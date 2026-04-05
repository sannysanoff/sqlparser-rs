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

// Package parseriface defines shared interfaces for the SQL parser.
//
// This package exists to break circular dependencies between the parser
// and dialects packages. It provides a common set of interfaces that
// both packages can depend on without creating import cycles.
//
// The interfaces defined here are intentionally minimal and focused
// on the contract needed for dialect-specific parsing behavior.
package parseriface

import (
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/datatype"
	"github.com/user/sqlparser/token"
)

// Parser defines the interface that the parser must implement
// to work with dialects. This breaks the circular dependency between
// the parser and dialects packages.
//
// This interface consolidates the previous ParserInterface (from parser/core.go)
// and ParserAccessor (from dialects/dialect.go) into a single, unified interface.
type Parser interface {
	// Token access methods
	PeekToken() token.TokenWithSpan
	PeekTokenRef() *token.TokenWithSpan
	PeekNthToken(n int) token.TokenWithSpan
	PeekNthTokenRef(n int) *token.TokenWithSpan
	PeekTokenNoSkip() token.TokenWithSpan
	PeekNthTokenNoSkip(n int) token.TokenWithSpan
	NextToken() token.TokenWithSpan
	NextTokenNoSkip() *token.TokenWithSpan
	AdvanceToken()
	PrevToken()
	GetCurrentToken() *token.TokenWithSpan
	GetCurrentIndex() int

	// Token consumption
	ConsumeToken(expected token.Token) bool
	ExpectToken(expected token.Token) (token.TokenWithSpan, error)

	// Keyword helpers
	ParseKeyword(expected string) bool
	PeekKeyword(expected string) bool
	PeekNthKeyword(n int, expected string) bool
	ParseOneOfKeywords(keywords []string) string
	PeekOneOfKeywords(keywords []string) string
	ParseKeywords(keywords []string) bool
	ExpectKeyword(expected string) (token.TokenWithSpan, error)
	ExpectKeywords(expected []string) error

	// Comma-separated parsing
	ParseCommaSeparated(parseFn func() error) error

	// Expression and statement parsing
	ParseExpression() (ast.Expr, error)
	ParseInsert() (ast.Statement, error)
	ParseQuery() (ast.Statement, error)
	ParseDataType() (datatype.DataType, error)

	// State management
	GetState() ParserState
	SetState(state ParserState)
	InColumnDefinitionState() bool

	// Dialect access - returns CompleteDialect for access to all capabilities
	// This will be set to dialects.CompleteDialect by the parser package
	GetDialect() CompleteDialect

	// Options access
	GetOptions() ParserOptions

	// Utility methods
	Expected(expected string, found token.TokenWithSpan) error
	ExpectedRef(expected string, found *token.TokenWithSpan) error
	ExpectedAt(expected string, index int) error
	SavePosition() func()
}

// ParserState represents the current state of the parser.
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
	// PrecedenceCollate - COLLATE operator (between AT TIME ZONE and ::)
	PrecedenceCollate Precedence = 42
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
// This is a minimal interface that defines the core dialect identifier.
// For capability-specific methods, use the sub-interfaces defined in the
// dialects package.
type Dialect interface {
	// Dialect returns the dialect name/identifier for identification purposes.
	Dialect() string

	// PrecValue returns the precedence value for a given Precedence level.
	PrecValue(prec Precedence) uint8

	// PrecUnknown returns the precedence when precedence is otherwise unknown.
	PrecUnknown() uint8

	// GetNextPrecedence returns the dialect-specific precedence override for
	// the next token. If returns 0, falls back to default behavior.
	GetNextPrecedence(parser Parser) (uint8, error)

	// GetNextPrecedenceDefault implements the default precedence logic.
	GetNextPrecedenceDefault(parser Parser) (uint8, error)

	// ParsePrefix is a dialect-specific prefix parser override.
	// If second return value is false, falls back to default behavior.
	ParsePrefix(parser Parser) (ast.Expr, bool, error)

	// ParseInfix is a dialect-specific infix parser override.
	// If second return value is false, falls back to default behavior.
	ParseInfix(parser Parser, expr ast.Expr, precedence uint8) (ast.Expr, bool, error)

	// ParseStatement is a dialect-specific statement parser override.
	// If second return value is false, falls back to default behavior.
	ParseStatement(parser Parser) (ast.Statement, bool, error)

	// ParseColumnOption is a dialect-specific column option parser override.
	// If second return value is false, falls back to default behavior.
	ParseColumnOption(parser Parser) (ColumnOption, bool, error)
}

// CompleteDialect is the full dialect interface with all capability methods.
// This is the interface that the parser works with when it needs to check
// dialect capabilities.
type CompleteDialect interface {
	Dialect

	// Identifier handling
	IsIdentifierStart(ch rune) bool
	IsIdentifierPart(ch rune) bool
	IsDelimitedIdentifierStart(ch rune) bool
	IsNestedDelimitedIdentifierStart(ch rune) bool
	PeekNestedDelimitedIdentifierQuotes(chars []rune, pos int) (*NestedIdentifierQuote, int)
	IdentifierQuoteStyle(identifier string) *rune
	IsCustomOperatorPart(ch rune) bool
	IsReservedForIdentifier(kw token.Keyword) bool
	IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool

	// String literals
	SupportsStringLiteralBackslashEscape() bool
	IgnoresWildcardEscapes() bool
	SupportsUnicodeStringLiteral() bool
	SupportsTripleQuotedString() bool
	SupportsStringLiteralConcatenation() bool
	SupportsStringLiteralConcatenationWithNewline() bool
	SupportsQuoteDelimitedString() bool
	SupportsStringEscapeConstant() bool

	// Aggregations and window functions
	SupportsFilterDuringAggregation() bool
	SupportsWithinAfterArrayAggregation() bool
	SupportsWindowClauseNamedWindowReference() bool
	SupportsWindowFunctionNullTreatmentArg() bool
	SupportsMatchRecognize() bool

	// GROUP BY support
	SupportsGroupByExpr() bool
	SupportsGroupByWithModifier() bool

	// JOIN support
	SupportsLeftAssociativeJoinsWithoutParens() bool
	SupportsOuterJoinOperator() bool
	SupportsCrossJoinConstraint() bool

	// CONNECT BY support
	SupportsConnectBy() bool

	// Transaction support
	SupportsStartTransactionModifier() bool
	SupportsEndTransactionModifier() bool

	// Named function arguments
	SupportsNamedFnArgsWithEqOperator() bool
	SupportsNamedFnArgsWithColonOperator() bool
	SupportsNamedFnArgsWithAssignmentOperator() bool
	SupportsNamedFnArgsWithRArrowOperator() bool
	SupportsNamedFnArgsWithExprName() bool

	// SET statement support
	SupportsParenthesizedSetVariables() bool
	SupportsCommaSeparatedSetAssignments() bool
	SupportsSetStmtWithoutOperator() bool
	SupportsSetNames() bool

	// SELECT clause support
	SupportsSelectWildcardExcept() bool
	SupportsSelectWildcardExclude() bool
	SupportsSelectExclude() bool
	SupportsSelectWildcardReplace() bool
	SupportsSelectWildcardIlike() bool
	SupportsSelectWildcardRename() bool
	SupportsSelectWildcardWithAlias() bool
	SupportsSelectExprStar() bool
	SupportsFromFirstSelect() bool
	SupportsEmptyProjections() bool
	SupportsSelectModifiers() bool
	SupportsPipeOperator() bool
	SupportsTrailingCommas() bool
	SupportsProjectionTrailingCommas() bool
	SupportsFromTrailingCommas() bool
	SupportsColumnDefinitionTrailingCommas() bool
	SupportsLimitComma() bool

	// Type conversion
	ConvertTypeBeforeValue() bool
	SupportsTryConvert() bool
	SupportsBinaryKwAsCast() bool

	// Object names and references
	SupportsObjectNameDoubleDotNotation() bool
	SupportsNumericPrefix() bool
	SupportsNumericLiteralUnderscores() bool

	// IN expressions
	SupportsInEmptyList() bool

	// Dictionary and literal syntax
	SupportsDictionarySyntax() bool
	SupportsMapLiteralSyntax() bool
	SupportsStructLiteral() bool
	SupportsArrayLiteralSyntax() bool
	SupportsLambdaFunctions() bool

	// Table definition support
	SupportsCreateTableMultiSchemaInfoSources() bool
	SupportsCreateTableLikeParenthesized() bool
	SupportsCreateTableSelect() bool
	SupportsCreateViewCommentSyntax() bool
	SupportsArrayTypedefWithoutElementType() bool
	SupportsArrayTypedefWithBrackets() bool
	SupportsParensAroundTableFactor() bool
	SupportsValuesAsTableFactor() bool
	SupportsUnnestTableFactor() bool
	SupportsSemanticViewTableFactor() bool
	SupportsTableVersioning() bool
	SupportsTableSampleBeforeAlias() bool
	SupportsTableHints() bool
	SupportsIndexHints() bool

	// Column definition support
	SupportsAscDescInColumnDefinition() bool
	SupportsSpaceSeparatedColumnOptions() bool
	SupportsConstraintKeywordWithoutName() bool
	SupportsKeyColumnOption() bool
	SupportsDataTypeSignedSuffix() bool

	// Comment support
	SupportsNestedComments() bool
	SupportsMultilineCommentHints() bool
	SupportsCommentOptimizerHint() bool
	SupportsCommentOn() bool
	RequiresSingleLineCommentWhitespace() bool

	// EXPLAIN support
	SupportsExplainWithUtilityOptions() bool

	// EXECUTE IMMEDIATE support
	SupportsExecuteImmediate() bool

	// Extract function support
	AllowExtractCustom() bool
	AllowExtractSingleQuotes() bool
	SupportsExtractCommaSyntax() bool

	// Subquery support
	SupportsSubqueryAsFunctionArg() bool

	// Placeholder support
	SupportsDollarPlaceholder() bool

	// Index support
	SupportsCreateIndexWithClause() bool

	// Interval support
	RequireIntervalQualifier() bool
	SupportsIntervalOptions() bool

	// Operator support
	SupportsFactorialOperator() bool
	SupportsBitwiseShiftOperators() bool
	SupportsNotnullOperator() bool
	SupportsBangNotOperator() bool
	SupportsDoubleAmpersandOperator() bool

	// MATCH AGAINST support
	SupportsMatchAgainst() bool

	// Grantee support
	SupportsUserHostGrantee() bool

	// LISTEN/NOTIFY support
	SupportsListenNotify() bool

	// LOAD support
	SupportsLoadData() bool
	SupportsLoadExtension() bool

	// TOP/DISTINCT ordering
	SupportsTopBeforeDistinct() bool

	// Boolean literals
	SupportsBooleanLiterals() bool

	// SHOW statement support
	SupportsShowLikeBeforeIn() bool

	// PartiQL support
	SupportsPartiQL() bool

	// Alias assignment
	SupportsEqAliasAssignment() bool

	// INSERT statement support
	SupportsInsertSet() bool
	SupportsInsertTableFunction() bool
	SupportsInsertTableQuery() bool
	SupportsInsertFormat() bool
	SupportsInsertTableAlias() bool

	// ALTER TABLE support
	SupportsAlterColumnTypeUsing() bool
	SupportsCommaSeparatedDropColumnList() bool

	// ORDER BY support
	SupportsOrderByAll() bool

	// Geometric types support
	SupportsGeometricTypes() bool

	// DESCRIBE support
	DescribeRequiresTableKeyword() bool

	// ClickHouse-specific support
	SupportsOptimizeTable() bool
	SupportsPrewhere() bool
	SupportsWithFill() bool
	SupportsLimitBy() bool
	SupportsInterpolate() bool
	SupportsSettings() bool
	SupportsSelectFormat() bool

	// DuckDB-specific support
	SupportsInstall() bool
	SupportsDetach() bool

	// TRIM support
	SupportsCommaSeparatedTrim() bool

	// Keyword helpers
	GetReservedKeywordsForSelectItemOperator() []token.Keyword
	GetReservedGranteesTypes() []GranteesType
	IsColumnAlias(kw token.Keyword, parser Parser) bool
	IsSelectItemAlias(explicit bool, kw token.Keyword, parser Parser) bool
	IsTableFactor(kw token.Keyword, parser Parser) bool
	IsTableAlias(kw token.Keyword, parser Parser) bool
	IsTableFactorAlias(explicit bool, kw token.Keyword, parser Parser) bool
}
