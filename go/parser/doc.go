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

// Package parser provides SQL parsing for the sqlparser library.
//
// This package implements all parsing methods transpiled from the
// Rust sqlparser-rs crate's parser module. It uses Pratt parsing (precedence climbing)
// for handling operator precedence correctly.
//
// The parser is organized into several files:
//
//   - core.go: Contains ParseExpr(), ParseExprWithPrecedence(), and GetNextPrecedence()
//     which form the foundation of expression parsing with Pratt parsing.
//
//   - prefix.go: Contains all prefix expression parsers like parseIdentifier(),
//     parseValue(), parseFunction(), parseCase(), parseCast(), etc.
//
//   - infix.go: Contains all infix expression parsers like parseBinaryOp(),
//     parseIsNull(), parseIn(), parseBetween(), parseLike(), etc.
//
//   - postfix.go: Contains postfix expression parsers like parseArraySubscript()
//     and parseCollate().
//
//   - special.go: Contains special expression parsers like parseWindowFunction(),
//     parseAggregateFunction(), parseLambda(), parseExists(), etc.
//
//   - helpers.go: Contains utility methods like parseCommaSeparatedExprs(),
//     parseParenthesizedExpr(), parseOptionalAlias(), etc.
//
//   - groupings.go: Contains GROUPING SETS, CUBE, and ROLLUP expression parsers.
//
// The parser supports multiple SQL dialects through the dialects.Dialect interface,
// allowing dialect-specific syntax variations to be handled correctly.
//
// Usage:
//
//	dialect := dialects.NewGenericDialect()
//	p := parser.New(dialect)
//	expr, err := p.ParseExpr()
package parser
