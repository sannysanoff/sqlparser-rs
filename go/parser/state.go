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

// ParserState represents the current state of the parser.
// The parser can be in different states depending on what
// context it is parsing (e.g., normal statements, column definitions,
// CONNECT BY expressions).
type ParserState int

const (
	// StateNormal is the default state of the parser.
	// In this state, the parser processes standard SQL statements
	// without any special context.
	StateNormal ParserState = iota

	// StateConnectBy is the state when parsing a CONNECT BY expression.
	// This allows parsing PRIOR expressions while still allowing PRIOR
	// as an identifier name in other contexts.
	//
	// Example:
	//   SELECT * FROM employees
	//   START WITH manager_id IS NULL
	//   CONNECT BY PRIOR employee_id = manager_id
	StateConnectBy

	// StateColumnDefinition is the state when parsing column definitions.
	// This state prohibits NOT NULL as an alias for IS NOT NULL
	// in column definitions.
	//
	// Example:
	//   CREATE TABLE foo (abc BIGINT NOT NULL);
	//
	// In this context, NOT NULL is a column constraint, not a boolean expression.
	StateColumnDefinition
)

// String returns the string representation of the parser state.
func (s ParserState) String() string {
	switch s {
	case StateNormal:
		return "Normal"
	case StateConnectBy:
		return "ConnectBy"
	case StateColumnDefinition:
		return "ColumnDefinition"
	default:
		return "Unknown"
	}
}
