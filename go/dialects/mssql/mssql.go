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

package mssql

import (
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/token"
	"github.com/user/sqlparser/tokenizer"
)

// MsSqlDialect is a dialect for Microsoft SQL Server SQL implementation.
// See https://www.microsoft.com/en-us/sql-server/
type MsSqlDialect struct{}

// NewMsSqlDialect creates a new instance of MsSqlDialect.
func NewMsSqlDialect() *MsSqlDialect {
	return &MsSqlDialect{}
}

// Dialect returns the dialect identifier.
func (d *MsSqlDialect) Dialect() string {
	return "mssql"
}

// IsIdentifierStart returns true if the character is a valid start character
// for an unquoted identifier.
// See https://docs.microsoft.com/en-us/sql/relational-databases/databases/database-identifiers?view=sql-server-2017#rules-for-regular-identifiers
func (d *MsSqlDialect) IsIdentifierStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		ch == '_' || ch == '#' || ch == '@'
}

// IsIdentifierPart returns true if the character is a valid unquoted
// identifier character (not necessarily at the start).
func (d *MsSqlDialect) IsIdentifierPart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '@' || ch == '$' || ch == '#' || ch == '_'
}

// IsDelimitedIdentifierStart returns true if the character starts a quoted identifier.
func (d *MsSqlDialect) IsDelimitedIdentifierStart(ch rune) bool {
	return ch == '"' || ch == '['
}

// IsNestedDelimitedIdentifierStart returns true if the character starts a
// potential nested quoted identifier.
func (d *MsSqlDialect) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return false
}

// PeekNestedDelimitedIdentifierQuotes returns the outer quote style and
// optional inner quote style if applicable.
func (d *MsSqlDialect) PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*dialects.NestedIdentifierQuote, int) {
	return nil, pos
}

// IdentifierQuoteStyle returns the character used to quote identifiers.
// Returns '[' for MSSQL square bracket identifiers.
func (d *MsSqlDialect) IdentifierQuoteStyle(identifier string) *rune {
	quote := '['
	return &quote
}

// IsCustomOperatorPart returns true if the character is part of a custom operator.
func (d *MsSqlDialect) IsCustomOperatorPart(ch rune) bool {
	return false
}

// IsReservedForIdentifier returns true if the keyword is reserved.
func (d *MsSqlDialect) IsReservedForIdentifier(kw token.Keyword) bool {
	return token.IsReservedForIdentifier(kw)
}

// GetReservedKeywordsForSelectItemOperator returns reserved keywords for select items.
func (d *MsSqlDialect) GetReservedKeywordsForSelectItemOperator() []token.Keyword {
	return nil
}

// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
// See https://learn.microsoft.com/en-us/sql/relational-databases/security/authentication-access/server-level-roles
func (d *MsSqlDialect) GetReservedGranteesTypes() []dialects.GranteesType {
	return []dialects.GranteesType{dialects.GranteeTypeRole}
}

// reservedForSelectItemAliasMSSQL contains keywords that cannot be used as select item aliases.
var reservedForSelectItemAliasMSSQL = []token.Keyword{
	"IF", "ELSE", "DECLARE", "EXEC", "EXECUTE", "INSERT", "UPDATE",
	"DELETE", "DROP", "CREATE", "ALTER", "TRUNCATE", "PRINT", "WHILE",
	"RETURN", "THROW", "RAISERROR", "MERGE",
}

// reservedForTableAliasMSSQL contains keywords that cannot be used as table aliases in MSSQL.
var reservedForTableAliasMSSQL = []token.Keyword{
	"IF", "ELSE", "DECLARE", "EXEC", "EXECUTE", "INSERT", "UPDATE",
	"DELETE", "DROP", "CREATE", "ALTER", "TRUNCATE", "PRINT", "WHILE",
	"RETURN", "THROW", "RAISERROR", "MERGE",
}

// IsColumnAlias returns true if the keyword should be parsed as a column alias.
func (d *MsSqlDialect) IsColumnAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == "AS" || token.IsReservedForColumnAlias(kw)
}

// IsSelectItemAlias returns true if the specified keyword should be parsed as a select item alias
func (d *MsSqlDialect) IsSelectItemAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	for _, reserved := range reservedForSelectItemAliasMSSQL {
		if kw == reserved {
			return false
		}
	}
	return explicit || d.IsColumnAlias(kw, parser)
}

// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier
func (d *MsSqlDialect) IsTableFactor(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == "TABLE" || kw == "LATERAL"
}

// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias
func (d *MsSqlDialect) IsTableAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return !token.IsReservedForTableAlias(kw)
}

// IsTableFactorAlias returns true if the specified keyword should be parsed as a table factor alias
func (d *MsSqlDialect) IsTableFactorAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	for _, reserved := range reservedForTableAliasMSSQL {
		if kw == reserved {
			return false
		}
	}
	return explicit || d.IsTableAlias(kw, parser)
}

// IsIdentifierGeneratingFunctionName returns true if the ident is a function that generates identifiers.
func (d *MsSqlDialect) IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool {
	return false
}

// SupportsStringLiteralBackslashEscape returns true if the dialect supports
// escaping characters via '\' in string literals.
func (d *MsSqlDialect) SupportsStringLiteralBackslashEscape() bool {
	return false
}

// IgnoresWildcardEscapes returns true if the dialect strips the backslash
// when escaping LIKE wildcards (%, _).
func (d *MsSqlDialect) IgnoresWildcardEscapes() bool {
	return false
}

// SupportsUnicodeStringLiteral returns true if the dialect supports string
// literals with U& prefix for Unicode code points.
func (d *MsSqlDialect) SupportsUnicodeStringLiteral() bool {
	return false
}

// SupportsTripleQuotedString returns true if the dialect supports triple
// quoted string literals (e.g., """abc""").
func (d *MsSqlDialect) SupportsTripleQuotedString() bool {
	return false
}

// SupportsStringLiteralConcatenation returns true if the dialect supports
// concatenating string literals (e.g., 'Hello ' 'world').
func (d *MsSqlDialect) SupportsStringLiteralConcatenation() bool {
	return false
}

// SupportsStringLiteralConcatenationWithNewline returns true if the dialect
// supports concatenating string literals with a newline between them.
func (d *MsSqlDialect) SupportsStringLiteralConcatenationWithNewline() bool {
	return false
}

// SupportsQuoteDelimitedString returns true if the dialect supports
// quote-delimited string literals (e.g., Q'{...}') for Oracle-style strings.
func (d *MsSqlDialect) SupportsQuoteDelimitedString() bool {
	return false
}

// SupportsStringEscapeConstant returns true if the dialect supports the
// E'...' syntax for string literals with escape sequences (Postgres).
func (d *MsSqlDialect) SupportsStringEscapeConstant() bool {
	return false
}

// SupportsFilterDuringAggregation returns true if the dialect supports
// FILTER (WHERE expr) for aggregate queries.
func (d *MsSqlDialect) SupportsFilterDuringAggregation() bool {
	return false
}

// SupportsWithinAfterArrayAggregation returns true if the dialect supports
// ARRAY_AGG() [WITHIN GROUP (ORDER BY)] expressions.
func (d *MsSqlDialect) SupportsWithinAfterArrayAggregation() bool {
	return false
}

// SupportsWindowClauseNamedWindowReference returns true if the dialect
// supports referencing another named window within a window clause declaration.
func (d *MsSqlDialect) SupportsWindowClauseNamedWindowReference() bool {
	return false
}

// SupportsWindowFunctionNullTreatmentArg returns true if the dialect supports
// specifying null treatment as part of a window function's parameter list.
func (d *MsSqlDialect) SupportsWindowFunctionNullTreatmentArg() bool {
	return false
}

// SupportsMatchRecognize returns true if the dialect supports the MATCH_RECOGNIZE operation.
func (d *MsSqlDialect) SupportsMatchRecognize() bool {
	return false
}

// SupportsGroupByExpr returns true if the dialect supports GROUP BY expressions
// like GROUPING SETS, ROLLUP, or CUBE.
func (d *MsSqlDialect) SupportsGroupByExpr() bool {
	return false
}

// SupportsGroupByWithModifier returns true if the dialect supports GROUP BY
// modifiers prefixed by a WITH keyword (e.g., GROUP BY value WITH ROLLUP).
func (d *MsSqlDialect) SupportsGroupByWithModifier() bool {
	return false
}

// SupportsLeftAssociativeJoinsWithoutParens returns true if the dialect supports
// left-associative join parsing by default when parentheses are omitted in nested joins.
func (d *MsSqlDialect) SupportsLeftAssociativeJoinsWithoutParens() bool {
	return false
}

// SupportsOuterJoinOperator returns true if the dialect supports the (+)
// syntax for OUTER JOIN (Oracle-style).
func (d *MsSqlDialect) SupportsOuterJoinOperator() bool {
	return true
}

// SupportsCrossJoinConstraint returns true if the dialect supports a join
// specification on CROSS JOIN.
func (d *MsSqlDialect) SupportsCrossJoinConstraint() bool {
	return false
}

// SupportsConnectBy returns true if the dialect supports CONNECT BY for
// hierarchical queries (Oracle-style).
func (d *MsSqlDialect) SupportsConnectBy() bool {
	return true
}

// SupportsStartTransactionModifier returns true if the dialect supports
// BEGIN {DEFERRED | IMMEDIATE | EXCLUSIVE | TRY | CATCH} [TRANSACTION].
func (d *MsSqlDialect) SupportsStartTransactionModifier() bool {
	return true
}

// SupportsEndTransactionModifier returns true if the dialect supports
// END {TRY | CATCH} statements.
func (d *MsSqlDialect) SupportsEndTransactionModifier() bool {
	return true
}

// SupportsNamedFnArgsWithEqOperator returns true if the dialect supports
// named arguments of the form FUN(a = '1', b = '2').
func (d *MsSqlDialect) SupportsNamedFnArgsWithEqOperator() bool {
	return false
}

// SupportsNamedFnArgsWithColonOperator returns true if the dialect supports
// named arguments of the form FUN(a : '1', b : '2').
func (d *MsSqlDialect) SupportsNamedFnArgsWithColonOperator() bool {
	return true
}

// SupportsNamedFnArgsWithAssignmentOperator returns true if the dialect supports
// named arguments of the form FUN(a := '1', b := '2').
func (d *MsSqlDialect) SupportsNamedFnArgsWithAssignmentOperator() bool {
	return false
}

// SupportsNamedFnArgsWithRArrowOperator returns true if the dialect supports
// named arguments of the form FUN(a => '1', b => '2').
func (d *MsSqlDialect) SupportsNamedFnArgsWithRArrowOperator() bool {
	return false
}

// SupportsNamedFnArgsWithExprName returns true if the dialect supports
// argument name as arbitrary expression.
func (d *MsSqlDialect) SupportsNamedFnArgsWithExprName() bool {
	return true
}

// SupportsParenthesizedSetVariables returns true if the dialect supports
// multiple variable assignment using parentheses in a SET variable declaration.
func (d *MsSqlDialect) SupportsParenthesizedSetVariables() bool {
	return false
}

// SupportsCommaSeparatedSetAssignments returns true if the dialect supports
// multiple SET statements in a single statement.
func (d *MsSqlDialect) SupportsCommaSeparatedSetAssignments() bool {
	return false
}

// SupportsSetStmtWithoutOperator returns true if the dialect supports
// SET statements without an explicit assignment operator (e.g., SET SHOWPLAN_XML ON).
// See https://learn.microsoft.com/en-us/sql/t-sql/statements/set-statements-transact-sql
func (d *MsSqlDialect) SupportsSetStmtWithoutOperator() bool {
	return true
}

// SupportsSetNames returns true if the dialect supports SET NAMES <charset_name> [COLLATE <collation_name>].
func (d *MsSqlDialect) SupportsSetNames() bool {
	return false
}

// SupportsSelectWildcardExcept returns true if the dialect supports EXCEPT
// clause following a wildcard in a select list.
func (d *MsSqlDialect) SupportsSelectWildcardExcept() bool {
	return false
}

// SupportsSelectWildcardExclude returns true if the dialect supports EXCLUDE
// option following a wildcard in a projection section.
func (d *MsSqlDialect) SupportsSelectWildcardExclude() bool {
	return false
}

// SupportsSelectExclude returns true if the dialect supports EXCLUDE as the
// last item in the projection section, not necessarily after a wildcard.
func (d *MsSqlDialect) SupportsSelectExclude() bool {
	return false
}

// SupportsSelectWildcardReplace returns true if the dialect supports REPLACE
// option in SELECT * wildcard expressions.
func (d *MsSqlDialect) SupportsSelectWildcardReplace() bool {
	return false
}

// SupportsSelectWildcardIlike returns true if the dialect supports ILIKE
// option in SELECT * wildcard expressions.
func (d *MsSqlDialect) SupportsSelectWildcardIlike() bool {
	return false
}

// SupportsSelectWildcardRename returns true if the dialect supports RENAME
// option in SELECT * wildcard expressions.
func (d *MsSqlDialect) SupportsSelectWildcardRename() bool {
	return false
}

// SupportsSelectWildcardWithAlias returns true if the dialect supports aliasing
// a wildcard select item.
func (d *MsSqlDialect) SupportsSelectWildcardWithAlias() bool {
	return false
}

// SupportsSelectExprStar returns true if the dialect supports wildcard expansion
// on arbitrary expressions in projections.
func (d *MsSqlDialect) SupportsSelectExprStar() bool {
	return false
}

// SupportsFromFirstSelect returns true if the dialect supports "FROM-first" selects.
func (d *MsSqlDialect) SupportsFromFirstSelect() bool {
	return false
}

// SupportsEmptyProjections returns true if the dialect supports empty projections
// in SELECT statements.
func (d *MsSqlDialect) SupportsEmptyProjections() bool {
	return false
}

// SupportsSelectModifiers returns true if the dialect supports MySQL-specific
// SELECT modifiers like HIGH_PRIORITY, STRAIGHT_JOIN, SQL_SMALL_RESULT, etc.
func (d *MsSqlDialect) SupportsSelectModifiers() bool {
	return false
}

// SupportsPipeOperator returns true if the dialect supports the pipe operator (|>).
func (d *MsSqlDialect) SupportsPipeOperator() bool {
	return false
}

// SupportsTrailingCommas returns true if the dialect supports trailing commas.
func (d *MsSqlDialect) SupportsTrailingCommas() bool {
	return false
}

// SupportsProjectionTrailingCommas returns true if the dialect supports trailing
// commas in the projection list.
func (d *MsSqlDialect) SupportsProjectionTrailingCommas() bool {
	return false
}

// SupportsFromTrailingCommas returns true if the dialect supports trailing commas
// in the FROM clause.
func (d *MsSqlDialect) SupportsFromTrailingCommas() bool {
	return false
}

// SupportsColumnDefinitionTrailingCommas returns true if the dialect supports trailing
// commas in column definitions.
func (d *MsSqlDialect) SupportsColumnDefinitionTrailingCommas() bool {
	return false
}

// SupportsLimitComma returns true if the dialect supports parsing LIMIT 1, 2
// as LIMIT 2 OFFSET 1.
func (d *MsSqlDialect) SupportsLimitComma() bool {
	return false
}

// ConvertTypeBeforeValue returns true if the dialect has a CONVERT function
// which accepts a type first and an expression second.
// SQL Server has CONVERT(type, value) instead of CONVERT(value, type)
// See https://learn.microsoft.com/en-us/sql/t-sql/functions/cast-and-convert-transact-sql?view=sql-server-ver16
func (d *MsSqlDialect) ConvertTypeBeforeValue() bool {
	return true
}

// SupportsTryConvert returns true if the dialect supports the TRY_CONVERT function.
func (d *MsSqlDialect) SupportsTryConvert() bool {
	return true
}

// SupportsBinaryKwAsCast returns true if the dialect supports casting an expression
// to a binary type using the BINARY <expr> syntax.
func (d *MsSqlDialect) SupportsBinaryKwAsCast() bool {
	return false
}

// SupportsObjectNameDoubleDotNotation returns true if the dialect supports
// double dot notation for object names (e.g., db_name..table_name).
// See https://learn.microsoft.com/en-us/sql/t-sql/queries/from-transact-sql
func (d *MsSqlDialect) SupportsObjectNameDoubleDotNotation() bool {
	return true
}

// SupportsNumericPrefix returns true if the dialect supports identifiers
// starting with a numeric prefix.
func (d *MsSqlDialect) SupportsNumericPrefix() bool {
	return false
}

// SupportsNumericLiteralUnderscores returns true if the dialect supports
// numbers containing underscores (e.g., 10_000_000).
func (d *MsSqlDialect) SupportsNumericLiteralUnderscores() bool {
	return false
}

// SupportsInEmptyList returns true if the dialect supports (NOT) IN ()
// expressions with empty lists.
func (d *MsSqlDialect) SupportsInEmptyList() bool {
	return false
}

// SupportsDictionarySyntax returns true if the dialect supports defining
// structs or objects using syntax like {'x': 1, 'y': 2, 'z': 3}.
func (d *MsSqlDialect) SupportsDictionarySyntax() bool {
	return false
}

// SupportsMapLiteralSyntax returns true if the dialect supports defining
// objects using syntax like Map {1: 10, 2: 20}.
func (d *MsSqlDialect) SupportsMapLiteralSyntax() bool {
	return false
}

// SupportsStructLiteral returns true if the dialect supports STRUCT literal syntax.
func (d *MsSqlDialect) SupportsStructLiteral() bool {
	return false
}

// SupportsArrayLiteralSyntax returns true if the dialect supports array literal syntax.
func (d *MsSqlDialect) SupportsArrayLiteralSyntax() bool {
	return false
}

// SupportsLambdaFunctions returns true if the dialect supports lambda functions.
func (d *MsSqlDialect) SupportsLambdaFunctions() bool {
	return false
}

// SupportsCreateTableMultiSchemaInfoSources returns true if the dialect supports
// specifying multiple options in a CREATE TABLE statement.
func (d *MsSqlDialect) SupportsCreateTableMultiSchemaInfoSources() bool {
	return false
}

// SupportsCreateTableLikeParenthesized returns true if the dialect supports
// specifying which table to copy the schema from inside parenthesis.
func (d *MsSqlDialect) SupportsCreateTableLikeParenthesized() bool {
	return false
}

// SupportsCreateTableSelect returns true if the dialect supports CREATE TABLE SELECT.
func (d *MsSqlDialect) SupportsCreateTableSelect() bool {
	return false
}

// SupportsCreateViewCommentSyntax returns true if the dialect supports the COMMENT
// clause in CREATE VIEW statements using COMMENT = 'comment' syntax.
func (d *MsSqlDialect) SupportsCreateViewCommentSyntax() bool {
	return false
}

// SupportsArrayTypedefWithoutElementType returns true if the dialect supports
// ARRAY type without specifying an element type.
func (d *MsSqlDialect) SupportsArrayTypedefWithoutElementType() bool {
	return false
}

// SupportsArrayTypedefWithBrackets returns true if the dialect supports array
// type definition with brackets with optional size.
func (d *MsSqlDialect) SupportsArrayTypedefWithBrackets() bool {
	return false
}

// SupportsParensAroundTableFactor returns true if the dialect supports extra
// parentheses around lone table names or derived tables in the FROM clause.
func (d *MsSqlDialect) SupportsParensAroundTableFactor() bool {
	return false
}

// SupportsValuesAsTableFactor returns true if the dialect supports VALUES
// as a table factor without requiring parentheses.
func (d *MsSqlDialect) SupportsValuesAsTableFactor() bool {
	return false
}

// SupportsUnnestTableFactor returns false for MsSqlDialect.
func (d *MsSqlDialect) SupportsUnnestTableFactor() bool {
	return false
}

// SupportsSemanticViewTableFactor returns true if the dialect supports
// SEMANTIC_VIEW() table functions.
func (d *MsSqlDialect) SupportsSemanticViewTableFactor() bool {
	return false
}

// SupportsTableVersioning returns true if the dialect supports querying
// historical table data by specifying which version to query.
// See https://learn.microsoft.com/en-us/sql/relational-databases/tables/querying-data-in-a-system-versioned-temporal-table
func (d *MsSqlDialect) SupportsTableVersioning() bool {
	return true
}

// SupportsTableSampleBeforeAlias returns true if the dialect supports the
// TABLESAMPLE option before the table alias option.
func (d *MsSqlDialect) SupportsTableSampleBeforeAlias() bool {
	return false
}

// SupportsTableHints returns true if the dialect supports table hints in the FROM clause.
func (d *MsSqlDialect) SupportsTableHints() bool {
	return false
}

// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
func (d *MsSqlDialect) SupportsIndexHints() bool {
	return false
}

// SupportsAscDescInColumnDefinition returns true if the dialect supports
// ASC and DESC in column definitions.
func (d *MsSqlDialect) SupportsAscDescInColumnDefinition() bool {
	return false
}

// SupportsSpaceSeparatedColumnOptions returns true if the dialect supports
// space-separated column options in CREATE TABLE statements.
func (d *MsSqlDialect) SupportsSpaceSeparatedColumnOptions() bool {
	return false
}

// SupportsConstraintKeywordWithoutName returns true if the dialect supports
// CONSTRAINT keyword without a name in table constraint definitions.
func (d *MsSqlDialect) SupportsConstraintKeywordWithoutName() bool {
	return false
}

// SupportsKeyColumnOption returns true if the dialect supports the KEY keyword
// as part of column-level constraints.
func (d *MsSqlDialect) SupportsKeyColumnOption() bool {
	return false
}

// SupportsDataTypeSignedSuffix returns true if the dialect allows an optional
// SIGNED suffix after integer data types.
func (d *MsSqlDialect) SupportsDataTypeSignedSuffix() bool {
	return false
}

// SupportsNestedComments returns true if the dialect supports nested comments.
// See https://learn.microsoft.com/en-us/sql/t-sql/language-elements/slash-star-comment-transact-sql?view=sql-server-ver16
func (d *MsSqlDialect) SupportsNestedComments() bool {
	return true
}

// SupportsMultilineCommentHints returns true if the dialect supports optimizer
// hints in multiline comments (e.g., /*!50110 KEY_BLOCK_SIZE = 1024*/).
func (d *MsSqlDialect) SupportsMultilineCommentHints() bool {
	return false
}

// SupportsCommentOptimizerHint returns true if the dialect supports query
// optimizer hints in single and multi-line comments.
func (d *MsSqlDialect) SupportsCommentOptimizerHint() bool {
	return false
}

// SupportsCommentOn returns true if the dialect supports the COMMENT statement.
func (d *MsSqlDialect) SupportsCommentOn() bool {
	return false
}

// RequiresSingleLineCommentWhitespace returns true if the dialect requires
// a whitespace character after -- to start a single line comment.
func (d *MsSqlDialect) RequiresSingleLineCommentWhitespace() bool {
	return false
}

// SupportsExplainWithUtilityOptions returns true if the dialect supports
// EXPLAIN statements with utility options.
func (d *MsSqlDialect) SupportsExplainWithUtilityOptions() bool {
	return false
}

// SupportsExecuteImmediate returns true if the dialect supports EXECUTE IMMEDIATE statements.
func (d *MsSqlDialect) SupportsExecuteImmediate() bool {
	return false
}

// AllowExtractCustom returns true if the dialect allows the EXTRACT function
// to use words other than keywords.
func (d *MsSqlDialect) AllowExtractCustom() bool {
	return false
}

// AllowExtractSingleQuotes returns true if the dialect allows the EXTRACT
// function to use single quotes in the part being extracted.
func (d *MsSqlDialect) AllowExtractSingleQuotes() bool {
	return false
}

// SupportsExtractCommaSyntax returns true if the dialect supports EXTRACT
// function with a comma separator instead of FROM.
func (d *MsSqlDialect) SupportsExtractCommaSyntax() bool {
	return false
}

// SupportsSubqueryAsFunctionArg returns true if the dialect supports a subquery
// passed to a function as the only argument without enclosing parentheses.
func (d *MsSqlDialect) SupportsSubqueryAsFunctionArg() bool {
	return false
}

// SupportsDollarPlaceholder returns true if the dialect allows dollar placeholders (e.g., $var).
func (d *MsSqlDialect) SupportsDollarPlaceholder() bool {
	return false
}

// SupportsCreateIndexWithClause returns true if the dialect supports WITH clause
// in CREATE INDEX statement.
func (d *MsSqlDialect) SupportsCreateIndexWithClause() bool {
	return false
}

// RequireIntervalQualifier returns true if the dialect requires units
// (qualifiers) to be specified in INTERVAL expressions.
func (d *MsSqlDialect) RequireIntervalQualifier() bool {
	return true
}

// SupportsIntervalOptions returns true if the dialect supports INTERVAL data
// type with Postgres-style options.
func (d *MsSqlDialect) SupportsIntervalOptions() bool {
	return false
}

// SupportsFactorialOperator returns true if the dialect supports a!
// expressions for factorial.
func (d *MsSqlDialect) SupportsFactorialOperator() bool {
	return false
}

// SupportsBitwiseShiftOperators returns true if the dialect supports
// << and >> shift operators.
func (d *MsSqlDialect) SupportsBitwiseShiftOperators() bool {
	return false
}

// SupportsNotnullOperator returns true if the dialect supports the x NOTNULL
// operator expression.
func (d *MsSqlDialect) SupportsNotnullOperator() bool {
	return false
}

// SupportsBangNotOperator returns true if the dialect supports !a syntax
// for boolean NOT expressions.
func (d *MsSqlDialect) SupportsBangNotOperator() bool {
	return false
}

// SupportsDoubleAmpersandOperator returns true if the dialect considers
// the && operator as a boolean AND operator.
func (d *MsSqlDialect) SupportsDoubleAmpersandOperator() bool {
	return false
}

// SupportsMatchAgainst returns true if the dialect supports MATCH() AGAINST() syntax.
func (d *MsSqlDialect) SupportsMatchAgainst() bool {
	return false
}

// SupportsUserHostGrantee returns true if the dialect supports MySQL-style
// 'user'@'host' grantee syntax.
func (d *MsSqlDialect) SupportsUserHostGrantee() bool {
	return false
}

// SupportsListenNotify returns true if the dialect supports LISTEN, UNLISTEN,
// and NOTIFY statements.
func (d *MsSqlDialect) SupportsListenNotify() bool {
	return false
}

// SupportsLoadData returns true if the dialect supports LOAD DATA statement.
func (d *MsSqlDialect) SupportsLoadData() bool {
	return false
}

// SupportsLoadExtension returns true if the dialect supports LOAD extension statement.
func (d *MsSqlDialect) SupportsLoadExtension() bool {
	return false
}

// SupportsTopBeforeDistinct returns true if the dialect expects the TOP option
// before the ALL/DISTINCT options in a SELECT statement.
func (d *MsSqlDialect) SupportsTopBeforeDistinct() bool {
	return true
}

// SupportsBooleanLiterals returns true if the dialect supports boolean
// literals (true and false).
// In MSSQL, there is no boolean type, and `true` and `false` are valid column names
func (d *MsSqlDialect) SupportsBooleanLiterals() bool {
	return false
}

// SupportsShowLikeBeforeIn returns true if the dialect supports the LIKE
// option in a SHOW statement before the IN option.
func (d *MsSqlDialect) SupportsShowLikeBeforeIn() bool {
	return false
}

// SupportsPartiQL returns true if the dialect supports PartiQL for querying
// semi-structured data.
func (d *MsSqlDialect) SupportsPartiQL() bool {
	return false
}

// SupportsEqAliasAssignment returns true if the dialect supports treating
// the equals operator = within a SelectItem as an alias assignment operator.
func (d *MsSqlDialect) SupportsEqAliasAssignment() bool {
	return true
}

// SupportsInsertSet returns true if the dialect supports INSERT INTO ... SET syntax.
func (d *MsSqlDialect) SupportsInsertSet() bool {
	return false
}

// SupportsInsertTableFunction returns true if the dialect supports table
// function in insertion.
func (d *MsSqlDialect) SupportsInsertTableFunction() bool {
	return false
}

// SupportsInsertTableQuery returns true if the dialect supports table queries
// in insertion (e.g., SELECT INTO (<query>) ...).
func (d *MsSqlDialect) SupportsInsertTableQuery() bool {
	return false
}

// SupportsInsertFormat returns true if the dialect supports insert formats
// (e.g., INSERT INTO ... FORMAT <format>).
func (d *MsSqlDialect) SupportsInsertFormat() bool {
	return false
}

// SupportsInsertTableAlias returns true if the dialect supports INSERT INTO
// table [[AS] alias] syntax.
func (d *MsSqlDialect) SupportsInsertTableAlias() bool {
	return false
}

// SupportsAlterColumnTypeUsing returns true if the dialect supports the USING
// clause in an ALTER COLUMN statement.
func (d *MsSqlDialect) SupportsAlterColumnTypeUsing() bool {
	return false
}

// SupportsCommaSeparatedDropColumnList returns true if the dialect supports
// ALTER TABLE tbl DROP COLUMN c1, ..., cn.
func (d *MsSqlDialect) SupportsCommaSeparatedDropColumnList() bool {
	return false
}

// SupportsOrderByAll returns true if the dialect supports ORDER BY ALL.
func (d *MsSqlDialect) SupportsOrderByAll() bool {
	return false
}

// SupportsGeometricTypes returns true if the dialect supports geometric types
// (Postgres geometric operations).
func (d *MsSqlDialect) SupportsGeometricTypes() bool {
	return false
}

// DescribeRequiresTableKeyword returns true if the dialect requires the TABLE
// keyword after DESCRIBE.
func (d *MsSqlDialect) DescribeRequiresTableKeyword() bool {
	return false
}

// SupportsOptimizeTable returns true if the dialect supports OPTIMIZE TABLE.
func (d *MsSqlDialect) SupportsOptimizeTable() bool {
	return false
}

// SupportsPrewhere returns true if the dialect supports PREWHERE clause.
func (d *MsSqlDialect) SupportsPrewhere() bool {
	return false
}

// SupportsWithFill returns true if the dialect supports WITH FILL clause.
func (d *MsSqlDialect) SupportsWithFill() bool {
	return false
}

// SupportsLimitBy returns true if the dialect supports LIMIT BY clause.
func (d *MsSqlDialect) SupportsLimitBy() bool {
	return false
}

// SupportsInterpolate returns true if the dialect supports INTERPOLATE clause.
func (d *MsSqlDialect) SupportsInterpolate() bool {
	return false
}

// SupportsSettings returns true if the dialect supports SETTINGS clause.
func (d *MsSqlDialect) SupportsSettings() bool {
	return false
}

// SupportsSelectFormat returns true if the dialect supports FORMAT clause in SELECT.
func (d *MsSqlDialect) SupportsSelectFormat() bool {
	return false
}

// SupportsInstall returns true if the dialect supports INSTALL statement.
func (d *MsSqlDialect) SupportsInstall() bool {
	return false
}

// SupportsDetach returns true if the dialect supports DETACH statement.
func (d *MsSqlDialect) SupportsDetach() bool {
	return false
}

// SupportsCommaSeparatedTrim returns true if the dialect supports the two-argument
// comma-separated form of TRIM function: TRIM(expr, characters).
func (d *MsSqlDialect) SupportsCommaSeparatedTrim() bool {
	return false
}

// PrecValue returns the precedence value for a given Precedence level.
func (d *MsSqlDialect) PrecValue(prec dialects.Precedence) uint8 {
	return uint8(prec)
}

// PrecUnknown returns the precedence when precedence is otherwise unknown.
func (d *MsSqlDialect) PrecUnknown() uint8 {
	return 0
}

// GetNextPrecedence returns the dialect-specific precedence override for the next token.
// Returns precedence for colon operator used in JSON/variant access.
func (d *MsSqlDialect) GetNextPrecedence(parser dialects.ParserAccessor) (uint8, error) {
	tokenWithSpan := parser.PeekTokenRef()
	if _, ok := tokenWithSpan.Token.(tokenizer.TokenColon); ok {
		return uint8(dialects.PrecedenceColon), nil
	}
	return 0, nil
}

// GetNextPrecedenceDefault implements the default precedence logic.
func (d *MsSqlDialect) GetNextPrecedenceDefault(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// ParsePrefix is a dialect-specific prefix parser override.
func (d *MsSqlDialect) ParsePrefix(parser dialects.ParserAccessor) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseInfix is a dialect-specific infix parser override.
func (d *MsSqlDialect) ParseInfix(parser dialects.ParserAccessor, expr ast.Expr, precedence uint8) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseStatement is a dialect-specific statement parser override.
// Handles BEGIN blocks, IF statements, and CREATE TRIGGER.
func (d *MsSqlDialect) ParseStatement(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	// This is a simplified version - full implementation would need parser methods
	// that aren't available in this interface
	return nil, false, nil
}

// ParseColumnOption is a dialect-specific column option parser override.
func (d *MsSqlDialect) ParseColumnOption(parser dialects.ParserAccessor) (dialects.ColumnOption, bool, error) {
	return nil, false, nil
}
