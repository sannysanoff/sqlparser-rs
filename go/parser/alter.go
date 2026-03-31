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

// parseAlter parses ALTER statements
// TODO: Implement full ALTER parsing
func parseAlter(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER statement parsing not yet implemented")
}

// All ALTER TABLE operations stubbed out
// These need proper implementation with correct types

func parseAlterTable(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER TABLE parsing not yet implemented")
}

func parseAlterView(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER VIEW parsing not yet implemented")
}

func parseAlterIndex(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER INDEX parsing not yet implemented")
}

func parseAlterRole(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER ROLE parsing not yet implemented")
}

func parseAlterUser(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER USER parsing not yet implemented")
}

func parseAlterSchema(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER SCHEMA parsing not yet implemented")
}

func parseAlterType(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER TYPE parsing not yet implemented")
}

func parseAlterPolicy(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER POLICY parsing not yet implemented")
}

func parseAlterConnector(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER CONNECTOR parsing not yet implemented")
}

func parseAlterOperator(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER OPERATOR parsing not yet implemented")
}

func parseAlterOperatorClass(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER OPERATOR CLASS parsing not yet implemented")
}

func parseAlterOperatorFamily(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER OPERATOR FAMILY parsing not yet implemented")
}
