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

// Package errors provides error types and handling for the SQL parser.
//
// This package defines ParserError and related types for representing
// errors that occur during SQL tokenization and parsing. Errors include
// source location information (line, column) for accurate error reporting.
//
// Error Types:
//
//   - TokenizerErrorType: Errors during lexical analysis (tokenization)
//   - SyntaxErrorType: Errors during parsing (unexpected tokens, syntax violations)
//   - RecursionLimitExceededType: Parser recursion depth exceeded safety limit
//
// Usage:
//
//	// Creating a parser error
//	err := errors.NewSyntaxError("Unexpected token: FOO", span)
//
//	// Checking error type
//	if err.Type == errors.SyntaxErrorType {
//	    // Handle syntax error
//	}
//
//	// Getting error with location
//	fmt.Println(err.Error())  // "ParserError at Line: 1, Column: 10: Unexpected token: FOO"
package errors
