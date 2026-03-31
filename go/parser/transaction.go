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
	"fmt"

	"github.com/user/sqlparser/ast"
)

// ParseStartTransaction parses START TRANSACTION or BEGIN
func ParseStartTransaction(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("START TRANSACTION parsing not yet fully implemented")
}

// ParseBegin parses BEGIN TRANSACTION statements
func ParseBegin(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("BEGIN TRANSACTION parsing not yet fully implemented")
}

// ParseCommit parses COMMIT statements
func ParseCommit(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("COMMIT parsing not yet fully implemented")
}

// ParseRollback parses ROLLBACK statements
func ParseRollback(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ROLLBACK parsing not yet fully implemented")
}

// ParseSavepoint parses SAVEPOINT statements
func ParseSavepoint(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("SAVEPOINT parsing not yet fully implemented")
}

// ParseRelease parses RELEASE SAVEPOINT statements
func ParseRelease(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("RELEASE SAVEPOINT parsing not yet fully implemented")
}
