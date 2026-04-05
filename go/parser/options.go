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

import "github.com/user/sqlparser/parseriface"

// ParserOptions controls how the Parser parses SQL text.
// These options allow you to mix & match behavior otherwise
// constrained to certain dialects (e.g. trailing commas).
type ParserOptions = parseriface.ParserOptions

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
func NewParserOptions(opts ...ParserOption) parseriface.ParserOptions {
	options := parseriface.ParserOptions{
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
type ParserOption func(*parseriface.ParserOptions)

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
	return func(o *parseriface.ParserOptions) {
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
	return func(o *parseriface.ParserOptions) {
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
	return func(o *parseriface.ParserOptions) {
		o.RequireSemicolon = required
	}
}

// CloneOptions returns a copy of the ParserOptions.
func CloneOptions(o parseriface.ParserOptions) parseriface.ParserOptions {
	return parseriface.ParserOptions{
		TrailingCommas:   o.TrailingCommas,
		Unescape:         o.Unescape,
		RequireSemicolon: o.RequireSemicolon,
	}
}
