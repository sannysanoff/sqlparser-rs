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

package bigquery

import (
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/token"
	"github.com/user/sqlparser/tokenizer"
)

// BigQueryDialect is a dialect for Google BigQuery SQL.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/
type BigQueryDialect struct{}

// NewBigQueryDialect creates a new instance of BigQueryDialect.
func NewBigQueryDialect() *BigQueryDialect {
	return &BigQueryDialect{}
}

// Dialect returns the dialect identifier.
func (d *BigQueryDialect) Dialect() string {
	return "bigquery"
}

// IsIdentifierStart returns true if the character is a valid start character
// for an unquoted identifier.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/lexical#identifiers
func (d *BigQueryDialect) IsIdentifierStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' ||
		// BigQuery supports `@@foo.bar` variable syntax in its procedural language.
		// https://cloud.google.com/bigquery/docs/reference/standard-sql/procedural-language#beginexceptionend
		ch == '@'
}

// IsIdentifierPart returns true if the character is a valid unquoted
// identifier character.
func (d *BigQueryDialect) IsIdentifierPart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_'
}

// IsDelimitedIdentifierStart returns true if the character starts a quoted identifier.
// BigQuery uses backticks for delimited identifiers.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/lexical#identifiers
func (d *BigQueryDialect) IsDelimitedIdentifierStart(ch rune) bool {
	return ch == '`'
}

// IsNestedDelimitedIdentifierStart returns true if the character starts a
// potential nested quoted identifier.
func (d *BigQueryDialect) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return false
}

// PeekNestedDelimitedIdentifierQuotes returns the outer quote style and
// optional inner quote style if applicable.
func (d *BigQueryDialect) PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*dialects.NestedIdentifierQuote, int) {
	return nil, pos
}

// IdentifierQuoteStyle returns the character used to quote identifiers.
func (d *BigQueryDialect) IdentifierQuoteStyle(identifier string) *rune {
	return nil
}

// IsCustomOperatorPart returns true if the character is part of a custom operator.
func (d *BigQueryDialect) IsCustomOperatorPart(ch rune) bool {
	return false
}

// IsReservedForIdentifier returns true if the keyword is reserved.
func (d *BigQueryDialect) IsReservedForIdentifier(kw token.Keyword) bool {
	return token.IsReservedForIdentifier(kw)
}

// GetReservedKeywordsForSelectItemOperator returns reserved keywords for select items.
func (d *BigQueryDialect) GetReservedKeywordsForSelectItemOperator() []token.Keyword {
	return nil
}

// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
func (d *BigQueryDialect) GetReservedGranteesTypes() []dialects.GranteesType {
	return nil
}

// reservedForColumnAlias contains keywords that are disallowed as column identifiers.
// Such that `SELECT 5 AS <col> FROM T` is rejected by BigQuery.
var reservedForColumnAlias = []token.Keyword{
	"WITH",
	"SELECT",
	"WHERE",
	"GROUP",
	"HAVING",
	"ORDER",
	"LATERAL",
	"LIMIT",
	"FETCH",
	"UNION",
	"EXCEPT",
	"INTERSECT",
	"FROM",
	"INTO",
	"END",
}

// IsColumnAlias returns true if the keyword should be parsed as a column alias.
func (d *BigQueryDialect) IsColumnAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	for _, reserved := range reservedForColumnAlias {
		if kw == reserved {
			return false
		}
	}
	return true
}

// IsSelectItemAlias returns true if the specified keyword should be parsed as a select item alias
func (d *BigQueryDialect) IsSelectItemAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsColumnAlias(kw, parser)
}

// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier
func (d *BigQueryDialect) IsTableFactor(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == "TABLE" || kw == "LATERAL" || kw == "PIVOT" || kw == "UNPIVOT"
}

// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias
func (d *BigQueryDialect) IsTableAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsTableFactor(kw, parser)
}

// IsTableFactorAlias returns true if the specified keyword should be parsed as a table factor alias
func (d *BigQueryDialect) IsTableFactorAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsTableAlias(kw, parser)
}

// IsIdentifierGeneratingFunctionName returns true if the ident is a function that generates identifiers.
func (d *BigQueryDialect) IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool {
	return false
}

// SupportsStringLiteralBackslashEscape returns true if the dialect supports
// escaping characters via '\' in string literals.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/lexical#escape_sequences
func (d *BigQueryDialect) SupportsStringLiteralBackslashEscape() bool {
	return true
}

// IgnoresWildcardEscapes returns false for BigQueryDialect.
func (d *BigQueryDialect) IgnoresWildcardEscapes() bool {
	return false
}

// SupportsUnicodeStringLiteral returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsUnicodeStringLiteral() bool {
	return false
}

// SupportsTripleQuotedString returns true for BigQueryDialect.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/lexical#quoted_literals
func (d *BigQueryDialect) SupportsTripleQuotedString() bool {
	return true
}

// SupportsStringLiteralConcatenation returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsStringLiteralConcatenation() bool {
	return false
}

// SupportsStringLiteralConcatenationWithNewline returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsStringLiteralConcatenationWithNewline() bool {
	return false
}

// SupportsQuoteDelimitedString returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsQuoteDelimitedString() bool {
	return false
}

// SupportsStringEscapeConstant returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsStringEscapeConstant() bool {
	return false
}

// SupportsFilterDuringAggregation returns true for BigQueryDialect.
func (d *BigQueryDialect) SupportsFilterDuringAggregation() bool {
	return true
}

// SupportsWithinAfterArrayAggregation returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsWithinAfterArrayAggregation() bool {
	return false
}

// SupportsWindowClauseNamedWindowReference returns true for BigQueryDialect.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/window-function-calls#ref_named_window
func (d *BigQueryDialect) SupportsWindowClauseNamedWindowReference() bool {
	return true
}

// SupportsWindowFunctionNullTreatmentArg returns true for BigQueryDialect.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/navigation_functions#first_value
func (d *BigQueryDialect) SupportsWindowFunctionNullTreatmentArg() bool {
	return true
}

// SupportsMatchRecognize returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsMatchRecognize() bool {
	return false
}

// SupportsGroupByExpr returns true for BigQueryDialect.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/query-syntax#group_by_clause
func (d *BigQueryDialect) SupportsGroupByExpr() bool {
	return true
}

// SupportsGroupByWithModifier returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsGroupByWithModifier() bool {
	return false
}

// SupportsLeftAssociativeJoinsWithoutParens returns true for BigQueryDialect.
func (d *BigQueryDialect) SupportsLeftAssociativeJoinsWithoutParens() bool {
	return true
}

// SupportsOuterJoinOperator returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsOuterJoinOperator() bool {
	return false
}

// SupportsCrossJoinConstraint returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsCrossJoinConstraint() bool {
	return false
}

// SupportsConnectBy returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsConnectBy() bool {
	return false
}

// SupportsStartTransactionModifier returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsStartTransactionModifier() bool {
	return false
}

// SupportsEndTransactionModifier returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsEndTransactionModifier() bool {
	return false
}

// SupportsNamedFnArgsWithEqOperator returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsNamedFnArgsWithEqOperator() bool {
	return false
}

// SupportsNamedFnArgsWithColonOperator returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsNamedFnArgsWithColonOperator() bool {
	return false
}

// SupportsNamedFnArgsWithAssignmentOperator returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsNamedFnArgsWithAssignmentOperator() bool {
	return false
}

// SupportsNamedFnArgsWithRArrowOperator returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsNamedFnArgsWithRArrowOperator() bool {
	return false
}

// SupportsNamedFnArgsWithExprName returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsNamedFnArgsWithExprName() bool {
	return false
}

// SupportsParenthesizedSetVariables returns true for BigQueryDialect.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/procedural-language#set
func (d *BigQueryDialect) SupportsParenthesizedSetVariables() bool {
	return true
}

// SupportsCommaSeparatedSetAssignments returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsCommaSeparatedSetAssignments() bool {
	return false
}

// SupportsSetStmtWithoutOperator returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsSetStmtWithoutOperator() bool {
	return false
}

// SupportsSetNames returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsSetNames() bool {
	return false
}

// SupportsSelectWildcardExcept returns true for BigQueryDialect.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/query-syntax#select_except
func (d *BigQueryDialect) SupportsSelectWildcardExcept() bool {
	return true
}

// SupportsSelectWildcardExclude returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsSelectWildcardExclude() bool {
	return false
}

// SupportsSelectExclude returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsSelectExclude() bool {
	return false
}

// SupportsSelectWildcardReplace returns true for BigQueryDialect.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/query-syntax#select_replace
func (d *BigQueryDialect) SupportsSelectWildcardReplace() bool {
	return true
}

// SupportsSelectWildcardIlike returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsSelectWildcardIlike() bool {
	return false
}

// SupportsSelectWildcardRename returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsSelectWildcardRename() bool {
	return false
}

// SupportsSelectWildcardWithAlias returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsSelectWildcardWithAlias() bool {
	return false
}

// SupportsSelectExprStar returns true for BigQueryDialect.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/query-syntax#select_expression_star
func (d *BigQueryDialect) SupportsSelectExprStar() bool {
	return true
}

// SupportsFromFirstSelect returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsFromFirstSelect() bool {
	return false
}

// SupportsEmptyProjections returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsEmptyProjections() bool {
	return false
}

// SupportsSelectModifiers returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsSelectModifiers() bool {
	return false
}

// SupportsPipeOperator returns true for BigQueryDialect.
func (d *BigQueryDialect) SupportsPipeOperator() bool {
	return true
}

// SupportsTrailingCommas returns true for BigQueryDialect.
func (d *BigQueryDialect) SupportsTrailingCommas() bool {
	return true
}

// SupportsProjectionTrailingCommas returns true for BigQueryDialect.
func (d *BigQueryDialect) SupportsProjectionTrailingCommas() bool {
	return true
}

// SupportsFromTrailingCommas returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsFromTrailingCommas() bool {
	return false
}

// SupportsColumnDefinitionTrailingCommas returns true for BigQueryDialect.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/data-definition-language#create_table_statement
func (d *BigQueryDialect) SupportsColumnDefinitionTrailingCommas() bool {
	return true
}

// SupportsLimitComma returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsLimitComma() bool {
	return false
}

// ConvertTypeBeforeValue returns false for BigQueryDialect.
func (d *BigQueryDialect) ConvertTypeBeforeValue() bool {
	return false
}

// SupportsTryConvert returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsTryConvert() bool {
	return false
}

// SupportsBinaryKwAsCast returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsBinaryKwAsCast() bool {
	return false
}

// SupportsObjectNameDoubleDotNotation returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsObjectNameDoubleDotNotation() bool {
	return false
}

// SupportsNumericPrefix returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsNumericPrefix() bool {
	return false
}

// SupportsNumericLiteralUnderscores returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsNumericLiteralUnderscores() bool {
	return false
}

// SupportsInEmptyList returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsInEmptyList() bool {
	return false
}

// SupportsDictionarySyntax returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsDictionarySyntax() bool {
	return false
}

// SupportsMapLiteralSyntax returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsMapLiteralSyntax() bool {
	return false
}

// SupportsStructLiteral returns true for BigQueryDialect.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/data-types#constructing_a_struct
func (d *BigQueryDialect) SupportsStructLiteral() bool {
	return true
}

// SupportsArrayLiteralSyntax returns true for BigQueryDialect.
func (d *BigQueryDialect) SupportsArrayLiteralSyntax() bool {
	return true
}

// SupportsLambdaFunctions returns true for BigQueryDialect.
func (d *BigQueryDialect) SupportsLambdaFunctions() bool {
	return true
}

// SupportsCreateTableMultiSchemaInfoSources returns true for BigQueryDialect.
func (d *BigQueryDialect) SupportsCreateTableMultiSchemaInfoSources() bool {
	return true
}

// SupportsCreateTableLikeParenthesized returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsCreateTableLikeParenthesized() bool {
	return false
}

// SupportsCreateTableSelect returns true for BigQueryDialect.
func (d *BigQueryDialect) SupportsCreateTableSelect() bool {
	return true
}

// SupportsCreateViewCommentSyntax returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsCreateViewCommentSyntax() bool {
	return false
}

// SupportsArrayTypedefWithoutElementType returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsArrayTypedefWithoutElementType() bool {
	return false
}

// SupportsArrayTypedefWithBrackets returns true for BigQueryDialect.
func (d *BigQueryDialect) SupportsArrayTypedefWithBrackets() bool {
	return true
}

// SupportsParensAroundTableFactor returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsParensAroundTableFactor() bool {
	return false
}

// SupportsValuesAsTableFactor returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsValuesAsTableFactor() bool {
	return false
}

// SupportsUnnestTableFactor returns true for BigQueryDialect.
func (d *BigQueryDialect) SupportsUnnestTableFactor() bool {
	return true
}

// SupportsSemanticViewTableFactor returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsSemanticViewTableFactor() bool {
	return false
}

// SupportsTableVersioning returns true for BigQueryDialect.
// See https://cloud.google.com/bigquery/docs/access-historical-data
func (d *BigQueryDialect) SupportsTableVersioning() bool {
	return true
}

// SupportsTableSampleBeforeAlias returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsTableSampleBeforeAlias() bool {
	return false
}

// SupportsTableHints returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsTableHints() bool {
	return false
}

// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
func (d *BigQueryDialect) SupportsIndexHints() bool {
	return false
}

// SupportsAscDescInColumnDefinition returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsAscDescInColumnDefinition() bool {
	return false
}

// SupportsSpaceSeparatedColumnOptions returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsSpaceSeparatedColumnOptions() bool {
	return false
}

// SupportsConstraintKeywordWithoutName returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsConstraintKeywordWithoutName() bool {
	return false
}

// SupportsKeyColumnOption returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsKeyColumnOption() bool {
	return false
}

// SupportsDataTypeSignedSuffix returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsDataTypeSignedSuffix() bool {
	return false
}

// SupportsNestedComments returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsNestedComments() bool {
	return false
}

// SupportsMultilineCommentHints returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsMultilineCommentHints() bool {
	return false
}

// SupportsCommentOptimizerHint returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsCommentOptimizerHint() bool {
	return false
}

// SupportsCommentOn returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsCommentOn() bool {
	return false
}

// RequiresSingleLineCommentWhitespace returns false for BigQueryDialect.
func (d *BigQueryDialect) RequiresSingleLineCommentWhitespace() bool {
	return false
}

// SupportsExplainWithUtilityOptions returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsExplainWithUtilityOptions() bool {
	return false
}

// SupportsExecuteImmediate returns true for BigQueryDialect.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/procedural-language#execute_immediate
func (d *BigQueryDialect) SupportsExecuteImmediate() bool {
	return true
}

// AllowExtractCustom returns false for BigQueryDialect.
func (d *BigQueryDialect) AllowExtractCustom() bool {
	return false
}

// AllowExtractSingleQuotes returns false for BigQueryDialect.
func (d *BigQueryDialect) AllowExtractSingleQuotes() bool {
	return false
}

// SupportsExtractCommaSyntax returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsExtractCommaSyntax() bool {
	return false
}

// SupportsSubqueryAsFunctionArg returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsSubqueryAsFunctionArg() bool {
	return false
}

// SupportsDollarPlaceholder returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsDollarPlaceholder() bool {
	return false
}

// SupportsCreateIndexWithClause returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsCreateIndexWithClause() bool {
	return false
}

// RequireIntervalQualifier returns true for BigQueryDialect.
func (d *BigQueryDialect) RequireIntervalQualifier() bool {
	return true
}

// SupportsIntervalOptions returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsIntervalOptions() bool {
	return false
}

// SupportsFactorialOperator returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsFactorialOperator() bool {
	return false
}

// SupportsBitwiseShiftOperators returns true for BigQueryDialect.
func (d *BigQueryDialect) SupportsBitwiseShiftOperators() bool {
	return true
}

// SupportsNotnullOperator returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsNotnullOperator() bool {
	return false
}

// SupportsBangNotOperator returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsBangNotOperator() bool {
	return false
}

// SupportsDoubleAmpersandOperator returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsDoubleAmpersandOperator() bool {
	return false
}

// SupportsMatchAgainst returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsMatchAgainst() bool {
	return false
}

// SupportsUserHostGrantee returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsUserHostGrantee() bool {
	return false
}

// SupportsListenNotify returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsListenNotify() bool {
	return false
}

// SupportsLoadData returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsLoadData() bool {
	return false
}

// SupportsLoadExtension returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsLoadExtension() bool {
	return false
}

// SupportsTopBeforeDistinct returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsTopBeforeDistinct() bool {
	return false
}

// SupportsBooleanLiterals returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsBooleanLiterals() bool {
	return false
}

// SupportsShowLikeBeforeIn returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsShowLikeBeforeIn() bool {
	return false
}

// SupportsPartiQL returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsPartiQL() bool {
	return false
}

// SupportsEqAliasAssignment returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsEqAliasAssignment() bool {
	return false
}

// SupportsInsertSet returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsInsertSet() bool {
	return false
}

// SupportsInsertTableFunction returns true for BigQueryDialect.
func (d *BigQueryDialect) SupportsInsertTableFunction() bool {
	return true
}

// SupportsInsertTableQuery returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsInsertTableQuery() bool {
	return false
}

// SupportsInsertFormat returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsInsertFormat() bool {
	return false
}

// SupportsInsertTableAlias returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsInsertTableAlias() bool {
	return false
}

// SupportsAlterColumnTypeUsing returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsAlterColumnTypeUsing() bool {
	return false
}

// SupportsCommaSeparatedDropColumnList returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsCommaSeparatedDropColumnList() bool {
	return false
}

// SupportsOrderByAll returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsOrderByAll() bool {
	return false
}

// SupportsGeometricTypes returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsGeometricTypes() bool {
	return false
}

// DescribeRequiresTableKeyword returns false for BigQueryDialect.
func (d *BigQueryDialect) DescribeRequiresTableKeyword() bool {
	return false
}

// SupportsOptimizeTable returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsOptimizeTable() bool {
	return false
}

// SupportsPrewhere returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsPrewhere() bool {
	return false
}

// SupportsWithFill returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsWithFill() bool {
	return false
}

// SupportsLimitBy returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsLimitBy() bool {
	return false
}

// SupportsInterpolate returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsInterpolate() bool {
	return false
}

// SupportsSettings returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsSettings() bool {
	return false
}

// SupportsSelectFormat returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsSelectFormat() bool {
	return false
}

// SupportsInstall returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsInstall() bool {
	return false
}

// SupportsDetach returns false for BigQueryDialect.
func (d *BigQueryDialect) SupportsDetach() bool {
	return false
}

// SupportsCommaSeparatedTrim returns true for BigQueryDialect.
func (d *BigQueryDialect) SupportsCommaSeparatedTrim() bool {
	return true
}

// PrecValue returns the precedence value for a given Precedence level.
func (d *BigQueryDialect) PrecValue(prec dialects.Precedence) uint8 {
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
func (d *BigQueryDialect) PrecUnknown() uint8 {
	return 0
}

// GetNextPrecedence returns 0 to fall back to default behavior.
func (d *BigQueryDialect) GetNextPrecedence(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// GetNextPrecedenceDefault implements the default precedence logic.
func (d *BigQueryDialect) GetNextPrecedenceDefault(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// ParsePrefix returns false to fall back to default behavior.
func (d *BigQueryDialect) ParsePrefix(parser dialects.ParserAccessor) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseInfix returns false to fall back to default behavior.
func (d *BigQueryDialect) ParseInfix(parser dialects.ParserAccessor, expr ast.Expr, precedence uint8) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseStatement handles BigQuery-specific statement parsing.
// Currently handles BEGIN...EXCEPTION...END blocks for procedural language.
func (d *BigQueryDialect) ParseStatement(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	// Check if this is a BEGIN block (not BEGIN TRANSACTION)
	if parser.ParseKeyword("BEGIN") {
		// If next token is TRANSACTION, semicolon, or EOF, it's not a procedural BEGIN
		if parser.PeekKeyword("TRANSACTION") {
			parser.PrevToken()
			return nil, false, nil
		}

		tok := parser.PeekToken()
		if _, isSemicolon := tok.Token.(tokenizer.TokenSemiColon); isSemicolon {
			parser.PrevToken()
			return nil, false, nil
		}
		if _, isEOF := tok.Token.(tokenizer.EOF); isEOF {
			parser.PrevToken()
			return nil, false, nil
		}

		// This is a procedural BEGIN...EXCEPTION...END block
		// The actual parsing would be handled by parser.ParseBeginExceptionEnd()
		// For now, we return false to let the default parser handle it
		parser.PrevToken()
		return nil, false, nil
	}

	return nil, false, nil
}

// ParseColumnOption returns false to fall back to default behavior.
func (d *BigQueryDialect) ParseColumnOption(parser dialects.ParserAccessor) (dialects.ColumnOption, bool, error) {
	return nil, false, nil
}
