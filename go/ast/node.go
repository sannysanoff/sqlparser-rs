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

// Package ast provides the Abstract Syntax Tree (AST) types for SQL parsing.
// This is the core interface hierarchy that all AST nodes implement.
package ast

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/span"
)

// Node is the base interface for all AST nodes.
// It is sealed to prevent external implementations.
type Node interface {
	node()
	// Span returns the source location of this node in the original SQL.
	Span() span.Span
	// String returns the SQL representation of this node.
	String() string
}

// Statement is the interface for all SQL statements.
// It extends Node and is also sealed.
type Statement interface {
	Node
	statement()
	// IsStatement is a marker method to distinguish statements from expressions.
	IsStatement()
}

// Expr is the interface for all SQL expressions.
// It extends Node and is also sealed.
type Expr interface {
	Node
	expr()
	// IsExpr is a marker method to distinguish expressions from statements.
	IsExpr()
}

// DataType is the interface for all SQL data types.
// It extends Node and is also sealed.
type DataType interface {
	Node
	dataType()
	// IsDataType is a marker method to identify data type nodes.
	IsDataType()
}

// Query represents a complete query expression, including WITH, UNION, ORDER BY, etc.
// It extends Node and is also sealed.
type Query interface {
	Node
	query()
	// IsQuery is a marker method to identify query nodes.
	IsQuery()
}

// BaseNode provides common functionality for all AST node implementations.
// It should be embedded in concrete node types.
type BaseNode struct {
	// span tracks the source location of this node
	span span.Span
}

// Span returns the source location of this node.
func (n *BaseNode) Span() span.Span {
	return n.span
}

// SetSpan sets the source location of this node.
func (n *BaseNode) SetSpan(s span.Span) {
	n.span = s
}

// node implements the Node interface (sealed).
func (n *BaseNode) node() {}

// BaseStatement provides common functionality for statement implementations.
// Types that embed BaseStatement will automatically implement the Statement interface.
type BaseStatement struct {
	BaseNode
}

// statement implements the Statement interface (sealed).
func (s *BaseStatement) statement() {}

// IsStatement marks this as a statement node.
func (s *BaseStatement) IsStatement() {}

// displaySeparated formats a slice of elements separated by a given string.
// This is the Go equivalent of display_separated in Rust.
func displaySeparated[T fmt.Stringer](slice []T, sep string) string {
	var parts []string
	for _, item := range slice {
		parts = append(parts, item.String())
	}
	return strings.Join(parts, sep)
}

// displayCommaSeparated formats a slice of elements separated by ", ".
// This is the Go equivalent of display_comma_separated in Rust.
func displayCommaSeparated[T fmt.Stringer](slice []T) string {
	return displaySeparated(slice, ", ")
}

// formatStatementList formats a slice of statements, each ending with a semicolon.
// This is the Go equivalent of format_statement_list in Rust.
func formatStatementList(statements []Statement) string {
	var parts []string
	for _, stmt := range statements {
		parts = append(parts, stmt.String())
	}
	result := strings.Join(parts, "; ")
	if len(result) > 0 {
		result += ";"
	}
	return result
}
