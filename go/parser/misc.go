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
		objects.ObjectType = statement.GrantObjectTypeAllViewsInSchema // Treat as views for now
		schemas, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return nil, nil, err
		}
		objects.Schemas = schemas
	} else if p.ParseKeywords([]string{"ALL", "EXTERNAL", "TABLES", "IN", "SCHEMA"}) {
		// Snowflake: ALL EXTERNAL TABLES IN SCHEMA
		objects.ObjectType = statement.GrantObjectTypeAllTablesInSchema // Treat as tables for now
		schemas, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return nil, nil, err
		}
		objects.Schemas = schemas
	} else if p.ParseKeywords([]string{"ALL", "FUNCTIONS", "IN", "SCHEMA"}) {
		// Snowflake: ALL FUNCTIONS IN SCHEMA
		objects.ObjectType = statement.GrantObjectTypeAllTablesInSchema // Generic handling
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
func parseSet(p *Parser) (ast.Statement, error) {
	// SET [ SESSION | LOCAL | GLOBAL ] variable = value [, ...]
	// SET [ SESSION | LOCAL ] TIME ZONE { value | LOCAL | DEFAULT }

	setStmt := &statement.Set{}

	// Parse optional scope modifiers
	if p.ParseKeyword("SESSION") {
		setStmt.Session = true
	} else if p.ParseKeyword("LOCAL") {
		setStmt.Local = true
	} else if p.ParseKeyword("GLOBAL") {
		setStmt.Global = true
	}

	// Check for SET TIME ZONE
	if p.PeekKeyword("TIME") {
		// Save position in case this isn't TIME ZONE
		if p.ParseKeyword("TIME") && p.ParseKeyword("ZONE") {
			setStmt.TimeZone = true
			// Parse the timezone value
			exprParser := NewExpressionParser(p)
			val, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			setStmt.Values = []expr.Expr{val}
			return setStmt, nil
		}
		// Backtrack if it wasn't TIME ZONE
		p.PrevToken()
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
	setStmt.Variable = varName

	// Parse = or TO
	if !p.ParseKeyword("TO") {
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return nil, err
		}
	}

	// Parse one or more values (comma-separated)
	exprParser := NewExpressionParser(p)
	for {
		val, err := exprParser.ParseExpr()
		if err != nil {
			return nil, err
		}
		setStmt.Values = append(setStmt.Values, val)

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return setStmt, nil
}

// parseUse parses USE statements
func parseUse(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("USE statement parsing not yet fully implemented")
}

// parseAnalyze parses ANALYZE statements
func parseAnalyze(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ANALYZE statement parsing not yet fully implemented")
}

// parseCall parses CALL statements
func parseCall(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("CALL statement parsing not yet fully implemented")
}

// parseCache parses CACHE statements
func parseCache(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("CACHE statement parsing not yet fully implemented")
}

// parseUncache parses UNCACHE statements
func parseUncache(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("UNCACHE statement parsing not yet fully implemented")
}

// parseMsck parses MSCK statements
func parseMsck(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("MSCK statement parsing not yet fully implemented")
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
func parseComment(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("COMMENT statement parsing not yet fully implemented")
}

// parseDeclare parses DECLARE statements
func parseDeclare(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("DECLARE statement parsing not yet fully implemented")
}

// parseClose parses CLOSE statements
func parseClose(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("CLOSE statement parsing not yet fully implemented")
}

// parseFetch parses FETCH statements
func parseFetch(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("FETCH statement parsing not yet fully implemented")
}
