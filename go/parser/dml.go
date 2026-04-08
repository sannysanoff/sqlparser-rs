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
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/query"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/token"
)

// parseInsert parses an INSERT statement
func parseInsert(p *Parser, tok token.TokenWithSpan) (ast.Statement, error) {
	return parseInsertInternal(p, tok)
}

// parseReplace parses a REPLACE statement (MySQL-specific)
// REPLACE works like INSERT but deletes the old row if a duplicate key exists
func parseReplace(p *Parser, tok token.TokenWithSpan) (ast.Statement, error) {
	// REPLACE is only supported by MySQL and Generic dialects
	dialectName := p.dialect.Dialect()
	if dialectName != "mysql" && dialectName != "generic" {
		return nil, p.expected("REPLACE not supported by this dialect", p.PeekToken())
	}

	// Parse as INSERT but with replace_into flag set
	stmt, err := parseInsertInternal(p, tok)
	if err != nil {
		return nil, err
	}

	// Set the ReplaceInto flag on the insert statement
	if insert, ok := stmt.(*statement.Insert); ok {
		insert.ReplaceInto = true
	}

	return stmt, nil
}

// parseInsertInternal parses an INSERT statement
// Reference: src/parser/mod.rs:17376
func parseInsertInternal(p *Parser, insertToken token.TokenWithSpan) (ast.Statement, error) {
	// Parse optimizer hints if present (MySQL style: INSERT /*+ hint */)
	hintsInterface, err := maybeParseOptimizerHints(p)
	if err != nil {
		return nil, err
	}

	// Convert optimizer hints to proper type
	var optimizerHints []*expr.OptimizerHint
	for _, h := range hintsInterface {
		if hint, ok := h.(*expr.OptimizerHint); ok {
			optimizerHints = append(optimizerHints, hint)
		}
	}

	// Parse optional OR conflict clause (SQLite style: INSERT OR REPLACE)
	var orConflict *expr.SqliteOnConflict
	if p.ParseKeywords([]string{"OR", "REPLACE"}) {
		orConflictVal := expr.SqliteOnConflictReplace
		orConflict = &orConflictVal
	} else if p.ParseKeywords([]string{"OR", "ROLLBACK"}) {
		orConflictVal := expr.SqliteOnConflictRollback
		orConflict = &orConflictVal
	} else if p.ParseKeywords([]string{"OR", "ABORT"}) {
		orConflictVal := expr.SqliteOnConflictAbort
		orConflict = &orConflictVal
	} else if p.ParseKeywords([]string{"OR", "FAIL"}) {
		orConflictVal := expr.SqliteOnConflictFail
		orConflict = &orConflictVal
	} else if p.ParseKeywords([]string{"OR", "IGNORE"}) {
		orConflictVal := expr.SqliteOnConflictIgnore
		orConflict = &orConflictVal
	}

	// Parse MySQL priority keywords (LOW_PRIORITY, DELAYED, HIGH_PRIORITY)
	var priority *expr.MysqlInsertPriority
	if p.dialect.SupportsInsertSet() || p.dialect.Dialect() == "mysql" || p.dialect.Dialect() == "generic" {
		if p.ParseKeyword("LOW_PRIORITY") {
			pVal := expr.MysqlInsertPriority(1) // LowPriority
			priority = &pVal
		} else if p.ParseKeyword("DELAYED") {
			pVal := expr.MysqlInsertPriority(2) // Delayed
			priority = &pVal
		} else if p.ParseKeyword("HIGH_PRIORITY") {
			pVal := expr.MysqlInsertPriority(3) // HighPriority
			priority = &pVal
		}
	}

	// Parse optional IGNORE keyword (MySQL)
	ignore := false
	if p.dialect.SupportsInsertSet() || p.dialect.Dialect() == "mysql" || p.dialect.Dialect() == "generic" {
		ignore = p.ParseKeyword("IGNORE")
	}

	// REPLACE INTO is handled by the caller (parseReplace), so replaceInto is always false here
	replaceInto := false

	// Parse optional OVERWRITE keyword (used in some dialects like Hive)
	overwrite := p.ParseKeyword("OVERWRITE")

	// Parse optional INTO keyword
	into := p.ParseKeyword("INTO")

	// Parse optional TABLE keyword (Hive allows TABLE here)
	hasTableKeyword := p.ParseKeyword("TABLE")

	// Parse the table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse optional table alias (PostgreSQL/MySQL allows this)
	var tableAlias *ast.Ident
	var tableAliasExplicit bool
	if p.dialect.SupportsInsertTableAlias() {
		// Try to parse AS alias or just implicit alias
		if p.ParseKeyword("AS") {
			alias, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			tableAlias = alias
			tableAliasExplicit = true
		} else {
			// Try implicit alias (not followed by certain keywords)
			if !isInsertReservedKeyword(p.PeekToken()) {
				alias, err := p.ParseIdentifier()
				if err == nil {
					tableAlias = alias
					tableAliasExplicit = false
				}
			}
		}
	}

	// Check for DEFAULT VALUES
	if p.ParseKeywords([]string{"DEFAULT", "VALUES"}) {
		// Parse optional ON CONFLICT clause
		var onInsert *expr.OnInsert
		if p.ParseKeyword("ON") {
			if p.ParseKeyword("CONFLICT") {
				onInsert, err = parseOnConflict(p)
				if err != nil {
					return nil, err
				}
			}
		}

		// Parse optional RETURNING clause
		var returning []*query.SelectItem
		if p.ParseKeyword("RETURNING") {
			items, err := parseProjection(p)
			if err != nil {
				return nil, err
			}
			for i := range items {
				item := items[i]
				returning = append(returning, &item)
			}
		}

		return finishInsert(p, insertToken, optimizerHints, orConflict, priority, ignore, replaceInto,
			overwrite, into, hasTableKeyword, tableName, tableAlias, tableAliasExplicit, nil, nil, true, nil, nil, onInsert, returning)
	}

	// Parse optional column list (col1, col2, ...) - can be empty ()
	var columns []*ast.Ident
	if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
		columns, err = p.ParseParenthesizedColumnList()
		if err != nil {
			return nil, err
		}
	}

	// Parse the source: VALUES, SELECT, or SET (MySQL)
	var source *query.Query
	var assignments []*expr.Assignment

	if p.PeekKeyword("VALUES") || p.PeekKeyword("VALUE") {
		// Parse VALUES clause
		valuesStmt, err := parseValues(p)
		if err != nil {
			return nil, err
		}
		source = valuesStmt.Query
	} else if p.PeekKeyword("SELECT") || p.PeekKeyword("WITH") {
		// Parse SELECT query (with optional CTE)
		selectStmt, err := parseQuery(p)
		if err != nil {
			return nil, err
		}
		if selectStmt != nil {
			// If parseQuery returned a *statement.Query (e.g., for UNION operations),
			// extract the inner *query.Query directly
			if queryStmt, ok := selectStmt.(*statement.Query); ok {
				source = queryStmt.Query
			} else {
				// For other statement types, wrap them in a Query
				source = &query.Query{
					Body: selectStmt,
				}
			}
		}
	} else if p.dialect.SupportsInsertSet() && p.ParseKeyword("SET") {
		// MySQL INSERT ... SET syntax
		assignments, err = parseAssignments(p)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, p.Expected("VALUES, SELECT, or SET", p.PeekToken())
	}

	// Parse optional AS alias (MySQL: INSERT INTO t VALUES (1) AS alias)
	var insertAlias *expr.InsertAliases
	if p.dialect.Dialect() == "mysql" || p.dialect.Dialect() == "generic" {
		if p.ParseKeyword("AS") {
			rowAlias, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			var colAliases []*ast.Ident
			if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
				colAliases, err = p.ParseParenthesizedColumnList()
				if err != nil {
					return nil, err
				}
			}
			insertAlias = &expr.InsertAliases{
				RowAlias:   rowAlias,
				ColAliases: colAliases,
			}
		}
	}

	// Parse optional ON CONFLICT or ON DUPLICATE KEY UPDATE clause
	var onInsert *expr.OnInsert
	if p.ParseKeyword("ON") {
		if p.ParseKeyword("CONFLICT") {
			// Parse ON CONFLICT (PostgreSQL/SQLite style)
			onInsert, err = parseOnConflict(p)
			if err != nil {
				return nil, err
			}
		} else if p.ParseKeyword("DUPLICATE") {
			// Parse ON DUPLICATE KEY UPDATE (MySQL style)
			if !p.ParseKeyword("KEY") {
				return nil, p.Expected("KEY", p.PeekToken())
			}
			if !p.ParseKeyword("UPDATE") {
				return nil, p.Expected("UPDATE", p.PeekToken())
			}
			assigns, err := parseAssignments(p)
			if err != nil {
				return nil, err
			}
			onInsert = &expr.OnInsert{
				DuplicateKeyUpdate: assigns,
			}
		}
	}

	// Parse optional RETURNING clause (PostgreSQL style)
	var returning []*query.SelectItem
	if p.ParseKeyword("RETURNING") {
		items, err := parseProjection(p)
		if err != nil {
			return nil, err
		}
		for i := range items {
			item := items[i]
			returning = append(returning, &item)
		}
	}

	return finishInsert(p, insertToken, optimizerHints, orConflict, priority, ignore, replaceInto,
		overwrite, into, hasTableKeyword, tableName, tableAlias, tableAliasExplicit, columns, source, false, assignments, insertAlias, onInsert, returning)
}

// parseOnConflict parses the ON CONFLICT clause
func parseOnConflict(p *Parser) (*expr.OnInsert, error) {
	var conflictTarget *expr.ConflictTarget

	if p.ParseKeywords([]string{"ON", "CONSTRAINT"}) {
		constraintName, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		conflictTarget = &expr.ConflictTarget{
			OnConstraint: constraintName,
		}
	} else if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
		columns, err := p.ParseParenthesizedColumnList()
		if err != nil {
			return nil, err
		}
		conflictTarget = &expr.ConflictTarget{
			Columns: columns,
		}
	}

	if !p.ParseKeyword("DO") {
		return nil, p.Expected("DO", p.PeekToken())
	}

	var action expr.OnConflictAction
	if p.ParseKeyword("NOTHING") {
		action = expr.OnConflictAction{DoNothing: true}
	} else if p.ParseKeyword("UPDATE") {
		if !p.ParseKeyword("SET") {
			return nil, p.Expected("SET", p.PeekToken())
		}
		assignments, err := parseAssignments(p)
		if err != nil {
			return nil, err
		}
		var selection expr.Expr
		if p.ParseKeyword("WHERE") {
			ep := NewExpressionParser(p)
			selection, err = ep.ParseExpr()
			if err != nil {
				return nil, err
			}
		}
		action = expr.OnConflictAction{
			DoUpdate: &expr.DoUpdate{
				Assignments: assignments,
				Selection:   selection,
			},
		}
	} else {
		return nil, p.Expected("NOTHING or UPDATE", p.PeekToken())
	}

	return &expr.OnInsert{
		OnConflict: &expr.OnConflict{
			ConflictTarget: conflictTarget,
			Action:         action,
		},
	}, nil
}

// finishInsert creates the final Insert statement
func finishInsert(p *Parser, insertToken token.TokenWithSpan, optimizerHints []*expr.OptimizerHint,
	orConflict *expr.SqliteOnConflict, priority *expr.MysqlInsertPriority, ignore, replaceInto, overwrite, into,
	hasTableKeyword bool, tableName *ast.ObjectName, tableAlias *ast.Ident, tableAliasExplicit bool,
	columns []*ast.Ident, source *query.Query, defaultValues bool, assignments []*expr.Assignment,
	insertAlias *expr.InsertAliases, onInsert *expr.OnInsert, returning []*query.SelectItem) (*statement.Insert, error) {

	insert := &statement.Insert{
		OptimizerHints:     optimizerHints,
		Or:                 orConflict,
		Priority:           priority,
		Ignore:             ignore,
		ReplaceInto:        replaceInto,
		Into:               into,
		Overwrite:          overwrite,
		Table:              tableName,
		TableAlias:         tableAlias,
		TableAliasExplicit: tableAliasExplicit,
		HasTableKeyword:    hasTableKeyword,
		Columns:            columns,
		Source:             source,
		DefaultValues:      defaultValues,
		Assignments:        assignments,
		InsertAlias:        insertAlias,
		On:                 onInsert,
		Returning:          returning,
	}
	insert.SetSpan(insertToken.Span)
	return insert, nil
}

// isInsertReservedKeyword checks if a token is a reserved keyword that cannot be a table alias
func isInsertReservedKeyword(tok token.TokenWithSpan) bool {
	if word, ok := tok.Token.(token.TokenWord); ok {
		kw := word.Word.Keyword
		// Keywords that signal the end of table name/alias
		reserved := map[string]bool{
			"VALUES":    true,
			"VALUE":     true,
			"SELECT":    true,
			"DEFAULT":   true,
			"SET":       true,
			"RETURNING": true,
			"WITH":      true, // CTE keyword
		}
		return reserved[string(kw)]
	}
	return false
}

// parseUpdate parses UPDATE statements
func parseUpdate(p *Parser, tok token.TokenWithSpan) (ast.Statement, error) {
	return parseUpdateInternal(p, tok)
}

// parseUpdateInternal parses an UPDATE statement
// Basic syntax: UPDATE table SET col = val [, col2 = val2] [WHERE condition]
// MySQL syntax: UPDATE table [AS alias] [JOIN ...] SET ... WHERE ...
// PostgreSQL syntax: UPDATE table SET ... FROM other_table WHERE ...
// Snowflake/MSSQL syntax: UPDATE FROM table SET ... WHERE ...
// Reference: src/parser/mod.rs:17715
func parseUpdateInternal(p *Parser, updateToken token.TokenWithSpan) (ast.Statement, error) {
	// Parse optimizer hints if present
	hintsInterface, err := maybeParseOptimizerHints(p)
	if err != nil {
		return nil, err
	}

	// Convert optimizer hints to proper type
	var optimizerHints []*expr.OptimizerHint
	for _, h := range hintsInterface {
		if hint, ok := h.(*expr.OptimizerHint); ok {
			optimizerHints = append(optimizerHints, hint)
		}
	}

	// Parse optional OR conflict clause (SQLite style: UPDATE OR REPLACE ...)
	// Reference: src/parser/mod.rs:17657 parse_conflict_clause
	var orConflict *expr.SqliteOnConflict
	if p.ParseKeywords([]string{"OR", "REPLACE"}) {
		orVal := expr.SqliteOnConflictReplace
		orConflict = &orVal
	} else if p.ParseKeywords([]string{"OR", "ROLLBACK"}) {
		orVal := expr.SqliteOnConflictRollback
		orConflict = &orVal
	} else if p.ParseKeywords([]string{"OR", "ABORT"}) {
		orVal := expr.SqliteOnConflictAbort
		orConflict = &orVal
	} else if p.ParseKeywords([]string{"OR", "FAIL"}) {
		orVal := expr.SqliteOnConflictFail
		orConflict = &orVal
	} else if p.ParseKeywords([]string{"OR", "IGNORE"}) {
		orVal := expr.SqliteOnConflictIgnore
		orConflict = &orVal
	}

	// Parse the table reference with joins
	table, err := parseTableAndJoins(p)
	if err != nil {
		return nil, err
	}

	// Check for FROM clause BEFORE SET (Snowflake/MSSQL style: UPDATE FROM t1 SET ...)
	var fromBeforeSet *[]query.TableWithJoins
	if p.ParseKeyword("FROM") {
		fromTables, err := parseTableWithJoinsList(p)
		if err != nil {
			return nil, err
		}
		fromBeforeSet = &fromTables
	}

	// Expect SET keyword
	if _, err := p.ExpectKeyword("SET"); err != nil {
		return nil, err
	}

	// Create expression parser for parsing assignment values
	ep := NewExpressionParser(p)

	// Parse comma-separated assignments
	assignments, err := parseAssignmentsWithParser(p, ep)
	if err != nil {
		return nil, err
	}

	// Parse optional OUTPUT clause (MSSQL style)
	output, err := maybeParseOutputClause(p)
	if err != nil {
		return nil, err
	}

	// Check for FROM clause AFTER SET (PostgreSQL style: UPDATE t1 SET ... FROM t2)
	var fromAfterSet *[]query.TableWithJoins
	if fromBeforeSet == nil && p.ParseKeyword("FROM") {
		fromTables, err := parseTableWithJoinsList(p)
		if err != nil {
			return nil, err
		}
		fromAfterSet = &fromTables
	}

	// Build the From field based on which position it was found
	var fromKind *query.UpdateTableFromKind
	if fromBeforeSet != nil {
		fromKind = &query.UpdateTableFromKind{BeforeSet: fromBeforeSet}
	} else if fromAfterSet != nil {
		fromKind = &query.UpdateTableFromKind{AfterSet: fromAfterSet}
	}

	// Parse optional WHERE clause
	var selection expr.Expr
	if p.ParseKeyword("WHERE") {
		selection, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	// Note: MySQL UPDATE with ORDER BY and LIMIT is not implemented yet
	// as it requires proper type conversion between query.OrderByExpr and expr.OrderByExpr

	// Parse optional RETURNING clause (PostgreSQL style)
	var returning []*query.SelectItem
	if p.ParseKeyword("RETURNING") {
		items, err := parseProjection(p)
		if err != nil {
			return nil, err
		}
		// Convert []query.SelectItem to []*query.SelectItem
		for i := range items {
			item := items[i]
			returning = append(returning, &item)
		}
	}

	// Extract the main table name and alias from the table
	var mainTable *ast.ObjectName
	var tableAlias *ast.Ident
	if table.Relation != nil {
		if tf, ok := table.Relation.(*query.TableTableFactor); ok {
			mainTable = queryObjectNameToAst(tf.Name)
			// Extract alias if present
			if tf.Alias != nil {
				tableAlias = &ast.Ident{Value: tf.Alias.Name.Value}
			}
		}
	}

	// Extract joins from the parsed table (MySQL-style UPDATE with JOINs)
	var joins []query.Join
	if table != nil {
		joins = table.Joins
	}

	update := &statement.Update{
		Table:           mainTable,
		TableAlias:      tableAlias,
		Assignments:     assignments,
		From:            fromKind,
		Selection:       selection,
		Returning:       returning,
		Output:          output,
		IsFromStatement: false,
		Joins:           joins,
		Or:              orConflict,
	}
	update.SetSpan(updateToken.Span)

	// Store optimizer hints if the field exists (it might not yet)
	_ = optimizerHints // Avoid unused variable error if field doesn't exist yet

	return update, nil
}

// parseDelete parses DELETE statements
func parseDelete(p *Parser, tok token.TokenWithSpan) (ast.Statement, error) {
	return parseDeleteInternal(p, tok)
}

// parseDeleteInternal parses a DELETE statement
// Reference: src/parser/mod.rs:13432
func parseDeleteInternal(p *Parser, deleteToken token.TokenWithSpan) (ast.Statement, error) {
	// Parse optimizer hints if present
	hintsInterface, err := maybeParseOptimizerHints(p)
	if err != nil {
		return nil, err
	}

	// Convert optimizer hints to proper type
	var optimizerHints []*expr.OptimizerHint
	for _, h := range hintsInterface {
		if hint, ok := h.(*expr.OptimizerHint); ok {
			optimizerHints = append(optimizerHints, hint)
		}
	}

	// FROM keyword is optional in BigQuery, Oracle, and Generic dialects
	// Reference: src/parser/mod.rs:13434-13446
	var tables []*ast.ObjectName
	var withFromKeyword bool

	if !p.ParseKeyword("FROM") {
		// FROM is optional in some dialects
		dialectName := p.dialect.Dialect()
		if dialectName == "bigquery" || dialectName == "oracle" || dialectName == "generic" {
			// BigQuery style: DELETE table WHERE ... (no FROM)
			withFromKeyword = false
		} else {
			// MySQL style multi-table delete: DELETE t1, t2 FROM t1, t2 WHERE ...
			// Parse table list first
			tableName, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			tables = append(tables, tableName)

			// Check for additional tables
			for p.ConsumeToken(token.TokenComma{}) {
				t, err := p.ParseObjectName()
				if err != nil {
					return nil, err
				}
				tables = append(tables, t)
			}

			// Expect FROM after table list
			if _, err := p.ExpectKeyword("FROM"); err != nil {
				return nil, err
			}
			withFromKeyword = true
		}
	} else {
		withFromKeyword = true
	}

	// Parse FROM tables (if not already parsed as multi-table list)
	var fromTables []query.TableWithJoins
	if len(tables) == 0 || withFromKeyword {
		from, err := parseTableWithJoinsList(p)
		if err != nil {
			return nil, err
		}
		fromTables = from
	}

	// If no tables were parsed (single-table delete), extract from FROM clause
	if len(tables) == 0 && len(fromTables) > 0 {
		for _, twj := range fromTables {
			if tf, ok := twj.Relation.(*query.TableTableFactor); ok {
				objName := queryObjectNameToAst(tf.Name)
				tables = append(tables, objName)
			}
		}
	}

	// Parse optional OUTPUT clause (SQL Server style)
	output, err := maybeParseOutputClause(p)
	if err != nil {
		return nil, err
	}

	// Parse optional USING clause (PostgreSQL style: DELETE FROM t USING t2 WHERE ...)
	var using []query.TableWithJoins
	if p.ParseKeyword("USING") {
		using, err = parseTableWithJoinsList(p)
		if err != nil {
			return nil, err
		}
	}

	// Create expression parser for parsing WHERE conditions
	ep := NewExpressionParser(p)

	// Parse optional WHERE clause
	var selection expr.Expr
	if p.ParseKeyword("WHERE") {
		selection, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	// Parse optional RETURNING clause (PostgreSQL style)
	var returning []*query.SelectItem
	if p.ParseKeyword("RETURNING") {
		items, err := parseProjection(p)
		if err != nil {
			return nil, err
		}
		for i := range items {
			item := items[i]
			returning = append(returning, &item)
		}
	}

	// Parse optional ORDER BY clause (MySQL)
	var orderBy []query.OrderByExpr
	if p.ParseKeyword("ORDER") {
		if !p.ParseKeyword("BY") {
			return nil, p.Expected("BY", p.PeekToken())
		}
		orderByExprs, err := parseOrderByExpressions(p)
		if err != nil {
			return nil, err
		}
		orderBy = orderByExprs
	}

	// Parse optional LIMIT clause (MySQL)
	var limit query.LimitClause
	if p.ParseKeyword("LIMIT") {
		firstExpr, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}

		if p.ConsumeToken(token.TokenComma{}) {
			// MySQL style: LIMIT offset, limit
			secondExpr, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			limit = &query.LimitOffset{
				Limit:  &queryExprWrapper{expr: secondExpr},
				Offset: &query.Offset{Value: &queryExprWrapper{expr: firstExpr}},
			}
		} else {
			// Standard LIMIT expr
			limit = &query.LimitOffset{
				Limit: &queryExprWrapper{expr: firstExpr},
			}
		}
	}

	// Convert using from []query.TableWithJoins to []*query.TableWithJoins
	var usingPtrs []*query.TableWithJoins
	for i := range using {
		usingPtrs = append(usingPtrs, &using[i])
	}

	// Build the Delete statement
	// Note: The Go AST doesn't have a separate FromTable type like Rust,
	// so we store the tables directly in the Tables field
	delete := &statement.Delete{
		Tables:    tables,
		Using:     usingPtrs,
		Selection: selection,
		Returning: returning,
		OrderBy:   orderBy,
		Limit:     limit,
		Output:    output,
	}
	delete.SetSpan(deleteToken.Span)

	// Store optimizer hints - they may need to be added to the AST struct
	_ = optimizerHints

	return delete, nil
}

// parseAssignments parses a comma-separated list of assignments (col = val)
// Used for ON CONFLICT DO UPDATE SET ...
func parseAssignments(p *Parser) ([]*expr.Assignment, error) {
	var assignments []*expr.Assignment

	err := p.ParseCommaSeparated(func() error {
		// Parse column identifier
		col, err := p.ParseIdentifier()
		if err != nil {
			return err
		}

		// Expect = token
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return err
		}

		// Parse expression using ExpressionParser to get expr.Expr
		ep := NewExpressionParser(p)
		val, err := ep.ParseExpr()
		if err != nil {
			return err
		}

		assignment := &expr.Assignment{
			Column: col,
			Value:  val,
		}
		assignments = append(assignments, assignment)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return assignments, nil
}

// parseAssignmentsWithParser parses a comma-separated list of assignments (col = val)
// using an existing ExpressionParser
func parseAssignmentsWithParser(p *Parser, ep *ExpressionParser) ([]*expr.Assignment, error) {
	var assignments []*expr.Assignment

	err := p.ParseCommaSeparated(func() error {
		// Parse column identifier (possibly compound like "t.col")
		col, err := p.ParseIdentifier()
		if err != nil {
			return err
		}

		// Check for compound identifier (table.column)
		for {
			if !p.ConsumeToken(token.TokenPeriod{}) {
				break
			}
			nextIdent, err := p.ParseIdentifier()
			if err != nil {
				return err
			}
			// Create a compound identifier
			col = &ast.Ident{
				Value: col.Value + "." + nextIdent.Value,
			}
		}

		// Expect = token
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return err
		}

		// Parse expression using ExpressionParser to get expr.Expr
		val, err := ep.ParseExpr()
		if err != nil {
			return err
		}

		assignment := &expr.Assignment{
			Column: col,
			Value:  val,
		}
		assignments = append(assignments, assignment)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return assignments, nil
}

// maybeParseOutputClause tries to parse an OUTPUT clause if present
// Returns nil if no OUTPUT clause is found
// Reference: src/parser/merge.rs:235
func maybeParseOutputClause(p *Parser) (*expr.OutputClause, error) {
	tok := p.PeekToken()
	word, ok := tok.Token.(token.TokenWord)
	if !ok {
		return nil, nil
	}

	if string(word.Word.Keyword) == "OUTPUT" {
		p.AdvanceToken()                          // consume OUTPUT
		return parseOutputClauseInternal(p, true) // true = is OUTPUT clause
	}

	return nil, nil
}
