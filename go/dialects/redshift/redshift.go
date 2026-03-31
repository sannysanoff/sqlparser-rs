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

// Package redshift provides an Amazon Redshift dialect implementation for the SQL parser.
//
// This dialect implements Redshift-specific parsing behavior including:
//   - Based on PostgreSQL with Redshift-specific extensions
//   - Bracket-quoted identifiers ([foo] and ["foo"])
//   - CONVERT(type, value) syntax (type before value)
//   - SUPER type for semi-structured data (PartiQL support)
//   - COPY and UNLOAD commands
//   - Supports CONNECT BY for hierarchical queries
//   - TOP before DISTINCT in SELECT
//   - Supports EXCLUDE/REPLACE in wildcards
//
// See https://docs.aws.amazon.com/redshift/
package redshift

import (
	"unicode"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/token"
	"github.com/user/sqlparser/tokenizer"
)

// RedshiftSqlDialect is a dialect for Amazon Redshift.
type RedshiftSqlDialect struct{}

// NewRedshiftSqlDialect creates a new instance of RedshiftSqlDialect.
func NewRedshiftSqlDialect() *RedshiftSqlDialect {
	return &RedshiftSqlDialect{}
}

// Dialect returns the dialect identifier.
func (d *RedshiftSqlDialect) Dialect() string {
	return "redshift"
}

// IsIdentifierStart returns true if the character is a valid start character
// for an unquoted identifier. Redshift supports letters, underscore, hash,
// and Unicode characters (via PostgreSQL-style).
func (d *RedshiftSqlDialect) IsIdentifierStart(ch rune) bool {
	// Letters and underscore
	if unicode.IsLetter(ch) || ch == '_' {
		return true
	}
	// Redshift supports # in identifiers
	if ch == '#' {
		return true
	}
	// UTF-8 multibyte characters are supported
	return ch > 127
}

// IsIdentifierPart returns true if the character is a valid unquoted
// identifier character. Redshift supports letters, digits, underscore, hash,
// and Unicode characters.
func (d *RedshiftSqlDialect) IsIdentifierPart(ch rune) bool {
	// Letters, digits, underscore
	if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
		return true
	}
	// Redshift supports # in identifiers
	if ch == '#' {
		return true
	}
	// UTF-8 multibyte characters are supported
	return ch > 127
}

// IsDelimitedIdentifierStart returns true if the character starts a quoted identifier.
// Redshift uses double quotes for delimited identifiers.
func (d *RedshiftSqlDialect) IsDelimitedIdentifierStart(ch rune) bool {
	return ch == '"'
}

// IsNestedDelimitedIdentifierStart returns true if the character starts a
// potential nested quoted identifier. Redshift supports brackets as nested delimiters.
func (d *RedshiftSqlDialect) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return ch == '['
}

// PeekNestedDelimitedIdentifierQuotes returns the outer quote style and
// optional inner quote style if applicable. Redshift supports ["foo"] and [foo] patterns.
func (d *RedshiftSqlDialect) PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*dialects.NestedIdentifierQuote, int) {
	if pos >= len(chars) || chars[pos] != '[' {
		return nil, pos
	}

	// Move past the '['
	pos++

	// Skip whitespace
	for pos < len(chars) && unicode.IsSpace(chars[pos]) {
		pos++
	}

	if pos >= len(chars) {
		return nil, pos
	}

	ch := chars[pos]

	// Check if it's ["foo"] pattern
	if ch == '"' {
		quote := rune('"')
		return &dialects.NestedIdentifierQuote{
			OuterQuote: '[',
			InnerQuote: &quote,
		}, pos
	}

	// Check if it's a valid identifier start after [ (e.g., [foo])
	if d.IsIdentifierStart(ch) {
		return &dialects.NestedIdentifierQuote{
			OuterQuote: '[',
			InnerQuote: nil,
		}, pos
	}

	return nil, pos
}

// IdentifierQuoteStyle returns the character used to quote identifiers.
// Redshift uses double quotes for identifier quoting.
func (d *RedshiftSqlDialect) IdentifierQuoteStyle(identifier string) *rune {
	quote := rune('"')
	return &quote
}

// IsCustomOperatorPart returns true if the character is part of a custom operator.
func (d *RedshiftSqlDialect) IsCustomOperatorPart(ch rune) bool {
	return false
}

// IsReservedForIdentifier returns true if the keyword is reserved.
func (d *RedshiftSqlDialect) IsReservedForIdentifier(kw token.Keyword) bool {
	return token.IsReservedForIdentifier(kw)
}

// GetReservedKeywordsForSelectItemOperator returns reserved keywords for select items.
func (d *RedshiftSqlDialect) GetReservedKeywordsForSelectItemOperator() []token.Keyword {
	return nil
}

// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
func (d *RedshiftSqlDialect) GetReservedGranteesTypes() []dialects.GranteesType {
	return nil
}

// IsColumnAlias returns true if the keyword should be parsed as a column alias.
func (d *RedshiftSqlDialect) IsColumnAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == token.AS || token.IsReservedForColumnAlias(kw)
}

// IsSelectItemAlias returns true if the specified keyword should be parsed as a select item alias.
func (d *RedshiftSqlDialect) IsSelectItemAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsColumnAlias(kw, parser)
}

// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier.
func (d *RedshiftSqlDialect) IsTableFactor(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == token.TABLE || kw == token.LATERAL || kw == token.PIVOT || kw == token.UNPIVOT
}

// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias.
func (d *RedshiftSqlDialect) IsTableAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsTableFactor(kw, parser)
}

// IsTableFactorAlias returns true if the specified keyword should be parsed as a table factor alias.
func (d *RedshiftSqlDialect) IsTableFactorAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return explicit || d.IsTableAlias(kw, parser)
}

// IsIdentifierGeneratingFunctionName returns true if the ident is a function that generates identifiers.
func (d *RedshiftSqlDialect) IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool {
	return false
}

// SupportsStringLiteralBackslashEscape returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsStringLiteralBackslashEscape() bool {
	return true
}

// IgnoresWildcardEscapes returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) IgnoresWildcardEscapes() bool {
	return false
}

// SupportsUnicodeStringLiteral returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsUnicodeStringLiteral() bool {
	return false
}

// SupportsTripleQuotedString returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsTripleQuotedString() bool {
	return false
}

// SupportsStringLiteralConcatenation returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsStringLiteralConcatenation() bool {
	return false
}

// SupportsStringLiteralConcatenationWithNewline returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsStringLiteralConcatenationWithNewline() bool {
	return true
}

// SupportsQuoteDelimitedString returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsQuoteDelimitedString() bool {
	return false
}

// SupportsStringEscapeConstant returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsStringEscapeConstant() bool {
	return true
}

// SupportsFilterDuringAggregation returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsFilterDuringAggregation() bool {
	return true
}

// SupportsWithinAfterArrayAggregation returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsWithinAfterArrayAggregation() bool {
	return false
}

// SupportsWindowClauseNamedWindowReference returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsWindowClauseNamedWindowReference() bool {
	return true
}

// SupportsWindowFunctionNullTreatmentArg returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsWindowFunctionNullTreatmentArg() bool {
	return true
}

// SupportsMatchRecognize returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsMatchRecognize() bool {
	return false
}

// SupportsGroupByExpr returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsGroupByExpr() bool {
	return true
}

// SupportsGroupByWithModifier returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsGroupByWithModifier() bool {
	return false
}

// SupportsLeftAssociativeJoinsWithoutParens returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsLeftAssociativeJoinsWithoutParens() bool {
	return false
}

// SupportsOuterJoinOperator returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsOuterJoinOperator() bool {
	return false
}

// SupportsCrossJoinConstraint returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsCrossJoinConstraint() bool {
	return false
}

// SupportsConnectBy returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsConnectBy() bool {
	return true
}

// SupportsStartTransactionModifier returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsStartTransactionModifier() bool {
	return false
}

// SupportsEndTransactionModifier returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsEndTransactionModifier() bool {
	return false
}

// SupportsNamedFnArgsWithEqOperator returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsNamedFnArgsWithEqOperator() bool {
	return false
}

// SupportsNamedFnArgsWithColonOperator returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsNamedFnArgsWithColonOperator() bool {
	return false
}

// SupportsNamedFnArgsWithAssignmentOperator returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsNamedFnArgsWithAssignmentOperator() bool {
	return false
}

// SupportsNamedFnArgsWithRArrowOperator returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsNamedFnArgsWithRArrowOperator() bool {
	return true
}

// SupportsNamedFnArgsWithExprName returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsNamedFnArgsWithExprName() bool {
	return false
}

// SupportsParenthesizedSetVariables returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsParenthesizedSetVariables() bool {
	return false
}

// SupportsCommaSeparatedSetAssignments returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsCommaSeparatedSetAssignments() bool {
	return false
}

// SupportsSetStmtWithoutOperator returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSetStmtWithoutOperator() bool {
	return false
}

// SupportsSetNames returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSetNames() bool {
	return true
}

// SupportsSelectWildcardExcept returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSelectWildcardExcept() bool {
	return false
}

// SupportsSelectWildcardExclude returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSelectWildcardExclude() bool {
	return true
}

// SupportsSelectExclude returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSelectExclude() bool {
	return true
}

// SupportsSelectWildcardReplace returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSelectWildcardReplace() bool {
	return false
}

// SupportsSelectWildcardIlike returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSelectWildcardIlike() bool {
	return false
}

// SupportsSelectWildcardRename returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSelectWildcardRename() bool {
	return false
}

// SupportsSelectWildcardWithAlias returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSelectWildcardWithAlias() bool {
	return true
}

// SupportsSelectExprStar returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSelectExprStar() bool {
	return false
}

// SupportsFromFirstSelect returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsFromFirstSelect() bool {
	return false
}

// SupportsEmptyProjections returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsEmptyProjections() bool {
	return false
}

// SupportsSelectModifiers returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSelectModifiers() bool {
	return false
}

// SupportsPipeOperator returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsPipeOperator() bool {
	return false
}

// SupportsTrailingCommas returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsTrailingCommas() bool {
	return false
}

// SupportsProjectionTrailingCommas returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsProjectionTrailingCommas() bool {
	return false
}

// SupportsFromTrailingCommas returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsFromTrailingCommas() bool {
	return false
}

// SupportsColumnDefinitionTrailingCommas returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsColumnDefinitionTrailingCommas() bool {
	return false
}

// SupportsLimitComma returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsLimitComma() bool {
	return false
}

// ConvertTypeBeforeValue returns true for RedshiftSqlDialect.
// Redshift has CONVERT(type, value) instead of CONVERT(value, type).
func (d *RedshiftSqlDialect) ConvertTypeBeforeValue() bool {
	return true
}

// SupportsTryConvert returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsTryConvert() bool {
	return false
}

// SupportsBinaryKwAsCast returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsBinaryKwAsCast() bool {
	return false
}

// SupportsObjectNameDoubleDotNotation returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsObjectNameDoubleDotNotation() bool {
	return false
}

// SupportsNumericPrefix returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsNumericPrefix() bool {
	return false
}

// SupportsNumericLiteralUnderscores returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsNumericLiteralUnderscores() bool {
	return true
}

// SupportsInEmptyList returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsInEmptyList() bool {
	return false
}

// SupportsDictionarySyntax returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsDictionarySyntax() bool {
	return false
}

// SupportsMapLiteralSyntax returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsMapLiteralSyntax() bool {
	return false
}

// SupportsStructLiteral returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsStructLiteral() bool {
	return false
}

// SupportsArrayLiteralSyntax returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsArrayLiteralSyntax() bool {
	return true
}

// SupportsLambdaFunctions returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsLambdaFunctions() bool {
	return false
}

// SupportsCreateTableMultiSchemaInfoSources returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsCreateTableMultiSchemaInfoSources() bool {
	return false
}

// SupportsCreateTableLikeParenthesized returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsCreateTableLikeParenthesized() bool {
	return true
}

// SupportsCreateTableSelect returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsCreateTableSelect() bool {
	return false
}

// SupportsCreateViewCommentSyntax returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsCreateViewCommentSyntax() bool {
	return false
}

// SupportsArrayTypedefWithoutElementType returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsArrayTypedefWithoutElementType() bool {
	return false
}

// SupportsArrayTypedefWithBrackets returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsArrayTypedefWithBrackets() bool {
	return true
}

// SupportsParensAroundTableFactor returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsParensAroundTableFactor() bool {
	return false
}

// SupportsValuesAsTableFactor returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsValuesAsTableFactor() bool {
	return false
}

// SupportsSemanticViewTableFactor returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSemanticViewTableFactor() bool {
	return false
}

// SupportsTableVersioning returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsTableVersioning() bool {
	return false
}

// SupportsTableSampleBeforeAlias returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsTableSampleBeforeAlias() bool {
	return false
}

// SupportsTableHints returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsTableHints() bool {
	return false
}

// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
func (d *RedshiftSqlDialect) SupportsIndexHints() bool {
	return false
}

// SupportsAscDescInColumnDefinition returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsAscDescInColumnDefinition() bool {
	return false
}

// SupportsSpaceSeparatedColumnOptions returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSpaceSeparatedColumnOptions() bool {
	return false
}

// SupportsConstraintKeywordWithoutName returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsConstraintKeywordWithoutName() bool {
	return false
}

// SupportsKeyColumnOption returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsKeyColumnOption() bool {
	return false
}

// SupportsDataTypeSignedSuffix returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsDataTypeSignedSuffix() bool {
	return false
}

// SupportsNestedComments returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsNestedComments() bool {
	return true
}

// SupportsMultilineCommentHints returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsMultilineCommentHints() bool {
	return false
}

// SupportsCommentOptimizerHint returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsCommentOptimizerHint() bool {
	return false
}

// SupportsCommentOn returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsCommentOn() bool {
	return true
}

// RequiresSingleLineCommentWhitespace returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) RequiresSingleLineCommentWhitespace() bool {
	return false
}

// SupportsExplainWithUtilityOptions returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsExplainWithUtilityOptions() bool {
	return true
}

// SupportsExecuteImmediate returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsExecuteImmediate() bool {
	return false
}

// AllowExtractCustom returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) AllowExtractCustom() bool {
	return true
}

// AllowExtractSingleQuotes returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) AllowExtractSingleQuotes() bool {
	return true
}

// SupportsExtractCommaSyntax returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsExtractCommaSyntax() bool {
	return false
}

// SupportsSubqueryAsFunctionArg returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSubqueryAsFunctionArg() bool {
	return false
}

// SupportsDollarPlaceholder returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsDollarPlaceholder() bool {
	return true
}

// SupportsCreateIndexWithClause returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsCreateIndexWithClause() bool {
	return true
}

// RequireIntervalQualifier returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) RequireIntervalQualifier() bool {
	return false
}

// SupportsIntervalOptions returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsIntervalOptions() bool {
	return true
}

// SupportsFactorialOperator returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsFactorialOperator() bool {
	return true
}

// SupportsBitwiseShiftOperators returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsBitwiseShiftOperators() bool {
	return true
}

// SupportsNotnullOperator returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsNotnullOperator() bool {
	return true
}

// SupportsBangNotOperator returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsBangNotOperator() bool {
	return false
}

// SupportsDoubleAmpersandOperator returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsDoubleAmpersandOperator() bool {
	return false
}

// SupportsMatchAgainst returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsMatchAgainst() bool {
	return false
}

// SupportsUserHostGrantee returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsUserHostGrantee() bool {
	return false
}

// SupportsListenNotify returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsListenNotify() bool {
	return false
}

// SupportsLoadData returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsLoadData() bool {
	return false
}

// SupportsLoadExtension returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsLoadExtension() bool {
	return false
}

// SupportsTopBeforeDistinct returns true for RedshiftSqlDialect.
// Redshift expects the TOP option before the ALL/DISTINCT options.
func (d *RedshiftSqlDialect) SupportsTopBeforeDistinct() bool {
	return true
}

// SupportsBooleanLiterals returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsBooleanLiterals() bool {
	return true
}

// SupportsShowLikeBeforeIn returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsShowLikeBeforeIn() bool {
	return false
}

// SupportsPartiQL returns true for RedshiftSqlDialect.
// Redshift supports PartiQL for querying semi-structured data (SUPER type).
func (d *RedshiftSqlDialect) SupportsPartiQL() bool {
	return true
}

// SupportsEqAliasAssignment returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsEqAliasAssignment() bool {
	return false
}

// SupportsInsertSet returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsInsertSet() bool {
	return false
}

// SupportsInsertTableFunction returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsInsertTableFunction() bool {
	return false
}

// SupportsInsertTableQuery returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsInsertTableQuery() bool {
	return false
}

// SupportsInsertFormat returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsInsertFormat() bool {
	return false
}

// SupportsInsertTableAlias returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsInsertTableAlias() bool {
	return true
}

// SupportsAlterColumnTypeUsing returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsAlterColumnTypeUsing() bool {
	return true
}

// SupportsCommaSeparatedDropColumnList returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsCommaSeparatedDropColumnList() bool {
	return false
}

// SupportsOrderByAll returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsOrderByAll() bool {
	return false
}

// SupportsGeometricTypes returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsGeometricTypes() bool {
	return true
}

// DescribeRequiresTableKeyword returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) DescribeRequiresTableKeyword() bool {
	return false
}

// SupportsOptimizeTable returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsOptimizeTable() bool {
	return false
}

// SupportsPrewhere returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsPrewhere() bool {
	return false
}

// SupportsWithFill returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsWithFill() bool {
	return false
}

// SupportsLimitBy returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsLimitBy() bool {
	return false
}

// SupportsInterpolate returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsInterpolate() bool {
	return false
}

// SupportsSettings returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSettings() bool {
	return false
}

// SupportsSelectFormat returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsSelectFormat() bool {
	return false
}

// SupportsInstall returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsInstall() bool {
	return false
}

// SupportsDetach returns false for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsDetach() bool {
	return false
}

// SupportsCommaSeparatedTrim returns true for RedshiftSqlDialect.
func (d *RedshiftSqlDialect) SupportsCommaSeparatedTrim() bool {
	return true
}

// Precedence constants for Redshift operators.
const (
	periodPrec      uint8 = 200
	doubleColonPrec uint8 = 140
	bracketPrec     uint8 = 130
	collatePrec     uint8 = 120
	atTzPrec        uint8 = 110
	caretPrec       uint8 = 100
	mulDivModPrec   uint8 = 90
	plusMinusPrec   uint8 = 80
	xorPrec         uint8 = 75
	pgOtherPrec     uint8 = 70
	betweenLikePrec uint8 = 60
	eqPrec          uint8 = 50
	isPrec          uint8 = 40
	notPrec         uint8 = 30
	andPrec         uint8 = 20
	orPrec          uint8 = 10
)

// PrecValue returns the precedence value for a given Precedence level.
func (d *RedshiftSqlDialect) PrecValue(prec dialects.Precedence) uint8 {
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
		return pgOtherPrec
	case prec == dialects.PrecedenceCaret:
		return caretPrec
	case prec == dialects.PrecedencePipe || prec == dialects.PrecedenceColon:
		return pgOtherPrec
	case prec == dialects.PrecedenceBetween:
		return betweenLikePrec
	case prec == dialects.PrecedenceEq:
		return eqPrec
	case prec == dialects.PrecedenceLike:
		return betweenLikePrec
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
func (d *RedshiftSqlDialect) PrecUnknown() uint8 {
	return 0
}

// GetNextPrecedence returns the dialect-specific precedence override for
// the next token. Returns 0 to fall back to default behavior.
func (d *RedshiftSqlDialect) GetNextPrecedence(parser dialects.ParserAccessor) (uint8, error) {
	tok := parser.PeekTokenRef()

	switch tok.Token.(type) {
	case tokenizer.TokenWord:
		if word, ok := tok.Token.(tokenizer.TokenWord); ok {
			if word.Keyword == token.COLLATE && !parser.InColumnDefinitionState() {
				return collatePrec, nil
			}
		}
	case tokenizer.TokenLBracket:
		return bracketPrec, nil
	case tokenizer.TokenArrow, tokenizer.TokenLongArrow, tokenizer.TokenHashArrow, tokenizer.TokenHashLongArrow,
		tokenizer.TokenAtArrow, tokenizer.TokenArrowAt, tokenizer.TokenHashMinus, tokenizer.TokenAtQuestion,
		tokenizer.TokenAtAt, tokenizer.TokenQuestion, tokenizer.TokenQuestionAnd, tokenizer.TokenQuestionPipe,
		tokenizer.TokenExclamationMark, tokenizer.TokenOverlap, tokenizer.TokenCaretAt, tokenizer.TokenStringConcat,
		tokenizer.TokenSharp, tokenizer.TokenShiftRight, tokenizer.TokenShiftLeft, tokenizer.TokenCustomBinaryOperator:
		return pgOtherPrec, nil
	case tokenizer.TokenColon:
		return d.PrecUnknown(), nil
	}

	return 0, nil
}

// GetNextPrecedenceDefault implements the default precedence logic.
func (d *RedshiftSqlDialect) GetNextPrecedenceDefault(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// ParsePrefix is a dialect-specific prefix parser override.
func (d *RedshiftSqlDialect) ParsePrefix(parser dialects.ParserAccessor) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseInfix is a dialect-specific infix parser override.
func (d *RedshiftSqlDialect) ParseInfix(parser dialects.ParserAccessor, expr ast.Expr, precedence uint8) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseStatement is a dialect-specific statement parser override.
func (d *RedshiftSqlDialect) ParseStatement(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	return nil, false, nil
}

// ParseColumnOption is a dialect-specific column option parser override.
func (d *RedshiftSqlDialect) ParseColumnOption(parser dialects.ParserAccessor) (dialects.ColumnOption, bool, error) {
	return nil, false, nil
}
