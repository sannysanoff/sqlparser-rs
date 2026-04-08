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

package parser

import (
	"github.com/user/sqlparser/parseriface"
)

// dialectAdapter wraps a parseriface.CompleteDialect to implement tokenizer.Dialect
type dialectAdapter struct {
	dialect parseriface.CompleteDialect
}

// newDialectAdapter creates a new adapter wrapping the given dialect
func newDialectAdapter(d parseriface.CompleteDialect) *dialectAdapter {
	return &dialectAdapter{dialect: d}
}

// IsIdentifierStart implements tokenizer.Dialect
func (a *dialectAdapter) IsIdentifierStart(ch rune) bool {
	return a.dialect.IsIdentifierStart(ch)
}

// IsIdentifierPart implements tokenizer.Dialect
func (a *dialectAdapter) IsIdentifierPart(ch rune) bool {
	return a.dialect.IsIdentifierPart(ch)
}

// IsDelimitedIdentifierStart implements tokenizer.Dialect
func (a *dialectAdapter) IsDelimitedIdentifierStart(ch rune) bool {
	return a.dialect.IsDelimitedIdentifierStart(ch)
}

// IsNestedDelimitedIdentifierStart implements tokenizer.Dialect
func (a *dialectAdapter) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return a.dialect.IsNestedDelimitedIdentifierStart(ch)
}

// PeekNestedDelimitedIdentifierQuotes implements tokenizer.Dialect
// Adapts the different signatures between dialects and tokenizer
func (a *dialectAdapter) PeekNestedDelimitedIdentifierQuotes(chars string) (startQuote byte, nestedQuote byte, ok bool) {
	// Convert string to []rune and call the dialect version
	// For now, just return false (no nested quotes)
	// This is a simplified implementation
	return 0, 0, false
}

// SupportsStringLiteralBackslashEscape implements tokenizer.Dialect
func (a *dialectAdapter) SupportsStringLiteralBackslashEscape() bool {
	return a.dialect.SupportsStringLiteralBackslashEscape()
}

// SupportsTripleQuotedString implements tokenizer.Dialect
func (a *dialectAdapter) SupportsTripleQuotedString() bool {
	return a.dialect.SupportsTripleQuotedString()
}

// SupportsDollarQuotedString implements tokenizer.Dialect
func (a *dialectAdapter) SupportsDollarQuotedString() bool {
	// Delegate to the underlying dialect
	return a.dialect.SupportsDollarQuotedString()
}

// SupportsDollarPlaceholder implements tokenizer.Dialect
func (a *dialectAdapter) SupportsDollarPlaceholder() bool {
	return a.dialect.SupportsDollarPlaceholder()
}

// SupportsNumericLiteralUnderscores implements tokenizer.Dialect
func (a *dialectAdapter) SupportsNumericLiteralUnderscores() bool {
	return a.dialect.SupportsNumericLiteralUnderscores()
}

// SupportsNumericPrefix implements tokenizer.Dialect
func (a *dialectAdapter) SupportsNumericPrefix() bool {
	return a.dialect.SupportsNumericPrefix()
}

// SupportsNestedComments implements tokenizer.Dialect
func (a *dialectAdapter) SupportsNestedComments() bool {
	return a.dialect.SupportsNestedComments()
}

// SupportsMultilineCommentHints implements tokenizer.Dialect
func (a *dialectAdapter) SupportsMultilineCommentHints() bool {
	return a.dialect.SupportsMultilineCommentHints()
}

// SupportsQuoteDelimitedString implements tokenizer.Dialect
func (a *dialectAdapter) SupportsQuoteDelimitedString() bool {
	return a.dialect.SupportsQuoteDelimitedString()
}

// SupportsStringEscapeConstant implements tokenizer.Dialect
func (a *dialectAdapter) SupportsStringEscapeConstant() bool {
	return a.dialect.SupportsStringEscapeConstant()
}

// SupportsUnicodeStringLiteral implements tokenizer.Dialect
func (a *dialectAdapter) SupportsUnicodeStringLiteral() bool {
	return a.dialect.SupportsUnicodeStringLiteral()
}

// SupportsGeometricTypes implements tokenizer.Dialect
func (a *dialectAdapter) SupportsGeometricTypes() bool {
	return a.dialect.SupportsGeometricTypes()
}

// SupportsPipeOperator implements tokenizer.Dialect
func (a *dialectAdapter) SupportsPipeOperator() bool {
	return a.dialect.SupportsPipeOperator()
}

// IgnoresWildcardEscapes implements tokenizer.Dialect
func (a *dialectAdapter) IgnoresWildcardEscapes() bool {
	return a.dialect.IgnoresWildcardEscapes()
}

// RequiresSingleLineCommentWhitespace implements tokenizer.Dialect
func (a *dialectAdapter) RequiresSingleLineCommentWhitespace() bool {
	// Default to false - most dialects don't require whitespace after --
	return false
}

// IsBigQueryDialect implements tokenizer.Dialect
func (a *dialectAdapter) IsBigQueryDialect() bool {
	return a.dialect.Dialect() == "bigquery"
}

// IsPostgreSqlDialect implements tokenizer.Dialect
func (a *dialectAdapter) IsPostgreSqlDialect() bool {
	return a.dialect.Dialect() == "postgresql"
}

// IsMySqlDialect implements tokenizer.Dialect
func (a *dialectAdapter) IsMySqlDialect() bool {
	return a.dialect.Dialect() == "mysql"
}

// IsSnowflakeDialect implements tokenizer.Dialect
func (a *dialectAdapter) IsSnowflakeDialect() bool {
	return a.dialect.Dialect() == "snowflake"
}

// IsDuckDbDialect implements tokenizer.Dialect
func (a *dialectAdapter) IsDuckDbDialect() bool {
	return a.dialect.Dialect() == "duckdb"
}

// IsGenericDialect implements tokenizer.Dialect
func (a *dialectAdapter) IsGenericDialect() bool {
	return a.dialect.Dialect() == "generic"
}

// IsHiveDialect implements tokenizer.Dialect
func (a *dialectAdapter) IsHiveDialect() bool {
	return a.dialect.Dialect() == "hive"
}

// IsCustomOperatorPart implements tokenizer.Dialect
func (a *dialectAdapter) IsCustomOperatorPart(ch rune) bool {
	return a.dialect.IsCustomOperatorPart(ch)
}
