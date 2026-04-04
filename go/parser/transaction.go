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
	"github.com/user/sqlparser/tokenizer"
)

// ParseStartTransaction parses START TRANSACTION statements
// Reference: src/parser/mod.rs:18612-18623
func ParseStartTransaction(p *Parser) (ast.Statement, error) {
	// Expect TRANSACTION keyword after START
	if !p.ParseKeyword("TRANSACTION") {
		return nil, fmt.Errorf("Expected TRANSACTION after START")
	}

	modes, err := parseTransactionModes(p)
	if err != nil {
		return nil, err
	}

	return &statement.StartTransaction{
		Modes: modes,
		Begin: false,
	}, nil
}

// ParseBegin parses BEGIN TRANSACTION statements
// Reference: src/parser/mod.rs:18645-18664
func ParseBegin(p *Parser) (ast.Statement, error) {
	// Parse optional transaction modifier (DEFERRED, IMMEDIATE, EXCLUSIVE, etc.)
	modifier := parseTransactionModifier(p)

	// Parse optional TRANSACTION/WORK/TRAN keyword
	var transaction *expr.BeginTransactionKind
	if p.ParseKeyword("TRANSACTION") {
		kind := expr.BeginTransactionKindTransaction
		transaction = &kind
	} else if p.ParseKeyword("WORK") {
		kind := expr.BeginTransactionKindWork
		transaction = &kind
	} else if p.ParseKeyword("TRAN") {
		kind := expr.BeginTransactionKindTran
		transaction = &kind
	}

	modes, err := parseTransactionModes(p)
	if err != nil {
		return nil, err
	}

	return &statement.StartTransaction{
		Modes:       modes,
		Begin:       true,
		Transaction: transaction,
		Modifier:    modifier,
	}, nil
}

// parseTransactionModifier parses an optional transaction modifier
// Reference: src/parser/mod.rs:18626-18642
func parseTransactionModifier(p *Parser) *expr.TransactionModifier {
	dialect := p.GetDialect()
	if !dialect.SupportsStartTransactionModifier() {
		return nil
	}

	var modifier expr.TransactionModifier
	if p.ParseKeyword("DEFERRED") {
		modifier = expr.TransactionModifierDeferred
		return &modifier
	} else if p.ParseKeyword("IMMEDIATE") {
		modifier = expr.TransactionModifierImmediate
		return &modifier
	} else if p.ParseKeyword("EXCLUSIVE") {
		modifier = expr.TransactionModifierExclusive
		return &modifier
	} else if p.ParseKeyword("TRY") {
		modifier = expr.TransactionModifierTry
		return &modifier
	} else if p.ParseKeyword("CATCH") {
		modifier = expr.TransactionModifierCatch
		return &modifier
	}
	return nil
}

// parseTransactionModes parses transaction modes like ISOLATION LEVEL, READ ONLY, etc.
// Reference: src/parser/mod.rs:18731-18767
func parseTransactionModes(p *Parser) ([]*expr.TransactionMode, error) {
	var modes []*expr.TransactionMode
	required := false

	for {
		var mode *expr.TransactionMode

		if p.PeekKeyword("ISOLATION") {
			// ISOLATION LEVEL <level>
			p.ParseKeyword("ISOLATION")
			if !p.ParseKeyword("LEVEL") {
				return nil, fmt.Errorf("Expected LEVEL after ISOLATION")
			}
			var isoLevel expr.TransactionMode
			if p.PeekKeyword("READ") {
				p.ParseKeyword("READ")
				if p.ParseKeyword("UNCOMMITTED") {
					isoLevel = expr.TransactionModeReadUncommitted
				} else if p.ParseKeyword("COMMITTED") {
					isoLevel = expr.TransactionModeReadCommitted
				} else {
					return nil, fmt.Errorf("Expected UNCOMMITTED or COMMITTED after READ")
				}
			} else if p.PeekKeyword("REPEATABLE") {
				p.ParseKeyword("REPEATABLE")
				if !p.ParseKeyword("READ") {
					return nil, fmt.Errorf("Expected READ after REPEATABLE")
				}
				isoLevel = expr.TransactionModeRepeatableRead
			} else if p.ParseKeyword("SERIALIZABLE") {
				isoLevel = expr.TransactionModeSerializable
			} else if p.ParseKeyword("SNAPSHOT") {
				isoLevel = expr.TransactionModeSnapshot
			} else {
				return nil, fmt.Errorf("Expected isolation level")
			}
			mode = &isoLevel
		} else if p.PeekKeyword("READ") {
			p.ParseKeyword("READ")
			if p.ParseKeyword("ONLY") {
				m := expr.TransactionModeReadOnly
				mode = &m
			} else if p.ParseKeyword("WRITE") {
				m := expr.TransactionModeReadWrite
				mode = &m
			} else {
				return nil, fmt.Errorf("Expected ONLY or WRITE after READ")
			}
		} else if required {
			return nil, fmt.Errorf("Expected transaction mode")
		} else {
			break
		}

		modes = append(modes, mode)

		// Optional comma between modes (PostgreSQL doesn't require it, but ANSI does)
		required = p.ConsumeToken(tokenizer.TokenComma{})
	}

	return modes, nil
}

// ParseCommit parses COMMIT statements
// Reference: src/parser/mod.rs:18770-18776
func ParseCommit(p *Parser) (ast.Statement, error) {
	chain, err := parseCommitRollbackChain(p)
	if err != nil {
		return nil, err
	}

	return &statement.Commit{
		Chain: chain,
		End:   false,
	}, nil
}

// ParseRollback parses ROLLBACK statements
// Reference: src/parser/mod.rs:18779-18784
func ParseRollback(p *Parser) (ast.Statement, error) {
	chain, err := parseCommitRollbackChain(p)
	if err != nil {
		return nil, err
	}

	savepoint, err := parseRollbackSavepoint(p)
	if err != nil {
		return nil, err
	}

	return &statement.Rollback{
		Chain:     chain,
		Savepoint: savepoint,
	}, nil
}

// parseCommitRollbackChain parses optional AND [NO] CHAIN clause
// Reference: src/parser/mod.rs:18787-18796
func parseCommitRollbackChain(p *Parser) (bool, error) {
	// Skip optional TRANSACTION/WORK/TRAN keywords
	if p.ParseKeyword("TRANSACTION") || p.ParseKeyword("WORK") || p.ParseKeyword("TRAN") {
		// Consumed
	}

	if p.ParseKeyword("AND") {
		chain := !p.ParseKeyword("NO")
		if !p.ParseKeyword("CHAIN") {
			return false, fmt.Errorf("Expected CHAIN after AND")
		}
		return chain, nil
	}
	return false, nil
}

// parseRollbackSavepoint parses optional TO SAVEPOINT clause
// Reference: src/parser/mod.rs:18799-18808
func parseRollbackSavepoint(p *Parser) (*ast.Ident, error) {
	if p.ParseKeyword("TO") {
		// Optional SAVEPOINT keyword
		p.ParseKeyword("SAVEPOINT")
		return p.ParseIdentifier()
	}
	return nil, nil
}

// ParseSavepoint parses SAVEPOINT statements
// Reference: src/parser/mod.rs:1430-1433
func ParseSavepoint(p *Parser) (ast.Statement, error) {
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	return &statement.Savepoint{
		Name: name,
	}, nil
}

// ParseRelease parses RELEASE SAVEPOINT statements
// Reference: src/parser/mod.rs:1436-1441
func ParseRelease(p *Parser) (ast.Statement, error) {
	// Optional SAVEPOINT keyword
	p.ParseKeyword("SAVEPOINT")

	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	return &statement.ReleaseSavepoint{
		Name: name,
	}, nil
}
