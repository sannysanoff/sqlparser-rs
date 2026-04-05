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

// parseCopy parses COPY statements
// Reference: src/parser/mod.rs parse_copy
func parseCopy(p *Parser) (ast.Statement, error) {
	copyStmt := &statement.Copy{}

	// Parse source: either (query) or table_name [ (columns) ]
	if p.ConsumeToken(token.TokenLParen{}) {
		// Parse query as source
		query, err := p.ParseQuery()
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		copyStmt.Source = &expr.CopySource{
			Query: query,
		}
	} else {
		// Parse table name
		tableName, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}

		// Parse optional column list
		var columns []*ast.Ident
		if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
			p.AdvanceToken() // consume (
			for {
				if _, ok := p.PeekToken().Token.(token.TokenRParen); ok {
					p.AdvanceToken() // consume )
					break
				}
				col, err := p.ParseIdentifier()
				if err != nil {
					return nil, err
				}
				columns = append(columns, col)
				if !p.ConsumeToken(token.TokenComma{}) {
					if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
						return nil, err
					}
					break
				}
			}
		}

		copyStmt.Source = &expr.CopySource{
			TableName: tableName,
			Columns:   columns,
		}
	}

	// Parse FROM or TO
	direction := p.ParseOneOfKeywords([]string{"FROM", "TO"})
	if direction == "" {
		return nil, p.expectedRef("FROM or TO", p.PeekTokenRef())
	}
	copyStmt.To = direction == "TO"

	// Parse target: STDIN, STDOUT, PROGRAM 'cmd', or 'filename'
	switch {
	case p.ParseKeyword("STDIN"):
		copyStmt.Target = &expr.CopyTarget{
			Kind: expr.CopyTargetKindStdin,
		}
	case p.ParseKeyword("STDOUT"):
		copyStmt.Target = &expr.CopyTarget{
			Kind: expr.CopyTargetKindStdout,
		}
	case p.ParseKeyword("PROGRAM"):
		cmd, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		copyStmt.Target = &expr.CopyTarget{
			Kind:    expr.CopyTargetKindProgram,
			Command: cmd,
		}
	default:
		// Must be a filename string
		filename, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		copyStmt.Target = &expr.CopyTarget{
			Kind:     expr.CopyTargetKindFile,
			Filename: filename,
		}
	}

	// Parse optional WITH (ignored for compatibility)
	p.ParseKeyword("WITH")

	// Parse options: (option, option, ...)
	if p.ConsumeToken(token.TokenLParen{}) {
		for {
			if _, ok := p.PeekToken().Token.(token.TokenRParen); ok {
				p.AdvanceToken() // consume )
				break
			}

			opt, err := parseCopyOption(p)
			if err != nil {
				return nil, err
			}
			copyStmt.Options = append(copyStmt.Options, opt)

			if !p.ConsumeToken(token.TokenComma{}) {
				if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
					return nil, err
				}
				break
			}
		}
	}

	// Parse legacy options (space-separated keywords with optional values)
	for {
		opt, err := tryParseCopyLegacyOption(p)
		if err != nil || opt == nil {
			break
		}
		copyStmt.LegacyOptions = append(copyStmt.LegacyOptions, opt)
	}

	// TODO: Parse TSV values if target is STDIN

	return copyStmt, nil
}

// parseCopyOption parses a single COPY option (PostgreSQL 9.0+ format)
func parseCopyOption(p *Parser) (*expr.CopyOption, error) {
	kw := p.ParseOneOfKeywords([]string{
		"FORMAT", "FREEZE", "DELIMITER", "NULL", "HEADER",
		"QUOTE", "ESCAPE", "FORCE_QUOTE", "FORCE_NOT_NULL", "FORCE_NULL", "ENCODING",
	})

	switch kw {
	case "FORMAT":
		ident, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionFormat,
			Value:      ident,
		}, nil
	case "FREEZE":
		val := true
		if p.ParseKeyword("FALSE") {
			val = false
		} else if p.ParseKeyword("TRUE") {
			val = true
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionFreeze,
			Value:      val,
		}, nil
	case "DELIMITER":
		ch, err := p.ParseLiteralChar()
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionDelimiter,
			Value:      string(ch),
		}, nil
	case "NULL":
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionNull,
			Value:      val,
		}, nil
	case "HEADER":
		val := true
		if p.ParseKeyword("FALSE") {
			val = false
		} else if p.ParseKeyword("TRUE") {
			val = true
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionHeader,
			Value:      val,
		}, nil
	case "QUOTE":
		ch, err := p.ParseLiteralChar()
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionQuote,
			Value:      string(ch),
		}, nil
	case "ESCAPE":
		ch, err := p.ParseLiteralChar()
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionEscape,
			Value:      string(ch),
		}, nil
	case "FORCE_QUOTE":
		cols, err := parseCopyColumnList(p)
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionForceQuote,
			Value:      cols,
		}, nil
	case "FORCE_NOT_NULL":
		cols, err := parseCopyColumnList(p)
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionForceNotNull,
			Value:      cols,
		}, nil
	case "FORCE_NULL":
		cols, err := parseCopyColumnList(p)
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionForceNull,
			Value:      cols,
		}, nil
	case "ENCODING":
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyOption{
			OptionType: expr.CopyOptionEncoding,
			Value:      val,
		}, nil
	}

	return nil, p.expectedRef("COPY option", p.PeekTokenRef())
}

// parseCopyColumnList parses a parenthesized column list for FORCE_QUOTE, etc.
func parseCopyColumnList(p *Parser) ([]*ast.Ident, error) {
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	var cols []*ast.Ident
	for {
		if _, ok := p.PeekToken().Token.(token.TokenRParen); ok {
			p.AdvanceToken() // consume )
			break
		}
		col, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		cols = append(cols, col)
		if !p.ConsumeToken(token.TokenComma{}) {
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
			break
		}
	}

	return cols, nil
}

// tryParseCopyLegacyOption tries to parse a legacy COPY option. Returns nil if no option found.
func tryParseCopyLegacyOption(p *Parser) (*expr.CopyLegacyOption, error) {
	// Check for FORMAT [ AS ] (handled specially at the beginning)
	if p.ParseKeyword("FORMAT") {
		p.ParseKeyword("AS")
	}

	kw := p.ParseOneOfKeywords([]string{
		"BINARY", "CSV", "DELIMITER", "ESCAPE", "HEADER", "JSON",
		"NULL", "PARQUET", "GZIP", "BZIP2", "ZSTD", "EMPTYASNULL", "BLANKSASNULL",
		"REMOVEQUOTES", "ADDQUOTES", "IGNOREHEADER", "DATEFORMAT", "TIMEFORMAT",
		"TRUNCATECOLUMNS", "COMPUPDATE", "STATUPDATE", "PARALLEL", "MAXFILESIZE",
		"REGION", "IAM_ROLE", "MANIFEST", "CREDENTIALS", "FIXEDWIDTH", "EXTENSION",
		"ACCEPTANYDATE", "ACCEPTINVCHARS", "ALLOWOVERWRITE", "CLEANPATH", "ENCRYPTED",
		"ROWGROUPSIZE", "PARTITION", "PARTITIONBY",
	})

	switch kw {
	case "BINARY":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionBinary}, nil
	case "CSV":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionCsv}, nil
	case "DELIMITER":
		p.ParseKeyword("AS")
		ch, err := p.ParseLiteralChar()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionDelimiter,
			Value:      string(ch),
		}, nil
	case "ESCAPE":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionEscape}, nil
	case "HEADER":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionHeader}, nil
	case "JSON":
		p.ParseKeyword("AS")
		if tok, ok := p.PeekToken().Token.(token.TokenSingleQuotedString); ok {
			p.AdvanceToken()
			return &expr.CopyLegacyOption{
				OptionType: expr.CopyLegacyOptionJson,
				Value:      tok.Value,
			}, nil
		}
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionJson}, nil
	case "NULL":
		p.ParseKeyword("AS")
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionNull,
			Value:      val,
		}, nil
	case "PARQUET":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionParquet}, nil
	case "GZIP":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionGzip}, nil
	case "BZIP2":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionBzip2}, nil
	case "ZSTD":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionZstd}, nil
	case "EMPTYASNULL":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionEmptyAsNull}, nil
	case "BLANKSASNULL":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionBlankAsNull}, nil
	case "REMOVEQUOTES":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionRemoveQuotes}, nil
	case "ADDQUOTES":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionAddQuotes}, nil
	case "IGNOREHEADER":
		p.ParseKeyword("AS")
		num, err := p.ParseNumber()
		if err != nil {
			return nil, err
		}
		// Convert to int
		val := 0
		fmt.Sscanf(num, "%d", &val)
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionIgnoreHeader,
			Value:      val,
		}, nil
	case "DATEFORMAT":
		p.ParseKeyword("AS")
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionDateFormat,
			Value:      val,
		}, nil
	case "TIMEFORMAT":
		p.ParseKeyword("AS")
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionTimeFormat,
			Value:      val,
		}, nil
	case "TRUNCATECOLUMNS":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionTruncateColumns}, nil
	case "COMPUPDATE":
		val := ""
		if p.ParseKeyword("PRESET") {
			val = "PRESET"
		} else if p.ParseKeyword("ON") || p.ParseKeyword("TRUE") {
			val = "TRUE"
		} else if p.ParseKeyword("OFF") || p.ParseKeyword("FALSE") {
			val = "FALSE"
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionCompUpdate,
			Value:      val,
		}, nil
	case "STATUPDATE":
		val := ""
		if p.ParseKeyword("ON") || p.ParseKeyword("TRUE") {
			val = "TRUE"
		} else if p.ParseKeyword("OFF") || p.ParseKeyword("FALSE") {
			val = "FALSE"
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionStatUpdate,
			Value:      val,
		}, nil
	case "PARALLEL":
		val := ""
		if p.ParseKeyword("ON") || p.ParseKeyword("TRUE") {
			val = "TRUE"
		} else if p.ParseKeyword("OFF") || p.ParseKeyword("FALSE") {
			val = "FALSE"
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionParallel,
			Value:      val,
		}, nil
	case "MAXFILESIZE":
		p.ParseKeyword("AS")
		val, err := p.ParseNumber()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionMaxFileSize,
			Value:      val,
		}, nil
	case "REGION":
		p.ParseKeyword("AS")
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionRegion,
			Value:      val,
		}, nil
	case "MANIFEST":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionManifest}, nil
	case "CREDENTIALS":
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionCredentials,
			Value:      val,
		}, nil
	case "FIXEDWIDTH":
		p.ParseKeyword("AS")
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionFixedWidth,
			Value:      val,
		}, nil
	case "EXTENSION":
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionExtension,
			Value:      val,
		}, nil
	case "ACCEPTANYDATE":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionAcceptAnyDate}, nil
	case "ACCEPTINVCHARS":
		p.ParseKeyword("AS")
		if tok, ok := p.PeekToken().Token.(token.TokenSingleQuotedString); ok {
			p.AdvanceToken()
			return &expr.CopyLegacyOption{
				OptionType: expr.CopyLegacyOptionAcceptInvChars,
				Value:      tok.Value,
			}, nil
		}
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionAcceptInvChars}, nil
	case "ALLOWOVERWRITE":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionAllowOverwrite}, nil
	case "CLEANPATH":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionCleanPath}, nil
	case "ENCRYPTED":
		return &expr.CopyLegacyOption{OptionType: expr.CopyLegacyOptionEncrypted}, nil
	case "ROWGROUPSIZE":
		p.ParseKeyword("AS")
		val, err := p.ParseNumber()
		if err != nil {
			return nil, err
		}
		return &expr.CopyLegacyOption{
			OptionType: expr.CopyLegacyOptionRowGroupSize,
			Value:      val,
		}, nil
	}

	return nil, nil // No legacy option found
}
