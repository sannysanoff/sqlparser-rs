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
	"github.com/user/sqlparser/tokenizer"
)

// ParseDrop parses DROP statements
// Reference: src/parser/mod.rs parse_drop
func ParseDrop(p *Parser) (ast.Statement, error) {
	// Check for TEMPORARY (MySQL style)
	temporary := p.ParseKeyword("TEMPORARY")

	// Determine object type based on next keyword
	switch {
	case p.PeekKeyword("TABLE"):
		return parseDropTable(p, temporary)
	case p.PeekKeyword("VIEW"):
		return parseDropView(p, temporary)
	case p.PeekKeyword("INDEX"):
		return parseDropIndex(p)
	case p.PeekKeyword("ROLE"):
		return parseDropRole(p)
	case p.PeekKeyword("DATABASE"):
		return parseDropDatabase(p)
	case p.PeekKeyword("SCHEMA"):
		return parseDropSchema(p)
	case p.PeekKeyword("SEQUENCE"):
		return parseDropSequence(p)
	case p.PeekKeyword("FUNCTION"):
		return parseDropFunction(p)
	default:
		return nil, p.ExpectedRef("TABLE, VIEW, INDEX, ROLE, DATABASE, SCHEMA, SEQUENCE, or FUNCTION after DROP", p.PeekTokenRef())
	}
}

// parseDropView parses DROP VIEW
// Reference: src/parser/mod.rs parse_drop (VIEW branch)
func parseDropView(p *Parser, temporary bool) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("VIEW"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated view names
	names, err := parseCommaSeparatedObjectNames(p)
	if err != nil {
		return nil, err
	}

	// Parse optional CASCADE/RESTRICT
	cascade := p.ParseKeyword("CASCADE")
	restrict := p.ParseKeyword("RESTRICT")

	return &statement.DropView{
		IfExists: ifExists,
		Names:    names,
		Cascade:  cascade,
		Restrict: restrict,
	}, nil
}

// parseDropTable parses DROP TABLE
func parseDropTable(p *Parser, temporary bool) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("TABLE"); err != nil {
		return nil, err
	}

	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	names, err := parseCommaSeparatedObjectNames(p)
	if err != nil {
		return nil, err
	}

	cascade := p.ParseKeyword("CASCADE")
	restrict := p.ParseKeyword("RESTRICT")

	return &statement.DropTable{
		Temporary: temporary,
		IfExists:  ifExists,
		Names:     names,
		Cascade:   cascade,
		Restrict:  restrict,
	}, nil
}

// parseGrantObjectNames parses a comma-separated list of object names for GRANT/REVOKE
// This version handles wildcards like "foo.*" or "*.*"
func parseGrantObjectNames(p *Parser) ([]*ast.ObjectName, error) {
	var names []*ast.ObjectName
	for {
		name, err := parseGrantObjectName(p)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}
	return names, nil
}

// parseGrantObjectName parses a single object name that may contain wildcards
// Handles patterns like "foo.*" or "*.*"
func parseGrantObjectName(p *Parser) (*ast.ObjectName, error) {
	var parts []ast.ObjectNamePart

	for {
		tok := p.PeekToken()

		// Handle wildcard *
		if _, ok := tok.Token.(tokenizer.TokenMul); ok {
			p.NextToken() // consume *
			// Create a wildcard identifier
			parts = append(parts, &ast.ObjectNamePartIdentifier{
				Ident: &ast.Ident{Value: "*"},
			})
		} else {
			// Parse regular identifier
			ident, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			parts = append(parts, &ast.ObjectNamePartIdentifier{Ident: ident})
		}

		// Check if there's a period - if so, continue to next part
		if p.ConsumeToken(tokenizer.TokenPeriod{}) {
			continue
		}

		// No more parts
		break
	}

	return &ast.ObjectName{Parts: parts}, nil
}
func parseCommaSeparatedObjectNames(p *Parser) ([]*ast.ObjectName, error) {
	var names []*ast.ObjectName
	for {
		name, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		names = append(names, name)
		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}
	return names, nil
}

// parseCommaSeparatedIdents parses a comma-separated list of identifiers
func parseCommaSeparatedIdents(p *Parser) ([]*ast.Ident, error) {
	var names []*ast.Ident
	for {
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		names = append(names, name)
		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}
	return names, nil
}

// parseDropIndex parses DROP INDEX
// Reference: src/parser/mod.rs parse_drop (INDEX branch)
// DROP INDEX [CONCURRENTLY] [IF EXISTS] name [, ...] [CASCADE | RESTRICT]
func parseDropIndex(p *Parser) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("INDEX"); err != nil {
		return nil, err
	}

	// Parse CONCURRENTLY (PostgreSQL specific)
	concurrently := p.ParseKeyword("CONCURRENTLY")

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated list of index names
	var names []*ast.ObjectName
	for {
		name, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		names = append(names, name)

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	// Parse optional CASCADE or RESTRICT
	cascade := p.ParseKeyword("CASCADE")
	restrict := p.ParseKeyword("RESTRICT")

	if cascade && restrict {
		return nil, fmt.Errorf("cannot specify both CASCADE and RESTRICT in DROP INDEX")
	}

	return &statement.DropIndex{
		Concurrently: concurrently,
		IfExists:     ifExists,
		Names:        names,
		Cascade:      cascade,
		Restrict:     restrict,
	}, nil
}

func parseDropRole(p *Parser) (ast.Statement, error) {
	return nil, p.ExpectedRef("DROP ROLE not yet implemented", p.PeekTokenRef())
}

func parseDropDatabase(p *Parser) (ast.Statement, error) {
	return nil, p.ExpectedRef("DROP DATABASE not yet implemented", p.PeekTokenRef())
}

func parseDropSchema(p *Parser) (ast.Statement, error) {
	// SCHEMA keyword is expected (already checked by caller)
	if _, err := p.ExpectKeyword("SCHEMA"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated schema names
	names, err := parseCommaSeparatedObjectNames(p)
	if err != nil {
		return nil, err
	}

	// Parse optional CASCADE/RESTRICT
	cascade := p.ParseKeyword("CASCADE")
	restrict := p.ParseKeyword("RESTRICT")

	return &statement.DropSchema{
		IfExists: ifExists,
		Names:    names,
		Cascade:  cascade,
		Restrict: restrict,
	}, nil
}

func parseDropSequence(p *Parser) (ast.Statement, error) {
	// Consume SEQUENCE keyword (already checked by caller)
	if _, err := p.ExpectKeyword("SEQUENCE"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated list of sequence names
	var names []*ast.ObjectName
	for {
		name, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		names = append(names, name)

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	// Parse optional CASCADE or RESTRICT
	cascade := p.ParseKeyword("CASCADE")
	restrict := p.ParseKeyword("RESTRICT")

	if cascade && restrict {
		return nil, fmt.Errorf("cannot specify both CASCADE and RESTRICT")
	}

	return &statement.DropSequence{
		IfExists: ifExists,
		Names:    names,
		Cascade:  cascade,
		Restrict: restrict,
	}, nil
}

func parseDropFunction(p *Parser) (ast.Statement, error) {
	return nil, p.ExpectedRef("DROP FUNCTION not yet implemented", p.PeekTokenRef())
}

// ParseTruncate parses TRUNCATE statements
func ParseTruncate(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("TRUNCATE statement parsing not yet fully implemented")
}

// parseDescribe parses DESCRIBE statements
func parseDescribe(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("DESCRIBE statement parsing not yet fully implemented")
}

// parseShow parses SHOW statements
// parseShow parses SHOW statements
func parseShow(p *Parser) (ast.Statement, error) {
	// Parse optional modifiers that can appear before the object type
	var full, extended, global, session bool

	// Try to parse optional modifiers
	for {
		switch {
		case !full && p.PeekKeyword("FULL"):
			full = p.ParseKeyword("FULL")
		case !extended && p.PeekKeyword("EXTENDED"):
			extended = p.ParseKeyword("EXTENDED")
		case !global && p.PeekKeyword("GLOBAL"):
			global = p.ParseKeyword("GLOBAL")
		case !session && p.PeekKeyword("SESSION"):
			session = p.ParseKeyword("SESSION")
		default:
			goto doneModifiers
		}
	}
doneModifiers:

	// Check for CREATE (e.g., SHOW CREATE TABLE)
	if p.PeekKeyword("CREATE") {
		return parseShowCreate(p)
	}

	// Determine the object type
	switch {
	case p.PeekKeyword("COLUMNS") || p.PeekKeyword("FIELDS"):
		return parseShowColumns(p, extended, full)
	case p.PeekKeyword("TABLES"):
		return parseShowTables(p, extended, full)
	case p.PeekKeyword("STATUS"):
		return parseShowStatus(p, global, session)
	case p.PeekKeyword("VARIABLES"):
		return parseShowVariables(p, global, session)
	case p.PeekKeyword("COLLATION"):
		return parseShowCollation(p)
	case p.PeekKeyword("CHARSET") || p.PeekKeyword("CHARACTER"):
		return parseShowCharset(p)
	default:
		return nil, p.ExpectedRef("COLUMNS, FIELDS, TABLES, STATUS, VARIABLES, CREATE, COLLATION, CHARSET, or CHARACTER after SHOW", p.PeekTokenRef())
	}
}

// parseShowColumns parses SHOW COLUMNS/FIELDS statements
func parseShowColumns(p *Parser, extended, full bool) (ast.Statement, error) {
	// Consume COLUMNS or FIELDS
	p.AdvanceToken()

	options := &expr.ShowStatementOptions{}

	// Parse FROM or IN clause
	if p.PeekKeyword("FROM") || p.PeekKeyword("IN") {
		clause := expr.ShowStatementInClauseFrom
		if p.PeekKeyword("IN") {
			clause = expr.ShowStatementInClauseIn
		}
		p.AdvanceToken()

		// Parse table name
		tableName, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}

		options.ShowIn = &expr.ShowStatementIn{
			Clause:     clause,
			ParentName: tableName,
		}

		// Check for optional FROM db_name
		if p.PeekKeyword("FROM") || p.PeekKeyword("IN") {
			p.AdvanceToken()
			dbName, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			// Combine db.table
			options.ShowIn.ParentName = &ast.ObjectName{
				Parts: []ast.ObjectNamePart{
					&ast.ObjectNamePartIdentifier{Ident: dbName},
					&ast.ObjectNamePartIdentifier{Ident: &ast.Ident{Value: tableName.Parts[0].(*ast.ObjectNamePartIdentifier).Ident.Value}},
				},
			}
		}
	}

	// Parse optional LIKE or WHERE clause
	if p.PeekKeyword("LIKE") {
		p.AdvanceToken()
		likePattern, err := p.ParseLikePattern()
		if err != nil {
			return nil, err
		}
		options.Filter = &expr.ShowStatementFilter{Like: &likePattern}
		options.FilterPosition = expr.ShowStatementFilterPositionSuffix
	} else if p.PeekKeyword("WHERE") {
		p.AdvanceToken()
		exprParser := NewExpressionParser(p)
		whereExpr, err := exprParser.ParseExpr()
		if err != nil {
			return nil, err
		}
		options.Filter = &expr.ShowStatementFilter{Where: whereExpr}
		options.FilterPosition = expr.ShowStatementFilterPositionSuffix
	}

	return &statement.ShowColumns{
		Extended:    extended,
		Full:        full,
		ShowOptions: options,
	}, nil
}

// parseShowTables parses SHOW TABLES statements
func parseShowTables(p *Parser, extended, full bool) (ast.Statement, error) {
	p.AdvanceToken()

	options := &expr.ShowStatementOptions{}

	// Parse optional FROM/IN clause
	if p.PeekKeyword("FROM") || p.PeekKeyword("IN") {
		clause := expr.ShowStatementInClauseFrom
		if p.PeekKeyword("IN") {
			clause = expr.ShowStatementInClauseIn
		}
		p.AdvanceToken()
		dbName, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		options.ShowIn = &expr.ShowStatementIn{
			Clause:     clause,
			ParentName: &ast.ObjectName{Parts: []ast.ObjectNamePart{&ast.ObjectNamePartIdentifier{Ident: dbName}}},
		}
	}

	// Parse optional LIKE or WHERE clause
	if p.PeekKeyword("LIKE") {
		p.AdvanceToken()
		likePattern, err := p.ParseLikePattern()
		if err != nil {
			return nil, err
		}
		options.Filter = &expr.ShowStatementFilter{Like: &likePattern}
	} else if p.PeekKeyword("WHERE") {
		p.AdvanceToken()
		exprParser := NewExpressionParser(p)
		whereExpr, err := exprParser.ParseExpr()
		if err != nil {
			return nil, err
		}
		options.Filter = &expr.ShowStatementFilter{Where: whereExpr}
	}

	return &statement.ShowTables{
		Extended:    extended,
		Full:        full,
		ShowOptions: options,
	}, nil
}

// parseShowStatus parses SHOW STATUS statements
func parseShowStatus(p *Parser, global, session bool) (ast.Statement, error) {
	p.AdvanceToken()
	filter := parseShowFilter(p)
	return &statement.ShowStatus{
		Session: session,
		Global:  global,
		Filter:  filter,
	}, nil
}

// parseShowVariables parses SHOW VARIABLES statements
func parseShowVariables(p *Parser, global, session bool) (ast.Statement, error) {
	p.AdvanceToken()
	filter := parseShowFilter(p)
	return &statement.ShowVariables{
		Session: session,
		Global:  global,
		Filter:  filter,
	}, nil
}

// parseShowCreate parses SHOW CREATE statements
func parseShowCreate(p *Parser) (ast.Statement, error) {
	p.AdvanceToken()

	var objType expr.ShowCreateObject
	switch {
	case p.PeekKeyword("TABLE"):
		objType = expr.ShowCreateObjectTable
		p.AdvanceToken()
	case p.PeekKeyword("TRIGGER"):
		objType = expr.ShowCreateObjectTrigger
		p.AdvanceToken()
	case p.PeekKeyword("EVENT"):
		objType = expr.ShowCreateObjectEvent
		p.AdvanceToken()
	case p.PeekKeyword("FUNCTION"):
		objType = expr.ShowCreateObjectFunction
		p.AdvanceToken()
	case p.PeekKeyword("PROCEDURE"):
		objType = expr.ShowCreateObjectProcedure
		p.AdvanceToken()
	case p.PeekKeyword("VIEW"):
		objType = expr.ShowCreateObjectView
		p.AdvanceToken()
	default:
		return nil, p.ExpectedRef("TABLE, TRIGGER, EVENT, FUNCTION, PROCEDURE, or VIEW after SHOW CREATE", p.PeekTokenRef())
	}

	objName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	return &statement.ShowCreate{
		ObjType: objType,
		ObjName: objName,
	}, nil
}

// parseShowCollation parses SHOW COLLATION statements
func parseShowCollation(p *Parser) (ast.Statement, error) {
	p.AdvanceToken()
	filter := parseShowFilter(p)
	return &statement.ShowCollation{Filter: filter}, nil
}

// parseShowCharset parses SHOW CHARSET/CHARACTER SET statements
func parseShowCharset(p *Parser) (ast.Statement, error) {
	useCharacterSet := false
	if p.PeekKeyword("CHARSET") {
		p.AdvanceToken()
	} else if p.PeekKeyword("CHARACTER") {
		p.AdvanceToken()
		useCharacterSet = true
		if !p.ParseKeyword("SET") {
			return nil, p.ExpectedRef("SET after CHARACTER", p.PeekTokenRef())
		}
	}
	filter := parseShowFilter(p)
	return &statement.ShowCharset{
		Filter:          filter,
		UseCharacterSet: useCharacterSet,
	}, nil
}

// parseShowFilter parses an optional LIKE or WHERE clause for SHOW statements
func parseShowFilter(p *Parser) *expr.ShowStatementFilter {
	if p.PeekKeyword("LIKE") {
		p.AdvanceToken()
		likePattern, err := p.ParseLikePattern()
		if err != nil {
			return nil
		}
		return &expr.ShowStatementFilter{Like: &likePattern}
	}

	if p.PeekKeyword("WHERE") {
		p.AdvanceToken()
		exprParser := NewExpressionParser(p)
		whereExpr, err := exprParser.ParseExpr()
		if err != nil {
			return nil
		}
		return &expr.ShowStatementFilter{Where: whereExpr}
	}

	return nil
}
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
	if _, isAtSign := tok.(tokenizer.TokenAtSign); isAtSign {
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
		if _, err := p.ExpectToken(tokenizer.TokenEq{}); err != nil {
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

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	return setStmt, nil
}

// parseBegin parses BEGIN statements
func parseBegin(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("BEGIN statement parsing not yet fully implemented")
}

// parseCommit parses COMMIT statements
func parseCommit(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("COMMIT statement parsing not yet fully implemented")
}

// parseRollback parses ROLLBACK statements
func parseRollback(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ROLLBACK statement parsing not yet fully implemented")
}

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

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	return actions, nil
}

// parseGrantPermission parses a single privilege action
// Reference: src/parser/mod.rs parse_grant_permission (line 17021)
func parseGrantPermission(p *Parser) (*statement.Action, error) {
	tok := p.PeekToken()
	word, ok := tok.Token.(tokenizer.TokenWord)
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
	}

	// Check for column list: SELECT(col1, col2)
	if _, ok := p.PeekToken().Token.(tokenizer.TokenLParen); ok {
		p.NextToken() // consume (
		columns, err := parseCommaSeparatedIdents(p)
		if err != nil {
			return nil, err
		}
		action.Columns = columns
		if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
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

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
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
	if p.ConsumeToken(tokenizer.TokenAtSign{}) {
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

// parseCopy parses COPY statements
func parseCopy(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("COPY statement parsing not yet fully implemented")
}

// parseExecute parses EXECUTE statements
func parseExecute(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("EXECUTE statement parsing not yet fully implemented")
}

// parsePrepare parses PREPARE statements
func parsePrepare(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("PREPARE statement parsing not yet fully implemented")
}

// parseDeallocate parses DEALLOCATE statements
func parseDeallocate(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("DEALLOCATE statement parsing not yet fully implemented")
}

// parseDeclare parses DECLARE statements
func parseDeclare(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("DECLARE statement parsing not yet fully implemented")
}

// parseFetchStmt parses FETCH statements
func ParseFetch(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("FETCH statement parsing not yet fully implemented")
}

// parseClose parses CLOSE statements
func parseClose(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("CLOSE statement parsing not yet fully implemented")
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

// parseExplain parses EXPLAIN statements
func parseExplain(p *Parser) (ast.Statement, error) {
	return parseExplainWithAlias(p, expr.DescribeAliasExplain)
}

// parseExplainWithAlias parses EXPLAIN/DESCRIBE/DESC statements
func parseExplainWithAlias(p *Parser, describeAlias expr.DescribeAlias) (ast.Statement, error) {
	// Check for utility options in parentheses (PostgreSQL style)
	// EXPLAIN (ANALYZE, VERBOSE) SELECT ...
	var options []*expr.UtilityOption
	if p.GetDialect().SupportsExplainWithUtilityOptions() {
		if _, ok := p.PeekToken().Token.(tokenizer.TokenLParen); ok {
			opts, err := parseUtilityOptions(p)
			if err != nil {
				return nil, err
			}
			options = opts
		}
	}

	// Check for QUERY PLAN syntax
	if p.ParseKeyword("QUERY") {
		// EXPLAIN QUERY PLAN
		if !p.ParseKeyword("PLAN") {
			return nil, fmt.Errorf("expected PLAN after QUERY, found: %v", p.PeekToken())
		}
		// Parse the statement after QUERY PLAN
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		return &statement.Explain{
			DescribeAlias: describeAlias,
			Analyze:       false,
			Verbose:       false,
			QueryPlan:     true,
			Estimate:      false,
			Statement:     stmt,
			Options:       options,
		}, nil
	}

	// Parse optional flags
	analyze := p.ParseKeyword("ANALYZE")
	verbose := p.ParseKeyword("VERBOSE")

	// Check for ESTIMATE (Snowflake)
	estimate := p.ParseKeyword("ESTIMATE")

	// Check if next token looks like a statement keyword
	if p.PeekKeyword("SELECT") || p.PeekKeyword("INSERT") || p.PeekKeyword("UPDATE") || p.PeekKeyword("DELETE") {
		// Try to parse as a statement
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		return &statement.Explain{
			DescribeAlias: describeAlias,
			Analyze:       analyze,
			Verbose:       verbose,
			QueryPlan:     false,
			Estimate:      estimate,
			Statement:     stmt,
			Options:       options,
		}, nil
	}

	// Parse as table name (DESCRIBE table_name)
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	return &statement.ExplainTable{
		DescribeAlias:   describeAlias,
		TableName:       tableName,
		HasTableKeyword: false,
	}, nil
}

// parseUtilityOptions parses utility options in the form of `(option1, option2 arg2, option3 arg3, ...)`
// Reference: src/parser/mod.rs parse_utility_options
func parseUtilityOptions(p *Parser) ([]*expr.UtilityOption, error) {
	// Expect opening parenthesis
	if _, ok := p.PeekToken().Token.(tokenizer.TokenLParen); !ok {
		return nil, fmt.Errorf("expected opening parenthesis, found: %v", p.PeekToken())
	}
	p.NextToken() // consume LParen

	var options []*expr.UtilityOption

	// Handle empty parentheses
	if _, ok := p.PeekToken().Token.(tokenizer.TokenRParen); ok {
		p.NextToken() // consume RParen
		return options, nil
	}

	for {
		// Parse option name
		_, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected identifier, found: %v", p.PeekToken())
		}

		// Check if there's an argument
		nextToken := p.PeekToken()
		if _, isComma := nextToken.Token.(tokenizer.TokenComma); isComma {
			// No argument, just the option name
			options = append(options, &expr.UtilityOption{})
		} else if _, isRParen := nextToken.Token.(tokenizer.TokenRParen); isRParen {
			// No argument, just the option name
			options = append(options, &expr.UtilityOption{})
		} else {
			// Has argument - just consume tokens until comma or closing paren
			// This is a simplified version - in practice we'd parse an expression
			for {
				next := p.PeekToken()
				if _, isComma := next.Token.(tokenizer.TokenComma); isComma {
					break
				}
				if _, isRParen := next.Token.(tokenizer.TokenRParen); isRParen {
					break
				}
				p.NextToken()
			}
			options = append(options, &expr.UtilityOption{})
		}

		// Check for comma or closing parenthesis
		nextToken = p.PeekToken()
		if _, isRParen := nextToken.Token.(tokenizer.TokenRParen); isRParen {
			p.NextToken() // consume RParen
			break
		} else if _, isComma := nextToken.Token.(tokenizer.TokenComma); isComma {
			p.NextToken() // consume Comma
		} else {
			return nil, fmt.Errorf("expected comma or closing parenthesis, found: %v", nextToken)
		}
	}

	return options, nil
}

// ParseExplain parses EXPLAIN statements
func ParseExplain(p *Parser) (ast.Statement, error) {
	return parseExplain(p)
}

// ParseDescribe parses DESCRIBE statements
func ParseDescribe(p *Parser) (ast.Statement, error) {
	return ParseDescribeWithAlias(p, "DESCRIBE")
}

// ParseDescribeWithAlias parses DESCRIBE/DESC statements with the given alias
func ParseDescribeWithAlias(p *Parser, keyword string) (ast.Statement, error) {
	// Determine the alias based on the keyword used
	var alias expr.DescribeAlias
	switch keyword {
	case "DESC":
		alias = expr.DescribeAliasDesc
	case "DESCRIBE":
		alias = expr.DescribeAliasDescribe
	default:
		alias = expr.DescribeAliasDescribe
	}

	// Check if this looks like a statement (SELECT, INSERT, UPDATE, DELETE)
	// DESCRIBE is an alias for EXPLAIN when followed by a statement
	if p.PeekKeyword("SELECT") || p.PeekKeyword("INSERT") || p.PeekKeyword("UPDATE") || p.PeekKeyword("DELETE") {
		return parseExplainWithAlias(p, alias)
	}

	// Check for utility options in parentheses (PostgreSQL style)
	if p.GetDialect().SupportsExplainWithUtilityOptions() {
		if _, ok := p.PeekToken().Token.(tokenizer.TokenLParen); ok {
			return parseExplainWithAlias(p, alias)
		}
	}

	// Check for TABLE keyword (optional in some dialects)
	hasTableKeyword := p.ParseKeyword("TABLE")

	// Check for optional Hive format (EXTENDED or FORMATTED)
	var hiveFormat *expr.HiveDescribeFormat
	if p.ParseKeyword("EXTENDED") {
		format := expr.HiveDescribeFormatExtended
		hiveFormat = &format
	} else if p.ParseKeyword("FORMATTED") {
		format := expr.HiveDescribeFormatFormatted
		hiveFormat = &format
	}

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	return &statement.ExplainTable{
		DescribeAlias:   alias,
		TableName:       tableName,
		HasTableKeyword: hasTableKeyword,
		HiveFormat:      hiveFormat,
	}, nil
}

// ParseShow parses SHOW statements
func ParseShow(p *Parser) (ast.Statement, error) {
	return parseShow(p)
}

// ParseSet parses SET statements
func ParseSet(p *Parser) (ast.Statement, error) {
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
	if _, isAtSign := tok.(tokenizer.TokenAtSign); isAtSign {
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
		if _, err := p.ExpectToken(tokenizer.TokenEq{}); err != nil {
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

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	return setStmt, nil
}

// ParseGrant parses GRANT statements
func ParseGrant(p *Parser) (ast.Statement, error) {
	return parseGrant(p)
}

// ParseRevoke parses REVOKE statements
func ParseRevoke(p *Parser) (ast.Statement, error) {
	return parseRevoke(p)
}

// ParseUse parses USE statements
func ParseUse(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("USE statement parsing not yet fully implemented")
}

// ParseAnalyze parses ANALYZE statements
func ParseAnalyze(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ANALYZE statement parsing not yet fully implemented")
}

// ParseCall parses CALL statements
func ParseCall(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("CALL statement parsing not yet fully implemented")
}

// ParseCopy parses COPY statements
// Reference: src/parser/mod.rs parse_copy
func ParseCopy(p *Parser) (ast.Statement, error) {
	copyStmt := &statement.Copy{}

	// Parse source: either (query) or table_name [ (columns) ]
	if p.ConsumeToken(tokenizer.TokenLParen{}) {
		// Parse query as source
		query, err := p.ParseQuery()
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
			return nil, err
		}
		copyStmt.Source = &expr.CopySource{
			Query: query,
		}
	} else {
		// Parse table name
		tableName, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}

		// Parse optional column list
		var columns []*ast.Ident
		if _, ok := p.PeekToken().Token.(tokenizer.TokenLParen); ok {
			p.AdvanceToken() // consume (
			for {
				if _, ok := p.PeekToken().Token.(tokenizer.TokenRParen); ok {
					p.AdvanceToken() // consume )
					break
				}
				col, err := p.ParseIdentifier()
				if err != nil {
					return nil, err
				}
				columns = append(columns, col)
				if !p.ConsumeToken(tokenizer.TokenComma{}) {
					if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
						return nil, err
					}
					break
				}
			}
		}

		copyStmt.Source = &expr.CopySource{
			TableName: tableName,
			Columns:   columns,
		}
	}

	// Parse FROM or TO
	direction := p.ParseOneOfKeywords([]string{"FROM", "TO"})
	if direction == "" {
		return nil, p.expectedRef("FROM or TO", p.PeekTokenRef())
	}
	copyStmt.To = direction == "TO"

	// Parse target: STDIN, STDOUT, PROGRAM 'cmd', or 'filename'
	switch {
	case p.ParseKeyword("STDIN"):
		copyStmt.Target = &expr.CopyTarget{
			Kind: expr.CopyTargetKindStdin,
		}
	case p.ParseKeyword("STDOUT"):
		copyStmt.Target = &expr.CopyTarget{
			Kind: expr.CopyTargetKindStdout,
		}
	case p.ParseKeyword("PROGRAM"):
		cmd, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		copyStmt.Target = &expr.CopyTarget{
			Kind:    expr.CopyTargetKindProgram,
			Command: cmd,
		}
	default:
		// Must be a filename string
		filename, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		copyStmt.Target = &expr.CopyTarget{
			Kind:     expr.CopyTargetKindFile,
			Filename: filename,
		}
	}

	// Parse optional WITH (ignored for compatibility)
	p.ParseKeyword("WITH")

	// Parse options: (option, option, ...)
	if p.ConsumeToken(tokenizer.TokenLParen{}) {
		for {
			if _, ok := p.PeekToken().Token.(tokenizer.TokenRParen); ok {
				p.AdvanceToken() // consume )
				break
			}

			opt, err := parseCopyOption(p)
			if err != nil {
				return nil, err
			}
			copyStmt.Options = append(copyStmt.Options, opt)

			if !p.ConsumeToken(tokenizer.TokenComma{}) {
				if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
					return nil, err
				}
				break
			}
		}
	}

	// Parse legacy options (space-separated keywords with optional values)
	for {
		opt, err := tryParseCopyLegacyOption(p)
		if err != nil || opt == nil {
			break
		}
		copyStmt.LegacyOptions = append(copyStmt.LegacyOptions, opt)
	}

	// TODO: Parse TSV values if target is STDIN

	return copyStmt, nil
}

// parseCopyOption parses a single COPY option (PostgreSQL 9.0+ format)
func parseCopyOption(p *Parser) (*expr.CopyOption, error) {
	kw := p.ParseOneOfKeywords([]string{
		"FORMAT", "FREEZE", "DELIMITER", "NULL", "HEADER",
		"QUOTE", "ESCAPE", "FORCE_QUOTE", "FORCE_NOT_NULL", "FORCE_NULL", "ENCODING",
	})

	switch kw {
	case "FORMAT":
		ident, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionFormat,
			Value:      ident,
		}, nil
	case "FREEZE":
		val := true
		if p.ParseKeyword("FALSE") {
			val = false
		} else if p.ParseKeyword("TRUE") {
			val = true
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionFreeze,
			Value:      val,
		}, nil
	case "DELIMITER":
		ch, err := p.ParseLiteralChar()
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionDelimiter,
			Value:      string(ch),
		}, nil
	case "NULL":
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionNull,
			Value:      val,
		}, nil
	case "HEADER":
		val := true
		if p.ParseKeyword("FALSE") {
			val = false
		} else if p.ParseKeyword("TRUE") {
			val = true
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionHeader,
			Value:      val,
		}, nil
	case "QUOTE":
		ch, err := p.ParseLiteralChar()
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionQuote,
			Value:      string(ch),
		}, nil
	case "ESCAPE":
		ch, err := p.ParseLiteralChar()
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionEscape,
			Value:      string(ch),
		}, nil
	case "FORCE_QUOTE":
		cols, err := parseCopyColumnList(p)
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionForceQuote,
			Value:      cols,
		}, nil
	case "FORCE_NOT_NULL":
		cols, err := parseCopyColumnList(p)
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionForceNotNull,
			Value:      cols,
		}, nil
	case "FORCE_NULL":
		cols, err := parseCopyColumnList(p)
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionForceNull,
			Value:      cols,
		}, nil
	case "ENCODING":
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionEncoding,
			Value:      val,
		}, nil
	}

	return nil, p.expectedRef("COPY option", p.PeekTokenRef())
}

// parseCopyColumnList parses a parenthesized column list for FORCE_QUOTE, etc.
func parseCopyColumnList(p *Parser) ([]*ast.Ident, error) {
	if _, err := p.ExpectToken(tokenizer.TokenLParen{}); err != nil {
		return nil, err
	}

	var cols []*ast.Ident
	for {
		if _, ok := p.PeekToken().Token.(tokenizer.TokenRParen); ok {
			p.AdvanceToken() // consume )
			break
		}
		col, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		cols = append(cols, col)
		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
				return nil, err
			}
			break
		}
	}

	return cols, nil
}

// tryParseCopyLegacyOption tries to parse a legacy COPY option. Returns nil if no option found.
func tryParseCopyLegacyOption(p *Parser) (*expr.CopyLegacyOption, error) {
	// Check for FORMAT [ AS ] (handled specially at the beginning)
	if p.ParseKeyword("FORMAT") {
		p.ParseKeyword("AS")
	}

	kw := p.ParseOneOfKeywords([]string{
		"BINARY", "CSV", "DELIMITER", "ESCAPE", "HEADER", "JSON",
		"NULL", "PARQUET", "GZIP", "BZIP2", "ZSTD", "EMPTYASNULL", "BLANKSASNULL",
		"REMOVEQUOTES", "ADDQUOTES", "IGNOREHEADER", "DATEFORMAT", "TIMEFORMAT",
		"TRUNCATECOLUMNS", "COMPUPDATE", "STATUPDATE", "PARALLEL", "MAXFILESIZE",
		"REGION", "IAM_ROLE", "MANIFEST", "CREDENTIALS", "FIXEDWIDTH", "EXTENSION",
		"ACCEPTANYDATE", "ACCEPTINVCHARS", "ALLOWOVERWRITE", "CLEANPATH", "ENCRYPTED",
		"ROWGROUPSIZE", "PARTITION", "PARTITIONBY",
	})

	switch kw {
	case "BINARY":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionBinary}, nil
	case "CSV":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionCsv}, nil
	case "DELIMITER":
		p.ParseKeyword("AS")
		ch, err := p.ParseLiteralChar()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionDelimiter,
			Value:      string(ch),
		}, nil
	case "ESCAPE":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionEscape}, nil
	case "HEADER":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionHeader}, nil
	case "JSON":
		p.ParseKeyword("AS")
		if tok, ok := p.PeekToken().Token.(tokenizer.TokenSingleQuotedString); ok {
			p.AdvanceToken()
			return &expr.CopyLegacyOption{
				OptionType: expr.CopyLegacyOptionJson,
				Value:      tok.Value,
			}, nil
		}
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionJson}, nil
	case "NULL":
		p.ParseKeyword("AS")
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionNull,
			Value:      val,
		}, nil
	case "PARQUET":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionParquet}, nil
	case "GZIP":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionGzip}, nil
	case "BZIP2":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionBzip2}, nil
	case "ZSTD":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionZstd}, nil
	case "EMPTYASNULL":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionEmptyAsNull}, nil
	case "BLANKSASNULL":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionBlankAsNull}, nil
	case "REMOVEQUOTES":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionRemoveQuotes}, nil
	case "ADDQUOTES":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionAddQuotes}, nil
	case "IGNOREHEADER":
		p.ParseKeyword("AS")
		num, err := p.ParseNumber()
		if err != nil {
			return nil, err
		}
		// Convert to int
		val := 0
		fmt.Sscanf(num, "%d", &val)
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionIgnoreHeader,
			Value:      val,
		}, nil
	case "DATEFORMAT":
		p.ParseKeyword("AS")
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionDateFormat,
			Value:      val,
		}, nil
	case "TIMEFORMAT":
		p.ParseKeyword("AS")
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionTimeFormat,
			Value:      val,
		}, nil
	case "TRUNCATECOLUMNS":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionTruncateColumns}, nil
	case "COMPUPDATE":
		val := ""
		if p.ParseKeyword("PRESET") {
			val = "PRESET"
		} else if p.ParseKeyword("ON") || p.ParseKeyword("TRUE") {
			val = "TRUE"
		} else if p.ParseKeyword("OFF") || p.ParseKeyword("FALSE") {
			val = "FALSE"
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionCompUpdate,
			Value:      val,
		}, nil
	case "STATUPDATE":
		val := ""
		if p.ParseKeyword("ON") || p.ParseKeyword("TRUE") {
			val = "TRUE"
		} else if p.ParseKeyword("OFF") || p.ParseKeyword("FALSE") {
			val = "FALSE"
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionStatUpdate,
			Value:      val,
		}, nil
	case "PARALLEL":
		val := ""
		if p.ParseKeyword("ON") || p.ParseKeyword("TRUE") {
			val = "TRUE"
		} else if p.ParseKeyword("OFF") || p.ParseKeyword("FALSE") {
			val = "FALSE"
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionParallel,
			Value:      val,
		}, nil
	case "MAXFILESIZE":
		p.ParseKeyword("AS")
		val, err := p.ParseNumber()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionMaxFileSize,
			Value:      val,
		}, nil
	case "REGION":
		p.ParseKeyword("AS")
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionRegion,
			Value:      val,
		}, nil
	case "MANIFEST":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionManifest}, nil
	case "CREDENTIALS":
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionCredentials,
			Value:      val,
		}, nil
	case "FIXEDWIDTH":
		p.ParseKeyword("AS")
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionFixedWidth,
			Value:      val,
		}, nil
	case "EXTENSION":
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionExtension,
			Value:      val,
		}, nil
	case "ACCEPTANYDATE":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionAcceptAnyDate}, nil
	case "ACCEPTINVCHARS":
		p.ParseKeyword("AS")
		if tok, ok := p.PeekToken().Token.(tokenizer.TokenSingleQuotedString); ok {
			p.AdvanceToken()
			return &expr.CopyLegacyOption{
				OptionType: expr.CopyLegacyOptionAcceptInvChars,
				Value:      tok.Value,
			}, nil
		}
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionAcceptInvChars}, nil
	case "ALLOWOVERWRITE":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionAllowOverwrite}, nil
	case "CLEANPATH":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionCleanPath}, nil
	case "ENCRYPTED":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionEncrypted}, nil
	case "ROWGROUPSIZE":
		p.ParseKeyword("AS")
		val, err := p.ParseNumber()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionRowGroupSize,
			Value:      val,
		}, nil
	}

	return nil, nil // No legacy option found
}

// ParseDeclare parses DECLARE statements
func ParseDeclare(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("DECLARE statement parsing not yet fully implemented")
}

// ParseClose parses CLOSE statements
func ParseClose(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("CLOSE statement parsing not yet fully implemented")
}
