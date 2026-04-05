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

// ParseDrop parses DROP statements
// Reference: src/parser/mod.rs parse_drop
func ParseDrop(p *Parser) (ast.Statement, error) {
	// Check for TEMPORARY (MySQL style)
	temporary := p.ParseKeyword("TEMPORARY")

	// Determine object type based on next keyword
	switch {
	case p.PeekKeyword("TABLE"):
		return parseDropTable(p, temporary)
	case p.PeekKeyword("VIEW"):
		return parseDropView(p, temporary)
	case p.PeekKeyword("INDEX"):
		return parseDropIndex(p)
	case p.PeekKeyword("ROLE"):
		return parseDropRole(p)
	case p.PeekKeyword("DATABASE"):
		return parseDropDatabase(p)
	case p.PeekKeyword("SCHEMA"):
		return parseDropSchema(p)
	case p.PeekKeyword("SEQUENCE"):
		return parseDropSequence(p)
	case p.PeekKeyword("FUNCTION"):
		return parseDropFunction(p)
	default:
		return nil, p.ExpectedRef("TABLE, VIEW, INDEX, ROLE, DATABASE, SCHEMA, SEQUENCE, or FUNCTION after DROP", p.PeekTokenRef())
	}
}

// parseDropView parses DROP VIEW
// Reference: src/parser/mod.rs parse_drop (VIEW branch)
func parseDropView(p *Parser, temporary bool) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("VIEW"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated view names
	names, err := parseCommaSeparatedObjectNames(p)
	if err != nil {
		return nil, err
	}

	// Parse optional CASCADE/RESTRICT
	cascade := p.ParseKeyword("CASCADE")
	restrict := p.ParseKeyword("RESTRICT")

	return &statement.DropView{
		IfExists: ifExists,
		Names:    names,
		Cascade:  cascade,
		Restrict: restrict,
	}, nil
}

// parseDropTable parses DROP TABLE
func parseDropTable(p *Parser, temporary bool) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("TABLE"); err != nil {
		return nil, err
	}

	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	names, err := parseCommaSeparatedObjectNames(p)
	if err != nil {
		return nil, err
	}

	cascade := p.ParseKeyword("CASCADE")
	restrict := p.ParseKeyword("RESTRICT")

	return &statement.DropTable{
		IfExists: ifExists,
		Names:    names,
		Cascade:  cascade,
		Restrict: restrict,
	}, nil
}

// parseCommaSeparatedObjectNames parses a comma-separated list of object names
func parseCommaSeparatedObjectNames(p *Parser) ([]*ast.ObjectName, error) {
	var names []*ast.ObjectName
	for {
		name, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		names = append(names, name)
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}
	return names, nil
}

// parseCommaSeparatedIdents parses a comma-separated list of identifiers
func parseCommaSeparatedIdents(p *Parser) ([]*ast.Ident, error) {
	var names []*ast.Ident
	for {
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		names = append(names, name)
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}
	return names, nil
}

// parseDropIndex parses DROP INDEX
// Reference: src/parser/mod.rs parse_drop (INDEX branch)
// DROP INDEX [CONCURRENTLY] [IF EXISTS] name [, ...] [CASCADE | RESTRICT]
func parseDropIndex(p *Parser) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("INDEX"); err != nil {
		return nil, err
	}

	// Parse CONCURRENTLY (PostgreSQL specific)
	concurrently := p.ParseKeyword("CONCURRENTLY")

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated list of index names
	var names []*ast.ObjectName
	for {
		name, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		names = append(names, name)

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	// Parse optional CASCADE or RESTRICT
	cascade := p.ParseKeyword("CASCADE")
	restrict := p.ParseKeyword("RESTRICT")

	if cascade && restrict {
		return nil, fmt.Errorf("cannot specify both CASCADE and RESTRICT in DROP INDEX")
	}

	return &statement.DropIndex{
		Concurrently: concurrently,
		IfExists:     ifExists,
		Names:        names,
		Cascade:      cascade,
		Restrict:     restrict,
	}, nil
}

func parseDropRole(p *Parser) (ast.Statement, error) {
	// ROLE keyword is expected (already checked by caller)
	if _, err := p.ExpectKeyword("ROLE"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated role names (identifiers) - convert to ObjectNames
	names := make([]*ast.ObjectName, 0)
	for {
		ident, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		names = append(names, ast.NewObjectNameFromIdents(ident))
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	// Parse optional CASCADE/RESTRICT
	cascade := p.ParseKeyword("CASCADE")
	restrict := p.ParseKeyword("RESTRICT")

	return &statement.Drop{
		ObjectType: expr.ObjectTypeRole,
		IfExists:   ifExists,
		Names:      names,
		Cascade:    cascade,
		Restrict:   restrict,
	}, nil
}

func parseDropDatabase(p *Parser) (ast.Statement, error) {
	return nil, p.ExpectedRef("DROP DATABASE not yet implemented", p.PeekTokenRef())
}

func parseDropSchema(p *Parser) (ast.Statement, error) {
	// SCHEMA keyword is expected (already checked by caller)
	if _, err := p.ExpectKeyword("SCHEMA"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated schema names
	names, err := parseCommaSeparatedObjectNames(p)
	if err != nil {
		return nil, err
	}

	// Parse optional CASCADE/RESTRICT
	cascade := p.ParseKeyword("CASCADE")
	restrict := p.ParseKeyword("RESTRICT")

	return &statement.DropSchema{
		IfExists: ifExists,
		Names:    names,
		Cascade:  cascade,
		Restrict: restrict,
	}, nil
}

func parseDropSequence(p *Parser) (ast.Statement, error) {
	// Consume SEQUENCE keyword (already checked by caller)
	if _, err := p.ExpectKeyword("SEQUENCE"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated list of sequence names
	var names []*ast.ObjectName
	for {
		name, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		names = append(names, name)

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	// Parse optional CASCADE or RESTRICT
	cascade := p.ParseKeyword("CASCADE")
	restrict := p.ParseKeyword("RESTRICT")

	if cascade && restrict {
		return nil, fmt.Errorf("cannot specify both CASCADE and RESTRICT")
	}

	return &statement.DropSequence{
		IfExists: ifExists,
		Names:    names,
		Cascade:  cascade,
		Restrict: restrict,
	}, nil
}

func parseDropFunction(p *Parser) (ast.Statement, error) {
	return nil, p.ExpectedRef("DROP FUNCTION not yet implemented", p.PeekTokenRef())
}

// ParseTruncate parses TRUNCATE statements
// Reference: src/parser/mod.rs:1091-1139
func ParseTruncate(p *Parser) (ast.Statement, error) {
	// Parse optional TABLE keyword
	_ = p.ParseKeyword("TABLE")

	// Parse optional IF EXISTS
	_ = p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated table names
	var tableNames []*ast.ObjectName
	for {
		name, err := p.ParseObjectName()
		if err != nil {
			return nil, fmt.Errorf("expected table name in TRUNCATE: %w", err)
		}
		tableNames = append(tableNames, name)

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	if len(tableNames) == 0 {
		return nil, fmt.Errorf("expected at least one table name in TRUNCATE")
	}

	// Parse optional PARTITION clause
	var partitions []expr.Expr
	if p.ParseKeyword("PARTITION") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, fmt.Errorf("expected ( after PARTITION: %w", err)
		}
		ep := NewExpressionParser(p)
		for {
			if _, ok := p.PeekToken().Token.(token.TokenRParen); ok {
				break
			}
			partExpr, err := ep.ParseExpr()
			if err != nil {
				return nil, fmt.Errorf("expected partition expression: %w", err)
			}
			partitions = append(partitions, partExpr)
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, fmt.Errorf("expected ) after partition expressions: %w", err)
		}
	}

	// Parse optional ON CLUSTER (ClickHouse)
	var onCluster *ast.Ident
	if p.ParseKeyword("ON") {
		if !p.ParseKeyword("CLUSTER") {
			return nil, fmt.Errorf("expected CLUSTER after ON in TRUNCATE")
		}
		clusterName, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected cluster name after ON CLUSTER: %w", err)
		}
		onCluster = clusterName
	}

	// TODO: Parse RESTART IDENTITY / CONTINUE IDENTITY / CASCADE / RESTRICT
	// These are dialect-specific (PostgreSQL)

	return &statement.Truncate{
		TableNames: tableNames,
		Partitions: partitions,
		OnCluster:  onCluster,
	}, nil
}
