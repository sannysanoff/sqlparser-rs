package parser

import (
	"fmt"

	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/token"
)

// parseJsonAccess parses JSON/semi-structured data access expressions like a:b or a.b:c:d[0]
// This is the Go equivalent of parse_json_access in Rust.
func (ep *ExpressionParser) parseJsonAccess(left expr.Expr) (expr.Expr, error) {
	path, err := ep.parseJsonPath()
	if err != nil {
		return nil, err
	}
	// Calculate span from the left expression to the end of the last path element
	spanEnd := left.Span()
	if len(path.Path) > 0 {
		lastElem := path.Path[len(path.Path)-1]
		// Get span from the element based on its type
		switch e := lastElem.(type) {
		case *expr.JsonPathBracket:
			spanEnd = e.Key.Span()
		case *expr.JsonPathColonBracket:
			spanEnd = e.Key.Span()
		}
	}
	return &expr.JsonAccess{
		SpanVal: mergeSpans(left.Span(), spanEnd),
		Value:   left,
		Path:    path,
	}, nil
}

// parseJsonPath parses a JSON path for semi-structured data access.
// This handles patterns like :key, :"quoted key", .field, [index], :[expression]
// This is the Go equivalent of parse_json_path in Rust.
func (ep *ExpressionParser) parseJsonPath() (*expr.JsonPath, error) {
	var pathElems []expr.JsonPathElem

	for {
		tok := ep.parser.NextToken()

		switch tok.Token.(type) {
		case token.TokenColon:
			if len(pathElems) == 0 {
				// First colon - check if it's :[expr] form
				nextTok := ep.parser.PeekTokenRef()
				if _, ok := nextTok.Token.(token.TokenLBracket); ok {
					// :[expr] form
					ep.parser.NextToken() // consume [
					key, err := ep.ParseExpr()
					if err != nil {
						return nil, err
					}
					if _, err := ep.parser.ExpectToken(token.TokenRBracket{}); err != nil {
						return nil, err
					}
					pathElems = append(pathElems, &expr.JsonPathColonBracket{Key: key})
				} else {
					// :key form (colon followed by identifier)
					elem, err := ep.parseJsonPathObjectKey()
					if err != nil {
						return nil, err
					}
					pathElems = append(pathElems, elem)
				}
			} else {
				// Subsequent colons are treated as path separators
				elem, err := ep.parseJsonPathObjectKey()
				if err != nil {
					return nil, err
				}
				pathElems = append(pathElems, elem)
			}

		case token.TokenPeriod:
			if len(pathElems) == 0 {
				// Period can't start a path (we need colon first)
				ep.parser.PrevToken()
				return nil, fmt.Errorf("expected colon to start JSON path, found period")
			}
			elem, err := ep.parseJsonPathObjectKey()
			if err != nil {
				return nil, err
			}
			pathElems = append(pathElems, elem)

		case token.TokenLBracket:
			key, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			if _, err := ep.parser.ExpectToken(token.TokenRBracket{}); err != nil {
				return nil, err
			}
			pathElems = append(pathElems, &expr.JsonPathBracket{Key: key})

		default:
			// Not a JSON path element, put back the token and exit
			ep.parser.PrevToken()
			if len(pathElems) == 0 {
				return nil, fmt.Errorf("expected JSON path after colon")
			}
			return &expr.JsonPath{Path: pathElems}, nil
		}
	}
}

// parseJsonPathObjectKey parses an object key in a JSON path (identifier or quoted string).
// This is the Go equivalent of parse_json_path_object_key in Rust.
func (ep *ExpressionParser) parseJsonPathObjectKey() (expr.JsonPathElem, error) {
	tok := ep.parser.NextToken()

	switch t := tok.Token.(type) {
	case token.TokenWord:
		// Unquoted or quoted identifier
		return &expr.JsonPathDot{
			Key:    t.Word.Value,
			Quoted: t.Word.QuoteStyle != nil,
		}, nil

	case token.TokenDoubleQuotedString:
		return &expr.JsonPathDot{
			Key:    t.Value,
			Quoted: true,
		}, nil

	case token.TokenSingleQuotedString:
		// Snowflake allows single-quoted strings as keys
		return &expr.JsonPathDot{
			Key:    t.Value,
			Quoted: true,
		}, nil

	default:
		return nil, fmt.Errorf("expected variant object key name, found: %s", tok.Token.String())
	}
}
