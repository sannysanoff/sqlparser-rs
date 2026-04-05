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
	privileges, objects, err := parseGrantDenyRevokePrivilegesObjects(p)
	if err != nil {
		return nil, err
	}
	if _, err := p.ExpectKeyword("TO"); err != nil {
		return nil, err
	}
	grantees, err := parseGrantees(p)
	if err != nil {
		return nil, err
	}
	withGrantOption := p.ParseKeywords([]string{"WITH", "GRANT", "OPTION"})

	var currentGrants *statement.CurrentGrantsKind
	if p.ParseKeywords([]string{"COPY", "CURRENT", "GRANTS"}) {
		kind := statement.CurrentGrantsCopy
		currentGrants = &kind
	} else if p.ParseKeywords([]string{"REVOKE", "CURRENT", "GRANTS"}) {
		kind := statement.CurrentGrantsRevoke
		currentGrants = &kind
	}

	var asGrantor, grantedBy *ast.Ident
	if p.ParseKeyword("AS") {
		if asGrantor, err = p.ParseIdentifier(); err != nil {
			return nil, err
		}
	}
	if p.ParseKeywords([]string{"GRANTED", "BY"}) {
		if grantedBy, err = p.ParseIdentifier(); err != nil {
			return nil, err
		}
	}

	return &statement.Grant{
		Privileges: privileges, Objects: objects, Grantees: grantees,
		WithGrantOption: withGrantOption, AsGrantor: asGrantor,
		GrantedBy: grantedBy, CurrentGrants: currentGrants,
	}, nil
}

// parseRevoke parses REVOKE statements
// Reference: src/parser/mod.rs parse_revoke (line 17322)
func parseRevoke(p *Parser) (ast.Statement, error) {
	privileges, objects, err := parseGrantDenyRevokePrivilegesObjects(p)
	if err != nil {
		return nil, err
	}
	if _, err := p.ExpectKeyword("FROM"); err != nil {
		return nil, err
	}
	grantees, err := parseGrantees(p)
	if err != nil {
		return nil, err
	}
	var grantedBy *ast.Ident
	if p.ParseKeywords([]string{"GRANTED", "BY"}) {
		if grantedBy, err = p.ParseIdentifier(); err != nil {
			return nil, err
		}
	}
	return &statement.Revoke{
		Privileges: privileges, Objects: objects, Grantees: grantees,
		GrantedBy: grantedBy, Cascade: p.ParseKeyword("CASCADE"),
		Restrict: p.ParseKeyword("RESTRICT"),
	}, nil
}

// parseDeny parses DENY statements
// Reference: src/parser/mod.rs parse_deny (line 17289)
func parseDeny(p *Parser) (ast.Statement, error) {
	privileges, objects, err := parseGrantDenyRevokePrivilegesObjects(p)
	if err != nil {
		return nil, err
	}
	if objects == nil {
		return nil, fmt.Errorf("DENY statements must specify an object")
	}
	if _, err := p.ExpectKeyword("TO"); err != nil {
		return nil, err
	}
	grantees, err := parseGrantees(p)
	if err != nil {
		return nil, err
	}
	var cascadeOpt statement.CascadeOption
	if p.ParseKeyword("CASCADE") {
		cascadeOpt = statement.Cascade
	}
	var grantedBy *ast.Ident
	if p.ParseKeyword("AS") {
		if grantedBy, err = p.ParseIdentifier(); err != nil {
			return nil, err
		}
	}
	return &statement.DenyStatement{
		Privileges: privileges, Objects: objects, Grantees: grantees,
		GrantedBy: grantedBy, Cascade: cascadeOpt,
	}, nil
}

// parseGrantDenyRevokePrivilegesObjects parses privileges and objects for GRANT/DENY/REVOKE
// Reference: src/parser/mod.rs parse_grant_deny_revoke_privileges_objects (line 16807)
func parseGrantDenyRevokePrivilegesObjects(p *Parser) (*statement.Privileges, *statement.GrantObjects, error) {
	var privileges *statement.Privileges
	if p.ParseKeyword("ALL") {
		privileges = &statement.Privileges{All: true, WithPrivilegesKeyword: p.ParseKeyword("PRIVILEGES")}
	} else {
		actions, err := parseActionsList(p)
		if err != nil {
			return nil, nil, err
		}
		privileges = &statement.Privileges{All: false, Actions: actions}
	}

	if !p.ParseKeyword("ON") {
		return privileges, nil, nil
	}

	objects := &statement.GrantObjects{}
	parseAllInSchema := func(keywords []string, objectType statement.GrantObjectType) (bool, error) {
		if !p.ParseKeywords(keywords) {
			return false, nil
		}
		objects.ObjectType = objectType
		schemas, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return false, err
		}
		objects.Schemas = schemas
		return true, nil
	}
	parseSingleObjectType := func(keyword string, objectType statement.GrantObjectType) (bool, error) {
		if !p.ParseKeyword(keyword) {
			return false, nil
		}
		objects.ObjectType = objectType
		tables, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return false, err
		}
		objects.Tables = tables
		return true, nil
	}

	type allInSchema struct {
		keywords   []string
		objectType statement.GrantObjectType
	}
	for _, spec := range []allInSchema{
		{[]string{"ALL", "TABLES", "IN", "SCHEMA"}, statement.GrantObjectTypeAllTablesInSchema},
		{[]string{"ALL", "SEQUENCES", "IN", "SCHEMA"}, statement.GrantObjectTypeAllSequencesInSchema},
		{[]string{"ALL", "VIEWS", "IN", "SCHEMA"}, statement.GrantObjectTypeAllViewsInSchema},
		{[]string{"ALL", "MATERIALIZED", "VIEWS", "IN", "SCHEMA"}, statement.GrantObjectTypeAllMaterializedViewsInSchema},
		{[]string{"ALL", "EXTERNAL", "TABLES", "IN", "SCHEMA"}, statement.GrantObjectTypeAllExternalTablesInSchema},
		{[]string{"ALL", "FUNCTIONS", "IN", "SCHEMA"}, statement.GrantObjectTypeAllFunctionsInSchema},
	} {
		if matched, err := parseAllInSchema(spec.keywords, spec.objectType); err != nil {
			return nil, nil, err
		} else if matched {
			return privileges, objects, nil
		}
	}

	type singleObj struct {
		keyword    string
		objectType statement.GrantObjectType
	}
	for _, spec := range []singleObj{
		{"SEQUENCE", statement.GrantObjectTypeSequences},
		{"TABLE", statement.GrantObjectTypeTables},
		{"DATABASE", statement.GrantObjectTypeDatabases},
		{"SCHEMA", statement.GrantObjectTypeSchemas},
		{"VIEW", statement.GrantObjectTypeViews},
	} {
		if matched, err := parseSingleObjectType(spec.keyword, spec.objectType); err != nil {
			return nil, nil, err
		} else if matched {
			return privileges, objects, nil
		}
	}

	objects.ObjectType = statement.GrantObjectTypeTables
	tables, err := parseGrantObjectNames(p)
	if err != nil {
		return nil, nil, err
	}
	objects.Tables = tables
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
	ident, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	name := ast.NewObjectNameFromIdents(ident)
	if p.ConsumeToken(token.TokenPeriod{}) {
		if p.ConsumeToken(token.TokenMul{}) {
			name.Parts = append(name.Parts, &ast.ObjectNamePartIdentifier{Ident: &ast.Ident{Value: "*"}})
		} else {
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
	p.NextToken()

	action := &statement.Action{ActionType: actionType, RawKeyword: word.Word.Value}
	if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
		p.NextToken()
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
	granteeTypes := []struct {
		keywords []string
		kind     statement.GranteesType
	}{
		{[]string{"DATABASE", "ROLE"}, statement.GranteesTypeDatabaseRole},
		{[]string{"APPLICATION", "ROLE"}, statement.GranteesTypeApplicationRole},
		{[]string{"ROLE"}, statement.GranteesTypeRole},
		{[]string{"USER"}, statement.GranteesTypeUser},
		{[]string{"SHARE"}, statement.GranteesTypeShare},
		{[]string{"GROUP"}, statement.GranteesTypeGroup},
		{[]string{"PUBLIC"}, statement.GranteesTypePublic},
		{[]string{"APPLICATION"}, statement.GranteesTypeApplication},
	}

	for {
		newGranteeType := granteeType
		for _, gt := range granteeTypes {
			if p.ParseKeywords(gt.keywords) {
				newGranteeType = gt.kind
				break
			}
		}
		if newGranteeType != granteeType {
			granteeType = newGranteeType
		}

		var name *statement.GranteeName
		if granteeType != statement.GranteesTypePublic {
			var err error
			if name, err = parseGranteeName(p); err != nil {
				return nil, err
			}
		}
		grantees = append(grantees, &statement.Grantee{GranteeType: granteeType, Name: name})
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}
	return grantees, nil
}

// parseGranteeName parses a grantee name (identifier or 'user'@'host')
// Reference: src/parser/mod.rs parse_grantee_name (line 17273)
func parseGranteeName(p *Parser) (*statement.GranteeName, error) {
	ident, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	if p.ConsumeToken(token.TokenAtSign{}) {
		host, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		return &statement.GranteeName{User: ident, Host: host}, nil
	}
	return &statement.GranteeName{ObjectName: ast.NewObjectNameFromIdents(ident)}, nil
}

// parseSet parses SET statements
// Reference: src/parser/mod.rs:14784
func parseSet(p *Parser) (ast.Statement, error) {
	hivevar := p.ParseKeyword("HIVEVAR")
	if hivevar {
		if _, err := p.ExpectToken(token.TokenColon{}); err != nil {
			return nil, err
		}
	}
	var session, local, global bool
	if !hivevar {
		session = p.ParseKeyword("SESSION")
		if !session {
			local = p.ParseKeyword("LOCAL")
			if !local {
				global = p.ParseKeyword("GLOBAL")
			}
		}
	}

	if p.PeekKeyword("TIME") {
		if p.ParseKeyword("TIME") && p.ParseKeyword("ZONE") {
			exprParser := NewExpressionParser(p)
			val, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			return &statement.Set{Session: session, Local: local, TimeZone: true, Values: []expr.Expr{val}, HiveVar: hivevar}, nil
		}
		p.PrevToken()
	}

	if p.ParseKeyword("TRANSACTION") {
		setTrans := &statement.SetTransaction{Session: session, Local: local}
		if p.ParseKeyword("SNAPSHOT") {
			snapshot, err := NewExpressionParser(p).ParseExpr()
			if err != nil {
				return nil, err
			}
			setTrans.Snapshot = snapshot
			return setTrans, nil
		}
		modes, err := p.parseTransactionModes()
		if err != nil {
			return nil, err
		}
		setTrans.Modes = modes
		return setTrans, nil
	}

	if p.ParseKeyword("CHARACTERISTICS") {
		if err := p.ExpectKeywords([]string{"AS", "TRANSACTION"}); err != nil {
			return nil, err
		}
		modes, err := p.parseTransactionModes()
		if err != nil {
			return nil, err
		}
		return &statement.SetTransaction{Session: true, Modes: modes}, nil
	}

	if p.GetDialect().SupportsSetNames() && p.ParseKeyword("NAMES") {
		if p.ParseKeyword("DEFAULT") {
			return &statement.SetNames{CharsetName: "DEFAULT"}, nil
		}
		charset, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected charset name after SET NAMES: %w", err)
		}
		stmt := &statement.SetNames{CharsetName: charset.Value}
		if p.ParseKeyword("COLLATE") {
			collation, err := p.ParseIdentifier()
			if err != nil {
				return nil, fmt.Errorf("expected collation name after COLLATE: %w", err)
			}
			stmt.CollationName = &collation.Value
		}
		return stmt, nil
	}

	if p.ParseKeyword("AUTHORIZATION") {
		if !session && !local {
			return nil, fmt.Errorf("expected SESSION, LOCAL, or other scope modifier before AUTHORIZATION")
		}
		setAuth := &statement.SetSessionAuthorization{Session: session, Local: local}
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

	if p.ParseKeyword("ROLE") {
		setRole := &statement.SetRole{Session: session, Local: local}
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

	if p.GetDialect().SupportsParenthesizedSetVariables() {
		if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
			p.AdvanceToken()
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
			if !p.ParseKeyword("TO") {
				if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
					return nil, err
				}
			}
			if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
				return nil, err
			}
			ep := NewExpressionParser(p)
			values := []expr.Expr{}
			for {
				val, err := ep.ParseExpr()
				if err != nil {
					return nil, err
				}
				values = append(values, val)
				if !p.ConsumeToken(token.TokenComma{}) {
					break
				}
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
			setStmt := &statement.Set{Session: session, Local: local, Global: global, HiveVar: hivevar, Values: values}
			if len(vars) > 0 {
				setStmt.Variable = &ast.ObjectName{Parts: []ast.ObjectNamePart{&ast.ObjectNamePartIdentifier{Ident: &ast.Ident{Value: vars[0]}}}}
			}
			return setStmt, nil
		}
	}

	var varName *ast.ObjectName
	tok := p.PeekToken().Token
	if _, isAtSign := tok.(token.TokenAtSign); isAtSign {
		p.AdvanceToken()
		ident, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		varName = &ast.ObjectName{Parts: []ast.ObjectNamePart{&ast.ObjectNamePartIdentifier{Ident: ident}}}
	} else {
		objName, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		varName = objName
	}
	if !p.ParseKeyword("TO") {
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return nil, err
		}
	}
	ep := NewExpressionParser(p)
	values := []expr.Expr{}
	for {
		val, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		values = append(values, val)
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}
	return &statement.Set{Variable: varName, Values: values, Local: local, Session: session, Global: global, HiveVar: hivevar}, nil
}

// parseTransactionModes parses transaction mode options
func (p *Parser) parseTransactionModes() ([]expr.TransactionMode, error) {
	modes := []expr.TransactionMode{}
	for {
		matched := false
		if p.ParseKeyword("ISOLATION") {
			if err := p.ExpectKeywordIs("LEVEL"); err != nil {
				return nil, err
			}
			for _, il := range []struct {
				keywords []string
				mode     expr.TransactionMode
			}{
				{[]string{"READ", "UNCOMMITTED"}, expr.TransactionModeReadUncommitted},
				{[]string{"READ", "COMMITTED"}, expr.TransactionModeReadCommitted},
				{[]string{"REPEATABLE", "READ"}, expr.TransactionModeRepeatableRead},
				{[]string{"SERIALIZABLE"}, expr.TransactionModeSerializable},
				{[]string{"SNAPSHOT"}, expr.TransactionModeSnapshot},
			} {
				if p.ParseKeywords(il.keywords) {
					modes = append(modes, il.mode)
					matched = true
					break
				}
			}
		} else if p.ParseKeyword("READ") {
			if p.ParseKeyword("ONLY") {
				modes = append(modes, expr.TransactionModeReadOnly)
				matched = true
			} else if p.ParseKeyword("WRITE") {
				modes = append(modes, expr.TransactionModeReadWrite)
				matched = true
			}
		} else if p.ParseKeywords([]string{"NOT", "DEFERRABLE"}) {
			modes = append(modes, expr.TransactionModeNotDeferrable)
			matched = true
		} else if p.ParseKeyword("DEFERRABLE") {
			modes = append(modes, expr.TransactionModeDeferrable)
			matched = true
		}
		if !matched {
			break
		}
		p.ConsumeToken(token.TokenComma{})
	}
	return modes, nil
}

// parseUse parses USE statements
// Reference: src/parser/mod.rs:15226
func parseUse(p *Parser) (ast.Statement, error) {
	dialect := p.GetDialect().Dialect()
	if dialect == "hive" && p.ParseKeyword("DEFAULT") {
		return &statement.Use{Default: true}, nil
	}

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
			if !p.ParseKeyword("ROLES") && !p.ParseKeyword("ROLE") {
				return nil, fmt.Errorf("expected ROLES or ROLE after SECONDARY")
			}
			secRoles := &statement.SecondaryRoles{}
			if p.ParseKeyword("NONE") {
				secRoles.None = true
			} else if p.ParseKeyword("ALL") {
				secRoles.All = true
			} else {
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
	hasTableKeyword := p.ParseKeyword("TABLE")
	var tableName *ast.ObjectName
	if !p.PeekKeyword("FOR") && !p.PeekKeyword("PARTITION") && !p.PeekKeyword("CACHE") &&
		!p.PeekKeyword("NOSCAN") && !p.PeekKeyword("COMPUTE") {
		tableName, _ = p.ParseObjectName()
	}
	var columns []*ast.Ident
	if tableName != nil {
		if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
			p.AdvanceToken()
			cols, err := parseCommaSeparatedIdents(p)
			if err != nil {
				return nil, err
			}
			p.ExpectToken(token.TokenRParen{})
			columns = cols
		}
	}
	var forColumns, cacheMetadata, noscan, computeStatistics bool
	var partitions []expr.Expr
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
			if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
				p.AdvanceToken()
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
		HasTableKeyword: hasTableKeyword, TableName: tableName, ForColumns: forColumns,
		Columns: columns, Partitions: partitions, CacheMetadata: cacheMetadata,
		Noscan: noscan, ComputeStatistics: computeStatistics,
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
	var objectType expr.CommentObject
	for _, kt := range []struct {
		keywords   []string
		objectType expr.CommentObject
	}{
		{[]string{"MATERIALIZED", "VIEW"}, expr.CommentMaterializedView},
		{[]string{"TABLE"}, expr.CommentTable},
		{[]string{"VIEW"}, expr.CommentView},
		{[]string{"COLUMN"}, expr.CommentColumn},
		{[]string{"SCHEMA"}, expr.CommentSchema},
		{[]string{"DATABASE"}, expr.CommentDatabase},
		{[]string{"INDEX"}, expr.CommentIndex},
		{[]string{"SEQUENCE"}, expr.CommentSequence},
		{[]string{"TYPE"}, expr.CommentType},
		{[]string{"DOMAIN"}, expr.CommentDomain},
		{[]string{"FUNCTION"}, expr.CommentFunction},
		{[]string{"PROCEDURE"}, expr.CommentProcedure},
		{[]string{"ROLE"}, expr.CommentRole},
	} {
		if p.ParseKeywords(kt.keywords) {
			objectType = kt.objectType
			break
		}
	}
	if objectType == 0 {
		return nil, fmt.Errorf("unexpected object type for COMMENT")
	}
	objName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}
	p.ExpectKeyword("IS")
	var comment *string
	nextTok := p.PeekToken()
	if word, ok := nextTok.Token.(token.TokenWord); ok && word.Word.Keyword == "NULL" {
		p.AdvanceToken()
	} else {
		ep := NewExpressionParser(p)
		e, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		if val, ok := e.(*expr.ValueExpr); ok {
			if str, ok := val.Value.(string); ok {
				comment = &str
			}
		}
	}
	return &statement.Comment{ObjectType: objectType, ObjectName: objName, Comment: comment}, nil
}

// parseDeclare parses DECLARE statements
// Reference: src/parser/mod.rs:7486
func parseDeclare(p *Parser) (ast.Statement, error) {
	dialect := p.GetDialect()
	dialectName := dialect.Dialect()
	if dialectName == "bigquery" {
		return parseBigQueryDeclare(p)
	}
	if dialectName == "snowflake" {
		return parseSnowflakeDeclare(p)
	}
	if dialectName == "mssql" {
		return parseMssqlDeclare(p)
	}
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	var binary, sensitive, scroll *bool
	if p.ParseKeyword("BINARY") {
		b := true
		binary = &b
	}
	if p.ParseKeyword("INSENSITIVE") {
		s := true
		sensitive = &s
	} else if p.ParseKeyword("ASENSITIVE") {
		s := false
		sensitive = &s
	}
	if p.ParseKeyword("SCROLL") {
		s := true
		scroll = &s
	} else if p.ParseKeywords([]string{"NO", "SCROLL"}) {
		s := false
		scroll = &s
	}
	p.ExpectKeyword("CURSOR")
	declareType := expr.DeclareTypeCursor
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
	p.ExpectKeyword("FOR")
	stmt, err := p.parseQuery()
	if err != nil {
		return nil, err
	}
	q := extractQueryFromStatement(stmt)
	return &statement.Declare{Stmts: []*expr.Declare{{
		Names: []*expr.Ident{exprFromAstIdent(name)}, DeclareType: &declareType,
		Binary: binary, Sensitive: sensitive, Scroll: scroll, Hold: hold, ForQuery: q,
	}}}, nil
}

// parseBigQueryDeclare parses BigQuery DECLARE statements
// Reference: src/parser/mod.rs:7559
// Syntax: DECLARE variable_name[, ...] [{ <variable_type> | <DEFAULT expression> }];
func parseBigQueryDeclare(p *Parser) (ast.Statement, error) {
	names, err := parseCommaSeparatedIdents(p)
	if err != nil {
		return nil, err
	}
	var dataType interface{}
	nextTok := p.PeekToken()
	if word, ok := nextTok.Token.(token.TokenWord); !ok || word.Word.Keyword != "DEFAULT" {
		dt, err := p.ParseDataType()
		if err != nil {
			return nil, err
		}
		dataType = dt
	}
	var assignment expr.Expr
	var assignType expr.DeclareAssignment
	if dataType != nil && p.ParseKeyword("DEFAULT") {
		assignment, err = NewExpressionParser(p).ParseExpr()
		if err != nil {
			return nil, err
		}
		assignType = expr.DeclareAssignmentDefault
	} else if dataType == nil {
		p.ExpectKeyword("DEFAULT")
		assignment, err = NewExpressionParser(p).ParseExpr()
		if err != nil {
			return nil, err
		}
		assignType = expr.DeclareAssignmentDefault
	}
	return &statement.Declare{Stmts: []*expr.Declare{{
		Names: exprIdentsFromAstIdents(names), DataType: dataType,
		Assignment: assignment, AssignmentType: assignType,
	}}}, nil
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
		if p.ParseKeyword("CURSOR") {
			dt := expr.DeclareTypeCursor
			declareType = &dt
			p.ExpectKeyword("FOR")
			nextTok := p.PeekToken()
			if word, ok := nextTok.Token.(token.TokenWord); ok && word.Word.Keyword == "SELECT" {
				stmt, err := p.parseQuery()
				if err != nil {
					return nil, err
				}
				forQuery = extractQueryFromStatement(stmt)
			} else {
				assignment, err = NewExpressionParser(p).ParseExpr()
				if err != nil {
					return nil, err
				}
				assignType = expr.DeclareAssignmentFor
			}
		} else if p.ParseKeyword("RESULTSET") {
			dt := expr.DeclareTypeResultSet
			declareType = &dt
			if p.ParseKeyword("DEFAULT") || p.ParseKeyword(":=") {
				p.ExpectToken(token.TokenLParen{})
				stmt, err := p.parseQuery()
				if err != nil {
					return nil, err
				}
				forQuery = extractQueryFromStatement(stmt)
				p.ExpectToken(token.TokenRParen{})
				if assignType != expr.DeclareAssignmentDuckAssignment {
					assignType = expr.DeclareAssignmentDefault
				}
			}
		} else if p.ParseKeyword("EXCEPTION") {
			dt := expr.DeclareTypeException
			declareType = &dt
			if p.ConsumeToken(token.TokenLParen{}) {
				ep := NewExpressionParser(p)
				if _, err = ep.ParseExpr(); err != nil {
					return nil, err
				}
				p.ExpectToken(token.TokenComma{})
				if _, err = ep.ParseExpr(); err != nil {
					return nil, err
				}
				p.ExpectToken(token.TokenRParen{})
			}
		} else {
			nextTok := p.PeekToken()
			if word, ok := nextTok.Token.(token.TokenWord); ok && isDataTypeKeyword(string(word.Word.Keyword)) {
				dataType, err = p.ParseDataType()
				if err != nil {
					return nil, err
				}
			}
			if p.ParseKeyword("DEFAULT") {
				assignment, err = NewExpressionParser(p).ParseExpr()
				if err != nil {
					return nil, err
				}
				assignType = expr.DeclareAssignmentDefault
			} else if p.ConsumeToken(token.TokenEq{}) {
				p.ExpectToken(token.TokenEq{})
				assignment, err = NewExpressionParser(p).ParseExpr()
				if err != nil {
					return nil, err
				}
				assignType = expr.DeclareAssignmentDuckAssignment
			}
		}
		stmts = append(stmts, &expr.Declare{
			Names: []*expr.Ident{exprFromAstIdent(name)}, DataType: dataType,
			Assignment: assignment, AssignmentType: assignType,
			DeclareType: declareType, ForQuery: forQuery,
		})
		if !p.ConsumeToken(token.TokenSemiColon{}) {
			break
		}
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
	var stmts []*expr.Declare
	for {
		tok := p.PeekToken()
		if _, ok := tok.Token.(token.TokenAtSign); !ok {
			break
		}
		p.AdvanceToken()
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		name.Value = "@" + name.Value
		var dataType interface{}
		var assignment expr.Expr
		var assignType expr.DeclareAssignment
		if p.ParseKeyword("AS") {
			dataType, err = p.ParseDataType()
			if err != nil {
				return nil, err
			}
		} else if word, ok := p.PeekToken().Token.(token.TokenWord); ok && isDataTypeKeyword(string(word.Word.Keyword)) {
			dataType, err = p.ParseDataType()
			if err != nil {
				return nil, err
			}
		}
		if p.ConsumeToken(token.TokenEq{}) {
			assignment, err = NewExpressionParser(p).ParseExpr()
			if err != nil {
				return nil, err
			}
			assignType = expr.DeclareAssignmentMsSqlAssignment
		}
		stmts = append(stmts, &expr.Declare{
			Names: []*expr.Ident{exprFromAstIdent(name)}, DataType: dataType,
			Assignment: assignment, AssignmentType: assignType,
		})
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}
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
	if !p.ParseKeyword("CURSOR") {
		return nil, fmt.Errorf("expected CURSOR in MSSQL DECLARE statement")
	}
	declareType := expr.DeclareTypeCursor
	var scroll *bool
	if p.ParseKeyword("SCROLL") {
		s := true
		scroll = &s
	}
	p.ExpectKeyword("FOR")
	stmt, err := p.parseQuery()
	if err != nil {
		return nil, err
	}
	q := extractQueryFromStatement(stmt)
	return &statement.Declare{Stmts: []*expr.Declare{{
		Names: []*expr.Ident{exprFromAstIdent(name)}, DeclareType: &declareType,
		Scroll: scroll, ForQuery: q,
	}}}, nil
}

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
		return &query.Query{Body: &s.Select}
	default:
		return nil
	}
}

func isDataTypeKeyword(keyword string) bool {
	for _, dt := range []string{"INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT", "VARCHAR", "NVARCHAR", "CHAR", "NCHAR", "TEXT", "DECIMAL", "NUMERIC", "FLOAT", "REAL", "DOUBLE", "DATE", "TIME", "TIMESTAMP", "DATETIME", "BOOLEAN", "ARRAY", "STRUCT", "VARIANT", "OBJECT"} {
		if keyword == dt {
			return true
		}
	}
	return false
}

func exprFromAstIdent(ident *ast.Ident) *expr.Ident {
	return &expr.Ident{SpanVal: ident.Span(), Value: ident.Value, QuoteStyle: ident.QuoteStyle}
}

func exprIdentsFromAstIdents(idents []*ast.Ident) []*expr.Ident {
	result := make([]*expr.Ident, len(idents))
	for i, ident := range idents {
		result[i] = exprFromAstIdent(ident)
	}
	return result
}

func parseClose(p *Parser) (ast.Statement, error) {
	var cursor *expr.CloseCursor
	if p.ParseKeyword("ALL") {
		cursor = &expr.CloseCursor{Kind: expr.CloseCursorAll}
	} else {
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		cursor = &expr.CloseCursor{Kind: expr.CloseCursorSpecific, Name: exprFromAstIdent(name)}
	}
	return &statement.Close{Cursor: cursor}, nil
}

func parseFetch(p *Parser) (ast.Statement, error) {
	var direction *expr.FetchDirection
	type fetchDir struct {
		keyword string
		kind    expr.FetchDirectionKind
		hasExpr bool
	}
	fetchDirections := []fetchDir{
		{"NEXT", expr.FetchDirectionNext, false},
		{"PRIOR", expr.FetchDirectionPrior, false},
		{"FIRST", expr.FetchDirectionFirst, false},
		{"LAST", expr.FetchDirectionLast, false},
		{"ABSOLUTE", expr.FetchDirectionAbsolute, true},
		{"RELATIVE", expr.FetchDirectionRelative, true},
	}
	parsed := false
	for _, fd := range fetchDirections {
		if p.ParseKeyword(fd.keyword) {
			dir := &expr.FetchDirection{Kind: fd.kind}
			if fd.hasExpr {
				limit, err := NewExpressionParser(p).ParseExpr()
				if err != nil {
					return nil, err
				}
				dir.Limit = &limit
			}
			direction = dir
			parsed = true
			break
		}
	}
	if !parsed {
		if p.ParseKeyword("FORWARD") {
			if p.ParseKeyword("ALL") {
				direction = &expr.FetchDirection{Kind: expr.FetchDirectionForwardAll}
			} else {
				limit, err := NewExpressionParser(p).ParseExpr()
				if err != nil {
					return nil, err
				}
				direction = &expr.FetchDirection{Kind: expr.FetchDirectionForward, Limit: &limit}
			}
			parsed = true
		} else if p.ParseKeyword("BACKWARD") {
			if p.ParseKeyword("ALL") {
				direction = &expr.FetchDirection{Kind: expr.FetchDirectionBackwardAll}
			} else {
				limit, err := NewExpressionParser(p).ParseExpr()
				if err != nil {
					return nil, err
				}
				direction = &expr.FetchDirection{Kind: expr.FetchDirectionBackward, Limit: &limit}
			}
			parsed = true
		} else if p.ParseKeyword("ALL") {
			direction = &expr.FetchDirection{Kind: expr.FetchDirectionAll}
			parsed = true
		}
	}
	if !parsed {
		limit, err := NewExpressionParser(p).ParseExpr()
		if err != nil {
			return nil, err
		}
		direction = &expr.FetchDirection{Kind: expr.FetchDirectionCount, Limit: &limit}
	}
	var position *expr.FetchPosition
	if p.PeekKeyword("FROM") {
		p.AdvanceToken()
		pos := expr.FetchPositionFrom
		position = &pos
	} else if p.PeekKeyword("IN") {
		p.AdvanceToken()
		pos := expr.FetchPositionIn
		position = &pos
	} else {
		return nil, p.Expected("FROM or IN", p.PeekToken())
	}
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	var into *ast.ObjectName
	if p.ParseKeyword("INTO") {
		into, err = p.ParseObjectName()
		if err != nil {
			return nil, err
		}
	}
	return &statement.Fetch{Name: name, Direction: direction, Position: position, Into: into}, nil
}

func parseCache(p *Parser) (ast.Statement, error) {
	var tableFlag *ast.ObjectName
	var options []*expr.SqlOption
	var q *query.Query
	hasTableKeyword := p.ParseKeyword("TABLE")
	if hasTableKeyword {
		tableName, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		if p.ParseKeyword("OPTIONS") {
			p.ExpectToken(token.TokenLParen{})
			for {
				key, err := p.ParseIdentifier()
				if err != nil {
					return nil, err
				}
				p.ExpectToken(token.TokenEq{})
				val, err := NewExpressionParser(p).ParseExpr()
				if err != nil {
					return nil, err
				}
				options = append(options, &expr.SqlOption{Name: exprFromAstIdent(key), Value: val})
				if !p.ConsumeToken(token.TokenComma{}) {
					break
				}
			}
			p.ExpectToken(token.TokenRParen{})
		}
		hasAs := p.ParseKeyword("AS")
		if hasAs {
			stmt, err := p.parseQuery()
			if err != nil {
				return nil, err
			}
			q = extractQueryFromStatement(stmt)
		}
		return &statement.Cache{TableFlag: tableFlag, TableName: tableName, HasAs: hasAs, Options: options, Query: q}, nil
	}
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
	return &statement.Cache{TableFlag: tableFlag, TableName: tableName}, nil
}

func parseUncache(p *Parser) (ast.Statement, error) {
	p.ExpectKeyword("TABLE")
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}
	return &statement.Uncache{TableName: tableName, IfExists: ifExists}, nil
}

func parseMsck(p *Parser) (ast.Statement, error) {
	msck := &statement.Msck{}
	if p.ParseKeyword("REPAIR") {
		msck.RepairPartitions = true
	}
	if p.ParseKeyword("ADD") {
		msck.AddPartitions = true
	} else if p.ParseKeyword("DROP") {
		msck.DropPartitions = true
	} else if p.ParseKeyword("SYNC") {
		msck.SyncPartitions = true
	}
	p.ExpectKeyword("TABLE")
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}
	msck.TableName = tableName
	if p.ConsumeToken(token.TokenLParen{}) {
		var partitionSpec []expr.Expr
		for {
			expr, err := NewExpressionParser(p).ParseExpr()
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

func parseIfStatement(p *Parser) (ast.Statement, error) {
	p.ExpectKeyword("IF")
	ifCond, err := NewExpressionParser(p).ParseExpr()
	if err != nil {
		return nil, err
	}
	p.ExpectKeyword("THEN")
	ifStmts, err := parseConditionalStatements(p, []string{"ELSEIF", "ELSE", "END"})
	if err != nil {
		return nil, err
	}
	conditions := []*expr.IfStatementCondition{{Condition: ifCond, Statements: ifStmts}}
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
		conditions = append(conditions, &expr.IfStatementCondition{Condition: elseifCond, Statements: elseifStmts})
	}
	var elseClause *expr.IfStatementElse
	if p.ParseKeyword("ELSE") {
		elseStmts, err := parseConditionalStatements(p, []string{"END"})
		if err != nil {
			return nil, err
		}
		elseClause = &expr.IfStatementElse{Statements: elseStmts}
	}
	p.ExpectKeyword("END")
	p.ExpectKeyword("IF")
	return &statement.IfStatement{Conditions: conditions, Else: elseClause}, nil
}

func parseConditionalStatements(p *Parser, terminalKeywords []string) ([]ast.Statement, error) {
	var stmts []ast.Statement
	for {
		for p.ConsumeToken(token.TokenSemiColon{}) {
		}
		tok := p.PeekToken()
		if word, ok := tok.Token.(token.TokenWord); ok {
			kw := string(word.Word.Keyword)
			for _, term := range terminalKeywords {
				if kw == term {
					return stmts, nil
				}
			}
		}
		if _, ok := tok.Token.(token.EOF); ok {
			return stmts, nil
		}
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}
}
