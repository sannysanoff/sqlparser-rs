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
	"fmt"
	"strings"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/query"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/token"
)

// SnowflakeDialect is a dialect for Snowflake SQL.
// See: https://www.snowflake.com/
//
// SnowflakeDialect implements the following capability interfaces:
//   - dialects.CoreDialect
//   - dialects.IdentifierDialect (with $ in identifiers, " quotes)
//   - dialects.KeywordDialect (with IDENTIFIER() function)
//   - dialects.StringLiteralDialect (with backslash escape)
//   - dialects.AggregationDialect (with FILTER, WITHIN GROUP, MATCH_RECOGNIZE)
//   - dialects.GroupByDialect (with GROUP BY Expr)
//   - dialects.JoinDialect (with outer join operator)
//   - dialects.ConnectByDialect (with CONNECT BY support)
//   - dialects.TransactionDialect
//   - dialects.NamedArgumentDialect
//   - dialects.SetDialect (with parenthesized SET variables)
//   - dialects.SelectDialect (with trailing commas, EXCLUDE, REPLACE, ILIKE, RENAME)
//   - dialects.TypeConversionDialect (with TRY_CONVERT)
//   - dialects.ObjectReferenceDialect (with double-dot notation)
//   - dialects.InExpressionDialect
//   - dialects.LiteralDialect (with dictionary, lambda)
//   - dialects.TableDefinitionDialect (with table versioning, VALUES, PIVOT/UNPIVOT)
//   - dialects.ColumnDefinitionDialect
//   - dialects.CommentDialect (with nested comments, COMMENT ON)
//   - dialects.ExplainDialect
//   - dialects.ExecuteDialect (EXECUTE IMMEDIATE)
//   - dialects.ExtractDialect (with comma syntax, single quotes)
//   - dialects.SubqueryDialect
//   - dialects.PlaceholderDialect (dollar placeholders)
//   - dialects.IndexDialect (WITH clause)
//   - dialects.IntervalDialect
//   - dialects.OperatorDialect (with bitwise shift)
//   - dialects.MatchDialect
//   - dialects.GranteeDialect
//   - dialects.ListenNotifyDialect
//   - dialects.LoadDialect
//   - dialects.TopDistinctDialect
//   - dialects.BooleanLiteralDialect
//   - dialects.ShowDialect (LIKE before IN)
//   - dialects.PartiQLDialect
//   - dialects.AliasDialect
//   - dialects.InsertDialect (with table alias)
//   - dialects.AlterTableDialect
//   - dialects.OrderByDialect
//   - dialects.GeometricDialect
//   - dialects.DescribeDialect
//   - dialects.ClickHouseDialect
//   - dialects.DuckDBDialect
//   - dialects.TrimDialect
//
// Compile-time interface checks:
var _ dialects.CompleteDialect = (*SnowflakeDialect)(nil)
var _ dialects.CoreDialect = (*SnowflakeDialect)(nil)
var _ dialects.IdentifierDialect = (*SnowflakeDialect)(nil)
var _ dialects.KeywordDialect = (*SnowflakeDialect)(nil)
var _ dialects.StringLiteralDialect = (*SnowflakeDialect)(nil)
var _ dialects.AggregationDialect = (*SnowflakeDialect)(nil)
var _ dialects.GroupByDialect = (*SnowflakeDialect)(nil)
var _ dialects.JoinDialect = (*SnowflakeDialect)(nil)
var _ dialects.ConnectByDialect = (*SnowflakeDialect)(nil)
var _ dialects.TransactionDialect = (*SnowflakeDialect)(nil)
var _ dialects.NamedArgumentDialect = (*SnowflakeDialect)(nil)
var _ dialects.SetDialect = (*SnowflakeDialect)(nil)
var _ dialects.SelectDialect = (*SnowflakeDialect)(nil)
var _ dialects.TypeConversionDialect = (*SnowflakeDialect)(nil)
var _ dialects.ObjectReferenceDialect = (*SnowflakeDialect)(nil)
var _ dialects.InExpressionDialect = (*SnowflakeDialect)(nil)
var _ dialects.LiteralDialect = (*SnowflakeDialect)(nil)
var _ dialects.TableDefinitionDialect = (*SnowflakeDialect)(nil)
var _ dialects.ColumnDefinitionDialect = (*SnowflakeDialect)(nil)
var _ dialects.CommentDialect = (*SnowflakeDialect)(nil)
var _ dialects.ExplainDialect = (*SnowflakeDialect)(nil)
var _ dialects.ExecuteDialect = (*SnowflakeDialect)(nil)
var _ dialects.ExtractDialect = (*SnowflakeDialect)(nil)
var _ dialects.SubqueryDialect = (*SnowflakeDialect)(nil)
var _ dialects.PlaceholderDialect = (*SnowflakeDialect)(nil)
var _ dialects.IndexDialect = (*SnowflakeDialect)(nil)
var _ dialects.IntervalDialect = (*SnowflakeDialect)(nil)
var _ dialects.OperatorDialect = (*SnowflakeDialect)(nil)
var _ dialects.MatchDialect = (*SnowflakeDialect)(nil)
var _ dialects.GranteeDialect = (*SnowflakeDialect)(nil)
var _ dialects.ListenNotifyDialect = (*SnowflakeDialect)(nil)
var _ dialects.LoadDialect = (*SnowflakeDialect)(nil)
var _ dialects.TopDistinctDialect = (*SnowflakeDialect)(nil)
var _ dialects.BooleanLiteralDialect = (*SnowflakeDialect)(nil)
var _ dialects.ShowDialect = (*SnowflakeDialect)(nil)
var _ dialects.PartiQLDialect = (*SnowflakeDialect)(nil)
var _ dialects.AliasDialect = (*SnowflakeDialect)(nil)
var _ dialects.InsertDialect = (*SnowflakeDialect)(nil)
var _ dialects.AlterTableDialect = (*SnowflakeDialect)(nil)
var _ dialects.OrderByDialect = (*SnowflakeDialect)(nil)
var _ dialects.GeometricDialect = (*SnowflakeDialect)(nil)
var _ dialects.DescribeDialect = (*SnowflakeDialect)(nil)
var _ dialects.ClickHouseDialect = (*SnowflakeDialect)(nil)
var _ dialects.DuckDBDialect = (*SnowflakeDialect)(nil)
var _ dialects.TrimDialect = (*SnowflakeDialect)(nil)

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
		if _, ok := parser.PeekTokenRef().Token.(token.TokenChar); ok {
			// Check if it's a left paren
			if charTok, ok := parser.PeekTokenRef().Token.(token.TokenChar); ok && charTok.Char == '(' {
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
func isCommaOrEOF(t *token.TokenWithSpan) bool {
	if _, ok := t.Token.(token.EOF); ok {
		return true
	}
	if charTok, ok := t.Token.(token.TokenChar); ok && charTok.Char == ',' {
		return true
	}
	return false
}

// Helper function to check if token is semicolon or EOF
func isSemicolonOrEOF(t *token.TokenWithSpan) bool {
	if _, ok := t.Token.(token.EOF); ok {
		return true
	}
	if charTok, ok := t.Token.(token.TokenChar); ok && charTok.Char == ';' {
		return true
	}
	return false
}

// peekForLimitOptions checks if the next token suggests a LIMIT/FETCH option
func (d *SnowflakeDialect) peekForLimitOptions(parser dialects.ParserAccessor) bool {
	tok := parser.PeekTokenRef()
	switch tok.Token.(type) {
	case token.TokenNumber, token.TokenPlaceholder:
		return true
	case token.TokenSingleQuotedString:
		if strTok, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			return strTok.Value == ""
		}
	case token.TokenDollarQuotedString:
		if dlrTok, ok := tok.Token.(token.TokenDollarQuotedString); ok {
			return dlrTok.Value == ""
		}
	default:
		if wordTok, ok := tok.Token.(token.TokenWord); ok && wordTok.Word.Keyword == token.NULL {
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

// SupportsDollarQuotedString returns false for Snowflake.
func (d *SnowflakeDialect) SupportsDollarQuotedString() bool {
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

// SupportsNamedFnArgsWithRArrowOperator returns true for Snowflake.
// Snowflake uses => for named function arguments in functions like FLATTEN.
// See: https://docs.snowflake.com/en/sql-reference/functions/flatten
func (d *SnowflakeDialect) SupportsNamedFnArgsWithRArrowOperator() bool {
	return true
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

// SupportsUnnestTableFactor returns false for SnowflakeDialect.
func (d *SnowflakeDialect) SupportsUnnestTableFactor() bool {
	return false
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
	} else if prec == dialects.PrecedenceCollate {
		return 42
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
	if _, ok := tok.Token.(token.TokenChar); ok {
		if charTok, ok := tok.Token.(token.TokenChar); ok && charTok.Char == ':' {
			return d.PrecValue(dialects.PrecedenceDoubleColon), nil
		}
	}
	return 0, nil
}

// GetNextPrecedenceDefault implements the default precedence logic.
// This method is NOT called by the expression parser - instead the parser
// uses ExpressionParser.GetNextPrecedenceDefault() which has the actual
// precedence logic for all standard operators including >, <, =, etc.
// This method exists only to satisfy the dialect interface.
func (d *SnowflakeDialect) GetNextPrecedenceDefault(parser dialects.ParserAccessor) (uint8, error) {
	// Return 0 to indicate no custom precedence - the expression parser's
	// default implementation will handle all standard operators
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

// ParseStatement implements Snowflake-specific statement parsing.
func (d *SnowflakeDialect) ParseStatement(parser dialects.ParserAccessor) (ast.Statement, bool, error) {
	// Check for COPY INTO statement
	if parser.PeekKeyword("COPY") {
		// Look ahead to see if next keyword is INTO
		if parser.PeekNthKeyword(1, "INTO") {
			stmt, err := d.parseCopyInto(parser)
			return stmt, true, err
		}
	}

	// Check for multi-table INSERT: INSERT [OVERWRITE] {ALL | FIRST}
	if parser.PeekKeyword("INSERT") {
		// Save position for potential rewind
		restore := parser.SavePosition()
		parser.AdvanceToken() // consume INSERT

		// Check for OVERWRITE
		overwrite := parser.ParseKeyword("OVERWRITE")

		// Check for ALL or FIRST
		var multiTableType *expr.MultiTableInsertType
		if parser.ParseKeyword("ALL") {
			t := expr.MultiTableInsertTypeAll
			multiTableType = &t
		} else if parser.ParseKeyword("FIRST") {
			t := expr.MultiTableInsertTypeFirst
			multiTableType = &t
		}

		if multiTableType != nil {
			// This is a multi-table INSERT
			stmt, err := d.parseMultiTableInsert(parser, overwrite, multiTableType)
			return stmt, true, err
		}

		// Not a multi-table insert, restore position
		restore()
	}

	// TODO: Implement other Snowflake-specific statements:
	// - SHOW OBJECTS
	// - LIST/LS, REMOVE/RM (file staging commands)
	// - CREATE STAGE
	return nil, false, nil
}

// parseCopyInto parses a Snowflake COPY INTO statement.
// Reference: src/dialect/snowflake.rs:parse_copy_into
func (d *SnowflakeDialect) parseCopyInto(parser dialects.ParserAccessor) (ast.Statement, error) {
	// Consume COPY INTO keywords
	parser.AdvanceToken() // COPY
	parser.AdvanceToken() // INTO

	// Determine the kind of COPY INTO based on the next token
	kind := expr.CopyIntoSnowflakeKindTable
	nextTok := parser.PeekTokenRef()
	switch tok := nextTok.Token.(type) {
	case token.TokenAtSign:
		// @ indicates an internal stage (location kind)
		kind = expr.CopyIntoSnowflakeKindLocation
	case token.TokenSingleQuotedString:
		// URL-like string indicates external stage (location kind)
		if strings.Contains(tok.Value, "://") {
			kind = expr.CopyIntoSnowflakeKindLocation
		}
	}

	// Parse the target (INTO clause)
	into, err := d.ParseSnowflakeStageName(parser)
	if err != nil {
		return nil, err
	}

	// For location kind, parse stage params after the target
	var stageParams *expr.StageParamsObject
	if kind == expr.CopyIntoSnowflakeKindLocation {
		stageParams, err = d.parseStageParams(parser)
		if err != nil {
			return nil, err
		}
	}

	// Parse optional column list for table kind
	var intoColumns []*ast.Ident
	if _, ok := parser.PeekTokenRef().Token.(token.TokenLParen); ok {
		parser.AdvanceToken() // consume (
		for {
			if _, ok := parser.PeekTokenRef().Token.(token.TokenRParen); ok {
				parser.AdvanceToken() // consume )
				break
			}
			col, err := d.parseIdentifier(parser)
			if err != nil {
				return nil, err
			}
			intoColumns = append(intoColumns, col)
			if !parser.ConsumeToken(token.TokenComma{}) {
				if _, err := parser.ExpectToken(token.TokenRParen{}); err != nil {
					return nil, err
				}
				break
			}
		}
	}

	// Expect FROM keyword
	if _, err := parser.ExpectKeyword("FROM"); err != nil {
		return nil, err
	}

	// Parse the source
	var fromObj *ast.ObjectName
	var fromObjAlias *ast.Ident
	var fromTransformations []*expr.StageLoadSelectItemWrapper
	var fromQuery *query.Query

	nextTok = parser.PeekTokenRef()
	if _, ok := nextTok.Token.(token.TokenLParen); ok {
		// Parenthesized source - could be query or transformations
		parser.AdvanceToken() // consume (

		if kind == expr.CopyIntoSnowflakeKindTable {
			// For table kind, expect SELECT for transformations
			if parser.PeekKeyword("SELECT") {
				parser.AdvanceToken() // consume SELECT
				// Parse transformations (Snowflake-specific select items)
				transforms, err := d.parseSelectItemsForDataLoad(parser)
				if err != nil {
					return nil, err
				}
				fromTransformations = transforms

				// Expect FROM after transformations
				if _, err := parser.ExpectKeyword("FROM"); err != nil {
					return nil, err
				}

				// Parse stage name
				fromObj, err = d.ParseSnowflakeStageName(parser)
				if err != nil {
					return nil, err
				}

				// Parse stage params
				stageParams, err = d.parseStageParams(parser)
				if err != nil {
					return nil, err
				}

				// Parse optional alias
				if parser.ParseKeyword("AS") {
					alias, err := d.parseIdentifier(parser)
					if err != nil {
						return nil, err
					}
					fromObjAlias = alias
				} else {
					// Try to parse as implicit alias
					if word, ok := parser.PeekTokenRef().Token.(token.TokenWord); ok {
						// Check if it looks like an identifier (not a keyword)
						// Also check that it's not a COPY INTO option keyword
						if !token.IsReservedForIdentifier(word.Word.Keyword) &&
							word.Word.Value != "PARTITION" &&
							word.Word.Value != "FILE_FORMAT" &&
							word.Word.Value != "FILES" &&
							word.Word.Value != "PATTERN" &&
							word.Word.Value != "VALIDATION_MODE" &&
							word.Word.Value != "COPY_OPTIONS" {
							parser.AdvanceToken()
							fromObjAlias = &ast.Ident{Value: word.Word.Value}
						}
					}
				}
			}
		} else {
			// For location kind, expect a query
			// Parse the query inside the parentheses
			queryStmt, err := parser.ParseQuery()
			if err != nil {
				return nil, fmt.Errorf("error parsing COPY INTO query: %w", err)
			}

			// Extract the query from the statement
			fromQuery = parser.ExtractQuery(queryStmt)

			// Expect the closing parenthesis
			if _, err := parser.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		}
	} else {
		// Non-parenthesized source - parse stage name
		fromObj, err = d.ParseSnowflakeStageName(parser)
		if err != nil {
			return nil, err
		}

		// Parse stage params
		stageParams, err = d.parseStageParams(parser)
		if err != nil {
			return nil, err
		}

		// Parse optional alias
		if parser.ParseKeyword("AS") {
			alias, err := d.parseIdentifier(parser)
			if err != nil {
				return nil, err
			}
			fromObjAlias = alias
		} else {
			// Try to parse as implicit alias
			if word, ok := parser.PeekTokenRef().Token.(token.TokenWord); ok {
				// Check if it looks like an identifier (not a reserved keyword)
				// Also check that it's not a COPY INTO option keyword
				if !token.IsReservedForIdentifier(word.Word.Keyword) &&
					word.Word.Value != "PARTITION" &&
					word.Word.Value != "FILE_FORMAT" &&
					word.Word.Value != "FILES" &&
					word.Word.Value != "PATTERN" &&
					word.Word.Value != "VALIDATION_MODE" &&
					word.Word.Value != "COPY_OPTIONS" {
					parser.AdvanceToken()
					fromObjAlias = &ast.Ident{Value: word.Word.Value}
				}
			}
		}
	}

	// Parse optional clauses: FILE_FORMAT, FILES, PATTERN, VALIDATION_MODE, COPY_OPTIONS
	var fileFormat, copyOptions *expr.KeyValueOptions
	var files []string
	var pattern, validationMode *string
	var partition ast.Expr

	for {
		nextTok := parser.PeekTokenRef()
		if _, ok := nextTok.Token.(token.EOF); ok {
			break
		}
		if charTok, ok := nextTok.Token.(token.TokenChar); ok && charTok.Char == ';' {
			break
		}

		// Check for FILE_FORMAT
		if parser.PeekKeyword("FILE_FORMAT") {
			parser.AdvanceToken() // consume FILE_FORMAT
			if !parser.ConsumeToken(token.TokenEq{}) {
				return nil, fmt.Errorf("expected = after FILE_FORMAT")
			}
			opts, err := d.parseKeyValueOptions(parser, true)
			if err != nil {
				return nil, err
			}
			fileFormat = opts
			continue
		}

		// Check for PARTITION BY
		if parser.PeekKeyword("PARTITION") {
			parser.AdvanceToken() // consume PARTITION
			if _, err := parser.ExpectKeyword("BY"); err != nil {
				return nil, err
			}
			// Parse partition expression
			expr, err := parser.ParseExpression()
			if err != nil {
				return nil, err
			}
			partition = expr
			continue
		}

		// Check for FILES
		if parser.PeekKeyword("FILES") {
			parser.AdvanceToken() // consume FILES
			if !parser.ConsumeToken(token.TokenEq{}) {
				return nil, fmt.Errorf("expected = after FILES")
			}
			if !parser.ConsumeToken(token.TokenLParen{}) {
				return nil, fmt.Errorf("expected ( after FILES =")
			}
			for {
				tok := parser.NextToken()
				if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
					files = append(files, str.Value)
				} else {
					return nil, fmt.Errorf("expected string literal in FILES list, got %v", tok)
				}
				if parser.ConsumeToken(token.TokenComma{}) {
					continue
				}
				if _, ok := parser.PeekTokenRef().Token.(token.TokenRParen); ok {
					parser.AdvanceToken()
					break
				}
			}
			continue
		}

		// Check for PATTERN
		if parser.PeekKeyword("PATTERN") {
			parser.AdvanceToken() // consume PATTERN
			if !parser.ConsumeToken(token.TokenEq{}) {
				return nil, fmt.Errorf("expected = after PATTERN")
			}
			tok := parser.NextToken()
			if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
				pattern = &str.Value
			} else {
				return nil, fmt.Errorf("expected string literal after PATTERN =, got %v", tok)
			}
			continue
		}

		// Check for VALIDATION_MODE
		if parser.PeekKeyword("VALIDATION_MODE") {
			parser.AdvanceToken() // consume VALIDATION_MODE
			if !parser.ConsumeToken(token.TokenEq{}) {
				return nil, fmt.Errorf("expected = after VALIDATION_MODE")
			}
			tok := parser.NextToken()
			if word, ok := tok.Token.(token.TokenWord); ok {
				val := word.Word.Value
				validationMode = &val
			} else {
				return nil, fmt.Errorf("expected identifier after VALIDATION_MODE =, got %v", tok)
			}
			continue
		}

		// Check for COPY_OPTIONS
		if parser.PeekKeyword("COPY_OPTIONS") {
			parser.AdvanceToken() // consume COPY_OPTIONS
			if !parser.ConsumeToken(token.TokenEq{}) {
				return nil, fmt.Errorf("expected = after COPY_OPTIONS")
			}
			opts, err := d.parseKeyValueOptions(parser, true)
			if err != nil {
				return nil, err
			}
			copyOptions = opts
			continue
		}

		// For location kind, also allow standalone key=value options
		if kind == expr.CopyIntoSnowflakeKindLocation {
			if word, ok := nextTok.Token.(token.TokenWord); ok {
				// Try to parse as key=value option
				opt, err := d.parseKeyValueOption(parser, &word)
				if err == nil && opt != nil {
					if copyOptions == nil {
						copyOptions = &expr.KeyValueOptions{
							Delimiter: expr.KeyValueOptionsDelimiterSpace,
						}
					}
					copyOptions.Options = append(copyOptions.Options, opt)
					continue
				}
			}
		}

		// If we didn't match any known option, break
		break
	}

	return &statement.CopyIntoSnowflake{
		BaseStatement:       statement.BaseStatement{},
		Kind:                &kind,
		Into:                into,
		IntoColumns:         intoColumns,
		FromObj:             fromObj,
		FromObjAlias:        fromObjAlias,
		StageParams:         stageParams,
		FromTransformations: fromTransformations,
		FromQuery:           fromQuery,
		Files:               files,
		Pattern:             pattern,
		FileFormat:          fileFormat,
		CopyOptions:         copyOptions,
		ValidationMode:      validationMode,
		Partition:           partition,
	}, nil
}

// ParseSnowflakeStageName parses a Snowflake stage name which can include:
// - @namespace.%table_name
// - @namespace.stage_name/path
// - @~/path
// - Regular table names
// This is exported for use by the parser package.
func (d *SnowflakeDialect) ParseSnowflakeStageName(parser dialects.ParserAccessor) (*ast.ObjectName, error) {
	parts := []ast.ObjectNamePart{}

	// Check for @ prefix (stage reference)
	if parser.ConsumeToken(token.TokenAtSign{}) {
		// Stage reference - parse the stage identifier
		// Can be: @namespace.stage_name, @stage_name, @~/path, @%table_name
		// Also supports special chars: = (for partitioning), : (for time), /, %, ~, +, -
		var stageName strings.Builder
		stageName.WriteString("@")

		for {
			tok := parser.NextToken()
			switch t := tok.Token.(type) {
			case token.TokenWord:
				stageName.WriteString(t.Word.Value)
				continue // Continue to next token
			case token.TokenPeriod:
				stageName.WriteRune('.')
				continue
			case token.TokenNumber:
				// Handle numbers with trailing periods (e.g., "23." in "23.parquet")
				// The tokenizer treats "23." as a single number token, but in stage paths
				// the period is part of the path separator, so we include it
				stageName.WriteString(t.Value)
				continue // Continue to next token
			case token.TokenChar:
				if t.Char == '/' || t.Char == '%' || t.Char == '~' {
					stageName.WriteRune(t.Char)
					continue
				} else {
					// Not part of stage name - put back the token and return
					parser.PrevToken()
					parts = append(parts, &ast.ObjectNamePartIdentifier{
						Ident: &ast.Ident{Value: stageName.String()},
					})
					return &ast.ObjectName{Parts: parts}, nil
				}
			case token.TokenColon:
				stageName.WriteRune(':')
				continue
			case token.TokenEq:
				stageName.WriteRune('=')
				continue
			case token.TokenPlus:
				stageName.WriteRune('+')
				continue
			case token.TokenMinus:
				stageName.WriteRune('-')
				continue
			case token.TokenMod:
				stageName.WriteRune('%')
				continue
			case token.TokenDiv:
				stageName.WriteRune('/')
				continue
			case token.TokenSingleQuotedString:
				stageName.WriteString("'")
				stageName.WriteString(t.Value)
				stageName.WriteString("'")
				continue
			case token.TokenWhitespace:
				// Whitespace ends the stage name but doesn't need to be put back
				parts = append(parts, &ast.ObjectNamePartIdentifier{
					Ident: &ast.Ident{Value: stageName.String()},
				})
				return &ast.ObjectName{Parts: parts}, nil
			default:
				// End of stage name - put back the token so caller can handle it
				parser.PrevToken()
				parts = append(parts, &ast.ObjectNamePartIdentifier{
					Ident: &ast.Ident{Value: stageName.String()},
				})
				return &ast.ObjectName{Parts: parts}, nil
			}
		}
	}

	// Regular object name (table or string literal for location)
	tok := parser.NextToken()
	switch t := tok.Token.(type) {
	case token.TokenWord:
		ident := &ast.Ident{Value: t.Word.Value}
		parts = append(parts, &ast.ObjectNamePartIdentifier{Ident: ident})

		// Check for more parts (schema.table or db.schema.table)
		for {
			if !parser.ConsumeToken(token.TokenPeriod{}) {
				break
			}
			nextTok := parser.NextToken()
			if word, ok := nextTok.Token.(token.TokenWord); ok {
				parts = append(parts, &ast.ObjectNamePartIdentifier{
					Ident: &ast.Ident{Value: word.Word.Value},
				})
			} else {
				return nil, fmt.Errorf("expected identifier after . in object name, got %v", nextTok)
			}
		}
		return &ast.ObjectName{Parts: parts}, nil

	case token.TokenSingleQuotedString:
		// String literal as object name (e.g., 's3://bucket/file.csv')
		parts = append(parts, &ast.ObjectNamePartIdentifier{
			Ident: &ast.Ident{Value: fmt.Sprintf("'%s'", t.Value)},
		})
		return &ast.ObjectName{Parts: parts}, nil

	default:
		return nil, fmt.Errorf("expected object name or stage reference, got %v", tok)
	}
}

// parseStageParams parses stage parameters like URL, CREDENTIALS, ENCRYPTION, etc.
func (d *SnowflakeDialect) parseStageParams(parser dialects.ParserAccessor) (*expr.StageParamsObject, error) {
	params := &expr.StageParamsObject{
		Credentials: &expr.KeyValueOptions{
			Delimiter: expr.KeyValueOptionsDelimiterSpace,
		},
		Encryption: &expr.KeyValueOptions{
			Delimiter: expr.KeyValueOptionsDelimiterSpace,
		},
	}

	for {
		if parser.PeekKeyword("FILE_FORMAT") ||
			parser.PeekKeyword("FILES") ||
			parser.PeekKeyword("PATTERN") ||
			parser.PeekKeyword("VALIDATION_MODE") ||
			parser.PeekKeyword("COPY_OPTIONS") ||
			parser.PeekKeyword("PARTITION") {
			// These are not stage params, break
			break
		}

		nextTok := parser.PeekTokenRef()
		if _, ok := nextTok.Token.(token.EOF); ok {
			break
		}
		if charTok, ok := nextTok.Token.(token.TokenChar); ok && charTok.Char == ';' {
			break
		}
		if _, ok := nextTok.Token.(token.TokenRParen); ok {
			break
		}

		if word, ok := nextTok.Token.(token.TokenWord); ok {
			switch word.Word.Keyword {
			case token.URL:
				parser.AdvanceToken()
				if !parser.ConsumeToken(token.TokenEq{}) {
					return nil, fmt.Errorf("expected = after URL")
				}
				tok := parser.NextToken()
				if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
					params.Url = &str.Value
				} else {
					return nil, fmt.Errorf("expected string literal after URL =")
				}

			case token.STORAGE_INTEGRATION:
				parser.AdvanceToken()
				if !parser.ConsumeToken(token.TokenEq{}) {
					return nil, fmt.Errorf("expected = after STORAGE_INTEGRATION")
				}
				tok := parser.NextToken()
				if word, ok := tok.Token.(token.TokenWord); ok {
					params.StorageIntegration = &word.Word.Value
				} else {
					return nil, fmt.Errorf("expected identifier after STORAGE_INTEGRATION =")
				}

			case token.ENDPOINT:
				parser.AdvanceToken()
				if !parser.ConsumeToken(token.TokenEq{}) {
					return nil, fmt.Errorf("expected = after ENDPOINT")
				}
				tok := parser.NextToken()
				if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
					params.Endpoint = &str.Value
				} else {
					return nil, fmt.Errorf("expected string literal after ENDPOINT =")
				}

			case token.CREDENTIALS:
				parser.AdvanceToken()
				if !parser.ConsumeToken(token.TokenEq{}) {
					return nil, fmt.Errorf("expected = after CREDENTIALS")
				}
				if !parser.ConsumeToken(token.TokenLParen{}) {
					return nil, fmt.Errorf("expected ( after CREDENTIALS =")
				}
				opts, err := d.parseKeyValueOptions(parser, false)
				if err != nil {
					return nil, err
				}
				params.Credentials = opts
				if !parser.ConsumeToken(token.TokenRParen{}) {
					return nil, fmt.Errorf("expected ) after CREDENTIALS options")
				}

			case token.ENCRYPTION:
				parser.AdvanceToken()
				if !parser.ConsumeToken(token.TokenEq{}) {
					return nil, fmt.Errorf("expected = after ENCRYPTION")
				}
				if !parser.ConsumeToken(token.TokenLParen{}) {
					return nil, fmt.Errorf("expected ( after ENCRYPTION =")
				}
				opts, err := d.parseKeyValueOptions(parser, false)
				if err != nil {
					return nil, err
				}
				params.Encryption = opts
				if !parser.ConsumeToken(token.TokenRParen{}) {
					return nil, fmt.Errorf("expected ) after ENCRYPTION options")
				}

			default:
				// Not a stage param, break
				return params, nil
			}
		} else {
			// Not a word token, break
			break
		}
	}

	return params, nil
}

// parseKeyValueOptions parses a list of key=value options.
func (d *SnowflakeDialect) parseKeyValueOptions(parser dialects.ParserAccessor, allowParens bool) (*expr.KeyValueOptions, error) {
	opts := &expr.KeyValueOptions{
		Delimiter: expr.KeyValueOptionsDelimiterSpace,
	}

	if allowParens {
		if parser.ConsumeToken(token.TokenLParen{}) {
			// Parenthesized options (key=value, key=value)
			opts.Delimiter = expr.KeyValueOptionsDelimiterComma
			for {
				if _, ok := parser.PeekTokenRef().Token.(token.TokenRParen); ok {
					parser.AdvanceToken()
					break
				}
				tok := parser.NextToken()
				if word, ok := tok.Token.(token.TokenWord); ok {
					opt, err := d.parseKeyValueOption(parser, &word)
					if err != nil {
						return nil, err
					}
					opts.Options = append(opts.Options, opt)
				} else {
					return nil, fmt.Errorf("expected option name, got %v", tok)
				}
				if !parser.ConsumeToken(token.TokenComma{}) {
					if _, ok := parser.PeekTokenRef().Token.(token.TokenRParen); ok {
						parser.AdvanceToken()
						break
					}
					return nil, fmt.Errorf("expected , or ) after option")
				}
			}
			return opts, nil
		}
	}

	// Non-parenthesized options key=value key=value
	for {
		tok := parser.PeekTokenRef()
		if _, ok := tok.Token.(token.TokenRParen); ok {
			break
		}
		if word, ok := tok.Token.(token.TokenWord); ok {
			parser.AdvanceToken()
			opt, err := d.parseKeyValueOption(parser, &word)
			if err != nil {
				// If we can't parse as key=value, it might be the end
				parser.PrevToken()
				break
			}
			opts.Options = append(opts.Options, opt)
		} else {
			break
		}
	}

	return opts, nil
}

// parseKeyValueOption parses a single key=value option.
func (d *SnowflakeDialect) parseKeyValueOption(parser dialects.ParserAccessor, keyWord *token.TokenWord) (*expr.KeyValueOption, error) {
	key := keyWord.Word.Value

	if !parser.ConsumeToken(token.TokenEq{}) {
		return nil, fmt.Errorf("expected = after %s", key)
	}

	tok := parser.NextToken()
	switch t := tok.Token.(type) {
	case token.TokenWord:
		return &expr.KeyValueOption{
			OptionName:  key,
			OptionValue: t.Word.Value,
			Kind:        expr.KeyValueOptionKindSingle,
		}, nil
	case token.TokenSingleQuotedString:
		return &expr.KeyValueOption{
			OptionName:  key,
			OptionValue: t.Value,
			Kind:        expr.KeyValueOptionKindSingle,
		}, nil
	case token.TokenNumber:
		return &expr.KeyValueOption{
			OptionName:  key,
			OptionValue: t.Value,
			Kind:        expr.KeyValueOptionKindSingle,
		}, nil
	case token.TokenLParen:
		// Multi-value or nested options
		parser.AdvanceToken() // consume (
		opts, err := d.parseKeyValueOptions(parser, true)
		if err != nil {
			return nil, err
		}
		if !parser.ConsumeToken(token.TokenRParen{}) {
			return nil, fmt.Errorf("expected ) after nested options")
		}
		return &expr.KeyValueOption{
			OptionName:  key,
			OptionValue: opts,
			Kind:        expr.KeyValueOptionKindNested,
		}, nil
	default:
		return nil, fmt.Errorf("unexpected token %v after = in key=value option", tok)
	}
}

// parseIdentifier parses a single identifier.
func (d *SnowflakeDialect) parseIdentifier(parser dialects.ParserAccessor) (*ast.Ident, error) {
	tok := parser.NextToken()
	if word, ok := tok.Token.(token.TokenWord); ok {
		return &ast.Ident{Value: word.Word.Value}, nil
	}
	return nil, fmt.Errorf("expected identifier, got %v", tok)
}

// parseSelectItemsForDataLoad parses select items for data loading transformations.
func (d *SnowflakeDialect) parseSelectItemsForDataLoad(parser dialects.ParserAccessor) ([]*expr.StageLoadSelectItemWrapper, error) {
	var items []*expr.StageLoadSelectItemWrapper

	for {
		item, err := d.tryParseStageLoadSelectItem(parser)
		if err != nil {
			return nil, err
		}
		if item != nil {
			items = append(items, &expr.StageLoadSelectItemWrapper{
				Kind: expr.StageLoadSelectItemKindStageLoad,
				Item: item,
			})
		} else {
			// Fall back to regular select item - for now just skip
			// In a full implementation, we'd call parser.ParseSelectItem()
			break
		}

		if !parser.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return items, nil
}

// tryParseStageLoadSelectItem tries to parse a Snowflake-specific stage load select item.
// Format: [<alias>.]$<file_col_num>[:<element>] [AS <alias>]
func (d *SnowflakeDialect) tryParseStageLoadSelectItem(parser dialects.ParserAccessor) (*expr.StageLoadSelectItem, error) {
	item := &expr.StageLoadSelectItem{}

	// Check for optional alias prefix
	tok := parser.PeekTokenRef()
	if word, ok := tok.Token.(token.TokenWord); ok {
		// Look ahead for .
		nextTok := parser.PeekNthToken(1)
		if _, ok := nextTok.Token.(token.TokenPeriod); ok {
			item.Alias = &ast.Ident{Value: word.Word.Value}
			parser.AdvanceToken() // consume alias
			parser.AdvanceToken() // consume .
		}
	}

	// Expect $column_number
	nextTok := parser.NextToken()
	if char, ok := nextTok.Token.(token.TokenChar); !ok || char.Char != '$' {
		// Not a stage load item, go back
		parser.PrevToken()
		if item.Alias != nil {
			parser.PrevToken() // also go back over alias.
		}
		return nil, nil
	}

	// Parse column number
	numTok := parser.NextToken()
	if num, ok := numTok.Token.(token.TokenNumber); ok {
		// Try to parse as integer
		val := 0
		fmt.Sscanf(num.Value, "%d", &val)
		item.FileColNum = int32(val)
	} else {
		return nil, fmt.Errorf("expected column number after $")
	}

	// Check for optional :element
	if parser.ConsumeToken(token.TokenChar{Char: ':'}) {
		elemTok := parser.NextToken()
		if word, ok := elemTok.Token.(token.TokenWord); ok {
			item.Element = &ast.Ident{Value: word.Word.Value}
		} else {
			return nil, fmt.Errorf("expected element name after :")
		}
	}

	// Check for optional AS alias
	if parser.PeekKeyword("AS") {
		parser.AdvanceToken() // consume AS
		alias, err := d.parseIdentifier(parser)
		if err != nil {
			return nil, err
		}
		item.ItemAs = alias
	}

	return item, nil
}

// ParseColumnOption parses Snowflake-specific column options.
// Supports: [WITH] MASKING POLICY, [WITH] PROJECTION POLICY, [WITH] TAG
func (d *SnowflakeDialect) ParseColumnOption(parser dialects.ParserAccessor) (dialects.ColumnOption, bool, error) {
	// Try to parse WITH prefix (optional for most options)
	with := parser.ParseKeyword("WITH")

	// Try MASKING POLICY
	if parser.ParseKeywords([]string{"MASKING", "POLICY"}) {
		policy, err := parseColumnPolicyProperty(parser, with)
		if err != nil {
			return nil, true, err
		}
		return policy, true, nil
	}

	// Try PROJECTION POLICY
	if parser.ParseKeywords([]string{"PROJECTION", "POLICY"}) {
		policy, err := parseColumnPolicyProperty(parser, with)
		if err != nil {
			return nil, true, err
		}
		return policy, true, nil
	}

	// Try TAG
	if parser.ParseKeyword("TAG") {
		tagOpt, err := parseColumnTags(parser, with)
		if err != nil {
			return nil, true, err
		}
		return tagOpt, true, nil
	}

	// If we consumed WITH but didn't find a recognized option, put it back
	if with {
		parser.PrevToken()
	}

	return nil, false, nil
}

// parseColumnPolicyProperty parses a policy property (policy name and optional USING clause).
func parseColumnPolicyProperty(parser dialects.ParserAccessor, with bool) (*expr.ColumnPolicy, error) {
	// Parse policy name - can be multi-part (schema.policy)
	policyName, err := parseObjectName(parser)
	if err != nil {
		return nil, err
	}

	// Check for optional USING (col1, col2, ...) clause
	var usingCols []*ast.Ident
	if parser.ParseKeyword("USING") {
		cols, err := parseParenthesizedIdentifierList(parser)
		if err != nil {
			return nil, err
		}
		usingCols = cols
	}

	return &expr.ColumnPolicy{
		With:       with,
		PolicyName: policyName,
		UsingCols:  usingCols,
	}, nil
}

// parseColumnTags parses TAG (tag_name = 'value', ...).
// Tag names can be multi-part: foo.bar.baz.pii='email'
func parseColumnTags(parser dialects.ParserAccessor, with bool) (*expr.TagsColumnOption, error) {
	// Expect opening parenthesis
	if _, err := parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	var tags []*expr.SnowflakeTag
	for {
		// Parse tag name (can be multi-part with dots)
		tagName, err := parseObjectName(parser)
		if err != nil {
			return nil, err
		}

		// Expect =
		if _, err := parser.ExpectToken(token.TokenEq{}); err != nil {
			return nil, err
		}

		// Parse tag value (string literal or expression)
		tagValue, err := parser.ParseExpression()
		if err != nil {
			return nil, err
		}

		// Convert ast.Expr to expr.Expr
		var exprValue expr.Expr
		if tagValue != nil {
			exprValue = astExprToExpr(tagValue)
		}

		tags = append(tags, &expr.SnowflakeTag{
			Name:  &ast.Ident{Value: tagName.String()},
			Value: exprValue,
		})

		// Check for comma or closing parenthesis
		if parser.ConsumeToken(token.TokenComma{}) {
			continue
		}
		if _, err := parser.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		break
	}

	return &expr.TagsColumnOption{
		With: with,
		Tags: tags,
	}, nil
}

// parseIdent parses a single identifier.
func parseIdent(parser dialects.ParserAccessor) (*ast.Ident, error) {
	tok := parser.NextToken()
	if word, ok := tok.Token.(token.TokenWord); ok {
		return &ast.Ident{Value: word.Word.Value}, nil
	}
	return nil, fmt.Errorf("expected identifier, got %v", tok)
}

// parseObjectName parses a potentially multi-part object name.
func parseObjectName(parser dialects.ParserAccessor) (*ast.ObjectName, error) {
	parts := []ast.ObjectNamePart{}

	// Parse first identifier
	first, err := parseIdent(parser)
	if err != nil {
		return nil, err
	}
	parts = append(parts, &ast.ObjectNamePartIdentifier{Ident: first})

	// Parse additional .identifier parts
	for {
		if !parser.ConsumeToken(token.TokenPeriod{}) {
			break
		}
		next, err := parseIdent(parser)
		if err != nil {
			return nil, err
		}
		parts = append(parts, &ast.ObjectNamePartIdentifier{Ident: next})
	}

	return &ast.ObjectName{Parts: parts}, nil
}

// parseParenthesizedIdentifierList parses (col1, col2, ...).
func parseParenthesizedIdentifierList(parser dialects.ParserAccessor) ([]*ast.Ident, error) {
	// Expect opening parenthesis
	if _, err := parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	var idents []*ast.Ident
	for {
		ident, err := parseIdent(parser)
		if err != nil {
			return nil, err
		}
		idents = append(idents, ident)

		if parser.ConsumeToken(token.TokenComma{}) {
			continue
		}
		if _, err := parser.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		break
	}

	return idents, nil
}

// astExprToExpr converts ast.Expr to expr.Expr.
// Since ast.Expr and expr.Expr are different interfaces, we wrap the string representation.
func astExprToExpr(e ast.Expr) expr.Expr {
	if e == nil {
		return nil
	}
	// Wrap the string representation in a ValueExpr
	return &expr.ValueExpr{Value: e.String()}
}

// parseMultiTableInsert parses a Snowflake multi-table INSERT statement.
// Syntax:
//   - INSERT [OVERWRITE] ALL intoClause [ ... ] subquery
//   - INSERT [OVERWRITE] {FIRST | ALL} {WHEN condition THEN intoClause [ ... ]} [ELSE intoClause] subquery
//
// Reference: src/dialect/snowflake.rs:parse_multi_table_insert
func (d *SnowflakeDialect) parseMultiTableInsert(
	parser dialects.ParserAccessor,
	overwrite bool,
	multiTableType *expr.MultiTableInsertType,
) (ast.Statement, error) {
	// Check if this is conditional (has WHEN clauses) or unconditional (direct INTO clauses)
	isConditional := parser.PeekKeyword("WHEN")

	var intoClauses []*expr.MultiTableInsertIntoClause
	var whenClauses []*expr.MultiTableInsertWhenClause
	var elseClause []*expr.MultiTableInsertIntoClause

	if isConditional {
		// Conditional multi-table insert: parse WHEN clauses
		var err error
		whenClauses, elseClause, err = d.parseMultiTableInsertWhenClauses(parser)
		if err != nil {
			return nil, err
		}
	} else {
		// Unconditional multi-table insert: parse direct INTO clauses
		var err error
		intoClauses, err = d.parseMultiTableInsertIntoClauses(parser)
		if err != nil {
			return nil, err
		}
	}

	// Parse the source query
	source, err := parser.ParseQuery()
	if err != nil {
		return nil, err
	}

	// Extract the query from the statement
	sourceQuery := parser.ExtractQuery(source)

	return &statement.Insert{
		Overwrite:             overwrite,
		MultiTableInsertType:  multiTableType,
		MultiTableIntoClauses: intoClauses,
		MultiTableWhenClauses: whenClauses,
		MultiTableElseClause:  elseClause,
		Source:                sourceQuery,
	}, nil
}

// parseMultiTableInsertIntoClauses parses one or more INTO clauses for multi-table INSERT.
func (d *SnowflakeDialect) parseMultiTableInsertIntoClauses(parser dialects.ParserAccessor) ([]*expr.MultiTableInsertIntoClause, error) {
	var clauses []*expr.MultiTableInsertIntoClause

	for parser.ParseKeyword("INTO") {
		clause, err := d.parseMultiTableInsertIntoClause(parser)
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, clause)
	}

	if len(clauses) == 0 {
		return nil, fmt.Errorf("expected INTO clause in multi-table INSERT")
	}

	return clauses, nil
}

// parseMultiTableInsertIntoClause parses a single INTO clause for multi-table INSERT.
// Syntax: INTO table_name [(columns)] [VALUES (values)]
func (d *SnowflakeDialect) parseMultiTableInsertIntoClause(parser dialects.ParserAccessor) (*expr.MultiTableInsertIntoClause, error) {
	// Parse table name as identifier
	tableIdent, err := d.parseIdentifier(parser)
	if err != nil {
		return nil, err
	}

	// Build expr.ObjectName from identifier
	exprTableName := &expr.ObjectName{
		Parts: []*expr.ObjectNamePart{
			{Ident: &expr.Ident{Value: tableIdent.Value}},
		},
	}

	clause := &expr.MultiTableInsertIntoClause{
		TableName: exprTableName,
	}

	// Parse optional column list: (col1, col2, ...)
	if _, ok := parser.PeekTokenRef().Token.(token.TokenLParen); ok {
		parser.AdvanceToken() // consume (
		for {
			if _, ok := parser.PeekTokenRef().Token.(token.TokenRParen); ok {
				parser.AdvanceToken() // consume )
				break
			}
			col, err := d.parseIdentifier(parser)
			if err != nil {
				return nil, err
			}
			exprCol := &expr.Ident{Value: col.Value}
			clause.Columns = append(clause.Columns, exprCol)
			if !parser.ConsumeToken(token.TokenComma{}) {
				if _, err := parser.ExpectToken(token.TokenRParen{}); err != nil {
					return nil, err
				}
				break
			}
		}
	}

	// Parse optional VALUES clause
	if parser.ParseKeyword("VALUES") {
		if _, err := parser.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}

		values := &expr.MultiTableInsertValues{}
		for {
			if _, ok := parser.PeekTokenRef().Token.(token.TokenRParen); ok {
				parser.AdvanceToken() // consume )
				break
			}

			var value expr.MultiTableInsertValue
			if parser.ParseKeyword("DEFAULT") {
				value.IsDefault = true
			} else {
				exprVal, err := parser.ParseExpression()
				if err != nil {
					return nil, err
				}
				value.Expr = exprVal // No type assertion needed - field is interface{}
			}
			values.Values = append(values.Values, value)

			if !parser.ConsumeToken(token.TokenComma{}) {
				if _, err := parser.ExpectToken(token.TokenRParen{}); err != nil {
					return nil, err
				}
				break
			}
		}
		clause.Values = values
	}

	return clause, nil
}

// parseMultiTableInsertWhenClauses parses WHEN clauses for conditional multi-table INSERT.
func (d *SnowflakeDialect) parseMultiTableInsertWhenClauses(parser dialects.ParserAccessor) ([]*expr.MultiTableInsertWhenClause, []*expr.MultiTableInsertIntoClause, error) {
	var whenClauses []*expr.MultiTableInsertWhenClause
	var elseClause []*expr.MultiTableInsertIntoClause

	// Parse WHEN clauses
	for parser.ParseKeyword("WHEN") {
		// Use the full expression parser to handle operators in the condition
		// We can't use parser.ParseExpression() because it's a simplified version
		// that doesn't handle operators like >, <, =, etc.
		ep := parser.NewExpressionParser()
		condition, err := ep.ParseExprInterface()
		if err != nil {
			return nil, nil, err
		}

		if _, err := parser.ExpectKeyword("THEN"); err != nil {
			return nil, nil, err
		}

		intoClauses, err := d.parseMultiTableInsertIntoClauses(parser)
		if err != nil {
			return nil, nil, err
		}

		whenClause := &expr.MultiTableInsertWhenClause{
			Condition:   condition, // No type assertion needed - field is interface{}
			IntoClauses: intoClauses,
		}
		whenClauses = append(whenClauses, whenClause)
	}

	if len(whenClauses) == 0 {
		return nil, nil, fmt.Errorf("expected at least one WHEN clause in conditional multi-table INSERT")
	}

	// Parse optional ELSE clause
	if parser.ParseKeyword("ELSE") {
		var err error
		elseClause, err = d.parseMultiTableInsertIntoClauses(parser)
		if err != nil {
			return nil, nil, err
		}
	}

	return whenClauses, elseClause, nil
}
