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

package dialects

import (
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/token"
	"github.com/user/sqlparser/tokenizer"
)

// ParserAccessor defines the interface that the parser must implement
// to work with dialects. This breaks the circular dependency between
// the parser and dialects packages.
type ParserAccessor interface {
	// Token access methods
	PeekToken() tokenizer.TokenWithSpan
	PeekTokenRef() *tokenizer.TokenWithSpan
	PeekNthToken(n int) tokenizer.TokenWithSpan
	PeekNthTokenRef(n int) *tokenizer.TokenWithSpan
	NextToken() tokenizer.TokenWithSpan
	NextTokenNoSkip() *tokenizer.TokenWithSpan
	AdvanceToken()
	PrevToken()
	GetCurrentToken() *tokenizer.TokenWithSpan
	GetCurrentIndex() int

	// Token consumption
	ConsumeToken(expected tokenizer.Token) bool
	ExpectToken(expected tokenizer.Token) (tokenizer.TokenWithSpan, error)

	// Keyword helpers
	ParseKeyword(expected string) bool
	PeekKeyword(expected string) bool
	PeekNthKeyword(n int, expected string) bool
	ParseOneOfKeywords(keywords []string) string
	PeekOneOfKeywords(keywords []string) string
	ExpectKeyword(expected string) (tokenizer.TokenWithSpan, error)
	ExpectKeywords(expected []string) error

	// Expression parsing
	ParseExpression() (ast.Expr, error)

	// Statement parsing
	ParseInsert() (ast.Statement, error)

	// Comma-separated parsing
	ParseCommaSeparated(parseFn func() error) error

	// State management
	GetState() ParserState
	SetState(state ParserState)
	InColumnDefinitionState() bool

	// Dialect access
	GetDialect() Dialect
	GetOptions() ParserOptions
}

// ParserState represents the current state of the parser.
// This is a copy of parser.ParserState to avoid importing the parser package.
type ParserState int

const (
	// StateNormal is the default state of the parser.
	StateNormal ParserState = iota
	// StateConnectBy is the state when parsing a CONNECT BY expression.
	StateConnectBy
	// StateColumnDefinition is the state when parsing column definitions.
	StateColumnDefinition
)

// ParserOptions represents parser options.
// This is a copy of parser.ParserOptions to avoid importing the parser package.
type ParserOptions struct {
	TrailingCommas   bool
	Unescape         bool
	RequireSemicolon bool
}

// Precedence defines the precedence levels for SQL operators.
// Higher values mean higher precedence (tighter binding).
type Precedence int

const (
	// PrecedencePeriod - Member access operator '.' (highest precedence)
	PrecedencePeriod Precedence = 100
	// PrecedenceDoubleColon - Postgres style type cast '::'
	PrecedenceDoubleColon Precedence = 50
	// PrecedenceAtTz - Timezone operator (e.g., 'AT TIME ZONE')
	PrecedenceAtTz Precedence = 41
	// PrecedenceMulDivModOp - Multiplication/Division/Modulo operators
	PrecedenceMulDivModOp Precedence = 40
	// PrecedencePlusMinus - Addition/Subtraction
	PrecedencePlusMinus Precedence = 30
	// PrecedenceXor - Bitwise XOR operator
	PrecedenceXor Precedence = 24
	// PrecedenceAmpersand - Bitwise AND operator
	PrecedenceAmpersand Precedence = 23
	// PrecedenceCaret - Bitwise CARET for some dialects
	PrecedenceCaret Precedence = 22
	// PrecedencePipe - Bitwise OR / pipe operator
	PrecedencePipe Precedence = 21
	// PrecedenceColon - ':' operator for json/variant access
	PrecedenceColon Precedence = 21
	// PrecedenceBetween - BETWEEN operator
	PrecedenceBetween Precedence = 20
	// PrecedenceEq - Equality operator
	PrecedenceEq Precedence = 20
	// PrecedenceLike - Pattern matching (LIKE)
	PrecedenceLike Precedence = 19
	// PrecedenceIs - IS operator (e.g., 'IS NULL')
	PrecedenceIs Precedence = 17
	// PrecedencePgOther - Other Postgres-specific operators
	PrecedencePgOther Precedence = 16
	// PrecedenceUnaryNot - Unary NOT
	PrecedenceUnaryNot Precedence = 15
	// PrecedenceAnd - Logical AND
	PrecedenceAnd Precedence = 10
	// PrecedenceOr - Logical OR (lowest precedence)
	PrecedenceOr Precedence = 5
)

// NestedIdentifierQuote represents a nested identifier with outer and optional inner quotes
type NestedIdentifierQuote struct {
	OuterQuote rune
	InnerQuote *rune
}

// GranteesType represents grantee types that can be reserved
type GranteesType int

const (
	// GranteeTypeNone represents no special grantee type
	GranteeTypeNone GranteesType = iota
	// GranteeTypeRole represents a role grantee
	GranteeTypeRole
	// GranteeTypeUser represents a user grantee
	GranteeTypeUser
)

// ColumnOption represents a column option in CREATE TABLE
type ColumnOption interface {
	IsColumnOption()
}

// Dialect encapsulates the differences between SQL implementations.
// SQL implementations deviate from one another, either due to custom extensions
// or various historical reasons. This interface encapsulates the parsing differences
// between dialects.
type Dialect interface {
	// ============================================================================
	// Dialect Identification
	// ============================================================================

	// Dialect returns the dialect name/identifier for identification purposes.
	Dialect() string

	// ============================================================================
	// Identifier Handling
	// ============================================================================

	// IsIdentifierStart returns true if the character is a valid start character
	// for an unquoted identifier.
	IsIdentifierStart(ch rune) bool

	// IsIdentifierPart returns true if the character is a valid unquoted
	// identifier character (not necessarily at the start).
	IsIdentifierPart(ch rune) bool

	// IsDelimitedIdentifierStart returns true if the character starts a
	// quoted/delimited identifier. The default implementation accepts "double quoted"
	// and backtick-quoted identifiers, which is ANSI-compliant and appropriate for
	// most dialects (with the notable exception of MySQL, MS SQL, and SQLite).
	IsDelimitedIdentifierStart(ch rune) bool

	// IsNestedDelimitedIdentifierStart returns true if the character starts a
	// potential nested quoted identifier.
	IsNestedDelimitedIdentifierStart(ch rune) bool

	// PeekNestedDelimitedIdentifierQuotes returns the outer quote style and
	// optional inner (nested) quote style if the next sequence of tokens
	// potentially represents a nested identifier.
	PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*NestedIdentifierQuote, int)

	// IdentifierQuoteStyle returns the character used to quote identifiers for
	// the given identifier string. Returns nil if no quoting is needed.
	IdentifierQuoteStyle(identifier string) *rune

	// IsCustomOperatorPart returns true if the character is part of a custom operator.
	IsCustomOperatorPart(ch rune) bool

	// ============================================================================
	// Reserved Keywords
	// ============================================================================

	// IsReservedForIdentifier returns true if the specified keyword is reserved
	// and cannot be used as an identifier without special handling like quoting.
	IsReservedForIdentifier(kw token.Keyword) bool

	// GetReservedKeywordsForSelectItemOperator returns reserved keywords that may
	// prefix a select item expression (e.g., CONNECT_BY_ROOT in Snowflake).
	GetReservedKeywordsForSelectItemOperator() []token.Keyword

	// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
	GetReservedGranteesTypes() []GranteesType

	// IsColumnAlias returns true if the specified keyword should be parsed as a column alias.
	IsColumnAlias(kw token.Keyword, parser ParserAccessor) bool

	// IsSelectItemAlias returns true if the specified keyword should be parsed as
	// a select item alias. When explicit is true, the keyword is preceded by an AS word.
	IsSelectItemAlias(explicit bool, kw token.Keyword, parser ParserAccessor) bool

	// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier.
	IsTableFactor(kw token.Keyword, parser ParserAccessor) bool

	// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias.
	IsTableAlias(kw token.Keyword, parser ParserAccessor) bool

	// IsTableFactorAlias returns true if the specified keyword should be parsed as
	// a table factor alias. When explicit is true, the keyword is preceded by an AS word.
	IsTableFactorAlias(explicit bool, kw token.Keyword, parser ParserAccessor) bool

	// IsIdentifierGeneratingFunctionName returns true if the dialect considers
	// the specified ident as a function that returns an identifier.
	IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool

	// ============================================================================
	// String Literals
	// ============================================================================

	// SupportsStringLiteralBackslashEscape returns true if the dialect supports
	// escaping characters via '\' in string literals.
	SupportsStringLiteralBackslashEscape() bool

	// IgnoresWildcardEscapes returns true if the dialect strips the backslash
	// when escaping LIKE wildcards (%, _).
	IgnoresWildcardEscapes() bool

	// SupportsUnicodeStringLiteral returns true if the dialect supports string
	// literals with U& prefix for Unicode code points.
	SupportsUnicodeStringLiteral() bool

	// SupportsTripleQuotedString returns true if the dialect supports triple
	// quoted string literals (e.g., """abc""").
	SupportsTripleQuotedString() bool

	// SupportsStringLiteralConcatenation returns true if the dialect supports
	// concatenating string literals (e.g., 'Hello ' 'world').
	SupportsStringLiteralConcatenation() bool

	// SupportsStringLiteralConcatenationWithNewline returns true if the dialect
	// supports concatenating string literals with a newline between them.
	SupportsStringLiteralConcatenationWithNewline() bool

	// SupportsQuoteDelimitedString returns true if the dialect supports
	// quote-delimited string literals (e.g., Q'{...}') for Oracle-style strings.
	SupportsQuoteDelimitedString() bool

	// SupportsStringEscapeConstant returns true if the dialect supports the
	// E'...' syntax for string literals with escape sequences (Postgres).
	SupportsStringEscapeConstant() bool

	// ============================================================================
	// Aggregations and Window Functions
	// ============================================================================

	// SupportsFilterDuringAggregation returns true if the dialect supports
	// FILTER (WHERE expr) for aggregate queries.
	SupportsFilterDuringAggregation() bool

	// SupportsWithinAfterArrayAggregation returns true if the dialect supports
	// ARRAY_AGG() [WITHIN GROUP (ORDER BY)] expressions.
	SupportsWithinAfterArrayAggregation() bool

	// SupportsWindowClauseNamedWindowReference returns true if the dialect
	// supports referencing another named window within a window clause declaration.
	SupportsWindowClauseNamedWindowReference() bool

	// SupportsWindowFunctionNullTreatmentArg returns true if the dialect supports
	// specifying null treatment as part of a window function's parameter list.
	SupportsWindowFunctionNullTreatmentArg() bool

	// SupportsMatchRecognize returns true if the dialect supports the MATCH_RECOGNIZE operation.
	SupportsMatchRecognize() bool

	// ============================================================================
	// GROUP BY Support
	// ============================================================================

	// SupportsGroupByExpr returns true if the dialect supports GROUP BY expressions
	// like GROUPING SETS, ROLLUP, or CUBE.
	SupportsGroupByExpr() bool

	// SupportsGroupByWithModifier returns true if the dialect supports GROUP BY
	// modifiers prefixed by a WITH keyword (e.g., GROUP BY value WITH ROLLUP).
	SupportsGroupByWithModifier() bool

	// ============================================================================
	// JOIN Support
	// ============================================================================

	// SupportsLeftAssociativeJoinsWithoutParens returns true if the dialect supports
	// left-associative join parsing by default when parentheses are omitted in nested joins.
	SupportsLeftAssociativeJoinsWithoutParens() bool

	// SupportsOuterJoinOperator returns true if the dialect supports the (+)
	// syntax for OUTER JOIN (Oracle-style).
	SupportsOuterJoinOperator() bool

	// SupportsCrossJoinConstraint returns true if the dialect supports a join
	// specification on CROSS JOIN.
	SupportsCrossJoinConstraint() bool

	// ============================================================================
	// CONNECT BY Support
	// ============================================================================

	// SupportsConnectBy returns true if the dialect supports CONNECT BY for
	// hierarchical queries (Oracle-style).
	SupportsConnectBy() bool

	// ============================================================================
	// Transaction Support
	// ============================================================================

	// SupportsStartTransactionModifier returns true if the dialect supports
	// BEGIN {DEFERRED | IMMEDIATE | EXCLUSIVE | TRY | CATCH} [TRANSACTION].
	SupportsStartTransactionModifier() bool

	// SupportsEndTransactionModifier returns true if the dialect supports
	// END {TRY | CATCH} statements.
	SupportsEndTransactionModifier() bool

	// ============================================================================
	// Named Function Arguments
	// ============================================================================

	// SupportsNamedFnArgsWithEqOperator returns true if the dialect supports
	// named arguments of the form FUN(a = '1', b = '2').
	SupportsNamedFnArgsWithEqOperator() bool

	// SupportsNamedFnArgsWithColonOperator returns true if the dialect supports
	// named arguments of the form FUN(a : '1', b : '2').
	SupportsNamedFnArgsWithColonOperator() bool

	// SupportsNamedFnArgsWithAssignmentOperator returns true if the dialect supports
	// named arguments of the form FUN(a := '1', b := '2').
	SupportsNamedFnArgsWithAssignmentOperator() bool

	// SupportsNamedFnArgsWithRArrowOperator returns true if the dialect supports
	// named arguments of the form FUN(a => '1', b => '2').
	SupportsNamedFnArgsWithRArrowOperator() bool

	// SupportsNamedFnArgsWithExprName returns true if the dialect supports
	// argument name as arbitrary expression.
	SupportsNamedFnArgsWithExprName() bool

	// ============================================================================
	// SET Statement Support
	// ============================================================================

	// SupportsParenthesizedSetVariables returns true if the dialect supports
	// multiple variable assignment using parentheses in a SET variable declaration.
	SupportsParenthesizedSetVariables() bool

	// SupportsCommaSeparatedSetAssignments returns true if the dialect supports
	// multiple SET statements in a single statement.
	SupportsCommaSeparatedSetAssignments() bool

	// SupportsSetStmtWithoutOperator returns true if the dialect supports
	// SET statements without an explicit assignment operator (e.g., SET SHOWPLAN_XML ON).
	SupportsSetStmtWithoutOperator() bool

	// SupportsSetNames returns true if the dialect supports SET NAMES <charset_name> [COLLATE <collation_name>].
	SupportsSetNames() bool

	// ============================================================================
	// SELECT Clause Support
	// ============================================================================

	// SupportsSelectWildcardExcept returns true if the dialect supports EXCEPT
	// clause following a wildcard in a select list.
	SupportsSelectWildcardExcept() bool

	// SupportsSelectWildcardExclude returns true if the dialect supports EXCLUDE
	// option following a wildcard in a projection section.
	SupportsSelectWildcardExclude() bool

	// SupportsSelectExclude returns true if the dialect supports EXCLUDE as the
	// last item in the projection section, not necessarily after a wildcard.
	SupportsSelectExclude() bool

	// SupportsSelectWildcardReplace returns true if the dialect supports REPLACE
	// option in SELECT * wildcard expressions.
	SupportsSelectWildcardReplace() bool

	// SupportsSelectWildcardIlike returns true if the dialect supports ILIKE
	// option in SELECT * wildcard expressions.
	SupportsSelectWildcardIlike() bool

	// SupportsSelectWildcardRename returns true if the dialect supports RENAME
	// option in SELECT * wildcard expressions.
	SupportsSelectWildcardRename() bool

	// SupportsSelectWildcardWithAlias returns true if the dialect supports aliasing
	// a wildcard select item.
	SupportsSelectWildcardWithAlias() bool

	// SupportsSelectExprStar returns true if the dialect supports wildcard expansion
	// on arbitrary expressions in projections.
	SupportsSelectExprStar() bool

	// SupportsFromFirstSelect returns true if the dialect supports "FROM-first" selects.
	SupportsFromFirstSelect() bool

	// SupportsEmptyProjections returns true if the dialect supports empty projections
	// in SELECT statements.
	SupportsEmptyProjections() bool

	// SupportsSelectModifiers returns true if the dialect supports MySQL-specific
	// SELECT modifiers like HIGH_PRIORITY, STRAIGHT_JOIN, SQL_SMALL_RESULT, etc.
	SupportsSelectModifiers() bool

	// SupportsPipeOperator returns true if the dialect supports the pipe operator (|>).
	SupportsPipeOperator() bool

	// SupportsTrailingCommas returns true if the dialect supports trailing commas.
	SupportsTrailingCommas() bool

	// SupportsProjectionTrailingCommas returns true if the dialect supports trailing
	// commas in the projection list.
	SupportsProjectionTrailingCommas() bool

	// SupportsFromTrailingCommas returns true if the dialect supports trailing commas
	// in the FROM clause.
	SupportsFromTrailingCommas() bool

	// SupportsColumnDefinitionTrailingCommas returns true if the dialect supports trailing
	// commas in column definitions.
	SupportsColumnDefinitionTrailingCommas() bool

	// SupportsLimitComma returns true if the dialect supports parsing LIMIT 1, 2
	// as LIMIT 2 OFFSET 1.
	SupportsLimitComma() bool

	// ============================================================================
	// Type Conversion
	// ============================================================================

	// ConvertTypeBeforeValue returns true if the dialect has a CONVERT function
	// which accepts a type first and an expression second.
	ConvertTypeBeforeValue() bool

	// SupportsTryConvert returns true if the dialect supports the TRY_CONVERT function.
	SupportsTryConvert() bool

	// SupportsBinaryKwAsCast returns true if the dialect supports casting an expression
	// to a binary type using the BINARY <expr> syntax.
	SupportsBinaryKwAsCast() bool

	// ============================================================================
	// Object Names and References
	// ============================================================================

	// SupportsObjectNameDoubleDotNotation returns true if the dialect supports
	// double dot notation for object names (e.g., db_name..table_name).
	SupportsObjectNameDoubleDotNotation() bool

	// SupportsNumericPrefix returns true if the dialect supports identifiers
	// starting with a numeric prefix.
	SupportsNumericPrefix() bool

	// SupportsNumericLiteralUnderscores returns true if the dialect supports
	// numbers containing underscores (e.g., 10_000_000).
	SupportsNumericLiteralUnderscores() bool

	// ============================================================================
	// IN Expressions
	// ============================================================================

	// SupportsInEmptyList returns true if the dialect supports (NOT) IN ()
	// expressions with empty lists.
	SupportsInEmptyList() bool

	// ============================================================================
	// Dictionary and Literal Syntax
	// ============================================================================

	// SupportsDictionarySyntax returns true if the dialect supports defining
	// structs or objects using syntax like {'x': 1, 'y': 2, 'z': 3}.
	SupportsDictionarySyntax() bool

	// SupportsMapLiteralSyntax returns true if the dialect supports defining
	// objects using syntax like Map {1: 10, 2: 20}.
	SupportsMapLiteralSyntax() bool

	// SupportsStructLiteral returns true if the dialect supports STRUCT literal syntax.
	SupportsStructLiteral() bool

	// SupportsArrayLiteralSyntax returns true if the dialect supports array literal syntax.
	SupportsArrayLiteralSyntax() bool

	// SupportsLambdaFunctions returns true if the dialect supports lambda functions.
	SupportsLambdaFunctions() bool

	// ============================================================================
	// Table Definition Support
	// ============================================================================

	// SupportsCreateTableMultiSchemaInfoSources returns true if the dialect supports
	// specifying multiple options in a CREATE TABLE statement.
	SupportsCreateTableMultiSchemaInfoSources() bool

	// SupportsCreateTableLikeParenthesized returns true if the dialect supports
	// specifying which table to copy the schema from inside parenthesis.
	SupportsCreateTableLikeParenthesized() bool

	// SupportsCreateTableSelect returns true if the dialect supports CREATE TABLE SELECT.
	SupportsCreateTableSelect() bool

	// SupportsCreateViewCommentSyntax returns true if the dialect supports the COMMENT
	// clause in CREATE VIEW statements using COMMENT = 'comment' syntax.
	SupportsCreateViewCommentSyntax() bool

	// SupportsArrayTypedefWithoutElementType returns true if the dialect supports
	// ARRAY type without specifying an element type.
	SupportsArrayTypedefWithoutElementType() bool

	// SupportsArrayTypedefWithBrackets returns true if the dialect supports array
	// type definition with brackets with optional size.
	SupportsArrayTypedefWithBrackets() bool

	// SupportsParensAroundTableFactor returns true if the dialect supports extra
	// parentheses around lone table names or derived tables in the FROM clause.
	SupportsParensAroundTableFactor() bool

	// SupportsValuesAsTableFactor returns true if the dialect supports VALUES
	// as a table factor without requiring parentheses.
	SupportsValuesAsTableFactor() bool

	// SupportsSemanticViewTableFactor returns true if the dialect supports
	// SEMANTIC_VIEW() table functions.
	SupportsSemanticViewTableFactor() bool

	// SupportsTableVersioning returns true if the dialect supports querying
	// historical table data by specifying which version to query.
	SupportsTableVersioning() bool

	// SupportsTableSampleBeforeAlias returns true if the dialect supports the
	// TABLESAMPLE option before the table alias option.
	SupportsTableSampleBeforeAlias() bool

	// SupportsTableHints returns true if the dialect supports table hints in the FROM clause.
	SupportsTableHints() bool

	// SupportsIndexHints returns true if the dialect supports index hints (e.g., USE INDEX).
	SupportsIndexHints() bool

	// ============================================================================
	// Column Definition Support
	// ============================================================================

	// SupportsAscDescInColumnDefinition returns true if the dialect supports
	// ASC and DESC in column definitions.
	SupportsAscDescInColumnDefinition() bool

	// SupportsSpaceSeparatedColumnOptions returns true if the dialect supports
	// space-separated column options in CREATE TABLE statements.
	SupportsSpaceSeparatedColumnOptions() bool

	// SupportsConstraintKeywordWithoutName returns true if the dialect supports
	// CONSTRAINT keyword without a name in table constraint definitions.
	SupportsConstraintKeywordWithoutName() bool

	// SupportsKeyColumnOption returns true if the dialect supports the KEY keyword
	// as part of column-level constraints.
	SupportsKeyColumnOption() bool

	// SupportsDataTypeSignedSuffix returns true if the dialect allows an optional
	// SIGNED suffix after integer data types.
	SupportsDataTypeSignedSuffix() bool

	// ============================================================================
	// Comment Support
	// ============================================================================

	// SupportsNestedComments returns true if the dialect supports nested comments.
	SupportsNestedComments() bool

	// SupportsMultilineCommentHints returns true if the dialect supports optimizer
	// hints in multiline comments (e.g., /*!50110 KEY_BLOCK_SIZE = 1024*/).
	SupportsMultilineCommentHints() bool

	// SupportsCommentOptimizerHint returns true if the dialect supports query
	// optimizer hints in single and multi-line comments.
	SupportsCommentOptimizerHint() bool

	// SupportsCommentOn returns true if the dialect supports the COMMENT statement.
	SupportsCommentOn() bool

	// RequiresSingleLineCommentWhitespace returns true if the dialect requires
	// a whitespace character after -- to start a single line comment.
	RequiresSingleLineCommentWhitespace() bool

	// ============================================================================
	// EXPLAIN Support
	// ============================================================================

	// SupportsExplainWithUtilityOptions returns true if the dialect supports
	// EXPLAIN statements with utility options.
	SupportsExplainWithUtilityOptions() bool

	// ============================================================================
	// EXECUTE IMMEDIATE Support
	// ============================================================================

	// SupportsExecuteImmediate returns true if the dialect supports EXECUTE IMMEDIATE statements.
	SupportsExecuteImmediate() bool

	// ============================================================================
	// Extract Function Support
	// ============================================================================

	// AllowExtractCustom returns true if the dialect allows the EXTRACT function
	// to use words other than keywords.
	AllowExtractCustom() bool

	// AllowExtractSingleQuotes returns true if the dialect allows the EXTRACT
	// function to use single quotes in the part being extracted.
	AllowExtractSingleQuotes() bool

	// SupportsExtractCommaSyntax returns true if the dialect supports EXTRACT
	// function with a comma separator instead of FROM.
	SupportsExtractCommaSyntax() bool

	// ============================================================================
	// Subquery Support
	// ============================================================================

	// SupportsSubqueryAsFunctionArg returns true if the dialect supports a subquery
	// passed to a function as the only argument without enclosing parentheses.
	SupportsSubqueryAsFunctionArg() bool

	// ============================================================================
	// Placeholder Support
	// ============================================================================

	// SupportsDollarPlaceholder returns true if the dialect allows dollar placeholders (e.g., $var).
	SupportsDollarPlaceholder() bool

	// ============================================================================
	// Index Support
	// ============================================================================

	// SupportsCreateIndexWithClause returns true if the dialect supports WITH clause
	// in CREATE INDEX statement.
	SupportsCreateIndexWithClause() bool

	// ============================================================================
	// Interval Support
	// ============================================================================

	// RequireIntervalQualifier returns true if the dialect requires units
	// (qualifiers) to be specified in INTERVAL expressions.
	RequireIntervalQualifier() bool

	// SupportsIntervalOptions returns true if the dialect supports INTERVAL data
	// type with Postgres-style options.
	SupportsIntervalOptions() bool

	// ============================================================================
	// Operator Support
	// ============================================================================

	// SupportsFactorialOperator returns true if the dialect supports a!
	// expressions for factorial.
	SupportsFactorialOperator() bool

	// SupportsBitwiseShiftOperators returns true if the dialect supports
	// << and >> shift operators.
	SupportsBitwiseShiftOperators() bool

	// SupportsNotnullOperator returns true if the dialect supports the x NOTNULL
	// operator expression.
	SupportsNotnullOperator() bool

	// SupportsBangNotOperator returns true if the dialect supports !a syntax
	// for boolean NOT expressions.
	SupportsBangNotOperator() bool

	// SupportsDoubleAmpersandOperator returns true if the dialect considers
	// the && operator as a boolean AND operator.
	SupportsDoubleAmpersandOperator() bool

	// ============================================================================
	// MATCH AGAINST Support
	// ============================================================================

	// SupportsMatchAgainst returns true if the dialect supports MATCH() AGAINST() syntax.
	SupportsMatchAgainst() bool

	// ============================================================================
	// Grantee Support
	// ============================================================================

	// SupportsUserHostGrantee returns true if the dialect supports MySQL-style
	// 'user'@'host' grantee syntax.
	SupportsUserHostGrantee() bool

	// ============================================================================
	// LISTEN/NOTIFY Support
	// ============================================================================

	// SupportsListenNotify returns true if the dialect supports LISTEN, UNLISTEN,
	// and NOTIFY statements.
	SupportsListenNotify() bool

	// ============================================================================
	// LOAD Support
	// ============================================================================

	// SupportsLoadData returns true if the dialect supports LOAD DATA statement.
	SupportsLoadData() bool

	// SupportsLoadExtension returns true if the dialect supports LOAD extension statement.
	SupportsLoadExtension() bool

	// ============================================================================
	// TOP/DISTINCT Ordering
	// ============================================================================

	// SupportsTopBeforeDistinct returns true if the dialect expects the TOP option
	// before the ALL/DISTINCT options in a SELECT statement.
	SupportsTopBeforeDistinct() bool

	// ============================================================================
	// Boolean Literals
	// ============================================================================

	// SupportsBooleanLiterals returns true if the dialect supports boolean
	// literals (true and false).
	SupportsBooleanLiterals() bool

	// ============================================================================
	// SHOW Statement Support
	// ============================================================================

	// SupportsShowLikeBeforeIn returns true if the dialect supports the LIKE
	// option in a SHOW statement before the IN option.
	SupportsShowLikeBeforeIn() bool

	// ============================================================================
	// PartiQL Support
	// ============================================================================

	// SupportsPartiQL returns true if the dialect supports PartiQL for querying
	// semi-structured data.
	SupportsPartiQL() bool

	// ============================================================================
	// Alias Assignment
	// ============================================================================

	// SupportsEqAliasAssignment returns true if the dialect supports treating
	// the equals operator = within a SelectItem as an alias assignment operator.
	SupportsEqAliasAssignment() bool

	// ============================================================================
	// INSERT Statement Support
	// ============================================================================

	// SupportsInsertSet returns true if the dialect supports INSERT INTO ... SET syntax.
	SupportsInsertSet() bool

	// SupportsInsertTableFunction returns true if the dialect supports table
	// function in insertion.
	SupportsInsertTableFunction() bool

	// SupportsInsertTableQuery returns true if the dialect supports table queries
	// in insertion (e.g., SELECT INTO (<query>) ...).
	SupportsInsertTableQuery() bool

	// SupportsInsertFormat returns true if the dialect supports insert formats
	// (e.g., INSERT INTO ... FORMAT <format>).
	SupportsInsertFormat() bool

	// SupportsInsertTableAlias returns true if the dialect supports INSERT INTO
	// table [[AS] alias] syntax.
	SupportsInsertTableAlias() bool

	// ============================================================================
	// ALTER TABLE Support
	// ============================================================================

	// SupportsAlterColumnTypeUsing returns true if the dialect supports the USING
	// clause in an ALTER COLUMN statement.
	SupportsAlterColumnTypeUsing() bool

	// SupportsCommaSeparatedDropColumnList returns true if the dialect supports
	// ALTER TABLE tbl DROP COLUMN c1, ..., cn.
	SupportsCommaSeparatedDropColumnList() bool

	// ============================================================================
	// ORDER BY Support
	// ============================================================================

	// SupportsOrderByAll returns true if the dialect supports ORDER BY ALL.
	SupportsOrderByAll() bool

	// ============================================================================
	// Geometric Types Support
	// ============================================================================

	// SupportsGeometricTypes returns true if the dialect supports geometric types
	// (Postgres geometric operations).
	SupportsGeometricTypes() bool

	// ============================================================================
	// DESCRIBE Support
	// ============================================================================

	// DescribeRequiresTableKeyword returns true if the dialect requires the TABLE
	// keyword after DESCRIBE.
	DescribeRequiresTableKeyword() bool

	// ============================================================================
	// ClickHouse-specific Support
	// ============================================================================

	// SupportsOptimizeTable returns true if the dialect supports OPTIMIZE TABLE.
	SupportsOptimizeTable() bool

	// SupportsPrewhere returns true if the dialect supports PREWHERE clause.
	SupportsPrewhere() bool

	// SupportsWithFill returns true if the dialect supports WITH FILL clause.
	SupportsWithFill() bool

	// SupportsLimitBy returns true if the dialect supports LIMIT BY clause.
	SupportsLimitBy() bool

	// SupportsInterpolate returns true if the dialect supports INTERPOLATE clause.
	SupportsInterpolate() bool

	// SupportsSettings returns true if the dialect supports SETTINGS clause.
	SupportsSettings() bool

	// SupportsSelectFormat returns true if the dialect supports FORMAT clause in SELECT.
	SupportsSelectFormat() bool

	// ============================================================================
	// DuckDB-specific Support
	// ============================================================================

	// SupportsInstall returns true if the dialect supports INSTALL statement.
	SupportsInstall() bool

	// SupportsDetach returns true if the dialect supports DETACH statement.
	SupportsDetach() bool

	// ============================================================================
	// TRIM Support
	// ============================================================================

	// SupportsCommaSeparatedTrim returns true if the dialect supports the two-argument
	// comma-separated form of TRIM function: TRIM(expr, characters).
	SupportsCommaSeparatedTrim() bool

	// ============================================================================
	// Precedence Handling
	// ============================================================================

	// PrecValue returns the precedence value for a given Precedence level.
	PrecValue(prec Precedence) uint8

	// PrecUnknown returns the precedence when precedence is otherwise unknown.
	PrecUnknown() uint8

	// GetNextPrecedence returns the dialect-specific precedence override for
	// the next token. If returns 0, falls back to default behavior.
	GetNextPrecedence(parser ParserAccessor) (uint8, error)

	// GetNextPrecedenceDefault implements the default precedence logic.
	GetNextPrecedenceDefault(parser ParserAccessor) (uint8, error)

	// ============================================================================
	// Custom Parsing Hooks
	// ============================================================================

	// ParsePrefix is a dialect-specific prefix parser override.
	// If second return value is false, falls back to default behavior.
	ParsePrefix(parser ParserAccessor) (ast.Expr, bool, error)

	// ParseInfix is a dialect-specific infix parser override.
	// If second return value is false, falls back to default behavior.
	ParseInfix(parser ParserAccessor, expr ast.Expr, precedence uint8) (ast.Expr, bool, error)

	// ParseStatement is a dialect-specific statement parser override.
	// If second return value is false, falls back to default behavior.
	ParseStatement(parser ParserAccessor) (ast.Statement, bool, error)

	// ParseColumnOption is a dialect-specific column option parser override.
	// If second return value is false, falls back to default behavior.
	ParseColumnOption(parser ParserAccessor) (ColumnOption, bool, error)
}
