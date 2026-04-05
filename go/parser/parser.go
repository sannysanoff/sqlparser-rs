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

// Package parser provides the core SQL parser implementation.
// This is the Go port of the sqlparser-rs parser module.
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/datatype"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/errors"
	"github.com/user/sqlparser/parseriface"
	"github.com/user/sqlparser/token"
)

// Compile-time interface check: ensure Parser implements parseriface.Parser
var _ parseriface.Parser = (*Parser)(nil)

// DefaultRemainingDepth is the default maximum recursion depth
const DefaultRemainingDepth = 50

// eofToken is a constant EOF token that can be referenced.
var eofToken = token.TokenWithSpan{
	Token: token.EOF{},
	Span:  token.Span{},
}

// Parser is the main SQL parser struct.
// It maintains state for parsing SQL statements from a token stream.
type Parser struct {
	// tokens is the slice of tokens to parse
	tokens []token.TokenWithSpan

	// index is the index of the first unprocessed token in tokens.
	// The "current" token is at index - 1
	// The "next" token is at index
	// The "previous" token is at index - 2
	index int

	// state is the current state of the parser
	state parseriface.ParserState

	// dialect is the SQL dialect to use for parsing
	dialect parseriface.CompleteDialect

	// options controls parser behavior
	options parseriface.ParserOptions

	// recursionCounter prevents stack overflow by limiting recursion depth
	recursionCounter RecursionCounter
}

// New creates a new Parser for the given dialect.
//
// Example:
//
//	dialect := dialects.NewGenericDialect()
//	parser := parser.New(dialect)
//	stmts, err := parser.TryWithSQL("SELECT * FROM foo").ParseStatements()
func New(dialect parseriface.CompleteDialect) *Parser {
	return &Parser{
		tokens:           make([]token.TokenWithSpan, 0),
		index:            0,
		state:            parseriface.StateNormal,
		dialect:          dialect,
		recursionCounter: NewRecursionCounter(DefaultRemainingDepth),
		options: NewParserOptions(
			WithTrailingCommas(dialect.SupportsTrailingCommas()),
			WithUnescape(true),
		),
	}
}

// WithRecursionLimit sets the maximum recursion limit while parsing.
// The parser prevents stack overflows by returning an error if the parser
// exceeds this depth while processing the query.
func (p *Parser) WithRecursionLimit(recursionLimit int) *Parser {
	p.recursionCounter = NewRecursionCounter(recursionLimit)
	return p
}

// WithOptions sets additional parser options.
func (p *Parser) WithOptions(options parseriface.ParserOptions) *Parser {
	p.options = options
	return p
}

// WithTokens resets this parser to parse the specified token stream.
func (p *Parser) WithTokens(tokens []token.TokenWithSpan) *Parser {
	p.tokens = tokens
	p.index = 0
	return p
}

// TryWithSQL tokenizes the SQL string and sets this parser's state to
// parse the resulting tokens. Returns an error if there was an error tokenizing the SQL string.
func (p *Parser) TryWithSQL(sql string) (*Parser, error) {
	// Wrap the dialects.Dialect in an adapter that implements token.Dialect
	tokDialect := newDialectAdapter(p.dialect)
	tok := token.NewTokenizer(tokDialect, sql)
	tokens, err := tok.TokenizeWithSpan()
	if err != nil {
		return nil, err
	}
	p.tokens = tokens
	p.index = 0
	return p, nil
}

// ParseSQL is a convenience method to parse a string with one or more SQL
// statements into an Abstract Syntax Tree (AST).
//
// Example:
//
//	dialect := dialects.NewGenericDialect()
//	statements, err := parser.ParseSQL(dialect, "SELECT * FROM foo")
//	// statements will contain 1 Statement
func ParseSQL(dialect parseriface.CompleteDialect, sql string) ([]ast.Statement, error) {
	p, err := New(dialect).TryWithSQL(sql)
	if err != nil {
		return nil, err
	}
	return p.ParseStatements()
}

// ParseStatements parses potentially multiple statements from the token stream.
//
// Example:
//
//	dialect := dialects.NewGenericDialect()
//	parser := parser.New(dialect)
//	stmts, err := parser.TryWithSQL("SELECT * FROM foo; SELECT * FROM bar;").ParseStatements()
//	// stmts will contain 2 Statements
func (p *Parser) ParseStatements() ([]ast.Statement, error) {
	var stmts []ast.Statement
	expectingStatementDelimiter := false

	for {
		// ignore empty statements (between successive statement delimiters)
		for p.ConsumeToken(token.TokenSemiColon{}) {
			expectingStatementDelimiter = false
		}

		if !p.options.RequireSemicolon {
			expectingStatementDelimiter = false
		}

		nextTok := p.PeekTokenRef()
		switch t := nextTok.Token.(type) {
		case token.EOF:
			return stmts, nil
		default:
			_ = t // silence unused variable warning for now
		}

		if expectingStatementDelimiter {
			// Check for END keyword as a statement terminator
			if wordTok, ok := nextTok.Token.(token.TokenWord); ok && wordTok.Word.Keyword == "END" {
				return stmts, nil
			}
			return nil, p.expectedRef("end of statement", nextTok)
		}

		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
		expectingStatementDelimiter = true
	}
}

// ParseStatement parses a single top-level statement (such as SELECT, INSERT, CREATE, etc.),
// stopping before the statement separator, if any.
func (p *Parser) ParseStatement() (ast.Statement, error) {
	if err := p.recursionCounter.TryDecrease(); err != nil {
		return nil, err
	}
	defer p.recursionCounter.Increase()

	// Allow the dialect to override statement parsing
	if stmt, parsed, err := p.dialect.ParseStatement(p); err != nil {
		return nil, err
	} else if parsed {
		return stmt, nil
	}

	nextToken := p.NextToken()

	switch tok := nextToken.Token.(type) {
	case token.TokenWord:
		// Dispatch based on keyword
		return p.parseStatementByKeyword(string(tok.Word.Keyword), nextToken)
	case token.TokenLParen:
		// Parenthesized expression - likely a subquery
		p.PrevToken()
		return p.parseQuery()
	default:
		return nil, p.expected("an SQL statement", nextToken)
	}
}

// parseStatementByKeyword dispatches to specific statement parsers based on the keyword.
func (p *Parser) parseStatementByKeyword(keyword string, tok token.TokenWithSpan) (ast.Statement, error) {
	switch keyword {
	case "SELECT", "WITH", "VALUES":
		p.PrevToken()
		return p.parseQuery()
	case "INSERT":
		return p.parseInsert(tok)
	case "REPLACE":
		return p.parseReplace(tok)
	case "UPDATE":
		return p.parseUpdate(tok)
	case "DELETE":
		return p.parseDelete(tok)
	case "CREATE":
		return p.parseCreate()
	case "DROP":
		return p.parseDrop()
	case "ALTER":
		return p.parseAlter()
	case "TRUNCATE":
		return p.parseTruncate()
	case "EXPLAIN":
		return p.parseExplain()
	case "DESCRIBE", "DESC":
		return p.parseDescribe(keyword)
	case "SHOW":
		return p.parseShow()
	case "SET":
		return p.parseSet()
	case "BEGIN", "START":
		return p.parseBegin()
	case "COMMIT":
		return p.parseCommit()
	case "ROLLBACK":
		return p.parseRollback()
	case "GRANT":
		return p.parseGrant()
	case "REVOKE":
		return p.parseRevoke()
	case "DENY":
		return p.parseDeny()
	case "USE":
		return p.parseUse()
	case "ANALYZE":
		return p.parseAnalyze()
	case "CALL":
		return p.parseCall()
	case "COPY":
		return p.parseCopy()
	case "MERGE":
		return p.parseMerge(tok)
	case "EXECUTE", "EXEC":
		return p.parseExecute()
	case "PREPARE":
		return p.parsePrepare()
	case "DEALLOCATE":
		return p.parseDeallocate()
	case "DECLARE":
		return p.parseDeclare()
	case "FETCH":
		return p.parseFetch()
	case "CLOSE":
		return p.parseClose()
	case "CACHE":
		return p.parseCache()
	case "UNCACHE":
		return p.parseUncache()
	case "MSCK":
		return p.parseMsck()
	case "FLUSH":
		return p.parseFlush()
	case "KILL":
		return p.parseKill()
	case "VACUUM":
		return p.parseVacuum()
	case "OPTIMIZE":
		if p.dialect.SupportsOptimizeTable() {
			return p.parseOptimize()
		}
	case "LOAD":
		return p.parseLoad()
	case "INSTALL":
		if p.dialect.SupportsInstall() {
			return p.parseInstall()
		}
	case "UNLOAD":
		return p.parseUnload()
	case "ATTACH":
		return p.parseAttach()
	case "DETACH":
		if p.dialect.SupportsDetach() {
			return p.parseDetach()
		}
	case "COMMENT":
		if p.dialect.SupportsCommentOn() {
			return p.parseComment()
		}
	case "LISTEN":
		if p.dialect.SupportsListenNotify() {
			return p.parseListen()
		}
	case "NOTIFY":
		if p.dialect.SupportsListenNotify() {
			return p.parseNotify()
		}
	case "UNLISTEN":
		if p.dialect.SupportsListenNotify() {
			return p.parseUnlisten()
		}
	case "PRAGMA":
		return p.parsePragma()
	case "ASSERT":
		return p.parseAssert()
	case "SAVEPOINT":
		return p.parseSavepoint()
	case "RELEASE":
		return p.parseRelease()
	case "LOCK":
		return p.parseLock()
	case "UNLOCK":
		return p.parseUnlock()
	case "RENAME":
		return p.parseRename()
	case "RESET":
		return p.parseReset()
	case "DISCARD":
		return p.parseDiscard()
	case "EXPORT":
		return p.parseExport()
	case "RETURN":
		return p.parseReturn()
	case "PRINT":
		return p.parsePrint()
	case "RAISERROR":
		return p.parseRaiserror()
	case "THROW":
		return p.parseThrow()
	case "WAITFOR":
		return p.parseWaitFor()
	case "OPEN":
		return p.parseOpen()
	case "END":
		return p.parseEnd()
	case "IF":
		p.PrevToken()
		return p.parseIfStatement()
	case "WHILE":
		p.PrevToken()
		return p.parseWhile()
	case "CASE":
		p.PrevToken()
		return p.parseCaseStatement()
	case "RAISE":
		p.PrevToken()
		return p.parseRaise()
	}

	return nil, p.expected("an SQL statement", tok)
}

// GetDialect returns the dialect used by this parser.
func (p *Parser) GetDialect() parseriface.CompleteDialect {
	return p.dialect
}

// GetState returns the current parser state (implements parseriface.Parser).
func (p *Parser) GetState() parseriface.ParserState {
	switch p.state {
	case StateConnectBy:
		return parseriface.StateConnectBy
	case StateColumnDefinition:
		return parseriface.StateColumnDefinition
	default:
		return parseriface.StateNormal
	}
}

// SetState sets the current parser state (implements parseriface.Parser).
func (p *Parser) SetState(state parseriface.ParserState) {
	switch state {
	case parseriface.StateConnectBy:
		p.state = StateConnectBy
	case parseriface.StateColumnDefinition:
		p.state = StateColumnDefinition
	default:
		p.state = StateNormal
	}
}

// GetOptions returns the parser options (implements parseriface.Parser).
func (p *Parser) GetOptions() parseriface.ParserOptions {
	return parseriface.ParserOptions{
		TrailingCommas:   p.options.TrailingCommas,
		Unescape:         p.options.Unescape,
		RequireSemicolon: p.options.RequireSemicolon,
	}
}

// WithState sets the parser state temporarily for the duration of a function call.
func (p *Parser) WithState(state parseriface.ParserState, f func() error) error {
	oldState := p.state
	p.state = state
	err := f()
	p.state = oldState
	return err
}

// InColumnDefinitionState returns true if the parser is in the ColumnDefinition state.
func (p *Parser) InColumnDefinitionState() bool {
	return p.state == parseriface.StateColumnDefinition
}

// expected creates a parser error for an unexpected token.
func (p *Parser) expected(expected string, found token.TokenWithSpan) error {
	return errors.NewParserError(
		fmt.Sprintf("Expected: %s, found: %s", expected, found.Token.String()),
		found.Span,
	)
}

// expectedRef creates a parser error for an unexpected token (by reference).
func (p *Parser) expectedRef(expected string, found *token.TokenWithSpan) error {
	return errors.NewParserError(
		fmt.Sprintf("Expected: %s, found: %s", expected, found.Token.String()),
		found.Span,
	)
}

// expectedAt creates a parser error for a token at a specific index.
func (p *Parser) expectedAt(expected string, index int) error {
	found := p.tokenAt(index)
	return errors.NewParserError(
		fmt.Sprintf("Expected: %s, found: %s", expected, found.Token.String()),
		found.Span,
	)
}

// tokenAt returns the token at the given location, or EOF if beyond the length.
func (p *Parser) tokenAt(index int) *token.TokenWithSpan {
	if index < 0 {
		return &eofToken
	}
	if index >= len(p.tokens) {
		return &eofToken
	}
	return &p.tokens[index]
}

// Statement parser methods - delegate to the statements package

// ParseQuery parses a SELECT or other query statement
func (p *Parser) ParseQuery() (ast.Statement, error) {
	return parseQuery(p)
}

func (p *Parser) parseQuery() (ast.Statement, error) {
	return parseQuery(p)
}

func (p *Parser) parseInsert(tok token.TokenWithSpan) (ast.Statement, error) {
	return parseInsert(p, tok)
}

func (p *Parser) parseReplace(tok token.TokenWithSpan) (ast.Statement, error) {
	return parseReplace(p, tok)
}

func (p *Parser) parseUpdate(tok token.TokenWithSpan) (ast.Statement, error) {
	return parseUpdate(p, tok)
}

func (p *Parser) parseDelete(tok token.TokenWithSpan) (ast.Statement, error) {
	return parseDelete(p, tok)
}

func (p *Parser) parseCreate() (ast.Statement, error) {
	return parseCreate(p)
}

func (p *Parser) parseDrop() (ast.Statement, error) {
	return parseDrop(p)
}

func (p *Parser) parseAlter() (ast.Statement, error) {
	return parseAlter(p)
}

func (p *Parser) parseTruncate() (ast.Statement, error) {
	return parseTruncate(p)
}

func (p *Parser) parseExplain() (ast.Statement, error) {
	return parseExplain(p)
}

func (p *Parser) parseDescribe(keyword string) (ast.Statement, error) {
	return parseDescribeWithAlias(p, keyword)
}

func (p *Parser) parseShow() (ast.Statement, error) {
	return parseShow(p)
}

func (p *Parser) parseSet() (ast.Statement, error) {
	return parseSet(p)
}

func (p *Parser) parseBegin() (ast.Statement, error) {
	return parseBegin(p)
}

func (p *Parser) parseCommit() (ast.Statement, error) {
	return parseCommit(p)
}

func (p *Parser) parseRollback() (ast.Statement, error) {
	return parseRollback(p)
}

func (p *Parser) parseGrant() (ast.Statement, error) {
	return parseGrant(p)
}

func (p *Parser) parseRevoke() (ast.Statement, error) {
	return parseRevoke(p)
}

func (p *Parser) parseDeny() (ast.Statement, error) {
	return parseDeny(p)
}

func (p *Parser) parseUse() (ast.Statement, error) {
	return parseUse(p)
}

func (p *Parser) parseAnalyze() (ast.Statement, error) {
	return parseAnalyze(p)
}

func (p *Parser) parseCall() (ast.Statement, error) {
	return parseCall(p)
}

func (p *Parser) parseCopy() (ast.Statement, error) {
	return parseCopy(p)
}

func (p *Parser) parseMerge(tok token.TokenWithSpan) (ast.Statement, error) {
	return parseMerge(p, tok)
}

func (p *Parser) parseExecute() (ast.Statement, error) {
	return parseExecute(p)
}

func (p *Parser) parsePrepare() (ast.Statement, error) {
	return parsePrepare(p)
}

func (p *Parser) parseDeallocate() (ast.Statement, error) {
	return parseDeallocate(p)
}

func (p *Parser) parseDeclare() (ast.Statement, error) {
	return parseDeclare(p)
}

func (p *Parser) parseFetch() (ast.Statement, error) {
	return parseFetch(p)
}

func (p *Parser) parseClose() (ast.Statement, error) {
	return parseClose(p)
}

func (p *Parser) parseCache() (ast.Statement, error) {
	return nil, p.expectedRef("CACHE not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseUncache() (ast.Statement, error) {
	return nil, p.expectedRef("UNCACHE not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseMsck() (ast.Statement, error) {
	return nil, p.expectedRef("MSCK not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseFlush() (ast.Statement, error) {
	return nil, p.expectedRef("FLUSH not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseKill() (ast.Statement, error) {
	// Parse optional modifier: CONNECTION, QUERY, MUTATION, or HARD
	var modifier *expr.KillType
	var hasHard bool

	kw := p.ParseOneOfKeywords([]string{"CONNECTION", "QUERY", "MUTATION", "HARD"})
	switch kw {
	case "CONNECTION":
		m := expr.KillTypeConnection
		modifier = &m
	case "QUERY":
		m := expr.KillTypeQuery
		modifier = &m
	case "MUTATION":
		// MUTATION is only supported for ClickHouse and Generic dialects
		if p.GetDialect().Dialect() == "clickhouse" || p.GetDialect().Dialect() == "generic" {
			m := expr.KillTypeMutation
			modifier = &m
		} else {
			return nil, fmt.Errorf("unsupported KILL type MUTATION for dialect %s", p.GetDialect().Dialect())
		}
	case "HARD":
		hasHard = true
		// Check for additional QUERY or CONNECTION modifier after HARD
		if p.ParseKeyword("QUERY") {
			m := expr.KillTypeQuery
			modifier = &m
		} else if p.ParseKeyword("CONNECTION") {
			m := expr.KillTypeConnection
			modifier = &m
		}
	}

	// Parse the process ID (uint)
	tok := p.NextToken()
	if num, ok := tok.Token.(token.TokenNumber); ok {
		id, err := strconv.ParseUint(num.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid process ID: %w", err)
		}
		return &statement.Kill{
			Modifier: modifier,
			Hard:     hasHard,
			ID:       id,
		}, nil
	}

	return nil, fmt.Errorf("expected process ID after KILL")
}

func (p *Parser) parseVacuum() (ast.Statement, error) {
	return nil, p.expectedRef("VACUUM not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseOptimize() (ast.Statement, error) {
	return nil, p.expectedRef("OPTIMIZE not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseLoad() (ast.Statement, error) {
	// Check for DuckDB LOAD extension syntax: LOAD extension_name
	if p.dialect.SupportsLoadExtension() {
		extensionName, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		return &statement.Load{
			ExtensionName: extensionName,
		}, nil
	}

	// Check for Hive LOAD DATA syntax: LOAD DATA [LOCAL] INPATH 'path' [OVERWRITE] INTO TABLE table_name
	if p.ParseKeyword("DATA") && p.dialect.SupportsLoadData() {
		// Parse optional LOCAL
		local := p.ParseKeyword("LOCAL")

		// Parse INPATH
		if _, err := p.ExpectKeyword("INPATH"); err != nil {
			return nil, err
		}

		// Parse path string
		pathTok, err := p.ExpectToken(token.TokenSingleQuotedString{})
		if err != nil {
			return nil, err
		}
		path := ""
		if str, ok := pathTok.Token.(token.TokenSingleQuotedString); ok {
			path = str.Value
		}

		// Parse optional OVERWRITE
		overwrite := p.ParseKeyword("OVERWRITE")

		// Parse INTO TABLE
		if _, err := p.ExpectKeyword("INTO"); err != nil {
			return nil, err
		}
		if _, err := p.ExpectKeyword("TABLE"); err != nil {
			return nil, err
		}

		// Parse table name
		tableName, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}

		// For simplicity, skip partition and table_format parsing for now
		// TODO: Implement partition and table_format parsing for full Hive support

		return &statement.LoadData{
			Local:     local,
			Inpath:    path,
			Overwrite: overwrite,
			TableName: tableName,
		}, nil
	}

	return nil, p.expectedRef("extension name or DATA", p.PeekTokenRef())
}

func (p *Parser) parseInstall() (ast.Statement, error) {
	return nil, p.expectedRef("INSTALL not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseUnload() (ast.Statement, error) {
	return nil, p.expectedRef("UNLOAD not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseAttach() (ast.Statement, error) {
	return nil, p.expectedRef("ATTACH not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseDetach() (ast.Statement, error) {
	return nil, p.expectedRef("DETACH not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseComment() (ast.Statement, error) {
	// Parse optional IF EXISTS
	ifExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse ON keyword
	if _, err := p.ExpectKeyword("ON"); err != nil {
		return nil, err
	}

	// Parse object type
	tok := p.PeekToken()
	var objectType expr.CommentObject
	var objectName *ast.ObjectName

	if word, ok := tok.Token.(token.TokenWord); ok {
		switch strings.ToUpper(string(word.Keyword)) {
		case "COLUMN":
			objectType = expr.CommentColumn
			p.AdvanceToken()
			name, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			objectName = name
		case "DATABASE":
			objectType = expr.CommentDatabase
			p.AdvanceToken()
			name, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			objectName = name
		case "DOMAIN":
			objectType = expr.CommentDomain
			p.AdvanceToken()
			name, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			objectName = name
		case "EXTENSION":
			objectType = expr.CommentExtension
			p.AdvanceToken()
			name, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			objectName = name
		case "FUNCTION":
			objectType = expr.CommentFunction
			p.AdvanceToken()
			name, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			objectName = name
		case "INDEX":
			objectType = expr.CommentIndex
			p.AdvanceToken()
			name, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			objectName = name
		case "MATERIALIZED":
			p.AdvanceToken()
			if _, err := p.ExpectKeyword("VIEW"); err != nil {
				return nil, err
			}
			objectType = expr.CommentMaterializedView
			name, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			objectName = name
		case "PROCEDURE":
			objectType = expr.CommentProcedure
			p.AdvanceToken()
			name, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			objectName = name
		case "ROLE":
			objectType = expr.CommentRole
			p.AdvanceToken()
			name, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			objectName = name
		case "SCHEMA":
			objectType = expr.CommentSchema
			p.AdvanceToken()
			name, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			objectName = name
		case "SEQUENCE":
			objectType = expr.CommentSequence
			p.AdvanceToken()
			name, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			objectName = name
		case "TABLE":
			objectType = expr.CommentTable
			p.AdvanceToken()
			name, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			objectName = name
		case "TYPE":
			objectType = expr.CommentType
			p.AdvanceToken()
			name, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			objectName = name
		case "USER":
			objectType = expr.CommentUser
			p.AdvanceToken()
			name, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			objectName = name
		case "VIEW":
			objectType = expr.CommentView
			p.AdvanceToken()
			name, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			objectName = name
		default:
			return nil, p.expectedRef("comment object type (COLUMN, TABLE, VIEW, etc.)", p.PeekTokenRef())
		}
	} else {
		return nil, p.expectedRef("comment object type", p.PeekTokenRef())
	}

	// Parse IS keyword
	if _, err := p.ExpectKeyword("IS"); err != nil {
		return nil, err
	}

	// Parse comment value (string literal or NULL)
	var comment *string
	if p.ParseKeyword("NULL") {
		comment = nil
	} else {
		tok, err := p.ExpectToken(token.TokenSingleQuotedString{})
		if err != nil {
			return nil, err
		}
		if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			comment = &str.Value
		}
	}

	return &statement.Comment{
		ObjectType: objectType,
		ObjectName: objectName,
		Comment:    comment,
		IfExists:   ifExists,
	}, nil
}

func (p *Parser) parseListen() (ast.Statement, error) {
	channel, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	return &statement.Listen{
		Channel: channel,
	}, nil
}

func (p *Parser) parseNotify() (ast.Statement, error) {
	channel, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	var payload *string
	if p.ConsumeToken(token.TokenComma{}) {
		tok, err := p.ExpectToken(token.TokenSingleQuotedString{})
		if err != nil {
			return nil, err
		}
		if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			payload = &str.Value
		}
	}

	return &statement.Notify{
		Channel: channel,
		Payload: payload,
	}, nil
}

func (p *Parser) parseUnlisten() (ast.Statement, error) {
	// Check for wildcard (*)
	tok := p.PeekToken()
	if _, ok := tok.Token.(token.TokenMul); ok {
		p.AdvanceToken()
		// Create identifier with * as the name
		return &statement.Unlisten{
			Channel: &ast.Ident{Value: "*"},
		}, nil
	}

	channel, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	return &statement.Unlisten{
		Channel: channel,
	}, nil
}

func (p *Parser) parsePragma() (ast.Statement, error) {
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	if p.ConsumeToken(token.TokenLParen{}) {
		// PRAGMA name(value)
		value, err := p.parsePragmaValue()
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		return &statement.Pragma{
			Name:  name,
			Value: value,
			IsEq:  false,
		}, nil
	} else if p.ConsumeToken(token.TokenEq{}) {
		// PRAGMA name = value
		value, err := p.parsePragmaValue()
		if err != nil {
			return nil, err
		}
		return &statement.Pragma{
			Name:  name,
			Value: value,
			IsEq:  true,
		}, nil
	}

	// PRAGMA name (no value)
	return &statement.Pragma{
		Name:  name,
		Value: nil,
		IsEq:  false,
	}, nil
}

// parsePragmaValue parses a value for PRAGMA (number, string, or placeholder)
func (p *Parser) parsePragmaValue() (expr.Expr, error) {
	tok := p.NextToken()
	switch t := tok.Token.(type) {
	case token.TokenSingleQuotedString:
		return &expr.ValueExpr{
			Value: ast.NewSingleQuotedString(t.Value),
		}, nil
	case token.TokenNumber:
		val, err := ast.NewNumber(t.Value, false)
		if err != nil {
			return nil, err
		}
		return &expr.ValueExpr{
			Value: val,
		}, nil
	case token.TokenPlaceholder:
		return &expr.ValueExpr{
			Value: ast.NewPlaceholder(t.Value),
		}, nil
	default:
		return nil, fmt.Errorf("Expected number, string, or placeholder for PRAGMA value, got %v", t)
	}
}

func (p *Parser) parseAssert() (ast.Statement, error) {
	// Parse condition expression
	condition, err := NewExpressionParser(p).ParseExpr()
	if err != nil {
		return nil, err
	}

	// Parse optional AS message
	var message expr.Expr
	if p.ParseKeyword("AS") {
		message, err = NewExpressionParser(p).ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	return &statement.Assert{
		Condition: condition,
		Message:   message,
	}, nil
}

func (p *Parser) parseSavepoint() (ast.Statement, error) {
	return parseSavepoint(p)
}

func (p *Parser) parseRelease() (ast.Statement, error) {
	return parseRelease(p)
}

func (p *Parser) parseLock() (ast.Statement, error) {
	return nil, p.expectedRef("LOCK not yet implemented for generic dialect", p.PeekTokenRef())
}

func (p *Parser) parseUnlock() (ast.Statement, error) {
	return nil, p.expectedRef("UNLOCK not yet implemented for generic dialect", p.PeekTokenRef())
}

func (p *Parser) parseRename() (ast.Statement, error) {
	return nil, p.expectedRef("RENAME not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseReset() (ast.Statement, error) {
	return nil, p.expectedRef("RESET not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseDiscard() (ast.Statement, error) {
	// Parse discard object type: ALL, PLANS, SEQUENCES, or TEMP
	var objectType expr.DiscardObject

	if word, ok := p.PeekTokenRef().Token.(token.TokenWord); ok {
		switch strings.ToUpper(word.Value) {
		case "ALL":
			objectType = expr.DiscardAll
			p.AdvanceToken()
		case "PLANS":
			objectType = expr.DiscardPlans
			p.AdvanceToken()
		case "SEQUENCES":
			objectType = expr.DiscardSequences
			p.AdvanceToken()
		case "TEMP", "TEMPORARY":
			objectType = expr.DiscardTemp
			p.AdvanceToken()
		default:
			return nil, p.expectedRef("ALL, PLANS, SEQUENCES, or TEMP", p.PeekTokenRef())
		}
	} else {
		return nil, p.expectedRef("ALL, PLANS, SEQUENCES, or TEMP", p.PeekTokenRef())
	}

	return &statement.Discard{
		ObjectType: objectType,
	}, nil
}

func (p *Parser) parseExport() (ast.Statement, error) {
	return nil, p.expectedRef("EXPORT not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseReturn() (ast.Statement, error) {
	return nil, p.expectedRef("RETURN not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parsePrint() (ast.Statement, error) {
	return nil, p.expectedRef("PRINT not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseRaiserror() (ast.Statement, error) {
	return nil, p.expectedRef("RAISERROR not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseThrow() (ast.Statement, error) {
	return nil, p.expectedRef("THROW not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseWaitFor() (ast.Statement, error) {
	return nil, p.expectedRef("WAITFOR not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseOpen() (ast.Statement, error) {
	return nil, p.expectedRef("OPEN not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseEnd() (ast.Statement, error) {
	return nil, p.expectedRef("END not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseIfStatement() (ast.Statement, error) {
	return nil, p.expectedRef("IF statement not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseWhile() (ast.Statement, error) {
	return nil, p.expectedRef("WHILE not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseCaseStatement() (ast.Statement, error) {
	return nil, p.expectedRef("CASE statement not yet implemented", p.PeekTokenRef())
}

func (p *Parser) parseRaise() (ast.Statement, error) {
	return nil, p.expectedRef("RAISE not yet implemented", p.PeekTokenRef())
}

// ParseExpression implements the dialects.ParserAccessor interface.
// TODO: Implement actual expression parsing.
func (p *Parser) ParseExpression() (ast.Expr, error) {
	return nil, fmt.Errorf("expression parsing not yet implemented")
}

// ParseInsert implements the dialects.ParserAccessor interface.
// TODO: Implement actual INSERT parsing.
func (p *Parser) ParseInsert() (ast.Statement, error) {
	return nil, fmt.Errorf("INSERT parsing not yet implemented")
}

// ExpectedRef returns an error with context about what was expected and what was found.
// This is the exported version of expectedRef.
func (p *Parser) ExpectedRef(expected string, found *token.TokenWithSpan) error {
	return p.expectedRef(expected, found)
}

// Expected returns an error with context about what was expected and what was found.
// This is the exported version of expected.
func (p *Parser) Expected(expected string, found token.TokenWithSpan) error {
	return p.expected(expected, found)
}

// ParseObjectName parses an object name (table name, column name, etc.)
// Handles multi-part names like "schema.table" or "db.schema.table"
func (p *Parser) ParseObjectName() (*ast.ObjectName, error) {
	var parts []ast.ObjectNamePart

	for {
		// Parse the next identifier
		ident, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		parts = append(parts, &ast.ObjectNamePartIdentifier{Ident: ident})

		// Check if there's a period - if so, continue to next part
		if p.ConsumeToken(token.TokenPeriod{}) {
			// Continue to parse the next part
			continue
		}

		// No more parts
		break
	}

	return &ast.ObjectName{Parts: parts}, nil
}

// ParseParenthesizedColumnList parses a parenthesized list of column names: (col1, col2, ...)
// Returns the list of identifiers and consumes the closing parenthesis.
func (p *Parser) ParseParenthesizedColumnList() ([]*ast.Ident, error) {
	// Expect opening parenthesis
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Check for empty list
	if _, ok := p.PeekToken().Token.(token.TokenRParen); ok {
		p.AdvanceToken()
		return nil, nil
	}

	var columns []*ast.Ident
	for {
		ident, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		columns = append(columns, ident)

		// Check for comma or closing parenthesis
		if p.ConsumeToken(token.TokenComma{}) {
			continue
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		break
	}

	return columns, nil
}

// ParseIdentifier parses a single identifier
// TODO: Implement proper identifier parsing
func (p *Parser) ParseIdentifier() (*ast.Ident, error) {
	tok := p.PeekToken()
	if word, ok := tok.Token.(token.TokenWord); ok {
		p.AdvanceToken()
		// Preserve original case for all dialects
		// This matches the Rust reference implementation
		ident := &ast.Ident{Value: word.Word.Value}
		// If the word has a quote style, set it on the identifier
		if word.Word.QuoteStyle != nil {
			quoteStyle := rune(*word.Word.QuoteStyle)
			ident.QuoteStyle = &quoteStyle
		}
		return ident, nil
	}
	return nil, fmt.Errorf("expected identifier, found %v", tok.Token)
}

// PeekNthKeyword implements dialects.ParserAccessor interface.
// Checks if the nth token is the expected keyword.
func (p *Parser) PeekNthKeyword(n int, expected string) bool {
	tok := p.PeekNthToken(n)
	if word, ok := tok.Token.(token.TokenWord); ok {
		return strings.EqualFold(word.Word.Value, expected)
	}
	return false
}

// TryDecreaseRecursion attempts to decrease the recursion counter.
// Returns an error if the recursion limit is exceeded.
func (p *Parser) TryDecreaseRecursion() error {
	return p.recursionCounter.TryDecrease()
}

// IncreaseRecursion increases the recursion counter (call after TryDecreaseRecursion).
func (p *Parser) IncreaseRecursion() {
	p.recursionCounter.Increase()
}

// ExpectedAt returns an error with context about what was expected at a specific index.
// This is the exported version of expectedAt.
func (p *Parser) ExpectedAt(expected string, index int) error {
	return p.expectedAt(expected, index)
}

// ParseDataType parses a SQL data type
// This is a simplified implementation that handles common data types
// like INT, TEXT, VARCHAR, etc.
func (p *Parser) ParseDataType() (datatype.DataType, error) {
	tok := p.PeekToken()
	word, ok := tok.Token.(token.TokenWord)
	if !ok {
		return nil, fmt.Errorf("expected data type keyword, found %v", tok.Token)
	}

	p.AdvanceToken()
	typeName := strings.ToUpper(word.Word.Value)

	switch typeName {
	case "INT", "INTEGER":
		return parseIntType(p, tok.Span)
	case "INT4":
		return parseInt4Type(p, tok.Span)
	case "INT8":
		return parseInt8Type(p, tok.Span)
	case "BIGINT":
		return parseBigIntType(p, tok.Span)
	case "SMALLINT":
		return parseSmallIntType(p, tok.Span)
	case "TINYINT":
		return parseTinyIntType(p, tok.Span)
	case "MEDIUMINT":
		return parseMediumIntType(p, tok.Span)
	case "TEXT":
		return &datatype.TextType{SpanVal: tok.Span}, nil
	case "VARCHAR":
		return parseVarcharType(p, tok.Span)
	case "NVARCHAR":
		return parseNvarcharType(p, tok.Span)
	case "CHAR", "CHARACTER":
		return parseCharType(p, tok.Span)
	case "NCHAR":
		return parseNcharType(p, tok.Span)
	case "VARCHAR2":
		return parseVarchar2Type(p, tok.Span)
	case "NVARCHAR2":
		return parseNvarchar2Type(p, tok.Span)
	case "BOOL", "BOOLEAN":
		return &datatype.BooleanType{SpanVal: tok.Span}, nil
	case "DATE":
		return &datatype.DateType{SpanVal: tok.Span}, nil
	case "TIME":
		return parseTimeType(p, tok.Span)
	case "TIMESTAMP":
		return parseTimestampType(p, tok.Span)
	case "DATETIME":
		return parseDatetimeType(p, tok.Span)
	case "FLOAT":
		return parseFloatType(p, tok.Span)
	case "DOUBLE":
		return parseDoubleType(p, tok.Span)
	case "NUMERIC":
		return parseNumericType(p, tok.Span)
	case "DECIMAL", "DEC":
		return parseDecimalType(p, tok.Span)
	case "REAL":
		return parseRealType(p, tok.Span)
	case "BYTEA":
		return &datatype.ByteaType{SpanVal: tok.Span}, nil
	case "JSON":
		return &datatype.JSONType{SpanVal: tok.Span}, nil
	case "JSONB":
		return &datatype.JSONBType{SpanVal: tok.Span}, nil
	case "UUID":
		return &datatype.UuidType{SpanVal: tok.Span}, nil
	case "CLOB":
		return parseClobType(p, tok.Span)
	case "BLOB":
		return parseBlobType(p, tok.Span)
	case "BINARY":
		return parseBinaryType(p, tok.Span)
	case "VARBINARY":
		return parseVarbinaryType(p, tok.Span)
	default:
		// For unknown types, return a custom type
		return &datatype.CustomType{
			SpanVal: tok.Span,
			Name: &expr.ObjectName{
				Parts: []*expr.ObjectNamePart{{Ident: &expr.Ident{Value: word.Word.Value}}},
			},
		}, nil
	}
}

// parseVarcharType parses VARCHAR [(n)]
func parseVarcharType(p *Parser, spanVal token.Span) (*datatype.VarcharType, error) {
	result := &datatype.VarcharType{
		SpanVal: spanVal,
	}

	// Check for optional size specification
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			length, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid VARCHAR length: %w", err)
			}
			result.Length = &datatype.CharacterLength{Length: length}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in VARCHAR length specification")
		}
	}

	return result, nil
}

// parseCharType parses CHAR [(n)]
func parseCharType(p *Parser, spanVal token.Span) (*datatype.CharType, error) {
	result := &datatype.CharType{
		SpanVal: spanVal,
	}

	// Check for optional size specification
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			length, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid CHAR length: %w", err)
			}
			result.Length = &datatype.CharacterLength{Length: length}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in CHAR length specification")
		}
	}

	return result, nil
}

// parseNvarcharType parses NVARCHAR [(n)]
func parseNvarcharType(p *Parser, spanVal token.Span) (*datatype.NvarcharType, error) {
	result := &datatype.NvarcharType{
		SpanVal: spanVal,
	}

	// Check for optional size specification
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			length, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid NVARCHAR length: %w", err)
			}
			result.Length = &datatype.CharacterLength{Length: length}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in NVARCHAR length specification")
		}
	}

	return result, nil
}

// parseNcharType parses NCHAR [(n)]
func parseNcharType(p *Parser, spanVal token.Span) (*datatype.NcharType, error) {
	result := &datatype.NcharType{
		SpanVal: spanVal,
	}

	// Check for optional size specification
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			length, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid NCHAR length: %w", err)
			}
			result.Length = &datatype.CharacterLength{Length: length}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in NCHAR length specification")
		}
	}

	return result, nil
}

// parseVarchar2Type parses VARCHAR2 [(n)] (Oracle-specific)
func parseVarchar2Type(p *Parser, spanVal token.Span) (*datatype.Varchar2Type, error) {
	result := &datatype.Varchar2Type{
		SpanVal: spanVal,
	}

	// Check for optional size specification
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			length, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid VARCHAR2 length: %w", err)
			}
			result.Length = &datatype.CharacterLength{Length: length}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in VARCHAR2 length specification")
		}
	}

	return result, nil
}

// parseNvarchar2Type parses NVARCHAR2 [(n)] (Oracle-specific)
func parseNvarchar2Type(p *Parser, spanVal token.Span) (*datatype.Nvarchar2Type, error) {
	result := &datatype.Nvarchar2Type{
		SpanVal: spanVal,
	}

	// Check for optional size specification
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			length, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid NVARCHAR2 length: %w", err)
			}
			result.Length = &datatype.CharacterLength{Length: length}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in NVARCHAR2 length specification")
		}
	}

	return result, nil
}

// parseTimeType parses TIME [WITH TIME ZONE]
func parseTimeType(p *Parser, spanVal token.Span) (*datatype.TimeType, error) {
	result := &datatype.TimeType{
		SpanVal: spanVal,
	}

	// Check for optional precision
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			precision, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid TIME precision: %w", err)
			}
			result.Precision = &precision
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		}
	}

	// Check for WITH TIME ZONE
	if p.ParseKeyword("WITH") {
		if p.ParseKeyword("TIME") && p.ParseKeyword("ZONE") {
			result.TimezoneInfo = datatype.WithTimeZone
		}
	}

	return result, nil
}

// parseOptionalPrecision parses optional (n) display width for integer types
func parseOptionalPrecision(p *Parser) *uint64 {
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); !isLParen {
		return nil
	}
	p.NextToken() // consume (
	tok := p.NextToken()
	if num, ok := tok.Token.(token.TokenNumber); ok {
		precision, err := strconv.ParseUint(num.Value, 10, 64)
		if err != nil {
			return nil
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil
		}
		return &precision
	}
	return nil
}

// parseIntType parses INT/INTEGER with optional display width and UNSIGNED modifier
// Reference: src/parser/mod.rs:11996-12006
func parseIntType(p *Parser, spanVal token.Span) (datatype.DataType, error) {
	displayWidth := parseOptionalPrecision(p)
	if p.ParseKeyword("UNSIGNED") {
		return &datatype.IntUnsignedType{DisplayWidth: displayWidth, SpanVal: spanVal}, nil
	}
	// MySQL allows optional SIGNED keyword
	if p.GetDialect().SupportsIndexHints() {
		p.ParseKeyword("SIGNED")
	}
	return &datatype.IntType{DisplayWidth: displayWidth, SpanVal: spanVal}, nil
}

// parseInt4Type parses INT4 with optional display width and UNSIGNED modifier
func parseInt4Type(p *Parser, spanVal token.Span) (datatype.DataType, error) {
	displayWidth := parseOptionalPrecision(p)
	if p.ParseKeyword("UNSIGNED") {
		return &datatype.Int4UnsignedType{DisplayWidth: displayWidth, SpanVal: spanVal}, nil
	}
	return &datatype.Int4Type{DisplayWidth: displayWidth, SpanVal: spanVal}, nil
}

// parseInt8Type parses INT8 with optional display width and UNSIGNED modifier
func parseInt8Type(p *Parser, spanVal token.Span) (datatype.DataType, error) {
	displayWidth := parseOptionalPrecision(p)
	if p.ParseKeyword("UNSIGNED") {
		return &datatype.Int8UnsignedType{DisplayWidth: displayWidth, SpanVal: spanVal}, nil
	}
	return &datatype.Int8Type{DisplayWidth: displayWidth, SpanVal: spanVal}, nil
}

// parseBigIntType parses BIGINT with optional display width and UNSIGNED modifier
// Reference: src/parser/mod.rs:12039-12049
func parseBigIntType(p *Parser, spanVal token.Span) (datatype.DataType, error) {
	displayWidth := parseOptionalPrecision(p)
	if p.ParseKeyword("UNSIGNED") {
		return &datatype.BigIntUnsignedType{DisplayWidth: displayWidth, SpanVal: spanVal}, nil
	}
	// MySQL allows optional SIGNED keyword
	if p.GetDialect().SupportsIndexHints() {
		p.ParseKeyword("SIGNED")
	}
	return &datatype.BigIntType{DisplayWidth: displayWidth, SpanVal: spanVal}, nil
}

// parseSmallIntType parses SMALLINT with optional display width and UNSIGNED modifier
// Reference: src/parser/mod.rs:11974-11984
func parseSmallIntType(p *Parser, spanVal token.Span) (datatype.DataType, error) {
	displayWidth := parseOptionalPrecision(p)
	if p.ParseKeyword("UNSIGNED") {
		return &datatype.SmallIntUnsignedType{DisplayWidth: displayWidth, SpanVal: spanVal}, nil
	}
	// MySQL allows optional SIGNED keyword
	if p.GetDialect().SupportsIndexHints() {
		p.ParseKeyword("SIGNED")
	}
	return &datatype.SmallIntType{DisplayWidth: displayWidth, SpanVal: spanVal}, nil
}

// parseTinyIntType parses TINYINT with optional display width and UNSIGNED modifier
// Reference: src/parser/mod.rs:11955-11965
func parseTinyIntType(p *Parser, spanVal token.Span) (datatype.DataType, error) {
	displayWidth := parseOptionalPrecision(p)
	if p.ParseKeyword("UNSIGNED") {
		return &datatype.TinyIntUnsignedType{DisplayWidth: displayWidth, SpanVal: spanVal}, nil
	}
	// MySQL allows optional SIGNED keyword
	if p.GetDialect().SupportsIndexHints() {
		p.ParseKeyword("SIGNED")
	}
	return &datatype.TinyIntType{DisplayWidth: displayWidth, SpanVal: spanVal}, nil
}

// parseMediumIntType parses MEDIUMINT with optional display width and UNSIGNED modifier
// Reference: src/parser/mod.rs:11985-11995
func parseMediumIntType(p *Parser, spanVal token.Span) (datatype.DataType, error) {
	displayWidth := parseOptionalPrecision(p)
	if p.ParseKeyword("UNSIGNED") {
		return &datatype.MediumIntUnsignedType{DisplayWidth: displayWidth, SpanVal: spanVal}, nil
	}
	// MySQL allows optional SIGNED keyword
	if p.GetDialect().SupportsIndexHints() {
		p.ParseKeyword("SIGNED")
	}
	return &datatype.MediumIntType{DisplayWidth: displayWidth, SpanVal: spanVal}, nil
}

// parseFloatType parses FLOAT with optional precision/scale and UNSIGNED modifier
// Reference: src/parser/mod.rs:11918-11926
func parseFloatType(p *Parser, spanVal token.Span) (datatype.DataType, error) {
	// Parse optional (p) or (p,s)
	var info datatype.ExactNumberInfo
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			prec, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid FLOAT precision: %w", err)
			}
			info.Precision = &prec
			// Check for optional scale
			if p.ConsumeToken(token.TokenComma{}) {
				tok = p.NextToken()
				if num, ok := tok.Token.(token.TokenNumber); ok {
					s, err := strconv.ParseInt(num.Value, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("invalid FLOAT scale: %w", err)
					}
					info.Scale = &s
				} else {
					return nil, fmt.Errorf("expected number for FLOAT scale")
				}
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in FLOAT precision specification")
		}
	}

	if p.ParseKeyword("UNSIGNED") {
		return &datatype.FloatUnsignedType{Info: info, SpanVal: spanVal}, nil
	}
	return &datatype.FloatType{Info: info, SpanVal: spanVal}, nil
}

// parseDoubleType parses DOUBLE with optional precision/scale and UNSIGNED modifier
// Reference: src/parser/mod.rs:11938-11954
func parseDoubleType(p *Parser, spanVal token.Span) (datatype.DataType, error) {
	// Check for PRECISION keyword (DOUBLE PRECISION)
	if p.ParseKeyword("PRECISION") {
		if p.ParseKeyword("UNSIGNED") {
			return &datatype.DoublePrecisionUnsignedType{SpanVal: spanVal}, nil
		}
		return &datatype.DoubleType{SpanVal: spanVal}, nil
	}

	// Parse optional (p) or (p,s)
	var info datatype.ExactNumberInfo
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			prec, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid DOUBLE precision: %w", err)
			}
			info.Precision = &prec
			// Check for optional scale
			if p.ConsumeToken(token.TokenComma{}) {
				tok = p.NextToken()
				if num, ok := tok.Token.(token.TokenNumber); ok {
					s, err := strconv.ParseInt(num.Value, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("invalid DOUBLE scale: %w", err)
					}
					info.Scale = &s
				} else {
					return nil, fmt.Errorf("expected number for DOUBLE scale")
				}
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in DOUBLE precision specification")
		}
	}

	if p.ParseKeyword("UNSIGNED") {
		return &datatype.DoubleUnsignedType{Info: info, SpanVal: spanVal}, nil
	}
	return &datatype.DoubleType{Info: info, SpanVal: spanVal}, nil
}

// parseNumericType parses NUMERIC [(p[,s])] with UNSIGNED modifier
// Reference: src/parser/mod.rs (similar to DECIMAL)
func parseNumericType(p *Parser, spanVal token.Span) (datatype.DataType, error) {
	var info datatype.ExactNumberInfo
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			prec, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid NUMERIC precision: %w", err)
			}
			info.Precision = &prec
			// Check for optional scale
			if p.ConsumeToken(token.TokenComma{}) {
				tok = p.NextToken()
				if num, ok := tok.Token.(token.TokenNumber); ok {
					s, err := strconv.ParseInt(num.Value, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("invalid NUMERIC scale: %w", err)
					}
					info.Scale = &s
				} else {
					return nil, fmt.Errorf("expected number for NUMERIC scale")
				}
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in NUMERIC precision specification")
		}
	}

	if p.ParseKeyword("UNSIGNED") {
		return &datatype.NumericType{Info: info, SpanVal: spanVal}, nil
	}
	return &datatype.NumericType{Info: info, SpanVal: spanVal}, nil
}

// parseDecimalType parses DECIMAL [(p[,s])] or DEC [(p[,s])] with UNSIGNED modifier
// Reference: src/parser/mod.rs:12181-12198
func parseDecimalType(p *Parser, spanVal token.Span) (datatype.DataType, error) {
	var info datatype.ExactNumberInfo
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			prec, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid DECIMAL precision: %w", err)
			}
			info.Precision = &prec
			// Check for optional scale
			if p.ConsumeToken(token.TokenComma{}) {
				tok = p.NextToken()
				if num, ok := tok.Token.(token.TokenNumber); ok {
					s, err := strconv.ParseInt(num.Value, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("invalid DECIMAL scale: %w", err)
					}
					info.Scale = &s
				} else {
					return nil, fmt.Errorf("expected number for DECIMAL scale")
				}
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in DECIMAL precision specification")
		}
	}

	if p.ParseKeyword("UNSIGNED") {
		return &datatype.DecimalUnsignedType{Info: info, SpanVal: spanVal}, nil
	}
	return &datatype.DecimalType{Info: info, SpanVal: spanVal}, nil
}

// parseRealType parses REAL with UNSIGNED modifier
// Reference: src/parser/mod.rs:11927-11933
func parseRealType(p *Parser, spanVal token.Span) (datatype.DataType, error) {
	if p.ParseKeyword("UNSIGNED") {
		return &datatype.RealUnsignedType{SpanVal: spanVal}, nil
	}
	return &datatype.RealType{SpanVal: spanVal}, nil
}

// parseTimestampType parses TIMESTAMP [(precision)]
func parseTimestampType(p *Parser, spanVal token.Span) (*datatype.TimestampType, error) {
	result := &datatype.TimestampType{
		SpanVal: spanVal,
	}

	// Check for optional precision specification
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			precision, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid TIMESTAMP precision: %w", err)
			}
			result.Precision = &precision
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in TIMESTAMP precision specification")
		}
	}

	return result, nil
}

// parseDatetimeType parses DATETIME [(precision)]
func parseDatetimeType(p *Parser, spanVal token.Span) (*datatype.DatetimeType, error) {
	result := &datatype.DatetimeType{
		SpanVal: spanVal,
	}

	// Check for optional precision specification
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			precision, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid DATETIME precision: %w", err)
			}
			result.Precision = &precision
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in DATETIME precision specification")
		}
	}

	return result, nil
}

// parseClobType parses CLOB [(n)]
func parseClobType(p *Parser, spanVal token.Span) (*datatype.ClobType, error) {
	result := &datatype.ClobType{
		SpanVal: spanVal,
	}

	// Check for optional length specification
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			length, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid CLOB length: %w", err)
			}
			result.Length = &length
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in CLOB length specification")
		}
	}

	return result, nil
}

// parseBlobType parses BLOB [(n)]
func parseBlobType(p *Parser, spanVal token.Span) (*datatype.BlobType, error) {
	result := &datatype.BlobType{
		SpanVal: spanVal,
	}

	// Check for optional length specification
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			length, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid BLOB length: %w", err)
			}
			result.Length = &length
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in BLOB length specification")
		}
	}

	return result, nil
}

// parseBinaryType parses BINARY [(n)]
func parseBinaryType(p *Parser, spanVal token.Span) (*datatype.BinaryType, error) {
	result := &datatype.BinaryType{
		SpanVal: spanVal,
	}

	// Check for optional length specification
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			length, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid BINARY length: %w", err)
			}
			result.Length = &length
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in BINARY length specification")
		}
	}

	return result, nil
}

// parseVarbinaryType parses VARBINARY [(n)]
func parseVarbinaryType(p *Parser, spanVal token.Span) (*datatype.VarbinaryType, error) {
	result := &datatype.VarbinaryType{
		SpanVal: spanVal,
	}

	// Check for optional length specification
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.NextToken() // consume (
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			length, err := strconv.ParseUint(num.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid VARBINARY length: %w", err)
			}
			result.Length = &datatype.BinaryLength{Length: length}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("expected number in VARBINARY length specification")
		}
	}

	return result, nil
}
