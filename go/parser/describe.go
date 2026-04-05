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
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/token"
)

// ParseExplain parses EXPLAIN statements
func ParseExplain(p *Parser) (ast.Statement, error) {
	return parseExplain(p)
}

// parseExplain parses EXPLAIN statements
func parseExplain(p *Parser) (ast.Statement, error) {
	return parseExplainWithAlias(p, expr.DescribeAliasExplain)
}

// ParseDescribe parses DESCRIBE statements
func ParseDescribe(p *Parser) (ast.Statement, error) {
	return ParseDescribeWithAlias(p, "DESCRIBE")
}

// ParseDescribeWithAlias parses DESCRIBE/DESC statements with the given alias
func ParseDescribeWithAlias(p *Parser, keyword string) (ast.Statement, error) {
	// Determine the alias based on the keyword used
	var alias expr.DescribeAlias
	switch keyword {
	case "DESC":
		alias = expr.DescribeAliasDesc
	case "DESCRIBE":
		alias = expr.DescribeAliasDescribe
	default:
		alias = expr.DescribeAliasDescribe
	}

	// Check if this looks like a statement (SELECT, INSERT, UPDATE, DELETE)
	// DESCRIBE is an alias for EXPLAIN when followed by a statement
	if p.PeekKeyword("SELECT") || p.PeekKeyword("INSERT") || p.PeekKeyword("UPDATE") || p.PeekKeyword("DELETE") {
		return parseExplainWithAlias(p, alias)
	}

	// Check for utility options in parentheses (PostgreSQL style)
	if p.GetDialect().SupportsExplainWithUtilityOptions() {
		if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
			return parseExplainWithAlias(p, alias)
		}
	}

	// Check for TABLE keyword (optional in some dialects)
	hasTableKeyword := p.ParseKeyword("TABLE")

	// Check for optional Hive format (EXTENDED or FORMATTED)
	var hiveFormat *expr.HiveDescribeFormat
	if p.ParseKeyword("EXTENDED") {
		format := expr.HiveDescribeFormatExtended
		hiveFormat = &format
	} else if p.ParseKeyword("FORMATTED") {
		format := expr.HiveDescribeFormatFormatted
		hiveFormat = &format
	}

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	return &statement.ExplainTable{
		DescribeAlias:   alias,
		TableName:       tableName,
		HasTableKeyword: hasTableKeyword,
		HiveFormat:      hiveFormat,
	}, nil
}

// parseExplainWithAlias parses EXPLAIN/DESCRIBE/DESC statements
func parseExplainWithAlias(p *Parser, describeAlias expr.DescribeAlias) (ast.Statement, error) {
	// Check for utility options in parentheses (PostgreSQL style)
	// EXPLAIN (ANALYZE, VERBOSE) SELECT ...
	var options []*expr.UtilityOption
	if p.GetDialect().SupportsExplainWithUtilityOptions() {
		if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
			opts, err := parseUtilityOptions(p)
			if err != nil {
				return nil, err
			}
			options = opts
		}
	}

	// Check for QUERY PLAN syntax
	if p.ParseKeyword("QUERY") {
		// EXPLAIN QUERY PLAN
		if !p.ParseKeyword("PLAN") {
			return nil, fmt.Errorf("expected PLAN after QUERY, found: %v", p.PeekToken())
		}
		// Parse the statement after QUERY PLAN
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		return &statement.Explain{
			DescribeAlias: describeAlias,
			Analyze:       false,
			Verbose:       false,
			QueryPlan:     true,
			Estimate:      false,
			Statement:     stmt,
			Options:       options,
		}, nil
	}

	// Parse optional flags
	analyze := p.ParseKeyword("ANALYZE")
	verbose := p.ParseKeyword("VERBOSE")

	// Check for ESTIMATE (Snowflake)
	estimate := p.ParseKeyword("ESTIMATE")

	// Check if next token looks like a statement keyword
	if p.PeekKeyword("SELECT") || p.PeekKeyword("INSERT") || p.PeekKeyword("UPDATE") || p.PeekKeyword("DELETE") {
		// Try to parse as a statement
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		return &statement.Explain{
			DescribeAlias: describeAlias,
			Analyze:       analyze,
			Verbose:       verbose,
			QueryPlan:     false,
			Estimate:      estimate,
			Statement:     stmt,
			Options:       options,
		}, nil
	}

	// Parse as table name (DESCRIBE table_name)
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	return &statement.ExplainTable{
		DescribeAlias:   describeAlias,
		TableName:       tableName,
		HasTableKeyword: false,
	}, nil
}

// parseUtilityOptions parses utility options in the form of `(option1, option2 arg2, option3 arg3, ...)`
// Reference: src/parser/mod.rs parse_utility_options
func parseUtilityOptions(p *Parser) ([]*expr.UtilityOption, error) {
	// Expect opening parenthesis
	if _, ok := p.PeekToken().Token.(token.TokenLParen); !ok {
		return nil, fmt.Errorf("expected opening parenthesis, found: %v", p.PeekToken())
	}
	p.NextToken() // consume LParen

	var options []*expr.UtilityOption

	// Handle empty parentheses
	if _, ok := p.PeekToken().Token.(token.TokenRParen); ok {
		p.NextToken() // consume RParen
		return options, nil
	}

	for {
		// Parse option name
		_, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected identifier, found: %v", p.PeekToken())
		}

		// Check if there's an argument
		nextToken := p.PeekToken()
		if _, isComma := nextToken.Token.(token.TokenComma); isComma {
			// No argument, just the option name
			options = append(options, &expr.UtilityOption{})
		} else if _, isRParen := nextToken.Token.(token.TokenRParen); isRParen {
			// No argument, just the option name
			options = append(options, &expr.UtilityOption{})
		} else {
			// Has argument - just consume tokens until comma or closing paren
			// This is a simplified version - in practice we'd parse an expression
			for {
				next := p.PeekToken()
				if _, isComma := next.Token.(token.TokenComma); isComma {
					break
				}
				if _, isRParen := next.Token.(token.TokenRParen); isRParen {
					break
				}
				p.NextToken()
			}
			options = append(options, &expr.UtilityOption{})
		}

		// Check for comma or closing parenthesis
		nextToken = p.PeekToken()
		if _, isRParen := nextToken.Token.(token.TokenRParen); isRParen {
			p.NextToken() // consume RParen
			break
		} else if _, isComma := nextToken.Token.(token.TokenComma); isComma {
			p.NextToken() // consume Comma
		} else {
			return nil, fmt.Errorf("expected comma or closing parenthesis, found: %v", nextToken)
		}
	}

	return options, nil
}

// parseDescribe parses DESCRIBE statements (internal helper)
func parseDescribe(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("DESCRIBE statement parsing not yet fully implemented")
}
