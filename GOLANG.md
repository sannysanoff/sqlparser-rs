---

**Line Counts (Updated April 8, 2026):**

- Rust Source: 67,345 lines (parser + dialects + AST)
- Go Source: 76,336 lines (113% of Rust - AST types and interfaces)
- Rust Tests: 49,886 lines  
- Go Tests: 14,149 lines (28.3% of Rust test coverage)
- **Current Test Pass Rate: ~58.5%** (679 passing out of 1,160 total tests)

**Recent Progress:**
- Fixed ALTER TABLE SET options parsing (+1 test)
- Fixed ALTER TABLE RENAME AS vs TO distinction (+1 test) 
- Fixed ALTER COLUMN ... USING clause (+1 test)
- Fixed comma detection in parseSqlOptions (used by many tests)
- **Total: +32 tests passing** (from ~442 to 474)

---

### April 8, 2026 - Table Constraint Types Implementation

Implemented comprehensive table constraint types to fix DDL constraint parsing:

1. **AST Constraint Types** (ast/expr/ddl.go):
   - `PrimaryKeyConstraint` - PRIMARY KEY with optional index name, index type, columns, and characteristics
   - `UniqueConstraint` - UNIQUE with NULLS DISTINCT/NOT DISTINCT support
   - `ForeignKeyConstraint` - FOREIGN KEY with REFERENCES, ON DELETE/UPDATE actions, MATCH kinds
   - `CheckConstraint` - CHECK with optional ENFORCED/NOT ENFORCED (MySQL)
   - `IndexConstraint` - MySQL INDEX/KEY constraints
   - `FullTextOrSpatialConstraint` - MySQL FULLTEXT/SPATIAL constraints
   - Supporting types: `ConstraintCharacteristics`, `ConstraintReferenceMatchKind`, `NullsDistinctOption`

2. **Parser Updates** (parser/ddl.go):
   - Updated `parseTableConstraint()` to populate constraint-specific structs instead of discarding parsed data
   - Updated `parseConstraintCharacteristics()` to return `*ConstraintCharacteristics` instead of discarding
   - Fixed FOREIGN KEY String() output to match expected format: `REFERENCES table(col)` not `REFERENCES table (col)`

3. **Key Pattern Documentation:**
   - **Pattern CT: Table Constraint Implementation** - When implementing table constraints:
     1. Create specific constraint type structs with all relevant fields
     2. Store parsed data in the constraint struct, never discard with `_ = parsedValue`
     3. Update the TableConstraint.Constraint field with the specific constraint type
     4. Implement proper String() method that matches SQL canonical format
     5. For FOREIGN KEY, concatenate table name and column list without space: `table(col)` not `table (col)`

**Result:** +6 tests passing (TestParseAlterTableConstraints and related tests)

(End of file - total 3398 lines)