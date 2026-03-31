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

package snowflake

import (
	"strings"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/token"
	"github.com/user/sqlparser/tokenizer"
)

// SnowflakeDialect is a dialect for Snowflake SQL.
// See: https://www.snowflake.com/
type SnowflakeDialect struct{}

// NewSnowflakeDialect creates a new instance of SnowflakeDialect.
func NewSnowflakeDialect() *SnowflakeDialect {
	return &SnowflakeDialect{}
}

// Dialect returns the dialect identifier.
func (d *SnowflakeDialect) Dialect() string {
	return "snowflake"
}

// IsIdentifierStart returns true if the character is a valid start character
// for an unquoted identifier.
// See: https://docs.snowflake.com/en/sql-reference/identifiers-syntax.html
func (d *SnowflakeDialect) IsIdentifierStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// IsIdentifierPart returns true if the character is a valid unquoted
// identifier character.
func (d *SnowflakeDialect) IsIdentifierPart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') || ch == '$' || ch == '_'
}

// IsDelimitedIdentifierStart returns true if the character starts a quoted identifier.
func (d *SnowflakeDialect) IsDelimitedIdentifierStart(ch rune) bool {
	return ch == '"'
}

// IsNestedDelimitedIdentifierStart returns true if the character starts a
// potential nested quoted identifier.
func (d *SnowflakeDialect) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return false
}

// PeekNestedDelimitedIdentifierQuotes returns the outer quote style and
// optional inner quote style if applicable.
func (d *SnowflakeDialect) PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*dialects.NestedIdentifierQuote, int) {
	return nil, pos
}

// IdentifierQuoteStyle returns the character used to quote identifiers.
func (d *SnowflakeDialect) IdentifierQuoteStyle(identifier string) *rune {
	return nil
}

// IsCustomOperatorPart returns true if the character is part of a custom operator.
func (d *SnowflakeDialect) IsCustomOperatorPart(ch rune) bool {
	return false
}

// Reserved keywords for Snowflake table factors
var reservedKeywordsForTableFactor = []token.Keyword{
	token.ALL,
	token.ALTER,
	token.AND,
	token.ANY,
	token.AS,
	token.BETWEEN,
	token.BY,
	token.CHECK,
	token.COLUMN,
	token.CONNECT,
	token.CREATE,
	token.CROSS,
	token.CURRENT,
	token.DELETE,
	token.DISTINCT,
	token.DROP,
	token.ELSE,
	token.EXISTS,
	token.FOLLOWING,
	token.FOR,
	token.FROM,
	token.FULL,
	token.GRANT,
	token.GROUP,
	token.HAVING,
	token.ILIKE,
	token.IN,
	token.INCREMENT,
	token.INNER,
	token.INSERT,
	token.INTERSECT,
	token.INTO,
	token.IS,
	token.JOIN,
	token.LEFT,
	token.LIKE,
	token.MINUS,
	token.NATURAL,
	token.NOT,
	token.NULL,
	token.OF,
	token.ON,
	token.OR,
	token.ORDER,
	token.QUALIFY,
	token.REGEXP,
	token.REVOKE,
	token.RIGHT,
	token.RLIKE,
	token.ROW,
	token.ROWS,
	token.SAMPLE,
	token.SELECT,
	token.SET,
	token.SOME,
	token.START,
	token.TABLE,
	token.TABLESAMPLE,
	token.THEN,
	token.TO,
	token.TRIGGER,
	token.UNION,
	token.UNIQUE,
	token.UPDATE,
	token.USING,
	token.VALUES,
	token.WHEN,
	token.WHENEVER,
	token.WHERE,
	token.WINDOW,
	token.WITH,
}

// IsReservedForIdentifier returns true if the keyword is reserved.
func (d *SnowflakeDialect) IsReservedForIdentifier(kw token.Keyword) bool {
	// Unreserve INTERVAL for Snowflake
	// See: https://docs.snowflake.com/en/sql-reference/reserved-keywords
	if kw == token.INTERVAL {
		return false
	}
	return token.IsReservedForIdentifier(kw)
}

// GetReservedKeywordsForSelectItemOperator returns reserved keywords for select items.
func (d *SnowflakeDialect) GetReservedKeywordsForSelectItemOperator() []token.Keyword {
	return []token.Keyword{token.CONNECT_BY_ROOT}
}

// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
func (d *SnowflakeDialect) GetReservedGranteesTypes() []dialects.GranteesType {
	return nil
}

// IsColumnAlias returns true if the keyword should be parsed as a column alias.
func (d *SnowflakeDialect) IsColumnAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	switch kw {
	// The following keywords can be considered an alias as long as
	// they are not followed by other tokens that may change their meaning
	case token.EXCEPT:
		// e.g. `SELECT * EXCEPT (col1) FROM tbl`
		if !isCommaOrEOF(parser.PeekTokenRef()) {
			return false
		}
	case token.RETURNING:
		// e.g. `INSERT INTO t SELECT 1 RETURNING *`
		if !isSemicolonOrEOF(parser.PeekTokenRef()) {
			return false
		}

	// e.g. `SELECT 1 LIMIT 5` - not an alias
	// e.g. `SELECT 1 OFFSET 5 ROWS` - not an alias
	case token.LIMIT, token.OFFSET:
		if d.peekForLimitOptions(parser) {
			return false
		}

	// `FETCH` can be considered an alias as long as it's not followed by `FIRST` or `NEXT`
	case token.FETCH:
		if parser.PeekOneOfKeywords([]string{"FIRST", "NEXT"}) != "" || d.peekForLimitOptions(parser) {
			return false
		}

	// Reserved keywords by the Snowflake dialect
	case token.FROM, token.GROUP, token.HAVING, token.INTERSECT,
		token.INTO, token.MINUS, token.ORDER, token.SELECT,
		token.UNION, token.WHERE, token.WITH:
		return false
	}

	// Any other word is considered an alias
	return true
}

// IsSelectItemAlias returns true if the specified keyword should be parsed as a select item alias
func (d *SnowflakeDialect) IsSelectItemAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return d.IsColumnAlias(kw, parser)
}

// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier
func (d *SnowflakeDialect) IsTableFactor(kw token.Keyword, parser dialects.ParserAccessor) bool {
	switch kw {
	case token.LIMIT:
		if d.peekForLimitOptions(parser) {
			return false
		}
	case token.TABLE:
		// Table function - check if followed by LParen
		if _, ok := parser.PeekTokenRef().Token.(tokenizer.TokenChar); ok {
			// Check if it's a left paren
			if charTok, ok := parser.PeekTokenRef().Token.(tokenizer.TokenChar); ok && charTok.Char == '(' {
				return true
			}
		}
	}

	for _, reserved := range reservedKeywordsForTableFactor {
		if kw == reserved {
			return false
		}
	}
	return true
}

// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias
func (d *SnowflakeDialect) IsTableAlias(kw token.Keyword, parser dialects.ParserAccessor) bool {
	switch kw {
	// The following keywords can be considered an alias as long as
	// they are not followed by other tokens that may change their meaning
	case token.RETURNING, token.PIVOT, token.UNPIVOT, token.EXCEPT,
		token.MATCH_RECOGNIZE:
		if !isSemicolonOrEOF(parser.PeekTokenRef()) {
			return false
		}

	// `LIMIT` can be considered an alias as long as it's not followed by a value
	case token.LIMIT, token.OFFSET:
		if d.peekForLimitOptions(parser) {
			return false
		}

	// `FETCH` can be considered an alias as long as it's not followed by `FIRST` or `NEXT`
	case token.FETCH:
		if parser.PeekOneOfKeywords([]string{"FIRST", "NEXT"}) != "" || d.peekForLimitOptions(parser) {
			return false
		}

	// All sorts of join-related keywords can be considered aliases unless additional
	// keywords change their meaning.
	case token.RIGHT, token.LEFT, token.SEMI, token.ANTI:
		if parser.PeekOneOfKeywords([]string{"JOIN", "OUTER"}) != "" {
			return false
		}

	case token.GLOBAL:
		if parser.PeekKeyword("FULL") {
			return false
		}

	// Reserved keywords by the Snowflake dialect
	case token.WITH, token.ORDER, token.SELECT, token.WHERE,
		token.GROUP, token.HAVING, token.LATERAL, token.UNION,
		token.INTERSECT, token.MINUS, token.ON, token.JOIN,
		token.INNER, token.CROSS, token.FULL,
		token.NATURAL, token.USING, token.ASOF,
		token.MATCH_CONDITION, token.SET, token.QUALIFY, token.FOR,
		token.START, token.CONNECT, token.SAMPLE, token.TABLESAMPLE,
		token.FROM:
		return false
	}

	// Any other word is considered an alias
	return true
}

// IsTableFactorAlias returns true if the specified keyword should be parsed as a table factor alias
func (d *SnowflakeDialect) IsTableFactorAlias(explicit bool, kw token.Keyword, parser dialects.ParserAccessor) bool {
	return explicit || d.IsTableAlias(kw, parser)
}

// IsIdentifierGeneratingFunctionName returns true if the ident is a function that generates identifiers.
func (d *SnowflakeDialect) IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool {
	if ident == nil || ident.Value == "" {
		return false
	}
	if ident.QuoteStyle != nil {
		return false
	}
	if !strings.EqualFold(ident.Value, "identifier") {
		return false
	}
	for _, p := range nameParts {
		if _, ok := p.(*ast.ObjectNamePartFunction); ok {
			return false
		}
	}
	return true
}

// Helper function to check if token is comma or EOF
func isCommaOrEOF(t *tokenizer.TokenWithSpan) bool {
	if _, ok := t.Token.(tokenizer.EOF); ok {
		return true
	}
	if charTok, ok := t.Token.(tokenizer.TokenChar); ok && charTok.Char == ',' {
		return true
	}
	return false
}

// Helper function to check if token is semicolon or EOF
func isSemicolonOrEOF(t *tokenizer.TokenWithSpan) bool {
	if _, ok := t.Token.(tokenizer.EOF); ok {
		return true
	}
	if charTok, ok := t.Token.(tokenizer.TokenChar); ok && charTok.Char == ';' {
		return true
	}
	return false
}

// peekForLimitOptions checks if the next token suggests a LIMIT/FETCH option
func (d *SnowflakeDialect) peekForLimitOptions(parser dialects.ParserAccessor) bool {
	tok := parser.PeekTokenRef()
	switch tok.Token.(type) {
	case tokenizer.TokenNumber, tokenizer.TokenPlaceholder:
		return true
	case tokenizer.TokenSingleQuotedString:
		if strTok, ok := tok.Token.(tokenizer.TokenSingleQuotedString); ok {
			return strTok.Value == ""
		}
	case tokenizer.TokenDollarQuotedString:
		if dlrTok, ok := tok.Token.(tokenizer.TokenDollarQuotedString); ok {
			return dlrTok.Value == ""
		}
	default:
		if wordTok, ok := tok.Token.(tokenizer.TokenWord); ok && wordTok.Word.Keyword == token.NULL {
			return true
		}
	}
	return false
}

// SupportsStringLiteralBackslashEscape returns true for Snowflake.
// See https://cloud.google.com/bigquery/docs/reference/standard-sql/lexical#escape_sequences
func (d *SnowflakeDialect) SupportsStringLiteralBackslashEscape() bool {
	return true
}

// IgnoresWildcardEscapes returns false for Snowflake.
func (d *SnowflakeDialect) IgnoresWildcardEscapes() bool {
	return false
}

// SupportsUnicodeStringLiteral returns false for Snowflake.
func (d *SnowflakeDialect) SupportsUnicodeStringLiteral() bool {
	return false
}

// SupportsTripleQuotedString returns false for Snowflake.
func (d *SnowflakeDialect) SupportsTripleQuotedString() bool {
	return false
}

// SupportsStringLiteralConcatenation returns false for Snowflake.
func (d *SnowflakeDialect) SupportsStringLiteralConcatenation() bool {
	return false
}

// SupportsStringLiteralConcatenationWithNewline returns false for Snowflake.
func (d *SnowflakeDialect) SupportsStringLiteralConcatenationWithNewline() bool {
	return false
}

// SupportsQuoteDelimitedString returns false for Snowflake.
func (d *SnowflakeDialect) SupportsQuoteDelimitedString() bool {
	return false
}

// SupportsStringEscapeConstant returns false for Snowflake.
func (d *SnowflakeDialect) SupportsStringEscapeConstant() bool {
	return false
}

// SupportsFilterDuringAggregation returns true for Snowflake.
func (d *SnowflakeDialect) SupportsFilterDuringAggregation() bool {
	return true
}

// SupportsWithinAfterArrayAggregation returns true for Snowflake.
func (d *SnowflakeDialect) SupportsWithinAfterArrayAggregation() bool {
	return true
}

// SupportsWindowClauseNamedWindowReference returns true for Snowflake.
func (d *SnowflakeDialect) SupportsWindowClauseNamedWindowReference() bool {
	return true
}

// SupportsWindowFunctionNullTreatmentArg returns true for Snowflake.
// Snowflake supports FIRST_VALUE(arg, { IGNORE | RESPECT } NULLS) inside the argument list.
func (d *SnowflakeDialect) SupportsWindowFunctionNullTreatmentArg() bool {
	return true
}

// SupportsMatchRecognize returns true for Snowflake.
func (d *SnowflakeDialect) SupportsMatchRecognize() bool {
	return true
}

// SupportsGroupByExpr returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/constructs/group-by
func (d *SnowflakeDialect) SupportsGroupByExpr() bool {
	return true
}

// SupportsGroupByWithModifier returns false for Snowflake.
func (d *SnowflakeDialect) SupportsGroupByWithModifier() bool {
	return false
}

// SupportsLeftAssociativeJoinsWithoutParens returns false for Snowflake.
func (d *SnowflakeDialect) SupportsLeftAssociativeJoinsWithoutParens() bool {
	return false
}

// SupportsOuterJoinOperator returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/constructs/where#joins-in-the-where-clause
func (d *SnowflakeDialect) SupportsOuterJoinOperator() bool {
	return true
}

// SupportsCrossJoinConstraint returns false for Snowflake.
func (d *SnowflakeDialect) SupportsCrossJoinConstraint() bool {
	return false
}

// SupportsConnectBy returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/constructs/connect-by
func (d *SnowflakeDialect) SupportsConnectBy() bool {
	return true
}

// SupportsStartTransactionModifier returns false for Snowflake.
func (d *SnowflakeDialect) SupportsStartTransactionModifier() bool {
	return false
}

// SupportsEndTransactionModifier returns false for Snowflake.
func (d *SnowflakeDialect) SupportsEndTransactionModifier() bool {
	return false
}

// SupportsNamedFnArgsWithEqOperator returns false for Snowflake.
func (d *SnowflakeDialect) SupportsNamedFnArgsWithEqOperator() bool {
	return false
}

// SupportsNamedFnArgsWithColonOperator returns false for Snowflake.
func (d *SnowflakeDialect) SupportsNamedFnArgsWithColonOperator() bool {
	return false
}

// SupportsNamedFnArgsWithAssignmentOperator returns false for Snowflake.
func (d *SnowflakeDialect) SupportsNamedFnArgsWithAssignmentOperator() bool {
	return false
}

// SupportsNamedFnArgsWithRArrowOperator returns false for Snowflake.
func (d *SnowflakeDialect) SupportsNamedFnArgsWithRArrowOperator() bool {
	return false
}

// SupportsNamedFnArgsWithExprName returns false for Snowflake.
func (d *SnowflakeDialect) SupportsNamedFnArgsWithExprName() bool {
	return false
}

// SupportsParenthesizedSetVariables returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/sql/set#syntax
func (d *SnowflakeDialect) SupportsParenthesizedSetVariables() bool {
	return true
}

// SupportsCommaSeparatedSetAssignments returns false for Snowflake.
func (d *SnowflakeDialect) SupportsCommaSeparatedSetAssignments() bool {
	return false
}

// SupportsSetStmtWithoutOperator returns false for Snowflake.
func (d *SnowflakeDialect) SupportsSetStmtWithoutOperator() bool {
	return false
}

// SupportsSetNames returns false for Snowflake.
func (d *SnowflakeDialect) SupportsSetNames() bool {
	return false
}

// SupportsSelectWildcardExcept returns false for Snowflake.
func (d *SnowflakeDialect) SupportsSelectWildcardExcept() bool {
	return false
}

// SupportsSelectWildcardExclude returns true for Snowflake.
func (d *SnowflakeDialect) SupportsSelectWildcardExclude() bool {
	return true
}

// SupportsSelectExclude returns false for Snowflake.
func (d *SnowflakeDialect) SupportsSelectExclude() bool {
	return false
}

// SupportsSelectWildcardReplace returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/sql/select#parameters
func (d *SnowflakeDialect) SupportsSelectWildcardReplace() bool {
	return true
}

// SupportsSelectWildcardIlike returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/sql/select#parameters
func (d *SnowflakeDialect) SupportsSelectWildcardIlike() bool {
	return true
}

// SupportsSelectWildcardRename returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/sql/select#parameters
func (d *SnowflakeDialect) SupportsSelectWildcardRename() bool {
	return true
}

// SupportsSelectWildcardWithAlias returns false for Snowflake.
func (d *SnowflakeDialect) SupportsSelectWildcardWithAlias() bool {
	return false
}

// SupportsSelectExprStar returns true for Snowflake.
// For example: `SELECT IDENTIFIER('alias1').* FROM tbl AS alias1`
func (d *SnowflakeDialect) SupportsSelectExprStar() bool {
	return true
}

// SupportsFromFirstSelect returns false for Snowflake.
func (d *SnowflakeDialect) SupportsFromFirstSelect() bool {
	return false
}

// SupportsEmptyProjections returns false for Snowflake.
func (d *SnowflakeDialect) SupportsEmptyProjections() bool {
	return false
}

// SupportsSelectModifiers returns false for Snowflake.
func (d *SnowflakeDialect) SupportsSelectModifiers() bool {
	return false
}

// SupportsPipeOperator returns false for Snowflake.
func (d *SnowflakeDialect) SupportsPipeOperator() bool {
	return false
}

// SupportsTrailingCommas returns true for Snowflake.
func (d *SnowflakeDialect) SupportsTrailingCommas() bool {
	return true
}

// SupportsProjectionTrailingCommas returns true for Snowflake.
func (d *SnowflakeDialect) SupportsProjectionTrailingCommas() bool {
	return true
}

// SupportsFromTrailingCommas returns true for Snowflake.
func (d *SnowflakeDialect) SupportsFromTrailingCommas() bool {
	return true
}

// SupportsColumnDefinitionTrailingCommas returns false for Snowflake.
func (d *SnowflakeDialect) SupportsColumnDefinitionTrailingCommas() bool {
	return false
}

// SupportsLimitComma returns false for Snowflake.
func (d *SnowflakeDialect) SupportsLimitComma() bool {
	return false
}

// ConvertTypeBeforeValue returns false for Snowflake.
func (d *SnowflakeDialect) ConvertTypeBeforeValue() bool {
	return false
}

// SupportsTryConvert returns true for Snowflake.
func (d *SnowflakeDialect) SupportsTryConvert() bool {
	return true
}

// SupportsBinaryKwAsCast returns false for Snowflake.
func (d *SnowflakeDialect) SupportsBinaryKwAsCast() bool {
	return false
}

// SupportsObjectNameDoubleDotNotation returns true for Snowflake.
// Snowflake supports double-dot notation when the schema name is not specified.
// In this case the default PUBLIC schema is used.
// See: https://docs.snowflake.com/en/sql-reference/name-resolution#resolution-when-schema-omitted-double-dot-notation
func (d *SnowflakeDialect) SupportsObjectNameDoubleDotNotation() bool {
	return true
}

// SupportsNumericPrefix returns false for Snowflake.
func (d *SnowflakeDialect) SupportsNumericPrefix() bool {
	return false
}

// SupportsNumericLiteralUnderscores returns false for Snowflake.
func (d *SnowflakeDialect) SupportsNumericLiteralUnderscores() bool {
	return false
}

// SupportsInEmptyList returns false for Snowflake.
func (d *SnowflakeDialect) SupportsInEmptyList() bool {
	return false
}

// SupportsDictionarySyntax returns true for Snowflake.
// Snowflake uses this syntax for "object constants".
// See: https://docs.snowflake.com/en/sql-reference/data-types-semistructured#label-object-constant
func (d *SnowflakeDialect) SupportsDictionarySyntax() bool {
	return true
}

// SupportsMapLiteralSyntax returns false for Snowflake.
func (d *SnowflakeDialect) SupportsMapLiteralSyntax() bool {
	return false
}

// SupportsStructLiteral returns false for Snowflake.
func (d *SnowflakeDialect) SupportsStructLiteral() bool {
	return false
}

// SupportsArrayLiteralSyntax returns false for Snowflake.
func (d *SnowflakeDialect) SupportsArrayLiteralSyntax() bool {
	return false
}

// SupportsLambdaFunctions returns true for Snowflake.
// See: https://docs.snowflake.com/en/user-guide/querying-semistructured#label-higher-order-functions
func (d *SnowflakeDialect) SupportsLambdaFunctions() bool {
	return true
}

// SupportsCreateTableMultiSchemaInfoSources returns true for Snowflake.
func (d *SnowflakeDialect) SupportsCreateTableMultiSchemaInfoSources() bool {
	return true
}

// SupportsCreateTableLikeParenthesized returns false for Snowflake.
func (d *SnowflakeDialect) SupportsCreateTableLikeParenthesized() bool {
	return false
}

// SupportsCreateTableSelect returns true for Snowflake.
func (d *SnowflakeDialect) SupportsCreateTableSelect() bool {
	return true
}

// SupportsCreateViewCommentSyntax returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/sql/create-view#optional-parameters
func (d *SnowflakeDialect) SupportsCreateViewCommentSyntax() bool {
	return true
}

// SupportsArrayTypedefWithoutElementType returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/data-types-semistructured#array
func (d *SnowflakeDialect) SupportsArrayTypedefWithoutElementType() bool {
	return true
}

// SupportsArrayTypedefWithBrackets returns false for Snowflake.
func (d *SnowflakeDialect) SupportsArrayTypedefWithBrackets() bool {
	return false
}

// SupportsParensAroundTableFactor returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/constructs/from
func (d *SnowflakeDialect) SupportsParensAroundTableFactor() bool {
	return true
}

// SupportsValuesAsTableFactor returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/constructs/values
func (d *SnowflakeDialect) SupportsValuesAsTableFactor() bool {
	return true
}

// SupportsSemanticViewTableFactor returns true for Snowflake.
func (d *SnowflakeDialect) SupportsSemanticViewTableFactor() bool {
	return true
}

// SupportsTableVersioning returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/constructs/at-before
func (d *SnowflakeDialect) SupportsTableVersioning() bool {
	return true
}

// SupportsTableSampleBeforeAlias returns false for Snowflake.
func (d *SnowflakeDialect) SupportsTableSampleBeforeAlias() bool {
	return false
}

// SupportsTableHints returns false for Snowflake.
func (d *SnowflakeDialect) SupportsTableHints() bool {
	return false
}

// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
func (d *SnowflakeDialect) SupportsIndexHints() bool {
	return false
}

// SupportsAscDescInColumnDefinition returns false for Snowflake.
func (d *SnowflakeDialect) SupportsAscDescInColumnDefinition() bool {
	return false
}

// SupportsSpaceSeparatedColumnOptions returns true for Snowflake.
func (d *SnowflakeDialect) SupportsSpaceSeparatedColumnOptions() bool {
	return true
}

// SupportsConstraintKeywordWithoutName returns true for Snowflake.
func (d *SnowflakeDialect) SupportsConstraintKeywordWithoutName() bool {
	return true
}

// SupportsKeyColumnOption returns false for Snowflake.
func (d *SnowflakeDialect) SupportsKeyColumnOption() bool {
	return false
}

// SupportsDataTypeSignedSuffix returns false for Snowflake.
func (d *SnowflakeDialect) SupportsDataTypeSignedSuffix() bool {
	return false
}

// SupportsNestedComments returns true for Snowflake.
func (d *SnowflakeDialect) SupportsNestedComments() bool {
	return true
}

// SupportsMultilineCommentHints returns false for Snowflake.
func (d *SnowflakeDialect) SupportsMultilineCommentHints() bool {
	return false
}

// SupportsCommentOptimizerHint returns false for Snowflake.
func (d *SnowflakeDialect) SupportsCommentOptimizerHint() bool {
	return false
}

// SupportsCommentOn returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/sql/comment
func (d *SnowflakeDialect) SupportsCommentOn() bool {
	return true
}

// RequiresSingleLineCommentWhitespace returns false for Snowflake.
func (d *SnowflakeDialect) RequiresSingleLineCommentWhitespace() bool {
	return false
}

// SupportsExplainWithUtilityOptions returns false for Snowflake.
func (d *SnowflakeDialect) SupportsExplainWithUtilityOptions() bool {
	return false
}

// SupportsExecuteImmediate returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/sql/execute-immediate
func (d *SnowflakeDialect) SupportsExecuteImmediate() bool {
	return true
}

// AllowExtractCustom returns true for Snowflake.
func (d *SnowflakeDialect) AllowExtractCustom() bool {
	return true
}

// AllowExtractSingleQuotes returns true for Snowflake.
func (d *SnowflakeDialect) AllowExtractSingleQuotes() bool {
	return true
}

// SupportsExtractCommaSyntax returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/functions/extract
func (d *SnowflakeDialect) SupportsExtractCommaSyntax() bool {
	return true
}

// SupportsSubqueryAsFunctionArg returns true for Snowflake.
// See: https://docs.snowflake.com/en/sql-reference/functions/flatten
func (d *SnowflakeDialect) SupportsSubqueryAsFunctionArg() bool {
	return true
}

// SupportsDollarPlaceholder returns true for Snowflake.
func (d *SnowflakeDialect) SupportsDollarPlaceholder() bool {
	return true
}

// SupportsCreateIndexWithClause returns true for Snowflake.
func (d *SnowflakeDialect) SupportsCreateIndexWithClause() bool {
	return true
}

// RequireIntervalQualifier returns false for Snowflake.
func (d *SnowflakeDialect) RequireIntervalQualifier() bool {
	return false
}

// SupportsIntervalOptions returns false for Snowflake.
func (d *SnowflakeDialect) SupportsIntervalOptions() bool {
	return false
}

// SupportsFactorialOperator returns false for Snowflake.
func (d *SnowflakeDialect) SupportsFactorialOperator() bool {
	return false
}

// SupportsBitwiseShiftOperators returns true for Snowflake.
func (d *SnowflakeDialect) SupportsBitwiseShiftOperators() bool {
	return true
}

// SupportsNotnullOperator returns false for Snowflake.
func (d *SnowflakeDialect) SupportsNotnullOperator() bool {
	return false
}

// SupportsBangNotOperator returns false for Snowflake.
func (d *SnowflakeDialect) SupportsBangNotOperator() bool {
	return false
}

// SupportsDoubleAmpersandOperator returns false for Snowflake.
func (d *SnowflakeDialect) SupportsDoubleAmpersandOperator() bool {
	return false
}

// SupportsMatchAgainst returns false for Snowflake.
func (d *SnowflakeDialect) SupportsMatchAgainst() bool {
	return false
}

// SupportsUserHostGrantee returns false for Snowflake.
func (d *SnowflakeDialect) SupportsUserHostGrantee() bool {
	return false
}

// SupportsListenNotify returns false for Snowflake.
func (d *SnowflakeDialect) SupportsListenNotify() bool {
	return false
}

// SupportsLoadData returns false for Snowflake.
func (d *SnowflakeDialect) SupportsLoadData() bool {
	return false
}

// SupportsLoadExtension returns false for Snowflake.
func (d *SnowflakeDialect) SupportsLoadExtension() bool {
	return false
}

// SupportsTopBeforeDistinct returns false for Snowflake.
func (d *SnowflakeDialect) SupportsTopBeforeDistinct() bool {
	return false
}

// SupportsBooleanLiterals returns true for Snowflake.
func (d *SnowflakeDialect) SupportsBooleanLiterals() bool {
	return true
}

// SupportsShowLikeBeforeIn returns true for Snowflake.
// Snowflake expects the LIKE option before the IN option.
// See: https://docs.snowflake.com/en/sql-reference/sql/show-views#syntax
func (d *SnowflakeDialect) SupportsShowLikeBeforeIn() bool {
	return true
}

// SupportsPartiQL returns true for Snowflake.
func (d *SnowflakeDialect) SupportsPartiQL() bool {
	return true
}

// SupportsEqAliasAssignment returns false for Snowflake.
func (d *SnowflakeDialect) SupportsEqAliasAssignment() bool {
	return false
}

// SupportsInsertSet returns false for Snowflake.
func (d *SnowflakeDialect) SupportsInsertSet() bool {
	return false
}

// SupportsInsertTableFunction returns false for Snowflake.
func (d *SnowflakeDialect) SupportsInsertTableFunction() bool {
	return false
}

// SupportsInsertTableQuery returns false for Snowflake.
func (d *SnowflakeDialect) SupportsInsertTableQuery() bool {
	return false
}

// SupportsInsertFormat returns false for Snowflake.
func (d *SnowflakeDialect) SupportsInsertFormat() bool {
	return false
}

// SupportsInsertTableAlias returns true for Snowflake.
func (d *SnowflakeDialect) SupportsInsertTableAlias() bool {
	return true
}

// SupportsAlterColumnTypeUsing returns false for Snowflake.
func (d *SnowflakeDialect) SupportsAlterColumnTypeUsing() bool {
	return false
}

// SupportsCommaSeparatedDropColumnList returns true for Snowflake.
func (d *SnowflakeDialect) SupportsCommaSeparatedDropColumnList() bool {
	return true
}

// SupportsOrderByAll returns false for Snowflake.
func (d *SnowflakeDialect) SupportsOrderByAll() bool {
	return false
}

// SupportsGeometricTypes returns false for Snowflake.
func (d *SnowflakeDialect) SupportsGeometricTypes() bool {
	return false
}

// DescribeRequiresTableKeyword returns true for Snowflake.
func (d *SnowflakeDialect) DescribeRequiresTableKeyword() bool {
	return true
}

// SupportsOptimizeTable returns false for Snowflake.
func (d *SnowflakeDialect) SupportsOptimizeTable() bool {
	return false
}

// SupportsPrewhere returns false for Snowflake.
func (d *SnowflakeDialect) SupportsPrewhere() bool {
	return false
}

// SupportsWithFill returns false for Snowflake.
func (d *SnowflakeDialect) SupportsWithFill() bool {
	return false
}

// SupportsLimitBy returns false for Snowflake.
func (d *SnowflakeDialect) SupportsLimitBy() bool {
	return false
}

// SupportsInterpolate returns false for Snowflake.
func (d *SnowflakeDialect) SupportsInterpolate() bool {
	return false
}

// SupportsSettings returns false for Snowflake.
func (d *SnowflakeDialect) SupportsSettings() bool {
	return false
}

// SupportsSelectFormat returns false for Snowflake.
func (d *SnowflakeDialect) SupportsSelectFormat() bool {
	return false
}

// SupportsInstall returns false for Snowflake.
func (d *SnowflakeDialect) SupportsInstall() bool {
	return false
}

// SupportsDetach returns false for Snowflake.
func (d *SnowflakeDialect) SupportsDetach() bool {
	return false
}

// SupportsCommaSeparatedTrim returns true for Snowflake.
func (d *SnowflakeDialect) SupportsCommaSeparatedTrim() bool {
	return true
}

// PrecValue returns the precedence value for a given Precedence level.
func (d *SnowflakeDialect) PrecValue(prec dialects.Precedence) uint8 {
	// Using if-else chain to avoid duplicate case issues with constants that have the same value
	if prec == dialects.PrecedencePeriod {
		return 100
	} else if prec == dialects.PrecedenceDoubleColon {
		return 50
	} else if prec == dialects.PrecedenceAtTz {
		return 41
	} else if prec == dialects.PrecedenceMulDivModOp {
		return 40
	} else if prec == dialects.PrecedencePlusMinus {
		return 30
	} else if prec == dialects.PrecedenceXor {
		return 24
	} else if prec == dialects.PrecedenceAmpersand {
		return 23
	} else if prec == dialects.PrecedenceCaret {
		return 22
	} else if prec == dialects.PrecedencePipe || prec == dialects.PrecedenceColon {
		return 21
	} else if prec == dialects.PrecedenceBetween || prec == dialects.PrecedenceEq {
		return 20
	} else if prec == dialects.PrecedenceLike {
		return 19
	} else if prec == dialects.PrecedenceIs {
		return 17
	} else if prec == dialects.PrecedencePgOther {
		return 16
	} else if prec == dialects.PrecedenceUnaryNot {
		return 15
	} else if prec == dialects.PrecedenceAnd {
		return 10
	} else if prec == dialects.PrecedenceOr {
		return 5
	}
	return d.PrecUnknown()
}

// PrecUnknown returns the precedence when precedence is otherwise unknown.
func (d *SnowflakeDialect) PrecUnknown() uint8 {
	return 0
}

// GetNextPrecedence returns the precedence override for the next token.
// Snowflake supports the `:` cast operator with higher precedence.
func (d *SnowflakeDialect) GetNextPrecedence(parser dialects.ParserAccessor) (uint8, error) {
	tok := parser.PeekTokenRef()
	if _, ok := tok.Token.(tokenizer.TokenChar); ok {
		if charTok, ok := tok.Token.(tokenizer.TokenChar); ok && charTok.Char == ':' {
			return d.PrecValue(dialects.PrecedenceDoubleColon), nil
		}
	}
	return 0, nil
}

// GetNextPrecedenceDefault implements the default precedence logic.
func (d *SnowflakeDialect) GetNextPrecedenceDefault(parser dialects.ParserAccessor) (uint8, error) {
	return 0, nil
}

// ParsePrefix returns false to fall back to default behavior.
func (d *SnowflakeDialect) ParsePrefix(parser dialects.ParserAccessor) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseInfix returns false to fall back to default behavior.
func (d *SnowflakeDialect) ParseInfix(parser dialects.ParserAccessor, expr ast.Expr, precedence uint8) (ast.Expr, bool, error) {
	return nil, false, nil
}

// ParseStatement returns false to fall back to default behavior.
// Snowflake has many custom statements that will need to be implemented here.
func (d *SnowflakeDialect) ParseStatement(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	// TODO: Implement Snowflake-specific statement parsing:
	// - BEGIN ... END blocks (not transactions)
	// - ALTER DYNAMIC TABLE
	// - ALTER EXTERNAL TABLE
	// - ALTER SESSION
	// - CREATE STAGE
	// - CREATE TABLE (with Snowflake-specific options)
	// - CREATE DATABASE
	// - COPY INTO
	// - LIST/LS, REMOVE/RM (file staging commands)
	// - SHOW OBJECTS
	// - Multi-table INSERT (INSERT ALL/FIRST)
	return nil, false, nil
}

// ParseColumnOption returns false to fall back to default behavior.
// Snowflake supports special column options like IDENTITY, AUTOINCREMENT,
// MASKING POLICY, PROJECTION POLICY, and TAG.
func (d *SnowflakeDialect) ParseColumnOption(parser dialects.ParserAccessor) (dialects.ColumnOption, bool, error) {
	// TODO: Implement Snowflake-specific column options:
	// - WITH IDENTITY / IDENTITY
	// - AUTOINCREMENT
	// - [WITH] MASKING POLICY
	// - [WITH] PROJECTION POLICY
	// - [WITH] TAG
	return nil, false, nil
}
