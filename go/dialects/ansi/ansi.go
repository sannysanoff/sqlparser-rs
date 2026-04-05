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

// Package ansi provides an ANSI SQL:2011 dialect implementation for the SQL parser.
//
// This dialect implements strict ANSI SQL:2011 compliance with:
//   - Minimal feature set
//   - Standard identifier rules (ASCII letters only)
//   - Standard operator precedence
//   - Double-quoted identifiers
//   - Nested block comments (per SQL standard)
//   - Required interval qualifiers
//
// See https://en.wikipedia.org/wiki/SQL:2011
package ansi

import (
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/token"
)

// AnsiDialect is a dialect for strict ANSI SQL:2011 compliance.
//
// AnsiDialect implements the following capability interfaces from the dialects package:
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
type AnsiDialect struct{}

// Compile-time interface checks to ensure AnsiDialect implements all required interfaces.
var _ dialects.CompleteDialect = (*AnsiDialect)(nil)
var _ dialects.IdentifierDialect = (*AnsiDialect)(nil)
var _ dialects.KeywordDialect = (*AnsiDialect)(nil)
var _ dialects.StringLiteralDialect = (*AnsiDialect)(nil)
var _ dialects.AggregationDialect = (*AnsiDialect)(nil)
var _ dialects.GroupByDialect = (*AnsiDialect)(nil)
var _ dialects.JoinDialect = (*AnsiDialect)(nil)
var _ dialects.TransactionDialect = (*AnsiDialect)(nil)
var _ dialects.NamedArgumentDialect = (*AnsiDialect)(nil)
var _ dialects.SetDialect = (*AnsiDialect)(nil)
var _ dialects.SelectDialect = (*AnsiDialect)(nil)
var _ dialects.TypeConversionDialect = (*AnsiDialect)(nil)
var _ dialects.ObjectReferenceDialect = (*AnsiDialect)(nil)
var _ dialects.InExpressionDialect = (*AnsiDialect)(nil)
var _ dialects.LiteralDialect = (*AnsiDialect)(nil)
var _ dialects.TableDefinitionDialect = (*AnsiDialect)(nil)
var _ dialects.ColumnDefinitionDialect = (*AnsiDialect)(nil)
var _ dialects.CommentDialect = (*AnsiDialect)(nil)
var _ dialects.ExplainDialect = (*AnsiDialect)(nil)
var _ dialects.ExecuteDialect = (*AnsiDialect)(nil)
var _ dialects.ExtractDialect = (*AnsiDialect)(nil)
var _ dialects.SubqueryDialect = (*AnsiDialect)(nil)
var _ dialects.PlaceholderDialect = (*AnsiDialect)(nil)
var _ dialects.IndexDialect = (*AnsiDialect)(nil)
var _ dialects.IntervalDialect = (*AnsiDialect)(nil)
var _ dialects.OperatorDialect = (*AnsiDialect)(nil)
var _ dialects.MatchDialect = (*AnsiDialect)(nil)
var _ dialects.GranteeDialect = (*AnsiDialect)(nil)
var _ dialects.ListenNotifyDialect = (*AnsiDialect)(nil)
var _ dialects.LoadDialect = (*AnsiDialect)(nil)
var _ dialects.TopDistinctDialect = (*AnsiDialect)(nil)
var _ dialects.BooleanLiteralDialect = (*AnsiDialect)(nil)
var _ dialects.ShowDialect = (*AnsiDialect)(nil)
var _ dialects.PartiQLDialect = (*AnsiDialect)(nil)
var _ dialects.AliasDialect = (*AnsiDialect)(nil)
var _ dialects.InsertDialect = (*AnsiDialect)(nil)
var _ dialects.AlterTableDialect = (*AnsiDialect)(nil)
var _ dialects.OrderByDialect = (*AnsiDialect)(nil)
var _ dialects.GeometricDialect = (*AnsiDialect)(nil)
var _ dialects.DescribeDialect = (*AnsiDialect)(nil)
var _ dialects.ClickHouseDialect = (*AnsiDialect)(nil)
var _ dialects.DuckDBDialect = (*AnsiDialect)(nil)
var _ dialects.TrimDialect = (*AnsiDialect)(nil)
var _ dialects.ConnectByDialect = (*AnsiDialect)(nil)

// NewAnsiDialect creates a new instance of AnsiDialect.
func NewAnsiDialect() *AnsiDialect {
	return &AnsiDialect{}
}

// Dialect returns the dialect identifier.
func (d *AnsiDialect) Dialect() string {
	return "ansi"
}

// IsIdentifierStart returns true if the character is a valid start character
// for an unquoted identifier. ANSI only supports ASCII letters (no underscore at start).
func (d *AnsiDialect) IsIdentifierStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// IsIdentifierPart returns true if the character is a valid unquoted
// identifier character. ANSI supports ASCII letters, digits, and underscore.
func (d *AnsiDialect) IsIdentifierPart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') || ch == '_'
}

// IsDelimitedIdentifierStart returns true if the character starts a quoted identifier.
// ANSI uses double quotes for delimited identifiers.
func (d *AnsiDialect) IsDelimitedIdentifierStart(ch rune) bool {
	return ch == '"'
}

// IsNestedDelimitedIdentifierStart returns true if the character starts a
// potential nested quoted identifier. ANSI doesn't support nested identifiers.
func (d *AnsiDialect) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return false
}

// PeekNestedDelimitedIdentifierQuotes returns the outer quote style and
// optional inner quote style if applicable. ANSI doesn't support nested identifiers.
func (d *AnsiDialect) PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*dialects.NestedIdentifierQuote, int) {
	return nil, pos
}

// IdentifierQuoteStyle returns the character used to quote identifiers.
// ANSI uses double quotes for identifier quoting.
func (d *AnsiDialect) IdentifierQuoteStyle(identifier string) *rune {
	quote := rune('"')
	return &quote
}

// IsCustomOperatorPart returns true if the character is part of a custom operator.
// ANSI doesn't support custom operators.
func (d *AnsiDialect) IsCustomOperatorPart(ch rune) bool {
	return false
}

// IsReservedForIdentifier returns true if the keyword is reserved.
func (d *AnsiDialect) IsReservedForIdentifier(kw token.Keyword) bool {
	return token.IsReservedForIdentifier(kw)
}

// GetReservedKeywordsForSelectItemOperator returns reserved keywords for select items.
func (d *AnsiDialect) GetReservedKeywordsForSelectItemOperator() []token.Keyword {
	return nil
}

// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
func (d *AnsiDialect) GetReservedGranteesTypes() []dialects.GranteesType {
	return nil
}

// IsColumnAlias returns true if the keyword should be parsed as a column alias.
func (d *AnsiDialect) IsColumnAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == token.AS || token.IsReservedForColumnAlias(kw)
}

// IsSelectItemAlias returns true if the specified keyword should be parsed as a select item alias.
func (d *AnsiDialect) IsSelectItemAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsColumnAlias(kw, parser)
}

// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier.
func (d *AnsiDialect) IsTableFactor(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == token.TABLE || kw == token.LATERAL
}

// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias.
func (d *AnsiDialect) IsTableAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsTableFactor(kw, parser)
}

// IsTableFactorAlias returns true if the specified keyword should be parsed as a table factor alias.
func (d *AnsiDialect) IsTableFactorAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return explicit || d.IsTableAlias(kw, parser)
}

// IsIdentifierGeneratingFunctionName returns true if the ident is a function that generates identifiers.
func (d *AnsiDialect) IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool {
	return false
}

// SupportsStringLiteralBackslashEscape returns false for AnsiDialect.
func (d *AnsiDialect) SupportsStringLiteralBackslashEscape() bool {
	return false
}

// IgnoresWildcardEscapes returns false for AnsiDialect.
func (d *AnsiDialect) IgnoresWildcardEscapes() bool {
	return false
}

// SupportsUnicodeStringLiteral returns false for AnsiDialect.
func (d *AnsiDialect) SupportsUnicodeStringLiteral() bool {
	return false
}

// SupportsTripleQuotedString returns false for AnsiDialect.
func (d *AnsiDialect) SupportsTripleQuotedString() bool {
	return false
}

// SupportsStringLiteralConcatenation returns false for AnsiDialect.
func (d *AnsiDialect) SupportsStringLiteralConcatenation() bool {
	return false
}

// SupportsStringLiteralConcatenationWithNewline returns false for AnsiDialect.
func (d *AnsiDialect) SupportsStringLiteralConcatenationWithNewline() bool {
	return false
}

// SupportsQuoteDelimitedString returns false for AnsiDialect.
func (d *AnsiDialect) SupportsQuoteDelimitedString() bool {
	return false
}

// SupportsStringEscapeConstant returns false for AnsiDialect.
func (d *AnsiDialect) SupportsStringEscapeConstant() bool {
	return false
}

// SupportsFilterDuringAggregation returns false for AnsiDialect.
func (d *AnsiDialect) SupportsFilterDuringAggregation() bool {
	return false
}

// SupportsWithinAfterArrayAggregation returns false for AnsiDialect.
func (d *AnsiDialect) SupportsWithinAfterArrayAggregation() bool {
	return false
}

// SupportsWindowClauseNamedWindowReference returns false for AnsiDialect.
func (d *AnsiDialect) SupportsWindowClauseNamedWindowReference() bool {
	return false
}

// SupportsWindowFunctionNullTreatmentArg returns false for AnsiDialect.
func (d *AnsiDialect) SupportsWindowFunctionNullTreatmentArg() bool {
	return false
}

// SupportsMatchRecognize returns false for AnsiDialect.
func (d *AnsiDialect) SupportsMatchRecognize() bool {
	return false
}

// SupportsGroupByExpr returns false for AnsiDialect.
func (d *AnsiDialect) SupportsGroupByExpr() bool {
	return false
}

// SupportsGroupByWithModifier returns false for AnsiDialect.
func (d *AnsiDialect) SupportsGroupByWithModifier() bool {
	return false
}

// SupportsLeftAssociativeJoinsWithoutParens returns false for AnsiDialect.
func (d *AnsiDialect) SupportsLeftAssociativeJoinsWithoutParens() bool {
	return false
}

// SupportsOuterJoinOperator returns false for AnsiDialect.
func (d *AnsiDialect) SupportsOuterJoinOperator() bool {
	return false
}

// SupportsCrossJoinConstraint returns false for AnsiDialect.
func (d *AnsiDialect) SupportsCrossJoinConstraint() bool {
	return false
}

// SupportsConnectBy returns false for AnsiDialect.
func (d *AnsiDialect) SupportsConnectBy() bool {
	return false
}

// SupportsStartTransactionModifier returns false for AnsiDialect.
func (d *AnsiDialect) SupportsStartTransactionModifier() bool {
	return false
}

// SupportsEndTransactionModifier returns false for AnsiDialect.
func (d *AnsiDialect) SupportsEndTransactionModifier() bool {
	return false
}

// SupportsNamedFnArgsWithEqOperator returns false for AnsiDialect.
func (d *AnsiDialect) SupportsNamedFnArgsWithEqOperator() bool {
	return false
}

// SupportsNamedFnArgsWithColonOperator returns false for AnsiDialect.
func (d *AnsiDialect) SupportsNamedFnArgsWithColonOperator() bool {
	return false
}

// SupportsNamedFnArgsWithAssignmentOperator returns false for AnsiDialect.
func (d *AnsiDialect) SupportsNamedFnArgsWithAssignmentOperator() bool {
	return false
}

// SupportsNamedFnArgsWithRArrowOperator returns false for AnsiDialect.
func (d *AnsiDialect) SupportsNamedFnArgsWithRArrowOperator() bool {
	return false
}

// SupportsNamedFnArgsWithExprName returns false for AnsiDialect.
func (d *AnsiDialect) SupportsNamedFnArgsWithExprName() bool {
	return false
}

// SupportsParenthesizedSetVariables returns false for AnsiDialect.
func (d *AnsiDialect) SupportsParenthesizedSetVariables() bool {
	return false
}

// SupportsCommaSeparatedSetAssignments returns false for AnsiDialect.
func (d *AnsiDialect) SupportsCommaSeparatedSetAssignments() bool {
	return false
}

// SupportsSetStmtWithoutOperator returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSetStmtWithoutOperator() bool {
	return false
}

// SupportsSetNames returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSetNames() bool {
	return false
}

// SupportsSelectWildcardExcept returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSelectWildcardExcept() bool {
	return false
}

// SupportsSelectWildcardExclude returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSelectWildcardExclude() bool {
	return false
}

// SupportsSelectExclude returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSelectExclude() bool {
	return false
}

// SupportsSelectWildcardReplace returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSelectWildcardReplace() bool {
	return false
}

// SupportsSelectWildcardIlike returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSelectWildcardIlike() bool {
	return false
}

// SupportsSelectWildcardRename returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSelectWildcardRename() bool {
	return false
}

// SupportsSelectWildcardWithAlias returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSelectWildcardWithAlias() bool {
	return false
}

// SupportsSelectExprStar returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSelectExprStar() bool {
	return false
}

// SupportsFromFirstSelect returns false for AnsiDialect.
func (d *AnsiDialect) SupportsFromFirstSelect() bool {
	return false
}

// SupportsEmptyProjections returns false for AnsiDialect.
func (d *AnsiDialect) SupportsEmptyProjections() bool {
	return false
}

// SupportsSelectModifiers returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSelectModifiers() bool {
	return false
}

// SupportsPipeOperator returns false for AnsiDialect.
func (d *AnsiDialect) SupportsPipeOperator() bool {
	return false
}

// SupportsTrailingCommas returns false for AnsiDialect.
func (d *AnsiDialect) SupportsTrailingCommas() bool {
	return false
}

// SupportsProjectionTrailingCommas returns false for AnsiDialect.
func (d *AnsiDialect) SupportsProjectionTrailingCommas() bool {
	return false
}

// SupportsFromTrailingCommas returns false for AnsiDialect.
func (d *AnsiDialect) SupportsFromTrailingCommas() bool {
	return false
}

// SupportsColumnDefinitionTrailingCommas returns false for AnsiDialect.
func (d *AnsiDialect) SupportsColumnDefinitionTrailingCommas() bool {
	return false
}

// SupportsLimitComma returns false for AnsiDialect.
func (d *AnsiDialect) SupportsLimitComma() bool {
	return false
}

// ConvertTypeBeforeValue returns false for AnsiDialect.
func (d *AnsiDialect) ConvertTypeBeforeValue() bool {
	return false
}

// SupportsTryConvert returns false for AnsiDialect.
func (d *AnsiDialect) SupportsTryConvert() bool {
	return false
}

// SupportsBinaryKwAsCast returns false for AnsiDialect.
func (d *AnsiDialect) SupportsBinaryKwAsCast() bool {
	return false
}

// SupportsObjectNameDoubleDotNotation returns false for AnsiDialect.
func (d *AnsiDialect) SupportsObjectNameDoubleDotNotation() bool {
	return false
}

// SupportsNumericPrefix returns false for AnsiDialect.
func (d *AnsiDialect) SupportsNumericPrefix() bool {
	return false
}

// SupportsNumericLiteralUnderscores returns false for AnsiDialect.
func (d *AnsiDialect) SupportsNumericLiteralUnderscores() bool {
	return false
}

// SupportsInEmptyList returns false for AnsiDialect.
func (d *AnsiDialect) SupportsInEmptyList() bool {
	return false
}

// SupportsDictionarySyntax returns false for AnsiDialect.
func (d *AnsiDialect) SupportsDictionarySyntax() bool {
	return false
}

// SupportsMapLiteralSyntax returns false for AnsiDialect.
func (d *AnsiDialect) SupportsMapLiteralSyntax() bool {
	return false
}

// SupportsStructLiteral returns false for AnsiDialect.
func (d *AnsiDialect) SupportsStructLiteral() bool {
	return false
}

// SupportsArrayLiteralSyntax returns false for AnsiDialect.
func (d *AnsiDialect) SupportsArrayLiteralSyntax() bool {
	return false
}

// SupportsLambdaFunctions returns false for AnsiDialect.
func (d *AnsiDialect) SupportsLambdaFunctions() bool {
	return false
}

// SupportsCreateTableMultiSchemaInfoSources returns false for AnsiDialect.
func (d *AnsiDialect) SupportsCreateTableMultiSchemaInfoSources() bool {
	return false
}

// SupportsCreateTableLikeParenthesized returns false for AnsiDialect.
func (d *AnsiDialect) SupportsCreateTableLikeParenthesized() bool {
	return false
}

// SupportsCreateTableSelect returns false for AnsiDialect.
func (d *AnsiDialect) SupportsCreateTableSelect() bool {
	return false
}

// SupportsCreateViewCommentSyntax returns false for AnsiDialect.
func (d *AnsiDialect) SupportsCreateViewCommentSyntax() bool {
	return false
}

// SupportsArrayTypedefWithoutElementType returns false for AnsiDialect.
func (d *AnsiDialect) SupportsArrayTypedefWithoutElementType() bool {
	return false
}

// SupportsArrayTypedefWithBrackets returns false for AnsiDialect.
func (d *AnsiDialect) SupportsArrayTypedefWithBrackets() bool {
	return false
}

// SupportsParensAroundTableFactor returns false for AnsiDialect.
func (d *AnsiDialect) SupportsParensAroundTableFactor() bool {
	return false
}

// SupportsValuesAsTableFactor returns false for AnsiDialect.
func (d *AnsiDialect) SupportsValuesAsTableFactor() bool {
	return false
}

// SupportsUnnestTableFactor returns false for AnsiDialect.
func (d *AnsiDialect) SupportsUnnestTableFactor() bool {
	return false
}

// SupportsSemanticViewTableFactor returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSemanticViewTableFactor() bool {
	return false
}

// SupportsTableVersioning returns false for AnsiDialect.
func (d *AnsiDialect) SupportsTableVersioning() bool {
	return false
}

// SupportsTableSampleBeforeAlias returns false for AnsiDialect.
func (d *AnsiDialect) SupportsTableSampleBeforeAlias() bool {
	return false
}

// SupportsTableHints returns false for AnsiDialect.
func (d *AnsiDialect) SupportsTableHints() bool {
	return false
}

// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
func (d *AnsiDialect) SupportsIndexHints() bool {
	return false
}

// SupportsAscDescInColumnDefinition returns false for AnsiDialect.
func (d *AnsiDialect) SupportsAscDescInColumnDefinition() bool {
	return false
}

// SupportsSpaceSeparatedColumnOptions returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSpaceSeparatedColumnOptions() bool {
	return false
}

// SupportsConstraintKeywordWithoutName returns false for AnsiDialect.
func (d *AnsiDialect) SupportsConstraintKeywordWithoutName() bool {
	return false
}

// SupportsKeyColumnOption returns false for AnsiDialect.
func (d *AnsiDialect) SupportsKeyColumnOption() bool {
	return false
}

// SupportsDataTypeSignedSuffix returns false for AnsiDialect.
func (d *AnsiDialect) SupportsDataTypeSignedSuffix() bool {
	return false
}

// SupportsNestedComments returns true for AnsiDialect.
// The SQL standard explicitly states that block comments nest.
func (d *AnsiDialect) SupportsNestedComments() bool {
	return true
}

// SupportsMultilineCommentHints returns false for AnsiDialect.
func (d *AnsiDialect) SupportsMultilineCommentHints() bool {
	return false
}

// SupportsCommentOptimizerHint returns false for AnsiDialect.
func (d *AnsiDialect) SupportsCommentOptimizerHint() bool {
	return false
}

// SupportsCommentOn returns false for AnsiDialect.
func (d *AnsiDialect) SupportsCommentOn() bool {
	return false
}

// RequiresSingleLineCommentWhitespace returns false for AnsiDialect.
func (d *AnsiDialect) RequiresSingleLineCommentWhitespace() bool {
	return false
}

// SupportsExplainWithUtilityOptions returns false for AnsiDialect.
func (d *AnsiDialect) SupportsExplainWithUtilityOptions() bool {
	return false
}

// SupportsExecuteImmediate returns false for AnsiDialect.
func (d *AnsiDialect) SupportsExecuteImmediate() bool {
	return false
}

// AllowExtractCustom returns false for AnsiDialect.
func (d *AnsiDialect) AllowExtractCustom() bool {
	return false
}

// AllowExtractSingleQuotes returns false for AnsiDialect.
func (d *AnsiDialect) AllowExtractSingleQuotes() bool {
	return false
}

// SupportsExtractCommaSyntax returns false for AnsiDialect.
func (d *AnsiDialect) SupportsExtractCommaSyntax() bool {
	return false
}

// SupportsSubqueryAsFunctionArg returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSubqueryAsFunctionArg() bool {
	return false
}

// SupportsDollarPlaceholder returns false for AnsiDialect.
func (d *AnsiDialect) SupportsDollarPlaceholder() bool {
	return false
}

// SupportsCreateIndexWithClause returns false for AnsiDialect.
func (d *AnsiDialect) SupportsCreateIndexWithClause() bool {
	return false
}

// RequireIntervalQualifier returns true for AnsiDialect.
// ANSI requires units to be specified in INTERVAL expressions.
func (d *AnsiDialect) RequireIntervalQualifier() bool {
	return true
}

// SupportsIntervalOptions returns false for AnsiDialect.
func (d *AnsiDialect) SupportsIntervalOptions() bool {
	return false
}

// SupportsFactorialOperator returns false for AnsiDialect.
func (d *AnsiDialect) SupportsFactorialOperator() bool {
	return false
}

// SupportsBitwiseShiftOperators returns false for AnsiDialect.
func (d *AnsiDialect) SupportsBitwiseShiftOperators() bool {
	return false
}

// SupportsNotnullOperator returns false for AnsiDialect.
func (d *AnsiDialect) SupportsNotnullOperator() bool {
	return false
}

// SupportsBangNotOperator returns false for AnsiDialect.
func (d *AnsiDialect) SupportsBangNotOperator() bool {
	return false
}

// SupportsDoubleAmpersandOperator returns false for AnsiDialect.
func (d *AnsiDialect) SupportsDoubleAmpersandOperator() bool {
	return false
}

// SupportsMatchAgainst returns false for AnsiDialect.
func (d *AnsiDialect) SupportsMatchAgainst() bool {
	return false
}

// SupportsUserHostGrantee returns false for AnsiDialect.
func (d *AnsiDialect) SupportsUserHostGrantee() bool {
	return false
}

// SupportsListenNotify returns false for AnsiDialect.
func (d *AnsiDialect) SupportsListenNotify() bool {
	return false
}

// SupportsLoadData returns false for AnsiDialect.
func (d *AnsiDialect) SupportsLoadData() bool {
	return false
}

// SupportsLoadExtension returns false for AnsiDialect.
func (d *AnsiDialect) SupportsLoadExtension() bool {
	return false
}

// SupportsTopBeforeDistinct returns false for AnsiDialect.
func (d *AnsiDialect) SupportsTopBeforeDistinct() bool {
	return false
}

// SupportsBooleanLiterals returns false for AnsiDialect.
func (d *AnsiDialect) SupportsBooleanLiterals() bool {
	return false
}

// SupportsShowLikeBeforeIn returns false for AnsiDialect.
func (d *AnsiDialect) SupportsShowLikeBeforeIn() bool {
	return false
}

// SupportsPartiQL returns false for AnsiDialect.
func (d *AnsiDialect) SupportsPartiQL() bool {
	return false
}

// SupportsEqAliasAssignment returns false for AnsiDialect.
func (d *AnsiDialect) SupportsEqAliasAssignment() bool {
	return false
}

// SupportsInsertSet returns false for AnsiDialect.
func (d *AnsiDialect) SupportsInsertSet() bool {
	return false
}

// SupportsInsertTableFunction returns false for AnsiDialect.
func (d *AnsiDialect) SupportsInsertTableFunction() bool {
	return false
}

// SupportsInsertTableQuery returns false for AnsiDialect.
func (d *AnsiDialect) SupportsInsertTableQuery() bool {
	return false
}

// SupportsInsertFormat returns false for AnsiDialect.
func (d *AnsiDialect) SupportsInsertFormat() bool {
	return false
}

// SupportsInsertTableAlias returns false for AnsiDialect.
func (d *AnsiDialect) SupportsInsertTableAlias() bool {
	return false
}

// SupportsAlterColumnTypeUsing returns false for AnsiDialect.
func (d *AnsiDialect) SupportsAlterColumnTypeUsing() bool {
	return false
}

// SupportsCommaSeparatedDropColumnList returns false for AnsiDialect.
func (d *AnsiDialect) SupportsCommaSeparatedDropColumnList() bool {
	return false
}

// SupportsOrderByAll returns false for AnsiDialect.
func (d *AnsiDialect) SupportsOrderByAll() bool {
	return false
}

// SupportsGeometricTypes returns false for AnsiDialect.
func (d *AnsiDialect) SupportsGeometricTypes() bool {
	return false
}

// DescribeRequiresTableKeyword returns false for AnsiDialect.
func (d *AnsiDialect) DescribeRequiresTableKeyword() bool {
	return false
}

// SupportsOptimizeTable returns false for AnsiDialect.
func (d *AnsiDialect) SupportsOptimizeTable() bool {
	return false
}

// SupportsPrewhere returns false for AnsiDialect.
func (d *AnsiDialect) SupportsPrewhere() bool {
	return false
}

// SupportsWithFill returns false for AnsiDialect.
func (d *AnsiDialect) SupportsWithFill() bool {
	return false
}

// SupportsLimitBy returns false for AnsiDialect.
func (d *AnsiDialect) SupportsLimitBy() bool {
	return false
}

// SupportsInterpolate returns false for AnsiDialect.
func (d *AnsiDialect) SupportsInterpolate() bool {
	return false
}

// SupportsSettings returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSettings() bool {
	return false
}

// SupportsSelectFormat returns false for AnsiDialect.
func (d *AnsiDialect) SupportsSelectFormat() bool {
	return false
}

// SupportsInstall returns false for AnsiDialect.
func (d *AnsiDialect) SupportsInstall() bool {
	return false
}

// SupportsDetach returns false for AnsiDialect.
func (d *AnsiDialect) SupportsDetach() bool {
	return false
}

// SupportsCommaSeparatedTrim returns false for AnsiDialect.
func (d *AnsiDialect) SupportsCommaSeparatedTrim() bool {
	return false
}

// Precedence constants for ANSI operators.
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
func (d *AnsiDialect) PrecValue(prec dialects.Precedence) uint8 {
	switch {
	case prec == dialects.PrecedencePeriod:
		return periodPrec
	case prec == dialects.PrecedenceDoubleColon:
		return doubleColonPrec
	case prec == dialects.PrecedenceCollate:
		return 42
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
func (d *AnsiDialect) PrecUnknown() uint8 {
	return 0
}

// GetNextPrecedence returns 0 to fall back to default behavior.
func (d *AnsiDialect) GetNextPrecedence(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// GetNextPrecedenceDefault implements the default precedence logic.
func (d *AnsiDialect) GetNextPrecedenceDefault(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// ParsePrefix returns false to fall back to default behavior.
func (d *AnsiDialect) ParsePrefix(parser dialects.ParserAccessor) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseInfix returns false to fall back to default behavior.
func (d *AnsiDialect) ParseInfix(parser dialects.ParserAccessor, expr ast.Expr, precedence uint8) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseStatement returns false to fall back to default behavior.
func (d *AnsiDialect) ParseStatement(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	return nil, false, nil
}

// ParseColumnOption returns false to fall back to default behavior.
func (d *AnsiDialect) ParseColumnOption(parser dialects.ParserAccessor) (dialects.ColumnOption, bool, error) {
	return nil, false, nil
}
