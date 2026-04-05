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

// Package dialects provides SQL dialect definitions and interfaces for the sqlparser.
//
// This package defines the Dialect interface and its sub-interfaces that
// encapsulate the differences between SQL implementations. The interfaces
// are organized using the Interface Segregation Principle, splitting a large
// monolithic interface into focused, cohesive capability interfaces.
//
// The interfaces defined here work with the parseriface package to avoid
// circular dependencies with the parser package.
package dialects

import (
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/parseriface"
	"github.com/user/sqlparser/token"
)

// Re-export types from parseriface for backward compatibility
type (
	// ParserAccessor is an alias for parseriface.Parser for backward compatibility.
	// Deprecated: Use parseriface.Parser directly.
	ParserAccessor = parseriface.Parser

	// ParserState is an alias for parseriface.ParserState.
	// Deprecated: Use parseriface.ParserState directly.
	ParserState = parseriface.ParserState

	// ParserOptions is an alias for parseriface.ParserOptions.
	// Deprecated: Use parseriface.ParserOptions directly.
	ParserOptions = parseriface.ParserOptions

	// Precedence is an alias for parseriface.Precedence.
	// Deprecated: Use parseriface.Precedence directly.
	Precedence = parseriface.Precedence

	// NestedIdentifierQuote is an alias for parseriface.NestedIdentifierQuote.
	// Deprecated: Use parseriface.NestedIdentifierQuote directly.
	NestedIdentifierQuote = parseriface.NestedIdentifierQuote

	// GranteesType is an alias for parseriface.GranteesType.
	// Deprecated: Use parseriface.GranteesType directly.
	GranteesType = parseriface.GranteesType

	// ColumnOption is an alias for parseriface.ColumnOption.
	// Deprecated: Use parseriface.ColumnOption directly.
	ColumnOption = parseriface.ColumnOption

	// Dialect is an alias for parseriface.Dialect.
	// Deprecated: Use CompleteDialect or specific capability interfaces.
	Dialect = CompleteDialect
)

// Re-export constants from parseriface for backward compatibility
const (
	// Parser states
	StateNormal           = parseriface.StateNormal
	StateConnectBy        = parseriface.StateConnectBy
	StateColumnDefinition = parseriface.StateColumnDefinition

	// Precedence levels
	PrecedencePeriod      = parseriface.PrecedencePeriod
	PrecedenceDoubleColon = parseriface.PrecedenceDoubleColon
	PrecedenceAtTz        = parseriface.PrecedenceAtTz
	PrecedenceMulDivModOp = parseriface.PrecedenceMulDivModOp
	PrecedencePlusMinus   = parseriface.PrecedencePlusMinus
	PrecedenceXor         = parseriface.PrecedenceXor
	PrecedenceAmpersand   = parseriface.PrecedenceAmpersand
	PrecedenceCaret       = parseriface.PrecedenceCaret
	PrecedencePipe        = parseriface.PrecedencePipe
	PrecedenceColon       = parseriface.PrecedenceColon
	PrecedenceBetween     = parseriface.PrecedenceBetween
	PrecedenceEq          = parseriface.PrecedenceEq
	PrecedenceLike        = parseriface.PrecedenceLike
	PrecedenceIs          = parseriface.PrecedenceIs
	PrecedencePgOther     = parseriface.PrecedencePgOther
	PrecedenceUnaryNot    = parseriface.PrecedenceUnaryNot
	PrecedenceAnd         = parseriface.PrecedenceAnd
	PrecedenceOr          = parseriface.PrecedenceOr
	PrecedenceCollate     = parseriface.PrecedenceCollate

	// Grantee types
	GranteeTypeNone = parseriface.GranteeTypeNone
	GranteeTypeRole = parseriface.GranteeTypeRole
	GranteeTypeUser = parseriface.GranteeTypeUser
)

// ============================================================================
// Capability Sub-Interfaces (Interface Segregation Principle)
// ============================================================================

// CoreDialect defines the minimal dialect interface with just the identifier.
type CoreDialect interface {
	// Dialect returns the dialect name/identifier for identification purposes.
	Dialect() string
}

// IdentifierDialect defines identifier handling capabilities.
type IdentifierDialect interface {
	CoreDialect

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

	// IsReservedForIdentifier returns true if the specified keyword is reserved
	// and cannot be used as an identifier without special handling like quoting.
	IsReservedForIdentifier(kw token.Keyword) bool

	// IsIdentifierGeneratingFunctionName returns true if the dialect considers
	// the specified ident as a function that returns an identifier.
	IsIdentifierGeneratingFunctionName(ident *ast.Ident, nameParts []ast.ObjectNamePart) bool
}

// KeywordDialect defines keyword and alias handling capabilities.
type KeywordDialect interface {
	CoreDialect

	// GetReservedKeywordsForSelectItemOperator returns reserved keywords that may
	// prefix a select item expression (e.g., CONNECT_BY_ROOT in Snowflake).
	GetReservedKeywordsForSelectItemOperator() []token.Keyword

	// GetReservedGranteesTypes returns grantee types that should be treated as identifiers.
	GetReservedGranteesTypes() []GranteesType

	// IsColumnAlias returns true if the specified keyword should be parsed as a column alias.
	IsColumnAlias(kw token.Keyword, parser parseriface.Parser) bool

	// IsSelectItemAlias returns true if the specified keyword should be parsed as
	// a select item alias. When explicit is true, the keyword is preceded by an AS word.
	IsSelectItemAlias(explicit bool, kw token.Keyword, parser parseriface.Parser) bool

	// IsTableFactor returns true if the specified keyword should be parsed as a table factor identifier.
	IsTableFactor(kw token.Keyword, parser parseriface.Parser) bool

	// IsTableAlias returns true if the specified keyword should be parsed as a table factor alias.
	IsTableAlias(kw token.Keyword, parser parseriface.Parser) bool

	// IsTableFactorAlias returns true if the specified keyword should be parsed as
	// a table factor alias. When explicit is true, the keyword is preceded by an AS word.
	IsTableFactorAlias(explicit bool, kw token.Keyword, parser parseriface.Parser) bool
}

// StringLiteralDialect defines string literal handling capabilities.
type StringLiteralDialect interface {
	CoreDialect

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
}

// AggregationDialect defines aggregation and window function capabilities.
type AggregationDialect interface {
	CoreDialect

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
}

// GroupByDialect defines GROUP BY clause capabilities.
type GroupByDialect interface {
	CoreDialect

	// SupportsGroupByExpr returns true if the dialect supports GROUP BY expressions
	// like GROUPING SETS, ROLLUP, or CUBE.
	SupportsGroupByExpr() bool

	// SupportsGroupByWithModifier returns true if the dialect supports GROUP BY
	// modifiers prefixed by a WITH keyword (e.g., GROUP BY value WITH ROLLUP).
	SupportsGroupByWithModifier() bool
}

// JoinDialect defines JOIN clause capabilities.
type JoinDialect interface {
	CoreDialect

	// SupportsLeftAssociativeJoinsWithoutParens returns true if the dialect supports
	// left-associative join parsing by default when parentheses are omitted in nested joins.
	SupportsLeftAssociativeJoinsWithoutParens() bool

	// SupportsOuterJoinOperator returns true if the dialect supports the (+)
	// syntax for OUTER JOIN (Oracle-style).
	SupportsOuterJoinOperator() bool

	// SupportsCrossJoinConstraint returns true if the dialect supports a join
	// specification on CROSS JOIN.
	SupportsCrossJoinConstraint() bool
}

// TransactionDialect defines transaction statement capabilities.
type TransactionDialect interface {
	CoreDialect

	// SupportsStartTransactionModifier returns true if the dialect supports
	// BEGIN {DEFERRED | IMMEDIATE | EXCLUSIVE | TRY | CATCH} [TRANSACTION].
	SupportsStartTransactionModifier() bool

	// SupportsEndTransactionModifier returns true if the dialect supports
	// END {TRY | CATCH} statements.
	SupportsEndTransactionModifier() bool
}

// NamedArgumentDialect defines named function argument capabilities.
type NamedArgumentDialect interface {
	CoreDialect

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
}

// SetDialect defines SET statement capabilities.
type SetDialect interface {
	CoreDialect

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
}

// SelectDialect defines SELECT clause capabilities.
type SelectDialect interface {
	CoreDialect

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
}

// TypeConversionDialect defines type conversion capabilities.
type TypeConversionDialect interface {
	CoreDialect

	// ConvertTypeBeforeValue returns true if the dialect has a CONVERT function
	// which accepts a type first and an expression second.
	ConvertTypeBeforeValue() bool

	// SupportsTryConvert returns true if the dialect supports the TRY_CONVERT function.
	SupportsTryConvert() bool

	// SupportsBinaryKwAsCast returns true if the dialect supports casting an expression
	// to a binary type using the BINARY <expr> syntax.
	SupportsBinaryKwAsCast() bool
}

// ObjectReferenceDialect defines object name and reference capabilities.
type ObjectReferenceDialect interface {
	CoreDialect

	// SupportsObjectNameDoubleDotNotation returns true if the dialect supports
	// double dot notation for object names (e.g., db_name..table_name).
	SupportsObjectNameDoubleDotNotation() bool

	// SupportsNumericPrefix returns true if the dialect supports identifiers
	// starting with a numeric prefix.
	SupportsNumericPrefix() bool

	// SupportsNumericLiteralUnderscores returns true if the dialect supports
	// numbers containing underscores (e.g., 10_000_000).
	SupportsNumericLiteralUnderscores() bool
}

// InExpressionDialect defines IN expression capabilities.
type InExpressionDialect interface {
	CoreDialect

	// SupportsInEmptyList returns true if the dialect supports (NOT) IN ()
	// expressions with empty lists.
	SupportsInEmptyList() bool
}

// LiteralDialect defines dictionary and literal syntax capabilities.
type LiteralDialect interface {
	CoreDialect

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
}

// TableDefinitionDialect defines table definition capabilities.
type TableDefinitionDialect interface {
	CoreDialect

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

	// SupportsUnnestTableFactor returns true if the dialect supports UNNEST
	// as a table factor (BigQuery/PostgreSQL).
	SupportsUnnestTableFactor() bool

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
}

// ColumnDefinitionDialect defines column definition capabilities.
type ColumnDefinitionDialect interface {
	CoreDialect

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
}

// CommentDialect defines comment handling capabilities.
type CommentDialect interface {
	CoreDialect

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
}

// ExplainDialect defines EXPLAIN statement capabilities.
type ExplainDialect interface {
	CoreDialect

	// SupportsExplainWithUtilityOptions returns true if the dialect supports
	// EXPLAIN statements with utility options.
	SupportsExplainWithUtilityOptions() bool
}

// ExecuteDialect defines EXECUTE IMMEDIATE statement capabilities.
type ExecuteDialect interface {
	CoreDialect

	// SupportsExecuteImmediate returns true if the dialect supports EXECUTE IMMEDIATE statements.
	SupportsExecuteImmediate() bool
}

// ExtractDialect defines EXTRACT function capabilities.
type ExtractDialect interface {
	CoreDialect

	// AllowExtractCustom returns true if the dialect allows the EXTRACT function
	// to use words other than keywords.
	AllowExtractCustom() bool

	// AllowExtractSingleQuotes returns true if the dialect allows the EXTRACT
	// function to use single quotes in the part being extracted.
	AllowExtractSingleQuotes() bool

	// SupportsExtractCommaSyntax returns true if the dialect supports EXTRACT
	// function with a comma separator instead of FROM.
	SupportsExtractCommaSyntax() bool
}

// SubqueryDialect defines subquery capabilities.
type SubqueryDialect interface {
	CoreDialect

	// SupportsSubqueryAsFunctionArg returns true if the dialect supports a subquery
	// passed to a function as the only argument without enclosing parentheses.
	SupportsSubqueryAsFunctionArg() bool
}

// PlaceholderDialect defines placeholder capabilities.
type PlaceholderDialect interface {
	CoreDialect

	// SupportsDollarPlaceholder returns true if the dialect allows dollar placeholders (e.g., $var).
	SupportsDollarPlaceholder() bool
}

// IndexDialect defines index statement capabilities.
type IndexDialect interface {
	CoreDialect

	// SupportsCreateIndexWithClause returns true if the dialect supports WITH clause
	// in CREATE INDEX statement.
	SupportsCreateIndexWithClause() bool
}

// IntervalDialect defines INTERVAL expression capabilities.
type IntervalDialect interface {
	CoreDialect

	// RequireIntervalQualifier returns true if the dialect requires units
	// (qualifiers) to be specified in INTERVAL expressions.
	RequireIntervalQualifier() bool

	// SupportsIntervalOptions returns true if the dialect supports INTERVAL data
	// type with Postgres-style options.
	SupportsIntervalOptions() bool
}

// OperatorDialect defines operator support capabilities.
type OperatorDialect interface {
	CoreDialect

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
}

// MatchDialect defines MATCH AGAINST capabilities.
type MatchDialect interface {
	CoreDialect

	// SupportsMatchAgainst returns true if the dialect supports MATCH() AGAINST() syntax.
	SupportsMatchAgainst() bool
}

// GranteeDialect defines grantee capabilities.
type GranteeDialect interface {
	CoreDialect

	// SupportsUserHostGrantee returns true if the dialect supports MySQL-style
	// 'user'@'host' grantee syntax.
	SupportsUserHostGrantee() bool
}

// ListenNotifyDialect defines LISTEN/NOTIFY statement capabilities.
type ListenNotifyDialect interface {
	CoreDialect

	// SupportsListenNotify returns true if the dialect supports LISTEN, UNLISTEN,
	// and NOTIFY statements.
	SupportsListenNotify() bool
}

// LoadDialect defines LOAD statement capabilities.
type LoadDialect interface {
	CoreDialect

	// SupportsLoadData returns true if the dialect supports LOAD DATA statement.
	SupportsLoadData() bool

	// SupportsLoadExtension returns true if the dialect supports LOAD extension statement.
	SupportsLoadExtension() bool
}

// TopDistinctDialect defines TOP/DISTINCT ordering capabilities.
type TopDistinctDialect interface {
	CoreDialect

	// SupportsTopBeforeDistinct returns true if the dialect expects the TOP option
	// before the ALL/DISTINCT options in a SELECT statement.
	SupportsTopBeforeDistinct() bool
}

// BooleanLiteralDialect defines boolean literal capabilities.
type BooleanLiteralDialect interface {
	CoreDialect

	// SupportsBooleanLiterals returns true if the dialect supports boolean
	// literals (true and false).
	SupportsBooleanLiterals() bool
}

// ShowDialect defines SHOW statement capabilities.
type ShowDialect interface {
	CoreDialect

	// SupportsShowLikeBeforeIn returns true if the dialect supports the LIKE
	// option in a SHOW statement before the IN option.
	SupportsShowLikeBeforeIn() bool
}

// PartiQLDialect defines PartiQL capabilities.
type PartiQLDialect interface {
	CoreDialect

	// SupportsPartiQL returns true if the dialect supports PartiQL for querying
	// semi-structured data.
	SupportsPartiQL() bool
}

// AliasDialect defines alias assignment capabilities.
type AliasDialect interface {
	CoreDialect

	// SupportsEqAliasAssignment returns true if the dialect supports treating
	// the equals operator = within a SelectItem as an alias assignment operator.
	SupportsEqAliasAssignment() bool
}

// InsertDialect defines INSERT statement capabilities.
type InsertDialect interface {
	CoreDialect

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
}

// AlterTableDialect defines ALTER TABLE statement capabilities.
type AlterTableDialect interface {
	CoreDialect

	// SupportsAlterColumnTypeUsing returns true if the dialect supports the USING
	// clause in an ALTER COLUMN statement.
	SupportsAlterColumnTypeUsing() bool

	// SupportsCommaSeparatedDropColumnList returns true if the dialect supports
	// ALTER TABLE tbl DROP COLUMN c1, ..., cn.
	SupportsCommaSeparatedDropColumnList() bool
}

// OrderByDialect defines ORDER BY clause capabilities.
type OrderByDialect interface {
	CoreDialect

	// SupportsOrderByAll returns true if the dialect supports ORDER BY ALL.
	SupportsOrderByAll() bool
}

// GeometricDialect defines geometric type capabilities.
type GeometricDialect interface {
	CoreDialect

	// SupportsGeometricTypes returns true if the dialect supports geometric types
	// (Postgres geometric operations).
	SupportsGeometricTypes() bool
}

// DescribeDialect defines DESCRIBE statement capabilities.
type DescribeDialect interface {
	CoreDialect

	// DescribeRequiresTableKeyword returns true if the dialect requires the TABLE
	// keyword after DESCRIBE.
	DescribeRequiresTableKeyword() bool
}

// ClickHouseDialect defines ClickHouse-specific capabilities.
type ClickHouseDialect interface {
	CoreDialect

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
}

// DuckDBDialect defines DuckDB-specific capabilities.
type DuckDBDialect interface {
	CoreDialect

	// SupportsInstall returns true if the dialect supports INSTALL statement.
	SupportsInstall() bool

	// SupportsDetach returns true if the dialect supports DETACH statement.
	SupportsDetach() bool
}

// TrimDialect defines TRIM function capabilities.
type TrimDialect interface {
	CoreDialect

	// SupportsCommaSeparatedTrim returns true if the dialect supports the two-argument
	// comma-separated form of TRIM function: TRIM(expr, characters).
	SupportsCommaSeparatedTrim() bool
}

// ConnectByDialect defines CONNECT BY clause capabilities.
type ConnectByDialect interface {
	CoreDialect

	// SupportsConnectBy returns true if the dialect supports CONNECT BY for
	// hierarchical queries (Oracle-style).
	SupportsConnectBy() bool
}

// ============================================================================
// Complete Dialect Interface (Embeds all capability interfaces)
// ============================================================================

// CompleteDialect is the full dialect interface that embeds all capability sub-interfaces.
// This is the interface that concrete dialect implementations should satisfy.
type CompleteDialect interface {
	CoreDialect
	parseriface.Dialect
	parseriface.CompleteDialect
	IdentifierDialect
	KeywordDialect
	StringLiteralDialect
	AggregationDialect
	GroupByDialect
	JoinDialect
	TransactionDialect
	NamedArgumentDialect
	SetDialect
	SelectDialect
	TypeConversionDialect
	ObjectReferenceDialect
	InExpressionDialect
	LiteralDialect
	TableDefinitionDialect
	ColumnDefinitionDialect
	CommentDialect
	ExplainDialect
	ExecuteDialect
	ExtractDialect
	SubqueryDialect
	PlaceholderDialect
	IndexDialect
	IntervalDialect
	OperatorDialect
	MatchDialect
	GranteeDialect
	ListenNotifyDialect
	LoadDialect
	TopDistinctDialect
	BooleanLiteralDialect
	ShowDialect
	PartiQLDialect
	AliasDialect
	InsertDialect
	AlterTableDialect
	OrderByDialect
	GeometricDialect
	DescribeDialect
	ClickHouseDialect
	DuckDBDialect
	TrimDialect
	ConnectByDialect
}
