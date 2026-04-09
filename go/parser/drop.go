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
	"strings"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/token"
)

// parseDrop parses DROP statements
// Reference: src/parser/mod.rs parse_drop
func parseDrop(p *Parser) (ast.Statement, error) {
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
	case p.PeekKeyword("TRIGGER"):
		return parseDropTrigger(p)
	case p.PeekKeyword("OPERATOR"):
		// Check if this is DROP OPERATOR FAMILY or DROP OPERATOR CLASS
		p.NextToken() // consume OPERATOR
		if p.PeekKeyword("FAMILY") {
			return parseDropOperatorFamily(p)
		} else if p.PeekKeyword("CLASS") {
			return parseDropOperatorClass(p)
		}
		// Put back OPERATOR token and parse as regular DROP OPERATOR
		p.PrevToken()
		return parseDropOperator(p)
	case p.PeekKeyword("STAGE"):
		return parseDropStage(p)
	case p.PeekKeyword("USER"):
		return parseDropUser(p)
	case p.PeekKeyword("STREAM"):
		return parseDropStream(p)
	case p.PeekKeyword("POLICY"):
		return parseDropPolicy(p)
	case p.PeekKeyword("CONNECTOR"):
		return parseDropConnector(p)
	case p.PeekKeyword("EXTENSION"):
		return parseDropExtension(p)
	case p.PeekKeyword("DOMAIN"):
		return parseDropDomain(p)
	case p.PeekKeyword("PROCEDURE"):
		return parseDropProcedure(p)
	default:
		return nil, p.ExpectedRef("TABLE, VIEW, INDEX, ROLE, DATABASE, SCHEMA, SEQUENCE, FUNCTION, TRIGGER, OPERATOR, STAGE, USER, STREAM, POLICY, CONNECTOR, EXTENSION, DOMAIN, or PROCEDURE after DROP", p.PeekTokenRef())
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
		Temporary: temporary,
		IfExists:  ifExists,
		Names:     names,
		Cascade:   cascade,
		Restrict:  restrict,
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
// MySQL: DROP INDEX [IF EXISTS] name ON tbl_name
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

	// MySQL: DROP INDEX ... ON table_name
	var tableName *ast.ObjectName
	if p.ParseKeyword("ON") {
		tblName, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		tableName = tblName
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
		OnTable:      tableName,
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
	// DATABASE keyword is already consumed by caller
	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated database names
	names, err := parseCommaSeparatedObjectNames(p)
	if err != nil {
		return nil, err
	}

	// Parse optional CASCADE/RESTRICT
	cascade := p.ParseKeyword("CASCADE")
	restrict := p.ParseKeyword("RESTRICT")

	return &statement.Drop{
		ObjectType: expr.ObjectTypeDatabase,
		IfExists:   ifExists,
		Names:      names,
		Cascade:    cascade,
		Restrict:   restrict,
	}, nil
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
	// DROP FUNCTION [ IF EXISTS ] function_name [ ( [ [ argmode ] [ argname ] argtype [, ...] ] ) ] [, ...]
	// [ CASCADE | RESTRICT ]

	// Consume FUNCTION keyword
	if _, err := p.ExpectKeyword("FUNCTION"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse function descriptions (comma-separated)
	var funcDescs []*expr.FunctionDesc
	for {
		funcDesc, err := parseFunctionDesc(p)
		if err != nil {
			return nil, err
		}
		funcDescs = append(funcDescs, funcDesc)

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	// Parse optional CASCADE or RESTRICT
	var dropBehavior *expr.DropBehavior
	if p.ParseKeyword("CASCADE") {
		behavior := expr.DropBehaviorCascade
		dropBehavior = &behavior
	} else if p.ParseKeyword("RESTRICT") {
		behavior := expr.DropBehaviorRestrict
		dropBehavior = &behavior
	}

	return &statement.DropFunction{
		IfExists:     ifExists,
		FuncDesc:     funcDescs,
		DropBehavior: dropBehavior,
	}, nil
}

// parseFunctionDesc parses a function/procedure description with optional argument list
// Reference: src/parser/mod.rs parse_function_desc, parse_function_arg
// Syntax: name [ ( [ [ argmode ] [ argname ] argtype [ = default ] [, ...] ] ) ]
func parseFunctionDesc(p *Parser) (*expr.FunctionDesc, error) {
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	funcDesc := &expr.FunctionDesc{
		Name:      name,
		Args:      nil, // No args if no parentheses
		HasParens: false,
	}

	// Check for optional argument list
	if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
		funcDesc.HasParens = true
		p.AdvanceToken() // consume (

		// Check for empty argument list: ()
		if _, ok := p.PeekToken().Token.(token.TokenRParen); ok {
			p.AdvanceToken() // consume )
			funcDesc.Args = []expr.Expr{}
		} else {
			// Parse argument list
			var args []expr.Expr
			for {
				// For DROP FUNCTION/PROCEDURE, we parse function arguments
				// The syntax is: [ IN | OUT | INOUT ] [ argname ] argtype [ = default | DEFAULT expr ]
				arg, err := parseFunctionArgForDrop(p)
				if err != nil {
					return nil, err
				}
				args = append(args, arg)

				if !p.ConsumeToken(token.TokenComma{}) {
					break
				}
			}

			// Expect closing )
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}

			funcDesc.Args = args
		}
	}

	return funcDesc, nil
}

// parseFunctionArgForDrop parses a function argument for DROP FUNCTION/PROCEDURE
// Reference: src/parser/mod.rs parse_function_arg
// Syntax: [ IN | OUT | INOUT ] [ argname ] argtype [ = default | DEFAULT expr ]
func parseFunctionArgForDrop(p *Parser) (expr.Expr, error) {
	// Check for argmode (IN, OUT, INOUT)
	argMode := ""
	if p.ParseKeyword("IN") {
		// Could be IN or INOUT - check further
		if p.ParseKeyword("OUT") {
			argMode = "INOUT"
		} else {
			argMode = "IN"
		}
	} else if p.ParseKeyword("OUT") {
		argMode = "OUT"
	} else if p.ParseKeyword("INOUT") {
		argMode = "INOUT"
	}

	// Collect all tokens that make up this argument until we hit comma or closing paren
	var argParts []string

	for {
		tok := p.PeekToken().Token

		// Check for end of argument
		if _, ok := tok.(token.TokenRParen); ok {
			break
		}
		if _, ok := tok.(token.TokenComma); ok {
			break
		}

		// Handle different token types
		switch t := tok.(type) {
		case token.TokenWord:
			// Check for DEFAULT keyword
			if t.Word.String() == "DEFAULT" {
				argParts = append(argParts, "DEFAULT")
				p.AdvanceToken()
				// Parse the default expression
				ep := NewExpressionParser(p)
				if defExpr, err := ep.ParseExpr(); err == nil {
					argParts = append(argParts, defExpr.String())
				}
				break
			}
			argParts = append(argParts, t.Word.String())
			p.AdvanceToken()
		case token.TokenChar:
			if t.Char == '=' {
				argParts = append(argParts, "=")
				p.AdvanceToken()
				// Parse the default expression after =
				ep := NewExpressionParser(p)
				if defExpr, err := ep.ParseExpr(); err == nil {
					argParts = append(argParts, defExpr.String())
				}
				break
			}
			argParts = append(argParts, string(t.Char))
			p.AdvanceToken()
		default:
			// For any other token, add its string representation
			argParts = append(argParts, tok.String())
			p.AdvanceToken()
		}
	}

	// Build the full argument string
	if argMode != "" && len(argParts) > 0 {
		// Insert mode at the beginning
		argParts = append([]string{argMode}, argParts...)
	}

	argStr := strings.Join(argParts, " ")
	return &expr.Identifier{
		SpanVal: token.Span{},
		Ident:   &expr.Ident{Value: argStr},
	}, nil
}

// parseTruncate parses TRUNCATE statements
// Reference: src/parser/mod.rs:1091-1139
func parseTruncate(p *Parser) (ast.Statement, error) {
	// Parse optional TABLE keyword
	hasTableKeyword := p.ParseKeyword("TABLE")

	// Parse optional IF EXISTS
	hasIfExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated table targets (with optional ONLY and asterisk for PostgreSQL)
	var tableTargets []*statement.TruncateTableTarget
	for {
		// Check for ONLY keyword (PostgreSQL)
		only := p.ParseKeyword("ONLY")

		name, err := p.ParseObjectName()
		if err != nil {
			return nil, fmt.Errorf("expected table name in TRUNCATE: %w", err)
		}

		// Check for asterisk after table name (PostgreSQL descendant tables)
		hasAsterisk := p.ConsumeToken(token.TokenMul{})

		tableTargets = append(tableTargets, &statement.TruncateTableTarget{
			Name:        name,
			Only:        only,
			HasAsterisk: hasAsterisk,
		})

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	if len(tableTargets) == 0 {
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

	// Parse PostgreSQL-specific options: RESTART IDENTITY | CONTINUE IDENTITY
	var identityOpt statement.TruncateIdentityOption
	if p.ParseKeywords([]string{"RESTART", "IDENTITY"}) {
		identityOpt = statement.TruncateIdentityRestart
	} else if p.ParseKeywords([]string{"CONTINUE", "IDENTITY"}) {
		identityOpt = statement.TruncateIdentityContinue
	}

	// Parse PostgreSQL-specific options: CASCADE | RESTRICT
	var cascadeOpt statement.TruncateCascadeOption
	if p.ParseKeyword("CASCADE") {
		cascadeOpt = statement.TruncateCascadeCascade
	} else if p.ParseKeyword("RESTRICT") {
		cascadeOpt = statement.TruncateCascadeRestrict
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

	return &statement.Truncate{
		TableNames: tableTargets,
		Partitions: partitions,
		OnCluster:  onCluster,
		Table:      hasTableKeyword,
		IfExists:   hasIfExists,
		Identity:   identityOpt,
		Cascade:    cascadeOpt,
	}, nil
}

// parseDropTrigger parses DROP TRIGGER
// Reference: src/parser/mod.rs parse_drop_trigger
func parseDropTrigger(p *Parser) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("TRIGGER"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse trigger name
	triggerName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse optional ON table_name
	var tableName *ast.ObjectName
	if p.ParseKeyword("ON") {
		tn, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		tableName = tn
	}

	// Parse optional CASCADE or RESTRICT
	var dropBehavior *expr.DropBehavior
	if p.ParseKeyword("CASCADE") {
		behavior := expr.DropBehaviorCascade
		dropBehavior = &behavior
	} else if p.ParseKeyword("RESTRICT") {
		behavior := expr.DropBehaviorRestrict
		dropBehavior = &behavior
	}

	return &statement.DropTrigger{
		IfExists:     ifExists,
		Name:         triggerName,
		TableName:    tableName,
		DropBehavior: dropBehavior,
	}, nil
}

// parseDropOperator parses DROP OPERATOR
// Reference: src/parser/mod.rs parse_drop_operator
func parseDropOperator(p *Parser) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("OPERATOR"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated operator signatures
	var signatures []*expr.DropOperatorSignature
	for {
		// Parse operator name - can be an identifier or an operator symbol (e.g., @@, <, >, ~)
		name, err := p.ParseOperatorName()
		if err != nil {
			return nil, err
		}

		// Parse operator signature: (type1 [, type2])
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}

		var argTypes []string
		for {
			if _, ok := p.PeekToken().Token.(token.TokenRParen); ok {
				break
			}
			// Parse data type as identifier
			dt, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			argTypes = append(argTypes, dt.Value)
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}

		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}

		signatures = append(signatures, &expr.DropOperatorSignature{
			Name:     name,
			ArgTypes: argTypes,
		})

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	// Parse optional CASCADE/RESTRICT
	var dropBehavior *expr.DropBehavior
	if p.ParseKeyword("CASCADE") {
		cascade := expr.DropBehaviorCascade
		dropBehavior = &cascade
	} else if p.ParseKeyword("RESTRICT") {
		restrict := expr.DropBehaviorRestrict
		dropBehavior = &restrict
	}

	return &statement.DropOperator{
		IfExists:     ifExists,
		Names:        signatures,
		DropBehavior: dropBehavior,
	}, nil
}

// parseDropOperatorFamily parses DROP OPERATOR FAMILY
// Reference: src/parser/mod.rs parse_drop_operator_family
// DROP OPERATOR FAMILY [ IF EXISTS ] name [, ...] USING index_method [ CASCADE | RESTRICT ]
func parseDropOperatorFamily(p *Parser) (ast.Statement, error) {
	// OPERATOR keyword was already consumed
	if _, err := p.ExpectKeyword("FAMILY"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated operator family names
	names, err := parseCommaSeparatedObjectNames(p)
	if err != nil {
		return nil, err
	}

	// Expect USING keyword
	if _, err := p.ExpectKeyword("USING"); err != nil {
		return nil, err
	}

	// Parse index method (btree, hash, gist, gin, etc.)
	indexMethod, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Parse optional CASCADE or RESTRICT
	var dropBehavior *expr.DropBehavior
	if p.ParseKeyword("CASCADE") {
		behavior := expr.DropBehaviorCascade
		dropBehavior = &behavior
	} else if p.ParseKeyword("RESTRICT") {
		behavior := expr.DropBehaviorRestrict
		dropBehavior = &behavior
	}

	return &statement.DropOperatorFamily{
		IfExists:     ifExists,
		Names:        names,
		IndexMethod:  indexMethod,
		DropBehavior: dropBehavior,
	}, nil
}

// parseDropOperatorClass parses DROP OPERATOR CLASS
// Reference: src/parser/mod.rs parse_drop_operator_class
// DROP OPERATOR CLASS [ IF EXISTS ] name [, ...] USING index_method [ CASCADE | RESTRICT ]
func parseDropOperatorClass(p *Parser) (ast.Statement, error) {
	// OPERATOR keyword was already consumed
	if _, err := p.ExpectKeyword("CLASS"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated operator class names
	names, err := parseCommaSeparatedObjectNames(p)
	if err != nil {
		return nil, err
	}

	// Expect USING keyword
	if _, err := p.ExpectKeyword("USING"); err != nil {
		return nil, err
	}

	// Parse index method (btree, hash, gist, gin, etc.)
	indexMethod, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Parse optional CASCADE or RESTRICT
	var dropBehavior *expr.DropBehavior
	if p.ParseKeyword("CASCADE") {
		behavior := expr.DropBehaviorCascade
		dropBehavior = &behavior
	} else if p.ParseKeyword("RESTRICT") {
		behavior := expr.DropBehaviorRestrict
		dropBehavior = &behavior
	}

	return &statement.DropOperatorClass{
		IfExists:     ifExists,
		Names:        names,
		IndexMethod:  indexMethod,
		DropBehavior: dropBehavior,
	}, nil
}

// parseDropStage parses DROP STAGE (Snowflake-specific)
// Reference: Snowflake documentation
func parseDropStage(p *Parser) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("STAGE"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse stage name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	return &statement.DropStage{
		IfExists: ifExists,
		Name:     name,
	}, nil
}

// parseDropUser parses DROP USER statement
// Reference: src/parser/mod.rs parse_drop (USER branch)
func parseDropUser(p *Parser) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("USER"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated user names (identifiers) - convert to ObjectNames
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

	return &statement.Drop{
		ObjectType: expr.ObjectTypeUser,
		IfExists:   ifExists,
		Names:      names,
	}, nil
}

// parseDropStream parses DROP STREAM statement
// Reference: src/parser/mod.rs parse_drop (STREAM branch)
func parseDropStream(p *Parser) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("STREAM"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse stream name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	return &statement.Drop{
		ObjectType: expr.ObjectTypeStream,
		IfExists:   ifExists,
		Names:      []*ast.ObjectName{name},
	}, nil
}

// parseDropPolicy parses DROP POLICY statement
// Reference: src/parser/mod.rs parse_drop_policy
func parseDropPolicy(p *Parser) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("POLICY"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse policy name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Expect ON keyword
	if _, err := p.ExpectKeyword("ON"); err != nil {
		return nil, err
	}

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse optional CASCADE/RESTRICT
	var dropBehavior *expr.DropBehavior
	if p.ParseKeyword("CASCADE") {
		db := expr.DropBehaviorCascade
		dropBehavior = &db
	} else if p.ParseKeyword("RESTRICT") {
		db := expr.DropBehaviorRestrict
		dropBehavior = &db
	}

	return &statement.DropPolicy{
		IfExists:     ifExists,
		Name:         name,
		TableName:    tableName,
		DropBehavior: dropBehavior,
	}, nil
}

// parseDropConnector parses DROP CONNECTOR statement
func parseDropConnector(p *Parser) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("CONNECTOR"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse connector name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	return &statement.DropConnector{
		IfExists: ifExists,
		Name:     name,
	}, nil
}

// parseDropExtension parses DROP EXTENSION statement
// Reference: src/parser/mod.rs:8053-8069
func parseDropExtension(p *Parser) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("EXTENSION"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated extension names
	names, err := parseCommaSeparatedIdents(p)
	if err != nil {
		return nil, err
	}

	// Parse optional CASCADE or RESTRICT
	var dropBehavior *expr.DropBehavior
	if p.ParseKeyword("CASCADE") {
		b := expr.DropBehaviorCascade
		dropBehavior = &b
	} else if p.ParseKeyword("RESTRICT") {
		b := expr.DropBehaviorRestrict
		dropBehavior = &b
	}

	return &statement.DropExtension{
		IfExists:     ifExists,
		Names:        names,
		DropBehavior: dropBehavior,
	}, nil
}

// parseDropDomain parses DROP DOMAIN statement
// Reference: src/parser/mod.rs parse_drop_domain
// DROP DOMAIN [ IF EXISTS ] name [, ...] [ CASCADE | RESTRICT ]
func parseDropDomain(p *Parser) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("DOMAIN"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse comma-separated domain names
	names, err := parseCommaSeparatedObjectNames(p)
	if err != nil {
		return nil, err
	}

	// Parse optional CASCADE or RESTRICT
	var dropBehavior *expr.DropBehavior
	if p.ParseKeyword("CASCADE") {
		b := expr.DropBehaviorCascade
		dropBehavior = &b
	} else if p.ParseKeyword("RESTRICT") {
		b := expr.DropBehaviorRestrict
		dropBehavior = &b
	}

	return &statement.DropDomain{
		IfExists:     ifExists,
		Names:        names,
		DropBehavior: dropBehavior,
	}, nil
}

// parseDropProcedure parses DROP PROCEDURE statement
// Reference: src/parser/mod.rs parse_drop_procedure
// DROP PROCEDURE [ IF EXISTS ] name [ ( [ [ argmode ] [ argname ] argtype [, ...] ] ) ] [, ...] [ CASCADE | RESTRICT ]
func parseDropProcedure(p *Parser) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("PROCEDURE"); err != nil {
		return nil, err
	}

	// Parse IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse procedure descriptions (comma-separated, similar to DROP FUNCTION)
	var procDescs []*expr.FunctionDesc
	for {
		funcDesc, err := parseFunctionDesc(p)
		if err != nil {
			return nil, err
		}
		procDescs = append(procDescs, funcDesc)

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	// Parse optional CASCADE or RESTRICT
	var dropBehavior *expr.DropBehavior
	if p.ParseKeyword("CASCADE") {
		b := expr.DropBehaviorCascade
		dropBehavior = &b
	} else if p.ParseKeyword("RESTRICT") {
		b := expr.DropBehaviorRestrict
		dropBehavior = &b
	}

	return &statement.DropProcedure{
		IfExists:     ifExists,
		ProcDesc:     procDescs,
		DropBehavior: dropBehavior,
	}, nil
}
