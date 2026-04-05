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

// parseCreate parses a CREATE statement
// Reference: src/parser/mod.rs:5095
func parseCreate(p *Parser) (ast.Statement, error) {
	// Check for OR REPLACE / OR ALTER
	orReplace := p.ParseKeywords([]string{"OR", "REPLACE"})
	_ = p.ParseKeywords([]string{"OR", "ALTER"}) // orAlter not used yet

	// Check for LOCAL/GLOBAL
	local := p.ParseKeyword("LOCAL")
	global := p.ParseKeyword("GLOBAL")
	var globalOpt *bool
	if global {
		globalOpt = &[]bool{true}[0]
	} else if local {
		globalOpt = &[]bool{false}[0]
	}

	// Check for TRANSIENT
	transient := p.ParseKeyword("TRANSIENT")

	// Check for TEMPORARY/TEMP
	temporary := p.ParseKeyword("TEMPORARY") || p.ParseKeyword("TEMP")

	// Check for PERSISTENT (DuckDB)
	// Note: Persistent is not stored separately, it's just a modifier
	p.ParseKeyword("PERSISTENT")

	// Try various CREATE targets
	switch {
	case p.PeekKeyword("TABLE"):
		return parseCreateTable(p, orReplace, temporary, globalOpt, transient)
	case p.PeekKeyword("VIEW"):
		return parseCreateView(p, orReplace, temporary)
	case p.PeekKeyword("INDEX"):
		return parseCreateIndex(p, false)
	case p.PeekKeyword("UNIQUE"):
		p.NextToken()
		return parseCreateIndex(p, true)
	case p.PeekKeyword("ROLE"):
		return parseCreateRole(p, orReplace)
	case p.PeekKeyword("DATABASE"):
		return parseCreateDatabase(p)
	case p.PeekKeyword("SCHEMA"):
		return parseCreateSchema(p)
	case p.PeekKeyword("SEQUENCE"):
		return parseCreateSequence(p, orReplace, temporary)
	case p.PeekKeyword("TYPE"):
		return parseCreateType(p)
	case p.PeekKeyword("DOMAIN"):
		return parseCreateDomain(p)
	case p.PeekKeyword("EXTENSION"):
		return parseCreateExtension(p)
	case p.PeekKeyword("TRIGGER"):
		return parseCreateTrigger(p, orReplace)
	case p.PeekKeyword("POLICY"):
		return parseCreatePolicy(p, orReplace)
	case p.PeekKeyword("FUNCTION"):
		return parseCreateFunction(p, orReplace, temporary)
	case p.PeekKeyword("VIRTUAL"):
		p.NextToken()
		return parseCreateVirtualTable(p)
	case p.PeekKeyword("MACRO"):
		return parseCreateMacro(p)
	case p.PeekKeyword("SECRET"):
		return parseCreateSecret(p, orReplace, temporary)
	case p.PeekKeyword("CONNECTOR"):
		return parseCreateConnector(p, orReplace)
	case p.PeekKeyword("OPERATOR"):
		return parseCreateOperator(p)
	case p.PeekKeyword("USER"):
		return parseCreateUser(p, orReplace)
	default:
		return nil, p.ExpectedRef("TABLE, VIEW, INDEX, FUNCTION, ROLE, or other CREATE target", p.PeekTokenRef())
	}
}

// parseCreateTable parses CREATE TABLE
// Reference: src/parser/mod.rs:8339
func parseCreateTable(p *Parser, orReplace, temporary bool, global *bool, transient bool) (ast.Statement, error) {
	// Consume TABLE keyword
	if _, err := p.ExpectKeyword("TABLE"); err != nil {
		return nil, err
	}

	// Check for IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse optional column list and constraints
	var columns []*expr.ColumnDef
	var constraints []*expr.TableConstraint

	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		cols, cons, err := parseCreateTableColumns(p)
		if err != nil {
			return nil, err
		}
		columns = cols
		constraints = cons
	}

	// Check for AS (CREATE TABLE ... AS SELECT)
	var asQuery *query.Query
	if p.PeekKeyword("AS") {
		p.AdvanceToken()
		innerQuery, err := p.ParseQuery()
		if err != nil {
			return nil, err
		}
		// Extract query from the returned statement
		switch q := innerQuery.(type) {
		case *QueryStatement:
			asQuery = q.Query
		case *SelectStatement:
			// Convert SelectStatement to query.Query
			asQuery = &query.Query{
				Body: &query.SelectSetExpr{
					Select: &q.Select,
				},
			}
		case *ValuesStatement:
			asQuery = q.Query
		}
	}

	// Check for LIKE (CREATE TABLE ... LIKE)
	var like *expr.CreateTableLikeKind
	if p.PeekKeyword("LIKE") {
		p.AdvanceToken()
		likeTable, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		like = &expr.CreateTableLikeKind{
			Name: likeTable,
			Kind: expr.CreateTableLikePlain,
		}
	}

	return &statement.CreateTable{
		OrReplace:   orReplace,
		Temporary:   temporary,
		Global:      global,
		Transient:   transient,
		IfNotExists: ifNotExists,
		Name:        tableName,
		Columns:     columns,
		Constraints: constraints,
		Query:       asQuery,
		Like:        like,
	}, nil
}

// parseCreateTableColumns parses the parenthesized column list in CREATE TABLE
// Format: (col_def [, col_def ...] [, table_constraint ...])
func parseCreateTableColumns(p *Parser) ([]*expr.ColumnDef, []*expr.TableConstraint, error) {
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, nil, err
	}

	var columns []*expr.ColumnDef
	var constraints []*expr.TableConstraint

	for {
		// Check for end of list
		if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
			p.NextToken() // consume )
			break
		}

		// Check if this is a table constraint (starts with CONSTRAINT or a constraint keyword)
		if isTableConstraint(p) {
			constraint, err := parseTableConstraint(p)
			if err != nil {
				return nil, nil, err
			}
			constraints = append(constraints, constraint)
		} else {
			// Parse column definition
			col, err := parseColumnDef(p)
			if err != nil {
				return nil, nil, err
			}
			columns = append(columns, col)
		}

		// Check for comma
		if !p.ConsumeToken(token.TokenComma{}) {
			// No comma, expect end of list
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, nil, err
			}
			break
		}

		// Handle trailing comma (DuckDB style)
		if p.GetDialect().SupportsTrailingCommas() {
			if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
				p.NextToken() // consume )
				break
			}
		}
	}

	return columns, constraints, nil
}

// isTableConstraint checks if the next token indicates a table constraint
func isTableConstraint(p *Parser) bool {
	// Check for CONSTRAINT keyword
	if p.PeekKeyword("CONSTRAINT") {
		return true
	}

	// Check for table constraint keywords
	tableConstraintKeywords := []string{
		"PRIMARY", "FOREIGN", "UNIQUE", "CHECK",
	}

	for _, kw := range tableConstraintKeywords {
		if p.PeekKeyword(kw) {
			return true
		}
	}

	// MySQL-specific: INDEX/KEY/FULLTEXT/SPATIAL inline index constraints
	// Reference: src/parser/mod.rs:9732-9760
	if p.GetDialect().SupportsIndexHints() {
		mysqlConstraintKeywords := []string{"INDEX", "KEY", "FULLTEXT", "SPATIAL"}
		for _, kw := range mysqlConstraintKeywords {
			if p.PeekKeyword(kw) {
				return true
			}
		}
	}

	return false
}

// parseCreateView parses CREATE VIEW
// Reference: src/parser/mod.rs parse_create_view
func parseCreateView(p *Parser, orReplace, temporary bool) (ast.Statement, error) {
	// VIEW keyword is expected (already checked by caller)
	if _, err := p.ExpectKeyword("VIEW"); err != nil {
		return nil, err
	}

	// Check for IF NOT EXISTS (before name)
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse view name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Check for IF NOT EXISTS after name (Snowflake style)
	if !ifNotExists {
		ifNotExists = p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})
	}

	// Parse optional column list: (col1, col2, ...)
	var columns []*ast.Ident
	if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
		columns, err = p.ParseParenthesizedColumnList()
		if err != nil {
			return nil, err
		}
	}

	// Expect AS keyword
	if _, err := p.ExpectKeyword("AS"); err != nil {
		return nil, err
	}

	// Parse the query
	stmt, err := p.ParseQuery()
	if err != nil {
		return nil, err
	}

	// Convert statement to *query.Query
	// The parser returns SelectStatement or QueryStatement (for CTE WITH clauses)
	// We need to wrap it in a query.Query for CreateView
	var q *query.Query
	switch s := stmt.(type) {
	case *SelectStatement:
		// Create a copy of the Select to get a pointer
		selectCopy := s.Select
		q = &query.Query{
			Body: &query.SelectSetExpr{Select: &selectCopy},
		}
	case *QueryStatement:
		// CTE query (WITH clause) - use the Query directly
		q = s.Query
	default:
		return nil, fmt.Errorf("expected SELECT query in CREATE VIEW, got %T", stmt)
	}

	return &statement.CreateView{
		OrReplace:   orReplace,
		Temporary:   temporary,
		IfNotExists: ifNotExists,
		Name:        name,
		Columns:     columns,
		Query:       q,
	}, nil
}

// parseCreateIndex parses CREATE INDEX
// Reference: src/parser/mod.rs parse_create_index
func parseCreateIndex(p *Parser, unique bool) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("INDEX"); err != nil {
		return nil, err
	}

	// Parse CONCURRENTLY (PostgreSQL specific)
	concurrently := p.ParseKeyword("CONCURRENTLY")

	// Parse IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Check if index name is provided
	// In PostgreSQL, the index name is optional: CREATE INDEX ON table_name (col)
	// MySQL requires the index name: CREATE INDEX name ON table_name (col)
	var indexName *ast.Ident
	var using *ast.Ident

	// Check if we have ON keyword (meaning no index name)
	if !p.PeekKeyword("ON") {
		// Parse index name
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		indexName = name

		// Check for USING after index name (MySQL style: CREATE INDEX name USING btree ON ...)
		if p.ParseKeyword("USING") {
			using, err = p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
		}

		// Expect ON keyword
		if _, err := p.ExpectKeyword("ON"); err != nil {
			return nil, err
		}
	}

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Check for USING after table name (PostgreSQL style: CREATE INDEX ON table USING btree ...)
	if using == nil && p.ParseKeyword("USING") {
		using, err = p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
	}

	// Parse column list: (col1, col2, ...)
	columns, err := parseIndexColumnList(p)
	if err != nil {
		return nil, err
	}

	// Parse INCLUDE clause (PostgreSQL 11+)
	var include []*ast.Ident
	if p.ParseKeyword("INCLUDE") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		include, err = parseCommaSeparatedIdents(p)
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	// Parse NULLS DISTINCT / NULLS NOT DISTINCT (PostgreSQL 15+)
	var nullsDistinct *bool
	if p.ParseKeyword("NULLS") {
		notDistinct := p.ParseKeyword("NOT")
		if !notDistinct && !p.ParseKeyword("DISTINCT") {
			return nil, p.ExpectedRef("NOT DISTINCT or DISTINCT", p.PeekTokenRef())
		}
		if notDistinct {
			if !p.ParseKeyword("DISTINCT") {
				return nil, p.ExpectedRef("DISTINCT after NULLS NOT", p.PeekTokenRef())
			}
			val := false
			nullsDistinct = &val
		} else {
			val := true
			nullsDistinct = &val
		}
	}

	// Parse WITH (storage_parameters) clause
	var withOpts []*expr.SqlOption
	if p.ParseKeyword("WITH") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		withOpts, err = parseSqlOptions(p)
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	// Parse TABLESPACE clause
	var tablespace *ast.Ident
	if p.ParseKeyword("TABLESPACE") {
		tablespace, err = p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
	}

	// Parse WHERE clause (partial index)
	var predicate expr.Expr
	if p.ParseKeyword("WHERE") {
		exprParser := NewExpressionParser(p)
		predicate, err = exprParser.ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	return &statement.CreateIndex{
		Unique:        unique,
		Concurrently:  concurrently,
		IfNotExists:   ifNotExists,
		Name:          indexName,
		TableName:     tableName,
		Using:         using,
		Columns:       columns,
		Include:       include,
		NullsDistinct: nullsDistinct,
		With:          withOpts,
		TableSpace:    tablespace,
		Predicate:     predicate,
	}, nil
}

// parseIndexColumnList parses a parenthesized list of index columns
func parseIndexColumnList(p *Parser) ([]*expr.IndexColumn, error) {
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	var columns []*expr.IndexColumn
	if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
		// Empty list
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		return columns, nil
	}

	for {
		col, err := parseIndexColumn(p)
		if err != nil {
			return nil, err
		}
		columns = append(columns, col)

		if !p.ParseKeyword(",") {
			break
		}
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return columns, nil
}

// parseIndexColumn parses a single index column expression
// Format: expression [ASC|DESC] [NULLS FIRST|LAST] or expression opclass [ASC|DESC]
func parseIndexColumn(p *Parser) (*expr.IndexColumn, error) {
	exprParser := NewExpressionParser(p)

	// Parse the expression
	colExpr, err := exprParser.ParseExpr()
	if err != nil {
		return nil, err
	}

	col := &expr.IndexColumn{
		Expr: colExpr,
	}

	// Check for operator class (only if next token is not a keyword like ASC, DESC, NULLS)
	if !p.PeekKeyword("ASC") && !p.PeekKeyword("DESC") && !p.PeekKeyword("NULLS") {
		// Try to parse as operator class
		if opclass, err := tryParseOpclass(p); err == nil && opclass != nil {
			col.Opclass = opclass
		}
	}

	// Check for ASC/DESC
	if p.ParseKeyword("ASC") {
		asc := true
		col.Asc = &asc
	} else if p.ParseKeyword("DESC") {
		asc := false
		col.Asc = &asc
	}

	// Check for NULLS FIRST/LAST
	if p.ParseKeyword("NULLS") {
		if p.ParseKeyword("FIRST") {
			nullsFirst := true
			col.NullsFirst = &nullsFirst
		} else if p.ParseKeyword("LAST") {
			nullsFirst := false
			col.NullsFirst = &nullsFirst
		}
	}

	return col, nil
}

// tryParseOpclass attempts to parse an operator class
func tryParseOpclass(p *Parser) (*ast.ObjectName, error) {
	// Save position
	restore := p.SavePosition()

	// Try to parse as object name
	name, err := p.ParseObjectName()
	if err != nil {
		restore()
		return nil, err
	}

	// Check if the next token is one of the keywords that would indicate
	// this is not an operator class but part of the expression
	if p.PeekKeyword("ASC") || p.PeekKeyword("DESC") || p.PeekKeyword("NULLS") || p.PeekKeyword(",") {
		// This is an operator class
		return name, nil
	}

	// Check for right parenthesis
	if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
		// This is an operator class
		return name, nil
	}

	// Not followed by expected tokens, restore position
	restore()
	return nil, nil
}

// parseSqlOptions parses a comma-separated list of SQL options like (key = value, ...)
func parseSqlOptions(p *Parser) ([]*expr.SqlOption, error) {
	var options []*expr.SqlOption
	if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
		return options, nil
	}

	for {
		// Parse option name
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}

		// Expect =
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return nil, err
		}

		// Parse value
		exprParser := NewExpressionParser(p)
		val, err := exprParser.ParseExpr()
		if err != nil {
			return nil, err
		}

		// Convert ast.Ident to expr.Ident
		exprName := &expr.Ident{
			SpanVal:    name.Span(),
			Value:      name.Value,
			QuoteStyle: name.QuoteStyle,
		}

		options = append(options, &expr.SqlOption{
			Name:  exprName,
			Value: val,
		})

		if !p.ParseKeyword(",") {
			break
		}
	}

	return options, nil
}

func parseCreateRole(p *Parser, orReplace bool) (ast.Statement, error) {
	// ROLE keyword is expected (already checked by caller)
	if _, err := p.ExpectKeyword("ROLE"); err != nil {
		return nil, err
	}

	// Parse IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse comma-separated role names (identifiers, not full object names)
	names, err := parseCommaSeparatedIdents(p)
	if err != nil {
		return nil, err
	}

	return &statement.CreateRole{
		IfNotExists: ifNotExists,
		Names:       names,
		Options:     nil, // Role options not yet implemented (Postgres/MSSQL specific)
	}, nil
}

func parseCreateDatabase(p *Parser) (ast.Statement, error) {
	return nil, p.expectedRef("CREATE DATABASE not yet implemented", p.PeekTokenRef())
}

func parseCreateSchema(p *Parser) (ast.Statement, error) {
	// SCHEMA keyword is expected (already checked by caller)
	if _, err := p.ExpectKeyword("SCHEMA"); err != nil {
		return nil, err
	}

	// Check for IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse schema name (handles AUTHORIZATION variants)
	schemaName, err := parseSchemaName(p)
	if err != nil {
		return nil, err
	}

	// Parse optional DEFAULT COLLATE (BigQuery)
	var defaultCollateSpec expr.Expr
	if p.ParseKeywords([]string{"DEFAULT", "COLLATE"}) {
		exprParser := NewExpressionParser(p)
		defaultCollateSpec, err = exprParser.ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	// Parse optional WITH options (Trino)
	var withOpts []*expr.SqlOption
	if p.PeekKeyword("WITH") {
		// For now, skip parsing options - just consume the tokens
		// Full implementation would parse key=value pairs
		p.NextToken() // consume WITH
		if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
			p.NextToken() // consume (
			// Parse until we hit )
			for {
				if _, ok := p.PeekToken().Token.(token.TokenRParen); ok {
					p.NextToken() // consume )
					break
				}
				p.NextToken()
				if p.PeekKeyword(",") {
					p.NextToken()
				}
			}
		}
	}

	// Parse optional OPTIONS (BigQuery)
	var options []*expr.SqlOption
	if p.PeekKeyword("OPTIONS") {
		// For now, skip parsing options - just consume the tokens
		p.NextToken() // consume OPTIONS
		if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
			p.NextToken() // consume (
			// Parse until we hit )
			for {
				if _, ok := p.PeekToken().Token.(token.TokenRParen); ok {
					p.NextToken() // consume )
					break
				}
				p.NextToken()
				if p.PeekKeyword(",") {
					p.NextToken()
				}
			}
		}
	}

	// Parse optional CLONE (Snowflake)
	var clone *ast.ObjectName
	if p.ParseKeyword("CLONE") {
		clone, err = p.ParseObjectName()
		if err != nil {
			return nil, err
		}
	}

	return &statement.CreateSchema{
		SchemaName:         schemaName,
		IfNotExists:        ifNotExists,
		With:               withOpts,
		Options:            options,
		DefaultCollateSpec: defaultCollateSpec,
		Clone:              clone,
	}, nil
}

// parseSchemaName parses schema name with optional AUTHORIZATION clause
// Reference: src/parser/mod.rs parse_schema_name
// Supports:
//   - Simple: <schema_name>
//   - UnnamedAuthorization: AUTHORIZATION <user>
//   - NamedAuthorization: <schema_name> AUTHORIZATION <user>
func parseSchemaName(p *Parser) (*expr.SchemaName, error) {
	// Check for AUTHORIZATION first (UnnamedAuthorization case)
	if p.ParseKeyword("AUTHORIZATION") {
		auth, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected authorization identifier after AUTHORIZATION: %w", err)
		}
		return &expr.SchemaName{
			Authorization:    auth,
			HasAuthorization: true,
		}, nil
	}

	// Parse the schema name (could be simple identifier or object name)
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected schema name: %w", err)
	}

	// Check for AUTHORIZATION after name (NamedAuthorization case)
	if p.ParseKeyword("AUTHORIZATION") {
		auth, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected authorization identifier after AUTHORIZATION: %w", err)
		}
		return &expr.SchemaName{
			Name:             name,
			Authorization:    auth,
			HasAuthorization: true,
		}, nil
	}

	// Simple case: just the name
	return &expr.SchemaName{
		Name: name,
	}, nil
}

func parseCreateSequence(p *Parser, orReplace, temporary bool) (ast.Statement, error) {
	// Consume SEQUENCE keyword (already checked by caller)
	if _, err := p.ExpectKeyword("SEQUENCE"); err != nil {
		return nil, err
	}

	// Parse IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse sequence name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse AS data_type (optional)
	var dataType string
	if p.ParseKeyword("AS") {
		dt, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		// Store the data type name as a string
		dataType = dt.Value
	}

	// Parse sequence options
	sequenceOptions, err := parseCreateSequenceOptions(p)
	if err != nil {
		return nil, err
	}

	// Parse OWNED BY clause
	var ownedBy *ast.ObjectName
	if p.ParseKeywords([]string{"OWNED", "BY"}) {
		if p.ParseKeyword("NONE") {
			// OWNED BY NONE - represented as a special identifier
			ownedBy = &ast.ObjectName{
				Parts: []ast.ObjectNamePart{&ast.ObjectNamePartIdentifier{Ident: &ast.Ident{Value: "NONE"}}},
			}
		} else {
			ownedBy, err = p.ParseObjectName()
			if err != nil {
				return nil, err
			}
		}
	}

	return &statement.CreateSequence{
		Temporary:       temporary,
		IfNotExists:     ifNotExists,
		Name:            name,
		DataType:        dataType,
		SequenceOptions: sequenceOptions,
		OwnedBy:         ownedBy,
	}, nil
}

// parseCreateSequenceOptions parses the various options for CREATE SEQUENCE
func parseCreateSequenceOptions(p *Parser) ([]*expr.SequenceOptions, error) {
	var sequenceOptions []*expr.SequenceOptions
	exprParser := NewExpressionParser(p)

	for {
		// INCREMENT [BY] increment
		if p.ParseKeyword("INCREMENT") {
			hasBy := p.ParseKeyword("BY")
			incExpr, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:        expr.SeqOptIncrementBy,
				Expr:        incExpr,
				HasByOrWith: hasBy,
			})
			continue
		}

		// MINVALUE minvalue | NO MINVALUE
		if p.ParseKeyword("MINVALUE") {
			minExpr, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptMinValue,
				Expr:    minExpr,
				NoValue: false,
			})
			continue
		} else if p.ParseKeywords([]string{"NO", "MINVALUE"}) {
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptMinValue,
				NoValue: true,
			})
			continue
		}

		// MAXVALUE maxvalue | NO MAXVALUE
		if p.ParseKeyword("MAXVALUE") {
			maxExpr, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptMaxValue,
				Expr:    maxExpr,
				NoValue: false,
			})
			continue
		} else if p.ParseKeywords([]string{"NO", "MAXVALUE"}) {
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptMaxValue,
				NoValue: true,
			})
			continue
		}

		// START [WITH] start
		if p.ParseKeyword("START") {
			hasWith := p.ParseKeyword("WITH")
			startExpr, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:        expr.SeqOptStartWith,
				Expr:        startExpr,
				HasByOrWith: hasWith,
			})
			continue
		}

		// CACHE cache
		if p.ParseKeyword("CACHE") {
			cacheExpr, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type: expr.SeqOptCache,
				Expr: cacheExpr,
			})
			continue
		}

		// [NO] CYCLE
		if p.ParseKeywords([]string{"NO", "CYCLE"}) {
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptCycle,
				NoCycle: true,
			})
			continue
		} else if p.ParseKeyword("CYCLE") {
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptCycle,
				NoCycle: false,
			})
			continue
		}

		// No more options
		break
	}

	return sequenceOptions, nil
}

func parseCreateType(p *Parser) (ast.Statement, error) {
	return nil, p.expectedRef("CREATE TYPE not yet implemented", p.PeekTokenRef())
}

func parseCreateDomain(p *Parser) (ast.Statement, error) {
	return nil, p.expectedRef("CREATE DOMAIN not yet implemented", p.PeekTokenRef())
}

func parseCreateExtension(p *Parser) (ast.Statement, error) {
	return nil, p.expectedRef("CREATE EXTENSION not yet implemented", p.PeekTokenRef())
}

func parseCreateTrigger(p *Parser, orReplace bool) (ast.Statement, error) {
	return nil, p.expectedRef("CREATE TRIGGER not yet implemented", p.PeekTokenRef())
}

func parseCreatePolicy(p *Parser, orReplace bool) (ast.Statement, error) {
	return nil, p.expectedRef("CREATE POLICY not yet implemented", p.PeekTokenRef())
}

func parseCreateFunction(p *Parser, orReplace, temporary bool) (ast.Statement, error) {
	return nil, p.expectedRef("CREATE FUNCTION not yet implemented", p.PeekTokenRef())
}

func parseCreateVirtualTable(p *Parser) (ast.Statement, error) {
	return nil, p.expectedRef("CREATE VIRTUAL TABLE not yet implemented", p.PeekTokenRef())
}

func parseCreateMacro(p *Parser) (ast.Statement, error) {
	return nil, p.expectedRef("CREATE MACRO not yet implemented", p.PeekTokenRef())
}

func parseCreateSecret(p *Parser, orReplace, temporary bool) (ast.Statement, error) {
	return nil, p.expectedRef("CREATE SECRET not yet implemented", p.PeekTokenRef())
}

func parseCreateConnector(p *Parser, orReplace bool) (ast.Statement, error) {
	return nil, p.expectedRef("CREATE CONNECTOR not yet implemented", p.PeekTokenRef())
}

func parseCreateOperator(p *Parser) (ast.Statement, error) {
	return nil, p.expectedRef("CREATE OPERATOR not yet implemented", p.PeekTokenRef())
}

func parseCreateUser(p *Parser, orReplace bool) (ast.Statement, error) {
	return nil, p.expectedRef("CREATE USER not yet implemented", p.PeekTokenRef())
}
