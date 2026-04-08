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

package dialects

// ============================================================================
// SELECT Clause Capability Helpers
// ============================================================================

// SupportsSelectWildcardExcept returns true if the dialect supports EXCEPT in SELECT *.
func SupportsSelectWildcardExcept(d CoreDialect) bool {
	if sd, ok := d.(SelectDialect); ok {
		return sd.SupportsSelectWildcardExcept()
	}
	return false
}

// SupportsSelectWildcardExclude returns true if the dialect supports EXCLUDE
// option following a wildcard in a projection section.
func SupportsSelectWildcardExclude(d CoreDialect) bool {
	if sd, ok := d.(SelectDialect); ok {
		return sd.SupportsSelectWildcardExclude()
	}
	return false
}

// SupportsSelectWildcardReplace returns true if the dialect supports REPLACE
// option in SELECT * wildcard expressions.
func SupportsSelectWildcardReplace(d CoreDialect) bool {
	if sd, ok := d.(SelectDialect); ok {
		return sd.SupportsSelectWildcardReplace()
	}
	return false
}

// SupportsSelectWildcardRename returns true if the dialect supports RENAME
// option in SELECT * wildcard expressions.
func SupportsSelectWildcardRename(d CoreDialect) bool {
	if sd, ok := d.(SelectDialect); ok {
		return sd.SupportsSelectWildcardRename()
	}
	return false
}

// SupportsSelectWildcardIlike returns true if the dialect supports ILIKE
// option in SELECT * wildcard expressions.
func SupportsSelectWildcardIlike(d CoreDialect) bool {
	if sd, ok := d.(SelectDialect); ok {
		return sd.SupportsSelectWildcardIlike()
	}
	return false
}

// SupportsSelectExprStar returns true if the dialect supports wildcard expansion
// on arbitrary expressions like IDENTIFIER('name').* (Snowflake-specific).
func SupportsSelectExprStar(d CoreDialect) bool {
	if sd, ok := d.(SelectDialect); ok {
		return sd.SupportsSelectExprStar()
	}
	return false
}

// SupportsTrailingCommas returns true if the dialect supports trailing commas.
func SupportsTrailingCommas(d CoreDialect) bool {
	if sd, ok := d.(SelectDialect); ok {
		return sd.SupportsTrailingCommas()
	}
	return false
}

// SupportsProjectionTrailingCommas returns true if the dialect supports trailing
// commas in the projection list.
func SupportsProjectionTrailingCommas(d CoreDialect) bool {
	if sd, ok := d.(SelectDialect); ok {
		return sd.SupportsProjectionTrailingCommas()
	}
	return false
}

// SupportsLimitComma returns true if the dialect supports parsing LIMIT 1, 2
// as LIMIT 2 OFFSET 1.
func SupportsLimitComma(d CoreDialect) bool {
	if sd, ok := d.(SelectDialect); ok {
		return sd.SupportsLimitComma()
	}
	return false
}

// SupportsFromFirstSelect returns true if the dialect supports "FROM-first" selects.
func SupportsFromFirstSelect(d CoreDialect) bool {
	if sd, ok := d.(SelectDialect); ok {
		return sd.SupportsFromFirstSelect()
	}
	return false
}

// SupportsPipeOperator returns true if the dialect supports the pipe operator (|>).
func SupportsPipeOperator(d CoreDialect) bool {
	if sd, ok := d.(SelectDialect); ok {
		return sd.SupportsPipeOperator()
	}
	return false
}

// ============================================================================
// JOIN Clause Capability Helpers
// ============================================================================

// SupportsOuterJoinOperator returns true if the dialect supports Oracle-style (+) join.
func SupportsOuterJoinOperator(d CoreDialect) bool {
	if jd, ok := d.(JoinDialect); ok {
		return jd.SupportsOuterJoinOperator()
	}
	return false
}

// SupportsLeftAssociativeJoinsWithoutParens returns true if the dialect supports
// left-associative join parsing by default when parentheses are omitted in nested joins.
func SupportsLeftAssociativeJoinsWithoutParens(d CoreDialect) bool {
	if jd, ok := d.(JoinDialect); ok {
		return jd.SupportsLeftAssociativeJoinsWithoutParens()
	}
	return false
}

// SupportsCrossJoinConstraint returns true if the dialect supports a join
// specification on CROSS JOIN.
func SupportsCrossJoinConstraint(d CoreDialect) bool {
	if jd, ok := d.(JoinDialect); ok {
		return jd.SupportsCrossJoinConstraint()
	}
	return false
}

// ============================================================================
// String Literal Capability Helpers
// ============================================================================

// SupportsStringLiteralBackslashEscape returns true if the dialect supports
// escaping characters via '\' in string literals.
func SupportsStringLiteralBackslashEscape(d CoreDialect) bool {
	if sld, ok := d.(StringLiteralDialect); ok {
		return sld.SupportsStringLiteralBackslashEscape()
	}
	return false
}

// SupportsUnicodeStringLiteral returns true if the dialect supports string
// literals with U& prefix for Unicode code points.
func SupportsUnicodeStringLiteral(d CoreDialect) bool {
	if sld, ok := d.(StringLiteralDialect); ok {
		return sld.SupportsUnicodeStringLiteral()
	}
	return false
}

// SupportsTripleQuotedString returns true if the dialect supports triple
// quoted string literals (e.g., """abc""").
func SupportsTripleQuotedString(d CoreDialect) bool {
	if sld, ok := d.(StringLiteralDialect); ok {
		return sld.SupportsTripleQuotedString()
	}
	return false
}

// SupportsStringEscapeConstant returns true if the dialect supports the
// E'...' syntax for string literals with escape sequences (Postgres).
func SupportsStringEscapeConstant(d CoreDialect) bool {
	if sld, ok := d.(StringLiteralDialect); ok {
		return sld.SupportsStringEscapeConstant()
	}
	return false
}

// SupportsStringLiteralConcatenation returns true if the dialect supports
// concatenating adjacent string literals (e.g., 'Hello ' 'world').
func SupportsStringLiteralConcatenation(d CoreDialect) bool {
	if sld, ok := d.(StringLiteralDialect); ok {
		return sld.SupportsStringLiteralConcatenation()
	}
	return false
}

// SupportsStringLiteralConcatenationWithNewline returns true if the dialect
// supports concatenating string literals separated by newlines.
func SupportsStringLiteralConcatenationWithNewline(d CoreDialect) bool {
	if sld, ok := d.(StringLiteralDialect); ok {
		return sld.SupportsStringLiteralConcatenationWithNewline()
	}
	return false
}

// ============================================================================
// Aggregation and Window Function Capability Helpers
// ============================================================================

// SupportsFilterDuringAggregation returns true if the dialect supports
// FILTER (WHERE expr) for aggregate queries.
func SupportsFilterDuringAggregation(d CoreDialect) bool {
	if ad, ok := d.(AggregationDialect); ok {
		return ad.SupportsFilterDuringAggregation()
	}
	return false
}

// SupportsWithinAfterArrayAggregation returns true if the dialect supports
// the WITHIN GROUP clause for array aggregations.
func SupportsWithinAfterArrayAggregation(d CoreDialect) bool {
	if ad, ok := d.(AggregationDialect); ok {
		return ad.SupportsWithinAfterArrayAggregation()
	}
	return false
}

// SupportsMatchRecognize returns true if the dialect supports the MATCH_RECOGNIZE operation.
func SupportsMatchRecognize(d CoreDialect) bool {
	if ad, ok := d.(AggregationDialect); ok {
		return ad.SupportsMatchRecognize()
	}
	return false
}

// SupportsWindowClauseNamedWindowReference returns true if the dialect
// supports referencing another named window within a window clause declaration.
func SupportsWindowClauseNamedWindowReference(d CoreDialect) bool {
	if ad, ok := d.(AggregationDialect); ok {
		return ad.SupportsWindowClauseNamedWindowReference()
	}
	return false
}

// ============================================================================
// GROUP BY Clause Capability Helpers
// ============================================================================

// SupportsGroupByExpr returns true if the dialect supports GROUP BY expressions
// like GROUPING SETS, ROLLUP, or CUBE.
func SupportsGroupByExpr(d CoreDialect) bool {
	if gbd, ok := d.(GroupByDialect); ok {
		return gbd.SupportsGroupByExpr()
	}
	return false
}

// SupportsGroupByWithModifier returns true if the dialect supports GROUP BY
// modifiers prefixed by a WITH keyword (e.g., GROUP BY value WITH ROLLUP).
func SupportsGroupByWithModifier(d CoreDialect) bool {
	if gbd, ok := d.(GroupByDialect); ok {
		return gbd.SupportsGroupByWithModifier()
	}
	return false
}

// ============================================================================
// Named Argument Capability Helpers
// ============================================================================

// SupportsNamedFnArgsWithRArrowOperator returns true if the dialect supports
// named arguments of the form FUN(a => '1', b => '2').
func SupportsNamedFnArgsWithRArrowOperator(d CoreDialect) bool {
	if nad, ok := d.(NamedArgumentDialect); ok {
		return nad.SupportsNamedFnArgsWithRArrowOperator()
	}
	return false
}

// SupportsNamedFnArgsWithEqOperator returns true if the dialect supports
// named arguments of the form FUN(a = '1', b = '2').
func SupportsNamedFnArgsWithEqOperator(d CoreDialect) bool {
	if nad, ok := d.(NamedArgumentDialect); ok {
		return nad.SupportsNamedFnArgsWithEqOperator()
	}
	return false
}

// SupportsNamedFnArgsWithColonOperator returns true if the dialect supports
// named arguments of the form FUN(a : '1', b : '2').
func SupportsNamedFnArgsWithColonOperator(d CoreDialect) bool {
	if nad, ok := d.(NamedArgumentDialect); ok {
		return nad.SupportsNamedFnArgsWithColonOperator()
	}
	return false
}

// SupportsNamedFnArgsWithAssignmentOperator returns true if the dialect supports
// named arguments of the form FUN(a := '1', b := '2').
func SupportsNamedFnArgsWithAssignmentOperator(d CoreDialect) bool {
	if nad, ok := d.(NamedArgumentDialect); ok {
		return nad.SupportsNamedFnArgsWithAssignmentOperator()
	}
	return false
}

// SupportsNamedFnArgsWithExprName returns true if the dialect supports
// named arguments with expression names like FUN(expr = value).
func SupportsNamedFnArgsWithExprName(d CoreDialect) bool {
	if nad, ok := d.(NamedArgumentDialect); ok {
		return nad.SupportsNamedFnArgsWithExprName()
	}
	return false
}

// ============================================================================
// Comment Capability Helpers
// ============================================================================

// SupportsNestedComments returns true if the dialect supports nested comments.
func SupportsNestedComments(d CoreDialect) bool {
	if cd, ok := d.(CommentDialect); ok {
		return cd.SupportsNestedComments()
	}
	return false
}

// SupportsMultilineCommentHints returns true if the dialect supports optimizer
// hints in multiline comments (e.g., /*!50110 KEY_BLOCK_SIZE = 1024*/).
func SupportsMultilineCommentHints(d CoreDialect) bool {
	if cd, ok := d.(CommentDialect); ok {
		return cd.SupportsMultilineCommentHints()
	}
	return false
}

// SupportsCommentOn returns true if the dialect supports the COMMENT statement.
func SupportsCommentOn(d CoreDialect) bool {
	if cd, ok := d.(CommentDialect); ok {
		return cd.SupportsCommentOn()
	}
	return false
}

// ============================================================================
// Table Definition Capability Helpers
// ============================================================================

// SupportsCreateTableSelect returns true if the dialect supports CREATE TABLE SELECT.
func SupportsCreateTableSelect(d CoreDialect) bool {
	if tdd, ok := d.(TableDefinitionDialect); ok {
		return tdd.SupportsCreateTableSelect()
	}
	return false
}

// SupportsValuesAsTableFactor returns true if the dialect supports VALUES
// as a table factor without requiring parentheses.
func SupportsValuesAsTableFactor(d CoreDialect) bool {
	if tdd, ok := d.(TableDefinitionDialect); ok {
		return tdd.SupportsValuesAsTableFactor()
	}
	return false
}

// SupportsUnnestTableFactor returns true if the dialect supports UNNEST
// as a table factor (BigQuery/PostgreSQL).
func SupportsUnnestTableFactor(d CoreDialect) bool {
	if tdd, ok := d.(TableDefinitionDialect); ok {
		return tdd.SupportsUnnestTableFactor()
	}
	return false
}

// SupportsTableHints returns true if the dialect supports table hints in the FROM clause.
func SupportsTableHints(d CoreDialect) bool {
	if tdd, ok := d.(TableDefinitionDialect); ok {
		return tdd.SupportsTableHints()
	}
	return false
}

// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
func SupportsIndexHints(d CoreDialect) bool {
	if tdd, ok := d.(TableDefinitionDialect); ok {
		return tdd.SupportsIndexHints()
	}
	return false
}

// ============================================================================
// Literal and Expression Capability Helpers
// ============================================================================

// SupportsDictionarySyntax returns true if the dialect supports defining
// structs or objects using syntax like {'x': 1, 'y': 2, 'z': 3}.
func SupportsDictionarySyntax(d CoreDialect) bool {
	if ld, ok := d.(LiteralDialect); ok {
		return ld.SupportsDictionarySyntax()
	}
	return false
}

// SupportsLambdaFunctions returns true if the dialect supports lambda functions.
func SupportsLambdaFunctions(d CoreDialect) bool {
	if ld, ok := d.(LiteralDialect); ok {
		return ld.SupportsLambdaFunctions()
	}
	return false
}

// SupportsInEmptyList returns true if the dialect supports (NOT) IN ()
// expressions with empty lists.
func SupportsInEmptyList(d CoreDialect) bool {
	if ied, ok := d.(InExpressionDialect); ok {
		return ied.SupportsInEmptyList()
	}
	return false
}

// SupportsSubqueryAsFunctionArg returns true if the dialect supports a subquery
// passed to a function as the only argument without enclosing parentheses.
func SupportsSubqueryAsFunctionArg(d CoreDialect) bool {
	if sqd, ok := d.(SubqueryDialect); ok {
		return sqd.SupportsSubqueryAsFunctionArg()
	}
	return false
}

// ============================================================================
// Operator Capability Helpers
// ============================================================================

// SupportsFactorialOperator returns true if the dialect supports a!
// expressions for factorial.
func SupportsFactorialOperator(d CoreDialect) bool {
	if od, ok := d.(OperatorDialect); ok {
		return od.SupportsFactorialOperator()
	}
	return false
}

// SupportsBitwiseShiftOperators returns true if the dialect supports
// << and >> shift operators.
func SupportsBitwiseShiftOperators(d CoreDialect) bool {
	if od, ok := d.(OperatorDialect); ok {
		return od.SupportsBitwiseShiftOperators()
	}
	return false
}

// SupportsBangNotOperator returns true if the dialect supports !a syntax
// for boolean NOT expressions.
func SupportsBangNotOperator(d CoreDialect) bool {
	if od, ok := d.(OperatorDialect); ok {
		return od.SupportsBangNotOperator()
	}
	return false
}

// SupportsNotnullOperator returns true if the dialect supports the x NOTNULL
// operator expression.
func SupportsNotnullOperator(d CoreDialect) bool {
	if od, ok := d.(OperatorDialect); ok {
		return od.SupportsNotnullOperator()
	}
	return false
}

// SupportsDoubleAmpersandOperator returns true if the dialect considers
// the && operator as a boolean AND operator.
func SupportsDoubleAmpersandOperator(d CoreDialect) bool {
	if od, ok := d.(OperatorDialect); ok {
		return od.SupportsDoubleAmpersandOperator()
	}
	return false
}

// ============================================================================
// Special Dialect Capability Helpers
// ============================================================================

// SupportsConnectBy returns true if the dialect supports CONNECT BY for
// hierarchical queries (Oracle-style).
func SupportsConnectBy(d CoreDialect) bool {
	if cbd, ok := d.(ConnectByDialect); ok {
		return cbd.SupportsConnectBy()
	}
	return false
}

// SupportsPartiQL returns true if the dialect supports PartiQL for querying
// semi-structured data.
func SupportsPartiQL(d CoreDialect) bool {
	if pqd, ok := d.(PartiQLDialect); ok {
		return pqd.SupportsPartiQL()
	}
	return false
}

// SupportsGeometricTypes returns true if the dialect supports geometric types
// (Postgres geometric operations).
func SupportsGeometricTypes(d CoreDialect) bool {
	if gd, ok := d.(GeometricDialect); ok {
		return gd.SupportsGeometricTypes()
	}
	return false
}

// SupportsLoadData returns true if the dialect supports LOAD DATA statement.
func SupportsLoadData(d CoreDialect) bool {
	if ld, ok := d.(LoadDialect); ok {
		return ld.SupportsLoadData()
	}
	return false
}

// SupportsMatchAgainst returns true if the dialect supports MATCH() AGAINST() syntax.
func SupportsMatchAgainst(d CoreDialect) bool {
	if md, ok := d.(MatchDialect); ok {
		return md.SupportsMatchAgainst()
	}
	return false
}

// SupportsExplainWithUtilityOptions returns true if the dialect supports
// EXPLAIN statements with utility options.
func SupportsExplainWithUtilityOptions(d CoreDialect) bool {
	if ed, ok := d.(ExplainDialect); ok {
		return ed.SupportsExplainWithUtilityOptions()
	}
	return false
}

// SupportsExecuteImmediate returns true if the dialect supports EXECUTE IMMEDIATE statements.
func SupportsExecuteImmediate(d CoreDialect) bool {
	if ed, ok := d.(ExecuteDialect); ok {
		return ed.SupportsExecuteImmediate()
	}
	return false
}

// SupportsListenNotify returns true if the dialect supports LISTEN, UNLISTEN,
// and NOTIFY statements.
func SupportsListenNotify(d CoreDialect) bool {
	if lnd, ok := d.(ListenNotifyDialect); ok {
		return lnd.SupportsListenNotify()
	}
	return false
}

// ============================================================================
// ClickHouse-Specific Capability Helpers
// ============================================================================

// SupportsPrewhere returns true if the dialect supports PREWHERE clause.
func SupportsPrewhere(d CoreDialect) bool {
	if chd, ok := d.(ClickHouseDialect); ok {
		return chd.SupportsPrewhere()
	}
	return false
}

// SupportsLimitBy returns true if the dialect supports LIMIT BY clause.
func SupportsLimitBy(d CoreDialect) bool {
	if chd, ok := d.(ClickHouseDialect); ok {
		return chd.SupportsLimitBy()
	}
	return false
}

// SupportsSettings returns true if the dialect supports SETTINGS clause.
func SupportsSettings(d CoreDialect) bool {
	if chd, ok := d.(ClickHouseDialect); ok {
		return chd.SupportsSettings()
	}
	return false
}

// SupportsOptimizeTable returns true if the dialect supports OPTIMIZE TABLE.
func SupportsOptimizeTable(d CoreDialect) bool {
	if chd, ok := d.(ClickHouseDialect); ok {
		return chd.SupportsOptimizeTable()
	}
	return false
}

// ============================================================================
// DuckDB-Specific Capability Helpers
// ============================================================================

// SupportsInstall returns true if the dialect supports INSTALL statement.
func SupportsInstall(d CoreDialect) bool {
	if ddd, ok := d.(DuckDBDialect); ok {
		return ddd.SupportsInstall()
	}
	return false
}

// SupportsDetach returns true if the dialect supports DETACH statement.
func SupportsDetach(d CoreDialect) bool {
	if ddd, ok := d.(DuckDBDialect); ok {
		return ddd.SupportsDetach()
	}
	return false
}

// ============================================================================
// INSERT Statement Capability Helpers
// ============================================================================

// SupportsInsertSet returns true if the dialect supports INSERT INTO ... SET syntax.
func SupportsInsertSet(d CoreDialect) bool {
	if id, ok := d.(InsertDialect); ok {
		return id.SupportsInsertSet()
	}
	return false
}

// SupportsInsertFormat returns true if the dialect supports insert formats
// (e.g., INSERT INTO ... FORMAT <format>).
func SupportsInsertFormat(d CoreDialect) bool {
	if id, ok := d.(InsertDialect); ok {
		return id.SupportsInsertFormat()
	}
	return false
}

// SupportsInsertTableAlias returns true if the dialect supports INSERT INTO
// table [[AS] alias] syntax.
func SupportsInsertTableAlias(d CoreDialect) bool {
	if id, ok := d.(InsertDialect); ok {
		return id.SupportsInsertTableAlias()
	}
	return false
}

// ============================================================================
// ALTER TABLE Capability Helpers
// ============================================================================

// SupportsAlterColumnTypeUsing returns true if the dialect supports the USING
// clause in an ALTER COLUMN statement.
func SupportsAlterColumnTypeUsing(d CoreDialect) bool {
	if atd, ok := d.(AlterTableDialect); ok {
		return atd.SupportsAlterColumnTypeUsing()
	}
	return false
}

// SupportsCommaSeparatedDropColumnList returns true if the dialect supports
// ALTER TABLE tbl DROP COLUMN c1, ..., cn.
func SupportsCommaSeparatedDropColumnList(d CoreDialect) bool {
	if atd, ok := d.(AlterTableDialect); ok {
		return atd.SupportsCommaSeparatedDropColumnList()
	}
	return false
}

// ============================================================================
// SET Statement Capability Helpers
// ============================================================================

// SupportsCommaSeparatedSetAssignments returns true if the dialect supports
// multiple SET statements in a single statement.
func SupportsCommaSeparatedSetAssignments(d CoreDialect) bool {
	if sd, ok := d.(SetDialect); ok {
		return sd.SupportsCommaSeparatedSetAssignments()
	}
	return false
}

// SupportsSetNames returns true if the dialect supports SET NAMES <charset_name> [COLLATE <collation_name>].
func SupportsSetNames(d CoreDialect) bool {
	if sd, ok := d.(SetDialect); ok {
		return sd.SupportsSetNames()
	}
	return false
}

// SupportsParenthesizedSetVariables returns true if the dialect supports
// multiple variable assignment using parentheses in a SET variable declaration.
func SupportsParenthesizedSetVariables(d CoreDialect) bool {
	if sd, ok := d.(SetDialect); ok {
		return sd.SupportsParenthesizedSetVariables()
	}
	return false
}

// ============================================================================
// Transaction Capability Helpers
// ============================================================================

// SupportsStartTransactionModifier returns true if the dialect supports
// BEGIN {DEFERRED | IMMEDIATE | EXCLUSIVE | TRY | CATCH} [TRANSACTION].
func SupportsStartTransactionModifier(d CoreDialect) bool {
	if td, ok := d.(TransactionDialect); ok {
		return td.SupportsStartTransactionModifier()
	}
	return false
}

// ============================================================================
// Type Conversion Capability Helpers
// ============================================================================

// ConvertTypeBeforeValue returns true if the dialect has a CONVERT function
// which accepts a type first and an expression second.
func ConvertTypeBeforeValue(d CoreDialect) bool {
	if tcd, ok := d.(TypeConversionDialect); ok {
		return tcd.ConvertTypeBeforeValue()
	}
	return false
}

// SupportsTryConvert returns true if the dialect supports the TRY_CONVERT function.
func SupportsTryConvert(d CoreDialect) bool {
	if tcd, ok := d.(TypeConversionDialect); ok {
		return tcd.SupportsTryConvert()
	}
	return false
}

// ============================================================================
// Interval Capability Helpers
// ============================================================================

// RequireIntervalQualifier returns true if the dialect requires units
// (qualifiers) to be specified in INTERVAL expressions.
func RequireIntervalQualifier(d CoreDialect) bool {
	if id, ok := d.(IntervalDialect); ok {
		return id.RequireIntervalQualifier()
	}
	return false
}

// ============================================================================
// Extract Capability Helpers
// ============================================================================

// SupportsExtractCommaSyntax returns true if the dialect supports EXTRACT
// function with a comma separator instead of FROM.
func SupportsExtractCommaSyntax(d CoreDialect) bool {
	if ed, ok := d.(ExtractDialect); ok {
		return ed.SupportsExtractCommaSyntax()
	}
	return false
}

// ============================================================================
// Trim Capability Helpers
// ============================================================================

// SupportsCommaSeparatedTrim returns true if the dialect supports the two-argument
// comma-separated form of TRIM function: TRIM(expr, characters).
func SupportsCommaSeparatedTrim(d CoreDialect) bool {
	if td, ok := d.(TrimDialect); ok {
		return td.SupportsCommaSeparatedTrim()
	}
	return false
}

// ============================================================================
// Literal Capability Helpers
// ============================================================================

// SupportsMapLiteralSyntax returns true if the dialect supports defining
// objects using syntax like Map {1: 10, 2: 20}.
func SupportsMapLiteralSyntax(d CoreDialect) bool {
	if ld, ok := d.(LiteralDialect); ok {
		return ld.SupportsMapLiteralSyntax()
	}
	return false
}

// SupportsStructLiteral returns true if the dialect supports STRUCT literal syntax.
func SupportsStructLiteral(d CoreDialect) bool {
	if ld, ok := d.(LiteralDialect); ok {
		return ld.SupportsStructLiteral()
	}
	return false
}

// ============================================================================
// Object Reference Capability Helpers
// ============================================================================

// SupportsObjectNameDoubleDotNotation returns true if the dialect supports
// double dot notation for object names (e.g., db_name..table_name).
func SupportsObjectNameDoubleDotNotation(d CoreDialect) bool {
	if ord, ok := d.(ObjectReferenceDialect); ok {
		return ord.SupportsObjectNameDoubleDotNotation()
	}
	return false
}

// SupportsNumericLiteralUnderscores returns true if the dialect supports
// numbers containing underscores (e.g., 10_000_000).
func SupportsNumericLiteralUnderscores(d CoreDialect) bool {
	if ord, ok := d.(ObjectReferenceDialect); ok {
		return ord.SupportsNumericLiteralUnderscores()
	}
	return false
}

// ============================================================================
// ORDER BY Capability Helpers
// ============================================================================

// SupportsOrderByAll returns true if the dialect supports ORDER BY ALL.
func SupportsOrderByAll(d CoreDialect) bool {
	if obd, ok := d.(OrderByDialect); ok {
		return obd.SupportsOrderByAll()
	}
	return false
}

// ============================================================================
// Boolean Literal Capability Helpers
// ============================================================================

// SupportsBooleanLiterals returns true if the dialect supports boolean
// literals (true and false).
func SupportsBooleanLiterals(d CoreDialect) bool {
	if bld, ok := d.(BooleanLiteralDialect); ok {
		return bld.SupportsBooleanLiterals()
	}
	return false
}

// ============================================================================
// Alias Capability Helpers
// ============================================================================

// SupportsEqAliasAssignment returns true if the dialect supports treating
// the equals operator = within a SelectItem as an alias assignment operator.
func SupportsEqAliasAssignment(d CoreDialect) bool {
	if ad, ok := d.(AliasDialect); ok {
		return ad.SupportsEqAliasAssignment()
	}
	return false
}
