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

package sqlite

import (
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/token"
)

// SQLiteDialect is a dialect for SQLite.
//
// This dialect allows columns in a CREATE TABLE statement with no
// type specified, as in `CREATE TABLE t1 (a)`. In the AST, these columns will
// have the data type Unspecified.
type SQLiteDialect struct{}

// NewSQLiteDialect creates a new instance of SQLiteDialect.
func NewSQLiteDialect() *SQLiteDialect {
	return &SQLiteDialect{}
}

// Dialect returns the dialect identifier.
func (d *SQLiteDialect) Dialect() string {
	return "sqlite"
}

// IsIdentifierStart returns true if the character is a valid start character
// for an unquoted identifier.
// See https://www.sqlite.org/draft/tokenreq.html
func (d *SQLiteDialect) IsIdentifierStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		ch == '_' ||
		(ch >= 0x007f && ch <= 0xffff)
}

// IsIdentifierPart returns true if the character is a valid unquoted
// identifier character.
func (d *SQLiteDialect) IsIdentifierPart(ch rune) bool {
	return d.IsIdentifierStart(ch) || (ch >= '0' && ch <= '9')
}

// IsDelimitedIdentifierStart returns true if the character starts a quoted identifier.
// See https://www.sqlite.org/lang_keywords.html
// parse `...`, [...] and "..." as identifier
func (d *SQLiteDialect) IsDelimitedIdentifierStart(ch rune) bool {
	return ch == '`' || ch == '"' || ch == '['
}

// IsNestedDelimitedIdentifierStart returns true if the character starts a
// potential nested quoted identifier.
func (d *SQLiteDialect) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return false
}

// PeekNestedDelimitedIdentifierQuotes returns the outer quote style and
// optional inner quote style if applicable.
func (d *SQLiteDialect) PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*dialects.NestedIdentifierQuote, int) {
	return nil, pos
}

// IdentifierQuoteStyle returns the character used to quote identifiers.
func (d *SQLiteDialect) IdentifierQuoteStyle(identifier string) *rune {
	quote := '`'
	return &quote
}

// IsCustomOperatorPart returns true if the character is part of a custom operator.
func (d *SQLiteDialect) IsCustomOperatorPart(ch rune) bool {
	return false
}

// IsReservedForIdentifier returns true if the keyword is reserved.
func (d *SQLiteDialect) IsReservedForIdentifier(kw token.Keyword) bool {
	return token.IsReservedForIdentifier(kw)
}

// GetReservedKeywordsForSelectItemOperator returns reserved keywords for select items.
func (d *SQLiteDialect) GetReservedKeywordsForSelectItemOperator() []token.Keyword {
	return nil
}

// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
func (d *SQLiteDialect) GetReservedGranteesTypes() []dialects.GranteesType {
	return nil
}

// IsColumnAlias returns true if the keyword should be parsed as a column alias.
func (d *SQLiteDialect) IsColumnAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == "AS" || token.IsReservedForColumnAlias(kw)
}

// IsSelectItemAlias returns true if the specified keyword should be parsed as a select item alias
func (d *SQLiteDialect) IsSelectItemAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsColumnAlias(kw, parser)
}

// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier
func (d *SQLiteDialect) IsTableFactor(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == "TABLE" || kw == "LATERAL" || kw == "PIVOT" || kw == "UNPIVOT"
}

// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias
func (d *SQLiteDialect) IsTableAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsTableFactor(kw, parser)
}

// IsTableFactorAlias returns true if the specified keyword should be parsed as a table factor alias
func (d *SQLiteDialect) IsTableFactorAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsTableAlias(kw, parser)
}

// IsIdentifierGeneratingFunctionName returns true if the ident is a function that generates identifiers.
func (d *SQLiteDialect) IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool {
	return false
}

// SupportsStringLiteralBackslashEscape returns false for SQLite.
// SQLite uses ” for escape, not backslash.
func (d *SQLiteDialect) SupportsStringLiteralBackslashEscape() bool {
	return false
}

// IgnoresWildcardEscapes returns false for SQLite.
func (d *SQLiteDialect) IgnoresWildcardEscapes() bool {
	return false
}

// SupportsUnicodeStringLiteral returns false for SQLite.
func (d *SQLiteDialect) SupportsUnicodeStringLiteral() bool {
	return false
}

// SupportsTripleQuotedString returns false for SQLite.
func (d *SQLiteDialect) SupportsTripleQuotedString() bool {
	return false
}

// SupportsStringLiteralConcatenation returns false for SQLite.
func (d *SQLiteDialect) SupportsStringLiteralConcatenation() bool {
	return false
}

// SupportsStringLiteralConcatenationWithNewline returns false for SQLite.
func (d *SQLiteDialect) SupportsStringLiteralConcatenationWithNewline() bool {
	return false
}

// SupportsQuoteDelimitedString returns false for SQLite.
func (d *SQLiteDialect) SupportsQuoteDelimitedString() bool {
	return false
}

// SupportsStringEscapeConstant returns false for SQLite.
func (d *SQLiteDialect) SupportsStringEscapeConstant() bool {
	return false
}

// SupportsFilterDuringAggregation returns true for SQLite.
func (d *SQLiteDialect) SupportsFilterDuringAggregation() bool {
	return true
}

// SupportsWithinAfterArrayAggregation returns false for SQLite.
func (d *SQLiteDialect) SupportsWithinAfterArrayAggregation() bool {
	return false
}

// SupportsWindowClauseNamedWindowReference returns true for SQLite.
func (d *SQLiteDialect) SupportsWindowClauseNamedWindowReference() bool {
	return true
}

// SupportsWindowFunctionNullTreatmentArg returns true for SQLite.
func (d *SQLiteDialect) SupportsWindowFunctionNullTreatmentArg() bool {
	return true
}

// SupportsMatchRecognize returns false for SQLite.
func (d *SQLiteDialect) SupportsMatchRecognize() bool {
	return false
}

// SupportsGroupByExpr returns true for SQLite.
func (d *SQLiteDialect) SupportsGroupByExpr() bool {
	return true
}

// SupportsGroupByWithModifier returns false for SQLite.
func (d *SQLiteDialect) SupportsGroupByWithModifier() bool {
	return false
}

// SupportsLeftAssociativeJoinsWithoutParens returns true for SQLite.
func (d *SQLiteDialect) SupportsLeftAssociativeJoinsWithoutParens() bool {
	return true
}

// SupportsOuterJoinOperator returns false for SQLite.
func (d *SQLiteDialect) SupportsOuterJoinOperator() bool {
	return false
}

// SupportsCrossJoinConstraint returns false for SQLite.
func (d *SQLiteDialect) SupportsCrossJoinConstraint() bool {
	return false
}

// SupportsConnectBy returns false for SQLite.
func (d *SQLiteDialect) SupportsConnectBy() bool {
	return false
}

// SupportsStartTransactionModifier returns true for SQLite.
// SQLite supports BEGIN {DEFERRED | IMMEDIATE | EXCLUSIVE} [TRANSACTION]
func (d *SQLiteDialect) SupportsStartTransactionModifier() bool {
	return true
}

// SupportsEndTransactionModifier returns false for SQLite.
func (d *SQLiteDialect) SupportsEndTransactionModifier() bool {
	return false
}

// SupportsNamedFnArgsWithEqOperator returns false for SQLite.
func (d *SQLiteDialect) SupportsNamedFnArgsWithEqOperator() bool {
	return false
}

// SupportsNamedFnArgsWithColonOperator returns false for SQLite.
func (d *SQLiteDialect) SupportsNamedFnArgsWithColonOperator() bool {
	return false
}

// SupportsNamedFnArgsWithAssignmentOperator returns false for SQLite.
func (d *SQLiteDialect) SupportsNamedFnArgsWithAssignmentOperator() bool {
	return false
}

// SupportsNamedFnArgsWithRArrowOperator returns false for SQLite.
func (d *SQLiteDialect) SupportsNamedFnArgsWithRArrowOperator() bool {
	return false
}

// SupportsNamedFnArgsWithExprName returns false for SQLite.
func (d *SQLiteDialect) SupportsNamedFnArgsWithExprName() bool {
	return false
}

// SupportsParenthesizedSetVariables returns false for SQLite.
func (d *SQLiteDialect) SupportsParenthesizedSetVariables() bool {
	return false
}

// SupportsCommaSeparatedSetAssignments returns false for SQLite.
func (d *SQLiteDialect) SupportsCommaSeparatedSetAssignments() bool {
	return false
}

// SupportsSetStmtWithoutOperator returns false for SQLite.
func (d *SQLiteDialect) SupportsSetStmtWithoutOperator() bool {
	return false
}

// SupportsSetNames returns false for SQLite.
func (d *SQLiteDialect) SupportsSetNames() bool {
	return false
}

// SupportsSelectWildcardExcept returns false for SQLite.
func (d *SQLiteDialect) SupportsSelectWildcardExcept() bool {
	return false
}

// SupportsSelectWildcardExclude returns false for SQLite.
func (d *SQLiteDialect) SupportsSelectWildcardExclude() bool {
	return false
}

// SupportsSelectExclude returns false for SQLite.
func (d *SQLiteDialect) SupportsSelectExclude() bool {
	return false
}

// SupportsSelectWildcardReplace returns false for SQLite.
func (d *SQLiteDialect) SupportsSelectWildcardReplace() bool {
	return false
}

// SupportsSelectWildcardIlike returns false for SQLite.
func (d *SQLiteDialect) SupportsSelectWildcardIlike() bool {
	return false
}

// SupportsSelectWildcardRename returns false for SQLite.
func (d *SQLiteDialect) SupportsSelectWildcardRename() bool {
	return false
}

// SupportsSelectWildcardWithAlias returns false for SQLite.
func (d *SQLiteDialect) SupportsSelectWildcardWithAlias() bool {
	return false
}

// SupportsSelectExprStar returns false for SQLite.
func (d *SQLiteDialect) SupportsSelectExprStar() bool {
	return false
}

// SupportsFromFirstSelect returns false for SQLite.
func (d *SQLiteDialect) SupportsFromFirstSelect() bool {
	return false
}

// SupportsEmptyProjections returns false for SQLite.
func (d *SQLiteDialect) SupportsEmptyProjections() bool {
	return false
}

// SupportsSelectModifiers returns false for SQLite.
func (d *SQLiteDialect) SupportsSelectModifiers() bool {
	return false
}

// SupportsPipeOperator returns false for SQLite.
func (d *SQLiteDialect) SupportsPipeOperator() bool {
	return false
}

// SupportsTrailingCommas returns false for SQLite.
func (d *SQLiteDialect) SupportsTrailingCommas() bool {
	return false
}

// SupportsProjectionTrailingCommas returns false for SQLite.
func (d *SQLiteDialect) SupportsProjectionTrailingCommas() bool {
	return false
}

// SupportsFromTrailingCommas returns false for SQLite.
func (d *SQLiteDialect) SupportsFromTrailingCommas() bool {
	return false
}

// SupportsColumnDefinitionTrailingCommas returns false for SQLite.
func (d *SQLiteDialect) SupportsColumnDefinitionTrailingCommas() bool {
	return false
}

// SupportsLimitComma returns true for SQLite.
// SQLite supports LIMIT 1, 2 as LIMIT 2 OFFSET 1
func (d *SQLiteDialect) SupportsLimitComma() bool {
	return true
}

// ConvertTypeBeforeValue returns false for SQLite.
func (d *SQLiteDialect) ConvertTypeBeforeValue() bool {
	return false
}

// SupportsTryConvert returns false for SQLite.
func (d *SQLiteDialect) SupportsTryConvert() bool {
	return false
}

// SupportsBinaryKwAsCast returns false for SQLite.
func (d *SQLiteDialect) SupportsBinaryKwAsCast() bool {
	return false
}

// SupportsObjectNameDoubleDotNotation returns false for SQLite.
func (d *SQLiteDialect) SupportsObjectNameDoubleDotNotation() bool {
	return false
}

// SupportsNumericPrefix returns false for SQLite.
func (d *SQLiteDialect) SupportsNumericPrefix() bool {
	return false
}

// SupportsNumericLiteralUnderscores returns false for SQLite.
func (d *SQLiteDialect) SupportsNumericLiteralUnderscores() bool {
	return false
}

// SupportsInEmptyList returns true for SQLite.
func (d *SQLiteDialect) SupportsInEmptyList() bool {
	return true
}

// SupportsDictionarySyntax returns false for SQLite.
func (d *SQLiteDialect) SupportsDictionarySyntax() bool {
	return false
}

// SupportsMapLiteralSyntax returns false for SQLite.
func (d *SQLiteDialect) SupportsMapLiteralSyntax() bool {
	return false
}

// SupportsStructLiteral returns false for SQLite.
func (d *SQLiteDialect) SupportsStructLiteral() bool {
	return false
}

// SupportsArrayLiteralSyntax returns false for SQLite.
func (d *SQLiteDialect) SupportsArrayLiteralSyntax() bool {
	return false
}

// SupportsLambdaFunctions returns false for SQLite.
func (d *SQLiteDialect) SupportsLambdaFunctions() bool {
	return false
}

// SupportsCreateTableMultiSchemaInfoSources returns false for SQLite.
func (d *SQLiteDialect) SupportsCreateTableMultiSchemaInfoSources() bool {
	return false
}

// SupportsCreateTableLikeParenthesized returns false for SQLite.
func (d *SQLiteDialect) SupportsCreateTableLikeParenthesized() bool {
	return false
}

// SupportsCreateTableSelect returns true for SQLite.
func (d *SQLiteDialect) SupportsCreateTableSelect() bool {
	return true
}

// SupportsCreateViewCommentSyntax returns false for SQLite.
func (d *SQLiteDialect) SupportsCreateViewCommentSyntax() bool {
	return false
}

// SupportsArrayTypedefWithoutElementType returns false for SQLite.
func (d *SQLiteDialect) SupportsArrayTypedefWithoutElementType() bool {
	return false
}

// SupportsArrayTypedefWithBrackets returns false for SQLite.
func (d *SQLiteDialect) SupportsArrayTypedefWithBrackets() bool {
	return false
}

// SupportsParensAroundTableFactor returns true for SQLite.
func (d *SQLiteDialect) SupportsParensAroundTableFactor() bool {
	return true
}

// SupportsValuesAsTableFactor returns true for SQLite.
func (d *SQLiteDialect) SupportsValuesAsTableFactor() bool {
	return true
}

// SupportsSemanticViewTableFactor returns false for SQLite.
func (d *SQLiteDialect) SupportsSemanticViewTableFactor() bool {
	return false
}

// SupportsTableVersioning returns false for SQLite.
func (d *SQLiteDialect) SupportsTableVersioning() bool {
	return false
}

// SupportsTableSampleBeforeAlias returns false for SQLite.
func (d *SQLiteDialect) SupportsTableSampleBeforeAlias() bool {
	return false
}

// SupportsTableHints returns false for SQLite.
func (d *SQLiteDialect) SupportsTableHints() bool {
	return false
}

// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
func (d *SQLiteDialect) SupportsIndexHints() bool {
	return false
}

// SupportsAscDescInColumnDefinition returns true for SQLite.
// SQLite supports ASC and DESC in column definitions for PRIMARY KEY
func (d *SQLiteDialect) SupportsAscDescInColumnDefinition() bool {
	return true
}

// SupportsSpaceSeparatedColumnOptions returns false for SQLite.
func (d *SQLiteDialect) SupportsSpaceSeparatedColumnOptions() bool {
	return false
}

// SupportsConstraintKeywordWithoutName returns true for SQLite.
func (d *SQLiteDialect) SupportsConstraintKeywordWithoutName() bool {
	return true
}

// SupportsKeyColumnOption returns false for SQLite.
func (d *SQLiteDialect) SupportsKeyColumnOption() bool {
	return false
}

// SupportsDataTypeSignedSuffix returns false for SQLite.
func (d *SQLiteDialect) SupportsDataTypeSignedSuffix() bool {
	return false
}

// SupportsNestedComments returns false for SQLite.
func (d *SQLiteDialect) SupportsNestedComments() bool {
	return false
}

// SupportsMultilineCommentHints returns false for SQLite.
func (d *SQLiteDialect) SupportsMultilineCommentHints() bool {
	return false
}

// SupportsCommentOptimizerHint returns false for SQLite.
func (d *SQLiteDialect) SupportsCommentOptimizerHint() bool {
	return false
}

// SupportsCommentOn returns false for SQLite.
func (d *SQLiteDialect) SupportsCommentOn() bool {
	return false
}

// RequiresSingleLineCommentWhitespace returns false for SQLite.
func (d *SQLiteDialect) RequiresSingleLineCommentWhitespace() bool {
	return false
}

// SupportsExplainWithUtilityOptions returns true for SQLite.
// SQLite supports EXPLAIN QUERY PLAN
func (d *SQLiteDialect) SupportsExplainWithUtilityOptions() bool {
	return true
}

// SupportsExecuteImmediate returns false for SQLite.
func (d *SQLiteDialect) SupportsExecuteImmediate() bool {
	return false
}

// AllowExtractCustom returns false for SQLite.
func (d *SQLiteDialect) AllowExtractCustom() bool {
	return false
}

// AllowExtractSingleQuotes returns false for SQLite.
func (d *SQLiteDialect) AllowExtractSingleQuotes() bool {
	return false
}

// SupportsExtractCommaSyntax returns false for SQLite.
func (d *SQLiteDialect) SupportsExtractCommaSyntax() bool {
	return false
}

// SupportsSubqueryAsFunctionArg returns false for SQLite.
func (d *SQLiteDialect) SupportsSubqueryAsFunctionArg() bool {
	return false
}

// SupportsDollarPlaceholder returns true for SQLite.
// SQLite allows dollar placeholders (e.g., $var)
func (d *SQLiteDialect) SupportsDollarPlaceholder() bool {
	return true
}

// SupportsCreateIndexWithClause returns false for SQLite.
func (d *SQLiteDialect) SupportsCreateIndexWithClause() bool {
	return false
}

// RequireIntervalQualifier returns false for SQLite.
func (d *SQLiteDialect) RequireIntervalQualifier() bool {
	return false
}

// SupportsIntervalOptions returns false for SQLite.
func (d *SQLiteDialect) SupportsIntervalOptions() bool {
	return false
}

// SupportsFactorialOperator returns false for SQLite.
func (d *SQLiteDialect) SupportsFactorialOperator() bool {
	return false
}

// SupportsBitwiseShiftOperators returns true for SQLite.
func (d *SQLiteDialect) SupportsBitwiseShiftOperators() bool {
	return true
}

// SupportsNotnullOperator returns true for SQLite.
// SQLite supports `NOTNULL` as aliases for `IS NOT NULL`
// See: https://sqlite.org/syntax/expr.html
func (d *SQLiteDialect) SupportsNotnullOperator() bool {
	return true
}

// SupportsBangNotOperator returns false for SQLite.
func (d *SQLiteDialect) SupportsBangNotOperator() bool {
	return false
}

// SupportsDoubleAmpersandOperator returns false for SQLite.
func (d *SQLiteDialect) SupportsDoubleAmpersandOperator() bool {
	return false
}

// SupportsMatchAgainst returns false for SQLite.
// Note: SQLite has MATCH operator but not MATCH() AGAINST() syntax
func (d *SQLiteDialect) SupportsMatchAgainst() bool {
	return false
}

// SupportsUserHostGrantee returns false for SQLite.
func (d *SQLiteDialect) SupportsUserHostGrantee() bool {
	return false
}

// SupportsListenNotify returns false for SQLite.
func (d *SQLiteDialect) SupportsListenNotify() bool {
	return false
}

// SupportsLoadData returns false for SQLite.
func (d *SQLiteDialect) SupportsLoadData() bool {
	return false
}

// SupportsLoadExtension returns true for SQLite.
// SQLite supports LOAD EXTENSION
func (d *SQLiteDialect) SupportsLoadExtension() bool {
	return true
}

// SupportsTopBeforeDistinct returns false for SQLite.
func (d *SQLiteDialect) SupportsTopBeforeDistinct() bool {
	return false
}

// SupportsBooleanLiterals returns true for SQLite.
func (d *SQLiteDialect) SupportsBooleanLiterals() bool {
	return true
}

// SupportsShowLikeBeforeIn returns false for SQLite.
func (d *SQLiteDialect) SupportsShowLikeBeforeIn() bool {
	return false
}

// SupportsPartiQL returns false for SQLite.
func (d *SQLiteDialect) SupportsPartiQL() bool {
	return false
}

// SupportsEqAliasAssignment returns false for SQLite.
func (d *SQLiteDialect) SupportsEqAliasAssignment() bool {
	return false
}

// SupportsInsertSet returns false for SQLite.
func (d *SQLiteDialect) SupportsInsertSet() bool {
	return false
}

// SupportsInsertTableFunction returns false for SQLite.
func (d *SQLiteDialect) SupportsInsertTableFunction() bool {
	return false
}

// SupportsInsertTableQuery returns false for SQLite.
func (d *SQLiteDialect) SupportsInsertTableQuery() bool {
	return false
}

// SupportsInsertFormat returns false for SQLite.
func (d *SQLiteDialect) SupportsInsertFormat() bool {
	return false
}

// SupportsInsertTableAlias returns false for SQLite.
func (d *SQLiteDialect) SupportsInsertTableAlias() bool {
	return false
}

// SupportsAlterColumnTypeUsing returns false for SQLite.
func (d *SQLiteDialect) SupportsAlterColumnTypeUsing() bool {
	return false
}

// SupportsCommaSeparatedDropColumnList returns true for SQLite.
// SQLite 3.35.0+ supports ALTER TABLE tbl DROP COLUMN c1, ..., cn
func (d *SQLiteDialect) SupportsCommaSeparatedDropColumnList() bool {
	return true
}

// SupportsOrderByAll returns false for SQLite.
func (d *SQLiteDialect) SupportsOrderByAll() bool {
	return false
}

// SupportsGeometricTypes returns false for SQLite.
func (d *SQLiteDialect) SupportsGeometricTypes() bool {
	return false
}

// DescribeRequiresTableKeyword returns false for SQLite.
func (d *SQLiteDialect) DescribeRequiresTableKeyword() bool {
	return false
}

// SupportsOptimizeTable returns false for SQLite.
func (d *SQLiteDialect) SupportsOptimizeTable() bool {
	return false
}

// SupportsPrewhere returns false for SQLite.
func (d *SQLiteDialect) SupportsPrewhere() bool {
	return false
}

// SupportsWithFill returns false for SQLite.
func (d *SQLiteDialect) SupportsWithFill() bool {
	return false
}

// SupportsLimitBy returns false for SQLite.
func (d *SQLiteDialect) SupportsLimitBy() bool {
	return false
}

// SupportsInterpolate returns false for SQLite.
func (d *SQLiteDialect) SupportsInterpolate() bool {
	return false
}

// SupportsSettings returns false for SQLite.
func (d *SQLiteDialect) SupportsSettings() bool {
	return false
}

// SupportsSelectFormat returns false for SQLite.
func (d *SQLiteDialect) SupportsSelectFormat() bool {
	return false
}

// SupportsInstall returns false for SQLite.
func (d *SQLiteDialect) SupportsInstall() bool {
	return false
}

// SupportsDetach returns true for SQLite.
// SQLite supports DETACH DATABASE
func (d *SQLiteDialect) SupportsDetach() bool {
	return true
}

// SupportsCommaSeparatedTrim returns true for SQLite.
// SQLite supports TRIM(expr, characters)
func (d *SQLiteDialect) SupportsCommaSeparatedTrim() bool {
	return true
}

// PrecValue returns the precedence value for a given Precedence level.
func (d *SQLiteDialect) PrecValue(prec dialects.Precedence) uint8 {
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
func (d *SQLiteDialect) PrecUnknown() uint8 {
	return 0
}

// GetNextPrecedence returns 0 to fall back to default behavior.
func (d *SQLiteDialect) GetNextPrecedence(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// GetNextPrecedenceDefault implements the default precedence logic.
func (d *SQLiteDialect) GetNextPrecedenceDefault(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// ParsePrefix returns false to fall back to default behavior.
func (d *SQLiteDialect) ParsePrefix(parser dialects.ParserAccessor) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseInfix parses MATCH and REGEXP as binary operators.
// See https://www.sqlite.org/lang_expr.html#the_like_glob_regexp_match_and_extract_operators
// TODO: Fix interface compatibility between ast.Expr and expr.Expr
func (d *SQLiteDialect) ParseInfix(parser dialects.ParserAccessor, leftExpr ast.Expr, precedence uint8) (ast.Expr, bool, error) {
	// TODO: Implement proper binary operation parsing
	// Currently disabled due to interface compatibility issues
	// between ast.Expr and expr.Expr
	return nil, false, nil
}

// ParseStatement handles REPLACE as an INSERT statement.
func (d *SQLiteDialect) ParseStatement(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	if parser.ParseKeyword("REPLACE") {
		parser.PrevToken()
		stmt, err := parser.ParseInsert()
		if err != nil {
			return nil, true, err
		}
		return stmt, true, nil
	}
	return nil, false, nil
}

// ParseColumnOption returns false to fall back to default behavior.
func (d *SQLiteDialect) ParseColumnOption(parser dialects.ParserAccessor) (dialects.ColumnOption, bool, error) {
	return nil, false, nil
}
