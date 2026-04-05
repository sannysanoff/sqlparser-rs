// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package parser

import (
	"fmt"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/query"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/token"
)

// parseGrant parses GRANT statements
// Reference: src/parser/mod.rs parse_grant (line 16697)
func parseGrant(p *Parser) (ast.Statement, error) {
	// Parse privileges and objects
	privileges, objects, err := parseGrantDenyRevokePrivilegesObjects(p)
	if err != nil {
		return nil, err
	}

	// Expect TO keyword
	if _, err := p.ExpectKeyword("TO"); err != nil {
		return nil, err
	}

	// Parse grantees
	grantees, err := parseGrantees(p)
	if err != nil {
		return nil, err
	}

	// Parse optional WITH GRANT OPTION
	withGrantOption := p.ParseKeywords([]string{"WITH", "GRANT", "OPTION"})

	// Parse optional COPY/REVOKE CURRENT GRANTS (Snowflake)
	var currentGrants *statement.CurrentGrantsKind
	if p.ParseKeywords([]string{"COPY", "CURRENT", "GRANTS"}) {
		kind := statement.CurrentGrantsCopy
		currentGrants = &kind
	} else if p.ParseKeywords([]string{"REVOKE", "CURRENT", "GRANTS"}) {
		kind := statement.CurrentGrantsRevoke
		currentGrants = &kind
	}

	// Parse optional AS clause
	var asGrantor *ast.Ident
	if p.ParseKeyword("AS") {
		ident, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		asGrantor = ident
	}

	// Parse optional GRANTED BY clause
	var grantedBy *ast.Ident
	if p.ParseKeywords([]string{"GRANTED", "BY"}) {
		ident, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		grantedBy = ident
	}

	return &statement.Grant{
		Privileges:      privileges,
		Objects:         objects,
		Grantees:        grantees,
		WithGrantOption: withGrantOption,
		AsGrantor:       asGrantor,
		GrantedBy:       grantedBy,
		CurrentGrants:   currentGrants,
	}, nil
}

// parseRevoke parses REVOKE statements
// Reference: src/parser/mod.rs parse_revoke (line 17322)
func parseRevoke(p *Parser) (ast.Statement, error) {
	// Parse privileges and objects
	privileges, objects, err := parseGrantDenyRevokePrivilegesObjects(p)
	if err != nil {
		return nil, err
	}

	// Expect FROM keyword
	if _, err := p.ExpectKeyword("FROM"); err != nil {
		return nil, err
	}

	// Parse grantees
	grantees, err := parseGrantees(p)
	if err != nil {
		return nil, err
	}

	// Parse optional GRANTED BY clause
	var grantedBy *ast.Ident
	if p.ParseKeywords([]string{"GRANTED", "BY"}) {
		ident, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		grantedBy = ident
	}

	// Parse optional CASCADE or RESTRICT
	cascade := p.ParseKeyword("CASCADE")
	restrict := p.ParseKeyword("RESTRICT")

	return &statement.Revoke{
		Privileges: privileges,
		Objects:    objects,
		Grantees:   grantees,
		GrantedBy:  grantedBy,
		Cascade:    cascade,
		Restrict:   restrict,
	}, nil
}

// parseDeny parses DENY statements
// Reference: src/parser/mod.rs parse_deny (line 17289)
func parseDeny(p *Parser) (ast.Statement, error) {
	// Parse privileges and objects
	privileges, objects, err := parseGrantDenyRevokePrivilegesObjects(p)
	if err != nil {
		return nil, err
	}

	// DENY requires objects to be specified
	if objects == nil {
		return nil, fmt.Errorf("DENY statements must specify an object")
	}

	// Expect TO keyword
	if _, err := p.ExpectKeyword("TO"); err != nil {
		return nil, err
	}

	// Parse grantees
	grantees, err := parseGrantees(p)
	if err != nil {
		return nil, err
	}

	// Parse optional CASCADE
	var cascadeOpt statement.CascadeOption
	if p.ParseKeyword("CASCADE") {
		cascadeOpt = statement.Cascade
	}

	// Parse optional AS clause (GRANTED BY in MSSQL)
	var grantedBy *ast.Ident
	if p.ParseKeyword("AS") {
		ident, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		grantedBy = ident
	}

	return &statement.DenyStatement{
		Privileges: privileges,
		Objects:    objects,
		Grantees:   grantees,
		GrantedBy:  grantedBy,
		Cascade:    cascadeOpt,
	}, nil
}

// parseGrantDenyRevokePrivilegesObjects parses privileges and objects for GRANT/DENY/REVOKE
// Reference: src/parser/mod.rs parse_grant_deny_revoke_privileges_objects (line 16807)
func parseGrantDenyRevokePrivilegesObjects(p *Parser) (*statement.Privileges, *statement.GrantObjects, error) {
	var privileges *statement.Privileges

	// Parse privilege level
	if p.ParseKeyword("ALL") {
		withPrivilegesKeyword := p.ParseKeyword("PRIVILEGES")
		privileges = &statement.Privileges{
			All:                   true,
			WithPrivilegesKeyword: withPrivilegesKeyword,
		}
	} else {
		// Parse specific actions
		actions, err := parseActionsList(p)
		if err != nil {
			return nil, nil, err
		}
		privileges = &statement.Privileges{
			All:     false,
			Actions: actions,
		}
	}

	// Parse optional ON clause for objects
	if !p.ParseKeyword("ON") {
		return privileges, nil, nil
	}

	// Parse object type
	objects := &statement.GrantObjects{}

	// Check for ALL TABLES IN SCHEMA, ALL SEQUENCES IN SCHEMA, etc.
	if p.ParseKeywords([]string{"ALL", "TABLES", "IN", "SCHEMA"}) {
		objects.ObjectType = statement.GrantObjectTypeAllTablesInSchema
		schemas, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return nil, nil, err
		}
		objects.Schemas = schemas
	} else if p.ParseKeywords([]string{"ALL", "SEQUENCES", "IN", "SCHEMA"}) {
		objects.ObjectType = statement.GrantObjectTypeAllSequencesInSchema
		schemas, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return nil, nil, err
		}
		objects.Schemas = schemas
	} else if p.ParseKeywords([]string{"ALL", "VIEWS", "IN", "SCHEMA"}) {
		objects.ObjectType = statement.GrantObjectTypeAllViewsInSchema
		schemas, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return nil, nil, err
		}
		objects.Schemas = schemas
	} else if p.ParseKeywords([]string{"ALL", "MATERIALIZED", "VIEWS", "IN", "SCHEMA"}) {
		// Snowflake: ALL MATERIALIZED VIEWS IN SCHEMA
		objects.ObjectType = statement.GrantObjectTypeAllMaterializedViewsInSchema
		schemas, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return nil, nil, err
		}
		objects.Schemas = schemas
	} else if p.ParseKeywords([]string{"ALL", "EXTERNAL", "TABLES", "IN", "SCHEMA"}) {
		// Snowflake: ALL EXTERNAL TABLES IN SCHEMA
		objects.ObjectType = statement.GrantObjectTypeAllExternalTablesInSchema
		schemas, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return nil, nil, err
		}
		objects.Schemas = schemas
	} else if p.ParseKeywords([]string{"ALL", "FUNCTIONS", "IN", "SCHEMA"}) {
		// Snowflake: ALL FUNCTIONS IN SCHEMA
		objects.ObjectType = statement.GrantObjectTypeAllFunctionsInSchema
		schemas, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return nil, nil, err
		}
		objects.Schemas = schemas
	} else if p.ParseKeyword("SEQUENCE") {
		objects.ObjectType = statement.GrantObjectTypeSequences
		tables, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return nil, nil, err
		}
		objects.Tables = tables
	} else if p.ParseKeyword("TABLE") {
		objects.ObjectType = statement.GrantObjectTypeTables
		tables, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return nil, nil, err
		}
		objects.Tables = tables
	} else if p.ParseKeyword("DATABASE") {
		objects.ObjectType = statement.GrantObjectTypeDatabases
		tables, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return nil, nil, err
		}
		objects.Tables = tables
	} else if p.ParseKeyword("SCHEMA") {
		objects.ObjectType = statement.GrantObjectTypeSchemas
		tables, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return nil, nil, err
		}
		objects.Tables = tables
	} else if p.ParseKeyword("VIEW") {
		objects.ObjectType = statement.GrantObjectTypeViews
		tables, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return nil, nil, err
		}
		objects.Tables = tables
	} else {
		// Regular table list - use parseGrantObjectNames to handle wildcards like foo.*
		objects.ObjectType = statement.GrantObjectTypeTables
		tables, err := parseGrantObjectNames(p)
		if err != nil {
			return nil, nil, err
		}
		objects.Tables = tables
	}

	return privileges, objects, nil
}

// parseGrantObjectNames parses object names for GRANT with wildcard support (e.g., schema.*)
// Reference: src/parser/mod.rs parse_grant_object_names (line 17003)
func parseGrantObjectNames(p *Parser) ([]*ast.ObjectName, error) {
	var names []*ast.ObjectName
	for {
		name, err := parseGrantObjectName(p)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}
	return names, nil
}

// parseGrantObjectName parses a single object name for GRANT (handles wildcards)
// Reference: src/parser/mod.rs parse_grant_object_name (line 16996)
func parseGrantObjectName(p *Parser) (*ast.ObjectName, error) {
	// Parse first identifier
	ident, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	name := ast.NewObjectNameFromIdents(ident)

	// Check for wildcard: schema.*
	if p.ConsumeToken(token.TokenPeriod{}) {
		if p.ConsumeToken(token.TokenMul{}) {
			// This is schema.* - add the wildcard as an identifier
			wildcard := &ast.Ident{Value: "*"}
			name.Parts = append(name.Parts, &ast.ObjectNamePartIdentifier{Ident: wildcard})
		} else {
			// Regular schema.table - parse the second identifier
			tableIdent, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			name.Parts = append(name.Parts, &ast.ObjectNamePartIdentifier{Ident: tableIdent})
		}
	}

	return name, nil
}

// parseActionsList parses a comma-separated list of privilege actions
// Reference: src/parser/mod.rs parse_actions_list
func parseActionsList(p *Parser) ([]*statement.Action, error) {
	var actions []*statement.Action

	for {
		action, err := parseGrantPermission(p)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return actions, nil
}

// parseGrantPermission parses a single privilege action
// Reference: src/parser/mod.rs parse_grant_permission (line 17021)
func parseGrantPermission(p *Parser) (*statement.Action, error) {
	tok := p.PeekToken()
	word, ok := tok.Token.(token.TokenWord)
	if !ok {
		return nil, p.expected("privilege name", tok)
	}

	actionType, found := statement.ParseActionType(word.Value)
	if !found {
		return nil, p.expected("privilege name", tok)
	}
	p.NextToken() // consume the keyword

	action := &statement.Action{
		ActionType: actionType,
		RawKeyword: word.Word.Value, // Preserve original keyword form
	}

	// Check for column list: SELECT(col1, col2)
	if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
		p.NextToken() // consume (
		columns, err := parseCommaSeparatedIdents(p)
		if err != nil {
			return nil, err
		}
		action.Columns = columns
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	return action, nil
}

// parseGrantees parses a comma-separated list of grantees
// Reference: src/parser/mod.rs parse_grantees (line 16738)
func parseGrantees(p *Parser) ([]*statement.Grantee, error) {
	var grantees []*statement.Grantee
	granteeType := statement.GranteesTypeNone

	for {
		newGranteeType := granteeType

		// Check for grantee type keywords
		if p.ParseKeyword("ROLE") {
			newGranteeType = statement.GranteesTypeRole
		} else if p.ParseKeyword("USER") {
			newGranteeType = statement.GranteesTypeUser
		} else if p.ParseKeyword("SHARE") {
			newGranteeType = statement.GranteesTypeShare
		} else if p.ParseKeyword("GROUP") {
			newGranteeType = statement.GranteesTypeGroup
		} else if p.ParseKeyword("PUBLIC") {
			newGranteeType = statement.GranteesTypePublic
		} else if p.ParseKeywords([]string{"DATABASE", "ROLE"}) {
			newGranteeType = statement.GranteesTypeDatabaseRole
		} else if p.ParseKeywords([]string{"APPLICATION", "ROLE"}) {
			newGranteeType = statement.GranteesTypeApplicationRole
		} else if p.ParseKeyword("APPLICATION") {
			newGranteeType = statement.GranteesTypeApplication
		}

		// Update grantee type if a new one was specified
		if newGranteeType != granteeType {
			granteeType = newGranteeType
		}

		// Handle PUBLIC grantee (no name needed)
		if granteeType == statement.GranteesTypePublic {
			grantees = append(grantees, &statement.Grantee{
				GranteeType: granteeType,
				Name:        nil,
			})
		} else {
			// Parse grantee name
			name, err := parseGranteeName(p)
			if err != nil {
				return nil, err
			}
			grantees = append(grantees, &statement.Grantee{
				GranteeType: granteeType,
				Name:        name,
			})
		}

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return grantees, nil
}

// parseGranteeName parses a grantee name (identifier or 'user'@'host')
// Reference: src/parser/mod.rs parse_grantee_name (line 17273)
func parseGranteeName(p *Parser) (*statement.GranteeName, error) {
	// Parse the first identifier (user name or object name)
	ident, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Check for @ symbol (MySQL user@host syntax)
	if p.ConsumeToken(token.TokenAtSign{}) {
		host, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		return &statement.GranteeName{
			User: ident,
			Host: host,
		}, nil
	}

	// Simple object name
	return &statement.GranteeName{
		ObjectName: ast.NewObjectNameFromIdents(ident),
	}, nil
}

// parseSet parses SET statements
// Reference: src/parser/mod.rs:14784
func parseSet(p *Parser) (ast.Statement, error) {
	// Check for HIVEVAR modifier first
	hivevar := p.ParseKeyword("HIVEVAR")
	if hivevar {
		// Expect colon after HIVEVAR
		if _, err := p.ExpectToken(token.TokenColon{}); err != nil {
			return nil, err
		}
	}

	// Parse optional scope modifiers (SESSION, LOCAL, GLOBAL)
	var session, local, global bool
	if !hivevar {
		if p.ParseKeyword("SESSION") {
			session = true
		} else if p.ParseKeyword("LOCAL") {
			local = true
		} else if p.ParseKeyword("GLOBAL") {
			global = true
		}
	}

	// Check for SET TIME ZONE
	if p.PeekKeyword("TIME") {
		if p.ParseKeyword("TIME") && p.ParseKeyword("ZONE") {
			// SET [ SESSION | LOCAL ] TIME ZONE { value | LOCAL | DEFAULT }
			exprParser := NewExpressionParser(p)
			val, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			return &statement.Set{
				Session:  session,
				Local:    local,
				TimeZone: true,
				Values:   []expr.Expr{val},
				HiveVar:  hivevar,
			}, nil
		}
		// Backtrack if it wasn't TIME ZONE
		p.PrevToken()
	}

	// Check for SET TRANSACTION SNAPSHOT or SET SESSION CHARACTERISTICS AS TRANSACTION
	if p.ParseKeyword("TRANSACTION") {
		setTrans := &statement.SetTransaction{
			Session: session,
			Local:   local,
		}

		// Check for SNAPSHOT
		if p.ParseKeyword("SNAPSHOT") {
			exprParser := NewExpressionParser(p)
			snapshot, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			setTrans.Snapshot = snapshot
			return setTrans, nil
		}

		// Parse transaction modes
		modes, err := p.parseTransactionModes()
		if err != nil {
			return nil, err
		}
		setTrans.Modes = modes
		return setTrans, nil
	}

	// Check for SET CHARACTERISTICS AS TRANSACTION
	if p.ParseKeyword("CHARACTERISTICS") {
		if err := p.ExpectKeywords([]string{"AS", "TRANSACTION"}); err != nil {
			return nil, err
		}
		setTrans := &statement.SetTransaction{
			Session: true, // SET CHARACTERISTICS AS TRANSACTION is always SESSION
		}
		modes, err := p.parseTransactionModes()
		if err != nil {
			return nil, err
		}
		setTrans.Modes = modes
		return setTrans, nil
	}

	// Check for SET NAMES (MySQL specific)
	if p.GetDialect().SupportsSetNames() && p.ParseKeyword("NAMES") {
		if p.ParseKeyword("DEFAULT") {
			return &statement.SetNames{
				CharsetName: "DEFAULT",
			}, nil
		}

		// Parse charset name
		charset, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected charset name after SET NAMES: %w", err)
		}

		setNamesStmt := &statement.SetNames{
			CharsetName: charset.Value,
		}

		// Parse optional COLLATE clause
		if p.ParseKeyword("COLLATE") {
			collation, err := p.ParseIdentifier()
			if err != nil {
				return nil, fmt.Errorf("expected collation name after COLLATE: %w", err)
			}
			collationStr := collation.Value
			setNamesStmt.CollationName = &collationStr
		}

		return setNamesStmt, nil
	}

	// Check for SET AUTHORIZATION
	if p.ParseKeyword("AUTHORIZATION") {
		// SET { SESSION | LOCAL } AUTHORIZATION { user_name | DEFAULT }
		if !session && !local {
			return nil, fmt.Errorf("expected SESSION, LOCAL, or other scope modifier before AUTHORIZATION")
		}

		setAuth := &statement.SetSessionAuthorization{
			Session: session,
			Local:   local,
		}

		if p.ParseKeyword("DEFAULT") {
			setAuth.Default = true
		} else {
			user, err := p.ParseIdentifier()
			if err != nil {
				return nil, fmt.Errorf("expected user name or DEFAULT after AUTHORIZATION: %w", err)
			}
			setAuth.User = user
		}
		return setAuth, nil
	}

	// Check for SET ROLE
	if p.ParseKeyword("ROLE") {
		// SET [ SESSION | LOCAL ] ROLE { role_name | NONE }
		setRole := &statement.SetRole{
			Session: session,
			Local:   local,
		}

		if p.ParseKeyword("NONE") {
			setRole.None = true
		} else {
			role, err := p.ParseIdentifier()
			if err != nil {
				return nil, fmt.Errorf("expected role name or NONE after ROLE: %w", err)
			}
			setRole.Role = role
		}
		return setRole, nil
	}

	// Check for parenthesized assignments: SET (a, b, c) = (1, 2, 3)
	if p.GetDialect().SupportsParenthesizedSetVariables() {
		if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
			// Parse parenthesized variable list
			p.AdvanceToken() // consume (
			vars := []string{}
			for {
				ident, err := p.ParseIdentifier()
				if err != nil {
					return nil, err
				}
				vars = append(vars, ident.Value)
				if !p.ConsumeToken(token.TokenComma{}) {
					break
				}
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}

			// Parse = or TO
			if !p.ParseKeyword("TO") {
				if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
					return nil, err
				}
			}

			// Expect (
			if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
				return nil, err
			}

			// Parse values
			exprParser := NewExpressionParser(p)
			values := []expr.Expr{}
			for {
				val, err := exprParser.ParseExpr()
				if err != nil {
					return nil, err
				}
				values = append(values, val)
				if !p.ConsumeToken(token.TokenComma{}) {
					break
				}
			}

			// Expect )
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}

			// Build SET statement with multiple variables
			setStmt := &statement.Set{
				Session: session,
				Local:   local,
				Global:  global,
				HiveVar: hivevar,
			}
			// Store the first variable name and values as a comma-separated list
			if len(vars) > 0 {
				setStmt.Variable = &ast.ObjectName{
					Parts: []ast.ObjectNamePart{
						&ast.ObjectNamePartIdentifier{Ident: &ast.Ident{Value: vars[0]}},
					},
				}
			}
			setStmt.Values = values
			return setStmt, nil
		}
	}

	// Standard SET variable = value [, ...]
	// Parse variable name - handle @variable syntax for MySQL/MS SQL
	var varName *ast.ObjectName
	tok := p.PeekToken().Token
	if _, isAtSign := tok.(token.TokenAtSign); isAtSign {
		p.AdvanceToken() // consume @
		ident, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		varName = &ast.ObjectName{
			Parts: []ast.ObjectNamePart{&ast.ObjectNamePartIdentifier{Ident: ident}},
		}
	} else {
		objName, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		varName = objName
	}

	// Parse = or TO
	if !p.ParseKeyword("TO") {
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return nil, err
		}
	}

	// Parse one or more values (comma-separated)
	exprParser := NewExpressionParser(p)
	values := []expr.Expr{}
	for {
		val, err := exprParser.ParseExpr()
		if err != nil {
			return nil, err
		}
		values = append(values, val)

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return &statement.Set{
		Variable: varName,
		Values:   values,
		Local:    local,
		Session:  session,
		Global:   global,
		HiveVar:  hivevar,
	}, nil
}

// parseTransactionModes parses transaction mode options
func (p *Parser) parseTransactionModes() ([]expr.TransactionMode, error) {
	modes := []expr.TransactionMode{}
	for {
		// Check for ISOLATION LEVEL
		if p.ParseKeyword("ISOLATION") {
			if err := p.ExpectKeywordIs("LEVEL"); err != nil {
				return nil, err
			}
			switch {
			case p.ParseKeyword("READ"):
				if p.ParseKeyword("UNCOMMITTED") {
					modes = append(modes, expr.TransactionModeReadUncommitted)
				} else if p.ParseKeyword("COMMITTED") {
					modes = append(modes, expr.TransactionModeReadCommitted)
				}
			case p.ParseKeyword("REPEATABLE"):
				if err := p.ExpectKeywordIs("READ"); err != nil {
					return nil, err
				}
				modes = append(modes, expr.TransactionModeRepeatableRead)
			case p.ParseKeyword("SERIALIZABLE"):
				modes = append(modes, expr.TransactionModeSerializable)
			case p.ParseKeyword("SNAPSHOT"):
				modes = append(modes, expr.TransactionModeSnapshot)
			}
		} else if p.ParseKeyword("READ") {
			// READ ONLY / READ WRITE
			if p.ParseKeyword("ONLY") {
				modes = append(modes, expr.TransactionModeReadOnly)
			} else if p.ParseKeyword("WRITE") {
				modes = append(modes, expr.TransactionModeReadWrite)
			}
		} else if p.ParseKeyword("NOT") {
			if p.ParseKeyword("DEFERRABLE") {
				modes = append(modes, expr.TransactionModeNotDeferrable)
			}
		} else if p.ParseKeyword("DEFERRABLE") {
			modes = append(modes, expr.TransactionModeDeferrable)
		} else {
			break
		}

		// Consume optional comma between modes
		p.ConsumeToken(token.TokenComma{})
	}
	return modes, nil
}

// parseUse parses USE statements
// Reference: src/parser/mod.rs:15226
func parseUse(p *Parser) (ast.Statement, error) {
	dialect := p.GetDialect().Dialect()

	// HiveDialect accepts USE DEFAULT
	if dialect == "hive" {
		if p.ParseKeyword("DEFAULT") {
			return &statement.Use{Default: true}, nil
		}
	}

	// Check for dialect-specific keywords
	var parsedKeyword string
	switch dialect {
	case "snowflake":
		if p.ParseKeyword("DATABASE") {
			parsedKeyword = "DATABASE"
		} else if p.ParseKeyword("SCHEMA") {
			parsedKeyword = "SCHEMA"
		} else if p.ParseKeyword("WAREHOUSE") {
			parsedKeyword = "WAREHOUSE"
		} else if p.ParseKeyword("ROLE") {
			parsedKeyword = "ROLE"
		} else if p.ParseKeyword("SECONDARY") {
			// Parse SECONDARY ROLES
			if !p.ParseKeyword("ROLES") && !p.ParseKeyword("ROLE") {
				return nil, fmt.Errorf("expected ROLES or ROLE after SECONDARY")
			}
			secRoles := &statement.SecondaryRoles{}
			if p.ParseKeyword("NONE") {
				secRoles.None = true
			} else if p.ParseKeyword("ALL") {
				secRoles.All = true
			} else {
				// Parse comma-separated role list
				roles := []*ast.Ident{}
				for {
					role, err := p.ParseIdentifier()
					if err != nil {
						return nil, err
					}
					roles = append(roles, role)
					if !p.ConsumeToken(token.TokenComma{}) {
						break
					}
				}
				secRoles.Roles = roles
			}
			return &statement.Use{SecondaryRoles: secRoles}, nil
		}
	case "databricks":
		if p.ParseKeyword("CATALOG") {
			parsedKeyword = "CATALOG"
		} else if p.ParseKeyword("DATABASE") {
			parsedKeyword = "DATABASE"
		} else if p.ParseKeyword("SCHEMA") {
			parsedKeyword = "SCHEMA"
		}
	}

	// Parse object name
	objName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	useStmt := &statement.Use{}
	switch parsedKeyword {
	case "CATALOG":
		useStmt.Catalog = objName
	case "DATABASE":
		useStmt.Database = objName
	case "SCHEMA":
		useStmt.Schema = objName
	case "WAREHOUSE":
		useStmt.Warehouse = objName
	case "ROLE":
		useStmt.Role = objName
	default:
		useStmt.Object = objName
	}

	return useStmt, nil
}

// parseAnalyze parses ANALYZE statements
// Reference: src/parser/mod.rs:1235-1297
func parseAnalyze(p *Parser) (ast.Statement, error) {
	// Check for TABLE keyword
	hasTableKeyword := p.ParseKeyword("TABLE")

	// Try to parse optional table name
	var tableName *ast.ObjectName
	if !p.PeekKeyword("FOR") && !p.PeekKeyword("PARTITION") && !p.PeekKeyword("CACHE") &&
		!p.PeekKeyword("NOSCAN") && !p.PeekKeyword("COMPUTE") {
		tableName, _ = p.ParseObjectName()
	}

	// Parse optional column list for PostgreSQL: ANALYZE t (col1, col2)
	var columns []*ast.Ident
	if tableName != nil {
		tok := p.PeekToken()
		if _, ok := tok.Token.(token.TokenLParen); ok {
			p.AdvanceToken() // consume (
			cols, err := parseCommaSeparatedIdents(p)
			if err != nil {
				return nil, err
			}
			p.ExpectToken(token.TokenRParen{})
			columns = cols
		}
	}

	var forColumns bool
	var cacheMetadata bool
	var noscan bool
	var partitions []expr.Expr
	var computeStatistics bool

	// Parse additional clauses in a loop
	for {
		switch {
		case p.ParseKeyword("PARTITION"):
			p.ExpectToken(token.TokenLParen{})
			parts, err := NewExpressionParser(p).parseCommaSeparatedExprs()
			if err != nil {
				return nil, err
			}
			partitions = parts
			p.ExpectToken(token.TokenRParen{})
		case p.ParseKeyword("NOSCAN"):
			noscan = true
		case p.ParseKeyword("FOR"):
			if !p.ParseKeyword("COLUMNS") {
				return nil, fmt.Errorf("Expected COLUMNS after FOR")
			}
			forColumns = true
			// Optional column list
			tok := p.PeekToken()
			if _, ok := tok.Token.(token.TokenLParen); ok {
				p.AdvanceToken() // consume (
				cols, err := parseCommaSeparatedIdents(p)
				if err != nil {
					return nil, err
				}
				columns = cols
				p.ExpectToken(token.TokenRParen{})
			}
		case p.ParseKeyword("CACHE"):
			if !p.ParseKeyword("METADATA") {
				return nil, fmt.Errorf("Expected METADATA after CACHE")
			}
			cacheMetadata = true
		case p.ParseKeyword("COMPUTE"):
			if !p.ParseKeyword("STATISTICS") {
				return nil, fmt.Errorf("Expected STATISTICS after COMPUTE")
			}
			computeStatistics = true
		default:
			goto done
		}
	}
done:

	return &statement.Analyze{
		HasTableKeyword:   hasTableKeyword,
		TableName:         tableName,
		ForColumns:        forColumns,
		Columns:           columns,
		Partitions:        partitions,
		CacheMetadata:     cacheMetadata,
		Noscan:            noscan,
		ComputeStatistics: computeStatistics,
	}, nil
}

// parseCall parses CALL statements
func parseCall(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("CALL statement parsing not yet fully implemented")
}

// parseFlush parses FLUSH statements
func parseFlush(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("FLUSH statement parsing not yet fully implemented")
}

// parseKill parses KILL statements
func parseKill(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("KILL statement parsing not yet fully implemented")
}

// parseVacuum parses VACUUM statements
func parseVacuum(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("VACUUM statement parsing not yet fully implemented")
}

// parseOptimize parses OPTIMIZE statements
func parseOptimize(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("OPTIMIZE statement parsing not yet fully implemented")
}

// parseLoad parses LOAD statements
func parseLoad(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("LOAD statement parsing not yet fully implemented")
}

// parseUnload parses UNLOAD statements
func parseUnload(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("UNLOAD statement parsing not yet fully implemented")
}

// parseAttach parses ATTACH statements
func parseAttach(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ATTACH statement parsing not yet fully implemented")
}

// parseDetach parses DETACH statements
func parseDetach(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("DETACH statement parsing not yet fully implemented")
}

// parseComment parses COMMENT statements
// Reference: src/parser/mod.rs:898
func parseComment(p *Parser) (ast.Statement, error) {
	p.ExpectKeyword("ON")

	// Parse the object type
	var objectType expr.CommentObject
	switch {
	case p.ParseKeyword("TABLE"):
		objectType = expr.CommentTable
	case p.ParseKeyword("VIEW"):
		objectType = expr.CommentView
	case p.ParseKeyword("COLUMN"):
		objectType = expr.CommentColumn
	case p.ParseKeyword("SCHEMA"):
		objectType = expr.CommentSchema
	case p.ParseKeyword("DATABASE"):
		objectType = expr.CommentDatabase
	case p.ParseKeyword("INDEX"):
		objectType = expr.CommentIndex
	case p.ParseKeyword("SEQUENCE"):
		objectType = expr.CommentSequence
	case p.ParseKeyword("MATERIALIZED") && p.ParseKeyword("VIEW"):
		objectType = expr.CommentMaterializedView
	case p.ParseKeyword("TYPE"):
		objectType = expr.CommentType
	case p.ParseKeyword("DOMAIN"):
		objectType = expr.CommentDomain
	case p.ParseKeyword("FUNCTION"):
		objectType = expr.CommentFunction
	case p.ParseKeyword("PROCEDURE"):
		objectType = expr.CommentProcedure
	case p.ParseKeyword("ROLE"):
		objectType = expr.CommentRole
	default:
		return nil, fmt.Errorf("unexpected object type for COMMENT")
	}

	// Parse object name
	objName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Expect IS keyword
	p.ExpectKeyword("IS")

	// Parse comment value (can be string literal or NULL)
	var comment *string
	nextTok := p.PeekToken()
	if word, ok := nextTok.Token.(token.TokenWord); ok && word.Word.Keyword == "NULL" {
		p.AdvanceToken()
		comment = nil
	} else {
		ep := NewExpressionParser(p)
		e, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		// Extract string value from the expression
		if val, ok := e.(*expr.ValueExpr); ok {
			if str, ok := val.Value.(string); ok {
				comment = &str
			}
		}
	}

	return &statement.Comment{
		ObjectType: objectType,
		ObjectName: objName,
		Comment:    comment,
	}, nil
}

// parseDeclare parses DECLARE statements
// Reference: src/parser/mod.rs:7486
func parseDeclare(p *Parser) (ast.Statement, error) {
	dialect := p.GetDialect()
	dialectName := dialect.Dialect()

	// Dispatch to dialect-specific parsers
	if dialectName == "bigquery" {
		return parseBigQueryDeclare(p)
	}
	if dialectName == "snowflake" {
		return parseSnowflakeDeclare(p)
	}
	if dialectName == "mssql" {
		return parseMssqlDeclare(p)
	}

	// Standard SQL cursor declaration (PostgreSQL-style)
	// DECLARE name [ BINARY ] [ ASENSITIVE | INSENSITIVE ] [ [ NO ] SCROLL ]
	//     CURSOR [ { WITH | WITHOUT } HOLD ] FOR query
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Parse optional BINARY
	var binary *bool
	if p.ParseKeyword("BINARY") {
		b := true
		binary = &b
	}

	// Parse ASENSITIVE | INSENSITIVE
	var sensitive *bool
	if p.ParseKeyword("INSENSITIVE") {
		s := true
		sensitive = &s
	} else if p.ParseKeyword("ASENSITIVE") {
		s := false
		sensitive = &s
	}

	// Parse [ NO ] SCROLL
	var scroll *bool
	if p.ParseKeyword("SCROLL") {
		s := true
		scroll = &s
	} else if p.ParseKeywords([]string{"NO", "SCROLL"}) {
		s := false
		scroll = &s
	}

	// Expect CURSOR
	p.ExpectKeyword("CURSOR")
	declareType := expr.DeclareTypeCursor

	// Parse { WITH | WITHOUT } HOLD
	var hold *bool
	if p.ParseKeyword("WITH") {
		p.ExpectKeyword("HOLD")
		h := true
		hold = &h
	} else if p.ParseKeyword("WITHOUT") {
		p.ExpectKeyword("HOLD")
		h := false
		hold = &h
	}

	// Expect FOR
	p.ExpectKeyword("FOR")

	// Parse query
	stmt, err := p.parseQuery()
	if err != nil {
		return nil, err
	}
	q := extractQueryFromStatement(stmt)

	return &statement.Declare{
		Stmts: []*expr.Declare{{
			Names:       []*expr.Ident{exprFromAstIdent(name)},
			DeclareType: &declareType,
			Binary:      binary,
			Sensitive:   sensitive,
			Scroll:      scroll,
			Hold:        hold,
			ForQuery:    q,
		}},
	}, nil
}

// parseBigQueryDeclare parses BigQuery DECLARE statements
// Reference: src/parser/mod.rs:7559
// Syntax: DECLARE variable_name[, ...] [{ <variable_type> | <DEFAULT expression> }];
func parseBigQueryDeclare(p *Parser) (ast.Statement, error) {
	// Parse comma-separated variable names
	names, err := parseCommaSeparatedIdents(p)
	if err != nil {
		return nil, err
	}

	// Check for data type or DEFAULT
	var dataType interface{}
	nextTok := p.PeekToken()
	if word, ok := nextTok.Token.(token.TokenWord); !ok || word.Word.Keyword != "DEFAULT" {
		// Parse data type
		dt, err := p.ParseDataType()
		if err != nil {
			return nil, err
		}
		dataType = dt
	}

	// Check for DEFAULT expression
	var assignment expr.Expr
	var assignType expr.DeclareAssignment
	if dataType != nil && p.ParseKeyword("DEFAULT") {
		ep := NewExpressionParser(p)
		assignment, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		assignType = expr.DeclareAssignmentDefault
	} else if dataType == nil {
		// No data type - DEFAULT expression is required
		p.ExpectKeyword("DEFAULT")
		ep := NewExpressionParser(p)
		assignment, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		assignType = expr.DeclareAssignmentDefault
	}

	return &statement.Declare{
		Stmts: []*expr.Declare{{
			Names:          exprIdentsFromAstIdents(names),
			DataType:       dataType,
			Assignment:     assignment,
			AssignmentType: assignType,
		}},
	}, nil
}

// parseSnowflakeDeclare parses Snowflake DECLARE statements
// Reference: src/parser/mod.rs:7619
func parseSnowflakeDeclare(p *Parser) (ast.Statement, error) {
	var stmts []*expr.Declare

	for {
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}

		var (
			declareType *expr.DeclareType
			forQuery    *query.Query
			assignment  expr.Expr
			assignType  expr.DeclareAssignment
			dataType    interface{}
		)

		// Check for CURSOR
		if p.ParseKeyword("CURSOR") {
			dt := expr.DeclareTypeCursor
			declareType = &dt
			p.ExpectKeyword("FOR")
			// Check if it's a SELECT query or a result set variable
			nextTok := p.PeekToken()
			if word, ok := nextTok.Token.(token.TokenWord); ok && word.Word.Keyword == "SELECT" {
				stmt, err := p.parseQuery()
				if err != nil {
					return nil, err
				}
				forQuery = extractQueryFromStatement(stmt)
			} else {
				// It's a reference to a result set variable
				ep := NewExpressionParser(p)
				assignment, err = ep.ParseExpr()
				if err != nil {
					return nil, err
				}
				assignType = expr.DeclareAssignmentFor
			}
		} else if p.ParseKeyword("RESULTSET") {
			// Result set declaration
			dt := expr.DeclareTypeResultSet
			declareType = &dt
			// Optional DEFAULT or := followed by ( query )
			if p.ParseKeyword("DEFAULT") || p.ParseKeyword(":=") {
				p.ExpectToken(token.TokenLParen{})
				stmt, err := p.parseQuery()
				if err != nil {
					return nil, err
				}
				forQuery = extractQueryFromStatement(stmt)
				p.ExpectToken(token.TokenRParen{})
				if assignType == expr.DeclareAssignmentDuckAssignment {
					assignType = expr.DeclareAssignmentDuckAssignment
				} else {
					assignType = expr.DeclareAssignmentDefault
				}
			}
		} else if p.ParseKeyword("EXCEPTION") {
			// Exception declaration
			dt := expr.DeclareTypeException
			declareType = &dt
			// Optional ( exception_number, 'exception_message' )
			if p.ConsumeToken(token.TokenLParen{}) {
				// Skip exception number for now
				ep := NewExpressionParser(p)
				_, err = ep.ParseExpr() // exception number
				if err != nil {
					return nil, err
				}
				p.ExpectToken(token.TokenComma{})
				// exception message
				_, err = ep.ParseExpr()
				if err != nil {
					return nil, err
				}
				p.ExpectToken(token.TokenRParen{})
			}
		} else {
			// Variable declaration with optional type and default
			nextTok := p.PeekToken()
			if word, ok := nextTok.Token.(token.TokenWord); ok {
				// Check if it's a type keyword
				if isDataTypeKeyword(string(word.Word.Keyword)) {
					dt, err := p.ParseDataType()
					if err != nil {
						return nil, err
					}
					dataType = dt
				}
			}

			// Check for assignment
			if p.ParseKeyword("DEFAULT") {
				ep := NewExpressionParser(p)
				assignment, err = ep.ParseExpr()
				if err != nil {
					return nil, err
				}
				assignType = expr.DeclareAssignmentDefault
			} else if p.ConsumeToken(token.TokenEq{}) {
				// Snowflake uses := not =
				p.ExpectToken(token.TokenEq{})
				ep := NewExpressionParser(p)
				assignment, err = ep.ParseExpr()
				if err != nil {
					return nil, err
				}
				assignType = expr.DeclareAssignmentDuckAssignment
			}
		}

		stmts = append(stmts, &expr.Declare{
			Names:          []*expr.Ident{exprFromAstIdent(name)},
			DataType:       dataType,
			Assignment:     assignment,
			AssignmentType: assignType,
			DeclareType:    declareType,
			ForQuery:       forQuery,
		})

		// Check for semicolon separator (Snowflake uses ; between declarations)
		if !p.ConsumeToken(token.TokenSemiColon{}) {
			break
		}

		// Check if next token is an identifier for another declaration
		nextTok := p.PeekToken()
		if _, ok := nextTok.Token.(token.TokenWord); !ok {
			break
		}
	}

	return &statement.Declare{Stmts: stmts}, nil
}

// parseMssqlDeclare parses MSSQL DECLARE statements
// Reference: src/parser/mod.rs:7722
func parseMssqlDeclare(p *Parser) (ast.Statement, error) {
	// MSSQL DECLARE can have multiple variables: DECLARE @a INT, @b VARCHAR(50)
	var stmts []*expr.Declare

	for {
		// Expect @variable_name
		tok := p.PeekToken()
		if _, ok := tok.Token.(token.TokenAtSign); !ok {
			// Not a variable, might be cursor declaration
			break
		}
		p.AdvanceToken() // consume @

		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}

		// Prepend @ to name
		name.Value = "@" + name.Value

		var dataType interface{}
		var assignment expr.Expr
		var assignType expr.DeclareAssignment

		// Check for AS keyword followed by data type
		if p.ParseKeyword("AS") {
			dt, err := p.ParseDataType()
			if err != nil {
				return nil, err
			}
			dataType = dt
		} else {
			// Try to parse data type directly
			nextTok := p.PeekToken()
			if word, ok := nextTok.Token.(token.TokenWord); ok && isDataTypeKeyword(string(word.Word.Keyword)) {
				dt, err := p.ParseDataType()
				if err != nil {
					return nil, err
				}
				dataType = dt
			}
		}

		// Check for assignment = expression
		if p.ConsumeToken(token.TokenEq{}) {
			ep := NewExpressionParser(p)
			assignment, err = ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			assignType = expr.DeclareAssignmentMsSqlAssignment
		}

		stmts = append(stmts, &expr.Declare{
			Names:          []*expr.Ident{exprFromAstIdent(name)},
			DataType:       dataType,
			Assignment:     assignment,
			AssignmentType: assignType,
		})

		// Check for comma separator
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	// If no variable declarations parsed, try cursor declaration
	if len(stmts) == 0 {
		return parseMssqlCursorDeclare(p)
	}

	return &statement.Declare{Stmts: stmts}, nil
}

// parseMssqlCursorDeclare parses MSSQL cursor declaration
func parseMssqlCursorDeclare(p *Parser) (ast.Statement, error) {
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Check for CURSOR keyword
	if !p.ParseKeyword("CURSOR") {
		return nil, fmt.Errorf("expected CURSOR in MSSQL DECLARE statement")
	}

	declareType := expr.DeclareTypeCursor

	// Parse optional cursor options
	var scroll *bool
	if p.ParseKeyword("SCROLL") {
		s := true
		scroll = &s
	}

	// Expect FOR
	p.ExpectKeyword("FOR")

	// Parse query
	stmt, err := p.parseQuery()
	if err != nil {
		return nil, err
	}
	q := extractQueryFromStatement(stmt)

	return &statement.Declare{
		Stmts: []*expr.Declare{{
			Names:       []*expr.Ident{exprFromAstIdent(name)},
			DeclareType: &declareType,
			Scroll:      scroll,
			ForQuery:    q,
		}},
	}, nil
}

// extractQueryFromStatement extracts a *query.Query from an ast.Statement
func extractQueryFromStatement(stmt ast.Statement) *query.Query {
	if stmt == nil {
		return nil
	}
	switch s := stmt.(type) {
	case *QueryStatement:
		return s.Query
	case *ValuesStatement:
		return s.Query
	case *SelectStatement:
		// Wrap Select in a Query
		return &query.Query{
			Body: &s.Select,
		}
	default:
		return nil
	}
}

// isDataTypeKeyword checks if a keyword is a data type
func isDataTypeKeyword(keyword string) bool {
	dataTypes := []string{
		"INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT",
		"VARCHAR", "NVARCHAR", "CHAR", "NCHAR", "TEXT",
		"DECIMAL", "NUMERIC", "FLOAT", "REAL", "DOUBLE",
		"DATE", "TIME", "TIMESTAMP", "DATETIME", "BOOLEAN",
		"ARRAY", "STRUCT", "VARIANT", "OBJECT", "VARIANT",
	}
	for _, dt := range dataTypes {
		if keyword == dt {
			return true
		}
	}
	return false
}

// Helper function to convert ast.Ident to expr.Ident
func exprFromAstIdent(ident *ast.Ident) *expr.Ident {
	return &expr.Ident{
		SpanVal:    ident.Span(),
		Value:      ident.Value,
		QuoteStyle: ident.QuoteStyle,
	}
}

// Helper function to convert []*ast.Ident to []*expr.Ident
func exprIdentsFromAstIdents(idents []*ast.Ident) []*expr.Ident {
	result := make([]*expr.Ident, len(idents))
	for i, ident := range idents {
		result[i] = exprFromAstIdent(ident)
	}
	return result
}

// parseClose parses CLOSE statements
// Reference: src/parser/mod.rs:parse_close
func parseClose(p *Parser) (ast.Statement, error) {
	var cursor *expr.CloseCursor

	if p.ParseKeyword("ALL") {
		cursor = &expr.CloseCursor{
			Kind: expr.CloseCursorAll,
		}
	} else {
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		cursor = &expr.CloseCursor{
			Kind: expr.CloseCursorSpecific,
			Name: exprFromAstIdent(name),
		}
	}

	return &statement.Close{Cursor: cursor}, nil
}

// parseFetch parses FETCH statements
// Reference: src/parser/mod.rs:7838
func parseFetch(p *Parser) (ast.Statement, error) {
	// Parse direction
	var direction *expr.FetchDirection

	if p.ParseKeyword("NEXT") {
		direction = &expr.FetchDirection{Kind: expr.FetchDirectionNext}
	} else if p.ParseKeyword("PRIOR") {
		direction = &expr.FetchDirection{Kind: expr.FetchDirectionPrior}
	} else if p.ParseKeyword("FIRST") {
		direction = &expr.FetchDirection{Kind: expr.FetchDirectionFirst}
	} else if p.ParseKeyword("LAST") {
		direction = &expr.FetchDirection{Kind: expr.FetchDirectionLast}
	} else if p.ParseKeyword("ABSOLUTE") {
		ep := NewExpressionParser(p)
		limit, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		direction = &expr.FetchDirection{Kind: expr.FetchDirectionAbsolute, Limit: &limit}
	} else if p.ParseKeyword("RELATIVE") {
		ep := NewExpressionParser(p)
		limit, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		direction = &expr.FetchDirection{Kind: expr.FetchDirectionRelative, Limit: &limit}
	} else if p.ParseKeyword("FORWARD") {
		if p.ParseKeyword("ALL") {
			direction = &expr.FetchDirection{Kind: expr.FetchDirectionForwardAll}
		} else {
			ep := NewExpressionParser(p)
			limit, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			direction = &expr.FetchDirection{Kind: expr.FetchDirectionForward, Limit: &limit}
		}
	} else if p.ParseKeyword("BACKWARD") {
		if p.ParseKeyword("ALL") {
			direction = &expr.FetchDirection{Kind: expr.FetchDirectionBackwardAll}
		} else {
			ep := NewExpressionParser(p)
			limit, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			direction = &expr.FetchDirection{Kind: expr.FetchDirectionBackward, Limit: &limit}
		}
	} else if p.ParseKeyword("ALL") {
		direction = &expr.FetchDirection{Kind: expr.FetchDirectionAll}
	} else {
		// Default: parse a count value
		ep := NewExpressionParser(p)
		limit, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		direction = &expr.FetchDirection{Kind: expr.FetchDirectionCount, Limit: &limit}
	}

	// Parse position (FROM or IN)
	var position *expr.FetchPosition
	if p.PeekKeyword("FROM") {
		p.AdvanceToken() // consume FROM
		pos := expr.FetchPositionFrom
		position = &pos
	} else if p.PeekKeyword("IN") {
		p.AdvanceToken() // consume IN
		pos := expr.FetchPositionIn
		position = &pos
	} else {
		return nil, p.Expected("FROM or IN", p.PeekToken())
	}

	// Parse cursor name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Parse optional INTO clause
	var into *ast.ObjectName
	if p.ParseKeyword("INTO") {
		into, err = p.ParseObjectName()
		if err != nil {
			return nil, err
		}
	}

	return &statement.Fetch{
		Name:      name,
		Direction: direction,
		Position:  position,
		Into:      into,
	}, nil
}

// parseCache parses CACHE TABLE statements
// Reference: src/parser/mod.rs:5277
func parseCache(p *Parser) (ast.Statement, error) {
	var tableFlag *ast.ObjectName
	var options []*expr.SqlOption
	hasAs := false
	var q *query.Query

	// Check for optional TABLE keyword
	hasTableKeyword := p.ParseKeyword("TABLE")

	if hasTableKeyword {
		// CACHE TABLE table_name ...
		tableName, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}

		// Parse optional OPTIONS
		if p.ParseKeyword("OPTIONS") {
			p.ExpectToken(token.TokenLParen{})
			for {
				key, err := p.ParseIdentifier()
				if err != nil {
					return nil, err
				}
				p.ExpectToken(token.TokenEq{})
				ep := NewExpressionParser(p)
				val, err := ep.ParseExpr()
				if err != nil {
					return nil, err
				}
				options = append(options, &expr.SqlOption{
					Name:  exprFromAstIdent(key),
					Value: val,
				})
				if !p.ConsumeToken(token.TokenComma{}) {
					break
				}
			}
			p.ExpectToken(token.TokenRParen{})
		}

		// Parse optional AS query
		nextTok := p.PeekToken()
		if _, ok := nextTok.Token.(token.EOF); !ok {
			if p.ParseKeyword("AS") {
				hasAs = true
				stmt, err := p.parseQuery()
				if err != nil {
					return nil, err
				}
				q = extractQueryFromStatement(stmt)
			}
		}

		return &statement.Cache{
			TableFlag: tableFlag,
			TableName: tableName,
			HasAs:     hasAs,
			Options:   options,
			Query:     q,
		}, nil
	}

	// CACHE table_name TABLE table_name ... (rare syntax)
	tf, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}
	tableFlag = tf

	if !p.ParseKeyword("TABLE") {
		return nil, fmt.Errorf("expected TABLE after table flag in CACHE statement")
	}

	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	return &statement.Cache{
		TableFlag: tableFlag,
		TableName: tableName,
		HasAs:     hasAs,
		Options:   options,
		Query:     q,
	}, nil
}

// parseUncache parses UNCACHE TABLE statements
// Reference: src/parser/mod.rs: parse_uncache_table (around line 5335)
func parseUncache(p *Parser) (ast.Statement, error) {
	p.ExpectKeyword("TABLE")

	// Parse optional IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	return &statement.Uncache{
		TableName: tableName,
		IfExists:  ifExists,
	}, nil
}

// parseMsck parses MSCK REPAIR TABLE statements (Hive)
// Reference: src/parser/mod.rs:1063
func parseMsck(p *Parser) (ast.Statement, error) {
	msck := &statement.Msck{}

	// Parse optional REPAIR
	if p.ParseKeyword("REPAIR") {
		msck.RepairPartitions = true
	}

	// Check for ADD/DROP/SYNC PARTITIONS
	if p.ParseKeyword("ADD") {
		msck.AddPartitions = true
	} else if p.ParseKeyword("DROP") {
		msck.DropPartitions = true
	} else if p.ParseKeyword("SYNC") {
		msck.SyncPartitions = true
	}

	// Expect TABLE
	p.ExpectKeyword("TABLE")

	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}
	msck.TableName = tableName

	// Parse optional partition specification
	if p.ConsumeToken(token.TokenLParen{}) {
		var partitionSpec []expr.Expr
		for {
			ep := NewExpressionParser(p)
			expr, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			partitionSpec = append(partitionSpec, expr)
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		p.ExpectToken(token.TokenRParen{})
		msck.PartitionSpec = partitionSpec
	}

	return msck, nil
}

// parseIfStatement parses IF statements
// Reference: src/parser/mod.rs:772-807
func parseIfStatement(p *Parser) (ast.Statement, error) {
	p.ExpectKeyword("IF")

	var conditions []*expr.IfStatementCondition

	// Parse initial IF condition
	ifCond, err := NewExpressionParser(p).ParseExpr()
	if err != nil {
		return nil, err
	}

	p.ExpectKeyword("THEN")

	// Parse statements until we hit ELSEIF, ELSE, or END
	ifStmts, err := parseConditionalStatements(p, []string{"ELSEIF", "ELSE", "END"})
	if err != nil {
		return nil, err
	}

	conditions = append(conditions, &expr.IfStatementCondition{
		Condition:  ifCond,
		Statements: ifStmts,
	})

	// Parse optional ELSEIF blocks
	for p.ParseKeyword("ELSEIF") {
		elseifCond, err := NewExpressionParser(p).ParseExpr()
		if err != nil {
			return nil, err
		}

		p.ExpectKeyword("THEN")

		elseifStmts, err := parseConditionalStatements(p, []string{"ELSEIF", "ELSE", "END"})
		if err != nil {
			return nil, err
		}

		conditions = append(conditions, &expr.IfStatementCondition{
			Condition:  elseifCond,
			Statements: elseifStmts,
		})
	}

	// Parse optional ELSE block
	var elseClause *expr.IfStatementElse
	if p.ParseKeyword("ELSE") {
		elseStmts, err := parseConditionalStatements(p, []string{"END"})
		if err != nil {
			return nil, err
		}
		elseClause = &expr.IfStatementElse{
			Statements: elseStmts,
		}
	}

	// Expect END IF
	p.ExpectKeyword("END")
	p.ExpectKeyword("IF")

	return &statement.IfStatement{
		Conditions: conditions,
		Else:       elseClause,
	}, nil
}

// parseConditionalStatements parses a sequence of statements until one of the terminal keywords is encountered
func parseConditionalStatements(p *Parser, terminalKeywords []string) ([]ast.Statement, error) {
	var stmts []ast.Statement

	for {
		// Skip any semicolons (statement separators)
		for p.ConsumeToken(token.TokenSemiColon{}) {
			// Keep consuming semicolons
		}

		// Check for terminal keywords
		tok := p.PeekToken()
		if word, ok := tok.Token.(token.TokenWord); ok {
			kw := string(word.Word.Keyword)
			for _, term := range terminalKeywords {
				if kw == term {
					return stmts, nil
				}
			}
		}

		// Check for EOF
		if _, ok := tok.Token.(token.EOF); ok {
			return stmts, nil
		}

		// Parse next statement
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}
}
