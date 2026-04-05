# Go Implementation Plan for sqlparser-rs

Complete re-implementation of sqlparser-rs in Go using automated transpilation.

**Project Scope:** ~117,000 lines of Rust → ~82,000 lines of Go (70%)
**Target:** Full feature parity with all 14 dialects and 1,260+ tests  
**Approach:** Automated transpilation with interface-based AST design

---

## Critical Implementation Rule

⚠️ **ALWAYS USE RUST IMPLEMENTATION AS REFERENCE** ⚠️

When implementing parser functionality:

1. **First, examine the Rust source** (`src/parser/mod.rs`, `src/ast/*.rs`, etc.)
2. **Port the logic directly** - do not reinvent or redesign
3. **Follow Rust naming conventions** where possible (e.g., `parse_create_view` → `parseCreateView`)
4. **Preserve behavior exactly** - edge cases, error messages, dialect-specific handling
5. **Comment with references** - cite the Rust file/line for complex logic

---

## Import Structure

The project uses a consolidated import structure after refactoring:

```go
import (
    "github.com/datafuselabs/sqlparser-go/ast"
    "github.com/datafuselabs/sqlparser-go/ast/expr"
    "github.com/datafuselabs/sqlparser-go/ast/statement"
    "github.com/datafuselabs/sqlparser-go/ast/query"
    "github.com/datafuselabs/sqlparser-go/parseriface"
    "github.com/datafuselabs/sqlparser-go/token"
    "github.com/datafuselabs/sqlparser-go/dialects"
)
```

The `token` package now contains both keywords and lexer/tokenizer functionality (merged from the old `tokenizer` and `span` packages). The `parseriface` package provides shared interfaces to avoid circular dependencies between `parser` and `dialects` packages.

---

## Typical Code Editing Errors and Patterns

This section documents common mistakes encountered when porting Rust to Go, categorized by pattern type. Understanding these patterns helps prevent similar bugs during implementation.

### Pattern A: Using Literal Tokens Instead of Constants

**Problem:** Hardcoding token values instead of using the token constant types.

**Example - parseObjectName (Dots vs Commas):**

```go
// INCORRECT - parsed comma-separated instead of dot-separated
func (ep *ExpressionParser) parseObjectName() (*expr.ObjectName, error) {
    idents, err := ep.parseCommaSeparatedIdents()  // Wrong! Uses commas
    ...
}

// CORRECT - parse dot-separated qualified names
func (ep *ExpressionParser) parseObjectName() (*expr.ObjectName, error) {
    var parts []*expr.ObjectNamePart
    for {
        ident, err := ep.parseIdentifier()
        if err != nil {
            return nil, err
        }
        parts = append(parts, &expr.ObjectNamePart{...})
        // Check for DOT token constant, not comma
        if !ep.parser.ConsumeToken(token.TokenPeriod{}) {
            break
        }
    }
    return &expr.ObjectName{Parts: parts}, nil
}
```

**Key Lesson:** Always use the proper token constant (`TokenPeriod`, `TokenComma`, etc.) rather than assuming separator behavior. Check Rust reference `src/parser/mod.rs:12715-12740` for `parse_object_name_inner`.

---

### Pattern B: Using Wrong Parser Function for Context

**Problem:** Using a generic identifier parser when the context requires qualified names.

**Example - JOIN USING Clause:**

```go
// INCORRECT - only parses simple identifiers
func parseJoinConstraint(p *Parser) (query.JoinConstraint, error) {
    cols, err := parseCommaSeparatedQueryIdents(p)  // Can only parse "col1"
    // FAILS on: JOIN tbl2 USING(t2.col1)
}

// CORRECT - parses qualified identifiers using proper function
func parseJoinConstraint(p *Parser) (query.JoinConstraint, error) {
    objNames, err := parseCommaSeparatedObjectNames(p)  // Can parse "t2.col1"
    // Convert []*ast.ObjectName to []query.ObjectName
    attrs := make([]query.ObjectName, len(objNames))
    for i, objName := range objNames {
        // ... conversion logic
    }
}
```

**Key Lesson:** When the SQL construct can contain qualified names (table.column), use `parseObjectName()` or `parseCommaSeparatedObjectNames()`, not simple identifier parsers. Reference: `src/parser/mod.rs:16687` for `parse_parenthesized_qualified_column_list`.

---

### Pattern C: Incorrect Token Type in Switch/Case

**Problem:** Using literal rune/byte comparisons instead of token type assertions.

**Example - Array Subscript Tokenizer Bug:**

```go
// INCORRECT - treated '[' as quoted identifier start
func (t *Tokenizer) NextToken() (Token, error) {
    switch state.Current() {
    case '`', '[':  // Wrong! '[' is not a quote character
        return t.tokenizeQuotedIdentifier(state)
    }
}

// CORRECT - '[' is a separate bracket token
func (t *Tokenizer) NextToken() (Token, error) {
    switch state.Current() {
    case '`':
        return t.tokenizeQuotedIdentifier(state)
    case '[':
        state.Next()
        return TokenLBracket{}, nil  // Return proper token type
    }
}
```

**Key Lesson:** Don't assume characters have the same token semantics as Rust. `[` is `Token::LBracket` in Rust (`src/tokenizer.rs:1691`), not a quoted identifier delimiter. Always verify token type mappings.

---

### Pattern D: Missing Dialect Capability Checks

**Problem:** Parsing dialect-specific syntax without checking if the dialect supports it.

```go
// INCORRECT - always tries to parse PostgreSQL syntax
func parseFunctionArgs(p *Parser) ([]expr.FunctionArg, error) {
    // Tries => syntax even for MySQL dialect
    if p.peekIsOperator("=>") {
        // parse named argument
    }
}

// CORRECT - check dialect first
func parseFunctionArgs(p *Parser) ([]expr.FunctionArg, error) {
    if p.GetDialect().SupportsNamedFnArgsWithRArrowOperator() {
        if p.peekIsOperator("=>") {
            // parse => syntax
        }
    }
    // ... fallback to regular args
}
```

**Key Lesson:** Always guard dialect-specific features with `dialect.SupportsXxx()` checks. Reference: `src/parser/mod.rs:17788-17836` for `parse_function_args()`.

---

### Pattern E: Unsafe Type Assertions

**Problem:** Direct type assertions without `ok` check cause panics.

```go
// INCORRECT - will panic if token is wrong type
func parseKeyword(p *Parser) string {
    word := p.PeekToken().Token.(token.TokenWord)  // Panic if not TokenWord
    return word.Value
}

// CORRECT - safe type assertion with ok check
func parseKeyword(p *Parser) (string, error) {
    tok := p.PeekToken()
    if word, ok := tok.Token.(token.TokenWord); ok {
        return word.Value, nil
    }
    return "", fmt.Errorf("expected keyword, got %T", tok.Token)
}
```

**Key Lesson:** Always use the `if x, ok := y.(Type); ok {` pattern. Go type assertions fail hard without the ok check.

---

### Pattern F: Case Sensitivity Assumptions

**Problem:** Assuming all dialects lowercase identifiers.

```go
// INCORRECT - always lowercases
func (i *Ident) String(dialect Dialect) string {
    return strings.ToLower(i.Value)  // Breaks MySQL tests
}

// CORRECT - preserve original case
func (i *Ident) String(dialect Dialect) string {
    // Rust preserves original case for ALL dialects to ensure AST comparison
    return i.Value  // Preserve original
}
```

**Key Lesson:** The Rust implementation preserves original case for ALL dialects to ensure consistent AST comparison. Don't normalize case in the parser. Reference: MySQL dialect in `src/dialect/mysql.rs`.

---

### Pattern G: Premature Token Consumption

**Problem:** Consuming tokens before validating the full pattern.

```go
// INCORRECT - consumes '(' then checks for empty
func parseWindowSpec(p *Parser) (*expr.WindowSpec, error) {
    p.ExpectToken(token.TokenLParen{})  // Consumes it
    if p.peekIsToken(token.TokenRParen{}) {
        // Error: we already consumed '('
        return &expr.WindowSpec{}, nil
    }
    // Now we need '(' again but it's gone
}

// CORRECT - peek first, consume only after validation
func parseWindowSpec(p *Parser) (*expr.WindowSpec, error) {
    // Check for empty spec: OVER ()
    tok := p.PeekToken()
    if _, ok := tok.Token.(token.TokenLParen); ok {
        // Check if next token after ( is )
        nextTok := p.PeekNthToken(1)
        if _, ok := nextTok.Token.(token.TokenRParen); ok {
            // Empty window specification
            p.AdvanceToken() // consume (
            p.AdvanceToken() // consume )
            return &expr.WindowSpec{}, nil
        }
    }
    // Continue parsing non-empty spec...
}
```

**Key Lesson:** Use `PeekToken()` and `PeekNthToken(n)` to look ahead before consuming. Only call `AdvanceToken()` or `ExpectToken()` after validation.

---

### Pattern H: Wrong Expression Return Type

**Problem:** Using `ast.Expr` interface instead of `expr.Expr` for statement fields.

```go
// INCORRECT - type mismatch
stmt := &statement.Pragma{
    Value: &expr.ValueExpr{...},  // *expr.ValueExpr doesn't implement ast.Expr
}

// CORRECT - use expr.Expr for expression fields
type Pragma struct {
    Value expr.Expr  // Not ast.Expr
}

stmt := &statement.Pragma{
    Value: &expr.ValueExpr{Value: ...},  // Works!
}
```

**Key Lesson:** The Go AST has two expression interfaces:

- `expr.Expr` - for expression types in `ast/expr/` (requires: `exprNode()`, `Span()`, `String()`)
- `ast.Expr` - sealed interface for all AST expressions (requires additional methods)

Use `expr.Expr` for statement fields that hold expressions.

---

### Pattern I: Stub Implementation Returns Empty

**Problem:** Stub functions return empty structs instead of parsing.

```go
// INCORRECT - silent failure with empty data
func parseIdentifier(p *Parser) (*ast.Ident, error) {
    return &ast.Ident{}, nil  // Returns empty! No error!
}

// Result: Table="", Alias="t" for "LOCK TABLES t READ"

// CORRECT - actually parse the token
func parseIdentifier(p *Parser) (*ast.Ident, error) {
    tok := p.PeekToken()
    if word, ok := tok.Token.(token.TokenWord); ok {
        p.AdvanceToken()
        return &ast.Ident{Value: word.Word.Value}, nil
    }
    return nil, fmt.Errorf("expected identifier, found %v", tok.Token)
}
```

**Key Lesson:** Always verify helper functions actually parse data. Empty returns without errors are silent bugs that show up as missing data later.

---

### Pattern J: Missing Statement Type Handling

**Problem:** Only handling one statement type when multiple are possible.

```go
// INCORRECT - only handles SelectStatement
func parseCTE(p *Parser) (*query.CTE, error) {
    innerQuery, err := p.parseQuery()
    if err != nil {
        return nil, err
    }

    if selStmt, ok := innerQuery.(*SelectStatement); ok {
        cte.Query = &query.Query{...}
    }
    // BUG: Query is nil if innerQuery is *QueryStatement (nested CTE)
}

// CORRECT - handle all possible query types
func parseCTE(p *Parser) (*query.CTE, error) {
    innerQuery, err := p.parseQuery()
    if err != nil {
        return nil, err
    }

    if selStmt, ok := innerQuery.(*SelectStatement); ok {
        cte.Query = &query.Query{...}
    } else if qStmt, ok := innerQuery.(*QueryStatement); ok {
        // Nested CTE (WITH clause inside CTE)
        cte.Query = qStmt.Query
    }
    // ... handle other cases
}
```

**Key Lesson:** When WITH clause is present, `parseQuery` can return `QueryStatement` instead of `SelectStatement`. All statement types that contain queries must handle both. Reference: `src/parser/mod.rs:13599-13610` for `parse_query()`.

---

### Pattern K: Hardcoded Case in Tokenizer Prefix Functions

**Problem:** Tokenizer functions that handle character-specific cases (like 'U' for Unicode strings) using hardcoded uppercase letters instead of preserving the original case.

**Example - tokenizeUnicodeStringLiteral:**

```go
// INCORRECT - hardcodes uppercase "U"
func (t *Tokenizer) tokenizeUnicodeStringLiteral(state *State) (Token, error) {
    // ...
    state.Next() // consume U/u
    // ...
    word := t.tokenizeWord(state, "U")  // Wrong! Always uses uppercase U
    return MakeWord(word, nil), nil
}

// Result: "uk_cities" becomes "Uk_cities" in the token value
```

```go
// CORRECT - preserve original case
func (t *Tokenizer) tokenizeUnicodeStringLiteral(state *State) (Token, error) {
    ch, _ := state.Peek()  // Save original character with its case
    state.Next() // consume U/u
    // ...
    word := t.tokenizeWord(state, string(ch))  // Use original case
    return MakeWord(word, nil), nil
}

// Result: "uk_cities" stays "uk_cities" - original case preserved
```

**Key Lesson:** When handling character-specific tokenizer cases (like Unicode string literals U&'...', byte strings B'...', etc.), always preserve the original character case. The tokenizer must not normalize case - that's the parser/AST's responsibility if needed. This affects identifiers starting with U, B, R, N, Q, E, X, etc. Reference: `src/tokenizer.rs` for how Rust handles these prefixes.

---

### Pattern L: Hardcoded Prefix Characters in Tokenizer Special Cases

**Problem:** Tokenizer functions that handle special character prefixes (like Q'...', E'...', X'...') using hardcoded uppercase characters instead of preserving the actual input character.

**Example - tokenizeQuoteDelimitedLiteral:**

```go
// INCORRECT - hardcodes uppercase "Q"
func (t *Tokenizer) tokenizeQuoteDelimitedLiteral(state *State) (Token, error) {
    state.Next() // consume Q/q
    next, ok := state.Peek()
    if !ok || next != '\'' {
        word := t.tokenizeWord(state, "Q")  // Wrong! Always uses uppercase Q
        return MakeWord(word, nil), nil
    }
    // ...
}

// Result: "quaternion" becomes "Quaternion" when not followed by quote
```

```go
// CORRECT - preserves original character case
func (t *Tokenizer) tokenizeQuoteDelimitedLiteral(state *State) (Token, error) {
    ch, _ := state.Peek()  // Save original character (Q or q)
    state.Next()           // consume Q/q
    next, ok := state.Peek()
    if !ok || next != '\'' {
        word := t.tokenizeWord(state, string(ch))  // Use original case
        return MakeWord(word, nil), nil
    }
    // ...
}

// Result: "quaternion" stays "quaternion" - original case preserved
```

**Key Lesson:** When tokenizer functions check for special prefixes (Q/q, E/e, X/x, etc.) and need to fall back to regular identifier tokenization, always use the actual consumed character rather than a hardcoded literal. The pattern of:
1. Check current character with `state.Peek()`
2. Save the character before consuming
3. Use `string(ch)` instead of hardcoded "Q"

This affects functions like `tokenizeQuoteDelimitedLiteral`, `tokenizeEscapedStringLiteral`, `tokenizeHexStringLiteral`, and others that handle special prefix characters. Reference: `src/tokenizer.rs` lines 544-628 for Rust reference implementation.

---

### Pattern M: Using Full ObjectName Parser for Simple Qualified Names

**Problem:** Using `ParseObjectName()` for qualified wildcards like `table.*` incorrectly consumes the dot and tries to parse additional identifier parts.

**Example - parseSelectItem (Qualified Wildcards):**

```go
// INCORRECT - consumes table.* incorrectly
func parseSelectItem(p *Parser) (query.SelectItem, error) {
    if isQualifiedWildcard(p) {
        tableName, err := p.ParseObjectName()  // Wrong! Consumes 'table.' and expects more
        // ... fails on 'inserted.*' because it tries to parse another identifier after '.'
    }
}

// CORRECT - parse single identifier only
func parseSelectItem(p *Parser) (query.SelectItem, error) {
    if isQualifiedWildcard(p) {
        // Don't use ParseObjectName() - it consumes the period and tries to parse more parts
        tableIdent, err := p.ParseIdentifier()  // Only parse the identifier
        if err != nil {
            return nil, err
        }
        p.ConsumeToken(token.TokenPeriod{}) // consume the . explicitly
        p.ConsumeToken(token.TokenMul{})    // consume the * explicitly
        
        // Build ObjectName from single identifier
        queryName := query.ObjectName{
            Parts: []query.Ident{{Value: tableIdent.Value}},
        }
        return &query.QualifiedWildcard{
            Kind: &query.ObjectNameWildcard{Name: queryName},
        }, nil
    }
}
```

**Key Lesson:** `ParseObjectName()` is designed for multi-part names like `schema.table.column` and will consume periods greedily. For qualified wildcards like `table.*`, use `ParseIdentifier()` directly and handle the period and star tokens explicitly. Reference: `src/parser/mod.rs` for how Rust handles select item parsing.

---

### Pattern N: FunctionCall Parentheses in String() Output

**Problem:** Function calls without arguments must still output empty parentheses `()` in their String() method for round-trip parsing compatibility.

**Example - FunctionDesc and TriggerExecBody:**

```go
// INCORRECT - outputs "EXECUTE FUNCTION emp_stamp" (no parens)
func (f *FunctionDesc) String() string {
    var sb strings.Builder
    sb.WriteString(f.Name.String())
    if len(f.Args) > 0 {  // Only adds parens when args exist
        sb.WriteString("(")
        // ... write args
        sb.WriteString(")")
    }
    return sb.String()
}
// Result: "CREATE TRIGGER ... EXECUTE FUNCTION emp_stamp" 
// Expected: "CREATE TRIGGER ... EXECUTE FUNCTION emp_stamp()"
// Test fails because re-parsed SQL doesn't match original

// CORRECT - always includes parentheses
func (f *FunctionDesc) String() string {
    var sb strings.Builder
    sb.WriteString(f.Name.String())
    sb.WriteString("(")  // Always add opening paren
    for i, arg := range f.Args {
        if i > 0 {
            sb.WriteString(", ")
        }
        sb.WriteString(arg.String())
    }
    sb.WriteString(")")  // Always add closing paren
    return sb.String()
}
// Result: "CREATE TRIGGER ... EXECUTE FUNCTION emp_stamp()"
// Test passes - SQL matches original
```

**Key Lesson:** Function calls in SQL always have parentheses, even for zero arguments. The String() method must always output `()` regardless of whether there are arguments. This ensures round-trip compatibility where `parse(stringify(AST)) == AST`. Reference: TriggerExecBody and FunctionDesc String() implementations.

---

### Pattern O: Incorrect Empty String vs nil Pointer in AST Fields

**Problem:** Setting a pointer to an empty string `&""` instead of leaving it as `nil` causes incorrect String() output.

**Example - CEIL/FLOOR parsing:**

```go
// INCORRECT - creates "CEIL(1.5 TO )" output
func parseCeilFloorExpr(isCeil bool) (expr.Expr, error) {
    // ...
    } else {
        // CEIL/FLOOR(expr) - simple case
        ceilExpr.Field.Kind = expr.CeilFloorDateTime
        empty := ""  // Wrong! Creates pointer to empty string
        ceilExpr.Field.DateTimeField = &empty
    }
    // String() checks: if Kind == CeilFloorDateTime && DateTimeField != nil
    // This passes because DateTimeField is not nil (points to empty string)
    // Result: fmt.Sprintf("CEIL(%s TO %s)", expr, *DateTimeField) outputs "CEIL(1.5 TO )"
}

// CORRECT - leaves DateTimeField as nil
func parseCeilFloorExpr(isCeil bool) (expr.Expr, error) {
    // ...
    } else {
        // CEIL/FLOOR(expr) - simple case, no DateTimeField
        ceilExpr.Field.Kind = expr.CeilFloorDateTime
        // Don't set DateTimeField - leave it nil so String() outputs simple form
    }
    // String() checks: if Kind == CeilFloorDateTime && DateTimeField != nil
    // This fails because DateTimeField is nil
    // Falls through to: fmt.Sprintf("CEIL(%s)", expr) outputs "CEIL(1.5)"
}
```

**Key Lesson:** When an AST field is optional (pointer type), leave it as `nil` when not present, don't set it to a pointer to an empty value. The String() methods typically check `!= nil` to determine if the field should be rendered. Reference: `src/parser/mod.rs` and AST String() implementations.

---

### Pattern O: Dollar-Quoted String Literals (PostgreSQL)

**Problem:** PostgreSQL dollar-quoted strings like `$$ ... $$` or `$tag$ ... $tag$` are not recognized by the tokenizer, causing parsing failures for function bodies.

**Example:**

```sql
-- INCORRECT - parser fails on $$
CREATE FUNCTION foo() RETURNS TEXT AS $$ SELECT 1 $$ LANGUAGE SQL;
-- Error: Expected: string literal, found: $$
```

**Key Lesson:** Dollar-quoted strings are a PostgreSQL-specific feature where the content between delimiters is treated as a raw string literal. The tokenizer needs to support these delimiters. Reference: `src/tokenizer.rs` for how Rust handles dollar-quoted strings.

---

### Pattern P: Parser Method Naming Conventions

**Problem:** Using non-existent methods like `p.PeekTokenIs()` or `parseDataType(p)` instead of the correct parser API.

**Example - Correct Parser API Usage:**

```go
// INCORRECT - using non-existent methods
if p.PeekTokenIs(token.TokenLParen{}) {  // Method doesn't exist!
    dataType, err := parseDataType(p)    // Function doesn't exist!
}

// CORRECT - using proper parser API
if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {  // Type assertion
    dataType, err := p.ParseDataType()  // Method on Parser struct
}
```

**Key Lesson:** The Go parser uses these conventions:
- **Token type checking**: Use type assertions on `p.PeekToken().Token`: `if _, ok := p.PeekToken().Token.(token.TokenLParen); ok`
- **Consuming tokens**: Use `p.ConsumeToken(token.TokenLParen{})` for optional tokens, `p.ExpectToken()` for required tokens
- **Data type parsing**: Use `p.ParseDataType()` (method on Parser, not standalone function)
- **Keyword checking**: Use `p.PeekKeyword("KEYWORD")` to check, `p.ExpectKeyword("KEYWORD")` to consume
- **Expression parsing**: Use `NewExpressionParser(p).ParseExpr()` for expressions

Reference: `parser/utils.go`, `parser/parser.go` for parser API methods.

---

### Pattern Q: Discarding Parsed Keywords Without Storing in AST

**Problem:** Parser consumes keywords like DISTINCT but doesn't store them, causing incorrect String() output.

**Example - SELECT DISTINCT:**

```go
// INCORRECT - keyword consumed but not stored
func parseSelect(p *Parser) (ast.Statement, error) {
    p.ExpectKeyword("SELECT")
    p.ParseKeyword("DISTINCT")  // Consumed but discarded!
    p.ParseKeyword("ALL")        // Consumed but discarded!
    projection, _ := parseProjection(p)
    return &SelectStatement{Select: query.Select{
        Projection: projection,  // Distinct field is nil!
    }}, nil
}
// Result: "SELECT DISTINCT name" becomes "SELECT name" in output

// CORRECT - store the parsed distinct/all values
func parseSelect(p *Parser) (ast.Statement, error) {
    p.ExpectKeyword("SELECT")
    var distinct *query.Distinct
    if p.ParseKeyword("DISTINCT") {
        d := query.DistinctDistinct
        distinct = &d
    } else if p.ParseKeyword("ALL") {
        d := query.DistinctAll
        distinct = &d
    }
    projection, _ := parseProjection(p)
    return &SelectStatement{Select: query.Select{
        Distinct:   distinct,     // Now properly stored!
        Projection: projection,
    }}, nil
}
// Result: "SELECT DISTINCT name" stays as "SELECT DISTINCT name"
```

**Key Lesson:** Always store parsed modifiers in the AST. Check the Rust implementation to see which fields should be set for each parsed keyword. Reference: `src/parser/mod.rs` for keyword handling.

---

### Pattern R: Not Preserving QuoteStyle in Identifiers

**Problem:** When parsing quoted identifiers like `"table"`, the QuoteStyle is not preserved in the AST, causing incorrect output.

**Example - Quoted Identifiers:**

```go
// INCORRECT - drops quote information
func (p *Parser) ParseIdentifier() (*ast.Ident, error) {
    tok := p.PeekToken()
    if word, ok := tok.Token.(token.TokenWord); ok {
        p.AdvanceToken()
        return &ast.Ident{Value: word.Word.Value}, nil  // QuoteStyle lost!
    }
    return nil, fmt.Errorf("expected identifier")
}
// Result: "table" (with quotes) becomes table (without quotes)

// CORRECT - preserve QuoteStyle
func (p *Parser) ParseIdentifier() (*ast.Ident, error) {
    tok := p.PeekToken()
    if word, ok := tok.Token.(token.TokenWord); ok {
        p.AdvanceToken()
        ident := &ast.Ident{Value: word.Word.Value}
        if word.Word.QuoteStyle != nil {
            quoteStyle := rune(*word.Word.QuoteStyle)
            ident.QuoteStyle = &quoteStyle  // Preserve it!
        }
        return ident, nil
    }
    return nil, fmt.Errorf("expected identifier")
}
// Result: "table" stays as "table" in String() output
```

**Key Lesson:** The tokenizer stores quote information in `TokenWord.Word.QuoteStyle`. Always copy this to `ast.Ident.QuoteStyle` for proper round-trip serialization. Reference: `ast/ident.go` for QuoteStyle handling.

---

### Pattern S: Multiple Parser Entry Points for Same Syntax

**Problem:** SQL syntax like INTERVAL can be parsed through multiple code paths (typed string literal vs. keyword expression), causing inconsistent behavior and missed validation.

**Example - INTERVAL Parsing:**

```go
// INCORRECT - only handling one entry point
func (ep *ExpressionParser) parseIntervalExpr() (expr.Expr, error) {
    // This handles INTERVAL keyword path
    // But INTERVAL '1 DAY' might be handled by tryParseTypedString!
    // Result: MySQL dialect doesn't error on 'SELECT INTERVAL '1 DAY''
}

// CORRECT - handling all entry points with consistent validation
func (ep *ExpressionParser) tryParseTypedString() (expr.Expr, bool) {
    // ... parse INTERVAL 'value' ...
    
    // For dialects requiring qualifiers, INTERVAL without external unit is invalid
    if !hasExternalUnit && dialects.RequireIntervalQualifier(dialect) {
        restore() // Undo token consumption
        return nil, false // Let parseIntervalExpr handle the error
    }
}

func (ep *ExpressionParser) parseIntervalExpr() (expr.Expr, error) {
    // Check for SECOND TO SECOND (not allowed)
    if leadingField == "SECOND" && hasTOClause {
        return nil, fmt.Errorf("syntax error at word: TO")
    }
    // ...
}
```

**Key Lesson:** When the same SQL syntax can be reached through multiple parser entry points:
1. **Identify all entry points** - e.g., INTERVAL can be parsed via `tryParseTypedString()` (typed literal) or `parseIntervalExpr()` (keyword)
2. **Centralize validation** or duplicate critical checks in all entry points
3. **Use `restore()` to backtrack** when a code path detects an error that should be handled elsewhere
4. **Always reference Rust** - `src/parser/mod.rs:3246-3312` for INTERVAL handling

---

### Pattern T: Whitespace-Skipping Token Methods

**Problem:** `PeekTokenRef()` and `AdvanceToken()` automatically skip whitespace tokens, causing issues when you need to detect newlines or other whitespace.

**Example - Newline-Based String Concatenation:**

```go
// INCORRECT - PeekTokenRef skips whitespace, never sees newlines
func (ep *ExpressionParser) maybeConcatStringLiteral(initial string, initialSpan token.Span) (string, token.Span) {
    for {
        nextTok := ep.parser.PeekTokenRef()  // Skips whitespace!
        switch t := nextTok.Token.(type) {
        case token.TokenWhitespace:
            // NEVER REACHED - PeekTokenRef already skipped whitespace
            if t.Whitespace.Type == token.Newline {
                afterNewline = true
            }
        }
    }
}

// CORRECT - Use PeekTokenNoSkip and NextTokenNoSkip
func (ep *ExpressionParser) maybeConcatStringLiteral(initial string, initialSpan token.Span) (string, token.Span) {
    for {
        nextTok := ep.parser.PeekTokenNoSkip()  // Sees whitespace
        switch t := nextTok.Token.(type) {
        case token.TokenWhitespace:
            if t.Whitespace.Type == token.Newline {
                afterNewline = true
            }
            ep.parser.NextTokenNoSkip()  // Consume whitespace
        case token.TokenSingleQuotedString:
            if afterNewline {
                result += t.Value
                ep.parser.NextTokenNoSkip()  // Consume string
            }
        }
    }
}
```

**Key Lesson:** The parser has two sets of token methods:
- **Whitespace-skipping**: `PeekToken()`, `PeekTokenRef()`, `AdvanceToken()`, `NextToken()` - skip `TokenWhitespace` tokens automatically
- **Whitespace-preserving**: `PeekTokenNoSkip()`, `PeekNthTokenNoSkip()`, `NextTokenNoSkip()` - return all tokens including whitespace

Use whitespace-preserving methods when you need to detect newlines, comments, or other whitespace. Reference: `parser/utils.go` for token method implementations.

---

### Pattern U: Adding New Precedence Levels

**Problem:** Adding a new precedence level (like COLLATE) requires updates to multiple files across the codebase.

**Example - Adding PrecedenceCollate:**

```go
// Step 1: Add to parseriface/parser.go
const (
    PrecedenceOr Precedence = 5
    PrecedenceCollate Precedence = 42  // NEW: between :: and AT TIME ZONE
)

// Step 2: Re-export from dialects/dialect.go
const (
    PrecedenceOr = parseriface.PrecedenceOr
    PrecedenceCollate = parseriface.PrecedenceCollate  // NEW
)

// Step 3: Update ALL dialects' PrecValue() functions
func (d *GenericDialect) PrecValue(prec dialects.Precedence) uint8 {
    switch {
    case prec == dialects.PrecedenceDoubleColon:
        return 50
    case prec == dialects.PrecedenceCollate:  // NEW
        return 42
    case prec == dialects.PrecedenceAtTz:
        return 41
    }
}

// Step 4: Add to GetNextPrecedenceDefault in parser/core.go
case "COLLATE":
    if !ep.parser.InColumnDefinitionState() {
        return dialect.PrecValue(parseriface.PrecedenceCollate), nil
    }

// Step 5: Add to parseWordInfix in parser/infix.go
case "COLLATE":
    collation, err := ep.parseObjectName()
    return &expr.Collate{Expr: base, Collation: collation}, nil
```

**Key Lesson:** When adding new precedence levels:
1. Add the constant to `parseriface/parser.go`
2. Re-export from `dialects/dialect.go`
3. Update **all** dialects' `PrecValue()` functions (13+ dialects)
4. Add handling to `GetNextPrecedenceDefault()` in `parser/core.go`
5. Add infix parsing case to handle the operator

Reference: Rust `src/parser/mod.rs` for precedence handling.

---

### Pattern V: Quoted Strings as Identifiers

**Problem:** `parseIdentifier()` only accepts `TokenWord`, but SQL allows quoted strings as identifiers in certain contexts (e.g., collation names).

**Example - Parsing Collation Names:**

```go
// INCORRECT - only accepts TokenWord
func (ep *ExpressionParser) parseIdentifier() (*expr.Ident, error) {
    tok := ep.parser.NextToken()
    word, ok := tok.Token.(token.TokenWord)
    if !ok {
        return nil, ep.parser.Expected("an identifier", tok)
    }
    return ep.wordToIdent(&word, tok.Span), nil
}
// Result: Fails on SELECT name COLLATE "de_DE" - "de_DE" is TokenDoubleQuotedString

// CORRECT - accept quoted strings as identifiers (following Rust)
func (ep *ExpressionParser) parseIdentifier() (*expr.Ident, error) {
    tok := ep.parser.NextToken()
    switch t := tok.Token.(type) {
    case token.TokenWord:
        return ep.wordToIdent(&t, tok.Span), nil
    case token.TokenSingleQuotedString:
        singleQuote := rune('\'')
        return &expr.Ident{SpanVal: tok.Span, Value: t.Value, QuoteStyle: &singleQuote}, nil
    case token.TokenDoubleQuotedString:
        doubleQuote := rune('"')
        return &expr.Ident{SpanVal: tok.Span, Value: t.Value, QuoteStyle: &doubleQuote}, nil
    }
    return nil, ep.parser.Expected("an identifier", tok)
}
```

**Key Lesson:** The Rust `parse_identifier()` in `src/parser/mod.rs:12926` accepts:
- `Token::Word(w)` - regular identifiers
- `Token::SingleQuotedString(s)` - single-quoted strings
- `Token::DoubleQuotedString(s)` - double-quoted strings

Always check the Rust reference when implementing identifier parsing.

---

### Pattern W: Clause Keywords vs Reserved Keywords for Column Aliases

**Problem:** The trailing comma check in projection parsing incorrectly uses `isReservedForColumnAlias()` which includes expression keywords like `NULL`, `TRUE`, `FALSE`, causing premature termination of the SELECT list.

**Example - SELECT with NULL expression:**

```go
// INCORRECT - stops after first expression when seeing NULL
func parseProjection(p *Parser) ([]query.SelectItem, error) {
    for {
        item, err := parseSelectItem(p)
        items = append(items, item)
        if !p.ConsumeToken(token.TokenComma{}) {
            break
        }
        // WRONG: isReservedForColumnAlias includes NULL, TRUE, etc.
        if isReservedForColumnAlias(word) {
            p.PrevToken()
            break  // Stops here incorrectly!
        }
    }
}
// Result: "SELECT a LIKE b, NULL FROM t" only parses "a LIKE b"
```

```go
// CORRECT - use isClauseKeyword for trailing comma check
func parseProjection(p *Parser) ([]query.SelectItem, error) {
    for {
        item, err := parseSelectItem(p)
        items = append(items, item)
        if !p.ConsumeToken(token.TokenComma{}) {
            break
        }
        // Check if next token is a CLAUSE keyword (FROM, WHERE, etc.)
        // NOT expression keywords like NULL
        if isClauseKeyword(word) {  // Only checks FROM, WHERE, GROUP, etc.
            p.PrevToken()
            break
        }
    }
}
// Result: "SELECT a LIKE b, NULL FROM t" parses both expressions
```

**Key Lesson:** Distinguish between:
- **Clause keywords** (FROM, WHERE, GROUP, ORDER, LIMIT, UNION, etc.) - these signal end of projection
- **Reserved keywords for column aliases** (NULL, TRUE, FALSE, LIKE, etc.) - these can appear in expressions

The trailing comma check should only use clause keywords, not all reserved keywords. Reference: `src/parser/mod.rs:4800-4807` for Rust's `parse_projection`.

---

## Current Status

**Overall Progress: ~42% Test Pass Rate** (~469 tests failing)

| Test Suite       | Status           | Passing | Total | Pass Rate |
| ---------------- | ---------------- | ------- | ----- | --------- |
| **TPC-H**        | ⚠️ Fixture issue  | 0       | 44    | **0%**    |
| **DDL Tests**    | 🔄 In Progress   | ~120    | ~300  | **40%**   |
| **DML Tests**    | 🔄 In Progress   | ~80     | ~150  | **53%**   |
| **Query Tests**  | 🔄 In Progress   | ~150    | ~350  | **43%**   |
| **MySQL**        | 🔄 In Progress   | ~60     | ~125  | **48%**   |
| **PostgreSQL**   | 🔄 In Progress   | ~45     | ~157  | **29%**   |
| **Snowflake**    | 🔄 In Progress   | ~18     | ~97   | **19%**   |
| **TOTAL**        | **~42% Complete** | **~339**| 813   | **~42%** |

**Line Counts:**

- Rust Source: 67,345 lines
- Go Source: 68,233 lines (101% of Rust - added data types)
- Rust Tests: 49,886 lines
- Go Tests: 14,131 lines (28%)

**Recent Focus:** CAST expressions with complex data types, named function arguments, and JSON_OBJECT parsing

---

### April 8, 2026 - Projection Comma Handling and CREATE POLICY/CONNECTOR

Implemented critical parser fixes and completed major missing DDL statement parsers:

1. **SELECT Projection Clause Keyword Fix** (parser/query.go):
   - Fixed `parseProjection()` to use `isClauseKeyword()` instead of `isReservedForColumnAlias()` for trailing comma detection
   - The bug caused premature termination of SELECT lists when expression keywords like `NULL`, `TRUE`, `FALSE` appeared after commas
   - Added new `isClauseKeyword()` function that only checks for clause-starting keywords (FROM, WHERE, GROUP, ORDER, LIMIT, UNION, etc.)
   - **+1 test passing** (TestParseNullLike now correctly parses `SELECT a LIKE NULL AS x, NULL LIKE b AS y FROM t`)
   - **Pattern W** documented: Clause keywords vs Reserved keywords distinction

2. **CREATE POLICY Statement** (parser/create.go, ast/statement/ddl.go, ast/expr/ddl.go):
   - Full implementation per Rust `parse_create_policy` (src/parser/mod.rs:6853)
   - Supports: `CREATE POLICY name ON table [AS PERMISSIVE|RESTRICTIVE] [FOR ALL|SELECT|INSERT|UPDATE|DELETE] [TO role[,...]] [USING (expr)] [WITH CHECK (expr)]`
   - Added new AST types: `CreatePolicyType` (Permissive/Restrictive), `CreatePolicyCommand` (All/Select/Insert/Update/Delete)
   - Added `Owner` type with support for CURRENT_USER, CURRENT_ROLE, SESSION_USER, and identifiers
   - Added `parseOwner()` helper function
   - Fixed `CreatePolicy` String() method to output complete SQL

3. **CREATE CONNECTOR Statement** (parser/create.go, ast/statement/ddl.go):
   - Full implementation per Rust `parse_create_connector` (src/parser/mod.rs:6938)
   - Supports: `CREATE CONNECTOR [IF NOT EXISTS] name [TYPE 'type'] [URL 'url'] [COMMENT 'comment'] [WITH DCPROPERTIES (k=v, ...)]`
   - Fixed `CreateConnector` struct to include all fields: ConnectorType, URL, Comment, WithDCProperties
   - Fixed String() method to output complete SQL with all optional clauses

**Result:** Major parser chunks now implemented. Fixed critical SELECT parsing bug that affected any SELECT with NULL expressions.

---

### April 8, 2026 - Aggregate Function Argument Clauses and Qualified Wildcards

Implemented critical fixes for aggregate/window function argument parsing and qualified wildcards:

1. **ORDER BY Clause in Function Arguments** (parser/special.go):
   - Fixed `parseOrderByClause()` to properly convert `[]*expr.OrderByExpr` to `[]expr.Expr`
   - The parsed ORDER BY expressions were being discarded, causing empty ORDER BY output
   - Now correctly outputs expressions like `FIRST_VALUE(x ORDER BY x)` instead of `FIRST_VALUE(x ORDER BY )`
   - **+4 tests passing** (TestParseAggWithOrderBy and related tests)

2. **Qualified Wildcard in Function Arguments** (parser/core.go):
   - Removed incorrect dialect check `dialects.SupportsSelectWildcardExcept(dialect)` 
   - Qualified wildcards like `COUNT(Employee.*)` are standard SQL and should work in all dialects
   - The check was meant for `SELECT * EXCEPT (cols)` syntax (BigQuery/DuckDB extension), not for `table.*`
   - **+1 test passing** (TestParseCountWildcard)

3. **IGNORE/RESPECT NULLS as FunctionArgumentClause** (parser/special.go):
   - Added `parseNullTreatmentClause()` to parse IGNORE/RESPECT NULLS inside function argument list
   - This is parsed as a `FunctionArgumentClause` (like ORDER BY), not as a suffix after the closing `)`
   - Added break check for IGNORE/RESPECT in the argument parsing loop
   - **+1 test passing** (TestParseWindowFunctionNullTreatmentArg)

**Key Pattern Documentation:**
- **Pattern Y: Function Argument Clauses** - Aggregate/window functions can have multiple clauses inside the parentheses: ORDER BY, LIMIT, IGNORE/RESPECT NULLS, SEPARATOR, ON OVERFLOW, HAVING, RETURNING. These must be parsed as `FunctionArgumentClause` items within `parseFunctionArgs()`, not as separate suffixes after the closing `)`.
- **Pattern Z: Qualified Wildcards Are Standard SQL** - The `table.*` syntax is standard SQL and works in all dialects. Don't gate it behind dialect-specific feature flags meant for extensions like `SELECT * EXCEPT (cols)`.

**Result:** +6 tests now passing. Pass rate improved from 42% to 43% (352/813 tests passing)

---

### April 8, 2026 - CREATE PROCEDURE and CREATE MATERIALIZED VIEW

Implemented parsers for missing CREATE statement types:

1. **CREATE MATERIALIZED VIEW** (parser/create.go):
   - Added `MATERIALIZED` case to `parseCreate()` to handle `CREATE MATERIALIZED VIEW` syntax
   - Updated `parseCreateView()` to accept optional `materialized` parameter
   - The `CreateView` AST type already had `Materialized` field - just needed parser support
   - **+2 tests passing** (TestParseCreateOrReplaceMaterializedView, TestParseCreateMaterializedView)

2. **CREATE PROCEDURE** (parser/create.go, ast/expr/ddl.go):
   - Implemented `parseCreateProcedure()` following Rust `parse_create_procedure` (src/parser/mod.rs:19319)
   - Added `parseProcedureParams()` to parse parameter lists: `(IN a INTEGER, OUT b TEXT)`
   - Added `parseProcedureParam()` to parse individual parameters with mode, name, type, and default
   - Updated `ProcedureParam` AST type to include proper fields: `Name`, `DataType`, `Mode`, `Default`
   - **Result:** Basic CREATE PROCEDURE parsing now works

**Key Pattern Documentation:**
- **Pattern AA: CREATE Statement Routing** - The `parseCreate()` function uses a switch statement on `p.PeekKeyword()` to route to specific CREATE parsers. When adding new CREATE types, add the case before the `default` clause that returns the error.

---

## Current Status

**Overall Progress: ~44% Test Pass Rate** (~457 tests failing)

| Test Suite       | Status           | Passing | Total | Pass Rate |
| ---------------- | ---------------- | ------- | ----- | --------- |
| **TPC-H**        | ⚠️ Fixture issue  | 0       | 44    | **0%**    |
| **DDL Tests**    | 🔄 In Progress   | 22      | 81    | **27%**   |
| **DML Tests**    | 🔄 In Progress   | 18      | 33    | **55%**   |
| **Query Tests**  | 🔄 In Progress   | 39      | 58    | **67%**   |
| **MySQL**        | 🔄 In Progress   | 60      | 125   | **48%**   |
| **PostgreSQL**   | 🔄 In Progress   | 42      | 157   | **27%**   |
| **Snowflake**    | 🔄 In Progress   | 17      | 97    | **18%**   |
| **TOTAL**        | **~44% Complete** | **356** | 813   | **~44%** |

**Line Counts:**

- Rust Source: 67,345 lines
- Go Source: 83,251 lines (124% of Rust - AST types and interfaces added)
- Rust Tests: 49,886 lines
- Go Tests: 14,131 lines (28%)

**Recent Focus:** RETURN, RESET, FLUSH, RENAME, LOCK, DROP FUNCTION statement parsers

---

## Recent Progress

### April 5, 2026 - Major Parser Implementations: RETURN, RESET, FLUSH, RENAME, LOCK, DROP FUNCTION

Implemented critical missing statement parsers to bring maximum test coverage:

1. **RETURN Statement** (parser/parser.go):
   - Implemented `parseReturn()` per Rust `parse_return` (src/parser/mod.rs:19767)
   - Parses optional expression: `RETURN` or `RETURN expr`
   - **+2 tests passing**

2. **RESET Statement** (parser/parser.go):
   - Implemented `parseReset()` per Rust `parse_reset` (src/parser/mod.rs:20076)
   - Supports: `RESET ALL` or `RESET configuration_parameter`
   - **+2 tests passing**

3. **FLUSH Statement** (parser/parser.go):
   - Implemented `parseFlush()` per Rust `parse_flush` (src/parser/mod.rs:972)
   - MySQL/Generic dialect support with FLUSH options: BINARY LOGS, ENGINE LOGS, ERROR LOGS, etc.
   - Supports NO_WRITE_TO_BINLOG and LOCAL modifiers
   - **+3 tests passing**

4. **RENAME TABLE Statement** (parser/parser.go):
   - Implemented `parseRename()` per Rust `parse_rename` (src/parser/mod.rs:1477)
   - Supports multiple table renames: `RENAME TABLE t1 TO t2, t3 TO t4`
   - **+2 tests passing**

5. **LOCK Statement** (parser/parser.go):
   - Implemented `parseLock()` per Rust `parse_lock_statement` (src/parser/mod.rs:18522)
   - PostgreSQL-style LOCK TABLE with optional IN ... MODE and NOWAIT
   - **+2 tests passing**

6. **DROP FUNCTION Statement** (parser/drop.go):
   - Implemented `parseDropFunction()` and `parseFunctionDesc()` per Rust `parse_drop_function` (src/parser/mod.rs:7362)
   - Supports: `DROP FUNCTION [IF EXISTS] name [(args)] [, ...] [CASCADE|RESTRICT]`
   - **+3 tests passing**

**Key Pattern Documentation:**
- **Pattern AB: Statement BaseStatement** - When creating statement structs, use zero values for embedded BaseStatement (e.g., `&statement.Return{Statement: ...}`) rather than explicitly setting `BaseStatement: ast.BaseStatement{}`. The Go AST types use embedding with zero values.
- **Pattern AC: Dialect Names as Strings** - Use string literals like "mysql", "generic" for dialect comparisons rather than constants. The dialect system returns string names via `Dialect()` method.

**Result:** +14 tests now passing. "Not yet implemented" errors eliminated for RETURN, RESET, FLUSH, RENAME, LOCK, DROP FUNCTION.

---

### April 7, 2026 - Data Type Parsing and JSON_OBJECT Named Arguments

Implemented critical fixes for data type parsing and named function argument support:

1. **DEC vs DECIMAL Type Preservation** (parser/parser.go):
   - Fixed `parseDecimalType()` to accept the original type name parameter
   - Now correctly returns `DecType` for "DEC" and `DecimalType` for "DECIMAL"
   - Updated `ParseDataType()` to pass the original keyword to `parseDecimalType()`
   - **+1 test passing** (TestParseCast - DEC now stays as DEC in output)

2. **UNSIGNED INTEGER Data Type Support** (parser/parser.go):
   - Added case for "UNSIGNED" keyword in `ParseDataType()` switch
   - Supports both `UNSIGNED INTEGER` and `UNSIGNED INT` syntax (MySQL style)
   - Falls back to plain `UNSIGNED` type if no INTEGER/INT follows
   - **+1 test passing** (TestParseCastIntegers - MySQL UNSIGNED INTEGER now works)

3. **Colon Operator for Named Function Arguments** (parser/core.go, parser/infix.go):
   - Added colon (`:`) to the expression parsing break list in `ParseExprWithPrecedence()`
   - The colon is now treated as an expression boundary, not an infix operator
   - This allows the function argument parser to handle `:` as a named argument separator
   - Required for MSSQL JSON_OBJECT syntax: `JSON_OBJECT('key' : value)`
   - JSON_OBJECT now parses successfully (though String() output uses `=>` instead of `:` - AST enhancement needed)

**Key Pattern Documentation:**
- **Pattern W: Data Type Name Preservation** - When multiple SQL keywords map to the same conceptual type (DEC/DECIMAL, INT/INTEGER), the parser must track the original keyword used to preserve round-trip serialization.
- **Pattern X: Token as Expression Boundary** - Tokens like `:` that have special meaning in specific contexts (function arguments) should be treated as expression boundaries in general expression parsing by adding them to the break list in `ParseExprWithPrecedence()`.

---

### April 7, 2026 - COLLATE and String Literal Concatenation

Implemented comprehensive support for COLLATE expressions and string literal concatenation:

1. **COLLATE Expression Support** (parser/core.go, parser/infix.go, parseriface/parser.go):
   - Added `PrecedenceCollate` constant to parseriface (value 42, between `::` and AT TIME ZONE)
   - Updated all dialects' `PrecValue()` functions to handle `PrecedenceCollate`
   - Added COLLATE case to `GetNextPrecedenceDefault()` in core.go
   - Added COLLATE case to `parseWordInfix()` in infix.go
   - Fixed `parseIdentifier()` in prefix.go to accept quoted strings as identifiers (required for collation names like `"de_DE"`)
   - **+2 tests passing** (TestParseCollate, TestParseCollateAfterParens)

2. **String Literal Concatenation** (parser/prefix.go, parseriface/parser.go, dialects/capabilities.go):
   - Implemented `maybeConcatStringLiteral()` with two modes:
     - **Adjacent concatenation** (MySQL, ClickHouse, etc.): `'a' 'b'` → `'ab'`
     - **Newline-based concatenation** (Redshift): `'a'\n'b'` → `'ab'`
   - Added helper functions `SupportsStringLiteralConcatenation()` and `SupportsStringLiteralConcatenationWithNewline()` to capabilities.go
   - **Critical fix**: Use `PeekTokenNoSkip()` and `NextTokenNoSkip()` for newline detection (regular `PeekTokenRef()` skips whitespace!)
   - Added `PeekTokenNoSkip()` and `PeekNthTokenNoSkip()` to parseriface.Parser interface
   - **+2 tests passing** (TestParseAdjacentStringLiteralConcatenation, TestParseStringLiteralConcatenationWithNewline)

**Key Pattern Documentation:**
- **Pattern T: Whitespace-Skipping Token Methods** - `PeekTokenRef()` and `AdvanceToken()` automatically skip whitespace tokens. To see whitespace (including newlines), use `PeekTokenNoSkip()` and `NextTokenNoSkip()`.
- **Pattern U: Precedence Constants** - When adding new precedence levels, add to parseriface.Precedence enum, re-export from dialects/dialect.go, and update ALL dialects' PrecValue() functions.
- **Pattern V: Quoted String Identifiers** - `parseIdentifier()` must accept `TokenSingleQuotedString` and `TokenDoubleQuotedString` as valid identifiers (not just `TokenWord`), following Rust `parse_identifier()` in `src/parser/mod.rs:12926`.

**Result:** +4 tests now passing (COLLATE and string concatenation)

---

### April 5, 2026 - CAST Expression, Data Types, and Table-Valued Functions

Implemented major parser fixes for CAST expressions, added comprehensive data type support, and fixed table-valued function parsing:

1. **CAST Expression Fix** (parser/helpers.go):
   - Fixed `parseCastExpr()` to use `ParseDataType()` instead of `parseIdentifier()`
   - This enables parsing complex types like `NVARCHAR(50)`, `CLOB(100)`, etc.
   - **Critical fix**: CAST now properly handles parameterized data types

2. **Extended Data Type Support** (parser/parser.go, ast/datatype/datatype.go):
   - Added `NVARCHAR`, `NCHAR`, `VARCHAR2`, `NVARCHAR2` character types
   - Added `CLOB`, `BLOB` large object types  
   - Added `BINARY`, `VARBINARY` binary types
   - Added `TIME` with optional precision and `WITH TIME ZONE` clause
   - Added `NcharType`, `Varchar2Type`, `Nvarchar2Type` structs to datatype package
   - All new types properly handle optional length/precision parameters

3. **Table-Valued Function Support** (parser/query.go):
   - Added support for table-valued functions in FROM clause: `SELECT * FROM fn()`
   - Updated `parseTableName()` to detect `(` after table name and parse as function
   - Added `parseTableFunctionArgs()` helper for parsing function arguments
   - Reference: `src/parser/mod.rs:15730-15735` for Rust implementation pattern

4. **Parser Backtracking Support** (parser/utils.go, parseriface/parser.go):
   - Added `SetCurrentIndex()` method to Parser interface for token position restoration
   - Enables proper backtracking when trying alternative parsing strategies
   - Used by `parsePositionExpr()` for Snowflake-style POSITION function handling

5. **POSITION Expression Fix** (parser/helpers.go):
   - Implemented backtracking in `parsePositionExpr()` for Snowflake 3-arg syntax
   - When special `POSITION(substr IN str)` syntax fails, falls back to function call
   - Uses `SavePosition()` pattern from Rust's `maybe_parse`

**Key Pattern Documentation:**
- **Pattern W: CAST Data Type Parsing** - CAST expressions must use `ParseDataType()` not `parseIdentifier()` to handle complex types like `NVARCHAR(50)` or `DECIMAL(10,2)`.
- **Pattern X: Table-Valued Functions** - After parsing a table name, check for `(` to detect table-valued functions. Reference: `src/parser/mod.rs:15730-15735`.
- **Pattern Y: Parser Backtracking** - Use `SavePosition()` or `GetCurrentIndex()/SetCurrentIndex()` to implement backtracking. This is the Go equivalent of Rust's `maybe_parse`.

**Result:** Table-valued functions now parsing correctly. CAST with complex data types working. +1 test passing (TestParseNullaryTableValuedFunction).

**Line Counts (Updated):**
- Rust Source: 67,345 lines
- Go Source: 68,201 lines (101% of Rust - increased due to type additions)
- Go Tests: 14,131 lines

---

Implemented comprehensive fixes for INTERVAL expression parsing following Rust reference (`src/parser/mod.rs:3246-3312`):

1. **INTERVAL Require Unit Fix** (parser/prefix.go, parser/helpers.go):
   - Fixed `tryParseTypedString` to check `RequireIntervalQualifier()` before returning INTERVAL without unit
   - For dialects like MySQL that require interval qualifiers, INTERVAL without external unit now correctly errors
   - **+1 test passing** (TestParseIntervalRequireUnit)

2. **INTERVAL TO Clause Support** (parser/prefix.go):
   - Added full support for YEAR TO MONTH, DAY TO HOUR, MINUTE TO SECOND, etc.
   - Properly parses precision on both sides: `MINUTE (5) TO SECOND (5)`
   - **+10 tests passing** (INTERVAL range expressions)

3. **SECOND Special Format** (parser/prefix.go, ast/expr/complex.go):
   - Implemented SQL-mandated special format for SECOND: `SECOND [( <leading precision> [ , <fractional seconds precision>] )]`
   - `SECOND TO SECOND` is now correctly rejected (must use `SECOND (n, m)` format)
   - Updated `IntervalExpr.String()` to output correct format for SECOND with both precisions
   - **+3 tests passing** (SECOND precision tests)

4. **Precision Validation** (parser/prefix.go, parser/helpers.go):
   - Added validation to reject precision on both leading and last field for non-SECOND units
   - `HOUR (1) TO HOUR (2)` now correctly errors
   - **+2 tests passing** (precision validation tests)

**Key Pattern Documentation:**
- **Multiple Parser Entry Points**: INTERVAL expressions can be parsed via `tryParseTypedString` (for typed literals) or `parseIntervalExpr` (for keyword-based). Both must handle dialect-specific validation.
- **SECOND is Special**: SECOND has unique SQL syntax rules that differ from other temporal units. Always check Rust reference for unit-specific handling.
- **Restore on Error**: When `tryParseTypedString` detects an error condition (like SECOND TO SECOND), use `restore()` to undo token consumption and return `(nil, false)` to let the normal parser handle the error.

**Result:** +16 INTERVAL tests now passing (all 7 INTERVAL test functions pass)

### April 6, 2026 - ASSERT, DISCARD, COMMENT, LOAD, and CREATE Statement Parsers

Implemented major missing statement parsers following Rust reference:

1. **AST Type Fixes** (go/ast/expr/ddl.go):
   - Completed `DiscardObject` enum with ALL, PLANS, SEQUENCES, TEMP values
   - Completed `CommentObject` enum with all 16 object types (COLUMN, TABLE, VIEW, etc.)
   - Added proper String() methods for both types

2. **ASSERT Statement** (parser.go):
   - Simple implementation: `ASSERT condition [AS message]`
   - Parses condition expression and optional AS message clause

3. **DISCARD Statement** (parser.go):
   - PostgreSQL-style: `DISCARD { ALL | PLANS | SEQUENCES | TEMP }`
   - Uses updated DiscardObject enum types

4. **LOAD Statement** (parser.go):
   - DuckDB variant: `LOAD extension_name`
   - Hive variant: `LOAD DATA [LOCAL] INPATH 'path' [OVERWRITE] INTO TABLE table_name`
   - Checks dialect capabilities before parsing

5. **COMMENT Statement** (parser.go):
   - Full implementation: `COMMENT ON [object_type] object_name IS 'comment' | NULL`
   - Supports all 16 object types including MATERIALIZED VIEW
   - Handles IF EXISTS modifier and NULL comments

6. **CREATE Statement Parsers** (create.go):
   - **CREATE USER**: Basic parsing with IF NOT EXISTS, name, and key=value options
   - **CREATE TYPE**: Supports AS ENUM, AS RANGE, and composite types (attr1 type1, ...)
   - **CREATE DOMAIN**: Parses AS data_type [COLLATE] [DEFAULT] [constraints]
   - **CREATE EXTENSION**: Parses IF NOT EXISTS, name, WITH [SCHEMA] [VERSION] [CASCADE]

**Key Pattern Documentation:**
- **AST Type Completion**: When enum types are incomplete (only having a `None` value), complete them with proper values and String() methods before implementing parsers that use them.
- **Token Type Checking**: Use `p.PeekToken().Token.(token.Type)` for type assertions, but use `p.PeekTokenRef()` when passing to functions expecting `*token.TokenWithSpan` (like `expectedRef`).
- **Keyword String Conversion**: `token.Keyword` is a named string type - use `string(word.Keyword)` to convert for functions expecting `string`.
- **Ident Type Conversion**: `p.ParseIdentifier()` returns `*ast.Ident`, but `expr.SqlOption.Name` expects `*expr.Ident`. Convert using: `&expr.Ident{SpanVal: key.Span(), Value: key.Value, QuoteStyle: key.QuoteStyle}`.

**Result:** +4 tests passing (473/1207 total, 39% pass rate)

Implemented critical parser fixes for SELECT DISTINCT, FETCH statement, and DELETE statement:

1. **SELECT DISTINCT/ALL Fix** (query.go):
   - Problem: `DISTINCT` and `ALL` keywords were being parsed but discarded (not stored in AST)
   - Root cause: Parser consumed keywords without setting `Select.Distinct` field
   - Fix: Store parsed DISTINCT/ALL in `*query.Distinct` and pass to Select struct
   - **+5 tests passing** (all DISTINCT tests now pass)
   - Pattern: Always store parsed keywords in appropriate AST fields, never discard

2. **FETCH Statement Implementation** (misc.go, expr/ddl.go):
   - Full parser per Rust `parse_fetch_statement` (src/parser/mod.rs:7838)
   - Supports all FETCH directions: NEXT, PRIOR, FIRST, LAST, ABSOLUTE n, RELATIVE n
   - Supports FORWARD/FORWARD ALL, BACKWARD/BACKWARD ALL, ALL, COUNT
   - Proper position parsing: FROM cursor_name, IN cursor_name
   - Optional INTO clause for storing results
   - Fixed `FetchDirection` and `FetchPosition` types to be proper structs with String() methods
   - **New functionality**: FETCH statements now parseable
   - Pattern: When implementing statement parsers, follow Rust's exact token order and error messages

3. **DELETE Statement Improvements** (dml.go):
   - Proper FROM keyword handling per dialect (optional in BigQuery/Oracle/Generic)
   - Multi-table DELETE support: `DELETE t1, t2 FROM t1, t2 WHERE ...`
   - USING clause support: `DELETE FROM t USING t2 WHERE ...`
   - OUTPUT clause support (SQL Server style): `DELETE ... OUTPUT deleted.* INTO @table`
   - Fixed table name extraction from TableWithJoins for single-table DELETE
   - Pattern: Dialect-specific features must check `dialect.Dialect()` or use dialect capability methods

4. **Quoted Identifier Fix** (parser.go):
   - Fixed `ParseIdentifier()` to preserve `QuoteStyle` from TokenWord
   - Double-quoted identifiers like `"table"` now preserve quotes in AST
   - String() output properly renders quoted identifiers
   - Pattern: When creating Ident from TokenWord, always copy QuoteStyle field

**Key Pattern Documentation:**
- **AST Field Preservation**: When parsing tokens that modify the statement (like DISTINCT), always store them in the AST. Never consume without storing.
- **Dialect-Specific Parsing**: Check dialect capabilities before parsing dialect-specific syntax. Use `p.dialect.Dialect()` for dialect name checks.
- **Token to AST Conversion**: When converting TokenWord to ast.Ident, preserve all metadata including QuoteStyle for proper round-trip serialization.

---

### April 5, 2026 - CREATE TRIGGER and CREATE OPERATOR Implementation

Implemented major CREATE statement parsers following Rust reference:

1. **CREATE TRIGGER** - Full parser per Rust `parse_create_trigger` (src/parser/mod.rs:6066):
   - Complete trigger AST types: TriggerPeriod, TriggerEvent, TriggerReferencing, TriggerExecBody
   - Support for: `BEFORE`/`AFTER`/`INSTEAD OF` periods, `INSERT`/`UPDATE`/`DELETE`/`TRUNCATE` events
   - Multiple events with OR separator: `INSERT OR UPDATE OR DELETE`
   - UPDATE OF column list support
   - REFERENCING clause with OLD TABLE/NEW TABLE
   - FOR [EACH] ROW/STATEMENT clause
   - WHEN condition clause
   - EXECUTE FUNCTION/PROCEDURE body
   - **+10 tests passing** (trigger tests)

2. **CREATE OPERATOR** - PostgreSQL operator parser per Rust `parse_create_operator` (src/parser/mod.rs:6993):
   - Complete operator AST types: OperatorOption, OperatorArgTypes, OperatorClassItem
   - Support for: FUNCTION/PROCEDURE parameters, LEFTARG/RIGHTARG data types
   - Operator options: HASHES, MERGES, COMMUTATOR, NEGATOR, RESTRICT, JOIN
   - CREATE OPERATOR FAMILY with USING clause
   - CREATE OPERATOR CLASS with OPERATOR/FUNCTION/STORAGE items
   - **+5 tests passing** (operator tests)

**Pattern Documentation:** When implementing CREATE statement parsers:
1. Reference Rust implementation directly - follow exact parsing order
2. Use `p.ParseDataType()` for data type parsing
3. Use type assertions for token checking: `if _, ok := p.PeekToken().Token.(token.TokenLParen); ok`
4. Always include empty parentheses in FunctionDesc.String() for function calls

---

## Recent Progress (Previous)

### April 5, 2026 - UPDATE FROM, CEIL/FLOOR Fix, Array Subscript Parser

Implemented major missing parser chunks following Rust reference:

1. **UPDATE FROM Clause** - Added full support per Rust `parse_update` (src/parser/mod.rs:17715):
   - PostgreSQL style: `UPDATE t1 SET ... FROM t2 WHERE ...`
   - Snowflake/MSSQL style: `UPDATE FROM t1 SET ... WHERE ...`
   - Support for multiple tables in FROM clause (comma-separated)
   - Updated `Update` AST struct to use `UpdateTableFromKind` (BeforeSet/AfterSet)
   - **+2 tests passing** (UPDATE FROM tests)

2. **CEIL/FLOOR String() Fix** - Fixed incorrect output format:
   - Problem: `CEIL(1.5)` was outputting as `CEIL(1.5 TO )` 
   - Root cause: Parser was setting `DateTimeField = &""` (empty string pointer) instead of leaving it nil
   - Fix: Removed the empty string assignment in `parseCeilFloorExpr` when no TO clause present
   - **+2 tests passing** (CEIL/FLOOR tests)

3. **Array Subscript Infix Parser** - Added missing `[` token handler:
   - Added `TokenLBracket` case to `parseInfix` in infix.go
   - Connected to existing `parseArraySubscript` function in postfix.go
   - Enables parsing of expressions like `arr[1]` or `matrix[i][j]`
   - **+1 test passing** (Array subscript test)

**Implementation Pattern:** When adding infix operators, always:
1. Add precedence case in `getPrecedence` (core.go)
2. Add handler case in `parseInfix` (infix.go) 
3. Ensure the handler consumes the token and returns proper expression type

---

### April 5, 2026 - MERGE OUTPUT/RETURNING and TRUNCATE Implementation

Implemented two major missing parser chunks:

1. **MERGE OUTPUT/RETURNING Clause** - Added support for:
   - SQL Server OUTPUT clause: `MERGE ... OUTPUT inserted.* INTO log_table`
   - PostgreSQL RETURNING clause: `MERGE ... RETURNING merge_action(), *`
   - Fixed qualified wildcard parsing (e.g., `inserted.*`, `w.*`) in projection parser
   - Added `parseOutputClauseInternal()` and `parseSelectIntoInternal()` functions
   - **+8 tests passing** (MERGE OUTPUT/RETURNING tests)

2. **TRUNCATE Statement** - Implemented full parser:
   - TRUNCATE [TABLE] name [, ...] syntax
   - IF EXISTS modifier
   - PARTITION clause support
   - ON CLUSTER (ClickHouse) support
   - **+2 tests passing** (TRUNCATE tests)

**Key Bug Fix:** The qualified wildcard parsing was incorrectly using `ParseObjectName()` which consumed the dot and tried to parse more identifiers. Fixed by parsing just a single identifier for the table name in qualified wildcards.

---

## Major Missing Parser Chunks (Priority Order)

Based on test failures analysis, the following major parser chunks need implementation:

1. **SET Statement Variants** (~10+ test failures)
   - SET TRANSACTION READ ONLY/READ WRITE/ISOLATION LEVEL
   - SET (a, b, c) = (1, 2, 3) parenthesized assignment syntax
   - SET TIME ZONE TO 'value' syntax
   - SET SESSION AUTHORIZATION
   - SET NAMES (MySQL)

2. **SHOW Statement Extensions** (~40+ test failures)
   - SHOW DATABASES, SHOW SCHEMAS
   - SHOW VIEWS, SHOW MATERIALIZED VIEWS
   - SHOW FUNCTIONS
   - TERSE, HISTORY, EXTERNAL modifiers
   - Filter position variations (LIKE/WHERE before IN/FROM)

3. **CREATE Statement Extensions** (~30+ "not yet implemented" errors)
   - CREATE MATERIALIZED VIEW
   - CREATE PROCEDURE
   - CREATE USER
   - CREATE ASSERT
   - CREATE TRIGGER, CREATE FUNCTION, CREATE DATABASE
   - CREATE POLICY, CREATE CONNECTOR, CREATE OPERATOR

4. **UPDATE FROM Clause** (~5+ test failures)
   - UPDATE ... SET ... FROM ... WHERE syntax (MSSQL/PostgreSQL)

5. **FETCH Statement** (~3+ test failures)
   - FETCH [FIRST|NEXT] n [ROW|ROWS] [ONLY|WITH TIES]

---

## Recent Progress (Concise)

### April 5, 2026 - CREATE DATABASE and CREATE FUNCTION Implementation

Implemented two major CREATE statement parsers:

1. **CREATE DATABASE** - Full parser per Rust reference (src/parser/mod.rs:5455):
   - `CREATE DATABASE [IF NOT EXISTS] name` syntax
   - MySQL-style `CHARACTER SET` and `COLLATE` options (with optional `DEFAULT`)
   - Hive/Databricks `LOCATION` and `MANAGEDLOCATION` clauses
   - Snowflake `CLONE` clause
   - **+8 CREATE DATABASE tests now passing**

2. **CREATE FUNCTION** - Basic PostgreSQL/generic parser per Rust reference (src/parser/mod.rs:5553):
   - `CREATE [OR REPLACE] FUNCTION name(args) RETURNS type ...` syntax
   - Function argument modes: `IN`, `OUT`, `INOUT`
   - Named parameters: `param_name TYPE` pattern
   - Function attributes: `LANGUAGE`, `AS`, `IMMUTABLE`, `STABLE`, `VOLATILE`
   - Null handling: `CALLED ON NULL INPUT`, `RETURNS NULL ON NULL INPUT`, `STRICT`
   - Parallel modes: `PARALLEL UNSAFE/RESTRICTED/SAFE`
   - Security modes: `SECURITY DEFINER/INVOKER`
   - SET parameters: `SET param = value` or `SET param FROM CURRENT`
   - **+6 CREATE FUNCTION tests with basic support**
   - **Remaining issues:** Dollar-quoted strings (`$$...$$`), data type normalization (INT vs INTEGER), RETURN vs AS RETURN syntax

3. **Fixed CREATE DATABASE String() output** to include `DEFAULT CHARACTER SET` and `DEFAULT COLLATE` for MySQL compatibility.

4. **Added Function AST Type Constants**:
   - `FunctionBehavior`: Immutable, Stable, Volatile
   - `FunctionCalledOnNull`: CalledOnNullInput, ReturnsNullOnNullInput, Strict
   - `FunctionParallel`: Unsafe, Restricted, Safe
   - `FunctionSecurity`: Definer, Invoker
   - `ArgMode`: In, Out, InOut

### April 5, 2026 - SHOW Statement Extensions and CREATE TABLE AS/LIKE

Implemented major missing parser chunks to bring maximum test coverage:

1. **SHOW Statement Extensions** - Added support for:
   - SHOW DATABASES, SHOW SCHEMAS with TERSE/HISTORY modifiers
   - SHOW VIEWS, SHOW MATERIALIZED VIEWS
   - SHOW FUNCTIONS, SHOW OBJECTS
   - EXTERNAL modifier for SHOW TABLES
   - Bare string literal suffix filters (Snowflake-style: `SHOW TABLES IN db1 'abc'`)
   - Added `SuffixString` field to `ShowStatementFilter` AST type

2. **CREATE TABLE AS/LIKE** - Added support for:
   - CREATE TABLE ... AS (query) - with proper query extraction from SelectStatement/QueryStatement
   - CREATE TABLE ... LIKE (table) - with CreateTableLikeKind
   - Parsing of LOCAL/GLOBAL/TRANSIENT modifiers before CREATE TABLE
   - Updated `ParseCreate` to handle modifier parsing order per Rust reference

**Result:** SHOW statements now working for Snowflake, MySQL, and Generic dialects. CREATE TABLE AS basic functionality operational. Common tests pass rate improved.

### April 4, 2026 - Analysis of Missing Parser Chunks

Implemented CREATE ROLE and DROP ROLE parsers, added DENY statement support, and fixed critical tokenizer case preservation bugs. The tokenizer was hardcoding uppercase prefixes in `tokenizeQuoteDelimitedLiteral`, `tokenizeEscapedStringLiteral`, and `tokenizeHexStringLiteral`, causing identifiers like "quaternion" to become "Quaternion" in GenericDialect. Added Pattern L documenting this widespread issue. **22 more tests passing** (356/858 total, 41% pass rate).

### April 4, 2026 - Tokenizer Case Preservation Fix

Fixed critical bug in `tokenizeUnicodeStringLiteral()` that was hardcoding uppercase "U" instead of preserving original case. This caused identifiers starting with 'u' (like `uk_cities`) to be incorrectly tokenized as "Uk_cities". Added Pattern K to "Typical Code Editing Errors" section documenting this case. **10 more tests passing** (334/858 total, 39% pass rate).

### April 4, 2026 - Named Argument Test Fix

Fixed `TestParseNamedArgumentFunctionWithEqOperator` to use `NewTestedDialectsWithFilter()` with predicate for `SupportsNamedFnArgsWithEqOperator()`, ensuring only dialects that support `=` for named arguments are tested. This aligns with the Rust reference test using `all_dialects_where()`.

### April 4, 2026 - Array Subscripts and CTEs

Implemented array subscript parsing, CTE in CREATE VIEW, UNNEST table factors, and VALUES as table factors. **6 more tests passing** (40% common tests).

### April 4, 2026 - Transaction Statements

Implemented COMMIT, ROLLBACK, SAVEPOINT, RELEASE, START TRANSACTION, LISTEN/NOTIFY, SQLite PRAGMA, and transaction modes. **52 more tests passing** (308 total).

### April 4, 2026 - INTERVAL, Typed Strings, TABLE Function

Fixed INTERVAL in window frames for dialects requiring qualifiers, added TIMESTAMPTZ/JSON/BIGNUMERIC support, implemented TABLE function parsing. **5 more tests passing**.

### April 4, 2026 - MERGE Statement

Implemented MERGE INTO statement with UPDATE SET, DELETE, INSERT actions and compound identifiers. **17 more tests passing** (251 total, 31%).

### April 4, 2026 - COPY and Window Functions

Implemented PostgreSQL COPY FROM/TO, fixed OVER clause for PARTITION BY/ORDER BY, added QUALIFY clause. **4 more tests passing**.

### April 4, 2026 - GRANT/REVOKE and LOCK TABLES

Implemented GRANT/REVOKE statements and MySQL LOCK TABLES with implicit AS. **3 more MySQL tests passing** (42%).

### April 4, 2026 - Named Arguments and CTE

Implemented PostgreSQL named arguments (`=>` syntax) and basic CTE (WITH clause) parsing. **2 more tests passing**.

---

## Project Structure

```
sqlparser-go/
├── token/                      # 800+ SQL keywords, lexer, position/span
├── errors/                     # Error types
├── ast/                        # Abstract Syntax Tree
│   ├── statement/             # Statement types (consolidated)
│   ├── expr/                  # Expression types
│   ├── datatype/              # DataType variants
│   ├── query/                 # Query-related types
│   ├── node.go                # Base Node types
│   ├── expr.go, expr_all.go   # Expression interfaces and all expressions
│   ├── statement.go, statement_all.go # Statement interfaces and types
│   ├── query.go, query_all.go # Query interfaces and types
│   └── types_all.go, operators_all.go # Consolidated type definitions
├── parser/                     # Parser (~10,000 lines, split by function)
│   ├── core.go, parser.go     # Core parser types and main entry points
│   ├── helpers.go, utils.go   # Parser utilities
│   ├── prefix.go, infix.go, postfix.go # Expression parsing
│   ├── create.go, alter.go, drop.go, truncate.go # DDL statements
│   ├── dml.go                 # INSERT, UPDATE, DELETE
│   ├── query.go               # SELECT, CTE parsing
│   ├── merge.go, copy.go, show.go, describe.go, misc.go # Other statements
│   └── transaction.go, prepared.go, special.go, groupings.go, options.go, state.go
├── parseriface/               # Parser interface definitions (resolves circular deps)
├── dialects/                  # 14 SQL dialects
│   ├── dialect.go, capabilities.go # Core dialect interfaces
│   ├── generic, postgresql, mysql, sqlite, bigquery
│   ├── snowflake, duckdb, clickhouse, hive, mssql
│   └── redshift, databricks, oracle, ansi
├── tests/                     # Test suites (TPC-H 100% ✅)
│   ├── ddl/                   # CREATE, ALTER, DROP, TRUNCATE tests
│   ├── query/                 # JOIN, CTE, set operations tests
│   └── snowflake/, mysql/, postgres/ # Dialect-specific tests
├── fuzz/                      # Fuzz testing
└── examples/                  # Usage examples
```

---

## Test Porting Status

| Test Suite       | Tests | Status                    |
| ---------------- | ----- | ------------------------- |
| **tests**        | ~50   | 🔄 Mixed results          |
| **tests/ddl**    | ~300  | 🔄 ~40% passing           |
| **tests/dml**    | ~150  | 🔄 ~53% passing           |
| **tests/query**  | ~350  | 🔄 ~43% passing           |
| **tests/mysql**  | ~125  | 🔄 ~48% passing           |
| **tests/postgres**| ~157 | 🔄 ~29% passing           |
| **tests/snowflake**| ~97  | 🔄 ~19% passing           |
| **tests/regression**| ~120 | 🔄 Mixed results          |
| **TOTAL**        | 1199  | **39% passing**           |

**Completed:** Tokenizer (29/29), TPC-H (44/44, 100%), AST types (131 statements, 69 expressions, 117 data types), 14 dialects, Fuzz testing.

---

## AST Interface Design

```go
// Sealed interface pattern

type Node interface {
    node() // Unexported prevents external implementation
}

type Statement interface {
    Node
    statementNode()
    String() string  // SQL regeneration
}

type Expr interface {
    Node
    exprNode()
    String() string
}

// Type assertion pattern (replaces Rust match)
func processStatement(stmt Statement) error {
    switch s := stmt.(type) {
    case *SelectStmt:
        return handleSelect(s)
    case *InsertStmt:
        return handleInsert(s)
    default:
        return fmt.Errorf("unknown statement: %T", stmt)
    }
}
```

---

## Running Tests

```bash
cd /Users/san/Fun/sqlparser-rs/go  # Must run from go/ directory

go test ./token/... -v               # Tokenizer tests (29 passing)
go test ./tests/... -v              # All test suites
go test ./tests/ddl/... -v          # DDL tests (CREATE, ALTER, DROP)
go test ./tests/query/... -v         # Query tests (JOIN, CTE)
go test ./tests/mysql/... -v         # MySQL dialect
go test ./tests/postgres/... -v      # PostgreSQL dialect
go build ./...                      # Build everything
```

---

### April 5, 2026 - PIVOT and UNPIVOT Table Factor Implementation

Implemented major missing parser chunks for PIVOT and UNPIVOT table operations following Rust reference (src/parser/mod.rs:16589-16678):

1. **PIVOT Table Factor Parser** (parser/query.go):
   - Implemented `parsePivotTableFactor()` for ClickHouse/Oracle-style PIVOT
   - Supports: `table PIVOT (agg_func FOR col IN (values))`
   - Handles multiple aggregate functions with optional aliases
   - Supports value columns (single or parenthesized list)
   - Supports three value source types:
     - Static value list: `IN ('JAN', 'FEB')`
     - Subquery: `IN (SELECT col FROM table)`
     - ANY with ORDER BY (Snowflake): `IN (ANY ORDER BY col)`
   - Supports DEFAULT ON NULL clause: `DEFAULT ON NULL (0)`
   - Properly handles table aliases after PIVOT operation

2. **UNPIVOT Table Factor Parser** (parser/query.go):
   - Implemented `parseUnpivotTableFactor()` for standard SQL UNPIVOT
   - Supports: `table UNPIVOT (value FOR name IN (cols))`
   - Handles optional INCLUDE/EXCLUDE NULLS modifier
   - Properly parses value expression and column list with optional aliases

3. **Integration with Table Factor Parsing**:
   - Added `parseTableNameWithPivot()` wrapper to check for PIVOT/UNPIVOT after table names
   - Updated `parseDerivedTableAfterParen()` to handle PIVOT/UNPIVOT after subqueries
   - Updated `isReservedForTableAlias()` to prevent PIVOT/UNPIVOT being parsed as table aliases
   - PIVOT/UNPIVOT operations can be chained: `table UNPIVOT (...) PIVOT (...)`

4. **Helper Functions Added**:
   - `parseCommaSeparatedPivotAggregates()` - parses comma-separated aggregate functions
   - `parsePivotAggregateFunction()` - parses single aggregate with optional alias
   - `parsePivotValueSource()` - handles ANY, subquery, or value list
   - `parseCommaSeparatedExprWithAlias()` - parses expressions with optional AS aliases
   - `parseParenthesizedExprWithAliasList()` - parses parenthesized expression list

**Key Pattern Documentation:**
- **Pattern AC: PIVOT Aggregate Parsing** - Use ExpressionParser to parse aggregate functions like SUM(amount) rather than manual function parsing. The expression parser already handles function calls correctly.
- **Pattern AD: Reserved Keywords for Table Aliases** - When PIVOT/UNPIVOT follows a table name, they must NOT be treated as table aliases. Add them to `isReservedForTableAlias()` to prevent incorrect parsing.
- **Pattern AE: Chained Table Operations** - PIVOT and UNPIVOT can chain: `(SELECT ...) PIVOT (...)`. The loop in `parseTableNameWithPivot()` and `parseDerivedTableAfterParen()` handles multiple operations.

**Result:** +1 test passing (TestParseUnpivotTable). PIVOT partially working - core parsing complete but Snowflake-specific features (subquery values, complex aggregates) need refinement.

---

## April 5, 2026 - SET Statement Enhancements and System Variable Support

Implemented major missing SET statement variants and fixed critical system variable tokenization issues:

1. **SET TRANSACTION with SNAPSHOT** (parser/misc.go, ast/statement/misc.go):
   - Added `SetTransaction` AST type with support for transaction modes and SNAPSHOT
   - Implemented `parseTransactionModes()` supporting ISOLATION LEVEL, READ ONLY/WRITE, DEFERRABLE
   - Syntax: `SET [SESSION|LOCAL] TRANSACTION [modes] [SNAPSHOT value]`
   - Also supports: `SET SESSION CHARACTERISTICS AS TRANSACTION [modes]`

2. **SET SESSION AUTHORIZATION** (parser/misc.go, ast/statement/misc.go):
   - Added `SetSessionAuthorization` AST type
   - Syntax: `SET {SESSION|LOCAL} AUTHORIZATION { username | DEFAULT }`
   - Updated `ParseIdentifier()` to accept quoted strings as identifiers (Pattern V)

3. **SET ROLE** (parser/misc.go, ast/statement/misc.go):
   - Added `SetRole` AST type
   - Syntax: `SET [SESSION|LOCAL] ROLE { rolename | NONE }`

4. **USE Statement** (parser/misc.go, ast/statement/dcl.go):
   - Extended `Use` AST type with comprehensive support for:
     - `USE DATABASE name`, `USE SCHEMA name`, `USE CATALOG name` (Databricks)
     - `USE WAREHOUSE name`, `USE ROLE name`, `USE SECONDARY ROLES {ALL|NONE|role,...}` (Snowflake)
     - `USE DEFAULT` (Hive)
     - `USE object_name` (Generic)
   - Added `SecondaryRoles` helper type

5. **@@ System Variable Tokenization Fix** (token/lexer.go, parser/prefix.go, ast/expr/basic.go):
   - Fixed critical bug where `@@sql_mode` was being tokenized incorrectly as `@@sl_mode` (missing characters)
   - Root cause: `tokenizeAtSign()` was calling `tokenizeIdentifierOrKeyword()` with wrong character consumption
   - Solution: Return `TokenAtAt` as separate token, then let parser combine with identifier
   - Added `SystemVariable` expression type to properly represent `@@var` and `@@global.var`
   - String() output correctly shows `@@var` instead of `@.@.var`

**Key Pattern Documentation:**
- **Pattern V: Quoted Strings as Identifiers** - `ParseIdentifier()` must accept `TokenSingleQuotedString` and `TokenDoubleQuotedString` as valid identifiers (not just `TokenWord`), following Rust `parse_identifier()` in `src/parser/mod.rs:12926`. This is required for `SET SESSION AUTHORIZATION 'username'`.
- **Pattern BF: Tokenizer Character Consumption** - When using `tokenizeIdentifierOrKeyword()`, be careful about character consumption. The function calls `state.Next()` twice to consume characters from the prefix. If the prefix is longer than 2 characters, the extra characters will be consumed from the input stream, causing character loss. Either use a 2-char prefix or handle tokenization differently.
- **Pattern BG: System Variable Tokenization** - For MySQL-style `@@var`, tokenizer should produce two separate tokens: `TokenAtAt` and `TokenWord`. The parser's prefix handler for `TokenAtAt` then combines them into a `SystemVariable` expression. Don't try to tokenize `@@var` as a single identifier - it causes character loss issues.

**Result:** +8 tests passing (TestParseSetVariables, TestParseSetSessionAuthorization, and 6 more). 448 failing subtests remaining (down from 456).

---

**Version:** 1.0  
**Last Updated:** April 5, 2026  
**Status:** TPC-H fixture issue, DDL ~27%, DML ~55%, Query ~67%, MySQL ~48%, PostgreSQL ~27%, Snowflake ~18%, **Total ~496/1207 (~41%)**

**Line Counts:**
- Rust Source: 67,345 lines  
- Go Source: 69,904 lines (104% of Rust)  
- Go Tests: 14,300 lines (29%)  
- Rust Tests: 49,886 lines
