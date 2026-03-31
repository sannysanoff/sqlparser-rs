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

// ParserOptions controls how the Parser parses SQL text.
// These options allow you to mix & match behavior otherwise
// constrained to certain dialects (e.g. trailing commas).
type ParserOptions struct {
	// TrailingCommas controls whether trailing commas are allowed in lists.
	// If this option is false (the default), the following SQL will not parse:
	//
	//   SELECT foo, bar, FROM baz
	//
	// If the option is true, the SQL will parse.
	//
	// See also:
	//   - BigQuery: https://cloud.google.com/bigquery/docs/reference/standard-sql/lexical#trailing_commas
	//   - Snowflake: https://docs.snowflake.com/en/release-notes/2024/8_11#select-supports-trailing-commas
	TrailingCommas bool

	// Unescape controls how literal values are unescaped during tokenization.
	// When true (default), escape sequences in string literals are processed.
	// For example, 'It\'s' becomes "It's".
	//
	// See Tokenizer.WithUnescape for more details.
	Unescape bool

	// RequireSemicolon controls if the parser expects a semicolon token
	// between statements. Default is true.
	//
	// When true, statements must be separated by semicolons:
	//   SELECT 1; SELECT 2;
	//
	// When false, semicolons are optional:
	//   SELECT 1 SELECT 2
	RequireSemicolon bool
}

// NewParserOptions creates a new ParserOptions with default values.
//
// Default values:
//   - TrailingCommas: false
//   - Unescape: true
//   - RequireSemicolon: true
//
// Use the functional options pattern to customize:
//
//	options := parser.NewParserOptions(
//	    parser.WithTrailingCommas(true),
//	    parser.WithUnescape(false),
//	)
func NewParserOptions(opts ...ParserOption) ParserOptions {
	options := ParserOptions{
		TrailingCommas:   false,
		Unescape:         true,
		RequireSemicolon: true,
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// ParserOption is a functional option for configuring ParserOptions.
type ParserOption func(*ParserOptions)

// WithTrailingCommas sets whether trailing commas are allowed.
//
// Example:
//
//	options := parser.NewParserOptions(
//	    parser.WithTrailingCommas(true),
//	)
//	parser := parser.New(dialect).WithOptions(options)
//	// Now this SQL will parse successfully:
//	// SELECT a, b, COUNT(*), FROM foo GROUP BY a, b,
func WithTrailingCommas(allowed bool) ParserOption {
	return func(o *ParserOptions) {
		o.TrailingCommas = allowed
	}
}

// WithUnescape sets whether literal values are unescaped.
//
// Example:
//
//	options := parser.NewParserOptions(
//	    parser.WithUnescape(false),  // Keep escape sequences as-is
//	)
func WithUnescape(unescape bool) ParserOption {
	return func(o *ParserOptions) {
		o.Unescape = unescape
	}
}

// WithRequireSemicolon sets whether semicolons are required between statements.
//
// Example:
//
//	options := parser.NewParserOptions(
//	    parser.WithRequireSemicolon(false),  // Allow statements without semicolons
//	)
func WithRequireSemicolon(required bool) ParserOption {
	return func(o *ParserOptions) {
		o.RequireSemicolon = required
	}
}

// Clone returns a copy of the ParserOptions.
func (o ParserOptions) Clone() ParserOptions {
	return ParserOptions{
		TrailingCommas:   o.TrailingCommas,
		Unescape:         o.Unescape,
		RequireSemicolon: o.RequireSemicolon,
	}
}
