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
	"strings"

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
	// Consume DATABASE keyword
	if _, err := p.ExpectKeyword("DATABASE"); err != nil {
		return nil, err
	}

	// Check for IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse database name
	dbName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse optional LOCATION/MANAGEDLOCATION (Hive/Databricks style)
	var location, managedLocation *string
	for {
		if p.PeekKeyword("LOCATION") {
			p.NextToken()
			loc, err := p.ParseStringLiteral()
			if err != nil {
				return nil, err
			}
			location = &loc
		} else if p.PeekKeyword("MANAGEDLOCATION") {
			p.NextToken()
			loc, err := p.ParseStringLiteral()
			if err != nil {
				return nil, err
			}
			managedLocation = &loc
		} else {
			break
		}
	}

	// Parse optional CLONE (Snowflake style)
	var clone *ast.ObjectName
	if p.PeekKeyword("CLONE") {
		p.NextToken()
		clone, err = p.ParseObjectName()
		if err != nil {
			return nil, err
		}
	}

	// Parse MySQL-style [DEFAULT] CHARACTER SET and [DEFAULT] COLLATE options
	var defaultCharset, defaultCollation *string
	for {
		hasDefault := p.PeekKeyword("DEFAULT")
		if hasDefault {
			p.NextToken()
		}

		if p.PeekKeyword("CHARACTER") {
			p.NextToken()
			if !p.ParseKeyword("SET") {
				// Not CHARACTER SET, put back and break
				break
			}
			p.ConsumeToken(token.TokenEq{}) // Optional =
			charset, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			defaultCharset = &charset.Value
		} else if p.PeekKeyword("CHARSET") {
			p.NextToken()
			p.ConsumeToken(token.TokenEq{}) // Optional =
			charset, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			defaultCharset = &charset.Value
		} else if p.PeekKeyword("COLLATE") {
			p.NextToken()
			p.ConsumeToken(token.TokenEq{}) // Optional =
			collation, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			defaultCollation = &collation.Value
		} else if hasDefault {
			// DEFAULT keyword not followed by CHARACTER SET, CHARSET, or COLLATE
			// Put it back and break
			break
		} else {
			break
		}
	}

	return &statement.CreateDatabase{
		DbName:           dbName,
		IfNotExists:      ifNotExists,
		Location:         location,
		ManagedLocation:  managedLocation,
		Clone:            clone,
		DefaultCharset:   defaultCharset,
		DefaultCollation: defaultCollation,
	}, nil
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
	// Parse type name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Check for AS keyword
	hasAs := p.ParseKeyword("AS")

	if !hasAs {
		// Simple CREATE TYPE name;
		return &statement.CreateType{
			Name: name,
		}, nil
	}

	// Parse AS variant
	if p.ParseKeyword("ENUM") {
		// CREATE TYPE name AS ENUM (...)
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		// Parse enum labels (comma-separated identifiers or string literals)
		for {
			if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
				break
			}
			// Just consume the token (identifier or string)
			p.AdvanceToken()
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		return &statement.CreateType{
			Name: name,
		}, nil
	}

	if p.ParseKeyword("RANGE") {
		// CREATE TYPE name AS RANGE (...)
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		// Parse range options - simplified
		for {
			if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
				break
			}
			p.AdvanceToken()
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		return &statement.CreateType{
			Name: name,
		}, nil
	}

	// Try composite type: CREATE TYPE name AS (attr1 type1, attr2 type2, ...)
	if p.ConsumeToken(token.TokenLParen{}) {
		// Parse composite attributes
		for {
			if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
				break
			}
			// Parse attribute name
			_, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			// Parse data type
			_, err = p.ParseDataType()
			if err != nil {
				return nil, err
			}
			// Optional COLLATE
			if p.ParseKeyword("COLLATE") {
				_, _ = p.ParseObjectName()
			}
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		return &statement.CreateType{
			Name: name,
		}, nil
	}

	return nil, p.expectedRef("ENUM, RANGE, or '(' after AS", p.PeekTokenRef())
}

func parseCreateDomain(p *Parser) (ast.Statement, error) {
	// Parse domain name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse AS keyword
	if _, err := p.ExpectKeyword("AS"); err != nil {
		return nil, err
	}

	// Parse data type
	_, err = p.ParseDataType()
	if err != nil {
		return nil, err
	}

	// Parse optional COLLATE
	var collation *ast.ObjectName
	if p.ParseKeyword("COLLATE") {
		collation, _ = p.ParseObjectName()
	}

	// Parse optional DEFAULT
	var defaultValue expr.Expr
	if p.ParseKeyword("DEFAULT") {
		defaultValue, _ = NewExpressionParser(p).ParseExpr()
	}

	// Parse optional constraints - simplified
	var constraints []*expr.DomainConstraint
	for {
		tok := p.PeekToken()
		// Check for EOF or end of statement
		if _, isEOF := tok.Token.(token.EOF); isEOF {
			break
		}
		// Check for semicolon using keyword-like check
		if word, ok := tok.Token.(token.TokenWord); ok && word.Value == ";" {
			break
		}
		// Try to parse a constraint
		// For now, we just skip unknown tokens
		if p.PeekKeyword("CONSTRAINT") || p.PeekKeyword("NOT") || p.PeekKeyword("NULL") ||
			p.PeekKeyword("UNIQUE") || p.PeekKeyword("PRIMARY") || p.PeekKeyword("CHECK") ||
			p.PeekKeyword("REFERENCES") {
			p.AdvanceToken()
		} else {
			break
		}
	}

	// Note: CreateDomain.DataType is *ast.DataType, but ParseDataType returns datatype.DataType
	// These are incompatible interfaces. For now, we leave DataType as nil.
	// TODO: Fix the AST type definition to use interface{} like ColumnDef does
	return &statement.CreateDomain{
		Name:         name,
		DataType:     nil,
		Collation:    collation,
		DefaultValue: defaultValue,
		Constraints:  constraints,
	}, nil
}

func parseCreateExtension(p *Parser) (ast.Statement, error) {
	// Check for OR REPLACE
	orReplace := p.ParseKeywords([]string{"OR", "REPLACE"})

	// Parse optional IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse extension name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Parse optional WITH clause
	var schema *ast.ObjectName
	var version *string
	cascade := false

	if p.ParseKeyword("WITH") {
		// Parse optional SCHEMA
		if p.ParseKeyword("SCHEMA") {
			schema, _ = p.ParseObjectName()
		}

		// Parse optional VERSION
		if p.ParseKeyword("VERSION") {
			verTok, err := p.ExpectToken(token.TokenSingleQuotedString{})
			if err == nil {
				if str, ok := verTok.Token.(token.TokenSingleQuotedString); ok {
					version = &str.Value
				}
			} else {
				// Try identifier
				verIdent, err := p.ParseIdentifier()
				if err == nil {
					v := verIdent.Value
					version = &v
				}
			}
		}

		// Parse optional CASCADE
		cascade = p.ParseKeyword("CASCADE")
	}

	return &statement.CreateExtension{
		OrReplace:   orReplace,
		IfNotExists: ifNotExists,
		Name:        name,
		Schema:      schema,
		Version:     version,
		Cascade:     cascade,
	}, nil
}

func parseCreateTrigger(p *Parser, orReplace bool) (ast.Statement, error) {
	// Consume TRIGGER keyword
	if _, err := p.ExpectKeyword("TRIGGER"); err != nil {
		return nil, err
	}

	// Parse trigger name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	trigger := &statement.CreateTrigger{
		Name:              name,
		OrReplace:         orReplace,
		PeriodBeforeTable: true, // Default to PostgreSQL/MySQL style (period before ON)
	}

	// Parse optional period (BEFORE, AFTER, INSTEAD OF, FOR)
	if p.PeekKeyword("BEFORE") {
		p.NextToken()
		period := expr.TriggerPeriodBefore
		trigger.Period = &period
	} else if p.PeekKeyword("AFTER") {
		p.NextToken()
		period := expr.TriggerPeriodAfter
		trigger.Period = &period
	} else if p.PeekKeyword("INSTEAD") {
		p.NextToken()
		if _, err := p.ExpectKeyword("OF"); err != nil {
			return nil, err
		}
		period := expr.TriggerPeriodInsteadOf
		trigger.Period = &period
	} else if p.PeekKeyword("FOR") {
		p.NextToken()
		period := expr.TriggerPeriodFor
		trigger.Period = &period
	}

	// Parse trigger events (can be OR-separated: INSERT OR UPDATE OR DELETE)
	for {
		event, eventCols, err := parseTriggerEvent(p)
		if err != nil {
			return nil, err
		}
		trigger.Events = append(trigger.Events, &expr.TriggerEventWithColumns{
			Event:   event,
			Columns: eventCols,
		})

		// Check for OR keyword to parse more events
		if !p.PeekKeyword("OR") {
			break
		}
		p.NextToken() // consume OR
	}

	// Expect ON keyword
	if _, err := p.ExpectKeyword("ON"); err != nil {
		return nil, err
	}

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}
	trigger.TableName = tableName

	// Parse optional FROM clause (referenced table)
	if p.PeekKeyword("FROM") {
		p.NextToken()
		refTable, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		trigger.ReferencedTableName = refTable
	}

	// Parse optional constraint characteristics (DEFERRABLE, etc.)
	// For now, just parse and discard - full implementation TODO
	parseConstraintCharacteristics(p)

	// Parse optional REFERENCING clause
	if p.PeekKeyword("REFERENCING") {
		p.NextToken()
		for {
			referencing, err := parseTriggerReferencing(p)
			if err != nil {
				return nil, err
			}
			if referencing == nil {
				break
			}
			trigger.Referencing = append(trigger.Referencing, referencing)
			// Check if next token could be another referencing clause
			if !p.PeekKeyword("OLD") && !p.PeekKeyword("NEW") {
				break
			}
		}
	}

	// Parse optional FOR [EACH] ROW/STATEMENT
	if p.PeekKeyword("FOR") {
		p.NextToken()
		kind := expr.TriggerObjectKindFor
		if p.PeekKeyword("EACH") {
			p.NextToken()
			kind = expr.TriggerObjectKindForEach
		}

		var obj expr.TriggerObject
		if p.PeekKeyword("ROW") {
			p.NextToken()
			obj = expr.TriggerObjectRow
		} else if p.PeekKeyword("STATEMENT") {
			p.NextToken()
			obj = expr.TriggerObjectStatement
		}
		trigger.TriggerObject = &expr.TriggerObjectKindWithObject{
			Kind:   kind,
			Object: obj,
		}
	}

	// Parse optional WHEN clause
	if p.PeekKeyword("WHEN") {
		p.NextToken()
		condition, err := NewExpressionParser(p).ParseExpr()
		if err != nil {
			return nil, err
		}
		trigger.Condition = condition
	}

	// Parse EXECUTE clause (FUNCTION or PROCEDURE)
	if p.PeekKeyword("EXECUTE") {
		p.NextToken()
		execBody, err := parseTriggerExecBody(p)
		if err != nil {
			return nil, err
		}
		trigger.ExecBody = execBody
	} else {
		// Parse statement body (for T-SQL style triggers)
		// For now, skip this - it's complex conditional statement parsing
	}

	return trigger, nil
}

func parseTriggerEvent(p *Parser) (expr.TriggerEvent, []*ast.Ident, error) {
	switch {
	case p.PeekKeyword("INSERT"):
		p.NextToken()
		return expr.TriggerEventInsert, nil, nil
	case p.PeekKeyword("UPDATE"):
		p.NextToken()
		var cols []*ast.Ident
		if p.PeekKeyword("OF") {
			p.NextToken()
			// Parse column list
			for {
				col, err := p.ParseIdentifier()
				if err != nil {
					return expr.TriggerEventNone, nil, err
				}
				cols = append(cols, col)
				if !p.ConsumeToken(token.TokenComma{}) {
					break
				}
			}
		}
		return expr.TriggerEventUpdate, cols, nil
	case p.PeekKeyword("DELETE"):
		p.NextToken()
		return expr.TriggerEventDelete, nil, nil
	case p.PeekKeyword("TRUNCATE"):
		p.NextToken()
		return expr.TriggerEventTruncate, nil, nil
	default:
		return expr.TriggerEventNone, nil, fmt.Errorf("expected INSERT, UPDATE, DELETE, or TRUNCATE, found %v", p.PeekToken())
	}
}

func parseTriggerReferencing(p *Parser) (*expr.TriggerReferencing, error) {
	var referType expr.TriggerReferencingType

	if p.PeekKeyword("OLD") {
		p.NextToken()
		if p.PeekKeyword("TABLE") {
			p.NextToken()
			referType = expr.TriggerReferencingTypeOldTable
		} else {
			// Not a valid referencing clause
			return nil, nil
		}
	} else if p.PeekKeyword("NEW") {
		p.NextToken()
		if p.PeekKeyword("TABLE") {
			p.NextToken()
			referType = expr.TriggerReferencingTypeNewTable
		} else {
			// Not a valid referencing clause
			return nil, nil
		}
	} else {
		return nil, nil
	}

	isAs := p.PeekKeyword("AS")
	if isAs {
		p.NextToken()
	}

	transitionName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	return &expr.TriggerReferencing{
		ReferType:              referType,
		IsAs:                   isAs,
		TransitionRelationName: transitionName,
	}, nil
}

func parseTriggerExecBody(p *Parser) (*expr.TriggerExecBody, error) {
	var execType expr.TriggerExecBodyType

	if p.PeekKeyword("FUNCTION") {
		p.NextToken()
		execType = expr.TriggerExecBodyTypeFunction
	} else if p.PeekKeyword("PROCEDURE") {
		p.NextToken()
		execType = expr.TriggerExecBodyTypeProcedure
	} else {
		return nil, fmt.Errorf("expected FUNCTION or PROCEDURE after EXECUTE")
	}

	// Parse function/procedure name and optional arguments
	funcName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	var args []expr.Expr
	if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
		p.NextToken() // consume (
		if _, ok := p.PeekToken().Token.(token.TokenRParen); !ok {
			for {
				arg, err := NewExpressionParser(p).ParseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				if !p.ConsumeToken(token.TokenComma{}) {
					break
				}
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	return &expr.TriggerExecBody{
		ExecType: execType,
		FuncDesc: &expr.FunctionDesc{
			Name: funcName,
			Args: args,
		},
	}, nil
}

func parseCreatePolicy(p *Parser, orReplace bool) (ast.Statement, error) {
	// Parse policy name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Parse ON keyword and table name
	if _, err := p.ExpectKeyword("ON"); err != nil {
		return nil, err
	}
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse optional AS PERMISSIVE/RESTRICTIVE
	var policyType *expr.CreatePolicyType
	if p.ParseKeyword("AS") {
		if p.ParseKeyword("PERMISSIVE") {
			pt := expr.CreatePolicyTypePermissive
			policyType = &pt
		} else if p.ParseKeyword("RESTRICTIVE") {
			pt := expr.CreatePolicyTypeRestrictive
			policyType = &pt
		} else {
			return nil, p.Expected("PERMISSIVE or RESTRICTIVE after AS", p.PeekToken())
		}
	}

	// Parse optional FOR ALL|SELECT|INSERT|UPDATE|DELETE
	var command *expr.CreatePolicyCommand
	if p.ParseKeyword("FOR") {
		if p.ParseKeyword("ALL") {
			cmd := expr.CreatePolicyCommandAll
			command = &cmd
		} else if p.ParseKeyword("SELECT") {
			cmd := expr.CreatePolicyCommandSelect
			command = &cmd
		} else if p.ParseKeyword("INSERT") {
			cmd := expr.CreatePolicyCommandInsert
			command = &cmd
		} else if p.ParseKeyword("UPDATE") {
			cmd := expr.CreatePolicyCommandUpdate
			command = &cmd
		} else if p.ParseKeyword("DELETE") {
			cmd := expr.CreatePolicyCommandDelete
			command = &cmd
		} else {
			return nil, p.Expected("ALL, SELECT, INSERT, UPDATE, or DELETE after FOR", p.PeekToken())
		}
	}

	// Parse optional TO clause (role names)
	var to []*expr.Owner
	if p.ParseKeyword("TO") {
		for {
			owner, err := parseOwner(p)
			if err != nil {
				return nil, err
			}
			to = append(to, owner)
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
	}

	// Parse optional USING (expression)
	var usingExpr expr.Expr
	if p.ParseKeyword("USING") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		ep := NewExpressionParser(p)
		usingExpr, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	// Parse optional WITH CHECK (expression)
	var withCheckExpr expr.Expr
	if p.ParseKeywords([]string{"WITH", "CHECK"}) {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		ep := NewExpressionParser(p)
		withCheckExpr, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	return &statement.CreatePolicy{
		OrReplace:  orReplace,
		Name:       name,
		TableName:  tableName,
		PolicyType: policyType,
		Command:    command,
		To:         to,
		Using:      usingExpr,
		WithCheck:  withCheckExpr,
	}, nil
}

// parseOwner parses an owner specification (CURRENT_USER, CURRENT_ROLE, SESSION_USER, or identifier)
// Reference: src/parser/mod.rs:6795
func parseOwner(p *Parser) (*expr.Owner, error) {
	if p.ParseKeyword("CURRENT_USER") {
		return &expr.Owner{Kind: expr.OwnerKindCurrentUser}, nil
	}
	if p.ParseKeyword("CURRENT_ROLE") {
		return &expr.Owner{Kind: expr.OwnerKindCurrentRole}, nil
	}
	if p.ParseKeyword("SESSION_USER") {
		return &expr.Owner{Kind: expr.OwnerKindSessionUser}, nil
	}

	// Otherwise, parse as identifier
	ident, err := p.ParseIdentifier()
	if err != nil {
		return nil, p.Expected("CURRENT_USER, CURRENT_ROLE, SESSION_USER, or identifier", p.PeekToken())
	}
	return &expr.Owner{Kind: expr.OwnerKindIdent, Ident: ident}, nil
}

func parseCreateFunction(p *Parser, orReplace, temporary bool) (ast.Statement, error) {
	// Consume FUNCTION keyword
	if _, err := p.ExpectKeyword("FUNCTION"); err != nil {
		return nil, err
	}

	// Parse function name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse function arguments: (arg1 TYPE, arg2 TYPE, ...)
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	var args []*expr.OperateFunctionArg
	if _, ok := p.PeekToken().Token.(token.TokenRParen); !ok {
		// Parse comma-separated arguments
		for {
			arg, err := parseFunctionArg(p)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)

			if p.ConsumeToken(token.TokenComma{}) {
				continue
			}
			break
		}
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// Parse optional RETURNS clause
	var returnType *expr.FunctionReturnType
	if p.PeekKeyword("RETURNS") {
		p.NextToken()
		returnType, err = parseFunctionReturnType(p)
		if err != nil {
			return nil, err
		}
	}

	// Parse function attributes (LANGUAGE, AS, IMMUTABLE, etc.)
	var language *ast.Ident
	var behavior *expr.FunctionBehavior
	var calledOnNull *expr.FunctionCalledOnNull
	var parallel *expr.FunctionParallel
	var security *expr.FunctionSecurity
	var body *expr.CreateFunctionBody
	var setParams []*expr.FunctionDefinitionSetParam

	for {
		if p.PeekKeyword("LANGUAGE") {
			p.NextToken()
			lang, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			language = lang
		} else if p.PeekKeyword("AS") {
			p.NextToken()
			// Parse function body as string literal or dollar-quoted string
			bodyStr, err := p.ParseStringLiteral()
			if err != nil {
				return nil, err
			}
			body = &expr.CreateFunctionBody{Value: bodyStr}
		} else if p.PeekKeyword("IMMUTABLE") {
			p.NextToken()
			b := expr.FunctionBehaviorImmutable
			behavior = &b
		} else if p.PeekKeyword("STABLE") {
			p.NextToken()
			b := expr.FunctionBehaviorStable
			behavior = &b
		} else if p.PeekKeyword("VOLATILE") {
			p.NextToken()
			b := expr.FunctionBehaviorVolatile
			behavior = &b
		} else if p.ParseKeywords([]string{"CALLED", "ON", "NULL", "INPUT"}) {
			c := expr.FunctionCalledOnNullCalledOnNullInput
			calledOnNull = &c
		} else if p.ParseKeywords([]string{"RETURNS", "NULL", "ON", "NULL", "INPUT"}) {
			c := expr.FunctionCalledOnNullReturnsNullOnNullInput
			calledOnNull = &c
		} else if p.PeekKeyword("STRICT") {
			p.NextToken()
			c := expr.FunctionCalledOnNullStrict
			calledOnNull = &c
		} else if p.PeekKeyword("PARALLEL") {
			p.NextToken()
			if p.PeekKeyword("UNSAFE") {
				p.NextToken()
				par := expr.FunctionParallelUnsafe
				parallel = &par
			} else if p.PeekKeyword("RESTRICTED") {
				p.NextToken()
				par := expr.FunctionParallelRestricted
				parallel = &par
			} else if p.PeekKeyword("SAFE") {
				p.NextToken()
				par := expr.FunctionParallelSafe
				parallel = &par
			}
		} else if p.ParseKeywords([]string{"SECURITY", "DEFINER"}) {
			s := expr.FunctionSecurityDefiner
			security = &s
		} else if p.ParseKeywords([]string{"SECURITY", "INVOKER"}) {
			s := expr.FunctionSecurityInvoker
			security = &s
		} else if p.PeekKeyword("SET") {
			p.NextToken()
			// Parse SET param_name = value or SET param_name FROM CURRENT
			paramName, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			var paramValue expr.FunctionSetValue
			if p.ParseKeywords([]string{"FROM", "CURRENT"}) {
				paramValue = expr.FunctionSetValue{Kind: expr.FunctionSetValueFromCurrent}
			} else {
				// Parse = or TO followed by values
				if !p.ConsumeToken(token.TokenEq{}) && !p.PeekKeyword("TO") {
					return nil, fmt.Errorf("expected = or TO after SET parameter name")
				}
				if p.PeekKeyword("TO") {
					p.NextToken()
				}
				// For simplicity, parse a single expression value
				exprParser := NewExpressionParser(p)
				value, err := exprParser.ParseExpr()
				if err != nil {
					return nil, err
				}
				paramValue = expr.FunctionSetValue{Kind: expr.FunctionSetValueExpr, Expr: value}
			}
			setParams = append(setParams, &expr.FunctionDefinitionSetParam{
				Name:  paramName,
				Value: paramValue,
			})
		} else if p.PeekKeyword("RETURN") {
			p.NextToken()
			exprParser := NewExpressionParser(p)
			retExpr, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			body = &expr.CreateFunctionBody{ReturnExpr: retExpr}
		} else {
			break
		}
	}

	return &statement.CreateFunction{
		OrReplace:    orReplace,
		Temporary:    temporary,
		Name:         name,
		Args:         args,
		ReturnType:   returnType,
		Language:     language,
		Behavior:     behavior,
		CalledOnNull: calledOnNull,
		Parallel:     parallel,
		Security:     security,
		Body:         body,
		Set:          setParams,
	}, nil
}

// parseFunctionArg parses a function argument like "name TYPE" or "IN name TYPE"
// Reference: src/parser/mod.rs:5972
func parseFunctionArg(p *Parser) (*expr.OperateFunctionArg, error) {
	// Check for IN/OUT/INOUT mode
	var mode *expr.ArgMode
	if p.PeekKeyword("IN") {
		p.NextToken()
		m := expr.ArgModeIn
		mode = &m
	} else if p.PeekKeyword("OUT") {
		p.NextToken()
		m := expr.ArgModeOut
		mode = &m
	} else if p.PeekKeyword("INOUT") {
		p.NextToken()
		m := expr.ArgModeInOut
		mode = &m
	}

	// Try to parse the first token - it could be either:
	// 1. A parameter name followed by a data type (e.g., "str1 VARCHAR")
	// 2. Just a data type (e.g., "INTEGER")

	// Save current position for potential backtracking
	savedIdx := p.index

	// Try to parse as data type first
	firstDataType, err := p.ParseDataType()
	if err != nil {
		// Failed to parse as data type, try as identifier then data type
		p.index = savedIdx
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		dataType, err := p.ParseDataType()
		if err != nil {
			return nil, err
		}

		// Check for DEFAULT or = value
		var defaultExpr expr.Expr
		if p.PeekKeyword("DEFAULT") || p.ConsumeToken(token.TokenEq{}) {
			if p.PeekKeyword("DEFAULT") {
				p.NextToken()
			}
			exprParser := NewExpressionParser(p)
			defaultExpr, err = exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
		}

		return &expr.OperateFunctionArg{
			Mode:        mode,
			Name:        name,
			DataType:    dataType,
			DefaultExpr: defaultExpr,
		}, nil
	}

	// We successfully parsed a data type. Now check if the next token could also be
	// a data type keyword (which would mean the first token was actually a parameter name).
	// This is the Rust approach: try to parse another data type, and if it succeeds,
	// treat the first as a name.

	// For simplicity, we check if next token looks like a type keyword
	// Common SQL type keywords: VARCHAR, INTEGER, INT, TEXT, BOOLEAN, etc.
	// If next token is one of these, the first token was likely a name.

	// Save position again
	secondIdx := p.index

	// Try to peek at the next token to see if it looks like a type
	nextTok := p.PeekToken()
	if word, ok := nextTok.Token.(token.TokenWord); ok {
		// Check if next token is a common SQL type
		typeKeywords := map[string]bool{
			"VARCHAR": true, "CHAR": true, "TEXT": true, "INTEGER": true,
			"INT": true, "BIGINT": true, "SMALLINT": true, "BOOLEAN": true,
			"BOOL": true, "REAL": true, "DOUBLE": true, "FLOAT": true,
			"DECIMAL": true, "NUMERIC": true, "DATE": true, "TIME": true,
			"TIMESTAMP": true, "INTERVAL": true, "ARRAY": true, "JSON": true,
			"JSONB": true, "BYTEA": true, "UUID": true, "SERIAL": true,
			"BIGSERIAL": true, "SMALLSERIAL": true, "MONEY": true,
		}
		upperWord := strings.ToUpper(word.Word.Value)
		if typeKeywords[upperWord] {
			// The next token is a type keyword, so first token was a name
			// We need to parse the second data type
			p.index = secondIdx
			secondDataType, err := p.ParseDataType()
			if err != nil {
				// If we fail to parse second, just use first as data type
				p.index = secondIdx

				// Check for DEFAULT or = value
				var defaultExpr expr.Expr
				if p.PeekKeyword("DEFAULT") || p.ConsumeToken(token.TokenEq{}) {
					if p.PeekKeyword("DEFAULT") {
						p.NextToken()
					}
					exprParser := NewExpressionParser(p)
					defaultExpr, err = exprParser.ParseExpr()
					if err != nil {
						return nil, err
					}
				}

				return &expr.OperateFunctionArg{
					Mode:        mode,
					DataType:    firstDataType,
					DefaultExpr: defaultExpr,
				}, nil
			}

			// Create identifier from first "data type" (which was actually a name)
			name := &ast.Ident{Value: firstDataType.(fmt.Stringer).String()}

			// Check for DEFAULT or = value
			var defaultExpr expr.Expr
			if p.PeekKeyword("DEFAULT") || p.ConsumeToken(token.TokenEq{}) {
				if p.PeekKeyword("DEFAULT") {
					p.NextToken()
				}
				exprParser := NewExpressionParser(p)
				defaultExpr, err = exprParser.ParseExpr()
				if err != nil {
					return nil, err
				}
			}

			return &expr.OperateFunctionArg{
				Mode:        mode,
				Name:        name,
				DataType:    secondDataType,
				DefaultExpr: defaultExpr,
			}, nil
		}
	}

	// Next token is not a type keyword, so firstDataType is the actual data type
	// Check for DEFAULT or = value
	var defaultExpr expr.Expr
	if p.PeekKeyword("DEFAULT") || p.ConsumeToken(token.TokenEq{}) {
		if p.PeekKeyword("DEFAULT") {
			p.NextToken()
		}
		exprParser := NewExpressionParser(p)
		defaultExpr, err = exprParser.ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	return &expr.OperateFunctionArg{
		Mode:        mode,
		DataType:    firstDataType,
		DefaultExpr: defaultExpr,
	}, nil
}

// parseFunctionReturnType parses a return type like "INTEGER" or "SETOF INTEGER"
func parseFunctionReturnType(p *Parser) (*expr.FunctionReturnType, error) {
	if p.PeekKeyword("SETOF") {
		p.NextToken()
		dataType, err := p.ParseDataType()
		if err != nil {
			return nil, err
		}
		return &expr.FunctionReturnType{Kind: expr.FunctionReturnTypeSetOf, DataType: dataType}, nil
	}

	dataType, err := p.ParseDataType()
	if err != nil {
		return nil, err
	}
	return &expr.FunctionReturnType{Kind: expr.FunctionReturnTypeDataType, DataType: dataType}, nil
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
	// Check for IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse connector name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Parse optional TYPE 'datasource_type'
	var connectorType *string
	if p.ParseKeyword("TYPE") {
		tok := p.PeekToken()
		if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			p.AdvanceToken()
			connectorType = &str.Value
		} else {
			return nil, p.Expected("string literal after TYPE", tok)
		}
	}

	// Parse optional URL 'datasource_url'
	var url *string
	if p.ParseKeyword("URL") {
		tok := p.PeekToken()
		if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			p.AdvanceToken()
			url = &str.Value
		} else {
			return nil, p.Expected("string literal after URL", tok)
		}
	}

	// Parse optional COMMENT 'comment'
	var comment *string
	if p.ParseKeyword("COMMENT") {
		// Optional = sign
		p.ParseKeyword("=")
		tok := p.PeekToken()
		if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			p.AdvanceToken()
			comment = &str.Value
		} else {
			return nil, p.Expected("string literal after COMMENT", tok)
		}
	}

	// Parse optional WITH DCPROPERTIES (property_name=property_value, ...)
	var dcProperties []*expr.SqlOption
	if p.ParseKeywords([]string{"WITH", "DCPROPERTIES"}) {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		for {
			// Parse property name
			propName, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			// Expect =
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			// Parse property value (identifier or string)
			var propValue expr.Expr
			tok := p.PeekToken()
			if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
				p.AdvanceToken()
				propValue = &expr.ValueExpr{
					Value: str.Value,
				}
			} else {
				propVal, err := p.ParseIdentifier()
				if err != nil {
					return nil, err
				}
				propValue = &expr.Ident{Value: propVal.Value}
			}
			dcProperties = append(dcProperties, &expr.SqlOption{
				Name:  &expr.Ident{Value: propName.Value},
				Value: propValue,
			})
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	return &statement.CreateConnector{
		IfNotExists:      ifNotExists,
		Name:             name,
		ConnectorType:    connectorType,
		URL:              url,
		Comment:          comment,
		WithDCProperties: dcProperties,
	}, nil
}

func parseCreateOperator(p *Parser) (ast.Statement, error) {
	// Consume OPERATOR keyword
	if _, err := p.ExpectKeyword("OPERATOR"); err != nil {
		return nil, err
	}

	// Check for FAMILY or CLASS
	if p.PeekKeyword("FAMILY") {
		return parseCreateOperatorFamily(p)
	}
	if p.PeekKeyword("CLASS") {
		return parseCreateOperatorClass(p)
	}

	// Parse operator name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Expect opening parenthesis
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	createOp := &statement.CreateOperator{
		Name: name,
	}

	// Parse operator parameters
	for {
		done := false
		switch {
		case p.PeekKeyword("FUNCTION"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			funcName, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			createOp.Function = funcName
			createOp.IsProcedure = false

		case p.PeekKeyword("PROCEDURE"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			funcName, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			createOp.Function = funcName
			createOp.IsProcedure = true

		case p.PeekKeyword("LEFTARG"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			// Parse data type
			dataType, err := p.ParseDataType()
			if err != nil {
				return nil, err
			}
			createOp.LeftArg = dataType

		case p.PeekKeyword("RIGHTARG"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			// Parse data type
			dataType, err := p.ParseDataType()
			if err != nil {
				return nil, err
			}
			createOp.RightArg = dataType

		case p.PeekKeyword("HASHES"):
			p.NextToken()
			createOp.Options = append(createOp.Options, &expr.OperatorOption{
				Kind: expr.OperatorOptionKindHashes,
			})

		case p.PeekKeyword("MERGES"):
			p.NextToken()
			createOp.Options = append(createOp.Options, &expr.OperatorOption{
				Kind: expr.OperatorOptionKindMerges,
			})

		case p.PeekKeyword("COMMUTATOR"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			var opName *ast.ObjectName
			if p.PeekKeyword("OPERATOR") {
				p.NextToken()
				if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
					return nil, err
				}
				opName, err = p.ParseObjectName()
				if err != nil {
					return nil, err
				}
				if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
					return nil, err
				}
			} else {
				opName, err = p.ParseObjectName()
				if err != nil {
					return nil, err
				}
			}
			createOp.Options = append(createOp.Options, &expr.OperatorOption{
				Kind: expr.OperatorOptionKindCommutator,
				Name: opName,
			})

		case p.PeekKeyword("NEGATOR"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			var opName *ast.ObjectName
			if p.PeekKeyword("OPERATOR") {
				p.NextToken()
				if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
					return nil, err
				}
				opName, err = p.ParseObjectName()
				if err != nil {
					return nil, err
				}
				if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
					return nil, err
				}
			} else {
				opName, err = p.ParseObjectName()
				if err != nil {
					return nil, err
				}
			}
			createOp.Options = append(createOp.Options, &expr.OperatorOption{
				Kind: expr.OperatorOptionKindNegator,
				Name: opName,
			})

		case p.PeekKeyword("RESTRICT"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			funcName, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			createOp.Options = append(createOp.Options, &expr.OperatorOption{
				Kind: expr.OperatorOptionKindRestrict,
				Name: funcName,
			})

		case p.PeekKeyword("JOIN"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			funcName, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			createOp.Options = append(createOp.Options, &expr.OperatorOption{
				Kind: expr.OperatorOptionKindJoin,
				Name: funcName,
			})

		default:
			done = true
		}

		if done {
			break
		}

		// Check for comma separator
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	// Expect closing parenthesis
	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// Validate that FUNCTION was specified
	if createOp.Function == nil {
		return nil, fmt.Errorf("CREATE OPERATOR requires FUNCTION parameter")
	}

	return createOp, nil
}

func parseCreateOperatorFamily(p *Parser) (ast.Statement, error) {
	// Consume FAMILY keyword
	if _, err := p.ExpectKeyword("FAMILY"); err != nil {
		return nil, err
	}

	// Parse family name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Expect USING keyword
	if _, err := p.ExpectKeyword("USING"); err != nil {
		return nil, err
	}

	// Parse index method
	indexMethod, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	return &statement.CreateOperatorFamily{
		Name:        name,
		IndexMethod: indexMethod,
	}, nil
}

func parseCreateOperatorClass(p *Parser) (ast.Statement, error) {
	// Consume CLASS keyword
	if _, err := p.ExpectKeyword("CLASS"); err != nil {
		return nil, err
	}

	// Parse class name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	createOpClass := &statement.CreateOperatorClass{
		Name: name,
	}

	// Check for DEFAULT
	if p.PeekKeyword("DEFAULT") {
		p.NextToken()
		createOpClass.IsDefault = true
	}

	// Expect FOR TYPE keywords
	if err := p.ExpectKeywords([]string{"FOR", "TYPE"}); err != nil {
		return nil, err
	}

	// Parse data type
	dataType, err := p.ParseDataType()
	if err != nil {
		return nil, err
	}
	createOpClass.DataType = dataType

	// Expect USING keyword
	if _, err := p.ExpectKeyword("USING"); err != nil {
		return nil, err
	}

	// Parse index method
	indexMethod, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	createOpClass.IndexMethod = indexMethod

	// Check for FAMILY clause
	if p.PeekKeyword("FAMILY") {
		p.NextToken()
		family, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		createOpClass.Family = family
	}

	// Expect AS keyword
	if _, err := p.ExpectKeyword("AS"); err != nil {
		return nil, err
	}

	// Parse operator class items
	for {
		item, err := parseOperatorClassItem(p)
		if err != nil {
			return nil, err
		}
		if item == nil {
			break
		}
		createOpClass.Items = append(createOpClass.Items, item)

		// Check for comma separator
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return createOpClass, nil
}

func parseOperatorClassItem(p *Parser) (*expr.OperatorClassItem, error) {
	if p.PeekKeyword("OPERATOR") {
		p.NextToken()

		// Parse strategy number
		stratNum, err := parseLiteralUint(p)
		if err != nil {
			return nil, err
		}

		// Parse operator name
		opName, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}

		item := &expr.OperatorClassItem{
			IsOperator:     true,
			StrategyNumber: stratNum,
			OperatorName:   opName,
		}

		// Check for optional argument types
		if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
			p.NextToken()
			leftType, err := p.ParseDataType()
			if err != nil {
				return nil, err
			}
			if _, err := p.ExpectToken(token.TokenComma{}); err != nil {
				return nil, err
			}
			rightType, err := p.ParseDataType()
			if err != nil {
				return nil, err
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
			item.OpTypes = &expr.OperatorArgTypes{
				Left:  leftType,
				Right: rightType,
			}
		}

		// Check for optional purpose (FOR SEARCH or FOR ORDER BY)
		if p.PeekKeyword("FOR") {
			p.NextToken()
			if p.PeekKeyword("SEARCH") {
				p.NextToken()
				item.Purpose = &expr.OperatorPurposeWithFamily{
					Purpose: expr.OperatorPurposeForSearch,
				}
			} else if p.PeekKeyword("ORDER") {
				p.NextToken()
				if _, err := p.ExpectKeyword("BY"); err != nil {
					return nil, err
				}
				sortFamily, err := p.ParseObjectName()
				if err != nil {
					return nil, err
				}
				item.Purpose = &expr.OperatorPurposeWithFamily{
					Purpose:    expr.OperatorPurposeForOrderBy,
					SortFamily: sortFamily,
				}
			}
		}

		return item, nil
	}

	if p.PeekKeyword("FUNCTION") {
		p.NextToken()

		// Parse support number
		supportNum, err := parseLiteralUint(p)
		if err != nil {
			return nil, err
		}

		// Parse function name
		funcName, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}

		item := &expr.OperatorClassItem{
			IsFunction:    true,
			SupportNumber: supportNum,
			FunctionName:  funcName,
		}

		// Parse argument types
		if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
			p.NextToken()
			if _, ok := p.PeekToken().Token.(token.TokenRParen); !ok {
				for {
					argType, err := p.ParseDataType()
					if err != nil {
						return nil, err
					}
					item.ArgumentTypes = append(item.ArgumentTypes, argType)
					if !p.ConsumeToken(token.TokenComma{}) {
						break
					}
				}
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		}

		return item, nil
	}

	if p.PeekKeyword("STORAGE") {
		p.NextToken()
		storageType, err := p.ParseDataType()
		if err != nil {
			return nil, err
		}
		return &expr.OperatorClassItem{
			IsStorage:   true,
			StorageType: storageType,
		}, nil
	}

	// No more items
	return nil, nil
}

func parseLiteralUint(p *Parser) (uint64, error) {
	tok := p.PeekToken()
	if numTok, ok := tok.Token.(token.TokenNumber); ok {
		var val uint64
		_, err := fmt.Sscanf(numTok.Value, "%d", &val)
		if err != nil {
			return 0, fmt.Errorf("expected unsigned integer, got %s", numTok.Value)
		}
		p.NextToken()
		return val, nil
	}
	// Also try parsing a word that represents a number
	if wordTok, ok := tok.Token.(token.TokenWord); ok {
		var val uint64
		_, err := fmt.Sscanf(wordTok.Word.Value, "%d", &val)
		if err == nil {
			p.NextToken()
			return val, nil
		}
	}
	return 0, fmt.Errorf("expected unsigned integer, got %v", tok)
}

func parseCreateUser(p *Parser, orReplace bool) (ast.Statement, error) {
	// Parse optional IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse user name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Parse optional WITH options (simplified - just look for key=value patterns)
	var options []*expr.SqlOption
	for {
		if p.PeekKeyword("WITH") || p.PeekKeyword("TAG") {
			break
		}
		// Try to parse key=value option
		key, err := p.ParseIdentifier()
		if err != nil {
			p.PrevToken()
			break
		}
		if !p.ConsumeToken(token.TokenEq{}) {
			p.PrevToken()
			break
		}
		value, err := NewExpressionParser(p).ParseExpr()
		if err != nil {
			p.PrevToken()
			p.PrevToken()
			break
		}
		// Convert ast.Ident to expr.Ident
		exprKey := &expr.Ident{
			SpanVal:    key.Span(),
			Value:      key.Value,
			QuoteStyle: key.QuoteStyle,
		}
		options = append(options, &expr.SqlOption{
			Name:  exprKey,
			Value: value,
		})
	}

	// Parse optional WITH TAG
	if p.ParseKeyword("WITH") {
		p.ParseKeyword("TAG")
	}

	return &statement.CreateUser{
		IfNotExists: ifNotExists,
		Name:        name,
		Options:     options,
	}, nil
}
