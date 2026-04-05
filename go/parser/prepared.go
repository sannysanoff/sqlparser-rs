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
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/datatype"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/token"
)

// parseDeallocate parses DEALLOCATE statements
// Reference: src/parser/mod.rs parse_deallocate
// DEALLOCATE [PREPARE] { name | ALL }
func parseDeallocate(p *Parser) (ast.Statement, error) {
	// Optional PREPARE keyword
	prepare := p.ParseKeyword("PREPARE")

	// Parse the name (identifier or ALL)
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	return &statement.Deallocate{
		Name:    name,
		Prepare: prepare,
	}, nil
}

// parseExecute parses EXECUTE statements
// Reference: src/parser/mod.rs parse_execute
// EXECUTE name [ (parameter [, ...]) ]
// EXECUTE IMMEDIATE (for some dialects like BigQuery/Snowflake)
func parseExecute(p *Parser) (ast.Statement, error) {
	executeStmt := &statement.Execute{}

	// Check for EXECUTE IMMEDIATE
	if p.ParseKeyword("IMMEDIATE") {
		executeStmt.Immediate = true
		// For EXECUTE IMMEDIATE, we would expect a string expression
		// For now, we return the basic structure
		return executeStmt, nil
	}

	// Track whether the name was wrapped in parentheses
	hasParentheses := p.ConsumeToken(token.TokenLParen{})

	// Parse the name (object name for procedure/function)
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}
	executeStmt.Name = name

	if hasParentheses {
		// Expect closing parenthesis
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		executeStmt.HasParentheses = true
		// When name has parens (EXEC (@sql)), it's dynamic SQL and takes no parameters
		return executeStmt, nil
	}

	// Check for parameter list in parentheses
	if p.ConsumeToken(token.TokenLParen{}) {
		executeStmt.HasParentheses = true

		// Parse comma-separated parameters
		exprParser := NewExpressionParser(p)
		for {
			// Check for empty list or end
			if p.ConsumeToken(token.TokenRParen{}) {
				break
			}

			// Parse parameter expression
			param, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			executeStmt.Parameters = append(executeStmt.Parameters, param)

			// Check for comma or closing parenthesis
			if p.ConsumeToken(token.TokenComma{}) {
				continue
			}

			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
			break
		}
	}

	// Check for USING clause (PostgreSQL style)
	if p.ParseKeyword("USING") {
		exprParser := NewExpressionParser(p)
		for {
			// Parse expression with optional alias
			exprVal, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			executeStmt.Using = append(executeStmt.Using, &expr.ExprWithAlias{Expr: exprVal})

			// Check for comma or end
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
	}

	return executeStmt, nil
}

// parsePrepare parses PREPARE statements
// Reference: src/parser/mod.rs parse_prepare
// PREPARE name [ ( data_type [, ...] ) ] AS statement
func parsePrepare(p *Parser) (ast.Statement, error) {
	// Parse the prepared statement name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	prepareStmt := &statement.Prepare{
		Name:      name,
		DataTypes: []datatype.DataType{},
	}

	// Parse optional data types list
	if p.ConsumeToken(token.TokenLParen{}) {
		// Parse comma-separated data types
		for {
			// Check for empty list or end
			if p.ConsumeToken(token.TokenRParen{}) {
				break
			}

			// Parse a data type
			dataType, err := p.ParseDataType()
			if err != nil {
				return nil, err
			}
			prepareStmt.DataTypes = append(prepareStmt.DataTypes, dataType)

			// Check for comma or closing parenthesis
			if p.ConsumeToken(token.TokenComma{}) {
				continue
			}

			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
			break
		}
	}

	// Expect AS keyword
	if _, err := p.ExpectKeyword("AS"); err != nil {
		return nil, err
	}

	// Parse the statement to prepare
	stmt, err := p.ParseStatement()
	if err != nil {
		return nil, err
	}
	prepareStmt.Statement = stmt

	return prepareStmt, nil
}
