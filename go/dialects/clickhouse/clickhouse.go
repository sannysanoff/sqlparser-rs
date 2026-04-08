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

package clickhouse

import (
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/token"
)

// ClickHouseDialect is a dialect for ClickHouse SQL implementation.
// See https://clickhouse.com/
//
// ClickHouseDialect implements the following capability interfaces:
//   - dialects.CoreDialect
//   - dialects.IdentifierDialect (with " and ` quotes)
//   - dialects.KeywordDialect
//   - dialects.StringLiteralDialect (with backslash escape, concatenation)
//   - dialects.AggregationDialect
//   - dialects.GroupByDialect (with GROUP BY Expr, WITH modifiers)
//   - dialects.JoinDialect
//   - dialects.ConnectByDialect
//   - dialects.TransactionDialect
//   - dialects.NamedArgumentDialect (with => operator)
//   - dialects.SetDialect
//   - dialects.SelectDialect (with LIMIT comma, wildcard EXCEPT/REPLACE, FROM-first)
//   - dialects.TypeConversionDialect
//   - dialects.ObjectReferenceDialect (with numeric literal underscores)
//   - dialects.InExpressionDialect
//   - dialects.LiteralDialect (with array, dictionary, lambda)
//   - dialects.TableDefinitionDialect (with array typedef, VALUES)
//   - dialects.ColumnDefinitionDialect
//   - dialects.CommentDialect (with nested comments)
//   - dialects.ExplainDialect
//   - dialects.ExecuteDialect
//   - dialects.ExtractDialect
//   - dialects.SubqueryDialect (subquery as function arg)
//   - dialects.PlaceholderDialect
//   - dialects.IndexDialect
//   - dialects.IntervalDialect
//   - dialects.OperatorDialect
//   - dialects.MatchDialect
//   - dialects.GranteeDialect
//   - dialects.ListenNotifyDialect
//   - dialects.LoadDialect
//   - dialects.TopDistinctDialect
//   - dialects.BooleanLiteralDialect
//   - dialects.ShowDialect
//   - dialects.PartiQLDialect
//   - dialects.AliasDialect
//   - dialects.InsertDialect (with table function, FORMAT)
//   - dialects.AlterTableDialect
//   - dialects.OrderByDialect (ORDER BY ALL)
//   - dialects.GeometricDialect
//   - dialects.DescribeDialect (requires TABLE keyword)
//   - dialects.ClickHouseDialect (OPTIMIZE, PREWHERE, WITH FILL, LIMIT BY, INTERPOLATE, SETTINGS, FORMAT)
//   - dialects.DuckDBDialect
//   - dialects.TrimDialect
//
// Compile-time interface checks:
var _ dialects.CompleteDialect = (*ClickHouseDialect)(nil)
var _ dialects.CoreDialect = (*ClickHouseDialect)(nil)
var _ dialects.IdentifierDialect = (*ClickHouseDialect)(nil)
var _ dialects.KeywordDialect = (*ClickHouseDialect)(nil)
var _ dialects.StringLiteralDialect = (*ClickHouseDialect)(nil)
var _ dialects.AggregationDialect = (*ClickHouseDialect)(nil)
var _ dialects.GroupByDialect = (*ClickHouseDialect)(nil)
var _ dialects.JoinDialect = (*ClickHouseDialect)(nil)
var _ dialects.ConnectByDialect = (*ClickHouseDialect)(nil)
var _ dialects.TransactionDialect = (*ClickHouseDialect)(nil)
var _ dialects.NamedArgumentDialect = (*ClickHouseDialect)(nil)
var _ dialects.SetDialect = (*ClickHouseDialect)(nil)
var _ dialects.SelectDialect = (*ClickHouseDialect)(nil)
var _ dialects.TypeConversionDialect = (*ClickHouseDialect)(nil)
var _ dialects.ObjectReferenceDialect = (*ClickHouseDialect)(nil)
var _ dialects.InExpressionDialect = (*ClickHouseDialect)(nil)
var _ dialects.LiteralDialect = (*ClickHouseDialect)(nil)
var _ dialects.TableDefinitionDialect = (*ClickHouseDialect)(nil)
var _ dialects.ColumnDefinitionDialect = (*ClickHouseDialect)(nil)
var _ dialects.CommentDialect = (*ClickHouseDialect)(nil)
var _ dialects.ExplainDialect = (*ClickHouseDialect)(nil)
var _ dialects.ExecuteDialect = (*ClickHouseDialect)(nil)
var _ dialects.ExtractDialect = (*ClickHouseDialect)(nil)
var _ dialects.SubqueryDialect = (*ClickHouseDialect)(nil)
var _ dialects.PlaceholderDialect = (*ClickHouseDialect)(nil)
var _ dialects.IndexDialect = (*ClickHouseDialect)(nil)
var _ dialects.IntervalDialect = (*ClickHouseDialect)(nil)
var _ dialects.OperatorDialect = (*ClickHouseDialect)(nil)
var _ dialects.MatchDialect = (*ClickHouseDialect)(nil)
var _ dialects.GranteeDialect = (*ClickHouseDialect)(nil)
var _ dialects.ListenNotifyDialect = (*ClickHouseDialect)(nil)
var _ dialects.LoadDialect = (*ClickHouseDialect)(nil)
var _ dialects.TopDistinctDialect = (*ClickHouseDialect)(nil)
var _ dialects.BooleanLiteralDialect = (*ClickHouseDialect)(nil)
var _ dialects.ShowDialect = (*ClickHouseDialect)(nil)
var _ dialects.PartiQLDialect = (*ClickHouseDialect)(nil)
var _ dialects.AliasDialect = (*ClickHouseDialect)(nil)
var _ dialects.InsertDialect = (*ClickHouseDialect)(nil)
var _ dialects.AlterTableDialect = (*ClickHouseDialect)(nil)
var _ dialects.OrderByDialect = (*ClickHouseDialect)(nil)
var _ dialects.GeometricDialect = (*ClickHouseDialect)(nil)
var _ dialects.DescribeDialect = (*ClickHouseDialect)(nil)
var _ dialects.ClickHouseDialect = (*ClickHouseDialect)(nil)
var _ dialects.DuckDBDialect = (*ClickHouseDialect)(nil)
var _ dialects.TrimDialect = (*ClickHouseDialect)(nil)

type ClickHouseDialect struct{}

// NewClickHouseDialect creates a new instance of ClickHouseDialect.
func NewClickHouseDialect() *ClickHouseDialect {
	return &ClickHouseDialect{}
}

// Dialect returns the dialect identifier.
func (d *ClickHouseDialect) Dialect() string {
	return "clickhouse"
}

// IsIdentifierStart returns true if the character is a valid start character
// for an unquoted identifier.
// See https://clickhouse.com/docs/en/sql-reference/syntax/#syntax-identifiers
func (d *ClickHouseDialect) IsIdentifierStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// IsIdentifierPart returns true if the character is a valid unquoted
// identifier character (not necessarily at the start).
func (d *ClickHouseDialect) IsIdentifierPart(ch rune) bool {
	return d.IsIdentifierStart(ch) || (ch >= '0' && ch <= '9')
}

// IsDelimitedIdentifierStart returns true if the character starts a quoted identifier.
func (d *ClickHouseDialect) IsDelimitedIdentifierStart(ch rune) bool {
	return ch == '"' || ch == '`'
}

// IsNestedDelimitedIdentifierStart returns true if the character starts a
// potential nested quoted identifier.
func (d *ClickHouseDialect) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return false
}

// PeekNestedDelimitedIdentifierQuotes returns the outer quote style and
// optional inner quote style if applicable.
func (d *ClickHouseDialect) PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*dialects.NestedIdentifierQuote, int) {
	return nil, pos
}

// IdentifierQuoteStyle returns the character used to quote identifiers.
func (d *ClickHouseDialect) IdentifierQuoteStyle(identifier string) *rune {
	return nil
}

// IsCustomOperatorPart returns true if the character is part of a custom operator.
func (d *ClickHouseDialect) IsCustomOperatorPart(ch rune) bool {
	return false
}

// IsReservedForIdentifier returns true if the keyword is reserved.
func (d *ClickHouseDialect) IsReservedForIdentifier(kw token.Keyword) bool {
	return token.IsReservedForIdentifier(kw)
}

// GetReservedKeywordsForSelectItemOperator returns reserved keywords for select items.
func (d *ClickHouseDialect) GetReservedKeywordsForSelectItemOperator() []token.Keyword {
	return nil
}

// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
func (d *ClickHouseDialect) GetReservedGranteesTypes() []dialects.GranteesType {
	return nil
}

// IsColumnAlias returns true if the keyword should be parsed as a column alias.
func (d *ClickHouseDialect) IsColumnAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == "AS" || token.IsReservedForColumnAlias(kw)
}

// IsSelectItemAlias returns true if the specified keyword should be parsed as a select item alias.
func (d *ClickHouseDialect) IsSelectItemAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsColumnAlias(kw, parser)
}

// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier.
func (d *ClickHouseDialect) IsTableFactor(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == "TABLE" || kw == "LATERAL"
}

// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias.
func (d *ClickHouseDialect) IsTableAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return !token.IsReservedForTableAlias(kw)
}

// IsTableFactorAlias returns true if the specified keyword should be parsed as a table factor alias.
func (d *ClickHouseDialect) IsTableFactorAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return explicit || d.IsTableAlias(kw, parser)
}

// IsIdentifierGeneratingFunctionName returns true if the ident is a function that generates identifiers.
func (d *ClickHouseDialect) IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool {
	return false
}

// SupportsStringLiteralBackslashEscape returns true if the dialect supports
// escaping characters via '\' in string literals.
func (d *ClickHouseDialect) SupportsStringLiteralBackslashEscape() bool {
	return true
}

// IgnoresWildcardEscapes returns true if the dialect strips the backslash
// when escaping LIKE wildcards (%, _).
func (d *ClickHouseDialect) IgnoresWildcardEscapes() bool {
	return false
}

// SupportsUnicodeStringLiteral returns true if the dialect supports string
// literals with U& prefix for Unicode code points.
func (d *ClickHouseDialect) SupportsUnicodeStringLiteral() bool {
	return false
}

// SupportsTripleQuotedString returns true if the dialect supports triple
// quoted string literals (e.g., """abc""").
func (d *ClickHouseDialect) SupportsTripleQuotedString() bool {
	return false
}

// SupportsStringLiteralConcatenation returns true if the dialect supports
// concatenating string literals (e.g., 'Hello ' 'world').
func (d *ClickHouseDialect) SupportsStringLiteralConcatenation() bool {
	return true
}

// SupportsStringLiteralConcatenationWithNewline returns true if the dialect
// supports concatenating string literals with a newline between them.
func (d *ClickHouseDialect) SupportsStringLiteralConcatenationWithNewline() bool {
	return true
}

// SupportsQuoteDelimitedString returns true if the dialect supports
// quote-delimited string literals (e.g., Q'{...}') for Oracle-style strings.
func (d *ClickHouseDialect) SupportsQuoteDelimitedString() bool {
	return false
}

// SupportsStringEscapeConstant returns true if the dialect supports the
// E'...' syntax for string literals with escape sequences (Postgres).
func (d *ClickHouseDialect) SupportsStringEscapeConstant() bool {
	return false
}

// SupportsDollarQuotedString returns true if the dialect supports
// dollar-quoted string literals (e.g., $$...$$) for PostgreSQL-style strings.
func (d *ClickHouseDialect) SupportsDollarQuotedString() bool {
	return false
}

// SupportsFilterDuringAggregation returns true if the dialect supports
// FILTER (WHERE expr) for aggregate queries.
func (d *ClickHouseDialect) SupportsFilterDuringAggregation() bool {
	return false
}

// SupportsWithinAfterArrayAggregation returns true if the dialect supports
// ARRAY_AGG() [WITHIN GROUP (ORDER BY)] expressions.
func (d *ClickHouseDialect) SupportsWithinAfterArrayAggregation() bool {
	return false
}

// SupportsWindowClauseNamedWindowReference returns true if the dialect
// supports referencing another named window within a window clause declaration.
func (d *ClickHouseDialect) SupportsWindowClauseNamedWindowReference() bool {
	return false
}

// SupportsWindowFunctionNullTreatmentArg returns true if the dialect supports
// specifying null treatment as part of a window function's parameter list.
func (d *ClickHouseDialect) SupportsWindowFunctionNullTreatmentArg() bool {
	return false
}

// SupportsMatchRecognize returns true if the dialect supports the MATCH_RECOGNIZE operation.
func (d *ClickHouseDialect) SupportsMatchRecognize() bool {
	return false
}

// SupportsGroupByExpr returns true if the dialect supports GROUP BY expressions
// like GROUPING SETS, ROLLUP, or CUBE.
// See https://clickhouse.com/docs/en/sql-reference/aggregate-functions/grouping_function#grouping-sets
func (d *ClickHouseDialect) SupportsGroupByExpr() bool {
	return true
}

// SupportsGroupByWithModifier returns true if the dialect supports GROUP BY
// modifiers prefixed by a WITH keyword (e.g., GROUP BY value WITH ROLLUP).
// See https://clickhouse.com/docs/en/sql-reference/statements/select/group-by#rollup-modifier
func (d *ClickHouseDialect) SupportsGroupByWithModifier() bool {
	return true
}

// SupportsLeftAssociativeJoinsWithoutParens returns true if the dialect supports
// left-associative join parsing by default when parentheses are omitted in nested joins.
func (d *ClickHouseDialect) SupportsLeftAssociativeJoinsWithoutParens() bool {
	return false
}

// SupportsOuterJoinOperator returns true if the dialect supports the (+)
// syntax for OUTER JOIN (Oracle-style).
func (d *ClickHouseDialect) SupportsOuterJoinOperator() bool {
	return false
}

// SupportsCrossJoinConstraint returns true if the dialect supports a join
// specification on CROSS JOIN.
func (d *ClickHouseDialect) SupportsCrossJoinConstraint() bool {
	return false
}

// SupportsConnectBy returns true if the dialect supports CONNECT BY for
// hierarchical queries (Oracle-style).
func (d *ClickHouseDialect) SupportsConnectBy() bool {
	return false
}

// SupportsStartTransactionModifier returns true if the dialect supports
// BEGIN {DEFERRED | IMMEDIATE | EXCLUSIVE | TRY | CATCH} [TRANSACTION].
func (d *ClickHouseDialect) SupportsStartTransactionModifier() bool {
	return false
}

// SupportsEndTransactionModifier returns true if the dialect supports
// END {TRY | CATCH} statements.
func (d *ClickHouseDialect) SupportsEndTransactionModifier() bool {
	return false
}

// SupportsNamedFnArgsWithEqOperator returns true if the dialect supports
// named arguments of the form FUN(a = '1', b = '2').
func (d *ClickHouseDialect) SupportsNamedFnArgsWithEqOperator() bool {
	return false
}

// SupportsNamedFnArgsWithColonOperator returns true if the dialect supports
// named arguments of the form FUN(a : '1', b : '2').
func (d *ClickHouseDialect) SupportsNamedFnArgsWithColonOperator() bool {
	return false
}

// SupportsNamedFnArgsWithAssignmentOperator returns true if the dialect supports
// named arguments of the form FUN(a := '1', b := '2').
func (d *ClickHouseDialect) SupportsNamedFnArgsWithAssignmentOperator() bool {
	return false
}

// SupportsNamedFnArgsWithRArrowOperator returns true if the dialect supports
// named arguments of the form FUN(a => '1', b => '2').
func (d *ClickHouseDialect) SupportsNamedFnArgsWithRArrowOperator() bool {
	return true
}

// SupportsNamedFnArgsWithExprName returns true if the dialect supports
// argument name as arbitrary expression.
func (d *ClickHouseDialect) SupportsNamedFnArgsWithExprName() bool {
	return false
}

// SupportsParenthesizedSetVariables returns true if the dialect supports
// multiple variable assignment using parentheses in a SET variable declaration.
func (d *ClickHouseDialect) SupportsParenthesizedSetVariables() bool {
	return false
}

// SupportsCommaSeparatedSetAssignments returns true if the dialect supports
// multiple SET statements in a single statement.
func (d *ClickHouseDialect) SupportsCommaSeparatedSetAssignments() bool {
	return false
}

// SupportsSetStmtWithoutOperator returns true if the dialect supports
// SET statements without an explicit assignment operator (e.g., SET SHOWPLAN_XML ON).
func (d *ClickHouseDialect) SupportsSetStmtWithoutOperator() bool {
	return false
}

// SupportsSetNames returns true if the dialect supports SET NAMES <charset_name> [COLLATE <collation_name>].
func (d *ClickHouseDialect) SupportsSetNames() bool {
	return false
}

// SupportsSelectWildcardExcept returns true if the dialect supports EXCEPT
// clause following a wildcard in a select list.
func (d *ClickHouseDialect) SupportsSelectWildcardExcept() bool {
	return true
}

// SupportsSelectWildcardExclude returns true if the dialect supports EXCLUDE
// option following a wildcard in a projection section.
func (d *ClickHouseDialect) SupportsSelectWildcardExclude() bool {
	return false
}

// SupportsSelectExclude returns true if the dialect supports EXCLUDE as the
// last item in the projection section, not necessarily after a wildcard.
func (d *ClickHouseDialect) SupportsSelectExclude() bool {
	return false
}

// SupportsSelectWildcardReplace returns true if the dialect supports REPLACE
// option in SELECT * wildcard expressions.
// See https://clickhouse.com/docs/sql-reference/statements/select#replace
func (d *ClickHouseDialect) SupportsSelectWildcardReplace() bool {
	return true
}

// SupportsSelectWildcardIlike returns true if the dialect supports ILIKE
// option in SELECT * wildcard expressions.
func (d *ClickHouseDialect) SupportsSelectWildcardIlike() bool {
	return false
}

// SupportsSelectWildcardRename returns true if the dialect supports RENAME
// option in SELECT * wildcard expressions.
func (d *ClickHouseDialect) SupportsSelectWildcardRename() bool {
	return false
}

// SupportsSelectWildcardWithAlias returns true if the dialect supports aliasing
// a wildcard select item.
func (d *ClickHouseDialect) SupportsSelectWildcardWithAlias() bool {
	return false
}

// SupportsSelectExprStar returns true if the dialect supports wildcard expansion
// on arbitrary expressions in projections.
func (d *ClickHouseDialect) SupportsSelectExprStar() bool {
	return false
}

// SupportsFromFirstSelect returns true if the dialect supports "FROM-first" selects.
func (d *ClickHouseDialect) SupportsFromFirstSelect() bool {
	return true
}

// SupportsEmptyProjections returns true if the dialect supports empty projections
// in SELECT statements.
func (d *ClickHouseDialect) SupportsEmptyProjections() bool {
	return false
}

// SupportsSelectModifiers returns true if the dialect supports MySQL-specific
// SELECT modifiers like HIGH_PRIORITY, STRAIGHT_JOIN, SQL_SMALL_RESULT, etc.
func (d *ClickHouseDialect) SupportsSelectModifiers() bool {
	return false
}

// SupportsPipeOperator returns true if the dialect supports the pipe operator (|>).
func (d *ClickHouseDialect) SupportsPipeOperator() bool {
	return false
}

// SupportsTrailingCommas returns true if the dialect supports trailing commas.
func (d *ClickHouseDialect) SupportsTrailingCommas() bool {
	return false
}

// SupportsProjectionTrailingCommas returns true if the dialect supports trailing
// commas in the projection list.
func (d *ClickHouseDialect) SupportsProjectionTrailingCommas() bool {
	return false
}

// SupportsFromTrailingCommas returns true if the dialect supports trailing commas
// in the FROM clause.
func (d *ClickHouseDialect) SupportsFromTrailingCommas() bool {
	return false
}

// SupportsColumnDefinitionTrailingCommas returns true if the dialect supports trailing
// commas in column definitions.
func (d *ClickHouseDialect) SupportsColumnDefinitionTrailingCommas() bool {
	return false
}

// SupportsLimitComma returns true if the dialect supports parsing LIMIT 1, 2
// as LIMIT 2 OFFSET 1.
func (d *ClickHouseDialect) SupportsLimitComma() bool {
	return true
}

// ConvertTypeBeforeValue returns true if the dialect has a CONVERT function
// which accepts a type first and an expression second.
func (d *ClickHouseDialect) ConvertTypeBeforeValue() bool {
	return false
}

// SupportsTryConvert returns true if the dialect supports the TRY_CONVERT function.
func (d *ClickHouseDialect) SupportsTryConvert() bool {
	return false
}

// SupportsBinaryKwAsCast returns true if the dialect supports casting an expression
// to a binary type using the BINARY <expr> syntax.
func (d *ClickHouseDialect) SupportsBinaryKwAsCast() bool {
	return false
}

// SupportsObjectNameDoubleDotNotation returns true if the dialect supports
// double dot notation for object names (e.g., db_name..table_name).
func (d *ClickHouseDialect) SupportsObjectNameDoubleDotNotation() bool {
	return false
}

// SupportsNumericPrefix returns true if the dialect supports identifiers
// starting with a numeric prefix.
func (d *ClickHouseDialect) SupportsNumericPrefix() bool {
	return false
}

// SupportsNumericLiteralUnderscores returns true if the dialect supports
// numbers containing underscores (e.g., 10_000_000).
func (d *ClickHouseDialect) SupportsNumericLiteralUnderscores() bool {
	return true
}

// SupportsInEmptyList returns true if the dialect supports (NOT) IN ()
// expressions with empty lists.
func (d *ClickHouseDialect) SupportsInEmptyList() bool {
	return false
}

// SupportsDictionarySyntax returns true if the dialect supports defining
// structs or objects using syntax like {'x': 1, 'y': 2, 'z': 3}.
// ClickHouse uses this for some FORMAT expressions in INSERT context.
// See https://clickhouse.com/docs/en/interfaces/formats
func (d *ClickHouseDialect) SupportsDictionarySyntax() bool {
	return true
}

// SupportsMapLiteralSyntax returns true if the dialect supports defining
// objects using syntax like Map {1: 10, 2: 20}.
func (d *ClickHouseDialect) SupportsMapLiteralSyntax() bool {
	return false
}

// SupportsStructLiteral returns true if the dialect supports STRUCT literal syntax.
func (d *ClickHouseDialect) SupportsStructLiteral() bool {
	return false
}

// SupportsArrayLiteralSyntax returns true if the dialect supports array literal syntax.
func (d *ClickHouseDialect) SupportsArrayLiteralSyntax() bool {
	return true
}

// SupportsLambdaFunctions returns true if the dialect supports lambda functions.
// See https://clickhouse.com/docs/en/sql-reference/functions#higher-order-functions---operator-and-lambdaparams-expr-function
func (d *ClickHouseDialect) SupportsLambdaFunctions() bool {
	return true
}

// SupportsCreateTableMultiSchemaInfoSources returns true if the dialect supports
// specifying multiple options in a CREATE TABLE statement.
func (d *ClickHouseDialect) SupportsCreateTableMultiSchemaInfoSources() bool {
	return false
}

// SupportsCreateTableLikeParenthesized returns true if the dialect supports
// specifying which table to copy the schema from inside parenthesis.
func (d *ClickHouseDialect) SupportsCreateTableLikeParenthesized() bool {
	return false
}

// SupportsCreateTableSelect returns true if the dialect supports CREATE TABLE SELECT.
func (d *ClickHouseDialect) SupportsCreateTableSelect() bool {
	return true
}

// SupportsCreateViewCommentSyntax returns true if the dialect supports the COMMENT
// clause in CREATE VIEW statements using COMMENT = 'comment' syntax.
func (d *ClickHouseDialect) SupportsCreateViewCommentSyntax() bool {
	return false
}

// SupportsArrayTypedefWithoutElementType returns true if the dialect supports
// ARRAY type without specifying an element type.
func (d *ClickHouseDialect) SupportsArrayTypedefWithoutElementType() bool {
	return true
}

// SupportsArrayTypedefWithBrackets returns true if the dialect supports array
// type definition with brackets with optional size.
func (d *ClickHouseDialect) SupportsArrayTypedefWithBrackets() bool {
	return true
}

// SupportsParensAroundTableFactor returns true if the dialect supports extra
// parentheses around lone table names or derived tables in the FROM clause.
func (d *ClickHouseDialect) SupportsParensAroundTableFactor() bool {
	return false
}

// SupportsValuesAsTableFactor returns true if the dialect supports VALUES
// as a table factor without requiring parentheses.
func (d *ClickHouseDialect) SupportsValuesAsTableFactor() bool {
	return true
}

// SupportsUnnestTableFactor returns false for ClickHouseDialect.
func (d *ClickHouseDialect) SupportsUnnestTableFactor() bool {
	return false
}

// SupportsSemanticViewTableFactor returns true if the dialect supports
// SEMANTIC_VIEW() table functions.
func (d *ClickHouseDialect) SupportsSemanticViewTableFactor() bool {
	return false
}

// SupportsTableVersioning returns true if the dialect supports querying
// historical table data by specifying which version to query.
func (d *ClickHouseDialect) SupportsTableVersioning() bool {
	return false
}

// SupportsTableSampleBeforeAlias returns true if the dialect supports the
// TABLESAMPLE option before the table alias option.
func (d *ClickHouseDialect) SupportsTableSampleBeforeAlias() bool {
	return false
}

// SupportsTableHints returns true if the dialect supports table hints in the FROM clause.
func (d *ClickHouseDialect) SupportsTableHints() bool {
	return false
}

// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
func (d *ClickHouseDialect) SupportsIndexHints() bool {
	return false
}

// SupportsAscDescInColumnDefinition returns true if the dialect supports
// ASC and DESC in column definitions.
func (d *ClickHouseDialect) SupportsAscDescInColumnDefinition() bool {
	return false
}

// SupportsSpaceSeparatedColumnOptions returns true if the dialect supports
// space-separated column options in CREATE TABLE statements.
func (d *ClickHouseDialect) SupportsSpaceSeparatedColumnOptions() bool {
	return false
}

// SupportsConstraintKeywordWithoutName returns true if the dialect supports
// CONSTRAINT keyword without a name in table constraint definitions.
func (d *ClickHouseDialect) SupportsConstraintKeywordWithoutName() bool {
	return false
}

// SupportsKeyColumnOption returns true if the dialect supports the KEY keyword
// as part of column-level constraints.
func (d *ClickHouseDialect) SupportsKeyColumnOption() bool {
	return false
}

// SupportsDataTypeSignedSuffix returns true if the dialect allows an optional
// SIGNED suffix after integer data types.
func (d *ClickHouseDialect) SupportsDataTypeSignedSuffix() bool {
	return false
}

// SupportsNestedComments returns true if the dialect supports nested comments.
// Supported since 2020.
// See https://clickhouse.com/docs/whats-new/changelog/2020#backward-incompatible-change-2
func (d *ClickHouseDialect) SupportsNestedComments() bool {
	return true
}

// SupportsMultilineCommentHints returns true if the dialect supports optimizer
// hints in multiline comments (e.g., /*!50110 KEY_BLOCK_SIZE = 1024*/).
func (d *ClickHouseDialect) SupportsMultilineCommentHints() bool {
	return false
}

// SupportsCommentOptimizerHint returns true if the dialect supports query
// optimizer hints in single and multi-line comments.
func (d *ClickHouseDialect) SupportsCommentOptimizerHint() bool {
	return false
}

// SupportsCommentOn returns true if the dialect supports the COMMENT statement.
func (d *ClickHouseDialect) SupportsCommentOn() bool {
	return false
}

// RequiresSingleLineCommentWhitespace returns true if the dialect requires
// a whitespace character after -- to start a single line comment.
func (d *ClickHouseDialect) RequiresSingleLineCommentWhitespace() bool {
	return false
}

// SupportsExplainWithUtilityOptions returns true if the dialect supports
// EXPLAIN statements with utility options.
func (d *ClickHouseDialect) SupportsExplainWithUtilityOptions() bool {
	return false
}

// SupportsExecuteImmediate returns true if the dialect supports EXECUTE IMMEDIATE statements.
func (d *ClickHouseDialect) SupportsExecuteImmediate() bool {
	return false
}

// AllowExtractCustom returns true if the dialect allows the EXTRACT function
// to use words other than keywords.
func (d *ClickHouseDialect) AllowExtractCustom() bool {
	return false
}

// AllowExtractSingleQuotes returns true if the dialect allows the EXTRACT
// function to use single quotes in the part being extracted.
func (d *ClickHouseDialect) AllowExtractSingleQuotes() bool {
	return false
}

// SupportsExtractCommaSyntax returns true if the dialect supports EXTRACT
// function with a comma separator instead of FROM.
func (d *ClickHouseDialect) SupportsExtractCommaSyntax() bool {
	return false
}

// SupportsSubqueryAsFunctionArg returns true if the dialect supports a subquery
// passed to a function as the only argument without enclosing parentheses.
func (d *ClickHouseDialect) SupportsSubqueryAsFunctionArg() bool {
	return true
}

// SupportsDollarPlaceholder returns true if the dialect allows dollar placeholders (e.g., $var).
func (d *ClickHouseDialect) SupportsDollarPlaceholder() bool {
	return false
}

// SupportsCreateIndexWithClause returns true if the dialect supports WITH clause
// in CREATE INDEX statement.
func (d *ClickHouseDialect) SupportsCreateIndexWithClause() bool {
	return false
}

// RequireIntervalQualifier returns true if the dialect requires units
// (qualifiers) to be specified in INTERVAL expressions.
func (d *ClickHouseDialect) RequireIntervalQualifier() bool {
	return true
}

// SupportsIntervalOptions returns true if the dialect supports INTERVAL data
// type with Postgres-style options.
func (d *ClickHouseDialect) SupportsIntervalOptions() bool {
	return false
}

// SupportsFactorialOperator returns true if the dialect supports a!
// expressions for factorial.
func (d *ClickHouseDialect) SupportsFactorialOperator() bool {
	return false
}

// SupportsBitwiseShiftOperators returns true if the dialect supports
// << and >> shift operators.
func (d *ClickHouseDialect) SupportsBitwiseShiftOperators() bool {
	return false
}

// SupportsNotnullOperator returns true if the dialect supports the x NOTNULL
// operator expression.
func (d *ClickHouseDialect) SupportsNotnullOperator() bool {
	return false
}

// SupportsBangNotOperator returns true if the dialect supports !a syntax
// for boolean NOT expressions.
func (d *ClickHouseDialect) SupportsBangNotOperator() bool {
	return false
}

// SupportsDoubleAmpersandOperator returns true if the dialect considers
// the && operator as a boolean AND operator.
func (d *ClickHouseDialect) SupportsDoubleAmpersandOperator() bool {
	return false
}

// SupportsMatchAgainst returns true if the dialect supports MATCH() AGAINST() syntax.
func (d *ClickHouseDialect) SupportsMatchAgainst() bool {
	return false
}

// SupportsUserHostGrantee returns true if the dialect supports MySQL-style
// 'user'@'host' grantee syntax.
func (d *ClickHouseDialect) SupportsUserHostGrantee() bool {
	return false
}

// SupportsListenNotify returns true if the dialect supports LISTEN, UNLISTEN,
// and NOTIFY statements.
func (d *ClickHouseDialect) SupportsListenNotify() bool {
	return false
}

// SupportsLoadData returns true if the dialect supports LOAD DATA statement.
func (d *ClickHouseDialect) SupportsLoadData() bool {
	return false
}

// SupportsLoadExtension returns true if the dialect supports LOAD extension statement.
func (d *ClickHouseDialect) SupportsLoadExtension() bool {
	return false
}

// SupportsTopBeforeDistinct returns true if the dialect expects the TOP option
// before the ALL/DISTINCT options in a SELECT statement.
func (d *ClickHouseDialect) SupportsTopBeforeDistinct() bool {
	return false
}

// SupportsBooleanLiterals returns true if the dialect supports boolean
// literals (true and false).
func (d *ClickHouseDialect) SupportsBooleanLiterals() bool {
	return true
}

// SupportsShowLikeBeforeIn returns true if the dialect supports the LIKE
// option in a SHOW statement before the IN option.
func (d *ClickHouseDialect) SupportsShowLikeBeforeIn() bool {
	return false
}

// SupportsPartiQL returns true if the dialect supports PartiQL for querying
// semi-structured data.
func (d *ClickHouseDialect) SupportsPartiQL() bool {
	return false
}

// SupportsEqAliasAssignment returns true if the dialect supports treating
// the equals operator = within a SelectItem as an alias assignment operator.
func (d *ClickHouseDialect) SupportsEqAliasAssignment() bool {
	return false
}

// SupportsInsertSet returns true if the dialect supports INSERT INTO ... SET syntax.
func (d *ClickHouseDialect) SupportsInsertSet() bool {
	return false
}

// SupportsInsertTableFunction returns true if the dialect supports table
// function in insertion.
func (d *ClickHouseDialect) SupportsInsertTableFunction() bool {
	return true
}

// SupportsInsertTableQuery returns true if the dialect supports table queries
// in insertion (e.g., SELECT INTO (<query>) ...).
func (d *ClickHouseDialect) SupportsInsertTableQuery() bool {
	return false
}

// SupportsInsertFormat returns true if the dialect supports insert formats
// (e.g., INSERT INTO ... FORMAT <format>).
func (d *ClickHouseDialect) SupportsInsertFormat() bool {
	return true
}

// SupportsInsertTableAlias returns true if the dialect supports INSERT INTO
// table [[AS] alias] syntax.
func (d *ClickHouseDialect) SupportsInsertTableAlias() bool {
	return false
}

// SupportsAlterColumnTypeUsing returns true if the dialect supports the USING
// clause in an ALTER COLUMN statement.
func (d *ClickHouseDialect) SupportsAlterColumnTypeUsing() bool {
	return false
}

// SupportsCommaSeparatedDropColumnList returns true if the dialect supports
// ALTER TABLE tbl DROP COLUMN c1, ..., cn.
func (d *ClickHouseDialect) SupportsCommaSeparatedDropColumnList() bool {
	return false
}

// SupportsRenameConstraint returns false for ClickHouseDialect.
func (d *ClickHouseDialect) SupportsRenameConstraint() bool {
	return false
}

// SupportsOrderByAll returns true if the dialect supports ORDER BY ALL.
// See https://clickhouse.com/docs/en/sql-reference/statements/select/order-by
func (d *ClickHouseDialect) SupportsOrderByAll() bool {
	return true
}

// SupportsGeometricTypes returns true if the dialect supports geometric types
// (Postgres geometric operations).
func (d *ClickHouseDialect) SupportsGeometricTypes() bool {
	return false
}

// DescribeRequiresTableKeyword returns true if the dialect requires the TABLE
// keyword after DESCRIBE.
func (d *ClickHouseDialect) DescribeRequiresTableKeyword() bool {
	return true
}

// SupportsOptimizeTable returns true if the dialect supports OPTIMIZE TABLE.
// See https://clickhouse.com/docs/en/sql-reference/statements/optimize
func (d *ClickHouseDialect) SupportsOptimizeTable() bool {
	return true
}

// SupportsPrewhere returns true if the dialect supports PREWHERE clause.
// See https://clickhouse.com/docs/en/sql-reference/statements/select/prewhere
func (d *ClickHouseDialect) SupportsPrewhere() bool {
	return true
}

// SupportsWithFill returns true if the dialect supports WITH FILL clause.
// See https://clickhouse.com/docs/en/sql-reference/statements/select/order-by#order-by-expr-with-fill-modifier
func (d *ClickHouseDialect) SupportsWithFill() bool {
	return true
}

// SupportsLimitBy returns true if the dialect supports LIMIT BY clause.
// See https://clickhouse.com/docs/en/sql-reference/statements/select/limit-by
func (d *ClickHouseDialect) SupportsLimitBy() bool {
	return true
}

// SupportsInterpolate returns true if the dialect supports INTERPOLATE clause.
// See https://clickhouse.com/docs/en/sql-reference/statements/select/order-by#order-by-expr-with-fill-modifier
func (d *ClickHouseDialect) SupportsInterpolate() bool {
	return true
}

// SupportsSettings returns true if the dialect supports SETTINGS clause.
// See https://clickhouse.com/docs/en/sql-reference/statements/select#settings-in-select-query
func (d *ClickHouseDialect) SupportsSettings() bool {
	return true
}

// SupportsSelectFormat returns true if the dialect supports FORMAT clause in SELECT.
// See https://clickhouse.com/docs/en/sql-reference/statements/select/format
func (d *ClickHouseDialect) SupportsSelectFormat() bool {
	return true
}

// SupportsInstall returns true if the dialect supports INSTALL statement.
func (d *ClickHouseDialect) SupportsInstall() bool {
	return false
}

// SupportsDetach returns true if the dialect supports DETACH statement.
func (d *ClickHouseDialect) SupportsDetach() bool {
	return false
}

// SupportsCommaSeparatedTrim returns true if the dialect supports the two-argument
// comma-separated form of TRIM function: TRIM(expr, characters).
func (d *ClickHouseDialect) SupportsCommaSeparatedTrim() bool {
	return true
}

// PrecValue returns the precedence value for a given Precedence level.
func (d *ClickHouseDialect) PrecValue(prec dialects.Precedence) uint8 {
	return uint8(prec)
}

// PrecUnknown returns the precedence when precedence is otherwise unknown.
func (d *ClickHouseDialect) PrecUnknown() uint8 {
	return 0
}

// GetNextPrecedence returns the dialect-specific precedence override for the next token.
func (d *ClickHouseDialect) GetNextPrecedence(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// GetNextPrecedenceDefault implements the default precedence logic.
func (d *ClickHouseDialect) GetNextPrecedenceDefault(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// ParsePrefix is a dialect-specific prefix parser override.
func (d *ClickHouseDialect) ParsePrefix(parser dialects.ParserAccessor) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseInfix is a dialect-specific infix parser override.
func (d *ClickHouseDialect) ParseInfix(parser dialects.ParserAccessor, expr ast.Expr, precedence uint8) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseStatement is a dialect-specific statement parser override.
func (d *ClickHouseDialect) ParseStatement(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	return nil, false, nil
}

// ParseColumnOption is a dialect-specific column option parser override.
func (d *ClickHouseDialect) ParseColumnOption(parser dialects.ParserAccessor) (dialects.ColumnOption, bool, error) {
	return nil, false, nil
}
