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

package duckdb

import (
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/token"
)

// DuckDbDialect is a dialect for DuckDB SQL implementation.
// See https://duckdb.org/
type DuckDbDialect struct{}

// NewDuckDbDialect creates a new instance of DuckDbDialect.
func NewDuckDbDialect() *DuckDbDialect {
	return &DuckDbDialect{}
}

// Dialect returns the dialect identifier.
func (d *DuckDbDialect) Dialect() string {
	return "duckdb"
}

// IsIdentifierStart returns true if the character is a valid start character
// for an unquoted identifier.
func (d *DuckDbDialect) IsIdentifierStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// IsIdentifierPart returns true if the character is a valid unquoted
// identifier character (not necessarily at the start).
func (d *DuckDbDialect) IsIdentifierPart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '$' || ch == '_'
}

// IsDelimitedIdentifierStart returns true if the character starts a quoted identifier.
func (d *DuckDbDialect) IsDelimitedIdentifierStart(ch rune) bool {
	return ch == '"'
}

// IsNestedDelimitedIdentifierStart returns true if the character starts a
// potential nested quoted identifier.
func (d *DuckDbDialect) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return false
}

// PeekNestedDelimitedIdentifierQuotes returns the outer quote style and
// optional inner quote style if applicable.
func (d *DuckDbDialect) PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*dialects.NestedIdentifierQuote, int) {
	return nil, pos
}

// IdentifierQuoteStyle returns the character used to quote identifiers.
func (d *DuckDbDialect) IdentifierQuoteStyle(identifier string) *rune {
	return nil
}

// IsCustomOperatorPart returns true if the character is part of a custom operator.
func (d *DuckDbDialect) IsCustomOperatorPart(ch rune) bool {
	return false
}

// IsReservedForIdentifier returns true if the keyword is reserved.
func (d *DuckDbDialect) IsReservedForIdentifier(kw token.Keyword) bool {
	return token.IsReservedForIdentifier(kw)
}

// GetReservedKeywordsForSelectItemOperator returns reserved keywords for select items.
func (d *DuckDbDialect) GetReservedKeywordsForSelectItemOperator() []token.Keyword {
	return nil
}

// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
func (d *DuckDbDialect) GetReservedGranteesTypes() []dialects.GranteesType {
	return nil
}

// IsColumnAlias returns true if the keyword should be parsed as a column alias.
func (d *DuckDbDialect) IsColumnAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == "AS" || token.IsReservedForColumnAlias(kw)
}

// IsSelectItemAlias returns true if the specified keyword should be parsed as a select item alias.
func (d *DuckDbDialect) IsSelectItemAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsColumnAlias(kw, parser)
}

// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier.
func (d *DuckDbDialect) IsTableFactor(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == "TABLE" || kw == "LATERAL"
}

// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias.
func (d *DuckDbDialect) IsTableAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return !token.IsReservedForTableAlias(kw)
}

// IsTableFactorAlias returns true if the specified keyword should be parsed as a table factor alias.
func (d *DuckDbDialect) IsTableFactorAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return explicit || d.IsTableAlias(kw, parser)
}

// IsIdentifierGeneratingFunctionName returns true if the ident is a function that generates identifiers.
func (d *DuckDbDialect) IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool {
	return false
}

// SupportsStringLiteralBackslashEscape returns true if the dialect supports
// escaping characters via '\' in string literals.
func (d *DuckDbDialect) SupportsStringLiteralBackslashEscape() bool {
	return true
}

// IgnoresWildcardEscapes returns true if the dialect strips the backslash
// when escaping LIKE wildcards (%, _).
func (d *DuckDbDialect) IgnoresWildcardEscapes() bool {
	return false
}

// SupportsUnicodeStringLiteral returns true if the dialect supports string
// literals with U& prefix for Unicode code points.
func (d *DuckDbDialect) SupportsUnicodeStringLiteral() bool {
	return true
}

// SupportsTripleQuotedString returns true if the dialect supports triple
// quoted string literals (e.g., """abc""").
func (d *DuckDbDialect) SupportsTripleQuotedString() bool {
	return false
}

// SupportsStringLiteralConcatenation returns true if the dialect supports
// concatenating string literals (e.g., 'Hello ' 'world').
func (d *DuckDbDialect) SupportsStringLiteralConcatenation() bool {
	return false
}

// SupportsStringLiteralConcatenationWithNewline returns true if the dialect
// supports concatenating string literals with a newline between them.
func (d *DuckDbDialect) SupportsStringLiteralConcatenationWithNewline() bool {
	return false
}

// SupportsQuoteDelimitedString returns true if the dialect supports
// quote-delimited string literals (e.g., Q'{...}') for Oracle-style strings.
func (d *DuckDbDialect) SupportsQuoteDelimitedString() bool {
	return false
}

// SupportsStringEscapeConstant returns true if the dialect supports the
// E'...' syntax for string literals with escape sequences (Postgres).
func (d *DuckDbDialect) SupportsStringEscapeConstant() bool {
	return true
}

// SupportsFilterDuringAggregation returns true if the dialect supports
// FILTER (WHERE expr) for aggregate queries.
func (d *DuckDbDialect) SupportsFilterDuringAggregation() bool {
	return true
}

// SupportsWithinAfterArrayAggregation returns true if the dialect supports
// ARRAY_AGG() [WITHIN GROUP (ORDER BY)] expressions.
func (d *DuckDbDialect) SupportsWithinAfterArrayAggregation() bool {
	return false
}

// SupportsWindowClauseNamedWindowReference returns true if the dialect
// supports referencing another named window within a window clause declaration.
func (d *DuckDbDialect) SupportsWindowClauseNamedWindowReference() bool {
	return true
}

// SupportsWindowFunctionNullTreatmentArg returns true if the dialect supports
// specifying null treatment as part of a window function's parameter list.
func (d *DuckDbDialect) SupportsWindowFunctionNullTreatmentArg() bool {
	return false
}

// SupportsMatchRecognize returns true if the dialect supports the MATCH_RECOGNIZE operation.
func (d *DuckDbDialect) SupportsMatchRecognize() bool {
	return false
}

// SupportsGroupByExpr returns true if the dialect supports GROUP BY expressions
// like GROUPING SETS, ROLLUP, or CUBE.
func (d *DuckDbDialect) SupportsGroupByExpr() bool {
	return true
}

// SupportsGroupByWithModifier returns true if the dialect supports GROUP BY
// modifiers prefixed by a WITH keyword (e.g., GROUP BY value WITH ROLLUP).
func (d *DuckDbDialect) SupportsGroupByWithModifier() bool {
	return false
}

// SupportsLeftAssociativeJoinsWithoutParens returns true if the dialect supports
// left-associative join parsing by default when parentheses are omitted in nested joins.
func (d *DuckDbDialect) SupportsLeftAssociativeJoinsWithoutParens() bool {
	return false
}

// SupportsOuterJoinOperator returns true if the dialect supports the (+)
// syntax for OUTER JOIN (Oracle-style).
func (d *DuckDbDialect) SupportsOuterJoinOperator() bool {
	return false
}

// SupportsCrossJoinConstraint returns true if the dialect supports a join
// specification on CROSS JOIN.
func (d *DuckDbDialect) SupportsCrossJoinConstraint() bool {
	return false
}

// SupportsConnectBy returns true if the dialect supports CONNECT BY for
// hierarchical queries (Oracle-style).
func (d *DuckDbDialect) SupportsConnectBy() bool {
	return false
}

// SupportsStartTransactionModifier returns true if the dialect supports
// BEGIN {DEFERRED | IMMEDIATE | EXCLUSIVE | TRY | CATCH} [TRANSACTION].
func (d *DuckDbDialect) SupportsStartTransactionModifier() bool {
	return false
}

// SupportsEndTransactionModifier returns true if the dialect supports
// END {TRY | CATCH} statements.
func (d *DuckDbDialect) SupportsEndTransactionModifier() bool {
	return false
}

// SupportsNamedFnArgsWithEqOperator returns true if the dialect supports
// named arguments of the form FUN(a = '1', b = '2').
func (d *DuckDbDialect) SupportsNamedFnArgsWithEqOperator() bool {
	return true
}

// SupportsNamedFnArgsWithColonOperator returns true if the dialect supports
// named arguments of the form FUN(a : '1', b : '2').
func (d *DuckDbDialect) SupportsNamedFnArgsWithColonOperator() bool {
	return false
}

// SupportsNamedFnArgsWithAssignmentOperator returns true if the dialect supports
// named arguments of the form FUN(a := '1', b := '2').
func (d *DuckDbDialect) SupportsNamedFnArgsWithAssignmentOperator() bool {
	return true
}

// SupportsNamedFnArgsWithRArrowOperator returns true if the dialect supports
// named arguments of the form FUN(a => '1', b => '2').
func (d *DuckDbDialect) SupportsNamedFnArgsWithRArrowOperator() bool {
	return true
}

// SupportsNamedFnArgsWithExprName returns true if the dialect supports
// argument name as arbitrary expression.
func (d *DuckDbDialect) SupportsNamedFnArgsWithExprName() bool {
	return false
}

// SupportsParenthesizedSetVariables returns true if the dialect supports
// multiple variable assignment using parentheses in a SET variable declaration.
func (d *DuckDbDialect) SupportsParenthesizedSetVariables() bool {
	return false
}

// SupportsCommaSeparatedSetAssignments returns true if the dialect supports
// multiple SET statements in a single statement.
func (d *DuckDbDialect) SupportsCommaSeparatedSetAssignments() bool {
	return false
}

// SupportsSetStmtWithoutOperator returns true if the dialect supports
// SET statements without an explicit assignment operator (e.g., SET SHOWPLAN_XML ON).
func (d *DuckDbDialect) SupportsSetStmtWithoutOperator() bool {
	return false
}

// SupportsSetNames returns true if the dialect supports SET NAMES <charset_name> [COLLATE <collation_name>].
func (d *DuckDbDialect) SupportsSetNames() bool {
	return false
}

// SupportsSelectWildcardExcept returns true if the dialect supports EXCEPT
// clause following a wildcard in a select list.
func (d *DuckDbDialect) SupportsSelectWildcardExcept() bool {
	return false
}

// SupportsSelectWildcardExclude returns true if the dialect supports EXCLUDE
// option following a wildcard in a projection section.
func (d *DuckDbDialect) SupportsSelectWildcardExclude() bool {
	return true
}

// SupportsSelectExclude returns true if the dialect supports EXCLUDE as the
// last item in the projection section, not necessarily after a wildcard.
func (d *DuckDbDialect) SupportsSelectExclude() bool {
	return false
}

// SupportsSelectWildcardReplace returns true if the dialect supports REPLACE
// option in SELECT * wildcard expressions.
func (d *DuckDbDialect) SupportsSelectWildcardReplace() bool {
	return true
}

// SupportsSelectWildcardIlike returns true if the dialect supports ILIKE
// option in SELECT * wildcard expressions.
func (d *DuckDbDialect) SupportsSelectWildcardIlike() bool {
	return false
}

// SupportsSelectWildcardRename returns true if the dialect supports RENAME
// option in SELECT * wildcard expressions.
func (d *DuckDbDialect) SupportsSelectWildcardRename() bool {
	return false
}

// SupportsSelectWildcardWithAlias returns true if the dialect supports aliasing
// a wildcard select item.
func (d *DuckDbDialect) SupportsSelectWildcardWithAlias() bool {
	return false
}

// SupportsSelectExprStar returns true if the dialect supports wildcard expansion
// on arbitrary expressions in projections.
func (d *DuckDbDialect) SupportsSelectExprStar() bool {
	return false
}

// SupportsFromFirstSelect returns true if the dialect supports "FROM-first" selects.
func (d *DuckDbDialect) SupportsFromFirstSelect() bool {
	return true
}

// SupportsEmptyProjections returns true if the dialect supports empty projections
// in SELECT statements.
func (d *DuckDbDialect) SupportsEmptyProjections() bool {
	return false
}

// SupportsSelectModifiers returns true if the dialect supports MySQL-specific
// SELECT modifiers like HIGH_PRIORITY, STRAIGHT_JOIN, SQL_SMALL_RESULT, etc.
func (d *DuckDbDialect) SupportsSelectModifiers() bool {
	return false
}

// SupportsPipeOperator returns true if the dialect supports the pipe operator (|>).
func (d *DuckDbDialect) SupportsPipeOperator() bool {
	return true
}

// SupportsTrailingCommas returns true if the dialect supports trailing commas.
func (d *DuckDbDialect) SupportsTrailingCommas() bool {
	return true
}

// SupportsProjectionTrailingCommas returns true if the dialect supports trailing
// commas in the projection list.
func (d *DuckDbDialect) SupportsProjectionTrailingCommas() bool {
	return false
}

// SupportsFromTrailingCommas returns true if the dialect supports trailing commas
// in the FROM clause.
func (d *DuckDbDialect) SupportsFromTrailingCommas() bool {
	return false
}

// SupportsColumnDefinitionTrailingCommas returns true if the dialect supports trailing
// commas in column definitions.
func (d *DuckDbDialect) SupportsColumnDefinitionTrailingCommas() bool {
	return false
}

// SupportsLimitComma returns true if the dialect supports parsing LIMIT 1, 2
// as LIMIT 2 OFFSET 1.
func (d *DuckDbDialect) SupportsLimitComma() bool {
	return false
}

// ConvertTypeBeforeValue returns true if the dialect has a CONVERT function
// which accepts a type first and an expression second.
func (d *DuckDbDialect) ConvertTypeBeforeValue() bool {
	return false
}

// SupportsTryConvert returns true if the dialect supports the TRY_CONVERT function.
func (d *DuckDbDialect) SupportsTryConvert() bool {
	return false
}

// SupportsBinaryKwAsCast returns true if the dialect supports casting an expression
// to a binary type using the BINARY <expr> syntax.
func (d *DuckDbDialect) SupportsBinaryKwAsCast() bool {
	return false
}

// SupportsObjectNameDoubleDotNotation returns true if the dialect supports
// double dot notation for object names (e.g., db_name..table_name).
func (d *DuckDbDialect) SupportsObjectNameDoubleDotNotation() bool {
	return false
}

// SupportsNumericPrefix returns true if the dialect supports identifiers
// starting with a numeric prefix.
func (d *DuckDbDialect) SupportsNumericPrefix() bool {
	return false
}

// SupportsNumericLiteralUnderscores returns true if the dialect supports
// numbers containing underscores (e.g., 10_000_000).
func (d *DuckDbDialect) SupportsNumericLiteralUnderscores() bool {
	return true
}

// SupportsInEmptyList returns true if the dialect supports (NOT) IN ()
// expressions with empty lists.
func (d *DuckDbDialect) SupportsInEmptyList() bool {
	return false
}

// SupportsDictionarySyntax returns true if the dialect supports defining
// structs or objects using syntax like {'x': 1, 'y': 2, 'z': 3}.
func (d *DuckDbDialect) SupportsDictionarySyntax() bool {
	return true
}

// SupportsMapLiteralSyntax returns true if the dialect supports defining
// objects using syntax like Map {1: 10, 2: 20}.
func (d *DuckDbDialect) SupportsMapLiteralSyntax() bool {
	return true
}

// SupportsStructLiteral returns true if the dialect supports STRUCT literal syntax.
func (d *DuckDbDialect) SupportsStructLiteral() bool {
	return true
}

// SupportsArrayLiteralSyntax returns true if the dialect supports array literal syntax.
func (d *DuckDbDialect) SupportsArrayLiteralSyntax() bool {
	return true
}

// SupportsLambdaFunctions returns true if the dialect supports lambda functions.
func (d *DuckDbDialect) SupportsLambdaFunctions() bool {
	return true
}

// SupportsCreateTableMultiSchemaInfoSources returns true if the dialect supports
// specifying multiple options in a CREATE TABLE statement.
func (d *DuckDbDialect) SupportsCreateTableMultiSchemaInfoSources() bool {
	return false
}

// SupportsCreateTableLikeParenthesized returns true if the dialect supports
// specifying which table to copy the schema from inside parenthesis.
func (d *DuckDbDialect) SupportsCreateTableLikeParenthesized() bool {
	return false
}

// SupportsCreateTableSelect returns true if the dialect supports CREATE TABLE SELECT.
func (d *DuckDbDialect) SupportsCreateTableSelect() bool {
	return true
}

// SupportsCreateViewCommentSyntax returns true if the dialect supports the COMMENT
// clause in CREATE VIEW statements using COMMENT = 'comment' syntax.
func (d *DuckDbDialect) SupportsCreateViewCommentSyntax() bool {
	return false
}

// SupportsArrayTypedefWithoutElementType returns true if the dialect supports
// ARRAY type without specifying an element type.
func (d *DuckDbDialect) SupportsArrayTypedefWithoutElementType() bool {
	return true
}

// SupportsArrayTypedefWithBrackets returns true if the dialect supports array
// type definition with brackets with optional size.
func (d *DuckDbDialect) SupportsArrayTypedefWithBrackets() bool {
	return true
}

// SupportsParensAroundTableFactor returns true if the dialect supports extra
// parentheses around lone table names or derived tables in the FROM clause.
func (d *DuckDbDialect) SupportsParensAroundTableFactor() bool {
	return false
}

// SupportsValuesAsTableFactor returns true if the dialect supports VALUES
// as a table factor without requiring parentheses.
func (d *DuckDbDialect) SupportsValuesAsTableFactor() bool {
	return false
}

// SupportsSemanticViewTableFactor returns true if the dialect supports
// SEMANTIC_VIEW() table functions.
func (d *DuckDbDialect) SupportsSemanticViewTableFactor() bool {
	return false
}

// SupportsTableVersioning returns true if the dialect supports querying
// historical table data by specifying which version to query.
func (d *DuckDbDialect) SupportsTableVersioning() bool {
	return false
}

// SupportsTableSampleBeforeAlias returns true if the dialect supports the
// TABLESAMPLE option before the table alias option.
func (d *DuckDbDialect) SupportsTableSampleBeforeAlias() bool {
	return false
}

// SupportsTableHints returns true if the dialect supports table hints in the FROM clause.
func (d *DuckDbDialect) SupportsTableHints() bool {
	return false
}

// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
func (d *DuckDbDialect) SupportsIndexHints() bool {
	return false
}

// SupportsAscDescInColumnDefinition returns true if the dialect supports
// ASC and DESC in column definitions.
func (d *DuckDbDialect) SupportsAscDescInColumnDefinition() bool {
	return false
}

// SupportsSpaceSeparatedColumnOptions returns true if the dialect supports
// space-separated column options in CREATE TABLE statements.
func (d *DuckDbDialect) SupportsSpaceSeparatedColumnOptions() bool {
	return false
}

// SupportsConstraintKeywordWithoutName returns true if the dialect supports
// CONSTRAINT keyword without a name in table constraint definitions.
func (d *DuckDbDialect) SupportsConstraintKeywordWithoutName() bool {
	return false
}

// SupportsKeyColumnOption returns true if the dialect supports the KEY keyword
// as part of column-level constraints.
func (d *DuckDbDialect) SupportsKeyColumnOption() bool {
	return false
}

// SupportsDataTypeSignedSuffix returns true if the dialect allows an optional
// SIGNED suffix after integer data types.
func (d *DuckDbDialect) SupportsDataTypeSignedSuffix() bool {
	return false
}

// SupportsNestedComments returns true if the dialect supports nested comments.
func (d *DuckDbDialect) SupportsNestedComments() bool {
	return true
}

// SupportsMultilineCommentHints returns true if the dialect supports optimizer
// hints in multiline comments (e.g., /*!50110 KEY_BLOCK_SIZE = 1024*/).
func (d *DuckDbDialect) SupportsMultilineCommentHints() bool {
	return false
}

// SupportsCommentOptimizerHint returns true if the dialect supports query
// optimizer hints in single and multi-line comments.
func (d *DuckDbDialect) SupportsCommentOptimizerHint() bool {
	return false
}

// SupportsCommentOn returns true if the dialect supports the COMMENT statement.
func (d *DuckDbDialect) SupportsCommentOn() bool {
	return false
}

// RequiresSingleLineCommentWhitespace returns true if the dialect requires
// a whitespace character after -- to start a single line comment.
func (d *DuckDbDialect) RequiresSingleLineCommentWhitespace() bool {
	return false
}

// SupportsExplainWithUtilityOptions returns true if the dialect supports
// EXPLAIN statements with utility options.
func (d *DuckDbDialect) SupportsExplainWithUtilityOptions() bool {
	return true
}

// SupportsExecuteImmediate returns true if the dialect supports EXECUTE IMMEDIATE statements.
func (d *DuckDbDialect) SupportsExecuteImmediate() bool {
	return false
}

// AllowExtractCustom returns true if the dialect allows the EXTRACT function
// to use words other than keywords.
func (d *DuckDbDialect) AllowExtractCustom() bool {
	return false
}

// AllowExtractSingleQuotes returns true if the dialect allows the EXTRACT
// function to use single quotes in the part being extracted.
func (d *DuckDbDialect) AllowExtractSingleQuotes() bool {
	return true
}

// SupportsExtractCommaSyntax returns true if the dialect supports EXTRACT
// function with a comma separator instead of FROM.
func (d *DuckDbDialect) SupportsExtractCommaSyntax() bool {
	return false
}

// SupportsSubqueryAsFunctionArg returns true if the dialect supports a subquery
// passed to a function as the only argument without enclosing parentheses.
func (d *DuckDbDialect) SupportsSubqueryAsFunctionArg() bool {
	return false
}

// SupportsDollarPlaceholder returns true if the dialect allows dollar placeholders (e.g., $var).
func (d *DuckDbDialect) SupportsDollarPlaceholder() bool {
	return false
}

// SupportsCreateIndexWithClause returns true if the dialect supports WITH clause
// in CREATE INDEX statement.
func (d *DuckDbDialect) SupportsCreateIndexWithClause() bool {
	return false
}

// RequireIntervalQualifier returns true if the dialect requires units
// (qualifiers) to be specified in INTERVAL expressions.
func (d *DuckDbDialect) RequireIntervalQualifier() bool {
	return true
}

// SupportsIntervalOptions returns true if the dialect supports INTERVAL data
// type with Postgres-style options.
func (d *DuckDbDialect) SupportsIntervalOptions() bool {
	return false
}

// SupportsFactorialOperator returns true if the dialect supports a!
// expressions for factorial.
func (d *DuckDbDialect) SupportsFactorialOperator() bool {
	return false
}

// SupportsBitwiseShiftOperators returns true if the dialect supports
// << and >> shift operators.
func (d *DuckDbDialect) SupportsBitwiseShiftOperators() bool {
	return true
}

// SupportsNotnullOperator returns true if the dialect supports the x NOTNULL
// operator expression.
func (d *DuckDbDialect) SupportsNotnullOperator() bool {
	return true
}

// SupportsBangNotOperator returns true if the dialect supports !a syntax
// for boolean NOT expressions.
func (d *DuckDbDialect) SupportsBangNotOperator() bool {
	return false
}

// SupportsDoubleAmpersandOperator returns true if the dialect considers
// the && operator as a boolean AND operator.
func (d *DuckDbDialect) SupportsDoubleAmpersandOperator() bool {
	return false
}

// SupportsMatchAgainst returns true if the dialect supports MATCH() AGAINST() syntax.
func (d *DuckDbDialect) SupportsMatchAgainst() bool {
	return false
}

// SupportsUserHostGrantee returns true if the dialect supports MySQL-style
// 'user'@'host' grantee syntax.
func (d *DuckDbDialect) SupportsUserHostGrantee() bool {
	return false
}

// SupportsListenNotify returns true if the dialect supports LISTEN, UNLISTEN,
// and NOTIFY statements.
func (d *DuckDbDialect) SupportsListenNotify() bool {
	return false
}

// SupportsLoadData returns true if the dialect supports LOAD DATA statement.
func (d *DuckDbDialect) SupportsLoadData() bool {
	return false
}

// SupportsLoadExtension returns true if the dialect supports LOAD extension statement.
// See https://duckdb.org/docs/sql/statements/load_and_install.html#load
func (d *DuckDbDialect) SupportsLoadExtension() bool {
	return true
}

// SupportsTopBeforeDistinct returns true if the dialect expects the TOP option
// before the ALL/DISTINCT options in a SELECT statement.
func (d *DuckDbDialect) SupportsTopBeforeDistinct() bool {
	return false
}

// SupportsBooleanLiterals returns true if the dialect supports boolean
// literals (true and false).
func (d *DuckDbDialect) SupportsBooleanLiterals() bool {
	return true
}

// SupportsShowLikeBeforeIn returns true if the dialect supports the LIKE
// option in a SHOW statement before the IN option.
func (d *DuckDbDialect) SupportsShowLikeBeforeIn() bool {
	return false
}

// SupportsPartiQL returns true if the dialect supports PartiQL for querying
// semi-structured data.
func (d *DuckDbDialect) SupportsPartiQL() bool {
	return false
}

// SupportsEqAliasAssignment returns true if the dialect supports treating
// the equals operator = within a SelectItem as an alias assignment operator.
func (d *DuckDbDialect) SupportsEqAliasAssignment() bool {
	return false
}

// SupportsInsertSet returns true if the dialect supports INSERT INTO ... SET syntax.
func (d *DuckDbDialect) SupportsInsertSet() bool {
	return false
}

// SupportsInsertTableFunction returns true if the dialect supports table
// function in insertion.
func (d *DuckDbDialect) SupportsInsertTableFunction() bool {
	return false
}

// SupportsInsertTableQuery returns true if the dialect supports table queries
// in insertion (e.g., SELECT INTO (<query>) ...).
func (d *DuckDbDialect) SupportsInsertTableQuery() bool {
	return false
}

// SupportsInsertFormat returns true if the dialect supports insert formats
// (e.g., INSERT INTO ... FORMAT <format>).
func (d *DuckDbDialect) SupportsInsertFormat() bool {
	return false
}

// SupportsInsertTableAlias returns true if the dialect supports INSERT INTO
// table [[AS] alias] syntax.
func (d *DuckDbDialect) SupportsInsertTableAlias() bool {
	return false
}

// SupportsAlterColumnTypeUsing returns true if the dialect supports the USING
// clause in an ALTER COLUMN statement.
func (d *DuckDbDialect) SupportsAlterColumnTypeUsing() bool {
	return false
}

// SupportsCommaSeparatedDropColumnList returns true if the dialect supports
// ALTER TABLE tbl DROP COLUMN c1, ..., cn.
func (d *DuckDbDialect) SupportsCommaSeparatedDropColumnList() bool {
	return false
}

// SupportsOrderByAll returns true if the dialect supports ORDER BY ALL.
// See https://duckdb.org/docs/sql/query_syntax/orderby.html#order-by-all-examples
func (d *DuckDbDialect) SupportsOrderByAll() bool {
	return true
}

// SupportsGeometricTypes returns true if the dialect supports geometric types
// (Postgres geometric operations).
func (d *DuckDbDialect) SupportsGeometricTypes() bool {
	return false
}

// DescribeRequiresTableKeyword returns true if the dialect requires the TABLE
// keyword after DESCRIBE.
func (d *DuckDbDialect) DescribeRequiresTableKeyword() bool {
	return false
}

// SupportsOptimizeTable returns true if the dialect supports OPTIMIZE TABLE.
func (d *DuckDbDialect) SupportsOptimizeTable() bool {
	return false
}

// SupportsPrewhere returns true if the dialect supports PREWHERE clause.
func (d *DuckDbDialect) SupportsPrewhere() bool {
	return false
}

// SupportsWithFill returns true if the dialect supports WITH FILL clause.
func (d *DuckDbDialect) SupportsWithFill() bool {
	return false
}

// SupportsLimitBy returns true if the dialect supports LIMIT BY clause.
func (d *DuckDbDialect) SupportsLimitBy() bool {
	return false
}

// SupportsInterpolate returns true if the dialect supports INTERPOLATE clause.
func (d *DuckDbDialect) SupportsInterpolate() bool {
	return false
}

// SupportsSettings returns true if the dialect supports SETTINGS clause.
func (d *DuckDbDialect) SupportsSettings() bool {
	return false
}

// SupportsSelectFormat returns true if the dialect supports FORMAT clause in SELECT.
func (d *DuckDbDialect) SupportsSelectFormat() bool {
	return false
}

// SupportsInstall returns true if the dialect supports INSTALL statement.
// See https://duckdb.org/docs/extensions/overview
func (d *DuckDbDialect) SupportsInstall() bool {
	return true
}

// SupportsDetach returns true if the dialect supports DETACH statement.
// See https://duckdb.org/docs/sql/statements/attach#detach-syntax
func (d *DuckDbDialect) SupportsDetach() bool {
	return true
}

// SupportsCommaSeparatedTrim returns true if the dialect supports the two-argument
// comma-separated form of TRIM function: TRIM(expr, characters).
func (d *DuckDbDialect) SupportsCommaSeparatedTrim() bool {
	return true
}

// PrecValue returns the precedence value for a given Precedence level.
func (d *DuckDbDialect) PrecValue(prec dialects.Precedence) uint8 {
	return uint8(prec)
}

// PrecUnknown returns the precedence when precedence is otherwise unknown.
func (d *DuckDbDialect) PrecUnknown() uint8 {
	return 0
}

// GetNextPrecedence returns the dialect-specific precedence override for the next token.
func (d *DuckDbDialect) GetNextPrecedence(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// GetNextPrecedenceDefault implements the default precedence logic.
func (d *DuckDbDialect) GetNextPrecedenceDefault(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// ParsePrefix is a dialect-specific prefix parser override.
func (d *DuckDbDialect) ParsePrefix(parser dialects.ParserAccessor) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseInfix is a dialect-specific infix parser override.
func (d *DuckDbDialect) ParseInfix(parser dialects.ParserAccessor, expr ast.Expr, precedence uint8) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseStatement is a dialect-specific statement parser override.
func (d *DuckDbDialect) ParseStatement(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	return nil, false, nil
}

// ParseColumnOption is a dialect-specific column option parser override.
func (d *DuckDbDialect) ParseColumnOption(parser dialects.ParserAccessor) (dialects.ColumnOption, bool, error) {
	return nil, false, nil
}
