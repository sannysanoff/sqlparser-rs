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

// Package postgresql provides a PostgreSQL dialect implementation for the SQL parser.
//
// This dialect implements PostgreSQL-specific parsing behavior including:
//   - Double-quoted identifiers (no backtick support)
//   - Unicode identifier support
//   - String escape constants (E'...')
//   - Dollar-quoted strings ($$...$$)
//   - Custom operators
//   - Array types and operations
//   - Geometric types (POINT, LINE, BOX, etc.)
//   - JSON/JSONB operators
//   - LISTEN/NOTIFY statements
//   - COPY statements
//   - PostgreSQL-specific precedence rules
package postgresql

import (
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/token"
)

// PostgreSqlDialect is a dialect for PostgreSQL.
// See https://www.postgresql.org/
type PostgreSqlDialect struct{}

// NewPostgreSqlDialect creates a new instance of PostgreSqlDialect.
func NewPostgreSqlDialect() *PostgreSqlDialect {
	return &PostgreSqlDialect{}
}

// Dialect returns the dialect identifier.
func (d *PostgreSqlDialect) Dialect() string {
	return "postgresql"
}

// IsIdentifierStart returns true if the character is a valid start character
// for an unquoted identifier. PostgreSQL supports letters, underscore, and
// Unicode characters.
func (d *PostgreSqlDialect) IsIdentifierStart(ch rune) bool {
	// Letters and underscore are valid start characters
	if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' {
		return true
	}
	// PostgreSQL implements Unicode characters in identifiers
	return ch > 127
}

// IsIdentifierPart returns true if the character is a valid unquoted
// identifier character. PostgreSQL supports letters, digits, underscore,
// dollar sign, and Unicode characters.
func (d *PostgreSqlDialect) IsIdentifierPart(ch rune) bool {
	// Letters, digits, underscore, and dollar sign
	if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') || ch == '_' || ch == '$' {
		return true
	}
	// PostgreSQL implements Unicode characters in identifiers
	return ch > 127
}

// IsDelimitedIdentifierStart returns true if the character starts a quoted identifier.
// PostgreSQL uses double quotes for delimited identifiers (no backticks).
func (d *PostgreSqlDialect) IsDelimitedIdentifierStart(ch rune) bool {
	return ch == '"'
}

// IsNestedDelimitedIdentifierStart returns true if the character starts a
// potential nested quoted identifier. PostgreSQL doesn't support nested identifiers.
func (d *PostgreSqlDialect) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return false
}

// PeekNestedDelimitedIdentifierQuotes returns the outer quote style and
// optional inner quote style if applicable. PostgreSQL doesn't support nested identifiers.
func (d *PostgreSqlDialect) PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*dialects.NestedIdentifierQuote, int) {
	return nil, pos
}

// IdentifierQuoteStyle returns the character used to quote identifiers.
// PostgreSQL always uses double quotes for identifier quoting.
func (d *PostgreSqlDialect) IdentifierQuoteStyle(identifier string) *rune {
	quote := rune('"')
	return &quote
}

// IsCustomOperatorPart returns true if the character is part of a custom operator.
// See https://www.postgresql.org/docs/current/sql-createoperator.html
func (d *PostgreSqlDialect) IsCustomOperatorPart(ch rune) bool {
	switch ch {
	case '+', '-', '*', '/', '<', '>', '=', '~', '!', '@', '#', '%', '^', '&', '|', '`', '?':
		return true
	default:
		return false
	}
}

// IsReservedForIdentifier returns true if the keyword is reserved.
// PostgreSQL treats INTERVAL as not reserved (can be used as identifier).
func (d *PostgreSqlDialect) IsReservedForIdentifier(kw token.Keyword) bool {
	// INTERVAL is not reserved in PostgreSQL
	if kw == token.INTERVAL {
		return false
	}
	return token.IsReservedForIdentifier(kw)
}

// GetReservedKeywordsForSelectItemOperator returns reserved keywords for select items.
func (d *PostgreSqlDialect) GetReservedKeywordsForSelectItemOperator() []token.Keyword {
	return nil
}

// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
func (d *PostgreSqlDialect) GetReservedGranteesTypes() []dialects.GranteesType {
	return nil
}

// IsColumnAlias returns true if the keyword should be parsed as a column alias.
func (d *PostgreSqlDialect) IsColumnAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == token.AS || token.IsReservedForColumnAlias(kw)
}

// IsSelectItemAlias returns true if the specified keyword should be parsed as a select item alias
func (d *PostgreSqlDialect) IsSelectItemAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsColumnAlias(kw, parser)
}

// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier
func (d *PostgreSqlDialect) IsTableFactor(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return kw == token.TABLE || kw == token.LATERAL || kw == token.PIVOT || kw == token.UNPIVOT
}

// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias
func (d *PostgreSqlDialect) IsTableAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsTableFactor(kw, parser)
}

// IsTableFactorAlias returns true if the specified keyword should be parsed as a table factor alias
func (d *PostgreSqlDialect) IsTableFactorAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return explicit || d.IsTableAlias(kw, parser)
}

// IsIdentifierGeneratingFunctionName returns true if the ident is a function that generates identifiers.
func (d *PostgreSqlDialect) IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool {
	return false
}

// SupportsStringLiteralBackslashEscape returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsStringLiteralBackslashEscape() bool {
	return true
}

// IgnoresWildcardEscapes returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) IgnoresWildcardEscapes() bool {
	return false
}

// SupportsUnicodeStringLiteral returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsUnicodeStringLiteral() bool {
	return true
}

// SupportsTripleQuotedString returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsTripleQuotedString() bool {
	return false
}

// SupportsStringLiteralConcatenation returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsStringLiteralConcatenation() bool {
	return false
}

// SupportsStringLiteralConcatenationWithNewline returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsStringLiteralConcatenationWithNewline() bool {
	return false
}

// SupportsQuoteDelimitedString returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsQuoteDelimitedString() bool {
	return false
}

// SupportsStringEscapeConstant returns true for PostgreSqlDialect.
// PostgreSQL supports E'...' syntax for string literals with escape sequences.
func (d *PostgreSqlDialect) SupportsStringEscapeConstant() bool {
	return true
}

// SupportsFilterDuringAggregation returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsFilterDuringAggregation() bool {
	return true
}

// SupportsWithinAfterArrayAggregation returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsWithinAfterArrayAggregation() bool {
	return false
}

// SupportsWindowClauseNamedWindowReference returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsWindowClauseNamedWindowReference() bool {
	return true
}

// SupportsWindowFunctionNullTreatmentArg returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsWindowFunctionNullTreatmentArg() bool {
	return true
}

// SupportsMatchRecognize returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsMatchRecognize() bool {
	return false
}

// SupportsGroupByExpr returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsGroupByExpr() bool {
	return true
}

// SupportsGroupByWithModifier returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsGroupByWithModifier() bool {
	return false
}

// SupportsLeftAssociativeJoinsWithoutParens returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsLeftAssociativeJoinsWithoutParens() bool {
	return false
}

// SupportsOuterJoinOperator returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsOuterJoinOperator() bool {
	return false
}

// SupportsCrossJoinConstraint returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsCrossJoinConstraint() bool {
	return false
}

// SupportsConnectBy returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsConnectBy() bool {
	return false
}

// SupportsStartTransactionModifier returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsStartTransactionModifier() bool {
	return false
}

// SupportsEndTransactionModifier returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsEndTransactionModifier() bool {
	return false
}

// SupportsNamedFnArgsWithEqOperator returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsNamedFnArgsWithEqOperator() bool {
	return false
}

// SupportsNamedFnArgsWithColonOperator returns true for PostgreSqlDialect.
// Required to support: SELECT json_object('a': 'b')
func (d *PostgreSqlDialect) SupportsNamedFnArgsWithColonOperator() bool {
	return true
}

// SupportsNamedFnArgsWithAssignmentOperator returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsNamedFnArgsWithAssignmentOperator() bool {
	return false
}

// SupportsNamedFnArgsWithRArrowOperator returns true for PostgreSqlDialect.
// PostgreSQL supports func(arg => val) syntax for named arguments.
func (d *PostgreSqlDialect) SupportsNamedFnArgsWithRArrowOperator() bool {
	return true
}

// SupportsNamedFnArgsWithExprName returns true for PostgreSqlDialect.
// Required to support: SELECT json_object('label': 'value')
func (d *PostgreSqlDialect) SupportsNamedFnArgsWithExprName() bool {
	return true
}

// SupportsParenthesizedSetVariables returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsParenthesizedSetVariables() bool {
	return false
}

// SupportsCommaSeparatedSetAssignments returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsCommaSeparatedSetAssignments() bool {
	return false
}

// SupportsSetStmtWithoutOperator returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSetStmtWithoutOperator() bool {
	return false
}

// SupportsSetNames returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSetNames() bool {
	return true
}

// SupportsSelectWildcardExcept returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSelectWildcardExcept() bool {
	return false
}

// SupportsSelectWildcardExclude returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSelectWildcardExclude() bool {
	return false
}

// SupportsSelectExclude returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSelectExclude() bool {
	return false
}

// SupportsSelectWildcardReplace returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSelectWildcardReplace() bool {
	return false
}

// SupportsSelectWildcardIlike returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSelectWildcardIlike() bool {
	return false
}

// SupportsSelectWildcardRename returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSelectWildcardRename() bool {
	return false
}

// SupportsSelectWildcardWithAlias returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSelectWildcardWithAlias() bool {
	return false
}

// SupportsSelectExprStar returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSelectExprStar() bool {
	return false
}

// SupportsFromFirstSelect returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsFromFirstSelect() bool {
	return false
}

// SupportsEmptyProjections returns true for PostgreSqlDialect.
// PostgreSQL supports: SELECT FROM table_name
func (d *PostgreSqlDialect) SupportsEmptyProjections() bool {
	return true
}

// SupportsSelectModifiers returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSelectModifiers() bool {
	return false
}

// SupportsPipeOperator returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsPipeOperator() bool {
	return false
}

// SupportsTrailingCommas returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsTrailingCommas() bool {
	return false
}

// SupportsProjectionTrailingCommas returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsProjectionTrailingCommas() bool {
	return false
}

// SupportsFromTrailingCommas returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsFromTrailingCommas() bool {
	return false
}

// SupportsColumnDefinitionTrailingCommas returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsColumnDefinitionTrailingCommas() bool {
	return false
}

// SupportsLimitComma returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsLimitComma() bool {
	return false
}

// ConvertTypeBeforeValue returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) ConvertTypeBeforeValue() bool {
	return false
}

// SupportsTryConvert returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsTryConvert() bool {
	return false
}

// SupportsBinaryKwAsCast returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsBinaryKwAsCast() bool {
	return false
}

// SupportsObjectNameDoubleDotNotation returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsObjectNameDoubleDotNotation() bool {
	return false
}

// SupportsNumericPrefix returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsNumericPrefix() bool {
	return false
}

// SupportsNumericLiteralUnderscores returns true for PostgreSqlDialect.
// PostgreSQL supports numbers like 1_000_000.
func (d *PostgreSqlDialect) SupportsNumericLiteralUnderscores() bool {
	return true
}

// SupportsInEmptyList returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsInEmptyList() bool {
	return false
}

// SupportsDictionarySyntax returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsDictionarySyntax() bool {
	return false
}

// SupportsMapLiteralSyntax returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsMapLiteralSyntax() bool {
	return false
}

// SupportsStructLiteral returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsStructLiteral() bool {
	return false
}

// SupportsArrayLiteralSyntax returns true for PostgreSqlDialect.
// PostgreSQL supports ARRAY[1,2,3] syntax.
func (d *PostgreSqlDialect) SupportsArrayLiteralSyntax() bool {
	return true
}

// SupportsLambdaFunctions returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsLambdaFunctions() bool {
	return false
}

// SupportsCreateTableMultiSchemaInfoSources returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsCreateTableMultiSchemaInfoSources() bool {
	return false
}

// SupportsCreateTableLikeParenthesized returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsCreateTableLikeParenthesized() bool {
	return true
}

// SupportsCreateTableSelect returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsCreateTableSelect() bool {
	return false
}

// SupportsCreateViewCommentSyntax returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsCreateViewCommentSyntax() bool {
	return false
}

// SupportsArrayTypedefWithoutElementType returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsArrayTypedefWithoutElementType() bool {
	return false
}

// SupportsArrayTypedefWithBrackets returns true for PostgreSqlDialect.
// See https://www.postgresql.org/docs/current/arrays.html#ARRAYS-DECLARATION
func (d *PostgreSqlDialect) SupportsArrayTypedefWithBrackets() bool {
	return true
}

// SupportsParensAroundTableFactor returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsParensAroundTableFactor() bool {
	return false
}

// SupportsValuesAsTableFactor returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsValuesAsTableFactor() bool {
	return false
}

// SupportsUnnestTableFactor returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsUnnestTableFactor() bool {
	return true
}

// SupportsSemanticViewTableFactor returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSemanticViewTableFactor() bool {
	return false
}

// SupportsTableVersioning returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsTableVersioning() bool {
	return false
}

// SupportsTableSampleBeforeAlias returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsTableSampleBeforeAlias() bool {
	return false
}

// SupportsTableHints returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsTableHints() bool {
	return false
}

// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
func (d *PostgreSqlDialect) SupportsIndexHints() bool {
	return false
}

// SupportsAscDescInColumnDefinition returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsAscDescInColumnDefinition() bool {
	return false
}

// SupportsSpaceSeparatedColumnOptions returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSpaceSeparatedColumnOptions() bool {
	return false
}

// SupportsConstraintKeywordWithoutName returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsConstraintKeywordWithoutName() bool {
	return false
}

// SupportsKeyColumnOption returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsKeyColumnOption() bool {
	return false
}

// SupportsDataTypeSignedSuffix returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsDataTypeSignedSuffix() bool {
	return false
}

// SupportsNestedComments returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsNestedComments() bool {
	return true
}

// SupportsMultilineCommentHints returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsMultilineCommentHints() bool {
	return false
}

// SupportsCommentOptimizerHint returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsCommentOptimizerHint() bool {
	return false
}

// SupportsCommentOn returns true for PostgreSqlDialect.
// See https://www.postgresql.org/docs/current/sql-comment.html
func (d *PostgreSqlDialect) SupportsCommentOn() bool {
	return true
}

// RequiresSingleLineCommentWhitespace returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) RequiresSingleLineCommentWhitespace() bool {
	return false
}

// SupportsExplainWithUtilityOptions returns true for PostgreSqlDialect.
// See https://www.postgresql.org/docs/current/sql-explain.html
func (d *PostgreSqlDialect) SupportsExplainWithUtilityOptions() bool {
	return true
}

// SupportsExecuteImmediate returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsExecuteImmediate() bool {
	return false
}

// AllowExtractCustom returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) AllowExtractCustom() bool {
	return true
}

// AllowExtractSingleQuotes returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) AllowExtractSingleQuotes() bool {
	return true
}

// SupportsExtractCommaSyntax returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsExtractCommaSyntax() bool {
	return false
}

// SupportsSubqueryAsFunctionArg returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSubqueryAsFunctionArg() bool {
	return false
}

// SupportsDollarPlaceholder returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsDollarPlaceholder() bool {
	return true
}

// SupportsCreateIndexWithClause returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsCreateIndexWithClause() bool {
	return true
}

// RequireIntervalQualifier returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) RequireIntervalQualifier() bool {
	return false
}

// SupportsIntervalOptions returns true for PostgreSqlDialect.
// See https://www.postgresql.org/docs/17/datatype-datetime.html
func (d *PostgreSqlDialect) SupportsIntervalOptions() bool {
	return true
}

// SupportsFactorialOperator returns true for PostgreSqlDialect.
// See https://www.postgresql.org/docs/13/functions-math.html
func (d *PostgreSqlDialect) SupportsFactorialOperator() bool {
	return true
}

// SupportsBitwiseShiftOperators returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsBitwiseShiftOperators() bool {
	return true
}

// SupportsNotnullOperator returns true for PostgreSqlDialect.
// PostgreSQL supports NOTNULL as an alias for IS NOT NULL.
// See https://www.postgresql.org/docs/17/functions-comparison.html
func (d *PostgreSqlDialect) SupportsNotnullOperator() bool {
	return true
}

// SupportsBangNotOperator returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsBangNotOperator() bool {
	return false
}

// SupportsDoubleAmpersandOperator returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsDoubleAmpersandOperator() bool {
	return false
}

// SupportsMatchAgainst returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsMatchAgainst() bool {
	return false
}

// SupportsUserHostGrantee returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsUserHostGrantee() bool {
	return false
}

// SupportsListenNotify returns true for PostgreSqlDialect.
// See https://www.postgresql.org/docs/current/sql-listen.html
// See https://www.postgresql.org/docs/current/sql-unlisten.html
// See https://www.postgresql.org/docs/current/sql-notify.html
func (d *PostgreSqlDialect) SupportsListenNotify() bool {
	return true
}

// SupportsLoadData returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsLoadData() bool {
	return false
}

// SupportsLoadExtension returns true for PostgreSqlDialect.
// See https://www.postgresql.org/docs/current/sql-load.html
func (d *PostgreSqlDialect) SupportsLoadExtension() bool {
	return true
}

// SupportsTopBeforeDistinct returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsTopBeforeDistinct() bool {
	return false
}

// SupportsBooleanLiterals returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsBooleanLiterals() bool {
	return true
}

// SupportsShowLikeBeforeIn returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsShowLikeBeforeIn() bool {
	return false
}

// SupportsPartiQL returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsPartiQL() bool {
	return false
}

// SupportsEqAliasAssignment returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsEqAliasAssignment() bool {
	return false
}

// SupportsInsertSet returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsInsertSet() bool {
	return false
}

// SupportsInsertTableFunction returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsInsertTableFunction() bool {
	return false
}

// SupportsInsertTableQuery returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsInsertTableQuery() bool {
	return false
}

// SupportsInsertFormat returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsInsertFormat() bool {
	return false
}

// SupportsInsertTableAlias returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsInsertTableAlias() bool {
	return true
}

// SupportsAlterColumnTypeUsing returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsAlterColumnTypeUsing() bool {
	return true
}

// SupportsCommaSeparatedDropColumnList returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsCommaSeparatedDropColumnList() bool {
	return false
}

// SupportsOrderByAll returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsOrderByAll() bool {
	return false
}

// SupportsGeometricTypes returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsGeometricTypes() bool {
	return true
}

// DescribeRequiresTableKeyword returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) DescribeRequiresTableKeyword() bool {
	return false
}

// SupportsOptimizeTable returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsOptimizeTable() bool {
	return false
}

// SupportsPrewhere returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsPrewhere() bool {
	return false
}

// SupportsWithFill returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsWithFill() bool {
	return false
}

// SupportsLimitBy returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsLimitBy() bool {
	return false
}

// SupportsInterpolate returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsInterpolate() bool {
	return false
}

// SupportsSettings returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSettings() bool {
	return false
}

// SupportsSelectFormat returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsSelectFormat() bool {
	return false
}

// SupportsInstall returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsInstall() bool {
	return false
}

// SupportsDetach returns false for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsDetach() bool {
	return false
}

// SupportsCommaSeparatedTrim returns true for PostgreSqlDialect.
func (d *PostgreSqlDialect) SupportsCommaSeparatedTrim() bool {
	return true
}

// Precedence constants for PostgreSQL operators.
// Higher values mean higher precedence (tighter binding).
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
func (d *PostgreSqlDialect) PrecValue(prec dialects.Precedence) uint8 {
	// Handle cases where multiple Precedence constants share the same value
	// by checking the actual constant, not just the underlying value
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
func (d *PostgreSqlDialect) PrecUnknown() uint8 {
	return 0
}

// GetNextPrecedence returns the dialect-specific precedence override for
// the next token. Returns 0 to fall back to default behavior.
func (d *PostgreSqlDialect) GetNextPrecedence(parser dialects.ParserAccessor) (uint8, error) {
	tok := parser.PeekTokenRef()

	// Check for PostgreSQL-specific tokens that need custom precedence
	switch tok.Token.(type) {
	case token.TokenWord:
		// Check if it's the COLLATE keyword
		if word, ok := tok.Token.(token.TokenWord); ok {
			if word.Keyword == token.COLLATE && !parser.InColumnDefinitionState() {
				return collatePrec, nil
			}
		}
	case token.TokenLBracket:
		return bracketPrec, nil
	case token.TokenArrow, token.TokenLongArrow, token.TokenHashArrow, token.TokenHashLongArrow,
		token.TokenAtArrow, token.TokenArrowAt, token.TokenHashMinus, token.TokenAtQuestion,
		token.TokenAtAt, token.TokenQuestion, token.TokenQuestionAnd, token.TokenQuestionPipe,
		token.TokenExclamationMark, token.TokenOverlap, token.TokenCaretAt, token.TokenStringConcat,
		token.TokenSharp, token.TokenShiftRight, token.TokenShiftLeft, token.TokenCustomBinaryOperator:
		return pgOtherPrec, nil
	case token.TokenColon:
		// Lowest precedence to prevent turning into a binary operator
		return d.PrecUnknown(), nil
	}

	return 0, nil
}

// GetNextPrecedenceDefault implements the default precedence logic.
func (d *PostgreSqlDialect) GetNextPrecedenceDefault(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// ParsePrefix is a dialect-specific prefix parser override.
// Returns false to fall back to default behavior.
func (d *PostgreSqlDialect) ParsePrefix(parser dialects.ParserAccessor) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseInfix is a dialect-specific infix parser override.
// Returns false to fall back to default behavior.
func (d *PostgreSqlDialect) ParseInfix(parser dialects.ParserAccessor, expr ast.Expr, precedence uint8) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseStatement is a dialect-specific statement parser override.
// Returns false to fall back to default behavior.
func (d *PostgreSqlDialect) ParseStatement(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	return nil, false, nil
}

// ParseColumnOption is a dialect-specific column option parser override.
// Returns false to fall back to default behavior.
func (d *PostgreSqlDialect) ParseColumnOption(parser dialects.ParserAccessor) (dialects.ColumnOption, bool, error) {
	return nil, false, nil
}
