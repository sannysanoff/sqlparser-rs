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

package generic

import (
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/token"
)

// GenericDialect is a permissive, general purpose dialect which parses a wide
// variety of SQL statements from many different dialects.
//
// GenericDialect implements the following capability interfaces from the dialects package:
//   - dialects.CoreDialect
//   - dialects.IdentifierDialect
//   - dialects.KeywordDialect
//   - dialects.StringLiteralDialect
//   - dialects.AggregationDialect
//   - dialects.GroupByDialect
//   - dialects.JoinDialect
//   - dialects.TransactionDialect
//   - dialects.NamedArgumentDialect
//   - dialects.SetDialect
//   - dialects.SelectDialect
//   - dialects.TypeConversionDialect
//   - dialects.ObjectReferenceDialect
//   - dialects.InExpressionDialect
//   - dialects.LiteralDialect
//   - dialects.TableDefinitionDialect
//   - dialects.ColumnDefinitionDialect
//   - dialects.CommentDialect
//   - dialects.ExplainDialect
//   - dialects.ExecuteDialect
//   - dialects.ExtractDialect
//   - dialects.SubqueryDialect
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
//   - dialects.InsertDialect
//   - dialects.AlterTableDialect
//   - dialects.OrderByDialect
//   - dialects.GeometricDialect
//   - dialects.DescribeDialect
//   - dialects.ClickHouseDialect
//   - dialects.DuckDBDialect
//   - dialects.TrimDialect
//   - dialects.ConnectByDialect
//   - dialects.CompleteDialect (via parseriface.CompleteDialect)
type GenericDialect struct{}

// Compile-time interface checks to ensure GenericDialect implements all required interfaces.
var _ dialects.CompleteDialect = (*GenericDialect)(nil)
var _ dialects.IdentifierDialect = (*GenericDialect)(nil)
var _ dialects.KeywordDialect = (*GenericDialect)(nil)
var _ dialects.StringLiteralDialect = (*GenericDialect)(nil)
var _ dialects.AggregationDialect = (*GenericDialect)(nil)
var _ dialects.GroupByDialect = (*GenericDialect)(nil)
var _ dialects.JoinDialect = (*GenericDialect)(nil)
var _ dialects.TransactionDialect = (*GenericDialect)(nil)
var _ dialects.NamedArgumentDialect = (*GenericDialect)(nil)
var _ dialects.SetDialect = (*GenericDialect)(nil)
var _ dialects.SelectDialect = (*GenericDialect)(nil)
var _ dialects.TypeConversionDialect = (*GenericDialect)(nil)
var _ dialects.ObjectReferenceDialect = (*GenericDialect)(nil)
var _ dialects.InExpressionDialect = (*GenericDialect)(nil)
var _ dialects.LiteralDialect = (*GenericDialect)(nil)
var _ dialects.TableDefinitionDialect = (*GenericDialect)(nil)
var _ dialects.ColumnDefinitionDialect = (*GenericDialect)(nil)
var _ dialects.CommentDialect = (*GenericDialect)(nil)
var _ dialects.ExplainDialect = (*GenericDialect)(nil)
var _ dialects.ExecuteDialect = (*GenericDialect)(nil)
var _ dialects.ExtractDialect = (*GenericDialect)(nil)
var _ dialects.SubqueryDialect = (*GenericDialect)(nil)
var _ dialects.PlaceholderDialect = (*GenericDialect)(nil)
var _ dialects.IndexDialect = (*GenericDialect)(nil)
var _ dialects.IntervalDialect = (*GenericDialect)(nil)
var _ dialects.OperatorDialect = (*GenericDialect)(nil)
var _ dialects.MatchDialect = (*GenericDialect)(nil)
var _ dialects.GranteeDialect = (*GenericDialect)(nil)
var _ dialects.ListenNotifyDialect = (*GenericDialect)(nil)
var _ dialects.LoadDialect = (*GenericDialect)(nil)
var _ dialects.TopDistinctDialect = (*GenericDialect)(nil)
var _ dialects.BooleanLiteralDialect = (*GenericDialect)(nil)
var _ dialects.ShowDialect = (*GenericDialect)(nil)
var _ dialects.PartiQLDialect = (*GenericDialect)(nil)
var _ dialects.AliasDialect = (*GenericDialect)(nil)
var _ dialects.InsertDialect = (*GenericDialect)(nil)
var _ dialects.AlterTableDialect = (*GenericDialect)(nil)
var _ dialects.OrderByDialect = (*GenericDialect)(nil)
var _ dialects.GeometricDialect = (*GenericDialect)(nil)
var _ dialects.DescribeDialect = (*GenericDialect)(nil)
var _ dialects.ClickHouseDialect = (*GenericDialect)(nil)
var _ dialects.DuckDBDialect = (*GenericDialect)(nil)
var _ dialects.TrimDialect = (*GenericDialect)(nil)
var _ dialects.ConnectByDialect = (*GenericDialect)(nil)

// NewGenericDialect creates a new instance of GenericDialect.
func NewGenericDialect() *GenericDialect {
	return &GenericDialect{}
}

// Dialect returns the dialect identifier.
func (d *GenericDialect) Dialect() string {
	return "generic"
}

// IsIdentifierStart returns true if the character is a valid start character
// for an unquoted identifier.
func (d *GenericDialect) IsIdentifierStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == '#' || ch == '@'
}

// IsIdentifierPart returns true if the character is a valid unquoted
// identifier character.
func (d *GenericDialect) IsIdentifierPart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '@' || ch == '$' || ch == '#' || ch == '_'
}

// IsDelimitedIdentifierStart returns true if the character starts a quoted identifier.
func (d *GenericDialect) IsDelimitedIdentifierStart(ch rune) bool {
	return ch == '"' || ch == '`'
}

// IsNestedDelimitedIdentifierStart returns true if the character starts a
// potential nested quoted identifier.
func (d *GenericDialect) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return false
}

// PeekNestedDelimitedIdentifierQuotes returns the outer quote style and
// optional inner quote style if applicable.
func (d *GenericDialect) PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*dialects.NestedIdentifierQuote, int) {
	return nil, pos
}

// IdentifierQuoteStyle returns the character used to quote identifiers.
func (d *GenericDialect) IdentifierQuoteStyle(identifier string) *rune {
	return nil
}

// IsCustomOperatorPart returns true if the character is part of a custom operator.
func (d *GenericDialect) IsCustomOperatorPart(ch rune) bool {
	return false
}

// IsReservedForIdentifier returns true if the keyword is reserved.
func (d *GenericDialect) IsReservedForIdentifier(kw token.Keyword) bool {
	return token.IsReservedForIdentifier(kw)
}

// GetReservedKeywordsForSelectItemOperator returns reserved keywords for select items.
func (d *GenericDialect) GetReservedKeywordsForSelectItemOperator() []token.Keyword {
	return nil
}

// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
func (d *GenericDialect) GetReservedGranteesTypes() []dialects.GranteesType {
	return nil
}

// IsColumnAlias returns true if the keyword should be parsed as a column alias.
func (d *GenericDialect) IsColumnAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == "AS" || token.IsReservedForColumnAlias(kw)
}

// IsSelectItemAlias returns true if the specified keyword should be parsed as a select item alias
func (d *GenericDialect) IsSelectItemAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsColumnAlias(kw, parser)
}

// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier
func (d *GenericDialect) IsTableFactor(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == "TABLE" || kw == "LATERAL" || kw == "PIVOT" || kw == "UNPIVOT"
}

// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias
func (d *GenericDialect) IsTableAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsTableFactor(kw, parser)
}

// IsTableFactorAlias returns true if the specified keyword should be parsed as a table factor alias
func (d *GenericDialect) IsTableFactorAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsTableAlias(kw, parser)
}

// IsIdentifierGeneratingFunctionName returns true if the ident is a function that generates identifiers.
func (d *GenericDialect) IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool {
	return false
}

// SupportsStringLiteralBackslashEscape returns true for GenericDialect.
func (d *GenericDialect) SupportsStringLiteralBackslashEscape() bool {
	return true
}

// IgnoresWildcardEscapes returns false for GenericDialect.
func (d *GenericDialect) IgnoresWildcardEscapes() bool {
	return false
}

// SupportsUnicodeStringLiteral returns true for GenericDialect.
func (d *GenericDialect) SupportsUnicodeStringLiteral() bool {
	return true
}

// SupportsTripleQuotedString returns false for GenericDialect.
func (d *GenericDialect) SupportsTripleQuotedString() bool {
	return false
}

// SupportsStringLiteralConcatenation returns false for GenericDialect.
func (d *GenericDialect) SupportsStringLiteralConcatenation() bool {
	return false
}

// SupportsStringLiteralConcatenationWithNewline returns false for GenericDialect.
func (d *GenericDialect) SupportsStringLiteralConcatenationWithNewline() bool {
	return false
}

// SupportsQuoteDelimitedString returns true for GenericDialect.
func (d *GenericDialect) SupportsQuoteDelimitedString() bool {
	return true
}

// SupportsStringEscapeConstant returns true for GenericDialect.
func (d *GenericDialect) SupportsStringEscapeConstant() bool {
	return true
}

// SupportsFilterDuringAggregation returns true for GenericDialect.
func (d *GenericDialect) SupportsFilterDuringAggregation() bool {
	return true
}

// SupportsWithinAfterArrayAggregation returns false for GenericDialect.
func (d *GenericDialect) SupportsWithinAfterArrayAggregation() bool {
	return false
}

// SupportsWindowClauseNamedWindowReference returns true for GenericDialect.
func (d *GenericDialect) SupportsWindowClauseNamedWindowReference() bool {
	return true
}

// SupportsWindowFunctionNullTreatmentArg returns true for GenericDialect.
func (d *GenericDialect) SupportsWindowFunctionNullTreatmentArg() bool {
	return true
}

// SupportsMatchRecognize returns true for GenericDialect.
func (d *GenericDialect) SupportsMatchRecognize() bool {
	return true
}

// SupportsGroupByExpr returns true for GenericDialect.
func (d *GenericDialect) SupportsGroupByExpr() bool {
	return true
}

// SupportsGroupByWithModifier returns true for GenericDialect.
func (d *GenericDialect) SupportsGroupByWithModifier() bool {
	return true
}

// SupportsLeftAssociativeJoinsWithoutParens returns true for GenericDialect.
func (d *GenericDialect) SupportsLeftAssociativeJoinsWithoutParens() bool {
	return true
}

// SupportsOuterJoinOperator returns false for GenericDialect.
func (d *GenericDialect) SupportsOuterJoinOperator() bool {
	return false
}

// SupportsCrossJoinConstraint returns false for GenericDialect.
func (d *GenericDialect) SupportsCrossJoinConstraint() bool {
	return false
}

// SupportsConnectBy returns true for GenericDialect.
func (d *GenericDialect) SupportsConnectBy() bool {
	return true
}

// SupportsStartTransactionModifier returns true for GenericDialect.
func (d *GenericDialect) SupportsStartTransactionModifier() bool {
	return true
}

// SupportsEndTransactionModifier returns false for GenericDialect.
func (d *GenericDialect) SupportsEndTransactionModifier() bool {
	return false
}

// SupportsNamedFnArgsWithEqOperator returns false for GenericDialect.
func (d *GenericDialect) SupportsNamedFnArgsWithEqOperator() bool {
	return false
}

// SupportsNamedFnArgsWithColonOperator returns false for GenericDialect.
func (d *GenericDialect) SupportsNamedFnArgsWithColonOperator() bool {
	return false
}

// SupportsNamedFnArgsWithAssignmentOperator returns true for GenericDialect.
func (d *GenericDialect) SupportsNamedFnArgsWithAssignmentOperator() bool {
	return true
}

// SupportsNamedFnArgsWithRArrowOperator returns true for GenericDialect.
func (d *GenericDialect) SupportsNamedFnArgsWithRArrowOperator() bool {
	return true
}

// SupportsNamedFnArgsWithExprName returns false for GenericDialect.
func (d *GenericDialect) SupportsNamedFnArgsWithExprName() bool {
	return false
}

// SupportsParenthesizedSetVariables returns true for GenericDialect.
func (d *GenericDialect) SupportsParenthesizedSetVariables() bool {
	return true
}

// SupportsCommaSeparatedSetAssignments returns true for GenericDialect.
func (d *GenericDialect) SupportsCommaSeparatedSetAssignments() bool {
	return true
}

// SupportsSetStmtWithoutOperator returns false for GenericDialect.
func (d *GenericDialect) SupportsSetStmtWithoutOperator() bool {
	return false
}

// SupportsSetNames returns true for GenericDialect.
func (d *GenericDialect) SupportsSetNames() bool {
	return true
}

// SupportsSelectWildcardExcept returns true for GenericDialect.
func (d *GenericDialect) SupportsSelectWildcardExcept() bool {
	return true
}

// SupportsSelectWildcardExclude returns true for GenericDialect.
func (d *GenericDialect) SupportsSelectWildcardExclude() bool {
	return true
}

// SupportsSelectExclude returns false for GenericDialect.
func (d *GenericDialect) SupportsSelectExclude() bool {
	return false
}

// SupportsSelectWildcardReplace returns true for GenericDialect.
func (d *GenericDialect) SupportsSelectWildcardReplace() bool {
	return true
}

// SupportsSelectWildcardIlike returns true for GenericDialect.
func (d *GenericDialect) SupportsSelectWildcardIlike() bool {
	return true
}

// SupportsSelectWildcardRename returns true for GenericDialect.
func (d *GenericDialect) SupportsSelectWildcardRename() bool {
	return true
}

// SupportsSelectWildcardWithAlias returns false for GenericDialect.
func (d *GenericDialect) SupportsSelectWildcardWithAlias() bool {
	return false
}

// SupportsSelectExprStar returns false for GenericDialect.
func (d *GenericDialect) SupportsSelectExprStar() bool {
	return false
}

// SupportsFromFirstSelect returns true for GenericDialect.
func (d *GenericDialect) SupportsFromFirstSelect() bool {
	return true
}

// SupportsEmptyProjections returns true for GenericDialect.
func (d *GenericDialect) SupportsEmptyProjections() bool {
	return true
}

// SupportsSelectModifiers returns false for GenericDialect.
func (d *GenericDialect) SupportsSelectModifiers() bool {
	return false
}

// SupportsPipeOperator returns true for GenericDialect.
func (d *GenericDialect) SupportsPipeOperator() bool {
	return true
}

// SupportsTrailingCommas returns true for GenericDialect.
func (d *GenericDialect) SupportsTrailingCommas() bool {
	return true
}

// SupportsProjectionTrailingCommas returns true for GenericDialect.
func (d *GenericDialect) SupportsProjectionTrailingCommas() bool {
	return true
}

// SupportsFromTrailingCommas returns false for GenericDialect.
func (d *GenericDialect) SupportsFromTrailingCommas() bool {
	return false
}

// SupportsColumnDefinitionTrailingCommas returns false for GenericDialect.
func (d *GenericDialect) SupportsColumnDefinitionTrailingCommas() bool {
	return false
}

// SupportsLimitComma returns true for GenericDialect.
func (d *GenericDialect) SupportsLimitComma() bool {
	return true
}

// ConvertTypeBeforeValue returns false for GenericDialect.
func (d *GenericDialect) ConvertTypeBeforeValue() bool {
	return false
}

// SupportsTryConvert returns true for GenericDialect.
func (d *GenericDialect) SupportsTryConvert() bool {
	return true
}

// SupportsBinaryKwAsCast returns false for GenericDialect.
func (d *GenericDialect) SupportsBinaryKwAsCast() bool {
	return false
}

// SupportsObjectNameDoubleDotNotation returns false for GenericDialect.
func (d *GenericDialect) SupportsObjectNameDoubleDotNotation() bool {
	return false
}

// SupportsNumericPrefix returns false for GenericDialect.
func (d *GenericDialect) SupportsNumericPrefix() bool {
	return false
}

// SupportsNumericLiteralUnderscores returns false for GenericDialect.
func (d *GenericDialect) SupportsNumericLiteralUnderscores() bool {
	return false
}

// SupportsInEmptyList returns false for GenericDialect.
func (d *GenericDialect) SupportsInEmptyList() bool {
	return false
}

// SupportsDictionarySyntax returns true for GenericDialect.
func (d *GenericDialect) SupportsDictionarySyntax() bool {
	return true
}

// SupportsMapLiteralSyntax returns true for GenericDialect.
func (d *GenericDialect) SupportsMapLiteralSyntax() bool {
	return true
}

// SupportsStructLiteral returns true for GenericDialect.
func (d *GenericDialect) SupportsStructLiteral() bool {
	return true
}

// SupportsArrayLiteralSyntax returns false for GenericDialect.
func (d *GenericDialect) SupportsArrayLiteralSyntax() bool {
	return false
}

// SupportsLambdaFunctions returns false for GenericDialect.
func (d *GenericDialect) SupportsLambdaFunctions() bool {
	return false
}

// SupportsCreateTableMultiSchemaInfoSources returns false for GenericDialect.
func (d *GenericDialect) SupportsCreateTableMultiSchemaInfoSources() bool {
	return false
}

// SupportsCreateTableLikeParenthesized returns false for GenericDialect.
func (d *GenericDialect) SupportsCreateTableLikeParenthesized() bool {
	return false
}

// SupportsCreateTableSelect returns false for GenericDialect.
func (d *GenericDialect) SupportsCreateTableSelect() bool {
	return false
}

// SupportsCreateViewCommentSyntax returns true for GenericDialect.
func (d *GenericDialect) SupportsCreateViewCommentSyntax() bool {
	return true
}

// SupportsArrayTypedefWithoutElementType returns false for GenericDialect.
func (d *GenericDialect) SupportsArrayTypedefWithoutElementType() bool {
	return false
}

// SupportsArrayTypedefWithBrackets returns true for GenericDialect.
func (d *GenericDialect) SupportsArrayTypedefWithBrackets() bool {
	return true
}

// SupportsParensAroundTableFactor returns true for GenericDialect.
func (d *GenericDialect) SupportsParensAroundTableFactor() bool {
	return true
}

// SupportsValuesAsTableFactor returns true for GenericDialect.
func (d *GenericDialect) SupportsValuesAsTableFactor() bool {
	return true
}

// SupportsUnnestTableFactor returns true for GenericDialect.
// Reference: src/parser/mod.rs:15646 - dialect_of!(self is BigQueryDialect | PostgreSqlDialect | GenericDialect)
func (d *GenericDialect) SupportsUnnestTableFactor() bool {
	return true
}

// SupportsSemanticViewTableFactor returns false for GenericDialect.
func (d *GenericDialect) SupportsSemanticViewTableFactor() bool {
	return false
}

// SupportsTableVersioning returns false for GenericDialect.
func (d *GenericDialect) SupportsTableVersioning() bool {
	return false
}

// SupportsTableSampleBeforeAlias returns false for GenericDialect.
func (d *GenericDialect) SupportsTableSampleBeforeAlias() bool {
	return false
}

// SupportsTableHints returns false for GenericDialect.
func (d *GenericDialect) SupportsTableHints() bool {
	return false
}

// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
func (d *GenericDialect) SupportsIndexHints() bool {
	return false
}

// SupportsAscDescInColumnDefinition returns true for GenericDialect.
func (d *GenericDialect) SupportsAscDescInColumnDefinition() bool {
	return true
}

// SupportsSpaceSeparatedColumnOptions returns false for GenericDialect.
func (d *GenericDialect) SupportsSpaceSeparatedColumnOptions() bool {
	return false
}

// SupportsConstraintKeywordWithoutName returns true for GenericDialect.
func (d *GenericDialect) SupportsConstraintKeywordWithoutName() bool {
	return true
}

// SupportsKeyColumnOption returns true for GenericDialect.
func (d *GenericDialect) SupportsKeyColumnOption() bool {
	return true
}

// SupportsDataTypeSignedSuffix returns true for GenericDialect.
func (d *GenericDialect) SupportsDataTypeSignedSuffix() bool {
	return true
}

// SupportsNestedComments returns true for GenericDialect.
func (d *GenericDialect) SupportsNestedComments() bool {
	return true
}

// SupportsMultilineCommentHints returns true for GenericDialect.
func (d *GenericDialect) SupportsMultilineCommentHints() bool {
	return true
}

// SupportsCommentOptimizerHint returns true for GenericDialect.
func (d *GenericDialect) SupportsCommentOptimizerHint() bool {
	return true
}

// SupportsCommentOn returns true for GenericDialect.
func (d *GenericDialect) SupportsCommentOn() bool {
	return true
}

// RequiresSingleLineCommentWhitespace returns false for GenericDialect.
func (d *GenericDialect) RequiresSingleLineCommentWhitespace() bool {
	return false
}

// SupportsExplainWithUtilityOptions returns true for GenericDialect.
func (d *GenericDialect) SupportsExplainWithUtilityOptions() bool {
	return true
}

// SupportsExecuteImmediate returns false for GenericDialect.
func (d *GenericDialect) SupportsExecuteImmediate() bool {
	return false
}

// AllowExtractCustom returns true for GenericDialect.
func (d *GenericDialect) AllowExtractCustom() bool {
	return true
}

// AllowExtractSingleQuotes returns true for GenericDialect.
func (d *GenericDialect) AllowExtractSingleQuotes() bool {
	return true
}

// SupportsExtractCommaSyntax returns true for GenericDialect.
func (d *GenericDialect) SupportsExtractCommaSyntax() bool {
	return true
}

// SupportsSubqueryAsFunctionArg returns false for GenericDialect.
func (d *GenericDialect) SupportsSubqueryAsFunctionArg() bool {
	return false
}

// SupportsDollarPlaceholder returns false for GenericDialect.
func (d *GenericDialect) SupportsDollarPlaceholder() bool {
	return false
}

// SupportsCreateIndexWithClause returns true for GenericDialect.
func (d *GenericDialect) SupportsCreateIndexWithClause() bool {
	return true
}

// RequireIntervalQualifier returns false for GenericDialect.
func (d *GenericDialect) RequireIntervalQualifier() bool {
	return false
}

// SupportsIntervalOptions returns true for GenericDialect.
func (d *GenericDialect) SupportsIntervalOptions() bool {
	return true
}

// SupportsFactorialOperator returns false for GenericDialect.
func (d *GenericDialect) SupportsFactorialOperator() bool {
	return false
}

// SupportsBitwiseShiftOperators returns true for GenericDialect.
func (d *GenericDialect) SupportsBitwiseShiftOperators() bool {
	return true
}

// SupportsNotnullOperator returns false for GenericDialect.
func (d *GenericDialect) SupportsNotnullOperator() bool {
	return false
}

// SupportsBangNotOperator returns false for GenericDialect.
func (d *GenericDialect) SupportsBangNotOperator() bool {
	return false
}

// SupportsDoubleAmpersandOperator returns false for GenericDialect.
func (d *GenericDialect) SupportsDoubleAmpersandOperator() bool {
	return false
}

// SupportsMatchAgainst returns true for GenericDialect.
func (d *GenericDialect) SupportsMatchAgainst() bool {
	return true
}

// SupportsUserHostGrantee returns true for GenericDialect.
func (d *GenericDialect) SupportsUserHostGrantee() bool {
	return true
}

// SupportsListenNotify returns false for GenericDialect.
func (d *GenericDialect) SupportsListenNotify() bool {
	return false
}

// SupportsLoadData returns false for GenericDialect.
func (d *GenericDialect) SupportsLoadData() bool {
	return false
}

// SupportsLoadExtension returns true for GenericDialect.
func (d *GenericDialect) SupportsLoadExtension() bool {
	return true
}

// SupportsTopBeforeDistinct returns false for GenericDialect.
func (d *GenericDialect) SupportsTopBeforeDistinct() bool {
	return false
}

// SupportsBooleanLiterals returns true for GenericDialect.
func (d *GenericDialect) SupportsBooleanLiterals() bool {
	return true
}

// SupportsShowLikeBeforeIn returns false for GenericDialect.
func (d *GenericDialect) SupportsShowLikeBeforeIn() bool {
	return false
}

// SupportsPartiQL returns false for GenericDialect.
func (d *GenericDialect) SupportsPartiQL() bool {
	return false
}

// SupportsEqAliasAssignment returns false for GenericDialect.
func (d *GenericDialect) SupportsEqAliasAssignment() bool {
	return false
}

// SupportsInsertSet returns false for GenericDialect.
func (d *GenericDialect) SupportsInsertSet() bool {
	return false
}

// SupportsInsertTableFunction returns false for GenericDialect.
func (d *GenericDialect) SupportsInsertTableFunction() bool {
	return false
}

// SupportsInsertTableQuery returns false for GenericDialect.
func (d *GenericDialect) SupportsInsertTableQuery() bool {
	return false
}

// SupportsInsertFormat returns false for GenericDialect.
func (d *GenericDialect) SupportsInsertFormat() bool {
	return false
}

// SupportsInsertTableAlias returns false for GenericDialect.
func (d *GenericDialect) SupportsInsertTableAlias() bool {
	return false
}

// SupportsAlterColumnTypeUsing returns false for GenericDialect.
func (d *GenericDialect) SupportsAlterColumnTypeUsing() bool {
	return false
}

// SupportsCommaSeparatedDropColumnList returns false for GenericDialect.
func (d *GenericDialect) SupportsCommaSeparatedDropColumnList() bool {
	return false
}

// SupportsOrderByAll returns false for GenericDialect.
func (d *GenericDialect) SupportsOrderByAll() bool {
	return false
}

// SupportsGeometricTypes returns false for GenericDialect.
func (d *GenericDialect) SupportsGeometricTypes() bool {
	return false
}

// DescribeRequiresTableKeyword returns false for GenericDialect.
func (d *GenericDialect) DescribeRequiresTableKeyword() bool {
	return false
}

// SupportsOptimizeTable returns true for GenericDialect.
func (d *GenericDialect) SupportsOptimizeTable() bool {
	return true
}

// SupportsPrewhere returns true for GenericDialect.
func (d *GenericDialect) SupportsPrewhere() bool {
	return true
}

// SupportsWithFill returns true for GenericDialect.
func (d *GenericDialect) SupportsWithFill() bool {
	return true
}

// SupportsLimitBy returns true for GenericDialect.
func (d *GenericDialect) SupportsLimitBy() bool {
	return true
}

// SupportsInterpolate returns true for GenericDialect.
func (d *GenericDialect) SupportsInterpolate() bool {
	return true
}

// SupportsSettings returns true for GenericDialect.
func (d *GenericDialect) SupportsSettings() bool {
	return true
}

// SupportsSelectFormat returns true for GenericDialect.
func (d *GenericDialect) SupportsSelectFormat() bool {
	return true
}

// SupportsInstall returns true for GenericDialect.
func (d *GenericDialect) SupportsInstall() bool {
	return true
}

// SupportsDetach returns true for GenericDialect.
func (d *GenericDialect) SupportsDetach() bool {
	return true
}

// SupportsCommaSeparatedTrim returns true for GenericDialect.
func (d *GenericDialect) SupportsCommaSeparatedTrim() bool {
	return true
}

// PrecValue returns the precedence value for a given Precedence level.
func (d *GenericDialect) PrecValue(prec dialects.Precedence) uint8 {
	switch {
	case prec == dialects.PrecedencePeriod:
		return 100
	case prec == dialects.PrecedenceDoubleColon:
		return 50
	case prec == dialects.PrecedenceCollate:
		return 42
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
func (d *GenericDialect) PrecUnknown() uint8 {
	return 0
}

// GetNextPrecedence returns 0 to fall back to default behavior.
func (d *GenericDialect) GetNextPrecedence(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// GetNextPrecedenceDefault implements the default precedence logic.
func (d *GenericDialect) GetNextPrecedenceDefault(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// ParsePrefix returns false to fall back to default behavior.
func (d *GenericDialect) ParsePrefix(parser dialects.ParserAccessor) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseInfix returns false to fall back to default behavior.
func (d *GenericDialect) ParseInfix(parser dialects.ParserAccessor, expr ast.Expr, precedence uint8) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseStatement returns false to fall back to default behavior.
func (d *GenericDialect) ParseStatement(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	return nil, false, nil
}

// ParseColumnOption returns false to fall back to default behavior.
func (d *GenericDialect) ParseColumnOption(parser dialects.ParserAccessor) (dialects.ColumnOption, bool, error) {
	return nil, false, nil
}
