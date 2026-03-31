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

// Package databricks provides a Databricks SQL dialect implementation for the SQL parser.
//
// This dialect implements Databricks-specific parsing behavior including:
//   - Based on Spark SQL
//   - Backtick-quoted identifiers (`identifier`)
//   - LATERAL VIEW support
//   - TRANSFORM support
//   - CLUSTER BY, DISTRIBUTE BY, SORT BY clauses
//   - Delta Lake features (table versioning, time travel)
//   - Supports EXCEPT in wildcards
//   - Lambda function support
//   - STRUCT literal syntax
//   - OPTIMIZE TABLE support
//   - Nested comments
//
// See https://docs.databricks.com/en/sql/language-manual/index.html
package databricks

import (
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/token"
)

// DatabricksDialect is a dialect for Databricks SQL.
type DatabricksDialect struct{}

// NewDatabricksDialect creates a new instance of DatabricksDialect.
func NewDatabricksDialect() *DatabricksDialect {
	return &DatabricksDialect{}
}

// Dialect returns the dialect identifier.
func (d *DatabricksDialect) Dialect() string {
	return "databricks"
}

// IsIdentifierStart returns true if the character is a valid start character
// for an unquoted identifier. Databricks supports letters and underscore.
// See https://docs.databricks.com/en/sql/language-manual/sql-ref-identifiers.html
func (d *DatabricksDialect) IsIdentifierStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// IsIdentifierPart returns true if the character is a valid unquoted
// identifier character. Databricks supports letters, digits, and underscore.
func (d *DatabricksDialect) IsIdentifierPart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') || ch == '_'
}

// IsDelimitedIdentifierStart returns true if the character starts a quoted identifier.
// Databricks uses backticks for delimited identifiers.
func (d *DatabricksDialect) IsDelimitedIdentifierStart(ch rune) bool {
	return ch == '`'
}

// IsNestedDelimitedIdentifierStart returns true if the character starts a
// potential nested quoted identifier. Databricks doesn't support nested identifiers.
func (d *DatabricksDialect) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return false
}

// PeekNestedDelimitedIdentifierQuotes returns the outer quote style and
// optional inner quote style if applicable. Databricks doesn't support nested identifiers.
func (d *DatabricksDialect) PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*dialects.NestedIdentifierQuote, int) {
	return nil, pos
}

// IdentifierQuoteStyle returns the character used to quote identifiers.
// Databricks uses backticks for identifier quoting.
func (d *DatabricksDialect) IdentifierQuoteStyle(identifier string) *rune {
	quote := rune('`')
	return &quote
}

// IsCustomOperatorPart returns true if the character is part of a custom operator.
func (d *DatabricksDialect) IsCustomOperatorPart(ch rune) bool {
	return false
}

// IsReservedForIdentifier returns true if the keyword is reserved.
func (d *DatabricksDialect) IsReservedForIdentifier(kw token.Keyword) bool {
	return token.IsReservedForIdentifier(kw)
}

// GetReservedKeywordsForSelectItemOperator returns reserved keywords for select items.
func (d *DatabricksDialect) GetReservedKeywordsForSelectItemOperator() []token.Keyword {
	return nil
}

// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
func (d *DatabricksDialect) GetReservedGranteesTypes() []dialects.GranteesType {
	return nil
}

// IsColumnAlias returns true if the keyword should be parsed as a column alias.
func (d *DatabricksDialect) IsColumnAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == token.AS || token.IsReservedForColumnAlias(kw)
}

// IsSelectItemAlias returns true if the specified keyword should be parsed as a select item alias.
func (d *DatabricksDialect) IsSelectItemAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsColumnAlias(kw, parser)
}

// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier.
func (d *DatabricksDialect) IsTableFactor(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == token.TABLE || kw == token.LATERAL || kw == token.PIVOT || kw == token.UNPIVOT
}

// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias.
func (d *DatabricksDialect) IsTableAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsTableFactor(kw, parser)
}

// IsTableFactorAlias returns true if the specified keyword should be parsed as a table factor alias.
func (d *DatabricksDialect) IsTableFactorAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return explicit || d.IsTableAlias(kw, parser)
}

// IsIdentifierGeneratingFunctionName returns true if the ident is a function that generates identifiers.
func (d *DatabricksDialect) IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool {
	return false
}

// SupportsStringLiteralBackslashEscape returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsStringLiteralBackslashEscape() bool {
	return false
}

// IgnoresWildcardEscapes returns false for DatabricksDialect.
func (d *DatabricksDialect) IgnoresWildcardEscapes() bool {
	return false
}

// SupportsUnicodeStringLiteral returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsUnicodeStringLiteral() bool {
	return false
}

// SupportsTripleQuotedString returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsTripleQuotedString() bool {
	return false
}

// SupportsStringLiteralConcatenation returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsStringLiteralConcatenation() bool {
	return false
}

// SupportsStringLiteralConcatenationWithNewline returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsStringLiteralConcatenationWithNewline() bool {
	return false
}

// SupportsQuoteDelimitedString returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsQuoteDelimitedString() bool {
	return false
}

// SupportsStringEscapeConstant returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsStringEscapeConstant() bool {
	return false
}

// SupportsFilterDuringAggregation returns true for DatabricksDialect.
// See https://docs.databricks.com/en/sql/language-manual/sql-ref-syntax-qry-select-groupby.html
func (d *DatabricksDialect) SupportsFilterDuringAggregation() bool {
	return true
}

// SupportsWithinAfterArrayAggregation returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsWithinAfterArrayAggregation() bool {
	return false
}

// SupportsWindowClauseNamedWindowReference returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsWindowClauseNamedWindowReference() bool {
	return true
}

// SupportsWindowFunctionNullTreatmentArg returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsWindowFunctionNullTreatmentArg() bool {
	return true
}

// SupportsMatchRecognize returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsMatchRecognize() bool {
	return false
}

// SupportsGroupByExpr returns true for DatabricksDialect.
// See https://docs.databricks.com/en/sql/language-manual/sql-ref-syntax-qry-select-groupby.html
func (d *DatabricksDialect) SupportsGroupByExpr() bool {
	return true
}

// SupportsGroupByWithModifier returns true for DatabricksDialect.
// See https://docs.databricks.com/en/sql/language-manual/sql-ref-syntax-qry-select-groupby.html
func (d *DatabricksDialect) SupportsGroupByWithModifier() bool {
	return true
}

// SupportsLeftAssociativeJoinsWithoutParens returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsLeftAssociativeJoinsWithoutParens() bool {
	return false
}

// SupportsOuterJoinOperator returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsOuterJoinOperator() bool {
	return false
}

// SupportsCrossJoinConstraint returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsCrossJoinConstraint() bool {
	return false
}

// SupportsConnectBy returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsConnectBy() bool {
	return false
}

// SupportsStartTransactionModifier returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsStartTransactionModifier() bool {
	return false
}

// SupportsEndTransactionModifier returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsEndTransactionModifier() bool {
	return false
}

// SupportsNamedFnArgsWithEqOperator returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsNamedFnArgsWithEqOperator() bool {
	return false
}

// SupportsNamedFnArgsWithColonOperator returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsNamedFnArgsWithColonOperator() bool {
	return false
}

// SupportsNamedFnArgsWithAssignmentOperator returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsNamedFnArgsWithAssignmentOperator() bool {
	return false
}

// SupportsNamedFnArgsWithRArrowOperator returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsNamedFnArgsWithRArrowOperator() bool {
	return true
}

// SupportsNamedFnArgsWithExprName returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsNamedFnArgsWithExprName() bool {
	return false
}

// SupportsParenthesizedSetVariables returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsParenthesizedSetVariables() bool {
	return false
}

// SupportsCommaSeparatedSetAssignments returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsCommaSeparatedSetAssignments() bool {
	return false
}

// SupportsSetStmtWithoutOperator returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsSetStmtWithoutOperator() bool {
	return false
}

// SupportsSetNames returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsSetNames() bool {
	return true
}

// SupportsSelectWildcardExcept returns true for DatabricksDialect.
// See https://docs.databricks.com/en/sql/language-manual/sql-ref-syntax-qry-select.html#syntax
func (d *DatabricksDialect) SupportsSelectWildcardExcept() bool {
	return true
}

// SupportsSelectWildcardExclude returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsSelectWildcardExclude() bool {
	return false
}

// SupportsSelectExclude returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsSelectExclude() bool {
	return false
}

// SupportsSelectWildcardReplace returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsSelectWildcardReplace() bool {
	return false
}

// SupportsSelectWildcardIlike returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsSelectWildcardIlike() bool {
	return false
}

// SupportsSelectWildcardRename returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsSelectWildcardRename() bool {
	return false
}

// SupportsSelectWildcardWithAlias returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsSelectWildcardWithAlias() bool {
	return false
}

// SupportsSelectExprStar returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsSelectExprStar() bool {
	return false
}

// SupportsFromFirstSelect returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsFromFirstSelect() bool {
	return false
}

// SupportsEmptyProjections returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsEmptyProjections() bool {
	return false
}

// SupportsSelectModifiers returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsSelectModifiers() bool {
	return false
}

// SupportsPipeOperator returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsPipeOperator() bool {
	return false
}

// SupportsTrailingCommas returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsTrailingCommas() bool {
	return false
}

// SupportsProjectionTrailingCommas returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsProjectionTrailingCommas() bool {
	return false
}

// SupportsFromTrailingCommas returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsFromTrailingCommas() bool {
	return false
}

// SupportsColumnDefinitionTrailingCommas returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsColumnDefinitionTrailingCommas() bool {
	return false
}

// SupportsLimitComma returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsLimitComma() bool {
	return false
}

// ConvertTypeBeforeValue returns false for DatabricksDialect.
func (d *DatabricksDialect) ConvertTypeBeforeValue() bool {
	return false
}

// SupportsTryConvert returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsTryConvert() bool {
	return true
}

// SupportsBinaryKwAsCast returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsBinaryKwAsCast() bool {
	return false
}

// SupportsObjectNameDoubleDotNotation returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsObjectNameDoubleDotNotation() bool {
	return false
}

// SupportsNumericPrefix returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsNumericPrefix() bool {
	return false
}

// SupportsNumericLiteralUnderscores returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsNumericLiteralUnderscores() bool {
	return true
}

// SupportsInEmptyList returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsInEmptyList() bool {
	return false
}

// SupportsDictionarySyntax returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsDictionarySyntax() bool {
	return true
}

// SupportsMapLiteralSyntax returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsMapLiteralSyntax() bool {
	return true
}

// SupportsStructLiteral returns true for DatabricksDialect.
// See https://docs.databricks.com/en/sql/language-manual/functions/struct.html
func (d *DatabricksDialect) SupportsStructLiteral() bool {
	return true
}

// SupportsArrayLiteralSyntax returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsArrayLiteralSyntax() bool {
	return true
}

// SupportsLambdaFunctions returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsLambdaFunctions() bool {
	return true
}

// SupportsCreateTableMultiSchemaInfoSources returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsCreateTableMultiSchemaInfoSources() bool {
	return false
}

// SupportsCreateTableLikeParenthesized returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsCreateTableLikeParenthesized() bool {
	return false
}

// SupportsCreateTableSelect returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsCreateTableSelect() bool {
	return true
}

// SupportsCreateViewCommentSyntax returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsCreateViewCommentSyntax() bool {
	return false
}

// SupportsArrayTypedefWithoutElementType returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsArrayTypedefWithoutElementType() bool {
	return false
}

// SupportsArrayTypedefWithBrackets returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsArrayTypedefWithBrackets() bool {
	return true
}

// SupportsParensAroundTableFactor returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsParensAroundTableFactor() bool {
	return true
}

// SupportsValuesAsTableFactor returns true for DatabricksDialect.
// See https://docs.databricks.com/en/sql/language-manual/sql-ref-syntax-qry-select-values.html
func (d *DatabricksDialect) SupportsValuesAsTableFactor() bool {
	return true
}

// SupportsSemanticViewTableFactor returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsSemanticViewTableFactor() bool {
	return false
}

// SupportsTableVersioning returns true for DatabricksDialect.
// See https://docs.databricks.com/gcp/en/delta/history#delta-time-travel-syntax
func (d *DatabricksDialect) SupportsTableVersioning() bool {
	return true
}

// SupportsTableSampleBeforeAlias returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsTableSampleBeforeAlias() bool {
	return false
}

// SupportsTableHints returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsTableHints() bool {
	return true
}

// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
func (d *DatabricksDialect) SupportsIndexHints() bool {
	return false
}

// SupportsAscDescInColumnDefinition returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsAscDescInColumnDefinition() bool {
	return false
}

// SupportsSpaceSeparatedColumnOptions returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsSpaceSeparatedColumnOptions() bool {
	return false
}

// SupportsConstraintKeywordWithoutName returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsConstraintKeywordWithoutName() bool {
	return false
}

// SupportsKeyColumnOption returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsKeyColumnOption() bool {
	return false
}

// SupportsDataTypeSignedSuffix returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsDataTypeSignedSuffix() bool {
	return false
}

// SupportsNestedComments returns true for DatabricksDialect.
// See https://docs.databricks.com/aws/en/sql/language-manual/sql-ref-syntax-comment
func (d *DatabricksDialect) SupportsNestedComments() bool {
	return true
}

// SupportsMultilineCommentHints returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsMultilineCommentHints() bool {
	return false
}

// SupportsCommentOptimizerHint returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsCommentOptimizerHint() bool {
	return false
}

// SupportsCommentOn returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsCommentOn() bool {
	return true
}

// RequiresSingleLineCommentWhitespace returns false for DatabricksDialect.
func (d *DatabricksDialect) RequiresSingleLineCommentWhitespace() bool {
	return false
}

// SupportsExplainWithUtilityOptions returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsExplainWithUtilityOptions() bool {
	return true
}

// SupportsExecuteImmediate returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsExecuteImmediate() bool {
	return false
}

// AllowExtractCustom returns true for DatabricksDialect.
func (d *DatabricksDialect) AllowExtractCustom() bool {
	return true
}

// AllowExtractSingleQuotes returns false for DatabricksDialect.
func (d *DatabricksDialect) AllowExtractSingleQuotes() bool {
	return false
}

// SupportsExtractCommaSyntax returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsExtractCommaSyntax() bool {
	return false
}

// SupportsSubqueryAsFunctionArg returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsSubqueryAsFunctionArg() bool {
	return false
}

// SupportsDollarPlaceholder returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsDollarPlaceholder() bool {
	return false
}

// SupportsCreateIndexWithClause returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsCreateIndexWithClause() bool {
	return true
}

// RequireIntervalQualifier returns true for DatabricksDialect.
func (d *DatabricksDialect) RequireIntervalQualifier() bool {
	return true
}

// SupportsIntervalOptions returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsIntervalOptions() bool {
	return false
}

// SupportsFactorialOperator returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsFactorialOperator() bool {
	return false
}

// SupportsBitwiseShiftOperators returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsBitwiseShiftOperators() bool {
	return true
}

// SupportsNotnullOperator returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsNotnullOperator() bool {
	return false
}

// SupportsBangNotOperator returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsBangNotOperator() bool {
	return false
}

// SupportsDoubleAmpersandOperator returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsDoubleAmpersandOperator() bool {
	return false
}

// SupportsMatchAgainst returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsMatchAgainst() bool {
	return false
}

// SupportsUserHostGrantee returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsUserHostGrantee() bool {
	return false
}

// SupportsListenNotify returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsListenNotify() bool {
	return false
}

// SupportsLoadData returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsLoadData() bool {
	return false
}

// SupportsLoadExtension returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsLoadExtension() bool {
	return false
}

// SupportsTopBeforeDistinct returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsTopBeforeDistinct() bool {
	return false
}

// SupportsBooleanLiterals returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsBooleanLiterals() bool {
	return true
}

// SupportsShowLikeBeforeIn returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsShowLikeBeforeIn() bool {
	return false
}

// SupportsPartiQL returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsPartiQL() bool {
	return false
}

// SupportsEqAliasAssignment returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsEqAliasAssignment() bool {
	return false
}

// SupportsInsertSet returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsInsertSet() bool {
	return false
}

// SupportsInsertTableFunction returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsInsertTableFunction() bool {
	return false
}

// SupportsInsertTableQuery returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsInsertTableQuery() bool {
	return false
}

// SupportsInsertFormat returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsInsertFormat() bool {
	return false
}

// SupportsInsertTableAlias returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsInsertTableAlias() bool {
	return true
}

// SupportsAlterColumnTypeUsing returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsAlterColumnTypeUsing() bool {
	return true
}

// SupportsCommaSeparatedDropColumnList returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsCommaSeparatedDropColumnList() bool {
	return true
}

// SupportsOrderByAll returns true for DatabricksDialect.
func (d *DatabricksDialect) SupportsOrderByAll() bool {
	return true
}

// SupportsGeometricTypes returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsGeometricTypes() bool {
	return false
}

// DescribeRequiresTableKeyword returns false for DatabricksDialect.
func (d *DatabricksDialect) DescribeRequiresTableKeyword() bool {
	return false
}

// SupportsOptimizeTable returns true for DatabricksDialect.
// See https://docs.databricks.com/en/sql/language-manual/delta-optimize.html
func (d *DatabricksDialect) SupportsOptimizeTable() bool {
	return true
}

// SupportsPrewhere returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsPrewhere() bool {
	return false
}

// SupportsWithFill returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsWithFill() bool {
	return false
}

// SupportsLimitBy returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsLimitBy() bool {
	return false
}

// SupportsInterpolate returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsInterpolate() bool {
	return false
}

// SupportsSettings returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsSettings() bool {
	return false
}

// SupportsSelectFormat returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsSelectFormat() bool {
	return false
}

// SupportsInstall returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsInstall() bool {
	return false
}

// SupportsDetach returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsDetach() bool {
	return false
}

// SupportsCommaSeparatedTrim returns false for DatabricksDialect.
func (d *DatabricksDialect) SupportsCommaSeparatedTrim() bool {
	return false
}

// Precedence constants for Databricks operators.
const (
	periodPrec      uint8 = 100
	doubleColonPrec uint8 = 50
	atTzPrec        uint8 = 41
	mulDivModPrec   uint8 = 40
	plusMinusPrec   uint8 = 30
	xorPrec         uint8 = 24
	ampersandPrec   uint8 = 23
	caretPrec       uint8 = 22
	pipePrec        uint8 = 21
	colonPrec       uint8 = 21
	betweenPrec     uint8 = 20
	eqPrec          uint8 = 20
	likePrec        uint8 = 19
	isPrec          uint8 = 17
	pgOtherPrec     uint8 = 16
	notPrec         uint8 = 15
	andPrec         uint8 = 10
	orPrec          uint8 = 5
)

// PrecValue returns the precedence value for a given Precedence level.
func (d *DatabricksDialect) PrecValue(prec dialects.Precedence) uint8 {
	switch {
	case prec == dialects.PrecedencePeriod:
		return periodPrec
	case prec == dialects.PrecedenceDoubleColon:
		return doubleColonPrec
	case prec == dialects.PrecedenceAtTz:
		return atTzPrec
	case prec == dialects.PrecedenceMulDivModOp:
		return mulDivModPrec
	case prec == dialects.PrecedencePlusMinus:
		return plusMinusPrec
	case prec == dialects.PrecedenceXor:
		return xorPrec
	case prec == dialects.PrecedenceAmpersand:
		return ampersandPrec
	case prec == dialects.PrecedenceCaret:
		return caretPrec
	case prec == dialects.PrecedencePipe:
		return pipePrec
	case prec == dialects.PrecedenceColon:
		return colonPrec
	case prec == dialects.PrecedenceBetween:
		return betweenPrec
	case prec == dialects.PrecedenceEq:
		return eqPrec
	case prec == dialects.PrecedenceLike:
		return likePrec
	case prec == dialects.PrecedenceIs:
		return isPrec
	case prec == dialects.PrecedencePgOther:
		return pgOtherPrec
	case prec == dialects.PrecedenceUnaryNot:
		return notPrec
	case prec == dialects.PrecedenceAnd:
		return andPrec
	case prec == dialects.PrecedenceOr:
		return orPrec
	default:
		return d.PrecUnknown()
	}
}

// PrecUnknown returns the precedence when precedence is otherwise unknown.
func (d *DatabricksDialect) PrecUnknown() uint8 {
	return 0
}

// GetNextPrecedence returns 0 to fall back to default behavior.
func (d *DatabricksDialect) GetNextPrecedence(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// GetNextPrecedenceDefault implements the default precedence logic.
func (d *DatabricksDialect) GetNextPrecedenceDefault(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// ParsePrefix returns false to fall back to default behavior.
func (d *DatabricksDialect) ParsePrefix(parser dialects.ParserAccessor) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseInfix returns false to fall back to default behavior.
func (d *DatabricksDialect) ParseInfix(parser dialects.ParserAccessor, expr ast.Expr, precedence uint8) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseStatement returns false to fall back to default behavior.
func (d *DatabricksDialect) ParseStatement(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	return nil, false, nil
}

// ParseColumnOption returns false to fall back to default behavior.
func (d *DatabricksDialect) ParseColumnOption(parser dialects.ParserAccessor) (dialects.ColumnOption, bool, error) {
	return nil, false, nil
}
