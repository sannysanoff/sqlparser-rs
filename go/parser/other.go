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
		IfExists: ifExists,
		Names:    names,
		Cascade:  cascade,
		Restrict: restrict,
	}, nil
}

// parseCommaSeparatedObjectNames parses a comma-separated list of object names
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
func parseShow(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("SHOW statement parsing not yet fully implemented")
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
func parseGrant(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("GRANT statement parsing not yet fully implemented")
}

// parseRevoke parses REVOKE statements
func parseRevoke(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("REVOKE statement parsing not yet fully implemented")
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
	// Parse optional flags
	analyze := p.ParseKeyword("ANALYZE")
	verbose := p.ParseKeyword("VERBOSE")

	// Check for QUERY PLAN
	queryPlan := false
	if p.ParseKeyword("QUERY") {
		if p.ParseKeyword("PLAN") {
			queryPlan = true
		}
	}

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
			DescribeAlias: expr.DescribeAliasExplain,
			Analyze:       analyze,
			Verbose:       verbose,
			QueryPlan:     queryPlan,
			Estimate:      estimate,
			Statement:     stmt,
		}, nil
	}

	// Parse as table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}
	return &statement.ExplainTable{
		DescribeAlias:   expr.DescribeAliasExplain,
		TableName:       tableName,
		HasTableKeyword: false,
	}, nil
}

// ParseAlter parses ALTER statements
func ParseAlter(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER statement parsing not yet fully implemented")
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

	return &statement.ExplainTable{
		DescribeAlias:   alias,
		TableName:       tableName,
		HasTableKeyword: hasTableKeyword,
		HiveFormat:      hiveFormat,
	}, nil
}

// ParseShow parses SHOW statements
func ParseShow(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("SHOW statement parsing not yet fully implemented")
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
	return nil, fmt.Errorf("GRANT statement parsing not yet fully implemented")
}

// ParseRevoke parses REVOKE statements
func ParseRevoke(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("REVOKE statement parsing not yet fully implemented")
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
func ParseCopy(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("COPY statement parsing not yet fully implemented")
}

// ParseDeclare parses DECLARE statements
func ParseDeclare(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("DECLARE statement parsing not yet fully implemented")
}

// ParseClose parses CLOSE statements
func ParseClose(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("CLOSE statement parsing not yet fully implemented")
}
