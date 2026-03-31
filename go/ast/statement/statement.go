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

// Package statement provides SQL statement AST types.
package statement

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/expr"
)

// DisplaySeparated is a helper for formatting slices with separators
type DisplaySeparated struct {
	Slice []fmt.Stringer
	Sep   string
}

func (d DisplaySeparated) String() string {
	var parts []string
	for _, s := range d.Slice {
		parts = append(parts, s.String())
	}
	return strings.Join(parts, d.Sep)
}

// DisplayCommaSeparated returns a comma-separated display helper
func DisplayCommaSeparated(slice []fmt.Stringer) DisplaySeparated {
	return DisplaySeparated{Slice: slice, Sep: ", "}
}

// BaseStatement provides common fields for all statements
type BaseStatement struct {
	ast.BaseStatement
}

// helper functions for SQL generation
func writeIfNotEmpty(f *strings.Builder, prefix, value string) {
	if value != "" {
		f.WriteString(prefix)
		f.WriteString(value)
	}
}

func writeIfTrue(f *strings.Builder, condition bool, value string) {
	if condition {
		f.WriteString(value)
	}
}

func writeOptional(f *strings.Builder, prefix string, value fmt.Stringer) {
	if value != nil {
		f.WriteString(prefix)
		f.WriteString(value.String())
	}
}

// Helper function to format expressions slice
func formatExprs(exprs []expr.Expr, sep string) string {
	if len(exprs) == 0 {
		return ""
	}
	parts := make([]string, len(exprs))
	for i, e := range exprs {
		parts[i] = e.String()
	}
	return strings.Join(parts, sep)
}

// Helper function to format idents slice
func formatIdents(idents []*ast.Ident, sep string) string {
	if len(idents) == 0 {
		return ""
	}
	parts := make([]string, len(idents))
	for i, ident := range idents {
		parts[i] = ident.String()
	}
	return strings.Join(parts, sep)
}

// Helper function to format object names slice
func formatObjectNames(names []*ast.ObjectName, sep string) string {
	if len(names) == 0 {
		return ""
	}
	parts := make([]string, len(names))
	for i, name := range names {
		parts[i] = name.String()
	}
	return strings.Join(parts, sep)
}
