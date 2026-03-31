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

package mysql

import (
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/token"
	"github.com/user/sqlparser/tokenizer"
)

// Reserved keywords for table alias in MySQL
var reservedForTableAliasMySQL = []token.Keyword{
	"USE",
	"IGNORE",
	"FORCE",
	"STRAIGHT_JOIN",
}

// MySqlDialect is a dialect for MySQL SQL implementation.
// See https://www.mysql.com/
type MySqlDialect struct{}

// NewMySqlDialect creates a new instance of MySqlDialect.
func NewMySqlDialect() *MySqlDialect {
	return &MySqlDialect{}
}

// Dialect returns the dialect identifier.
func (d *MySqlDialect) Dialect() string {
	return "mysql"
}

// IsIdentifierStart returns true if the character is a valid start character
// for an unquoted identifier.
// See https://dev.mysql.com/doc/refman/8.0/en/identifiers.html.
func (d *MySqlDialect) IsIdentifierStart(ch rune) bool {
	// MySQL also implements non-ascii utf-8 characters
	// Identifiers which begin with a digit are recognized while tokenizing numbers,
	// so they can be distinguished from exponent numeric literals.
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		ch == '_' ||
		ch == '$' ||
		ch == '@' ||
		(ch >= '\u0080' && ch <= '\uffff') ||
		!isASCII(ch)
}

// IsIdentifierPart returns true if the character is a valid unquoted
// identifier character (not necessarily at the start).
func (d *MySqlDialect) IsIdentifierPart(ch rune) bool {
	return d.IsIdentifierStart(ch) ||
		(ch >= '0' && ch <= '9') ||
		// MySQL implements Unicode characters in identifiers.
		!isASCII(ch)
}

// IsDelimitedIdentifierStart returns true if the character starts a quoted identifier.
func (d *MySqlDialect) IsDelimitedIdentifierStart(ch rune) bool {
	return ch == '`'
}

// IsNestedDelimitedIdentifierStart returns true if the character starts a
// potential nested quoted identifier.
func (d *MySqlDialect) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return false
}

// PeekNestedDelimitedIdentifierQuotes returns the outer quote style and
// optional inner quote style if applicable.
func (d *MySqlDialect) PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*dialects.NestedIdentifierQuote, int) {
	return nil, pos
}

// IdentifierQuoteStyle returns the character used to quote identifiers.
func (d *MySqlDialect) IdentifierQuoteStyle(identifier string) *rune {
	quote := '`'
	return &quote
}

// IsCustomOperatorPart returns true if the character is part of a custom operator.
func (d *MySqlDialect) IsCustomOperatorPart(ch rune) bool {
	return false
}

// isASCII checks if a rune is an ASCII character.
func isASCII(ch rune) bool {
	return ch >= 0 && ch <= 127
}

// IsReservedForIdentifier returns true if the keyword is reserved.
func (d *MySqlDialect) IsReservedForIdentifier(kw token.Keyword) bool {
	return token.IsReservedForIdentifier(kw)
}

// GetReservedKeywordsForSelectItemOperator returns reserved keywords for select items.
func (d *MySqlDialect) GetReservedKeywordsForSelectItemOperator() []token.Keyword {
	return nil
}

// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
func (d *MySqlDialect) GetReservedGranteesTypes() []dialects.GranteesType {
	return nil
}

// IsColumnAlias returns true if the keyword should be parsed as a column alias.
func (d *MySqlDialect) IsColumnAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == "AS" || token.IsReservedForColumnAlias(kw)
}

// IsSelectItemAlias returns true if the specified keyword should be parsed as a select item alias
func (d *MySqlDialect) IsSelectItemAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsColumnAlias(kw, parser)
}

// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier
func (d *MySqlDialect) IsTableFactor(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == "TABLE" || kw == "LATERAL"
}

// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias
func (d *MySqlDialect) IsTableAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return !token.IsReservedForTableAlias(kw)
}

// IsTableFactorAlias returns true if the specified keyword should be parsed as a table factor alias
func (d *MySqlDialect) IsTableFactorAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return explicit ||
		(!token.IsReservedForTableAlias(kw) &&
			!isReservedForTableAliasMySQL(kw))
}

// isReservedForTableAliasMySQL checks if a keyword is reserved for table alias in MySQL.
func isReservedForTableAliasMySQL(kw token.Keyword) bool {
	for _, reserved := range reservedForTableAliasMySQL {
		if kw == reserved {
			return true
		}
	}
	return false
}

// IsIdentifierGeneratingFunctionName returns true if the ident is a function that generates identifiers.
func (d *MySqlDialect) IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool {
	return false
}

// SupportsStringLiteralBackslashEscape returns true if the dialect supports
// escaping characters via '\' in string literals.
// See https://dev.mysql.com/doc/refman/8.0/en/string-literals.html#character-escape-sequences
func (d *MySqlDialect) SupportsStringLiteralBackslashEscape() bool {
	return true
}

// IgnoresWildcardEscapes returns true if the dialect strips the backslash
// when escaping LIKE wildcards (%, _).
func (d *MySqlDialect) IgnoresWildcardEscapes() bool {
	return true
}

// SupportsUnicodeStringLiteral returns true if the dialect supports string
// literals with U& prefix for Unicode code points.
func (d *MySqlDialect) SupportsUnicodeStringLiteral() bool {
	return false
}

// SupportsTripleQuotedString returns true if the dialect supports triple
// quoted string literals (e.g., """abc""").
func (d *MySqlDialect) SupportsTripleQuotedString() bool {
	return false
}

// SupportsStringLiteralConcatenation returns true if the dialect supports
// concatenating string literals (e.g., 'Hello ' 'world').
// See https://dev.mysql.com/doc/refman/8.4/en/string-functions.html#function_concat
func (d *MySqlDialect) SupportsStringLiteralConcatenation() bool {
	return true
}

// SupportsStringLiteralConcatenationWithNewline returns true if the dialect
// supports concatenating string literals with a newline between them.
func (d *MySqlDialect) SupportsStringLiteralConcatenationWithNewline() bool {
	return false
}

// SupportsQuoteDelimitedString returns true if the dialect supports
// quote-delimited string literals (e.g., Q'{...}') for Oracle-style strings.
func (d *MySqlDialect) SupportsQuoteDelimitedString() bool {
	return false
}

// SupportsStringEscapeConstant returns true if the dialect supports the
// E'...' syntax for string literals with escape sequences (Postgres).
func (d *MySqlDialect) SupportsStringEscapeConstant() bool {
	return false
}

// SupportsFilterDuringAggregation returns true if the dialect supports
// FILTER (WHERE expr) for aggregate queries.
func (d *MySqlDialect) SupportsFilterDuringAggregation() bool {
	return false
}

// SupportsWithinAfterArrayAggregation returns true if the dialect supports
// ARRAY_AGG() [WITHIN GROUP (ORDER BY)] expressions.
func (d *MySqlDialect) SupportsWithinAfterArrayAggregation() bool {
	return false
}

// SupportsWindowClauseNamedWindowReference returns true if the dialect
// supports referencing another named window within a window clause declaration.
func (d *MySqlDialect) SupportsWindowClauseNamedWindowReference() bool {
	return true
}

// SupportsWindowFunctionNullTreatmentArg returns true if the dialect supports
// specifying null treatment as part of a window function's parameter list.
func (d *MySqlDialect) SupportsWindowFunctionNullTreatmentArg() bool {
	return true
}

// SupportsMatchRecognize returns true if the dialect supports the MATCH_RECOGNIZE operation.
func (d *MySqlDialect) SupportsMatchRecognize() bool {
	return false
}

// SupportsGroupByExpr returns true if the dialect supports GROUP BY expressions
// like GROUPING SETS, ROLLUP, or CUBE.
func (d *MySqlDialect) SupportsGroupByExpr() bool {
	return true
}

// SupportsGroupByWithModifier returns true if the dialect supports GROUP BY
// modifiers prefixed by a WITH keyword (e.g., GROUP BY value WITH ROLLUP).
func (d *MySqlDialect) SupportsGroupByWithModifier() bool {
	return true
}

// SupportsLeftAssociativeJoinsWithoutParens returns true if the dialect supports
// left-associative join parsing by default when parentheses are omitted in nested joins.
func (d *MySqlDialect) SupportsLeftAssociativeJoinsWithoutParens() bool {
	return true
}

// SupportsOuterJoinOperator returns true if the dialect supports the (+)
// syntax for OUTER JOIN (Oracle-style).
func (d *MySqlDialect) SupportsOuterJoinOperator() bool {
	return false
}

// SupportsCrossJoinConstraint returns true if the dialect supports a join
// specification on CROSS JOIN.
func (d *MySqlDialect) SupportsCrossJoinConstraint() bool {
	return true
}

// SupportsConnectBy returns true if the dialect supports CONNECT BY for
// hierarchical queries (Oracle-style).
func (d *MySqlDialect) SupportsConnectBy() bool {
	return false
}

// SupportsStartTransactionModifier returns true if the dialect supports
// BEGIN {DEFERRED | IMMEDIATE | EXCLUSIVE | TRY | CATCH} [TRANSACTION].
func (d *MySqlDialect) SupportsStartTransactionModifier() bool {
	return false
}

// SupportsEndTransactionModifier returns true if the dialect supports
// END {TRY | CATCH} statements.
func (d *MySqlDialect) SupportsEndTransactionModifier() bool {
	return false
}

// SupportsNamedFnArgsWithEqOperator returns true if the dialect supports
// named arguments of the form FUN(a = '1', b = '2').
func (d *MySqlDialect) SupportsNamedFnArgsWithEqOperator() bool {
	return false
}

// SupportsNamedFnArgsWithColonOperator returns true if the dialect supports
// named arguments of the form FUN(a : '1', b : '2').
func (d *MySqlDialect) SupportsNamedFnArgsWithColonOperator() bool {
	return false
}

// SupportsNamedFnArgsWithAssignmentOperator returns true if the dialect supports
// named arguments of the form FUN(a := '1', b := '2').
func (d *MySqlDialect) SupportsNamedFnArgsWithAssignmentOperator() bool {
	return false
}

// SupportsNamedFnArgsWithRArrowOperator returns true if the dialect supports
// named arguments of the form FUN(a => '1', b => '2').
func (d *MySqlDialect) SupportsNamedFnArgsWithRArrowOperator() bool {
	return false
}

// SupportsNamedFnArgsWithExprName returns true if the dialect supports
// argument name as arbitrary expression.
func (d *MySqlDialect) SupportsNamedFnArgsWithExprName() bool {
	return false
}

// SupportsParenthesizedSetVariables returns true if the dialect supports
// multiple variable assignment using parentheses in a SET variable declaration.
func (d *MySqlDialect) SupportsParenthesizedSetVariables() bool {
	return false
}

// SupportsCommaSeparatedSetAssignments returns true if the dialect supports
// multiple SET statements in a single statement.
func (d *MySqlDialect) SupportsCommaSeparatedSetAssignments() bool {
	return true
}

// SupportsSetStmtWithoutOperator returns true if the dialect supports
// SET statements without an explicit assignment operator (e.g., SET SHOWPLAN_XML ON).
func (d *MySqlDialect) SupportsSetStmtWithoutOperator() bool {
	return false
}

// SupportsSetNames returns true if the dialect supports SET NAMES <charset_name> [COLLATE <collation_name>].
func (d *MySqlDialect) SupportsSetNames() bool {
	return true
}

// SupportsSelectWildcardExcept returns true if the dialect supports EXCEPT
// clause following a wildcard in a select list.
func (d *MySqlDialect) SupportsSelectWildcardExcept() bool {
	return false
}

// SupportsSelectWildcardExclude returns true if the dialect supports EXCLUDE
// option following a wildcard in a projection section.
func (d *MySqlDialect) SupportsSelectWildcardExclude() bool {
	return false
}

// SupportsSelectExclude returns true if the dialect supports EXCLUDE as the
// last item in the projection section, not necessarily after a wildcard.
func (d *MySqlDialect) SupportsSelectExclude() bool {
	return false
}

// SupportsSelectWildcardReplace returns true if the dialect supports REPLACE
// option in SELECT * wildcard expressions.
func (d *MySqlDialect) SupportsSelectWildcardReplace() bool {
	return false
}

// SupportsSelectWildcardIlike returns true if the dialect supports ILIKE
// option in SELECT * wildcard expressions.
func (d *MySqlDialect) SupportsSelectWildcardIlike() bool {
	return false
}

// SupportsSelectWildcardRename returns true if the dialect supports RENAME
// option in SELECT * wildcard expressions.
func (d *MySqlDialect) SupportsSelectWildcardRename() bool {
	return false
}

// SupportsSelectWildcardWithAlias returns true if the dialect supports aliasing
// a wildcard select item.
func (d *MySqlDialect) SupportsSelectWildcardWithAlias() bool {
	return false
}

// SupportsSelectExprStar returns true if the dialect supports wildcard expansion
// on arbitrary expressions in projections.
func (d *MySqlDialect) SupportsSelectExprStar() bool {
	return false
}

// SupportsFromFirstSelect returns true if the dialect supports "FROM-first" selects.
func (d *MySqlDialect) SupportsFromFirstSelect() bool {
	return false
}

// SupportsEmptyProjections returns true if the dialect supports empty projections
// in SELECT statements.
func (d *MySqlDialect) SupportsEmptyProjections() bool {
	return false
}

// SupportsSelectModifiers returns true if the dialect supports MySQL-specific
// SELECT modifiers like HIGH_PRIORITY, STRAIGHT_JOIN, SQL_SMALL_RESULT, etc.
func (d *MySqlDialect) SupportsSelectModifiers() bool {
	return true
}

// SupportsPipeOperator returns true if the dialect supports the pipe operator (|>).
func (d *MySqlDialect) SupportsPipeOperator() bool {
	return false
}

// SupportsTrailingCommas returns true if the dialect supports trailing commas.
func (d *MySqlDialect) SupportsTrailingCommas() bool {
	return false
}

// SupportsProjectionTrailingCommas returns true if the dialect supports trailing
// commas in the projection list.
func (d *MySqlDialect) SupportsProjectionTrailingCommas() bool {
	return false
}

// SupportsFromTrailingCommas returns true if the dialect supports trailing commas
// in the FROM clause.
func (d *MySqlDialect) SupportsFromTrailingCommas() bool {
	return false
}

// SupportsColumnDefinitionTrailingCommas returns true if the dialect supports trailing
// commas in column definitions.
func (d *MySqlDialect) SupportsColumnDefinitionTrailingCommas() bool {
	return false
}

// SupportsLimitComma returns true if the dialect supports parsing LIMIT 1, 2
// as LIMIT 2 OFFSET 1.
func (d *MySqlDialect) SupportsLimitComma() bool {
	return true
}

// ConvertTypeBeforeValue returns true if the dialect has a CONVERT function
// which accepts a type first and an expression second.
func (d *MySqlDialect) ConvertTypeBeforeValue() bool {
	return true
}

// SupportsTryConvert returns true if the dialect supports the TRY_CONVERT function.
func (d *MySqlDialect) SupportsTryConvert() bool {
	return false
}

// SupportsBinaryKwAsCast returns true if the dialect supports casting an expression
// to a binary type using the BINARY <expr> syntax.
// See https://dev.mysql.com/doc/refman/8.4/en/cast-functions.html#operator_binary
func (d *MySqlDialect) SupportsBinaryKwAsCast() bool {
	return true
}

// SupportsObjectNameDoubleDotNotation returns true if the dialect supports
// double dot notation for object names (e.g., db_name..table_name).
func (d *MySqlDialect) SupportsObjectNameDoubleDotNotation() bool {
	return false
}

// SupportsNumericPrefix returns true if the dialect supports identifiers
// starting with a numeric prefix.
func (d *MySqlDialect) SupportsNumericPrefix() bool {
	return true
}

// SupportsNumericLiteralUnderscores returns true if the dialect supports
// numbers containing underscores (e.g., 10_000_000).
func (d *MySqlDialect) SupportsNumericLiteralUnderscores() bool {
	return false
}

// SupportsInEmptyList returns true if the dialect supports (NOT) IN ()
// expressions with empty lists.
func (d *MySqlDialect) SupportsInEmptyList() bool {
	return false
}

// SupportsDictionarySyntax returns true if the dialect supports defining
// structs or objects using syntax like {'x': 1, 'y': 2, 'z': 3}.
func (d *MySqlDialect) SupportsDictionarySyntax() bool {
	return false
}

// SupportsMapLiteralSyntax returns true if the dialect supports defining
// objects using syntax like Map {1: 10, 2: 20}.
func (d *MySqlDialect) SupportsMapLiteralSyntax() bool {
	return false
}

// SupportsStructLiteral returns true if the dialect supports STRUCT literal syntax.
func (d *MySqlDialect) SupportsStructLiteral() bool {
	return false
}

// SupportsArrayLiteralSyntax returns true if the dialect supports array literal syntax.
func (d *MySqlDialect) SupportsArrayLiteralSyntax() bool {
	return false
}

// SupportsLambdaFunctions returns true if the dialect supports lambda functions.
func (d *MySqlDialect) SupportsLambdaFunctions() bool {
	return false
}

// SupportsCreateTableMultiSchemaInfoSources returns true if the dialect supports
// specifying multiple options in a CREATE TABLE statement.
func (d *MySqlDialect) SupportsCreateTableMultiSchemaInfoSources() bool {
	return false
}

// SupportsCreateTableLikeParenthesized returns true if the dialect supports
// specifying which table to copy the schema from inside parenthesis.
func (d *MySqlDialect) SupportsCreateTableLikeParenthesized() bool {
	return false
}

// SupportsCreateTableSelect returns true if the dialect supports CREATE TABLE SELECT.
// See https://dev.mysql.com/doc/refman/8.4/en/create-table-select.html
func (d *MySqlDialect) SupportsCreateTableSelect() bool {
	return true
}

// SupportsCreateViewCommentSyntax returns true if the dialect supports the COMMENT
// clause in CREATE VIEW statements using COMMENT = 'comment' syntax.
func (d *MySqlDialect) SupportsCreateViewCommentSyntax() bool {
	return false
}

// SupportsArrayTypedefWithoutElementType returns true if the dialect supports
// ARRAY type without specifying an element type.
func (d *MySqlDialect) SupportsArrayTypedefWithoutElementType() bool {
	return false
}

// SupportsArrayTypedefWithBrackets returns true if the dialect supports array
// type definition with brackets with optional size.
func (d *MySqlDialect) SupportsArrayTypedefWithBrackets() bool {
	return false
}

// SupportsParensAroundTableFactor returns true if the dialect supports extra
// parentheses around lone table names or derived tables in the FROM clause.
func (d *MySqlDialect) SupportsParensAroundTableFactor() bool {
	return false
}

// SupportsValuesAsTableFactor returns true if the dialect supports VALUES
// as a table factor without requiring parentheses.
func (d *MySqlDialect) SupportsValuesAsTableFactor() bool {
	return false
}

// SupportsSemanticViewTableFactor returns true if the dialect supports
// SEMANTIC_VIEW() table functions.
func (d *MySqlDialect) SupportsSemanticViewTableFactor() bool {
	return false
}

// SupportsTableVersioning returns true if the dialect supports querying
// historical table data by specifying which version to query.
func (d *MySqlDialect) SupportsTableVersioning() bool {
	return false
}

// SupportsTableSampleBeforeAlias returns true if the dialect supports the
// TABLESAMPLE option before the table alias option.
func (d *MySqlDialect) SupportsTableSampleBeforeAlias() bool {
	return false
}

// SupportsTableHints returns true if the dialect supports table hints in the FROM clause.
func (d *MySqlDialect) SupportsTableHints() bool {
	return true
}

// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
func (d *MySqlDialect) SupportsIndexHints() bool {
	return true
}

// SupportsAscDescInColumnDefinition returns true if the dialect supports
// ASC and DESC in column definitions.
func (d *MySqlDialect) SupportsAscDescInColumnDefinition() bool {
	return false
}

// SupportsSpaceSeparatedColumnOptions returns true if the dialect supports
// space-separated column options in CREATE TABLE statements.
func (d *MySqlDialect) SupportsSpaceSeparatedColumnOptions() bool {
	return false
}

// SupportsConstraintKeywordWithoutName returns true if the dialect supports
// CONSTRAINT keyword without a name in table constraint definitions.
// See https://dev.mysql.com/doc/refman/8.4/en/create-table.html
func (d *MySqlDialect) SupportsConstraintKeywordWithoutName() bool {
	return true
}

// SupportsKeyColumnOption returns true if the dialect supports the KEY keyword
// as part of column-level constraints.
// See https://dev.mysql.com/doc/refman/8.4/en/create-table.html
func (d *MySqlDialect) SupportsKeyColumnOption() bool {
	return true
}

// SupportsDataTypeSignedSuffix returns true if the dialect allows an optional
// SIGNED suffix after integer data types.
func (d *MySqlDialect) SupportsDataTypeSignedSuffix() bool {
	return true
}

// SupportsNestedComments returns true if the dialect supports nested comments.
func (d *MySqlDialect) SupportsNestedComments() bool {
	return false
}

// SupportsMultilineCommentHints returns true if the dialect supports optimizer
// hints in multiline comments (e.g., /*!50110 KEY_BLOCK_SIZE = 1024*/).
// See https://dev.mysql.com/doc/refman/8.4/en/comments.html
func (d *MySqlDialect) SupportsMultilineCommentHints() bool {
	return true
}

// SupportsCommentOptimizerHint returns true if the dialect supports query
// optimizer hints in single and multi-line comments.
func (d *MySqlDialect) SupportsCommentOptimizerHint() bool {
	return true
}

// SupportsCommentOn returns true if the dialect supports the COMMENT statement.
func (d *MySqlDialect) SupportsCommentOn() bool {
	return false
}

// RequiresSingleLineCommentWhitespace returns true if the dialect requires
// a whitespace character after -- to start a single line comment.
func (d *MySqlDialect) RequiresSingleLineCommentWhitespace() bool {
	return true
}

// SupportsExplainWithUtilityOptions returns true if the dialect supports
// EXPLAIN statements with utility options.
func (d *MySqlDialect) SupportsExplainWithUtilityOptions() bool {
	return false
}

// SupportsExecuteImmediate returns true if the dialect supports EXECUTE IMMEDIATE statements.
func (d *MySqlDialect) SupportsExecuteImmediate() bool {
	return false
}

// AllowExtractCustom returns true if the dialect allows the EXTRACT function
// to use words other than keywords.
func (d *MySqlDialect) AllowExtractCustom() bool {
	return false
}

// AllowExtractSingleQuotes returns true if the dialect allows the EXTRACT
// function to use single quotes in the part being extracted.
func (d *MySqlDialect) AllowExtractSingleQuotes() bool {
	return false
}

// SupportsExtractCommaSyntax returns true if the dialect supports EXTRACT
// function with a comma separator instead of FROM.
func (d *MySqlDialect) SupportsExtractCommaSyntax() bool {
	return false
}

// SupportsSubqueryAsFunctionArg returns true if the dialect supports a subquery
// passed to a function as the only argument without enclosing parentheses.
func (d *MySqlDialect) SupportsSubqueryAsFunctionArg() bool {
	return false
}

// SupportsDollarPlaceholder returns true if the dialect allows dollar placeholders (e.g., $var).
func (d *MySqlDialect) SupportsDollarPlaceholder() bool {
	return false
}

// SupportsCreateIndexWithClause returns true if the dialect supports WITH clause
// in CREATE INDEX statement.
func (d *MySqlDialect) SupportsCreateIndexWithClause() bool {
	return false
}

// RequireIntervalQualifier returns true if the dialect requires units
// (qualifiers) to be specified in INTERVAL expressions.
func (d *MySqlDialect) RequireIntervalQualifier() bool {
	return true
}

// SupportsIntervalOptions returns true if the dialect supports INTERVAL data
// type with Postgres-style options.
func (d *MySqlDialect) SupportsIntervalOptions() bool {
	return false
}

// SupportsFactorialOperator returns true if the dialect supports a!
// expressions for factorial.
func (d *MySqlDialect) SupportsFactorialOperator() bool {
	return false
}

// SupportsBitwiseShiftOperators returns true if the dialect supports
// << and >> shift operators.
func (d *MySqlDialect) SupportsBitwiseShiftOperators() bool {
	return true
}

// SupportsNotnullOperator returns true if the dialect supports the x NOTNULL
// operator expression.
func (d *MySqlDialect) SupportsNotnullOperator() bool {
	return false
}

// SupportsBangNotOperator returns true if the dialect supports !a syntax
// for boolean NOT expressions.
func (d *MySqlDialect) SupportsBangNotOperator() bool {
	return false
}

// SupportsDoubleAmpersandOperator returns true if the dialect considers
// the && operator as a boolean AND operator.
// See https://dev.mysql.com/doc/refman/8.4/en/expressions.html
func (d *MySqlDialect) SupportsDoubleAmpersandOperator() bool {
	return true
}

// SupportsMatchAgainst returns true if the dialect supports MATCH() AGAINST() syntax.
func (d *MySqlDialect) SupportsMatchAgainst() bool {
	return true
}

// SupportsUserHostGrantee returns true if the dialect supports MySQL-style
// 'user'@'host' grantee syntax.
func (d *MySqlDialect) SupportsUserHostGrantee() bool {
	return true
}

// SupportsListenNotify returns true if the dialect supports LISTEN, UNLISTEN,
// and NOTIFY statements.
func (d *MySqlDialect) SupportsListenNotify() bool {
	return false
}

// SupportsLoadData returns true if the dialect supports LOAD DATA statement.
func (d *MySqlDialect) SupportsLoadData() bool {
	return true
}

// SupportsLoadExtension returns true if the dialect supports LOAD extension statement.
func (d *MySqlDialect) SupportsLoadExtension() bool {
	return false
}

// SupportsTopBeforeDistinct returns true if the dialect expects the TOP option
// before the ALL/DISTINCT options in a SELECT statement.
func (d *MySqlDialect) SupportsTopBeforeDistinct() bool {
	return false
}

// SupportsBooleanLiterals returns true if the dialect supports boolean
// literals (true and false).
func (d *MySqlDialect) SupportsBooleanLiterals() bool {
	return true
}

// SupportsShowLikeBeforeIn returns true if the dialect supports the LIKE
// option in a SHOW statement before the IN option.
func (d *MySqlDialect) SupportsShowLikeBeforeIn() bool {
	return true
}

// SupportsPartiQL returns true if the dialect supports PartiQL for querying
// semi-structured data.
func (d *MySqlDialect) SupportsPartiQL() bool {
	return false
}

// SupportsEqAliasAssignment returns true if the dialect supports treating
// the equals operator = within a SelectItem as an alias assignment operator.
func (d *MySqlDialect) SupportsEqAliasAssignment() bool {
	return false
}

// SupportsInsertSet returns true if the dialect supports INSERT INTO ... SET syntax.
// See https://dev.mysql.com/doc/refman/8.4/en/insert.html
func (d *MySqlDialect) SupportsInsertSet() bool {
	return true
}

// SupportsInsertTableFunction returns true if the dialect supports table
// function in insertion.
func (d *MySqlDialect) SupportsInsertTableFunction() bool {
	return false
}

// SupportsInsertTableQuery returns true if the dialect supports table queries
// in insertion (e.g., SELECT INTO (<query>) ...).
func (d *MySqlDialect) SupportsInsertTableQuery() bool {
	return false
}

// SupportsInsertFormat returns true if the dialect supports insert formats
// (e.g., INSERT INTO ... FORMAT <format>).
func (d *MySqlDialect) SupportsInsertFormat() bool {
	return false
}

// SupportsInsertTableAlias returns true if the dialect supports INSERT INTO
// table [[AS] alias] syntax.
func (d *MySqlDialect) SupportsInsertTableAlias() bool {
	return false
}

// SupportsAlterColumnTypeUsing returns true if the dialect supports the USING
// clause in an ALTER COLUMN statement.
func (d *MySqlDialect) SupportsAlterColumnTypeUsing() bool {
	return false
}

// SupportsCommaSeparatedDropColumnList returns true if the dialect supports
// ALTER TABLE tbl DROP COLUMN c1, ..., cn.
func (d *MySqlDialect) SupportsCommaSeparatedDropColumnList() bool {
	return false
}

// SupportsOrderByAll returns true if the dialect supports ORDER BY ALL.
func (d *MySqlDialect) SupportsOrderByAll() bool {
	return false
}

// SupportsGeometricTypes returns true if the dialect supports geometric types
// (Postgres geometric operations).
func (d *MySqlDialect) SupportsGeometricTypes() bool {
	return false
}

// DescribeRequiresTableKeyword returns true if the dialect requires the TABLE
// keyword after DESCRIBE.
func (d *MySqlDialect) DescribeRequiresTableKeyword() bool {
	return false
}

// SupportsOptimizeTable returns true if the dialect supports OPTIMIZE TABLE.
func (d *MySqlDialect) SupportsOptimizeTable() bool {
	return true
}

// SupportsPrewhere returns true if the dialect supports PREWHERE clause.
func (d *MySqlDialect) SupportsPrewhere() bool {
	return false
}

// SupportsWithFill returns true if the dialect supports WITH FILL clause.
func (d *MySqlDialect) SupportsWithFill() bool {
	return false
}

// SupportsLimitBy returns true if the dialect supports LIMIT BY clause.
func (d *MySqlDialect) SupportsLimitBy() bool {
	return false
}

// SupportsInterpolate returns true if the dialect supports INTERPOLATE clause.
func (d *MySqlDialect) SupportsInterpolate() bool {
	return false
}

// SupportsSettings returns true if the dialect supports SETTINGS clause.
func (d *MySqlDialect) SupportsSettings() bool {
	return false
}

// SupportsSelectFormat returns true if the dialect supports FORMAT clause in SELECT.
func (d *MySqlDialect) SupportsSelectFormat() bool {
	return false
}

// SupportsInstall returns true if the dialect supports INSTALL statement.
func (d *MySqlDialect) SupportsInstall() bool {
	return false
}

// SupportsDetach returns true if the dialect supports DETACH statement.
func (d *MySqlDialect) SupportsDetach() bool {
	return false
}

// SupportsCommaSeparatedTrim returns true if the dialect supports the two-argument
// comma-separated form of TRIM function: TRIM(expr, characters).
func (d *MySqlDialect) SupportsCommaSeparatedTrim() bool {
	return false
}

// PrecValue returns the precedence value for a given Precedence level.
func (d *MySqlDialect) PrecValue(prec dialects.Precedence) uint8 {
	switch {
	case prec == dialects.PrecedencePeriod:
		return 100
	case prec == dialects.PrecedenceDoubleColon:
		return 50
	case prec == dialects.PrecedenceAtTz:
		return 41
	case prec == dialects.PrecedenceMulDivModOp:
		return 40
	case prec == dialects.PrecedencePlusMinus:
		return 30
	case prec == dialects.PrecedenceXor:
		return 24
	case prec == dialects.PrecedenceAmpersand:
		return 23
	case prec == dialects.PrecedenceCaret:
		return 22
	case prec == dialects.PrecedencePipe:
		return 21
	case prec == dialects.PrecedenceColon:
		return 21
	case prec == dialects.PrecedenceBetween:
		return 20
	case prec == dialects.PrecedenceEq:
		return 20
	case prec == dialects.PrecedenceLike:
		return 19
	case prec == dialects.PrecedenceIs:
		return 17
	case prec == dialects.PrecedencePgOther:
		return 16
	case prec == dialects.PrecedenceUnaryNot:
		return 15
	case prec == dialects.PrecedenceAnd:
		return 10
	case prec == dialects.PrecedenceOr:
		return 5
	default:
		return d.PrecUnknown()
	}
}

// PrecUnknown returns the precedence when precedence is otherwise unknown.
func (d *MySqlDialect) PrecUnknown() uint8 {
	return 0
}

// GetNextPrecedence returns 0 to fall back to default behavior.
func (d *MySqlDialect) GetNextPrecedence(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// GetNextPrecedenceDefault implements the default precedence logic.
func (d *MySqlDialect) GetNextPrecedenceDefault(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// ParsePrefix returns false to fall back to default behavior.
func (d *MySqlDialect) ParsePrefix(parser dialects.ParserAccessor) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseInfix is a dialect-specific infix parser override.
// Parses DIV as an operator for MySQL.
// TODO: Fix interface compatibility between ast.Expr and expr.Expr
func (d *MySqlDialect) ParseInfix(parser dialects.ParserAccessor, leftExpr ast.Expr, precedence uint8) (ast.Expr, bool, error) {
	// TODO: Implement DIV operator parsing
	// Currently disabled due to interface compatibility issues
	return nil, false, nil
}

// parseExpression attempts to parse an expression from the parser.
// This is a helper function to call the parser's expression parsing.
func parseExpression(parser dialects.ParserAccessor, precedence uint8) (ast.Expr, error) {
	// Note: In a full implementation, this would call into the parser's
	// expression parsing logic. For now, we return nil as this requires
	// access to the parser's internal state which isn't available in the
	// dialect interface.
	return nil, nil
}

// ParseStatement is a dialect-specific statement parser override.
// Handles MySQL-specific statements like LOCK TABLES and UNLOCK TABLES.
func (d *MySqlDialect) ParseStatement(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	// Check for LOCK TABLES
	if parser.PeekKeyword("LOCK") && parser.PeekNthKeyword(1, "TABLES") {
		parser.ParseKeyword("LOCK")
		parser.ParseKeyword("TABLES")
		return d.parseLockTables(parser)
	}

	// Check for UNLOCK TABLES
	if parser.PeekKeyword("UNLOCK") && parser.PeekNthKeyword(1, "TABLES") {
		parser.ParseKeyword("UNLOCK")
		parser.ParseKeyword("TABLES")
		return d.parseUnlockTables(parser)
	}

	return nil, false, nil
}

// parseLockTables parses the LOCK TABLES statement.
// https://dev.mysql.com/doc/refman/8.0/en/lock-tables.html
func (d *MySqlDialect) parseLockTables(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	tables, err := parseCommaSeparatedLockTables(parser)
	if err != nil {
		return nil, true, err
	}

	return &statement.LockTablesStmt{
		Tables: tables,
	}, true, nil
}

// parseCommaSeparatedLockTables parses a comma-separated list of lock table specifications.
func parseCommaSeparatedLockTables(parser dialects.ParserAccessor) ([]*statement.LockTable, error) {
	var tables []*statement.LockTable

	for {
		table, err := parseLockTable(parser)
		if err != nil {
			return nil, err
		}
		tables = append(tables, table)

		if !parser.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	return tables, nil
}

// parseLockTable parses a single lock table specification.
// tbl_name [[AS] alias] lock_type
func parseLockTable(parser dialects.ParserAccessor) (*statement.LockTable, error) {
	table, err := parseIdentifier(parser)
	if err != nil {
		return nil, err
	}

	alias, err := parseOptionalAlias(parser, []token.Keyword{"READ", "WRITE", "LOW_PRIORITY"})
	if err != nil {
		return nil, err
	}

	lockType, err := parseLockTablesType(parser)
	if err != nil {
		return nil, err
	}

	return &statement.LockTable{
		Table:    table,
		Alias:    alias,
		LockType: lockType,
	}, nil
}

// parseIdentifier parses an identifier from the parser.
func parseIdentifier(parser dialects.ParserAccessor) (*ast.Ident, error) {
	// This would typically call into the parser's identifier parsing
	// For now, return an empty identifier
	return &ast.Ident{}, nil
}

// parseOptionalAlias parses an optional alias after a table name.
func parseOptionalAlias(parser dialects.ParserAccessor, reservedKeywords []token.Keyword) (*ast.Ident, error) {
	// Check for AS keyword
	if parser.ParseKeyword("AS") {
		return parseIdentifier(parser)
	}

	// Check if next token is an identifier and not a reserved keyword
	tok := parser.PeekToken()
	if word, ok := tok.Token.(tokenizer.TokenWord); ok {
		// Check if it's not a reserved keyword
		kw := token.Keyword(word.Word.Value)
		for _, reserved := range reservedKeywords {
			if kw == reserved {
				return nil, nil
			}
		}
		parser.AdvanceToken()
		return &ast.Ident{Value: word.Word.Value}, nil
	}

	return nil, nil
}

// parseLockTablesType parses the lock type for LOCK TABLES.
// READ [LOCAL] | [LOW_PRIORITY] WRITE
func parseLockTablesType(parser dialects.ParserAccessor) (statement.LockTableType, error) {
	if parser.ParseKeyword("READ") {
		if parser.ParseKeyword("LOCAL") {
			return statement.LockTableTypeReadLocal, nil
		}
		return statement.LockTableTypeRead, nil
	} else if parser.ParseKeyword("LOW_PRIORITY") {
		if parser.ParseKeyword("WRITE") {
			return statement.LockTableTypeWriteLowPriority, nil
		}
	} else if parser.ParseKeyword("WRITE") {
		return statement.LockTableTypeWrite, nil
	}

	return statement.LockTableTypeRead, nil
}

// parseUnlockTables parses the UNLOCK TABLES statement.
// https://dev.mysql.com/doc/refman/8.0/en/lock-tables.html
func (d *MySqlDialect) parseUnlockTables(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	return &statement.UnlockTablesStmt{}, true, nil
}

// ParseColumnOption returns false to fall back to default behavior.
func (d *MySqlDialect) ParseColumnOption(parser dialects.ParserAccessor) (dialects.ColumnOption, bool, error) {
	return nil, false, nil
}
