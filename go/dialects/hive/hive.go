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

package hive

import (
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/token"
)

// HiveDialect is a dialect for Apache Hive SQL implementation.
// See https://hive.apache.org/
type HiveDialect struct{}

// NewHiveDialect creates a new instance of HiveDialect.
func NewHiveDialect() *HiveDialect {
	return &HiveDialect{}
}

// Dialect returns the dialect identifier.
func (d *HiveDialect) Dialect() string {
	return "hive"
}

// IsIdentifierStart returns true if the character is a valid start character
// for an unquoted identifier.
func (d *HiveDialect) IsIdentifierStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') || ch == '$'
}

// IsIdentifierPart returns true if the character is a valid unquoted
// identifier character (not necessarily at the start).
func (d *HiveDialect) IsIdentifierPart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_' || ch == '$' || ch == '{' || ch == '}'
}

// IsDelimitedIdentifierStart returns true if the character starts a quoted identifier.
func (d *HiveDialect) IsDelimitedIdentifierStart(ch rune) bool {
	return ch == '"' || ch == '`'
}

// IsNestedDelimitedIdentifierStart returns true if the character starts a
// potential nested quoted identifier.
func (d *HiveDialect) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return false
}

// PeekNestedDelimitedIdentifierQuotes returns the outer quote style and
// optional inner quote style if applicable.
func (d *HiveDialect) PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*dialects.NestedIdentifierQuote, int) {
	return nil, pos
}

// IdentifierQuoteStyle returns the character used to quote identifiers.
func (d *HiveDialect) IdentifierQuoteStyle(identifier string) *rune {
	return nil
}

// IsCustomOperatorPart returns true if the character is part of a custom operator.
func (d *HiveDialect) IsCustomOperatorPart(ch rune) bool {
	return false
}

// IsReservedForIdentifier returns true if the keyword is reserved.
func (d *HiveDialect) IsReservedForIdentifier(kw token.Keyword) bool {
	return token.IsReservedForIdentifier(kw)
}

// GetReservedKeywordsForSelectItemOperator returns reserved keywords for select items.
func (d *HiveDialect) GetReservedKeywordsForSelectItemOperator() []token.Keyword {
	return nil
}

// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
func (d *HiveDialect) GetReservedGranteesTypes() []dialects.GranteesType {
	return nil
}

// IsColumnAlias returns true if the keyword should be parsed as a column alias.
func (d *HiveDialect) IsColumnAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == "AS" || token.IsReservedForColumnAlias(kw)
}

// IsSelectItemAlias returns true if the specified keyword should be parsed as a select item alias.
func (d *HiveDialect) IsSelectItemAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsColumnAlias(kw, parser)
}

// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier.
func (d *HiveDialect) IsTableFactor(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == "TABLE" || kw == "LATERAL"
}

// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias.
func (d *HiveDialect) IsTableAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return !token.IsReservedForTableAlias(kw)
}

// IsTableFactorAlias returns true if the specified keyword should be parsed as a table factor alias.
func (d *HiveDialect) IsTableFactorAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return explicit || d.IsTableAlias(kw, parser)
}

// IsIdentifierGeneratingFunctionName returns true if the ident is a function that generates identifiers.
func (d *HiveDialect) IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool {
	return false
}

// SupportsStringLiteralBackslashEscape returns true if the dialect supports
// escaping characters via '\' in string literals.
func (d *HiveDialect) SupportsStringLiteralBackslashEscape() bool {
	return true
}

// IgnoresWildcardEscapes returns true if the dialect strips the backslash
// when escaping LIKE wildcards (%, _).
func (d *HiveDialect) IgnoresWildcardEscapes() bool {
	return true
}

// SupportsUnicodeStringLiteral returns true if the dialect supports string
// literals with U& prefix for Unicode code points.
func (d *HiveDialect) SupportsUnicodeStringLiteral() bool {
	return false
}

// SupportsTripleQuotedString returns true if the dialect supports triple
// quoted string literals (e.g., """abc""").
func (d *HiveDialect) SupportsTripleQuotedString() bool {
	return false
}

// SupportsStringLiteralConcatenation returns true if the dialect supports
// concatenating string literals (e.g., 'Hello ' 'world').
func (d *HiveDialect) SupportsStringLiteralConcatenation() bool {
	return true
}

// SupportsStringLiteralConcatenationWithNewline returns true if the dialect
// supports concatenating string literals with a newline between them.
func (d *HiveDialect) SupportsStringLiteralConcatenationWithNewline() bool {
	return true
}

// SupportsQuoteDelimitedString returns true if the dialect supports
// quote-delimited string literals (e.g., Q'{...}') for Oracle-style strings.
func (d *HiveDialect) SupportsQuoteDelimitedString() bool {
	return false
}

// SupportsStringEscapeConstant returns true if the dialect supports the
// E'...' syntax for string literals with escape sequences (Postgres).
func (d *HiveDialect) SupportsStringEscapeConstant() bool {
	return false
}

// SupportsFilterDuringAggregation returns true if the dialect supports
// FILTER (WHERE expr) for aggregate queries.
func (d *HiveDialect) SupportsFilterDuringAggregation() bool {
	return true
}

// SupportsWithinAfterArrayAggregation returns true if the dialect supports
// ARRAY_AGG() [WITHIN GROUP (ORDER BY)] expressions.
func (d *HiveDialect) SupportsWithinAfterArrayAggregation() bool {
	return false
}

// SupportsWindowClauseNamedWindowReference returns true if the dialect
// supports referencing another named window within a window clause declaration.
func (d *HiveDialect) SupportsWindowClauseNamedWindowReference() bool {
	return false
}

// SupportsWindowFunctionNullTreatmentArg returns true if the dialect supports
// specifying null treatment as part of a window function's parameter list.
func (d *HiveDialect) SupportsWindowFunctionNullTreatmentArg() bool {
	return false
}

// SupportsMatchRecognize returns true if the dialect supports the MATCH_RECOGNIZE operation.
func (d *HiveDialect) SupportsMatchRecognize() bool {
	return false
}

// SupportsGroupByExpr returns true if the dialect supports GROUP BY expressions
// like GROUPING SETS, ROLLUP, or CUBE.
func (d *HiveDialect) SupportsGroupByExpr() bool {
	return false
}

// SupportsGroupByWithModifier returns true if the dialect supports GROUP BY
// modifiers prefixed by a WITH keyword (e.g., GROUP BY value WITH ROLLUP).
// See https://cwiki.apache.org/confluence/pages/viewpage.action?pageId=30151323#EnhancedAggregation,Cube,GroupingandRollup-CubesandRollupsr
func (d *HiveDialect) SupportsGroupByWithModifier() bool {
	return true
}

// SupportsLeftAssociativeJoinsWithoutParens returns true if the dialect supports
// left-associative join parsing by default when parentheses are omitted in nested joins.
func (d *HiveDialect) SupportsLeftAssociativeJoinsWithoutParens() bool {
	return false
}

// SupportsOuterJoinOperator returns true if the dialect supports the (+)
// syntax for OUTER JOIN (Oracle-style).
func (d *HiveDialect) SupportsOuterJoinOperator() bool {
	return false
}

// SupportsCrossJoinConstraint returns true if the dialect supports a join
// specification on CROSS JOIN.
func (d *HiveDialect) SupportsCrossJoinConstraint() bool {
	return false
}

// SupportsConnectBy returns true if the dialect supports CONNECT BY for
// hierarchical queries (Oracle-style).
func (d *HiveDialect) SupportsConnectBy() bool {
	return false
}

// SupportsStartTransactionModifier returns true if the dialect supports
// BEGIN {DEFERRED | IMMEDIATE | EXCLUSIVE | TRY | CATCH} [TRANSACTION].
func (d *HiveDialect) SupportsStartTransactionModifier() bool {
	return false
}

// SupportsEndTransactionModifier returns true if the dialect supports
// END {TRY | CATCH} statements.
func (d *HiveDialect) SupportsEndTransactionModifier() bool {
	return false
}

// SupportsNamedFnArgsWithEqOperator returns true if the dialect supports
// named arguments of the form FUN(a = '1', b = '2').
func (d *HiveDialect) SupportsNamedFnArgsWithEqOperator() bool {
	return false
}

// SupportsNamedFnArgsWithColonOperator returns true if the dialect supports
// named arguments of the form FUN(a : '1', b : '2').
func (d *HiveDialect) SupportsNamedFnArgsWithColonOperator() bool {
	return false
}

// SupportsNamedFnArgsWithAssignmentOperator returns true if the dialect supports
// named arguments of the form FUN(a := '1', b := '2').
func (d *HiveDialect) SupportsNamedFnArgsWithAssignmentOperator() bool {
	return false
}

// SupportsNamedFnArgsWithRArrowOperator returns true if the dialect supports
// named arguments of the form FUN(a => '1', b => '2').
func (d *HiveDialect) SupportsNamedFnArgsWithRArrowOperator() bool {
	return false
}

// SupportsNamedFnArgsWithExprName returns true if the dialect supports
// argument name as arbitrary expression.
func (d *HiveDialect) SupportsNamedFnArgsWithExprName() bool {
	return false
}

// SupportsParenthesizedSetVariables returns true if the dialect supports
// multiple variable assignment using parentheses in a SET variable declaration.
func (d *HiveDialect) SupportsParenthesizedSetVariables() bool {
	return false
}

// SupportsCommaSeparatedSetAssignments returns true if the dialect supports
// multiple SET statements in a single statement.
func (d *HiveDialect) SupportsCommaSeparatedSetAssignments() bool {
	return false
}

// SupportsSetStmtWithoutOperator returns true if the dialect supports
// SET statements without an explicit assignment operator (e.g., SET SHOWPLAN_XML ON).
func (d *HiveDialect) SupportsSetStmtWithoutOperator() bool {
	return false
}

// SupportsSetNames returns true if the dialect supports SET NAMES <charset_name> [COLLATE <collation_name>].
func (d *HiveDialect) SupportsSetNames() bool {
	return false
}

// SupportsSelectWildcardExcept returns true if the dialect supports EXCEPT
// clause following a wildcard in a select list.
func (d *HiveDialect) SupportsSelectWildcardExcept() bool {
	return false
}

// SupportsSelectWildcardExclude returns true if the dialect supports EXCLUDE
// option following a wildcard in a projection section.
func (d *HiveDialect) SupportsSelectWildcardExclude() bool {
	return false
}

// SupportsSelectExclude returns true if the dialect supports EXCLUDE as the
// last item in the projection section, not necessarily after a wildcard.
func (d *HiveDialect) SupportsSelectExclude() bool {
	return false
}

// SupportsSelectWildcardReplace returns true if the dialect supports REPLACE
// option in SELECT * wildcard expressions.
func (d *HiveDialect) SupportsSelectWildcardReplace() bool {
	return false
}

// SupportsSelectWildcardIlike returns true if the dialect supports ILIKE
// option in SELECT * wildcard expressions.
func (d *HiveDialect) SupportsSelectWildcardIlike() bool {
	return false
}

// SupportsSelectWildcardRename returns true if the dialect supports RENAME
// option in SELECT * wildcard expressions.
func (d *HiveDialect) SupportsSelectWildcardRename() bool {
	return false
}

// SupportsSelectWildcardWithAlias returns true if the dialect supports aliasing
// a wildcard select item.
func (d *HiveDialect) SupportsSelectWildcardWithAlias() bool {
	return false
}

// SupportsSelectExprStar returns true if the dialect supports wildcard expansion
// on arbitrary expressions in projections.
func (d *HiveDialect) SupportsSelectExprStar() bool {
	return false
}

// SupportsFromFirstSelect returns true if the dialect supports "FROM-first" selects.
func (d *HiveDialect) SupportsFromFirstSelect() bool {
	return false
}

// SupportsEmptyProjections returns true if the dialect supports empty projections
// in SELECT statements.
func (d *HiveDialect) SupportsEmptyProjections() bool {
	return false
}

// SupportsSelectModifiers returns true if the dialect supports MySQL-specific
// SELECT modifiers like HIGH_PRIORITY, STRAIGHT_JOIN, SQL_SMALL_RESULT, etc.
func (d *HiveDialect) SupportsSelectModifiers() bool {
	return false
}

// SupportsPipeOperator returns true if the dialect supports the pipe operator (|>).
func (d *HiveDialect) SupportsPipeOperator() bool {
	return false
}

// SupportsTrailingCommas returns true if the dialect supports trailing commas.
func (d *HiveDialect) SupportsTrailingCommas() bool {
	return false
}

// SupportsProjectionTrailingCommas returns true if the dialect supports trailing
// commas in the projection list.
func (d *HiveDialect) SupportsProjectionTrailingCommas() bool {
	return false
}

// SupportsFromTrailingCommas returns true if the dialect supports trailing commas
// in the FROM clause.
func (d *HiveDialect) SupportsFromTrailingCommas() bool {
	return false
}

// SupportsColumnDefinitionTrailingCommas returns true if the dialect supports trailing
// commas in column definitions.
func (d *HiveDialect) SupportsColumnDefinitionTrailingCommas() bool {
	return false
}

// SupportsLimitComma returns true if the dialect supports parsing LIMIT 1, 2
// as LIMIT 2 OFFSET 1.
func (d *HiveDialect) SupportsLimitComma() bool {
	return false
}

// ConvertTypeBeforeValue returns true if the dialect has a CONVERT function
// which accepts a type first and an expression second.
func (d *HiveDialect) ConvertTypeBeforeValue() bool {
	return false
}

// SupportsTryConvert returns true if the dialect supports the TRY_CONVERT function.
func (d *HiveDialect) SupportsTryConvert() bool {
	return false
}

// SupportsBinaryKwAsCast returns true if the dialect supports casting an expression
// to a binary type using the BINARY <expr> syntax.
func (d *HiveDialect) SupportsBinaryKwAsCast() bool {
	return false
}

// SupportsObjectNameDoubleDotNotation returns true if the dialect supports
// double dot notation for object names (e.g., db_name..table_name).
func (d *HiveDialect) SupportsObjectNameDoubleDotNotation() bool {
	return false
}

// SupportsNumericPrefix returns true if the dialect supports identifiers
// starting with a numeric prefix.
// See https://cwiki.apache.org/confluence/pages/viewpage.action?pageId=27362061#Tutorial-BuiltInOperators
func (d *HiveDialect) SupportsNumericPrefix() bool {
	return true
}

// SupportsNumericLiteralUnderscores returns true if the dialect supports
// numbers containing underscores (e.g., 10_000_000).
func (d *HiveDialect) SupportsNumericLiteralUnderscores() bool {
	return false
}

// SupportsInEmptyList returns true if the dialect supports (NOT) IN ()
// expressions with empty lists.
func (d *HiveDialect) SupportsInEmptyList() bool {
	return false
}

// SupportsDictionarySyntax returns true if the dialect supports defining
// structs or objects using syntax like {'x': 1, 'y': 2, 'z': 3}.
func (d *HiveDialect) SupportsDictionarySyntax() bool {
	return false
}

// SupportsMapLiteralSyntax returns true if the dialect supports defining
// objects using syntax like Map {1: 10, 2: 20}.
func (d *HiveDialect) SupportsMapLiteralSyntax() bool {
	return false
}

// SupportsStructLiteral returns true if the dialect supports STRUCT literal syntax.
func (d *HiveDialect) SupportsStructLiteral() bool {
	return false
}

// SupportsArrayLiteralSyntax returns true if the dialect supports array literal syntax.
func (d *HiveDialect) SupportsArrayLiteralSyntax() bool {
	return false
}

// SupportsLambdaFunctions returns true if the dialect supports lambda functions.
func (d *HiveDialect) SupportsLambdaFunctions() bool {
	return false
}

// SupportsCreateTableMultiSchemaInfoSources returns true if the dialect supports
// specifying multiple options in a CREATE TABLE statement.
func (d *HiveDialect) SupportsCreateTableMultiSchemaInfoSources() bool {
	return false
}

// SupportsCreateTableLikeParenthesized returns true if the dialect supports
// specifying which table to copy the schema from inside parenthesis.
func (d *HiveDialect) SupportsCreateTableLikeParenthesized() bool {
	return false
}

// SupportsCreateTableSelect returns true if the dialect supports CREATE TABLE SELECT.
func (d *HiveDialect) SupportsCreateTableSelect() bool {
	return true
}

// SupportsCreateViewCommentSyntax returns true if the dialect supports the COMMENT
// clause in CREATE VIEW statements using COMMENT = 'comment' syntax.
func (d *HiveDialect) SupportsCreateViewCommentSyntax() bool {
	return false
}

// SupportsArrayTypedefWithoutElementType returns true if the dialect supports
// ARRAY type without specifying an element type.
func (d *HiveDialect) SupportsArrayTypedefWithoutElementType() bool {
	return false
}

// SupportsArrayTypedefWithBrackets returns true if the dialect supports array
// type definition with brackets with optional size.
func (d *HiveDialect) SupportsArrayTypedefWithBrackets() bool {
	return false
}

// SupportsParensAroundTableFactor returns true if the dialect supports extra
// parentheses around lone table names or derived tables in the FROM clause.
func (d *HiveDialect) SupportsParensAroundTableFactor() bool {
	return false
}

// SupportsValuesAsTableFactor returns true if the dialect supports VALUES
// as a table factor without requiring parentheses.
func (d *HiveDialect) SupportsValuesAsTableFactor() bool {
	return false
}

// SupportsSemanticViewTableFactor returns true if the dialect supports
// SEMANTIC_VIEW() table functions.
func (d *HiveDialect) SupportsSemanticViewTableFactor() bool {
	return false
}

// SupportsTableVersioning returns true if the dialect supports querying
// historical table data by specifying which version to query.
func (d *HiveDialect) SupportsTableVersioning() bool {
	return false
}

// SupportsTableSampleBeforeAlias returns true if the dialect supports the
// TABLESAMPLE option before the table alias option.
// See https://cwiki.apache.org/confluence/display/hive/languagemanual+sampling
func (d *HiveDialect) SupportsTableSampleBeforeAlias() bool {
	return true
}

// SupportsTableHints returns true if the dialect supports table hints in the FROM clause.
func (d *HiveDialect) SupportsTableHints() bool {
	return false
}

// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
func (d *HiveDialect) SupportsIndexHints() bool {
	return false
}

// SupportsAscDescInColumnDefinition returns true if the dialect supports
// ASC and DESC in column definitions.
func (d *HiveDialect) SupportsAscDescInColumnDefinition() bool {
	return false
}

// SupportsSpaceSeparatedColumnOptions returns true if the dialect supports
// space-separated column options in CREATE TABLE statements.
func (d *HiveDialect) SupportsSpaceSeparatedColumnOptions() bool {
	return false
}

// SupportsConstraintKeywordWithoutName returns true if the dialect supports
// CONSTRAINT keyword without a name in table constraint definitions.
func (d *HiveDialect) SupportsConstraintKeywordWithoutName() bool {
	return false
}

// SupportsKeyColumnOption returns true if the dialect supports the KEY keyword
// as part of column-level constraints.
func (d *HiveDialect) SupportsKeyColumnOption() bool {
	return false
}

// SupportsDataTypeSignedSuffix returns true if the dialect allows an optional
// SIGNED suffix after integer data types.
func (d *HiveDialect) SupportsDataTypeSignedSuffix() bool {
	return false
}

// SupportsNestedComments returns true if the dialect supports nested comments.
func (d *HiveDialect) SupportsNestedComments() bool {
	return false
}

// SupportsMultilineCommentHints returns true if the dialect supports optimizer
// hints in multiline comments (e.g., /*!50110 KEY_BLOCK_SIZE = 1024*/).
func (d *HiveDialect) SupportsMultilineCommentHints() bool {
	return true
}

// SupportsCommentOptimizerHint returns true if the dialect supports query
// optimizer hints in single and multi-line comments.
func (d *HiveDialect) SupportsCommentOptimizerHint() bool {
	return false
}

// SupportsCommentOn returns true if the dialect supports the COMMENT statement.
func (d *HiveDialect) SupportsCommentOn() bool {
	return false
}

// RequiresSingleLineCommentWhitespace returns true if the dialect requires
// a whitespace character after -- to start a single line comment.
func (d *HiveDialect) RequiresSingleLineCommentWhitespace() bool {
	return false
}

// SupportsExplainWithUtilityOptions returns true if the dialect supports
// EXPLAIN statements with utility options.
func (d *HiveDialect) SupportsExplainWithUtilityOptions() bool {
	return false
}

// SupportsExecuteImmediate returns true if the dialect supports EXECUTE IMMEDIATE statements.
func (d *HiveDialect) SupportsExecuteImmediate() bool {
	return false
}

// AllowExtractCustom returns true if the dialect allows the EXTRACT function
// to use words other than keywords.
func (d *HiveDialect) AllowExtractCustom() bool {
	return false
}

// AllowExtractSingleQuotes returns true if the dialect allows the EXTRACT
// function to use single quotes in the part being extracted.
func (d *HiveDialect) AllowExtractSingleQuotes() bool {
	return false
}

// SupportsExtractCommaSyntax returns true if the dialect supports EXTRACT
// function with a comma separator instead of FROM.
func (d *HiveDialect) SupportsExtractCommaSyntax() bool {
	return false
}

// SupportsSubqueryAsFunctionArg returns true if the dialect supports a subquery
// passed to a function as the only argument without enclosing parentheses.
func (d *HiveDialect) SupportsSubqueryAsFunctionArg() bool {
	return false
}

// SupportsDollarPlaceholder returns true if the dialect allows dollar placeholders (e.g., $var).
func (d *HiveDialect) SupportsDollarPlaceholder() bool {
	return false
}

// SupportsCreateIndexWithClause returns true if the dialect supports WITH clause
// in CREATE INDEX statement.
func (d *HiveDialect) SupportsCreateIndexWithClause() bool {
	return false
}

// RequireIntervalQualifier returns true if the dialect requires units
// (qualifiers) to be specified in INTERVAL expressions.
func (d *HiveDialect) RequireIntervalQualifier() bool {
	return true
}

// SupportsIntervalOptions returns true if the dialect supports INTERVAL data
// type with Postgres-style options.
func (d *HiveDialect) SupportsIntervalOptions() bool {
	return false
}

// SupportsFactorialOperator returns true if the dialect supports a!
// expressions for factorial.
func (d *HiveDialect) SupportsFactorialOperator() bool {
	return false
}

// SupportsBitwiseShiftOperators returns true if the dialect supports
// << and >> shift operators.
func (d *HiveDialect) SupportsBitwiseShiftOperators() bool {
	return true
}

// SupportsNotnullOperator returns true if the dialect supports the x NOTNULL
// operator expression.
func (d *HiveDialect) SupportsNotnullOperator() bool {
	return false
}

// SupportsBangNotOperator returns true if the dialect supports !a syntax
// for boolean NOT expressions.
// See https://cwiki.apache.org/confluence/pages/viewpage.action?pageId=27362061#Tutorial-BuiltInOperators
func (d *HiveDialect) SupportsBangNotOperator() bool {
	return true
}

// SupportsDoubleAmpersandOperator returns true if the dialect considers
// the && operator as a boolean AND operator.
func (d *HiveDialect) SupportsDoubleAmpersandOperator() bool {
	return false
}

// SupportsMatchAgainst returns true if the dialect supports MATCH() AGAINST() syntax.
func (d *HiveDialect) SupportsMatchAgainst() bool {
	return false
}

// SupportsUserHostGrantee returns true if the dialect supports MySQL-style
// 'user'@'host' grantee syntax.
func (d *HiveDialect) SupportsUserHostGrantee() bool {
	return false
}

// SupportsListenNotify returns true if the dialect supports LISTEN, UNLISTEN,
// and NOTIFY statements.
func (d *HiveDialect) SupportsListenNotify() bool {
	return false
}

// SupportsLoadData returns true if the dialect supports LOAD DATA statement.
// See https://cwiki.apache.org/confluence/pages/viewpage.action?pageId=27362036#LanguageManualDML-Loadingfilesintotables
func (d *HiveDialect) SupportsLoadData() bool {
	return true
}

// SupportsLoadExtension returns true if the dialect supports LOAD extension statement.
func (d *HiveDialect) SupportsLoadExtension() bool {
	return false
}

// SupportsTopBeforeDistinct returns true if the dialect expects the TOP option
// before the ALL/DISTINCT options in a SELECT statement.
func (d *HiveDialect) SupportsTopBeforeDistinct() bool {
	return false
}

// SupportsBooleanLiterals returns true if the dialect supports boolean
// literals (true and false).
func (d *HiveDialect) SupportsBooleanLiterals() bool {
	return true
}

// SupportsShowLikeBeforeIn returns true if the dialect supports the LIKE
// option in a SHOW statement before the IN option.
func (d *HiveDialect) SupportsShowLikeBeforeIn() bool {
	return false
}

// SupportsPartiQL returns true if the dialect supports PartiQL for querying
// semi-structured data.
func (d *HiveDialect) SupportsPartiQL() bool {
	return false
}

// SupportsEqAliasAssignment returns true if the dialect supports treating
// the equals operator = within a SelectItem as an alias assignment operator.
func (d *HiveDialect) SupportsEqAliasAssignment() bool {
	return false
}

// SupportsInsertSet returns true if the dialect supports INSERT INTO ... SET syntax.
func (d *HiveDialect) SupportsInsertSet() bool {
	return false
}

// SupportsInsertTableFunction returns true if the dialect supports table
// function in insertion.
func (d *HiveDialect) SupportsInsertTableFunction() bool {
	return false
}

// SupportsInsertTableQuery returns true if the dialect supports table queries
// in insertion (e.g., SELECT INTO (<query>) ...).
func (d *HiveDialect) SupportsInsertTableQuery() bool {
	return false
}

// SupportsInsertFormat returns true if the dialect supports insert formats
// (e.g., INSERT INTO ... FORMAT <format>).
func (d *HiveDialect) SupportsInsertFormat() bool {
	return false
}

// SupportsInsertTableAlias returns true if the dialect supports INSERT INTO
// table [[AS] alias] syntax.
func (d *HiveDialect) SupportsInsertTableAlias() bool {
	return false
}

// SupportsAlterColumnTypeUsing returns true if the dialect supports the USING
// clause in an ALTER COLUMN statement.
func (d *HiveDialect) SupportsAlterColumnTypeUsing() bool {
	return false
}

// SupportsCommaSeparatedDropColumnList returns true if the dialect supports
// ALTER TABLE tbl DROP COLUMN c1, ..., cn.
func (d *HiveDialect) SupportsCommaSeparatedDropColumnList() bool {
	return false
}

// SupportsOrderByAll returns true if the dialect supports ORDER BY ALL.
func (d *HiveDialect) SupportsOrderByAll() bool {
	return false
}

// SupportsGeometricTypes returns true if the dialect supports geometric types
// (Postgres geometric operations).
func (d *HiveDialect) SupportsGeometricTypes() bool {
	return false
}

// DescribeRequiresTableKeyword returns true if the dialect requires the TABLE
// keyword after DESCRIBE.
func (d *HiveDialect) DescribeRequiresTableKeyword() bool {
	return false
}

// SupportsOptimizeTable returns true if the dialect supports OPTIMIZE TABLE.
func (d *HiveDialect) SupportsOptimizeTable() bool {
	return false
}

// SupportsPrewhere returns true if the dialect supports PREWHERE clause.
func (d *HiveDialect) SupportsPrewhere() bool {
	return false
}

// SupportsWithFill returns true if the dialect supports WITH FILL clause.
func (d *HiveDialect) SupportsWithFill() bool {
	return false
}

// SupportsLimitBy returns true if the dialect supports LIMIT BY clause.
func (d *HiveDialect) SupportsLimitBy() bool {
	return false
}

// SupportsInterpolate returns true if the dialect supports INTERPOLATE clause.
func (d *HiveDialect) SupportsInterpolate() bool {
	return false
}

// SupportsSettings returns true if the dialect supports SETTINGS clause.
func (d *HiveDialect) SupportsSettings() bool {
	return false
}

// SupportsSelectFormat returns true if the dialect supports FORMAT clause in SELECT.
func (d *HiveDialect) SupportsSelectFormat() bool {
	return false
}

// SupportsInstall returns true if the dialect supports INSTALL statement.
func (d *HiveDialect) SupportsInstall() bool {
	return false
}

// SupportsDetach returns true if the dialect supports DETACH statement.
func (d *HiveDialect) SupportsDetach() bool {
	return false
}

// SupportsCommaSeparatedTrim returns true if the dialect supports the two-argument
// comma-separated form of TRIM function: TRIM(expr, characters).
func (d *HiveDialect) SupportsCommaSeparatedTrim() bool {
	return false
}

// PrecValue returns the precedence value for a given Precedence level.
func (d *HiveDialect) PrecValue(prec dialects.Precedence) uint8 {
	return uint8(prec)
}

// PrecUnknown returns the precedence when precedence is otherwise unknown.
func (d *HiveDialect) PrecUnknown() uint8 {
	return 0
}

// GetNextPrecedence returns the dialect-specific precedence override for the next token.
func (d *HiveDialect) GetNextPrecedence(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// GetNextPrecedenceDefault implements the default precedence logic.
func (d *HiveDialect) GetNextPrecedenceDefault(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// ParsePrefix is a dialect-specific prefix parser override.
func (d *HiveDialect) ParsePrefix(parser dialects.ParserAccessor) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseInfix is a dialect-specific infix parser override.
func (d *HiveDialect) ParseInfix(parser dialects.ParserAccessor, expr ast.Expr, precedence uint8) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseStatement is a dialect-specific statement parser override.
func (d *HiveDialect) ParseStatement(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	return nil, false, nil
}

// ParseColumnOption is a dialect-specific column option parser override.
func (d *HiveDialect) ParseColumnOption(parser dialects.ParserAccessor) (dialects.ColumnOption, bool, error) {
	return nil, false, nil
}
