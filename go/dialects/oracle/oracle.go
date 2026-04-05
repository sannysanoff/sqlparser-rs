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

// Package oracle provides an Oracle Database dialect implementation for the SQL parser.
//
// This dialect implements Oracle-specific parsing behavior including:
//   - Double-quoted identifiers
//   - CONNECT BY for hierarchical queries
//   - (+) outer join operator syntax
//   - DUAL table support
//   - PL/SQL blocks
//   - Oracle-specific functions
//   - String concatenation operator (||)
//   - Quote-delimited string literals (Q'[...]')
//   - EXECUTE IMMEDIATE statements
//   - MATCH_RECOGNIZE support
//   - Window function NULL treatment in args
//   - Comment optimizer hints
//
// See https://docs.oracle.com/en/database/oracle/oracle-database/21/sqlrf/index.html
package oracle

import (
	"unicode"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/token"
)

// Reserved keyword for CONNECT_BY_ROOT
var reservedKeywordsForSelectItemOperator = []token.Keyword{token.CONNECT_BY_ROOT}

// OracleDialect is a dialect for Oracle Database.
//
// OracleDialect implements the following capability interfaces:
//   - dialects.CoreDialect
//   - dialects.IdentifierDialect
//   - dialects.KeywordDialect
//   - dialects.StringLiteralDialect (with quote-delimited strings)
//   - dialects.AggregationDialect (with FILTER, MATCH_RECOGNIZE, window function null treatment)
//   - dialects.GroupByDialect (with GROUP BY expressions)
//   - dialects.JoinDialect (with outer join operator (+))
//   - dialects.TransactionDialect
//   - dialects.NamedArgumentDialect (with assignment operator :=)
//   - dialects.SetDialect (with SET without operator)
//   - dialects.SelectDialect
//   - dialects.TypeConversionDialect
//   - dialects.ObjectReferenceDialect
//   - dialects.InExpressionDialect
//   - dialects.LiteralDialect
//   - dialects.TableDefinitionDialect (with CREATE TABLE SELECT)
//   - dialects.ColumnDefinitionDialect
//   - dialects.CommentDialect (with optimizer hints)
//   - dialects.ExplainDialect
//   - dialects.ExecuteDialect (EXECUTE IMMEDIATE)
//   - dialects.ExtractDialect (with custom extract)
//   - dialects.SubqueryDialect
//   - dialects.PlaceholderDialect
//   - dialects.IndexDialect (CREATE INDEX WITH clause)
//   - dialects.IntervalDialect
//   - dialects.OperatorDialect (with && operator)
//   - dialects.MatchDialect
//   - dialects.GranteeDialect
//   - dialects.ListenNotifyDialect
//   - dialects.LoadDialect
//   - dialects.TopDistinctDialect
//   - dialects.BooleanLiteralDialect
//   - dialects.ShowDialect
//   - dialects.PartiQLDialect
//   - dialects.AliasDialect
//   - dialects.InsertDialect (with table alias and table query)
//   - dialects.AlterTableDialect
//   - dialects.OrderByDialect
//   - dialects.GeometricDialect
//   - dialects.DescribeDialect
//   - dialects.ClickHouseDialect
//   - dialects.DuckDBDialect
//   - dialects.TrimDialect
//   - dialects.ConnectByDialect (CONNECT BY hierarchical queries)
type OracleDialect struct{}

// Compile-time interface checks
var _ dialects.CoreDialect = (*OracleDialect)(nil)
var _ dialects.IdentifierDialect = (*OracleDialect)(nil)
var _ dialects.KeywordDialect = (*OracleDialect)(nil)
var _ dialects.StringLiteralDialect = (*OracleDialect)(nil)
var _ dialects.AggregationDialect = (*OracleDialect)(nil)
var _ dialects.GroupByDialect = (*OracleDialect)(nil)
var _ dialects.JoinDialect = (*OracleDialect)(nil)
var _ dialects.TransactionDialect = (*OracleDialect)(nil)
var _ dialects.NamedArgumentDialect = (*OracleDialect)(nil)
var _ dialects.SetDialect = (*OracleDialect)(nil)
var _ dialects.SelectDialect = (*OracleDialect)(nil)
var _ dialects.TypeConversionDialect = (*OracleDialect)(nil)
var _ dialects.ObjectReferenceDialect = (*OracleDialect)(nil)
var _ dialects.InExpressionDialect = (*OracleDialect)(nil)
var _ dialects.LiteralDialect = (*OracleDialect)(nil)
var _ dialects.TableDefinitionDialect = (*OracleDialect)(nil)
var _ dialects.ColumnDefinitionDialect = (*OracleDialect)(nil)
var _ dialects.CommentDialect = (*OracleDialect)(nil)
var _ dialects.ExplainDialect = (*OracleDialect)(nil)
var _ dialects.ExecuteDialect = (*OracleDialect)(nil)
var _ dialects.ExtractDialect = (*OracleDialect)(nil)
var _ dialects.SubqueryDialect = (*OracleDialect)(nil)
var _ dialects.PlaceholderDialect = (*OracleDialect)(nil)
var _ dialects.IndexDialect = (*OracleDialect)(nil)
var _ dialects.IntervalDialect = (*OracleDialect)(nil)
var _ dialects.OperatorDialect = (*OracleDialect)(nil)
var _ dialects.MatchDialect = (*OracleDialect)(nil)
var _ dialects.GranteeDialect = (*OracleDialect)(nil)
var _ dialects.ListenNotifyDialect = (*OracleDialect)(nil)
var _ dialects.LoadDialect = (*OracleDialect)(nil)
var _ dialects.TopDistinctDialect = (*OracleDialect)(nil)
var _ dialects.BooleanLiteralDialect = (*OracleDialect)(nil)
var _ dialects.ShowDialect = (*OracleDialect)(nil)
var _ dialects.PartiQLDialect = (*OracleDialect)(nil)
var _ dialects.AliasDialect = (*OracleDialect)(nil)
var _ dialects.InsertDialect = (*OracleDialect)(nil)
var _ dialects.AlterTableDialect = (*OracleDialect)(nil)
var _ dialects.OrderByDialect = (*OracleDialect)(nil)
var _ dialects.GeometricDialect = (*OracleDialect)(nil)
var _ dialects.DescribeDialect = (*OracleDialect)(nil)
var _ dialects.ClickHouseDialect = (*OracleDialect)(nil)
var _ dialects.DuckDBDialect = (*OracleDialect)(nil)
var _ dialects.TrimDialect = (*OracleDialect)(nil)
var _ dialects.ConnectByDialect = (*OracleDialect)(nil)

// NewOracleDialect creates a new instance of OracleDialect.
func NewOracleDialect() *OracleDialect {
	return &OracleDialect{}
}

// Dialect returns the dialect identifier.
func (d *OracleDialect) Dialect() string {
	return "oracle"
}

// IsIdentifierStart returns true if the character is a valid start character
// for an unquoted identifier. Oracle supports alphabetic characters.
func (d *OracleDialect) IsIdentifierStart(ch rune) bool {
	return unicode.IsLetter(ch)
}

// IsIdentifierPart returns true if the character is a valid unquoted
// identifier character. Oracle supports alphanumeric, underscore, $, #, and @.
func (d *OracleDialect) IsIdentifierPart(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '$' || ch == '#' || ch == '@'
}

// IsDelimitedIdentifierStart returns true if the character starts a quoted identifier.
// Oracle uses double quotes for delimited identifiers.
func (d *OracleDialect) IsDelimitedIdentifierStart(ch rune) bool {
	return ch == '"'
}

// IsNestedDelimitedIdentifierStart returns true if the character starts a
// potential nested quoted identifier. Oracle doesn't support nested identifiers.
func (d *OracleDialect) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return false
}

// PeekNestedDelimitedIdentifierQuotes returns the outer quote style and
// optional inner quote style if applicable. Oracle doesn't support nested identifiers.
func (d *OracleDialect) PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*dialects.NestedIdentifierQuote, int) {
	return nil, pos
}

// IdentifierQuoteStyle returns the character used to quote identifiers.
// Oracle uses double quotes for identifier quoting.
func (d *OracleDialect) IdentifierQuoteStyle(identifier string) *rune {
	quote := rune('"')
	return &quote
}

// IsCustomOperatorPart returns true if the character is part of a custom operator.
func (d *OracleDialect) IsCustomOperatorPart(ch rune) bool {
	return false
}

// IsReservedForIdentifier returns true if the keyword is reserved.
func (d *OracleDialect) IsReservedForIdentifier(kw token.Keyword) bool {
	return token.IsReservedForIdentifier(kw)
}

// GetReservedKeywordsForSelectItemOperator returns reserved keywords for select items.
// Oracle reserves CONNECT_BY_ROOT.
func (d *OracleDialect) GetReservedKeywordsForSelectItemOperator() []token.Keyword {
	return reservedKeywordsForSelectItemOperator
}

// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
func (d *OracleDialect) GetReservedGranteesTypes() []dialects.GranteesType {
	return nil
}

// IsColumnAlias returns true if the keyword should be parsed as a column alias.
func (d *OracleDialect) IsColumnAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == token.AS || token.IsReservedForColumnAlias(kw)
}

// IsSelectItemAlias returns true if the specified keyword should be parsed as a select item alias.
func (d *OracleDialect) IsSelectItemAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsColumnAlias(kw, parser)
}

// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier.
func (d *OracleDialect) IsTableFactor(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == token.TABLE || kw == token.LATERAL || kw == token.PIVOT || kw == token.UNPIVOT
}

// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias.
func (d *OracleDialect) IsTableAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsTableFactor(kw, parser)
}

// IsTableFactorAlias returns true if the specified keyword should be parsed as a table factor alias.
func (d *OracleDialect) IsTableFactorAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return explicit || d.IsTableAlias(kw, parser)
}

// IsIdentifierGeneratingFunctionName returns true if the ident is a function that generates identifiers.
func (d *OracleDialect) IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool {
	return false
}

// SupportsStringLiteralBackslashEscape returns false for OracleDialect.
func (d *OracleDialect) SupportsStringLiteralBackslashEscape() bool {
	return false
}

// IgnoresWildcardEscapes returns false for OracleDialect.
func (d *OracleDialect) IgnoresWildcardEscapes() bool {
	return false
}

// SupportsUnicodeStringLiteral returns false for OracleDialect.
func (d *OracleDialect) SupportsUnicodeStringLiteral() bool {
	return false
}

// SupportsTripleQuotedString returns false for OracleDialect.
func (d *OracleDialect) SupportsTripleQuotedString() bool {
	return false
}

// SupportsStringLiteralConcatenation returns false for OracleDialect.
func (d *OracleDialect) SupportsStringLiteralConcatenation() bool {
	return false
}

// SupportsStringLiteralConcatenationWithNewline returns false for OracleDialect.
func (d *OracleDialect) SupportsStringLiteralConcatenationWithNewline() bool {
	return false
}

// SupportsQuoteDelimitedString returns true for OracleDialect.
// Oracle supports Q'[...]' style string literals.
func (d *OracleDialect) SupportsQuoteDelimitedString() bool {
	return true
}

// SupportsStringEscapeConstant returns false for OracleDialect.
func (d *OracleDialect) SupportsStringEscapeConstant() bool {
	return false
}

// SupportsFilterDuringAggregation returns true for OracleDialect.
func (d *OracleDialect) SupportsFilterDuringAggregation() bool {
	return true
}

// SupportsWithinAfterArrayAggregation returns false for OracleDialect.
func (d *OracleDialect) SupportsWithinAfterArrayAggregation() bool {
	return false
}

// SupportsWindowClauseNamedWindowReference returns true for OracleDialect.
func (d *OracleDialect) SupportsWindowClauseNamedWindowReference() bool {
	return true
}

// SupportsWindowFunctionNullTreatmentArg returns true for OracleDialect.
func (d *OracleDialect) SupportsWindowFunctionNullTreatmentArg() bool {
	return true
}

// SupportsMatchRecognize returns true for OracleDialect.
func (d *OracleDialect) SupportsMatchRecognize() bool {
	return true
}

// SupportsGroupByExpr returns true for OracleDialect.
func (d *OracleDialect) SupportsGroupByExpr() bool {
	return true
}

// SupportsGroupByWithModifier returns false for OracleDialect.
func (d *OracleDialect) SupportsGroupByWithModifier() bool {
	return false
}

// SupportsLeftAssociativeJoinsWithoutParens returns false for OracleDialect.
func (d *OracleDialect) SupportsLeftAssociativeJoinsWithoutParens() bool {
	return false
}

// SupportsOuterJoinOperator returns true for OracleDialect.
// Oracle supports the (+) syntax for outer joins.
func (d *OracleDialect) SupportsOuterJoinOperator() bool {
	return true
}

// SupportsCrossJoinConstraint returns false for OracleDialect.
func (d *OracleDialect) SupportsCrossJoinConstraint() bool {
	return false
}

// SupportsConnectBy returns true for OracleDialect.
// Oracle supports CONNECT BY for hierarchical queries.
func (d *OracleDialect) SupportsConnectBy() bool {
	return true
}

// SupportsStartTransactionModifier returns false for OracleDialect.
func (d *OracleDialect) SupportsStartTransactionModifier() bool {
	return false
}

// SupportsEndTransactionModifier returns false for OracleDialect.
func (d *OracleDialect) SupportsEndTransactionModifier() bool {
	return false
}

// SupportsNamedFnArgsWithEqOperator returns false for OracleDialect.
func (d *OracleDialect) SupportsNamedFnArgsWithEqOperator() bool {
	return false
}

// SupportsNamedFnArgsWithColonOperator returns false for OracleDialect.
func (d *OracleDialect) SupportsNamedFnArgsWithColonOperator() bool {
	return false
}

// SupportsNamedFnArgsWithAssignmentOperator returns true for OracleDialect.
func (d *OracleDialect) SupportsNamedFnArgsWithAssignmentOperator() bool {
	return true
}

// SupportsNamedFnArgsWithRArrowOperator returns false for OracleDialect.
func (d *OracleDialect) SupportsNamedFnArgsWithRArrowOperator() bool {
	return false
}

// SupportsNamedFnArgsWithExprName returns false for OracleDialect.
func (d *OracleDialect) SupportsNamedFnArgsWithExprName() bool {
	return false
}

// SupportsParenthesizedSetVariables returns false for OracleDialect.
func (d *OracleDialect) SupportsParenthesizedSetVariables() bool {
	return false
}

// SupportsCommaSeparatedSetAssignments returns false for OracleDialect.
func (d *OracleDialect) SupportsCommaSeparatedSetAssignments() bool {
	return false
}

// SupportsSetStmtWithoutOperator returns true for OracleDialect.
func (d *OracleDialect) SupportsSetStmtWithoutOperator() bool {
	return true
}

// SupportsSetNames returns false for OracleDialect.
func (d *OracleDialect) SupportsSetNames() bool {
	return false
}

// SupportsSelectWildcardExcept returns false for OracleDialect.
func (d *OracleDialect) SupportsSelectWildcardExcept() bool {
	return false
}

// SupportsSelectWildcardExclude returns false for OracleDialect.
func (d *OracleDialect) SupportsSelectWildcardExclude() bool {
	return false
}

// SupportsSelectExclude returns false for OracleDialect.
func (d *OracleDialect) SupportsSelectExclude() bool {
	return false
}

// SupportsSelectWildcardReplace returns false for OracleDialect.
func (d *OracleDialect) SupportsSelectWildcardReplace() bool {
	return false
}

// SupportsSelectWildcardIlike returns false for OracleDialect.
func (d *OracleDialect) SupportsSelectWildcardIlike() bool {
	return false
}

// SupportsSelectWildcardRename returns false for OracleDialect.
func (d *OracleDialect) SupportsSelectWildcardRename() bool {
	return false
}

// SupportsSelectWildcardWithAlias returns false for OracleDialect.
func (d *OracleDialect) SupportsSelectWildcardWithAlias() bool {
	return false
}

// SupportsSelectExprStar returns false for OracleDialect.
func (d *OracleDialect) SupportsSelectExprStar() bool {
	return false
}

// SupportsFromFirstSelect returns false for OracleDialect.
func (d *OracleDialect) SupportsFromFirstSelect() bool {
	return false
}

// SupportsEmptyProjections returns false for OracleDialect.
func (d *OracleDialect) SupportsEmptyProjections() bool {
	return false
}

// SupportsSelectModifiers returns false for OracleDialect.
func (d *OracleDialect) SupportsSelectModifiers() bool {
	return false
}

// SupportsPipeOperator returns false for OracleDialect.
func (d *OracleDialect) SupportsPipeOperator() bool {
	return false
}

// SupportsTrailingCommas returns false for OracleDialect.
func (d *OracleDialect) SupportsTrailingCommas() bool {
	return false
}

// SupportsProjectionTrailingCommas returns false for OracleDialect.
func (d *OracleDialect) SupportsProjectionTrailingCommas() bool {
	return false
}

// SupportsFromTrailingCommas returns false for OracleDialect.
func (d *OracleDialect) SupportsFromTrailingCommas() bool {
	return false
}

// SupportsColumnDefinitionTrailingCommas returns false for OracleDialect.
func (d *OracleDialect) SupportsColumnDefinitionTrailingCommas() bool {
	return false
}

// SupportsLimitComma returns false for OracleDialect.
func (d *OracleDialect) SupportsLimitComma() bool {
	return false
}

// ConvertTypeBeforeValue returns false for OracleDialect.
func (d *OracleDialect) ConvertTypeBeforeValue() bool {
	return false
}

// SupportsTryConvert returns false for OracleDialect.
func (d *OracleDialect) SupportsTryConvert() bool {
	return false
}

// SupportsBinaryKwAsCast returns false for OracleDialect.
func (d *OracleDialect) SupportsBinaryKwAsCast() bool {
	return false
}

// SupportsObjectNameDoubleDotNotation returns false for OracleDialect.
func (d *OracleDialect) SupportsObjectNameDoubleDotNotation() bool {
	return false
}

// SupportsNumericPrefix returns false for OracleDialect.
func (d *OracleDialect) SupportsNumericPrefix() bool {
	return false
}

// SupportsNumericLiteralUnderscores returns false for OracleDialect.
func (d *OracleDialect) SupportsNumericLiteralUnderscores() bool {
	return false
}

// SupportsInEmptyList returns false for OracleDialect.
func (d *OracleDialect) SupportsInEmptyList() bool {
	return false
}

// SupportsDictionarySyntax returns false for OracleDialect.
func (d *OracleDialect) SupportsDictionarySyntax() bool {
	return false
}

// SupportsMapLiteralSyntax returns false for OracleDialect.
func (d *OracleDialect) SupportsMapLiteralSyntax() bool {
	return false
}

// SupportsStructLiteral returns false for OracleDialect.
func (d *OracleDialect) SupportsStructLiteral() bool {
	return false
}

// SupportsArrayLiteralSyntax returns false for OracleDialect.
func (d *OracleDialect) SupportsArrayLiteralSyntax() bool {
	return false
}

// SupportsLambdaFunctions returns false for OracleDialect.
func (d *OracleDialect) SupportsLambdaFunctions() bool {
	return false
}

// SupportsCreateTableMultiSchemaInfoSources returns false for OracleDialect.
func (d *OracleDialect) SupportsCreateTableMultiSchemaInfoSources() bool {
	return false
}

// SupportsCreateTableLikeParenthesized returns false for OracleDialect.
func (d *OracleDialect) SupportsCreateTableLikeParenthesized() bool {
	return false
}

// SupportsCreateTableSelect returns true for OracleDialect.
func (d *OracleDialect) SupportsCreateTableSelect() bool {
	return true
}

// SupportsCreateViewCommentSyntax returns false for OracleDialect.
func (d *OracleDialect) SupportsCreateViewCommentSyntax() bool {
	return false
}

// SupportsArrayTypedefWithoutElementType returns false for OracleDialect.
func (d *OracleDialect) SupportsArrayTypedefWithoutElementType() bool {
	return false
}

// SupportsArrayTypedefWithBrackets returns false for OracleDialect.
func (d *OracleDialect) SupportsArrayTypedefWithBrackets() bool {
	return false
}

// SupportsParensAroundTableFactor returns false for OracleDialect.
func (d *OracleDialect) SupportsParensAroundTableFactor() bool {
	return false
}

// SupportsValuesAsTableFactor returns false for OracleDialect.
func (d *OracleDialect) SupportsValuesAsTableFactor() bool {
	return false
}

// SupportsUnnestTableFactor returns false for OracleDialect.
func (d *OracleDialect) SupportsUnnestTableFactor() bool {
	return false
}

// SupportsSemanticViewTableFactor returns false for OracleDialect.
func (d *OracleDialect) SupportsSemanticViewTableFactor() bool {
	return false
}

// SupportsTableVersioning returns false for OracleDialect.
func (d *OracleDialect) SupportsTableVersioning() bool {
	return false
}

// SupportsTableSampleBeforeAlias returns false for OracleDialect.
func (d *OracleDialect) SupportsTableSampleBeforeAlias() bool {
	return false
}

// SupportsTableHints returns false for OracleDialect.
func (d *OracleDialect) SupportsTableHints() bool {
	return false
}

// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
func (d *OracleDialect) SupportsIndexHints() bool {
	return false
}

// SupportsAscDescInColumnDefinition returns false for OracleDialect.
func (d *OracleDialect) SupportsAscDescInColumnDefinition() bool {
	return false
}

// SupportsSpaceSeparatedColumnOptions returns false for OracleDialect.
func (d *OracleDialect) SupportsSpaceSeparatedColumnOptions() bool {
	return false
}

// SupportsConstraintKeywordWithoutName returns false for OracleDialect.
func (d *OracleDialect) SupportsConstraintKeywordWithoutName() bool {
	return false
}

// SupportsKeyColumnOption returns false for OracleDialect.
func (d *OracleDialect) SupportsKeyColumnOption() bool {
	return false
}

// SupportsDataTypeSignedSuffix returns false for OracleDialect.
func (d *OracleDialect) SupportsDataTypeSignedSuffix() bool {
	return false
}

// SupportsNestedComments returns false for OracleDialect.
func (d *OracleDialect) SupportsNestedComments() bool {
	return false
}

// SupportsMultilineCommentHints returns false for OracleDialect.
func (d *OracleDialect) SupportsMultilineCommentHints() bool {
	return false
}

// SupportsCommentOptimizerHint returns true for OracleDialect.
// Oracle supports query optimizer hints in comments.
func (d *OracleDialect) SupportsCommentOptimizerHint() bool {
	return true
}

// SupportsCommentOn returns true for OracleDialect.
func (d *OracleDialect) SupportsCommentOn() bool {
	return true
}

// RequiresSingleLineCommentWhitespace returns false for OracleDialect.
func (d *OracleDialect) RequiresSingleLineCommentWhitespace() bool {
	return false
}

// SupportsExplainWithUtilityOptions returns true for OracleDialect.
func (d *OracleDialect) SupportsExplainWithUtilityOptions() bool {
	return true
}

// SupportsExecuteImmediate returns true for OracleDialect.
// Oracle supports EXECUTE IMMEDIATE statements.
func (d *OracleDialect) SupportsExecuteImmediate() bool {
	return true
}

// AllowExtractCustom returns true for OracleDialect.
func (d *OracleDialect) AllowExtractCustom() bool {
	return true
}

// AllowExtractSingleQuotes returns false for OracleDialect.
func (d *OracleDialect) AllowExtractSingleQuotes() bool {
	return false
}

// SupportsExtractCommaSyntax returns false for OracleDialect.
func (d *OracleDialect) SupportsExtractCommaSyntax() bool {
	return false
}

// SupportsSubqueryAsFunctionArg returns false for OracleDialect.
func (d *OracleDialect) SupportsSubqueryAsFunctionArg() bool {
	return false
}

// SupportsDollarPlaceholder returns false for OracleDialect.
func (d *OracleDialect) SupportsDollarPlaceholder() bool {
	return false
}

// SupportsCreateIndexWithClause returns true for OracleDialect.
func (d *OracleDialect) SupportsCreateIndexWithClause() bool {
	return true
}

// RequireIntervalQualifier returns false for OracleDialect.
func (d *OracleDialect) RequireIntervalQualifier() bool {
	return false
}

// SupportsIntervalOptions returns false for OracleDialect.
func (d *OracleDialect) SupportsIntervalOptions() bool {
	return false
}

// SupportsFactorialOperator returns false for OracleDialect.
func (d *OracleDialect) SupportsFactorialOperator() bool {
	return false
}

// SupportsBitwiseShiftOperators returns false for OracleDialect.
func (d *OracleDialect) SupportsBitwiseShiftOperators() bool {
	return false
}

// SupportsNotnullOperator returns false for OracleDialect.
func (d *OracleDialect) SupportsNotnullOperator() bool {
	return false
}

// SupportsBangNotOperator returns false for OracleDialect.
func (d *OracleDialect) SupportsBangNotOperator() bool {
	return false
}

// SupportsDoubleAmpersandOperator returns true for OracleDialect.
// Oracle uses && for boolean AND in some contexts.
func (d *OracleDialect) SupportsDoubleAmpersandOperator() bool {
	return true
}

// SupportsMatchAgainst returns false for OracleDialect.
func (d *OracleDialect) SupportsMatchAgainst() bool {
	return false
}

// SupportsUserHostGrantee returns false for OracleDialect.
func (d *OracleDialect) SupportsUserHostGrantee() bool {
	return false
}

// SupportsListenNotify returns false for OracleDialect.
func (d *OracleDialect) SupportsListenNotify() bool {
	return false
}

// SupportsLoadData returns false for OracleDialect.
func (d *OracleDialect) SupportsLoadData() bool {
	return false
}

// SupportsLoadExtension returns false for OracleDialect.
func (d *OracleDialect) SupportsLoadExtension() bool {
	return false
}

// SupportsTopBeforeDistinct returns false for OracleDialect.
func (d *OracleDialect) SupportsTopBeforeDistinct() bool {
	return false
}

// SupportsBooleanLiterals returns false for OracleDialect.
// Oracle doesn't support TRUE/FALSE literals (uses 1/0).
func (d *OracleDialect) SupportsBooleanLiterals() bool {
	return false
}

// SupportsShowLikeBeforeIn returns false for OracleDialect.
func (d *OracleDialect) SupportsShowLikeBeforeIn() bool {
	return false
}

// SupportsPartiQL returns false for OracleDialect.
func (d *OracleDialect) SupportsPartiQL() bool {
	return false
}

// SupportsEqAliasAssignment returns false for OracleDialect.
func (d *OracleDialect) SupportsEqAliasAssignment() bool {
	return false
}

// SupportsInsertSet returns false for OracleDialect.
func (d *OracleDialect) SupportsInsertSet() bool {
	return false
}

// SupportsInsertTableFunction returns false for OracleDialect.
func (d *OracleDialect) SupportsInsertTableFunction() bool {
	return false
}

// SupportsInsertTableQuery returns true for OracleDialect.
// Oracle supports: INSERT INTO ... SELECT INTO (<query>) ...
func (d *OracleDialect) SupportsInsertTableQuery() bool {
	return true
}

// SupportsInsertFormat returns false for OracleDialect.
func (d *OracleDialect) SupportsInsertFormat() bool {
	return false
}

// SupportsInsertTableAlias returns true for OracleDialect.
func (d *OracleDialect) SupportsInsertTableAlias() bool {
	return true
}

// SupportsAlterColumnTypeUsing returns false for OracleDialect.
func (d *OracleDialect) SupportsAlterColumnTypeUsing() bool {
	return false
}

// SupportsCommaSeparatedDropColumnList returns false for OracleDialect.
func (d *OracleDialect) SupportsCommaSeparatedDropColumnList() bool {
	return false
}

// SupportsOrderByAll returns false for OracleDialect.
func (d *OracleDialect) SupportsOrderByAll() bool {
	return false
}

// SupportsGeometricTypes returns false for OracleDialect.
func (d *OracleDialect) SupportsGeometricTypes() bool {
	return false
}

// DescribeRequiresTableKeyword returns false for OracleDialect.
func (d *OracleDialect) DescribeRequiresTableKeyword() bool {
	return false
}

// SupportsOptimizeTable returns false for OracleDialect.
func (d *OracleDialect) SupportsOptimizeTable() bool {
	return false
}

// SupportsPrewhere returns false for OracleDialect.
func (d *OracleDialect) SupportsPrewhere() bool {
	return false
}

// SupportsWithFill returns false for OracleDialect.
func (d *OracleDialect) SupportsWithFill() bool {
	return false
}

// SupportsLimitBy returns false for OracleDialect.
func (d *OracleDialect) SupportsLimitBy() bool {
	return false
}

// SupportsInterpolate returns false for OracleDialect.
func (d *OracleDialect) SupportsInterpolate() bool {
	return false
}

// SupportsSettings returns false for OracleDialect.
func (d *OracleDialect) SupportsSettings() bool {
	return false
}

// SupportsSelectFormat returns false for OracleDialect.
func (d *OracleDialect) SupportsSelectFormat() bool {
	return false
}

// SupportsInstall returns false for OracleDialect.
func (d *OracleDialect) SupportsInstall() bool {
	return false
}

// SupportsDetach returns false for OracleDialect.
func (d *OracleDialect) SupportsDetach() bool {
	return false
}

// SupportsCommaSeparatedTrim returns false for OracleDialect.
func (d *OracleDialect) SupportsCommaSeparatedTrim() bool {
	return false
}

// Precedence constants for Oracle operators.
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
func (d *OracleDialect) PrecValue(prec dialects.Precedence) uint8 {
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
func (d *OracleDialect) PrecUnknown() uint8 {
	return 0
}

// GetNextPrecedence returns the dialect-specific precedence override for
// the next token. Oracle handles string concatenation (||) specially.
func (d *OracleDialect) GetNextPrecedence(parser dialects.ParserAccessor) (uint8, error) {
	tok := parser.PeekTokenRef()

	switch tok.Token.(type) {
	case token.TokenStringConcat:
		// String concatenation operator || has same precedence as PlusMinus
		return plusMinusPrec, nil
	}

	return 0, nil
}

// GetNextPrecedenceDefault implements the default precedence logic.
func (d *OracleDialect) GetNextPrecedenceDefault(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// ParsePrefix returns false to fall back to default behavior.
func (d *OracleDialect) ParsePrefix(parser dialects.ParserAccessor) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseInfix returns false to fall back to default behavior.
func (d *OracleDialect) ParseInfix(parser dialects.ParserAccessor, expr ast.Expr, precedence uint8) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseStatement returns false to fall back to default behavior.
func (d *OracleDialect) ParseStatement(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	return nil, false, nil
}

// ParseColumnOption returns false to fall back to default behavior.
func (d *OracleDialect) ParseColumnOption(parser dialects.ParserAccessor) (dialects.ColumnOption, bool, error) {
	return nil, false, nil
}
