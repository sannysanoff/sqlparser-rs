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

// parseShow parses SHOW statements
// Reference: src/parser/mod.rs:15030
func parseShow(p *Parser) (ast.Statement, error) {
	// Parse optional modifiers that can appear before the object type
	var terse, full, extended, global, session, external bool

	// Try to parse optional modifiers
	for {
		switch {
		case !terse && p.PeekKeyword("TERSE"):
			terse = p.ParseKeyword("TERSE")
		case !full && p.PeekKeyword("FULL"):
			full = p.ParseKeyword("FULL")
		case !extended && p.PeekKeyword("EXTENDED"):
			extended = p.ParseKeyword("EXTENDED")
		case !global && p.PeekKeyword("GLOBAL"):
			global = p.ParseKeyword("GLOBAL")
		case !session && p.PeekKeyword("SESSION"):
			session = p.ParseKeyword("SESSION")
		case !external && p.PeekKeyword("EXTERNAL"):
			external = p.ParseKeyword("EXTERNAL")
		default:
			goto doneModifiers
		}
	}
doneModifiers:

	// Check for CREATE (e.g., SHOW CREATE TABLE)
	// EXTENDED and FULL are not valid with CREATE
	if p.PeekKeyword("CREATE") {
		if extended || full {
			return nil, fmt.Errorf("EXTENDED/FULL are not supported with SHOW CREATE")
		}
		return parseShowCreate(p)
	}

	// Check for COLUMNS/FIELDS
	if p.PeekKeyword("COLUMNS") || p.PeekKeyword("FIELDS") {
		return parseShowColumns(p, extended, full)
	}

	// Check for TABLES
	if p.PeekKeyword("TABLES") {
		return parseShowTables(p, terse, extended, full, external)
	}

	// Check for MATERIALIZED VIEWS
	if p.PeekKeyword("MATERIALIZED") {
		p.AdvanceToken()
		if !p.PeekKeyword("VIEWS") {
			return nil, p.ExpectedRef("VIEWS after MATERIALIZED", p.PeekTokenRef())
		}
		return parseShowViews(p, terse, true)
	}

	// Check for VIEWS
	if p.PeekKeyword("VIEWS") {
		return parseShowViews(p, terse, false)
	}

	// Check for FUNCTIONS
	if p.PeekKeyword("FUNCTIONS") {
		return parseShowFunctions(p)
	}

	// Check for DATABASES
	if p.PeekKeyword("DATABASES") {
		return parseShowDatabases(p, terse)
	}

	// Check for SCHEMAS
	if p.PeekKeyword("SCHEMAS") {
		return parseShowSchemas(p, terse)
	}

	// Check for STATUS
	if p.PeekKeyword("STATUS") {
		if extended || full {
			return nil, fmt.Errorf("EXTENDED/FULL are not supported with SHOW STATUS")
		}
		return parseShowStatus(p, global, session)
	}

	// Check for VARIABLES
	if p.PeekKeyword("VARIABLES") {
		if extended || full {
			return nil, fmt.Errorf("EXTENDED/FULL are not supported with SHOW VARIABLES")
		}
		return parseShowVariables(p, global, session)
	}

	// Check for COLLATION
	if p.PeekKeyword("COLLATION") {
		if extended || full {
			return nil, fmt.Errorf("EXTENDED/FULL are not supported with SHOW COLLATION")
		}
		return parseShowCollation(p)
	}

	// Check for CHARSET/CHARACTER SET
	if p.PeekKeyword("CHARSET") || p.PeekKeyword("CHARACTER") {
		return parseShowCharset(p)
	}

	// Check for OBJECTS (Snowflake)
	if p.PeekKeyword("OBJECTS") {
		return parseShowObjects(p, terse)
	}

	// Extended/full without valid target
	if extended || full {
		return nil, fmt.Errorf("EXTENDED/FULL are not supported with this type of SHOW query")
	}

	// SHOW variable (MySQL style: SHOW ENGINE INNODB STATUS)
	return parseShowVariable(p)
}

// parseShowColumns parses SHOW COLUMNS/FIELDS statements
func parseShowColumns(p *Parser, extended, full bool) (ast.Statement, error) {
	// Consume COLUMNS or FIELDS
	p.AdvanceToken()

	options := &expr.ShowStatementOptions{}

	// Parse optional LIKE/WHERE before FROM/IN (for Snowflake-style: SHOW COLUMNS LIKE 'x' IN TABLE t)
	if p.PeekKeyword("LIKE") || p.PeekKeyword("WHERE") {
		options.Filter = parseShowFilter(p)
		// Mark filter as infix position (before IN/FROM)
		options.FilterPosition = expr.ShowStatementFilterPositionInfix
	}

	// Parse FROM or IN clause
	if p.PeekKeyword("FROM") || p.PeekKeyword("IN") {
		clause := expr.ShowStatementInClauseFrom
		if p.PeekKeyword("IN") {
			clause = expr.ShowStatementInClauseIn
		}
		p.AdvanceToken()

		// Check for special keyword TABLE/VIEW after IN (Snowflake-style: SHOW COLUMNS IN TABLE tbl)
		parentType := ""
		if p.ParseKeyword("TABLE") {
			parentType = "TABLE"
		} else if p.ParseKeyword("VIEW") {
			parentType = "VIEW"
		}

		// Parse table/view name (optional - may not be present)
		tableName, _ := p.ParseObjectName()
		if tableName != nil {
			options.ShowIn = &expr.ShowStatementIn{
				Clause:     clause,
				ParentName: tableName,
			}
			if parentType != "" {
				options.ShowIn.ParentType = &ast.Ident{Value: parentType}
			}
		} else {
			// No table name - just have the clause and parent type
			options.ShowIn = &expr.ShowStatementIn{
				Clause: clause,
			}
			if parentType != "" {
				options.ShowIn.ParentType = &ast.Ident{Value: parentType}
			}
		}

		// MySQL also supports a second FROM clause for the database:
		// SHOW COLUMNS FROM table FROM db -> SHOW COLUMNS FROM db.table
		if p.PeekKeyword("FROM") {
			p.AdvanceToken()
			dbName, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			// Prepend the database name to the table's parts
			newParts := []ast.ObjectNamePart{&ast.ObjectNamePartIdentifier{Ident: dbName}}
			newParts = append(newParts, tableName.Parts...)
			options.ShowIn.ParentName = &ast.ObjectName{Parts: newParts}
		}
	}

	// Parse optional LIKE/WHERE if not already parsed (suffix position - after IN/FROM)
	if options.Filter == nil {
		options.Filter = parseShowFilter(p)
		// FilterPosition stays as default (None/Suffix)
	}

	return &statement.ShowColumns{
		Extended:    extended,
		Full:        full,
		ShowOptions: options,
	}, nil
}

// parseShowTables parses SHOW TABLES statements
// Reference: src/parser/mod.rs:15079
func parseShowTables(p *Parser, terse, extended, full, external bool) (ast.Statement, error) {
	p.AdvanceToken()

	// Parse optional HISTORY keyword (not for EXTERNAL TABLES)
	history := !external && p.ParseKeyword("HISTORY")

	options := parseShowStmtOptions(p)

	return &statement.ShowTables{
		Terse:       terse,
		Extended:    extended,
		Full:        full,
		External:    external,
		History:     history,
		ShowOptions: options,
	}, nil
}

// parseShowStatus parses SHOW STATUS statements
func parseShowStatus(p *Parser, global, session bool) (ast.Statement, error) {
	p.AdvanceToken()

	filter := parseShowFilter(p)

	return &statement.ShowStatus{
		Global:  global,
		Session: session,
		Filter:  filter,
	}, nil
}

// parseShowVariables parses SHOW VARIABLES statements
func parseShowVariables(p *Parser, global, session bool) (ast.Statement, error) {
	p.AdvanceToken()

	filter := parseShowFilter(p)

	return &statement.ShowVariables{
		Global:  global,
		Session: session,
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

// parseShowDatabases parses SHOW DATABASES statements
// Reference: src/parser/mod.rs:15097
func parseShowDatabases(p *Parser, terse bool) (ast.Statement, error) {
	p.AdvanceToken()
	history := p.ParseKeyword("HISTORY")
	options := parseShowStmtOptions(p)
	return &statement.ShowDatabases{
		Terse:       terse,
		History:     history,
		ShowOptions: options,
	}, nil
}

// parseShowSchemas parses SHOW SCHEMAS statements
// Reference: src/parser/mod.rs:15107
func parseShowSchemas(p *Parser, terse bool) (ast.Statement, error) {
	p.AdvanceToken()
	history := p.ParseKeyword("HISTORY")
	options := parseShowStmtOptions(p)
	return &statement.ShowSchemas{
		Terse:       terse,
		History:     history,
		ShowOptions: options,
	}, nil
}

// parseShowViews parses SHOW VIEWS or SHOW MATERIALIZED VIEWS statements
// Reference: src/parser/mod.rs:15176
func parseShowViews(p *Parser, terse bool, materialized bool) (ast.Statement, error) {
	p.AdvanceToken()
	options := parseShowStmtOptions(p)
	return &statement.ShowViews{
		Terse:        terse,
		Materialized: materialized,
		ShowOptions:  options,
	}, nil
}

// parseShowFunctions parses SHOW FUNCTIONS statements
// Reference: src/parser/mod.rs:15190
func parseShowFunctions(p *Parser) (ast.Statement, error) {
	p.AdvanceToken()
	filter := parseShowFilter(p)
	return &statement.ShowFunctions{Filter: filter}, nil
}

// parseShowObjects parses SHOW OBJECTS statements (Snowflake)
func parseShowObjects(p *Parser, terse bool) (ast.Statement, error) {
	p.AdvanceToken()
	options := parseShowStmtOptions(p)
	return &statement.ShowObjects{
		Terse:       terse,
		ShowOptions: options,
	}, nil
}

// parseShowVariable parses SHOW variable statements (MySQL style: SHOW ENGINE INNODB STATUS)
func parseShowVariable(p *Parser) (ast.Statement, error) {
	// Parse one or more identifiers (e.g., "ENGINE INNODB STATUS")
	var identifiers []*ast.Ident
	for {
		ident, err := p.ParseIdentifier()
		if err != nil {
			break
		}
		identifiers = append(identifiers, ident)
	}
	return &statement.ShowVariable{Variable: identifiers}, nil
}

// parseShowStmtOptions parses SHOW statement options (LIKE, WHERE, IN, FROM, LIMIT, STARTS WITH)
// Reference: src/parser/mod.rs:19845
func parseShowStmtOptions(p *Parser) *expr.ShowStatementOptions {
	options := &expr.ShowStatementOptions{}

	// Parse optional LIKE/WHERE (infix position - before IN/FROM for some dialects)
	if p.GetDialect().SupportsShowLikeBeforeIn() {
		if p.PeekKeyword("LIKE") {
			p.AdvanceToken()
			likePattern, err := p.ParseLikePattern()
			if err == nil {
				options.Filter = &expr.ShowStatementFilter{Like: &likePattern}
				options.FilterPosition = expr.ShowStatementFilterPositionInfix
			}
		} else if p.PeekKeyword("WHERE") {
			p.AdvanceToken()
			exprParser := NewExpressionParser(p)
			whereExpr, err := exprParser.ParseExpr()
			if err == nil {
				options.Filter = &expr.ShowStatementFilter{Where: whereExpr}
				options.FilterPosition = expr.ShowStatementFilterPositionInfix
			}
		}
	}

	// Parse optional IN/FROM clause with support for parent types (ACCOUNT, DATABASE, SCHEMA, TABLE, VIEW)
	if p.PeekKeyword("FROM") || p.PeekKeyword("IN") {
		clause := expr.ShowStatementInClauseFrom
		if p.PeekKeyword("IN") {
			clause = expr.ShowStatementInClauseIn
		}
		p.AdvanceToken()

		// Check for special parent type keywords (ACCOUNT, DATABASE, SCHEMA, TABLE, VIEW)
		parentTypeKeywords := []string{"ACCOUNT", "DATABASE", "SCHEMA", "TABLE", "VIEW"}
		parentType := ""
		for _, kw := range parentTypeKeywords {
			if p.ParseKeyword(kw) {
				parentType = kw
				break
			}
		}

		// Parse optional parent name (object name)
		// This handles cases like "SHOW DATABASES IN ACCOUNT myaccount"
		// or "SHOW TABLES IN SCHEMA" (no name)
		var parentName *ast.ObjectName
		if parentType != "" {
			// Try to parse an identifier as the parent name, but it's optional
			if ident, err := p.ParseIdentifier(); err == nil {
				parentName = &ast.ObjectName{Parts: []ast.ObjectNamePart{&ast.ObjectNamePartIdentifier{Ident: ident}}}
			}
		} else {
			// No parent type keyword found, parse identifier as the parent name
			if ident, err := p.ParseIdentifier(); err == nil {
				parentName = &ast.ObjectName{Parts: []ast.ObjectNamePart{&ast.ObjectNamePartIdentifier{Ident: ident}}}
			}
		}

		options.ShowIn = &expr.ShowStatementIn{
			Clause:     clause,
			ParentName: parentName,
		}
		if parentType != "" {
			options.ShowIn.ParentType = &ast.Ident{Value: parentType}
		}
	}

	// Parse optional LIKE/WHERE (suffix position - after IN/FROM for some dialects)
	// Also try parsing a bare string literal for Snowflake-style suffix filter
	if !p.GetDialect().SupportsShowLikeBeforeIn() || options.Filter == nil {
		if p.PeekKeyword("LIKE") {
			p.AdvanceToken()
			likePattern, err := p.ParseLikePattern()
			if err == nil {
				options.Filter = &expr.ShowStatementFilter{Like: &likePattern}
				options.FilterPosition = expr.ShowStatementFilterPositionSuffix
			}
		} else if p.PeekKeyword("WHERE") {
			p.AdvanceToken()
			exprParser := NewExpressionParser(p)
			whereExpr, err := exprParser.ParseExpr()
			if err == nil {
				options.Filter = &expr.ShowStatementFilter{Where: whereExpr}
				options.FilterPosition = expr.ShowStatementFilterPositionSuffix
			}
		} else {
			// Try to parse a bare string literal (Snowflake-style: SHOW TABLES IN db1 'abc')
			tok := p.PeekToken()
			if strTok, ok := tok.Token.(token.TokenSingleQuotedString); ok {
				p.AdvanceToken()
				suffixStr := strTok.Value
				options.Filter = &expr.ShowStatementFilter{SuffixString: &suffixStr}
				options.FilterPosition = expr.ShowStatementFilterPositionSuffix
			}
		}
	}

	// Parse optional STARTS WITH (Snowflake)
	if p.ParseKeyword("STARTS") {
		if p.ParseKeyword("WITH") {
			prefix, err := p.ParseIdentifier()
			if err == nil {
				options.StartsWith = &prefix.Value
			}
		}
	}

	// Parse optional LIMIT (Snowflake)
	if p.ParseKeyword("LIMIT") {
		limitStr, err := p.ParseNumber()
		if err == nil {
			// Convert string to int
			var limit int
			if _, err := fmt.Sscanf(limitStr, "%d", &limit); err == nil {
				options.Limit = &limit
			}
		}
	}

	// Parse optional FROM (Snowflake pagination marker: FROM 'marker')
	// This is different from the IN/FROM clause - it comes after LIMIT
	if p.ParseKeyword("FROM") {
		// Can be either an identifier or a string literal
		tok := p.PeekToken()
		if strTok, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			p.AdvanceToken()
			options.LimitFrom = &strTok.Value
		} else if ident, err := p.ParseIdentifier(); err == nil {
			options.LimitFrom = &ident.Value
		}
	}

	return options
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
