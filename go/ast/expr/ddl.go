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

// Package expr provides SQL expression types for the sqlparser AST.
// This file contains DDL-related expression types.
package expr

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/query"
	"github.com/user/sqlparser/token"
)

// ============================================================================
// ColumnDef - Column definition for CREATE TABLE
// ============================================================================

// ColumnDef represents a column definition in a CREATE TABLE statement.
type ColumnDef struct {
	Name     *ast.Ident
	DataType interface{} // datatype.DataType - using interface{} to avoid import cycle
	Options  []*ColumnOptionDef
	SpanVal  token.Span
}

func (c *ColumnDef) exprNode() {}
func (c *ColumnDef) expr()     {}
func (c *ColumnDef) IsExpr()   {}

// Span returns the source span for this expression.
func (c *ColumnDef) Span() token.Span {
	return c.SpanVal
}

// String returns the SQL representation.
func (c *ColumnDef) String() string {
	var parts []string
	if c.Name != nil {
		parts = append(parts, c.Name.String())
	}
	if c.DataType != nil {
		if s, ok := c.DataType.(fmt.Stringer); ok {
			parts = append(parts, s.String())
		}
	}
	for _, opt := range c.Options {
		parts = append(parts, opt.String())
	}
	return strings.Join(parts, " ")
}

// ============================================================================
// TableConstraint - Table constraint for CREATE TABLE
// ============================================================================

// TableConstraint represents a constraint on a table.
type TableConstraint struct {
	Name       *ast.Ident
	Constraint interface{} // TODO: Define specific constraint types
	SpanVal    token.Span
}

func (t *TableConstraint) exprNode() {}
func (t *TableConstraint) expr()     {}
func (t *TableConstraint) IsExpr()   {}

// Span returns the source span for this expression.
func (t *TableConstraint) Span() token.Span {
	return t.SpanVal
}

// String returns the SQL representation.
func (t *TableConstraint) String() string {
	var parts []string
	if t.Name != nil {
		parts = append(parts, fmt.Sprintf("CONSTRAINT %s", t.Name.String()))
	}
	if t.Constraint != nil {
		switch c := t.Constraint.(type) {
		case *PrimaryKeyConstraint:
			parts = append(parts, c.String())
		case *UniqueConstraint:
			parts = append(parts, c.String())
		case *ForeignKeyConstraint:
			parts = append(parts, c.String())
		case *CheckConstraint:
			parts = append(parts, c.String())
		case *IndexConstraint:
			parts = append(parts, c.String())
		case *FullTextOrSpatialConstraint:
			parts = append(parts, c.String())
		case fmt.Stringer:
			parts = append(parts, c.String())
		}
	}
	return strings.Join(parts, " ")
}

// ============================================================================
// Constraint Types
// ============================================================================

// PrimaryKeyConstraint represents a PRIMARY KEY constraint.
// [CONSTRAINT [name]] PRIMARY KEY [index_name] [USING index_type] (columns) [index_options] [characteristics]
type PrimaryKeyConstraint struct {
	IndexName       *ast.Ident
	IndexType       *IndexType
	Columns         []*IndexColumn
	IndexOptions    []*IndexOption
	Characteristics *ConstraintCharacteristics
}

func (p *PrimaryKeyConstraint) String() string {
	var parts []string
	parts = append(parts, "PRIMARY KEY")
	if p.IndexName != nil {
		parts = append(parts, p.IndexName.String())
	}
	if p.IndexType != nil {
		parts = append(parts, fmt.Sprintf("USING %s", p.IndexType.String()))
	}
	if len(p.Columns) > 0 {
		colStrs := make([]string, len(p.Columns))
		for i, col := range p.Columns {
			colStrs[i] = col.String()
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(colStrs, ", ")))
	}
	for _, opt := range p.IndexOptions {
		parts = append(parts, opt.String())
	}
	if p.Characteristics != nil {
		parts = append(parts, p.Characteristics.String())
	}
	return strings.Join(parts, " ")
}

// UniqueConstraint represents a UNIQUE constraint.
// [CONSTRAINT [name]] UNIQUE [NULLS [NOT] DISTINCT] [INDEX|KEY] [index_name] [USING index_type] (columns) [index_options] [characteristics]
type UniqueConstraint struct {
	NullsDistinct   NullsDistinctOption
	HasIndexKeyword bool // true if INDEX/KEY keyword was explicitly specified (MySQL style)
	IndexName       *ast.Ident
	IndexType       *IndexType
	Columns         []*IndexColumn
	IndexOptions    []*IndexOption
	Characteristics *ConstraintCharacteristics
}

func (u *UniqueConstraint) String() string {
	var parts []string
	parts = append(parts, "UNIQUE")
	if u.NullsDistinct != NullsDistinctOptionNone {
		parts = append(parts, u.NullsDistinct.String())
	}
	// Include INDEX keyword if explicitly specified (MySQL style)
	if u.HasIndexKeyword {
		parts = append(parts, "INDEX")
	}
	if u.IndexName != nil {
		parts = append(parts, u.IndexName.String())
	}
	if u.IndexType != nil {
		parts = append(parts, fmt.Sprintf("USING %s", u.IndexType.String()))
	}
	if len(u.Columns) > 0 {
		colStrs := make([]string, len(u.Columns))
		for i, col := range u.Columns {
			colStrs[i] = col.String()
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(colStrs, ", ")))
	}
	for _, opt := range u.IndexOptions {
		parts = append(parts, opt.String())
	}
	if u.Characteristics != nil {
		parts = append(parts, u.Characteristics.String())
	}
	return strings.Join(parts, " ")
}

// ForeignKeyConstraint represents a FOREIGN KEY constraint.
// [CONSTRAINT [name]] FOREIGN KEY (columns) REFERENCES table [(cols)] [MATCH kind] [ON DELETE action] [ON UPDATE action] [characteristics]
type ForeignKeyConstraint struct {
	IndexName       *ast.Ident
	Columns         []*ast.Ident
	ForeignTable    *ast.ObjectName
	ReferredColumns []*ast.Ident
	MatchKind       *ConstraintReferenceMatchKind
	OnDelete        ReferentialAction
	OnUpdate        ReferentialAction
	Characteristics *ConstraintCharacteristics
}

func (f *ForeignKeyConstraint) String() string {
	var parts []string
	parts = append(parts, "FOREIGN KEY")
	if f.IndexName != nil {
		parts = append(parts, f.IndexName.String())
	}
	if len(f.Columns) > 0 {
		colStrs := make([]string, len(f.Columns))
		for i, col := range f.Columns {
			colStrs[i] = col.String()
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(colStrs, ", ")))
	}
	if f.ForeignTable != nil {
		// Build REFERENCES clause: REFERENCES table_name[(cols)] without space before (cols)
		refClause := "REFERENCES " + f.ForeignTable.String()
		if len(f.ReferredColumns) > 0 {
			colStrs := make([]string, len(f.ReferredColumns))
			for i, col := range f.ReferredColumns {
				colStrs[i] = col.String()
			}
			refClause += fmt.Sprintf("(%s)", strings.Join(colStrs, ", "))
		}
		parts = append(parts, refClause)
	}
	if f.MatchKind != nil {
		parts = append(parts, f.MatchKind.String())
	}
	if f.OnDelete != ReferentialActionNone {
		parts = append(parts, "ON DELETE", f.OnDelete.String())
	}
	if f.OnUpdate != ReferentialActionNone {
		parts = append(parts, "ON UPDATE", f.OnUpdate.String())
	}
	if f.Characteristics != nil {
		parts = append(parts, f.Characteristics.String())
	}
	return strings.Join(parts, " ")
}

// CheckConstraint represents a CHECK constraint.
// [CONSTRAINT [name]] CHECK (expr) [[NOT] ENFORCED]
type CheckConstraint struct {
	Expr     Expr
	Enforced *bool // nil = not specified, true = ENFORCED, false = NOT ENFORCED
}

func (c *CheckConstraint) String() string {
	var parts []string
	parts = append(parts, "CHECK")
	if c.Expr != nil {
		parts = append(parts, fmt.Sprintf("(%s)", c.Expr.String()))
	} else {
		parts = append(parts, "()")
	}
	if c.Enforced != nil {
		if *c.Enforced {
			parts = append(parts, "ENFORCED")
		} else {
			parts = append(parts, "NOT ENFORCED")
		}
	}
	return strings.Join(parts, " ")
}

// IndexConstraint represents an INDEX constraint (MySQL-specific).
// {INDEX | KEY} [index_name] [USING index_type] (columns) [index_options]
type IndexConstraint struct {
	DisplayAsKey bool // true = KEY, false = INDEX
	Name         *ast.Ident
	IndexType    *IndexType
	Columns      []*IndexColumn
	IndexOptions []*IndexOption
}

func (i *IndexConstraint) String() string {
	var parts []string
	if i.DisplayAsKey {
		parts = append(parts, "KEY")
	} else {
		parts = append(parts, "INDEX")
	}
	if i.Name != nil {
		parts = append(parts, i.Name.String())
	}
	if i.IndexType != nil {
		parts = append(parts, fmt.Sprintf("USING %s", i.IndexType.String()))
	}
	if len(i.Columns) > 0 {
		colStrs := make([]string, len(i.Columns))
		for j, col := range i.Columns {
			colStrs[j] = col.String()
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(colStrs, ", ")))
	}
	for _, opt := range i.IndexOptions {
		parts = append(parts, opt.String())
	}
	return strings.Join(parts, " ")
}

// FullTextOrSpatialConstraint represents a FULLTEXT or SPATIAL constraint (MySQL-specific).
// {FULLTEXT | SPATIAL} [INDEX | KEY] [index_name] (columns)
type FullTextOrSpatialConstraint struct {
	Fulltext bool // true = FULLTEXT, false = SPATIAL
	Columns  []*IndexColumn
}

func (f *FullTextOrSpatialConstraint) String() string {
	var parts []string
	if f.Fulltext {
		parts = append(parts, "FULLTEXT")
	} else {
		parts = append(parts, "SPATIAL")
	}
	if len(f.Columns) > 0 {
		colStrs := make([]string, len(f.Columns))
		for i, col := range f.Columns {
			colStrs[i] = col.String()
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(colStrs, ", ")))
	}
	return strings.Join(parts, " ")
}

// NullsDistinctOption represents NULLS DISTINCT option for UNIQUE constraints.
type NullsDistinctOption int

const (
	NullsDistinctOptionNone NullsDistinctOption = iota
	NullsDistinctOptionNullsDistinct
	NullsDistinctOptionNullsNotDistinct
)

func (n NullsDistinctOption) String() string {
	switch n {
	case NullsDistinctOptionNullsDistinct:
		return "NULLS DISTINCT"
	case NullsDistinctOptionNullsNotDistinct:
		return "NULLS NOT DISTINCT"
	default:
		return ""
	}
}

// ConstraintReferenceMatchKind represents MATCH kind for foreign key constraints.
type ConstraintReferenceMatchKind int

const (
	ConstraintReferenceMatchKindNone ConstraintReferenceMatchKind = iota
	ConstraintReferenceMatchKindFull
	ConstraintReferenceMatchKindPartial
	ConstraintReferenceMatchKindSimple
)

func (c ConstraintReferenceMatchKind) String() string {
	switch c {
	case ConstraintReferenceMatchKindFull:
		return "MATCH FULL"
	case ConstraintReferenceMatchKindPartial:
		return "MATCH PARTIAL"
	case ConstraintReferenceMatchKindSimple:
		return "MATCH SIMPLE"
	default:
		return ""
	}
}

// ConstraintCharacteristics represents constraint characteristics (DEFERRABLE, etc.).
type ConstraintCharacteristics struct {
	Deferrable *bool // nil = not specified, true = DEFERRABLE, false = NOT DEFERRABLE
	Initially  *ConstraintInitiallyOption
	NotValid   bool
	Enforced   *bool // For MySQL CHECK constraints
}

func (c *ConstraintCharacteristics) String() string {
	var parts []string
	if c.Deferrable != nil {
		if *c.Deferrable {
			parts = append(parts, "DEFERRABLE")
		} else {
			parts = append(parts, "NOT DEFERRABLE")
		}
	}
	if c.Initially != nil {
		parts = append(parts, c.Initially.String())
	}
	if c.NotValid {
		parts = append(parts, "NOT VALID")
	}
	if c.Enforced != nil {
		if *c.Enforced {
			parts = append(parts, "ENFORCED")
		} else {
			parts = append(parts, "NOT ENFORCED")
		}
	}
	return strings.Join(parts, " ")
}

// ConstraintInitiallyOption represents INITIALLY option.
type ConstraintInitiallyOption int

const (
	ConstraintInitiallyOptionNone ConstraintInitiallyOption = iota
	ConstraintInitiallyOptionDeferred
	ConstraintInitiallyOptionImmediate
)

func (c ConstraintInitiallyOption) String() string {
	switch c {
	case ConstraintInitiallyOptionDeferred:
		return "INITIALLY DEFERRED"
	case ConstraintInitiallyOptionImmediate:
		return "INITIALLY IMMEDIATE"
	default:
		return ""
	}
}

// IndexOption represents an index option.
type IndexOption struct {
	Name  string
	Value Expr
}

func (i *IndexOption) String() string {
	if i.Value != nil {
		// Special handling for USING option: output "USING BTREE" not "USING = BTREE"
		if i.Name == "USING" {
			return fmt.Sprintf("USING %s", i.Value.String())
		}
		// Special handling for COMMENT option: output "COMMENT 'string'" with quotes
		if i.Name == "COMMENT" {
			if strVal, ok := i.Value.(*ValueExpr); ok {
				if str, ok := strVal.Value.(string); ok {
					return fmt.Sprintf("COMMENT '%s'", str)
				}
			}
			return fmt.Sprintf("COMMENT '%s'", i.Value.String())
		}
		return fmt.Sprintf("%s = %s", i.Name, i.Value.String())
	}
	return i.Name
}

// ============================================================================
// FileFormat - File format for external tables
// ============================================================================

// FileFormat represents file format options for external tables.
type FileFormat int

const (
	FileFormatTEXTFILE FileFormat = iota
	FileFormatSEQUENCEFILE
	FileFormatORC
	FileFormatPARQUET
	FileFormatAVRO
	FileFormatRCFILE
	FileFormatJSONFILE
)

func (f FileFormat) String() string {
	switch f {
	case FileFormatTEXTFILE:
		return "TEXTFILE"
	case FileFormatSEQUENCEFILE:
		return "SEQUENCEFILE"
	case FileFormatORC:
		return "ORC"
	case FileFormatPARQUET:
		return "PARQUET"
	case FileFormatAVRO:
		return "AVRO"
	case FileFormatRCFILE:
		return "RCFILE"
	case FileFormatJSONFILE:
		return "JSONFILE"
	default:
		return ""
	}
}

// ============================================================================
// HiveDistributionStyle - Hive distribution options
// ============================================================================

// HiveDistributionStyle represents Hive table distribution style options.
type HiveDistributionStyle struct {
	Type                HiveDistributionType
	PartitionColumns    []*ColumnDef
	SkewColumns         []*ColumnDef
	SkewOnColumns       []*ColumnDef
	StoredAsDirectories bool
}

// HiveDistributionType represents the type of Hive distribution.
type HiveDistributionType int

const (
	HiveDistributionNONE HiveDistributionType = iota
	HiveDistributionPARTITIONED
	HiveDistributionSKEWED
)

func (h *HiveDistributionStyle) exprNode() {}
func (h *HiveDistributionStyle) expr()     {}
func (h *HiveDistributionStyle) IsExpr()   {}

// Span returns the source span for this expression.
func (h *HiveDistributionStyle) Span() token.Span {
	return token.Span{}
}

// String returns the SQL representation.
func (h *HiveDistributionStyle) String() string {
	return "DISTRIBUTION"
}

// ============================================================================
// HiveFormat - Hive format specification
// ============================================================================

// HiveFormat represents Hive table format and storage-related options.
type HiveFormat struct {
	RowFormat       *HiveRowFormat
	SerdeProperties []*SqlOption
	Storage         *HiveIOFormat
	Location        *string
}

// HiveRowFormat represents row format specification for Hive tables.
type HiveRowFormat struct {
	SerdeClass *string
	Delimited  bool
	Delimiters []*HiveRowDelimiter
}

// HiveRowDelimiter represents a single row delimiter specification.
type HiveRowDelimiter struct {
	Delimiter string
	Value     string
}

// HiveIOFormat represents input/output storage format details.
type HiveIOFormat struct {
	InputFormat  string
	OutputFormat string
}

func (h *HiveFormat) exprNode() {}
func (h *HiveFormat) expr()     {}
func (h *HiveFormat) IsExpr()   {}

// Span returns the source span for this expression.
func (h *HiveFormat) Span() token.Span {
	return token.Span{}
}

// String returns the SQL representation.
func (h *HiveFormat) String() string {
	return "FORMAT"
}

// ============================================================================
// CreateTableOptions - Options for CREATE TABLE
// ============================================================================

// CreateTableOptions represents options within a CREATE TABLE statement.
type CreateTableOptions struct {
	Type    CreateTableOptionsType
	Options []*SqlOption
}

// CreateTableOptionsType represents the type of CREATE TABLE options.
type CreateTableOptionsType int

const (
	CreateTableOptionsNone CreateTableOptionsType = iota
	CreateTableOptionsWith
	CreateTableOptionsOptions
	CreateTableOptionsPlain
	CreateTableOptionsTableProperties
)

func (c *CreateTableOptions) exprNode() {}
func (c *CreateTableOptions) expr()     {}
func (c *CreateTableOptions) IsExpr()   {}

// Span returns the source span for this expression.
func (c *CreateTableOptions) Span() token.Span {
	return token.Span{}
}

// String returns the SQL representation.
func (c *CreateTableOptions) String() string {
	switch c.Type {
	case CreateTableOptionsWith:
		return fmt.Sprintf("WITH (%s)", joinSqlOptions(c.Options))
	case CreateTableOptionsOptions:
		return fmt.Sprintf("OPTIONS(%s)", joinSqlOptions(c.Options))
	case CreateTableOptionsTableProperties:
		return fmt.Sprintf("TBLPROPERTIES (%s)", joinSqlOptions(c.Options))
	case CreateTableOptionsPlain:
		return joinSqlOptions(c.Options)
	default:
		return ""
	}
}

func joinSqlOptions(opts []*SqlOption) string {
	var parts []string
	for _, opt := range opts {
		parts = append(parts, opt.String())
	}
	return strings.Join(parts, ", ")
}

// ============================================================================
// CreateTableLikeKind - LIKE clause options
// ============================================================================

// CreateTableLikeKind represents the LIKE clause of a CREATE TABLE statement.
type CreateTableLikeKind struct {
	Kind     CreateTableLikeType
	Name     *ast.ObjectName
	Defaults *CreateTableLikeDefaults
}

// CreateTableLikeType represents the type of LIKE clause.
type CreateTableLikeType int

const (
	CreateTableLikeParenthesized CreateTableLikeType = iota
	CreateTableLikePlain
)

// CreateTableLikeDefaults represents whether defaults are included.
type CreateTableLikeDefaults int

const (
	CreateTableLikeDefaultsNone CreateTableLikeDefaults = iota
	CreateTableLikeDefaultsIncluding
	CreateTableLikeDefaultsExcluding
)

func (c *CreateTableLikeKind) exprNode() {}
func (c *CreateTableLikeKind) expr()     {}
func (c *CreateTableLikeKind) IsExpr()   {}

// Span returns the source span for this expression.
func (c *CreateTableLikeKind) Span() token.Span {
	return token.Span{}
}

// String returns the SQL representation.
func (c *CreateTableLikeKind) String() string {
	var sb strings.Builder

	// For parenthesized format, wrap in parentheses
	if c.Kind == CreateTableLikeParenthesized {
		sb.WriteString("(")
	}

	sb.WriteString("LIKE")
	if c.Name != nil {
		sb.WriteString(" ")
		sb.WriteString(c.Name.String())
	}

	// Add INCLUDING/EXCLUDING DEFAULTS if present
	if c.Defaults != nil {
		sb.WriteString(" ")
		switch *c.Defaults {
		case CreateTableLikeDefaultsIncluding:
			sb.WriteString("INCLUDING DEFAULTS")
		case CreateTableLikeDefaultsExcluding:
			sb.WriteString("EXCLUDING DEFAULTS")
		}
	}

	if c.Kind == CreateTableLikeParenthesized {
		sb.WriteString(")")
	}

	return sb.String()
}

// ============================================================================
// TableVersion - Table version selection
// ============================================================================

// TableVersion specifies a table version selection, e.g., FOR SYSTEM_TIME AS OF.
type TableVersion struct {
	Type TableVersionType
	Expr ast.Expr
}

// TableVersionType represents the type of table version.
type TableVersionType int

const (
	TableVersionForSystemTimeAsOf TableVersionType = iota
	TableVersionTimestampAsOf
	TableVersionVersionAsOf
	TableVersionFunction
)

func (t *TableVersion) exprNode() {}
func (t *TableVersion) expr()     {}
func (t *TableVersion) IsExpr()   {}

// Span returns the source span for this expression.
func (t *TableVersion) Span() token.Span {
	return token.Span{}
}

// String returns the SQL representation.
func (t *TableVersion) String() string {
	switch t.Type {
	case TableVersionForSystemTimeAsOf:
		return fmt.Sprintf("FOR SYSTEM_TIME AS OF %s", t.Expr.String())
	case TableVersionTimestampAsOf:
		return fmt.Sprintf("TIMESTAMP AS OF %s", t.Expr.String())
	case TableVersionVersionAsOf:
		return fmt.Sprintf("VERSION AS OF %s", t.Expr.String())
	case TableVersionFunction:
		return t.Expr.String()
	default:
		return ""
	}
}

// ============================================================================
// CommentDef - Comment with/without equals sign
// ============================================================================

// CommentDef represents a comment with optional equals sign.
type CommentDef struct {
	Type    CommentDefType
	Comment string
}

// CommentDefType represents the type of comment.
type CommentDefType int

const (
	CommentDefWithEq CommentDefType = iota
	CommentDefWithoutEq
)

func (c *CommentDef) exprNode() {}
func (c *CommentDef) expr()     {}
func (c *CommentDef) IsExpr()   {}

// Span returns the source span for this expression.
func (c *CommentDef) Span() token.Span {
	return token.Span{}
}

// String returns the SQL representation.
func (c *CommentDef) String() string {
	return c.Comment
}

// ============================================================================
// OnCommit - ON COMMIT actions
// ============================================================================

// OnCommit represents actions to take ON COMMIT for temporary tables.
type OnCommit int

const (
	OnCommitNone OnCommit = iota
	OnCommitDeleteRows
	OnCommitPreserveRows
	OnCommitDrop
)

func (o OnCommit) String() string {
	switch o {
	case OnCommitDeleteRows:
		return "ON COMMIT DELETE ROWS"
	case OnCommitPreserveRows:
		return "ON COMMIT PRESERVE ROWS"
	case OnCommitDrop:
		return "ON COMMIT DROP"
	default:
		return ""
	}
}

// ColumnOptionReferences stores details for a REFERENCES column constraint.
// Used for inline foreign key constraints like: REFERENCES table_name(col1, col2) [MATCH kind] ON DELETE CASCADE
type ColumnOptionReferences struct {
	Table     *ast.ObjectName
	Columns   []*ast.Ident
	MatchKind *ConstraintReferenceMatchKind
	OnDelete  ReferentialAction
	OnUpdate  ReferentialAction
}

func (c *ColumnOptionReferences) exprNode()        {}
func (c *ColumnOptionReferences) expr()            {}
func (c *ColumnOptionReferences) IsExpr()          {}
func (c *ColumnOptionReferences) Span() token.Span { return token.Span{} }
func (c *ColumnOptionReferences) String() string {
	var sb strings.Builder
	if c.Table != nil {
		sb.WriteString(c.Table.String())
	}
	if len(c.Columns) > 0 {
		// For inline column REFERENCES, add space before column list
		sb.WriteString(" ")
		colStrs := make([]string, len(c.Columns))
		for i, col := range c.Columns {
			colStrs[i] = col.String()
		}
		sb.WriteString("(")
		sb.WriteString(strings.Join(colStrs, ", "))
		sb.WriteString(")")
	}
	if c.MatchKind != nil {
		sb.WriteString(" ")
		sb.WriteString(c.MatchKind.String())
	}
	if c.OnDelete != ReferentialActionNone {
		sb.WriteString(" ON DELETE ")
		sb.WriteString(c.OnDelete.String())
	}
	if c.OnUpdate != ReferentialActionNone {
		sb.WriteString(" ON UPDATE ")
		sb.WriteString(c.OnUpdate.String())
	}
	return sb.String()
}

// ColumnOptionDef represents a column option definition.
// This corresponds to Rust's ColumnOptionDef with name (constraint name) and option.
type ColumnOptionDef struct {
	ConstraintName  *ast.Ident                 // Optional constraint name (e.g., "pkey" in "CONSTRAINT pkey PRIMARY KEY")
	Name            string                     // The option type (e.g., "PRIMARY KEY", "CHECK", "NOT NULL")
	Value           Expr                       // Optional value/expression for the option
	Characteristics *ConstraintCharacteristics // Optional constraint characteristics (DEFERRABLE, ENFORCED, etc.)
}

func (c *ColumnOptionDef) String() string {
	var sb strings.Builder

	// Write CONSTRAINT name if present
	if c.ConstraintName != nil {
		sb.WriteString("CONSTRAINT ")
		sb.WriteString(c.ConstraintName.String())
		sb.WriteString(" ")
	}

	// Handle GENERATED ALWAYS AS with STORED/VIRTUAL
	if strings.HasPrefix(c.Name, "GENERATED ALWAYS AS") && c.Value != nil {
		sb.WriteString("GENERATED ALWAYS AS ")
		sb.WriteString(c.Value.String())
		if strings.HasSuffix(c.Name, " STORED") {
			sb.WriteString(" STORED")
		} else if strings.HasSuffix(c.Name, " VIRTUAL") {
			sb.WriteString(" VIRTUAL")
		}
		return sb.String()
	}

	// Handle COMMENT with quoted string value
	if c.Name == "COMMENT" && c.Value != nil {
		sb.WriteString("COMMENT '")
		sb.WriteString(c.Value.String())
		sb.WriteString("'")
		return sb.String()
	}

	// Handle CHECK constraint - add parentheses around expression
	if c.Name == "CHECK" && c.Value != nil {
		sb.WriteString("CHECK (")
		sb.WriteString(c.Value.String())
		sb.WriteString(")")
		// Add constraint characteristics if present
		if c.Characteristics != nil {
			sb.WriteString(" ")
			sb.WriteString(c.Characteristics.String())
		}
		return sb.String()
	}

	// Handle REFERENCES constraint
	if c.Name == "REFERENCES" {
		sb.WriteString("REFERENCES")
		if c.Value != nil {
			if ref, ok := c.Value.(*ColumnOptionReferences); ok {
				sb.WriteString(" ")
				sb.WriteString(ref.String())
			}
		}
		return sb.String()
	}

	// Handle IDENTITY and AUTOINCREMENT/AUTO_INCREMENT column options
	if c.Name == "IDENTITY" || c.Name == "AUTOINCREMENT" || c.Name == "AUTO_INCREMENT" {
		if c.Value != nil {
			if ident, ok := c.Value.(*ColumnIdentity); ok {
				sb.WriteString(ident.String())
				return sb.String()
			}
		}
		sb.WriteString(c.Name)
		return sb.String()
	}

	// Handle GENERATED ALWAYS/BY DEFAULT AS IDENTITY
	if c.Name == "GENERATED AS IDENTITY" {
		if c.Value != nil {
			if gen, ok := c.Value.(*GeneratedIdentity); ok {
				sb.WriteString(gen.String())
				return sb.String()
			}
		}
		return sb.String()
	}

	// Handle Snowflake TAG option
	if c.Name == "TAG" {
		if c.Value != nil {
			if tagOpt, ok := c.Value.(*TagsColumnOption); ok {
				sb.WriteString(tagOpt.String())
				return sb.String()
			}
		}
		return sb.String()
	}

	// Handle Snowflake MASKING POLICY and PROJECTION POLICY options
	if c.Name == "MASKING POLICY" || c.Name == "PROJECTION POLICY" {
		if c.Value != nil {
			if policy, ok := c.Value.(*ColumnPolicy); ok {
				sb.WriteString(policy.String())
				return sb.String()
			}
		}
		return sb.String()
	}

	// Handle AS (expr) [STORED|VIRTUAL] - MySQL generated column shorthand
	if c.Name == "AS" || c.Name == "AS STORED" || c.Name == "AS VIRTUAL" {
		sb.WriteString("AS (")
		if c.Value != nil {
			sb.WriteString(c.Value.String())
		}
		sb.WriteString(")")
		if c.Name == "AS STORED" {
			sb.WriteString(" STORED")
		} else if c.Name == "AS VIRTUAL" {
			sb.WriteString(" VIRTUAL")
		}
		return sb.String()
	}

	// Handle UNIQUE KEY (MySQL)
	if c.Name == "UNIQUE KEY" {
		sb.WriteString("UNIQUE KEY")
		if c.Characteristics != nil {
			sb.WriteString(" ")
			sb.WriteString(c.Characteristics.String())
		}
		return sb.String()
	}

	// Handle SRID (MySQL spatial reference)
	if c.Name == "SRID" {
		sb.WriteString("SRID")
		if c.Value != nil {
			sb.WriteString(" ")
			sb.WriteString(c.Value.String())
		}
		return sb.String()
	}

	// Handle INVISIBLE/VISIBLE (MySQL column visibility)
	if c.Name == "INVISIBLE" || c.Name == "VISIBLE" {
		sb.WriteString(c.Name)
		return sb.String()
	}

	// Handle CHARACTER SET (MySQL column option)
	if c.Name == "CHARACTER SET" {
		sb.WriteString("CHARACTER SET ")
		if c.Value != nil {
			sb.WriteString(c.Value.String())
		}
		return sb.String()
	}

	// Default: Name + Value
	sb.WriteString(c.Name)
	if c.Value != nil {
		sb.WriteString(" ")
		sb.WriteString(c.Value.String())
	}
	// Add constraint characteristics if present (for PRIMARY KEY, UNIQUE, REFERENCES, etc.)
	if c.Characteristics != nil {
		sb.WriteString(" ")
		sb.WriteString(c.Characteristics.String())
	}
	return sb.String()
}

// GeneratedColumnOption represents a generated/virtual column option.
type GeneratedColumnOption struct {
	Expression    Expr
	GeneratedType string // "STORED", "VIRTUAL", or ""
}

func (g *GeneratedColumnOption) exprNode()        {}
func (g *GeneratedColumnOption) expr()            {}
func (g *GeneratedColumnOption) IsExpr()          {}
func (g *GeneratedColumnOption) Span() token.Span { return token.Span{} }
func (g *GeneratedColumnOption) String() string {
	var sb strings.Builder
	sb.WriteString("GENERATED ALWAYS AS ")
	if g.Expression != nil {
		sb.WriteString(g.Expression.String())
		sb.WriteString(" ")
	}
	if g.GeneratedType != "" {
		sb.WriteString(g.GeneratedType)
	}
	return strings.TrimSpace(sb.String())
}

// ============================================================================
// Snowflake Column Policy and Tag Types
// ============================================================================

// ColumnPolicyType represents the type of column policy (masking or projection).
type ColumnPolicyType int

const (
	ColumnPolicyTypeMasking ColumnPolicyType = iota
	ColumnPolicyTypeProjection
)

// ColumnPolicy represents a column policy (MASKING POLICY or PROJECTION POLICY).
type ColumnPolicy struct {
	Type       ColumnPolicyType
	With       bool // Whether WITH prefix was used
	PolicyName *ast.ObjectName
	UsingCols  []*ast.Ident // Optional USING (col1, col2, ...)
}

func (c *ColumnPolicy) exprNode()        {}
func (c *ColumnPolicy) expr()            {}
func (c *ColumnPolicy) IsExpr()          {}
func (c *ColumnPolicy) Span() token.Span { return token.Span{} }
func (c *ColumnPolicy) IsColumnOption()  {}

func (c *ColumnPolicy) String() string {
	var sb strings.Builder
	if c.With {
		sb.WriteString("WITH ")
	}
	if c.Type == ColumnPolicyTypeMasking {
		sb.WriteString("MASKING POLICY ")
	} else {
		sb.WriteString("PROJECTION POLICY ")
	}
	if c.PolicyName != nil {
		sb.WriteString(c.PolicyName.String())
	}
	if len(c.UsingCols) > 0 {
		sb.WriteString(" USING (")
		for i, col := range c.UsingCols {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(col.String())
		}
		sb.WriteString(")")
	}
	return sb.String()
}

// SnowflakeTag represents a Snowflake tag assignment (tag_name = 'tag_value').
// This is distinct from the generic Tag type used elsewhere.
type SnowflakeTag struct {
	Name  *ast.Ident
	Value Expr
}

func (t *SnowflakeTag) exprNode()        {}
func (t *SnowflakeTag) expr()            {}
func (t *SnowflakeTag) IsExpr()          {}
func (t *SnowflakeTag) Span() token.Span { return token.Span{} }

func (t *SnowflakeTag) String() string {
	var sb strings.Builder
	if t.Name != nil {
		sb.WriteString(t.Name.String())
	}
	if t.Value != nil {
		sb.WriteString("=")
		sb.WriteString(t.Value.String())
	}
	return sb.String()
}

// TagsColumnOption represents Snowflake TAG option for columns.
type TagsColumnOption struct {
	With bool            // Whether WITH prefix was used
	Tags []*SnowflakeTag // List of tag assignments
}

func (t *TagsColumnOption) exprNode()        {}
func (t *TagsColumnOption) expr()            {}
func (t *TagsColumnOption) IsExpr()          {}
func (t *TagsColumnOption) Span() token.Span { return token.Span{} }
func (t *TagsColumnOption) IsColumnOption()  {}

func (t *TagsColumnOption) String() string {
	var sb strings.Builder
	if t.With {
		sb.WriteString("WITH ")
	}
	sb.WriteString("TAG (")
	for i, tag := range t.Tags {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(tag.String())
	}
	sb.WriteString(")")
	return sb.String()
}

// ============================================================================
// View Column Definition (for CREATE VIEW with column options)
// ============================================================================

// ViewColumnDef represents a column definition in a CREATE VIEW statement.
// Unlike table columns, view columns can have options like TAG and POLICY (Snowflake).
type ViewColumnDef struct {
	Name     *ast.Ident
	DataType interface{}        // Optional data type (ClickHouse)
	Options  []*ColumnOptionDef // Column options (TAG, POLICY, COMMENT, etc.)
	SpanVal  token.Span
}

func (v *ViewColumnDef) exprNode()        {}
func (v *ViewColumnDef) expr()            {}
func (v *ViewColumnDef) IsExpr()          {}
func (v *ViewColumnDef) Span() token.Span { return v.SpanVal }

func (v *ViewColumnDef) String() string {
	var sb strings.Builder
	if v.Name != nil {
		sb.WriteString(v.Name.String())
	}
	for _, opt := range v.Options {
		sb.WriteString(" ")
		sb.WriteString(opt.String())
	}
	return sb.String()
}

// ============================================================================
// Additional DDL Types (Stubs for compilation)
// ============================================================================

// OneOrManyWithParens represents one or many items with parentheses.
type OneOrManyWithParens struct {
	Items []Expr
}

func (o *OneOrManyWithParens) exprNode()        {}
func (o *OneOrManyWithParens) expr()            {}
func (o *OneOrManyWithParens) IsExpr()          {}
func (o *OneOrManyWithParens) Span() token.Span { return token.Span{} }
func (o *OneOrManyWithParens) String() string   { return "(...)" }

// WrappedCollection represents a wrapped collection of items.
type WrappedCollection struct {
	Items []Expr
}

func (w *WrappedCollection) exprNode()        {}
func (w *WrappedCollection) expr()            {}
func (w *WrappedCollection) IsExpr()          {}
func (w *WrappedCollection) Span() token.Span { return token.Span{} }
func (w *WrappedCollection) String() string   { return "(...)" }

// ClusteredBy represents CLUSTER BY clause.
type ClusteredBy struct {
	Columns []*ast.Ident
}

func (c *ClusteredBy) exprNode()        {}
func (c *ClusteredBy) expr()            {}
func (c *ClusteredBy) IsExpr()          {}
func (c *ClusteredBy) Span() token.Span { return token.Span{} }
func (c *ClusteredBy) String() string   { return "CLUSTERED BY" }

// ForValuesKind represents the kind of FOR VALUES clause.
type ForValuesKind int

const (
	ForValuesKindIn ForValuesKind = iota
	ForValuesKindFrom
	ForValuesKindWith
	ForValuesKindDefault
)

// ForValues represents FOR VALUES clause for PostgreSQL PARTITION OF.
type ForValues struct {
	Kind      ForValuesKind
	Values    []Expr                // For IN
	From      []PartitionBoundValue // For FROM...TO
	To        []PartitionBoundValue // For FROM...TO
	Modulus   uint64                // For WITH (MODULUS, REMAINDER)
	Remainder uint64                // For WITH (MODULUS, REMAINDER)
}

func (f *ForValues) exprNode()        {}
func (f *ForValues) expr()            {}
func (f *ForValues) IsExpr()          {}
func (f *ForValues) Span() token.Span { return token.Span{} }
func (f *ForValues) String() string {
	switch f.Kind {
	case ForValuesKindDefault:
		return "DEFAULT"
	case ForValuesKindIn:
		var parts []string
		for _, v := range f.Values {
			parts = append(parts, v.String())
		}
		return fmt.Sprintf("FOR VALUES IN (%s)", strings.Join(parts, ", "))
	case ForValuesKindFrom:
		var fromParts []string
		for _, v := range f.From {
			fromParts = append(fromParts, v.String())
		}
		var toParts []string
		for _, v := range f.To {
			toParts = append(toParts, v.String())
		}
		return fmt.Sprintf("FOR VALUES FROM (%s) TO (%s)", strings.Join(fromParts, ", "), strings.Join(toParts, ", "))
	case ForValuesKindWith:
		return fmt.Sprintf("FOR VALUES WITH (MODULUS %d, REMAINDER %d)", f.Modulus, f.Remainder)
	default:
		return "FOR VALUES"
	}
}

// PartitionBoundValue represents a partition bound value (expression, MINVALUE, or MAXVALUE).
type PartitionBoundValue struct {
	IsMinValue bool
	IsMaxValue bool
	Expr       Expr
}

func (p *PartitionBoundValue) exprNode()        {}
func (p *PartitionBoundValue) expr()            {}
func (p *PartitionBoundValue) IsExpr()          {}
func (p *PartitionBoundValue) Span() token.Span { return token.Span{} }
func (p *PartitionBoundValue) String() string {
	if p.IsMinValue {
		return "MINVALUE"
	}
	if p.IsMaxValue {
		return "MAXVALUE"
	}
	if p.Expr != nil {
		return p.Expr.String()
	}
	return ""
}

// RowAccessPolicy represents row access policy.
type RowAccessPolicy struct {
	Name *ast.ObjectName
}

func (r *RowAccessPolicy) exprNode()        {}
func (r *RowAccessPolicy) expr()            {}
func (r *RowAccessPolicy) IsExpr()          {}
func (r *RowAccessPolicy) Span() token.Span { return token.Span{} }
func (r *RowAccessPolicy) String() string   { return "ROW ACCESS POLICY" }

// StorageLifecyclePolicy represents storage lifecycle policy.
type StorageLifecyclePolicy struct {
	Name string
}

func (s *StorageLifecyclePolicy) exprNode()        {}
func (s *StorageLifecyclePolicy) expr()            {}
func (s *StorageLifecyclePolicy) IsExpr()          {}
func (s *StorageLifecyclePolicy) Span() token.Span { return token.Span{} }
func (s *StorageLifecyclePolicy) String() string   { return "STORAGE LIFECYCLE POLICY" }

// Tag represents a tag.
type Tag struct {
	Name  string
	Value string
}

func (t *Tag) exprNode()        {}
func (t *Tag) expr()            {}
func (t *Tag) IsExpr()          {}
func (t *Tag) Span() token.Span { return token.Span{} }
func (t *Tag) String() string   { return fmt.Sprintf("%s=%s", t.Name, t.Value) }

// StorageSerializationPolicy represents storage serialization policy.
type StorageSerializationPolicy int

const (
	StorageSerializationPolicyNone StorageSerializationPolicy = iota
)

func (s StorageSerializationPolicy) String() string { return "" }

// RefreshModeKind represents refresh mode kind.
type RefreshModeKind int

const (
	RefreshModeKindNone RefreshModeKind = iota
)

func (r RefreshModeKind) String() string { return "" }

// InitializeKind represents initialization kind.
type InitializeKind int

const (
	InitializeKindNone InitializeKind = iota
	InitializeKindOnCreate
	InitializeKindOnSchedule
)

func (i InitializeKind) String() string {
	switch i {
	case InitializeKindOnCreate:
		return "ON_CREATE"
	case InitializeKindOnSchedule:
		return "ON_SCHEDULE"
	default:
		return ""
	}
}

// ViewEnvelope represents view envelope.
type ViewEnvelope struct{}

func (v *ViewEnvelope) exprNode()        {}
func (v *ViewEnvelope) expr()            {}
func (v *ViewEnvelope) IsExpr()          {}
func (v *ViewEnvelope) Span() token.Span { return token.Span{} }
func (v *ViewEnvelope) String() string   { return "" }

// IndexColumn represents an index column.
type IndexColumn struct {
	Expr       Expr
	Opclass    *ast.ObjectName
	Asc        *bool // nil means not specified
	NullsFirst *bool // nil means not specified
}

func (i *IndexColumn) exprNode()        {}
func (i *IndexColumn) expr()            {}
func (i *IndexColumn) IsExpr()          {}
func (i *IndexColumn) Span() token.Span { return token.Span{} }
func (i *IndexColumn) String() string {
	var parts []string
	if i.Expr != nil {
		parts = append(parts, i.Expr.String())
	}
	if i.Opclass != nil {
		parts = append(parts, i.Opclass.String())
	}
	if i.Asc != nil {
		if *i.Asc {
			parts = append(parts, "ASC")
		} else {
			parts = append(parts, "DESC")
		}
	}
	if i.NullsFirst != nil {
		if *i.NullsFirst {
			parts = append(parts, "NULLS FIRST")
		} else {
			parts = append(parts, "NULLS LAST")
		}
	}
	return strings.Join(parts, " ")
}

// ArgMode represents function argument mode (IN, OUT, INOUT)
type ArgMode int

const (
	ArgModeNone ArgMode = iota
	ArgModeIn
	ArgModeOut
	ArgModeInOut
)

func (a ArgMode) String() string {
	switch a {
	case ArgModeIn:
		return "IN"
	case ArgModeOut:
		return "OUT"
	case ArgModeInOut:
		return "INOUT"
	default:
		return ""
	}
}

// OperateFunctionArg represents operate function argument.
type OperateFunctionArg struct {
	Mode        *ArgMode
	Name        *ast.Ident
	DataType    interface{} // datatype.DataType - using interface{} to avoid import cycle
	DefaultExpr Expr
	DefaultOp   string // "=" or "DEFAULT" or "", empty means no default
}

func (o *OperateFunctionArg) exprNode()        {}
func (o *OperateFunctionArg) expr()            {}
func (o *OperateFunctionArg) IsExpr()          {}
func (o *OperateFunctionArg) Span() token.Span { return token.Span{} }
func (o *OperateFunctionArg) String() string {
	var f strings.Builder
	if o.Mode != nil && *o.Mode != ArgModeNone {
		f.WriteString(o.Mode.String())
		f.WriteString(" ")
	}
	if o.Name != nil {
		f.WriteString(o.Name.String())
		f.WriteString(" ")
	}
	if o.DataType != nil {
		if s, ok := o.DataType.(fmt.Stringer); ok {
			f.WriteString(s.String())
		}
	}
	if o.DefaultExpr != nil {
		op := o.DefaultOp
		if op == "" {
			op = "="
		}
		f.WriteString(" ")
		f.WriteString(op)
		f.WriteString(" ")
		f.WriteString(o.DefaultExpr.String())
	}
	return f.String()
}

// FunctionReturnTypeKind represents the kind of function return type
type FunctionReturnTypeKind int

const (
	FunctionReturnTypeDataType FunctionReturnTypeKind = iota
	FunctionReturnTypeSetOf
)

// FunctionReturnType represents function return type.
type FunctionReturnType struct {
	Kind     FunctionReturnTypeKind
	DataType interface{} // datatype.DataType - using interface{} to avoid import cycle
}

func (f *FunctionReturnType) exprNode()        {}
func (f *FunctionReturnType) expr()            {}
func (f *FunctionReturnType) IsExpr()          {}
func (f *FunctionReturnType) Span() token.Span { return token.Span{} }
func (f *FunctionReturnType) String() string {
	if f.DataType == nil {
		return ""
	}
	s, ok := f.DataType.(fmt.Stringer)
	if !ok {
		return ""
	}
	if f.Kind == FunctionReturnTypeSetOf {
		return "SETOF " + s.String()
	}
	return s.String()
}

// FunctionBehavior represents function behavior.
type FunctionBehavior int

const (
	FunctionBehaviorNone FunctionBehavior = iota
	FunctionBehaviorImmutable
	FunctionBehaviorStable
	FunctionBehaviorVolatile
)

func (f FunctionBehavior) String() string {
	switch f {
	case FunctionBehaviorImmutable:
		return "IMMUTABLE"
	case FunctionBehaviorStable:
		return "STABLE"
	case FunctionBehaviorVolatile:
		return "VOLATILE"
	default:
		return ""
	}
}

// FunctionCalledOnNull represents function called on null.
type FunctionCalledOnNull int

const (
	FunctionCalledOnNullNone FunctionCalledOnNull = iota
	FunctionCalledOnNullCalledOnNullInput
	FunctionCalledOnNullReturnsNullOnNullInput
	FunctionCalledOnNullStrict
)

func (f FunctionCalledOnNull) String() string {
	switch f {
	case FunctionCalledOnNullCalledOnNullInput:
		return "CALLED ON NULL INPUT"
	case FunctionCalledOnNullReturnsNullOnNullInput:
		return "RETURNS NULL ON NULL INPUT"
	case FunctionCalledOnNullStrict:
		return "STRICT"
	default:
		return ""
	}
}

// FunctionParallel represents function parallel.
type FunctionParallel int

const (
	FunctionParallelNone FunctionParallel = iota
	FunctionParallelUnsafe
	FunctionParallelRestricted
	FunctionParallelSafe
)

func (f FunctionParallel) String() string {
	switch f {
	case FunctionParallelUnsafe:
		return "PARALLEL UNSAFE"
	case FunctionParallelRestricted:
		return "PARALLEL RESTRICTED"
	case FunctionParallelSafe:
		return "PARALLEL SAFE"
	default:
		return ""
	}
}

// FunctionSecurity represents function security.
type FunctionSecurity int

const (
	FunctionSecurityNone FunctionSecurity = iota
	FunctionSecurityDefiner
	FunctionSecurityInvoker
)

func (f FunctionSecurity) String() string {
	switch f {
	case FunctionSecurityDefiner:
		return "SECURITY DEFINER"
	case FunctionSecurityInvoker:
		return "SECURITY INVOKER"
	default:
		return ""
	}
}

// FunctionDeterminismSpecifier represents function determinism specifier.
type FunctionDeterminismSpecifier int

const (
	FunctionDeterminismSpecifierNone FunctionDeterminismSpecifier = iota
)

func (f FunctionDeterminismSpecifier) String() string { return "" }

// FunctionSetValueKind represents the kind of function SET value
type FunctionSetValueKind int

const (
	FunctionSetValueFromCurrent FunctionSetValueKind = iota
	FunctionSetValueExpr
)

// FunctionSetValue represents a function SET parameter value
type FunctionSetValue struct {
	Kind  FunctionSetValueKind
	Exprs []Expr // For multiple values like SET param = value1, value2, ...
}

// CreateFunctionBody represents function body.
type CreateFunctionBody struct {
	Value          string // For AS 'string' syntax
	ReturnExpr     Expr   // For RETURN expr syntax
	IsDollarQuoted bool   // Whether the original was a dollar-quoted string
}

func (c *CreateFunctionBody) exprNode()        {}
func (c *CreateFunctionBody) expr()            {}
func (c *CreateFunctionBody) IsExpr()          {}
func (c *CreateFunctionBody) Span() token.Span { return token.Span{} }
func (c *CreateFunctionBody) String() string {
	if c.ReturnExpr != nil {
		return "RETURN " + c.ReturnExpr.String()
	}
	// Add quotes around the value if it's a string body
	if c.IsDollarQuoted {
		return "$$" + c.Value + "$$"
	}
	return "'" + c.Value + "'"
}

// FunctionDefinitionSetParam represents function definition set parameter.
type FunctionDefinitionSetParam struct {
	Name  *ast.Ident
	Value FunctionSetValue
}

func (f *FunctionDefinitionSetParam) exprNode()        {}
func (f *FunctionDefinitionSetParam) expr()            {}
func (f *FunctionDefinitionSetParam) IsExpr()          {}
func (f *FunctionDefinitionSetParam) Span() token.Span { return token.Span{} }
func (f *FunctionDefinitionSetParam) String() string {
	var b strings.Builder
	b.WriteString("SET ")
	b.WriteString(f.Name.String())
	if f.Value.Kind == FunctionSetValueFromCurrent {
		b.WriteString(" FROM CURRENT")
	} else if len(f.Value.Exprs) > 0 {
		b.WriteString(" = ")
		for i, expr := range f.Value.Exprs {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(expr.String())
		}
	}
	return b.String()
}

// SqlSecurity represents SQL security.
type SqlSecurity int

const (
	SqlSecurityNone SqlSecurity = iota
)

func (s SqlSecurity) String() string { return "" }

// RemoteProperty represents remote property.
type RemoteProperty struct{}

func (r *RemoteProperty) exprNode()        {}
func (r *RemoteProperty) expr()            {}
func (r *RemoteProperty) IsExpr()          {}
func (r *RemoteProperty) Span() token.Span { return token.Span{} }
func (r *RemoteProperty) String() string   { return "" }

// ProcedureParam represents procedure parameter.
type ProcedureParam struct {
	SpanVal  token.Span
	Name     *Ident
	DataType interface{} // datatype.DataType - using interface{} to avoid import cycle
	Mode     *ArgMode    // IN, OUT, INOUT
	Default  Expr        // Optional default value
}

func (p *ProcedureParam) exprNode()        {}
func (p *ProcedureParam) expr()            {}
func (p *ProcedureParam) IsExpr()          {}
func (p *ProcedureParam) Span() token.Span { return p.SpanVal }
func (p *ProcedureParam) String() string {
	var sb strings.Builder
	if p.Mode != nil {
		sb.WriteString(p.Mode.String())
		sb.WriteString(" ")
	}
	sb.WriteString(p.Name.String())
	if p.DataType != nil {
		sb.WriteString(" ")
		if dt, ok := p.DataType.(fmt.Stringer); ok {
			sb.WriteString(dt.String())
		}
	}
	if p.Default != nil {
		sb.WriteString(" = ")
		sb.WriteString(p.Default.String())
	}
	return sb.String()
}

// ExecuteAs represents EXECUTE AS clause.
type ExecuteAs int

const (
	ExecuteAsNone ExecuteAs = iota
)

func (e ExecuteAs) String() string { return "" }

// ReplicaIdentityType represents the type of replica identity for ALTER TABLE REPLICA IDENTITY
type ReplicaIdentityType int

const (
	ReplicaIdentityNothing ReplicaIdentityType = iota
	ReplicaIdentityFull
	ReplicaIdentityDefault
	ReplicaIdentityIndex
)

func (r ReplicaIdentityType) String() string {
	switch r {
	case ReplicaIdentityNothing:
		return "NOTHING"
	case ReplicaIdentityFull:
		return "FULL"
	case ReplicaIdentityDefault:
		return "DEFAULT"
	case ReplicaIdentityIndex:
		return "USING INDEX"
	}
	return ""
}

// AlterTableOperation represents ALTER TABLE operation.
type AlterTableOperation struct {
	// Operation type
	Op AlterTableOpType

	// Fields for AddColumn
	AddColumnKeyword  bool
	AddIfNotExists    bool
	AddColumnDef      *ColumnDef
	AddColumnDefs     []*ColumnDef         // For multiple columns: ADD COLUMN (c1 INT, c2 INT)
	AddColumnPosition *MySQLColumnPosition // MySQL: FIRST or AFTER column

	// Fields for DropColumn
	DropColumnKeyword bool
	DropIfExists      bool
	DropColumnNames   []*ast.Ident
	DropBehavior      DropBehavior

	// Fields for AddConstraint
	Constraint         *TableConstraint
	ConstraintNotValid bool

	// Fields for DropConstraint
	DropConstraintIfExists bool
	DropConstraintName     *ast.Ident

	// Fields for RenameConstraint (PostgreSQL: RENAME CONSTRAINT)
	RenameConstraintOldName *ast.Ident
	RenameConstraintNewName *ast.Ident

	// Fields for ValidateConstraint (PostgreSQL: VALIDATE CONSTRAINT)
	ValidateConstraintName *ast.Ident

	// Fields for RenameColumn
	RenameOldColumn *ast.Ident
	RenameNewColumn *ast.Ident

	// Fields for RenameTable
	NewTableName      *ast.ObjectName
	RenameTableAsKind RenameTableAsKind // Whether RENAME AS or RENAME TO was used

	// Fields for AlterColumn
	AlterColumnName     *ast.Ident
	AlterColumnOp       AlterColumnOpType
	AlterDefault        Expr
	AlterDataType       interface{} // datatype.DataType
	AlterUsing          Expr        // USING expression for SET DATA TYPE
	AlterDataTypeHadSet bool        // Whether SET DATA TYPE (true) or just TYPE (false)

	// Fields for AlterColumn ADD GENERATED (PostgreSQL)
	AlterGeneratedAs           GeneratedAs        // ALWAYS, BY DEFAULT, or None
	AlterGeneratedSequenceOpts []*SequenceOptions // Sequence options within parentheses
	AlterGeneratedHasParens    bool               // Whether () was present (even if empty)

	// Fields for SetTblProperties
	TblProperties []*SqlOption

	// Fields for SetOptions (MySQL: AUTO_INCREMENT, ALGORITHM, LOCK)
	AutoIncrementValue string     // For AUTO_INCREMENT = N
	AlgorithmValue     *ast.Ident // For ALGORITHM = {COPY|INPLACE|INSTANT}
	LockValue          *ast.Ident // For LOCK = {DEFAULT|NONE|SHARED|EXCLUSIVE}

	// Fields for SetOptionsParens (PostgreSQL: SET (key=value, ...))
	SetOptions []*SqlOption // For SET (key=value, ...)

	// Fields for ChangeColumn (MySQL)
	ChangeOldName        *ast.Ident
	ChangeNewName        *ast.Ident
	ChangeDataType       interface{} // datatype.DataType
	ChangeOptions        []*ColumnOption
	ChangeColumnPosition *MySQLColumnPosition

	// Fields for ModifyColumn (MySQL)
	ModifyColumnName     *ast.Ident
	ModifyDataType       interface{} // datatype.DataType
	ModifyOptions        []*ColumnOption
	ModifyColumnPosition *MySQLColumnPosition

	// Fields for Enable/Disable operations (PostgreSQL)
	DisableEnableName *ast.Ident // Name of rule or trigger

	// Fields for ReplicaIdentity (PostgreSQL)
	ReplicaIdentity      ReplicaIdentityType
	ReplicaIdentityIndex *ast.Ident

	// Fields for SwapWith (Snowflake)
	SwapWithTableName *ast.ObjectName

	// Fields for ClusterBy (Snowflake)
	ClusterBy []Expr

	// Fields for Refresh (Snowflake - ALTER EXTERNAL TABLE)
	RefreshSubpath *string

	// Fields for OwnerTo (PostgreSQL)
	NewOwner *OwnerToTarget

	// Span
	SpanVal token.Span
}

// OwnerToTarget represents the new owner in OWNER TO statements
// Can be an identifier or a special value like CURRENT_USER, CURRENT_ROLE, SESSION_USER
type OwnerToTarget struct {
	Ident     *ast.Ident // The owner identifier (preserves quote style)
	Name      string     // The owner name (for special values or when ident is nil)
	IsSpecial bool       // Whether this is a special value like CURRENT_USER
}

func (o *OwnerToTarget) String() string {
	if o.Ident != nil {
		return o.Ident.String()
	}
	return o.Name
}

// AlterTableOpType represents the type of ALTER TABLE operation
type AlterTableOpType int

const (
	AlterTableOpAddColumn AlterTableOpType = iota
	AlterTableOpDropColumn
	AlterTableOpAddConstraint
	AlterTableOpDropConstraint
	AlterTableOpRenameConstraint
	AlterTableOpRenameColumn
	AlterTableOpRenameTable
	AlterTableOpAlterColumn
	AlterTableOpSetTblProperties
	AlterTableOpSetOptionsParens
	AlterTableOpSetOptions // MySQL: AUTO_INCREMENT, ALGORITHM, LOCK
	AlterTableOpChangeColumn
	AlterTableOpModifyColumn
	AlterTableOpDropPrimaryKey
	AlterTableOpDropForeignKey
	AlterTableOpDropIndex
	AlterTableOpDisableRowLevelSecurity
	AlterTableOpEnableRowLevelSecurity
	AlterTableOpForceRowLevelSecurity
	AlterTableOpNoForceRowLevelSecurity
	AlterTableOpDisableTrigger
	AlterTableOpEnableTrigger
	AlterTableOpDisableRule
	AlterTableOpEnableRule
	AlterTableOpEnableAlwaysRule
	AlterTableOpEnableReplicaRule
	AlterTableOpEnableAlwaysTrigger
	AlterTableOpEnableReplicaTrigger
	AlterTableOpReplicaIdentity
	AlterTableOpValidateConstraint
	// PostgreSQL OWNER TO
	AlterTableOpOwnerTo
	// Snowflake-specific clustering operations
	AlterTableOpDropClusteringKey
	AlterTableOpSuspendRecluster
	AlterTableOpResumeRecluster
	// Snowflake-specific SWAP WITH operation
	AlterTableOpSwapWith
	// Snowflake-specific CLUSTER BY operation
	AlterTableOpClusterBy
	// Snowflake-specific dynamic/external table operations
	AlterTableOpRefresh
	AlterTableOpSuspend
	AlterTableOpResume
)

// AlterTableType represents the type of table for ALTER TABLE (e.g., ICEBERG, DYNAMIC, EXTERNAL)
type AlterTableType int

const (
	AlterTableTypeNone AlterTableType = iota
	AlterTableTypeIceberg
	AlterTableTypeDynamic
	AlterTableTypeExternal
)

func (a AlterTableType) String() string {
	switch a {
	case AlterTableTypeIceberg:
		return "ICEBERG"
	case AlterTableTypeDynamic:
		return "DYNAMIC"
	case AlterTableTypeExternal:
		return "EXTERNAL"
	default:
		return ""
	}
}

// RenameTableAsKind represents whether RENAME AS or RENAME TO was used
type RenameTableAsKind int

const (
	RenameTableTo RenameTableAsKind = iota
	RenameTableAs
)

// AlterColumnOpType represents operations on a column via ALTER COLUMN
type AlterColumnOpType int

const (
	AlterColumnOpSetNotNull AlterColumnOpType = iota
	AlterColumnOpDropNotNull
	AlterColumnOpSetDefault
	AlterColumnOpDropDefault
	AlterColumnOpSetDataType
	AlterColumnOpAddGenerated // PostgreSQL: ADD GENERATED {ALWAYS|BY DEFAULT} AS IDENTITY
)

func (a *AlterTableOperation) exprNode()        {}
func (a *AlterTableOperation) expr()            {}
func (a *AlterTableOperation) IsExpr()          {}
func (a *AlterTableOperation) Span() token.Span { return a.SpanVal }
func (a *AlterTableOperation) String() string {
	switch a.Op {
	case AlterTableOpAddColumn:
		var buf strings.Builder
		buf.WriteString("ADD ")
		if a.AddColumnKeyword {
			buf.WriteString("COLUMN ")
		}
		if a.AddIfNotExists {
			buf.WriteString("IF NOT EXISTS ")
		}
		if len(a.AddColumnDefs) > 0 {
			// Multiple columns in parentheses
			buf.WriteString("(")
			for i, colDef := range a.AddColumnDefs {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(colDef.String())
			}
			buf.WriteString(")")
		} else if a.AddColumnDef != nil {
			buf.WriteString(a.AddColumnDef.String())
		}
		if a.AddColumnPosition != nil {
			buf.WriteString(" ")
			buf.WriteString(a.AddColumnPosition.String())
		}
		return buf.String()
	case AlterTableOpDropColumn:
		var buf strings.Builder
		buf.WriteString("DROP ")
		if a.DropColumnKeyword {
			buf.WriteString("COLUMN ")
		}
		if a.DropIfExists {
			buf.WriteString("IF EXISTS ")
		}
		for i, name := range a.DropColumnNames {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(name.String())
		}
		if a.DropBehavior != DropBehaviorNone {
			buf.WriteString(" ")
			buf.WriteString(a.DropBehavior.String())
		}
		return buf.String()
	case AlterTableOpAddConstraint:
		var buf strings.Builder
		buf.WriteString("ADD ")
		if a.Constraint != nil {
			buf.WriteString(a.Constraint.String())
		}
		if a.ConstraintNotValid {
			buf.WriteString(" NOT VALID")
		}
		return buf.String()
	case AlterTableOpDropConstraint:
		var buf strings.Builder
		buf.WriteString("DROP CONSTRAINT ")
		if a.DropConstraintIfExists {
			buf.WriteString("IF EXISTS ")
		}
		if a.DropConstraintName != nil {
			buf.WriteString(a.DropConstraintName.String())
		}
		if a.DropBehavior != DropBehaviorNone {
			buf.WriteString(" ")
			buf.WriteString(a.DropBehavior.String())
		}
		return buf.String()
	case AlterTableOpRenameConstraint:
		var buf strings.Builder
		buf.WriteString("RENAME CONSTRAINT ")
		if a.RenameConstraintOldName != nil {
			buf.WriteString(a.RenameConstraintOldName.String())
		}
		buf.WriteString(" TO ")
		if a.RenameConstraintNewName != nil {
			buf.WriteString(a.RenameConstraintNewName.String())
		}
		return buf.String()
	case AlterTableOpRenameColumn:
		var buf strings.Builder
		buf.WriteString("RENAME COLUMN ")
		if a.RenameOldColumn != nil {
			buf.WriteString(a.RenameOldColumn.String())
		}
		buf.WriteString(" TO ")
		if a.RenameNewColumn != nil {
			buf.WriteString(a.RenameNewColumn.String())
		}
		return buf.String()
	case AlterTableOpRenameTable:
		var buf strings.Builder
		if a.RenameTableAsKind == RenameTableAs {
			buf.WriteString("RENAME AS ")
		} else {
			buf.WriteString("RENAME TO ")
		}
		if a.NewTableName != nil {
			buf.WriteString(a.NewTableName.String())
		}
		return buf.String()
	case AlterTableOpAlterColumn:
		var buf strings.Builder
		buf.WriteString("ALTER COLUMN ")
		if a.AlterColumnName != nil {
			buf.WriteString(a.AlterColumnName.String())
		}
		switch a.AlterColumnOp {
		case AlterColumnOpSetNotNull:
			buf.WriteString(" SET NOT NULL")
		case AlterColumnOpDropNotNull:
			buf.WriteString(" DROP NOT NULL")
		case AlterColumnOpSetDefault:
			buf.WriteString(" SET DEFAULT ")
			if a.AlterDefault != nil {
				buf.WriteString(a.AlterDefault.String())
			}
		case AlterColumnOpDropDefault:
			buf.WriteString(" DROP DEFAULT")
		case AlterColumnOpSetDataType:
			// Output either "SET DATA TYPE" or just "TYPE" depending on what was parsed
			if a.AlterDataTypeHadSet {
				buf.WriteString(" SET DATA TYPE")
			} else {
				buf.WriteString(" TYPE")
			}
			if a.AlterDataType != nil {
				if dt, ok := a.AlterDataType.(fmt.Stringer); ok {
					buf.WriteString(" ")
					buf.WriteString(dt.String())
				}
			}
			// Output USING clause if present (PostgreSQL)
			if a.AlterUsing != nil {
				buf.WriteString(" USING ")
				buf.WriteString(a.AlterUsing.String())
			}
		case AlterColumnOpAddGenerated:
			buf.WriteString(" ADD GENERATED")
			if a.AlterGeneratedAs != GeneratedAsNone {
				buf.WriteString(" ")
				buf.WriteString(a.AlterGeneratedAs.String())
			}
			buf.WriteString(" AS IDENTITY")
			// Output parentheses if present (with options or empty)
			if a.AlterGeneratedHasParens {
				buf.WriteString(" (")
				for i, opt := range a.AlterGeneratedSequenceOpts {
					if i > 0 {
						buf.WriteString(" ")
					}
					buf.WriteString(opt.String())
				}
				buf.WriteString(")")
			}
		}
		return buf.String()
	case AlterTableOpSetTblProperties:
		var buf strings.Builder
		buf.WriteString("SET TBLPROPERTIES(")
		for i, opt := range a.TblProperties {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(fmt.Sprintf("'%s' = '%s'", opt.Name.String(), opt.Value.String()))
		}
		buf.WriteString(")")
		return buf.String()
	case AlterTableOpSetOptions:
		var buf strings.Builder
		if a.AutoIncrementValue != "" {
			buf.WriteString("AUTO_INCREMENT = ")
			buf.WriteString(a.AutoIncrementValue)
		} else if a.AlgorithmValue != nil {
			buf.WriteString("ALGORITHM = ")
			// MySQL algorithm values should be uppercase per Rust reference
			// src/ast/ddl.rs:606-615 - INPLACE, INSTANT, COPY, DEFAULT
			buf.WriteString(strings.ToUpper(a.AlgorithmValue.Value))
		} else if a.LockValue != nil {
			buf.WriteString("LOCK = ")
			// MySQL lock values should be uppercase per Rust reference
			// src/ast/ddl.rs:635-644 - EXCLUSIVE, SHARED, NONE, DEFAULT
			buf.WriteString(strings.ToUpper(a.LockValue.Value))
		}
		return buf.String()
	case AlterTableOpSetOptionsParens:
		var buf strings.Builder
		buf.WriteString("SET (")
		for i, opt := range a.SetOptions {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(opt.Name.String())
			buf.WriteString(" = ")
			if opt.Value != nil {
				buf.WriteString(opt.Value.String())
			}
		}
		buf.WriteString(")")
		return buf.String()
	case AlterTableOpDropPrimaryKey:
		var buf strings.Builder
		buf.WriteString("DROP PRIMARY KEY")
		if a.DropBehavior != DropBehaviorNone {
			buf.WriteString(" ")
			buf.WriteString(a.DropBehavior.String())
		}
		return buf.String()
	case AlterTableOpDropForeignKey:
		var buf strings.Builder
		buf.WriteString("DROP FOREIGN KEY ")
		if len(a.DropColumnNames) > 0 {
			buf.WriteString(a.DropColumnNames[0].String())
		}
		if a.DropBehavior != DropBehaviorNone {
			buf.WriteString(" ")
			buf.WriteString(a.DropBehavior.String())
		}
		return buf.String()
	case AlterTableOpDropIndex:
		var buf strings.Builder
		buf.WriteString("DROP INDEX ")
		if len(a.DropColumnNames) > 0 {
			buf.WriteString(a.DropColumnNames[0].String())
		}
		return buf.String()
	case AlterTableOpChangeColumn:
		var buf strings.Builder
		buf.WriteString("CHANGE COLUMN ")
		if a.ChangeOldName != nil {
			buf.WriteString(a.ChangeOldName.String())
			buf.WriteString(" ")
		}
		if a.ChangeNewName != nil {
			buf.WriteString(a.ChangeNewName.String())
			buf.WriteString(" ")
		}
		if a.ChangeDataType != nil {
			if dt, ok := a.ChangeDataType.(fmt.Stringer); ok {
				buf.WriteString(dt.String())
			}
		}
		if a.ChangeColumnPosition != nil {
			buf.WriteString(" ")
			buf.WriteString(a.ChangeColumnPosition.String())
		}
		return buf.String()
	case AlterTableOpModifyColumn:
		var buf strings.Builder
		buf.WriteString("MODIFY COLUMN ")
		if a.ModifyColumnName != nil {
			buf.WriteString(a.ModifyColumnName.String())
			buf.WriteString(" ")
		}
		if a.ModifyDataType != nil {
			if dt, ok := a.ModifyDataType.(fmt.Stringer); ok {
				buf.WriteString(dt.String())
			}
		}
		if a.ModifyColumnPosition != nil {
			buf.WriteString(" ")
			buf.WriteString(a.ModifyColumnPosition.String())
		}
		return buf.String()
	case AlterTableOpDisableRowLevelSecurity:
		return "DISABLE ROW LEVEL SECURITY"
	case AlterTableOpEnableRowLevelSecurity:
		return "ENABLE ROW LEVEL SECURITY"
	case AlterTableOpForceRowLevelSecurity:
		return "FORCE ROW LEVEL SECURITY"
	case AlterTableOpNoForceRowLevelSecurity:
		return "NO FORCE ROW LEVEL SECURITY"
	case AlterTableOpDisableTrigger:
		var buf strings.Builder
		buf.WriteString("DISABLE TRIGGER ")
		if a.DisableEnableName != nil {
			buf.WriteString(a.DisableEnableName.String())
		}
		return buf.String()
	case AlterTableOpEnableTrigger:
		var buf strings.Builder
		buf.WriteString("ENABLE TRIGGER ")
		if a.DisableEnableName != nil {
			buf.WriteString(a.DisableEnableName.String())
		}
		return buf.String()
	case AlterTableOpEnableAlwaysTrigger:
		var buf strings.Builder
		buf.WriteString("ENABLE ALWAYS TRIGGER ")
		if a.DisableEnableName != nil {
			buf.WriteString(a.DisableEnableName.String())
		}
		return buf.String()
	case AlterTableOpEnableReplicaTrigger:
		var buf strings.Builder
		buf.WriteString("ENABLE REPLICA TRIGGER ")
		if a.DisableEnableName != nil {
			buf.WriteString(a.DisableEnableName.String())
		}
		return buf.String()
	case AlterTableOpDisableRule:
		var buf strings.Builder
		buf.WriteString("DISABLE RULE ")
		if a.DisableEnableName != nil {
			buf.WriteString(a.DisableEnableName.String())
		}
		return buf.String()
	case AlterTableOpEnableRule:
		var buf strings.Builder
		buf.WriteString("ENABLE RULE ")
		if a.DisableEnableName != nil {
			buf.WriteString(a.DisableEnableName.String())
		}
		return buf.String()
	case AlterTableOpEnableAlwaysRule:
		var buf strings.Builder
		buf.WriteString("ENABLE ALWAYS RULE ")
		if a.DisableEnableName != nil {
			buf.WriteString(a.DisableEnableName.String())
		}
		return buf.String()
	case AlterTableOpEnableReplicaRule:
		var buf strings.Builder
		buf.WriteString("ENABLE REPLICA RULE ")
		if a.DisableEnableName != nil {
			buf.WriteString(a.DisableEnableName.String())
		}
		return buf.String()
	case AlterTableOpReplicaIdentity:
		var buf strings.Builder
		buf.WriteString("REPLICA IDENTITY ")
		buf.WriteString(a.ReplicaIdentity.String())
		if a.ReplicaIdentity == ReplicaIdentityIndex && a.ReplicaIdentityIndex != nil {
			buf.WriteString(" ")
			buf.WriteString(a.ReplicaIdentityIndex.String())
		}
		return buf.String()
	case AlterTableOpValidateConstraint:
		var buf strings.Builder
		buf.WriteString("VALIDATE CONSTRAINT ")
		if a.ValidateConstraintName != nil {
			buf.WriteString(a.ValidateConstraintName.String())
		}
		return buf.String()
	// PostgreSQL OWNER TO
	case AlterTableOpOwnerTo:
		var buf strings.Builder
		buf.WriteString("OWNER TO ")
		if a.NewOwner != nil {
			buf.WriteString(a.NewOwner.String())
		}
		return buf.String()
	// Snowflake-specific clustering operations
	case AlterTableOpDropClusteringKey:
		return "DROP CLUSTERING KEY"
	case AlterTableOpSuspendRecluster:
		return "SUSPEND RECLUSTER"
	case AlterTableOpResumeRecluster:
		return "RESUME RECLUSTER"
	// Snowflake-specific SWAP WITH operation
	case AlterTableOpSwapWith:
		var buf strings.Builder
		buf.WriteString("SWAP WITH ")
		if a.SwapWithTableName != nil {
			buf.WriteString(a.SwapWithTableName.String())
		}
		return buf.String()
	// Snowflake-specific CLUSTER BY operation
	case AlterTableOpClusterBy:
		var buf strings.Builder
		buf.WriteString("CLUSTER BY (")
		for i, expr := range a.ClusterBy {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(expr.String())
		}
		buf.WriteString(")")
		return buf.String()
	// Snowflake-specific dynamic/external table operations
	case AlterTableOpRefresh:
		var buf strings.Builder
		buf.WriteString("REFRESH")
		if a.RefreshSubpath != nil {
			buf.WriteString(" '")
			buf.WriteString(*a.RefreshSubpath)
			buf.WriteString("'")
		}
		return buf.String()
	case AlterTableOpSuspend:
		return "SUSPEND"
	case AlterTableOpResume:
		return "RESUME"
	default:
		return ""
	}
}

// DropBehavior represents DROP behavior.
type DropBehavior int

const (
	DropBehaviorNone DropBehavior = iota
	DropBehaviorRestrict
	DropBehaviorCascade
)

func (d DropBehavior) String() string {
	switch d {
	case DropBehaviorRestrict:
		return "RESTRICT"
	case DropBehaviorCascade:
		return "CASCADE"
	default:
		return ""
	}
}

// MySQLColumnPosition represents MySQL ALTER TABLE column position (FIRST or AFTER column).
type MySQLColumnPosition struct {
	// If true, place column first
	IsFirst bool
	// If not first, the column to place after
	AfterColumn *ast.Ident
}

func (m *MySQLColumnPosition) String() string {
	if m.IsFirst {
		return "FIRST"
	}
	if m.AfterColumn != nil {
		return "AFTER " + m.AfterColumn.String()
	}
	return ""
}

// HiveSetLocation represents Hive SET LOCATION.
type HiveSetLocation struct{}

func (h *HiveSetLocation) exprNode()        {}
func (h *HiveSetLocation) expr()            {}
func (h *HiveSetLocation) IsExpr()          {}
func (h *HiveSetLocation) Span() token.Span { return token.Span{} }
func (h *HiveSetLocation) String() string   { return "" }

// AlterIndexOperation represents ALTER INDEX operation.
type AlterIndexOperation struct {
	// RenameTo is the new name for RENAME TO operation
	RenameTo *ObjectName
}

func (a *AlterIndexOperation) exprNode()        {}
func (a *AlterIndexOperation) expr()            {}
func (a *AlterIndexOperation) IsExpr()          {}
func (a *AlterIndexOperation) Span() token.Span { return token.Span{} }
func (a *AlterIndexOperation) String() string {
	if a == nil || a.RenameTo == nil {
		return ""
	}
	return "RENAME TO " + a.RenameTo.String()
}

// AlterSchemaOperation represents ALTER SCHEMA operation.
// This is an interface that can be one of:
// - *AlterSchemaRenameTo
// - *AlterSchemaOwnerTo
type AlterSchemaOperation interface {
	Expr
	IsAlterSchemaOperation()
}

// AlterSchemaRenameTo represents ALTER SCHEMA RENAME TO operation
type AlterSchemaRenameTo struct {
	NewName *ast.ObjectName
}

func (a *AlterSchemaRenameTo) exprNode()               {}
func (a *AlterSchemaRenameTo) expr()                   {}
func (a *AlterSchemaRenameTo) IsExpr()                 {}
func (a *AlterSchemaRenameTo) Span() token.Span        { return token.Span{} }
func (a *AlterSchemaRenameTo) IsAlterSchemaOperation() {}

func (a *AlterSchemaRenameTo) String() string {
	return "RENAME TO " + a.NewName.String()
}

// AlterSchemaOwnerTo represents ALTER SCHEMA OWNER TO operation
type AlterSchemaOwnerTo struct {
	Owner Owner
}

func (a *AlterSchemaOwnerTo) exprNode()               {}
func (a *AlterSchemaOwnerTo) expr()                   {}
func (a *AlterSchemaOwnerTo) IsExpr()                 {}
func (a *AlterSchemaOwnerTo) Span() token.Span        { return token.Span{} }
func (a *AlterSchemaOwnerTo) IsAlterSchemaOperation() {}

func (a *AlterSchemaOwnerTo) String() string {
	return "OWNER TO " + a.Owner.String()
}

// AlterTypeOperation represents ALTER TYPE operation.
// This is an interface that can be one of:
// - *AlterTypeRename
// - *AlterTypeAddValue
// - *AlterTypeRenameValue
type AlterTypeOperation interface {
	Expr
	IsAlterTypeOperation()
}

// AlterTypeRename represents ALTER TYPE RENAME TO operation
type AlterTypeRename struct {
	NewName *ast.Ident
}

func (a *AlterTypeRename) exprNode()             {}
func (a *AlterTypeRename) expr()                 {}
func (a *AlterTypeRename) IsExpr()               {}
func (a *AlterTypeRename) Span() token.Span      { return token.Span{} }
func (a *AlterTypeRename) IsAlterTypeOperation() {}

func (a *AlterTypeRename) String() string {
	return "RENAME TO " + a.NewName.String()
}

// AlterTypeAddValuePosition represents the position for ADD VALUE
type AlterTypeAddValuePosition interface {
	Expr
	IsAddValuePosition()
}

// AlterTypeAddValuePositionBefore represents BEFORE position
type AlterTypeAddValuePositionBefore struct {
	NeighborValue *ast.Ident
}

func (a *AlterTypeAddValuePositionBefore) exprNode()           {}
func (a *AlterTypeAddValuePositionBefore) expr()               {}
func (a *AlterTypeAddValuePositionBefore) IsExpr()             {}
func (a *AlterTypeAddValuePositionBefore) Span() token.Span    { return token.Span{} }
func (a *AlterTypeAddValuePositionBefore) IsAddValuePosition() {}

func (a *AlterTypeAddValuePositionBefore) String() string {
	return "BEFORE " + a.NeighborValue.String()
}

// AlterTypeAddValuePositionAfter represents AFTER position
type AlterTypeAddValuePositionAfter struct {
	NeighborValue *ast.Ident
}

func (a *AlterTypeAddValuePositionAfter) exprNode()           {}
func (a *AlterTypeAddValuePositionAfter) expr()               {}
func (a *AlterTypeAddValuePositionAfter) IsExpr()             {}
func (a *AlterTypeAddValuePositionAfter) Span() token.Span    { return token.Span{} }
func (a *AlterTypeAddValuePositionAfter) IsAddValuePosition() {}

func (a *AlterTypeAddValuePositionAfter) String() string {
	return "AFTER " + a.NeighborValue.String()
}

// AlterTypeAddValue represents ALTER TYPE ADD VALUE operation
type AlterTypeAddValue struct {
	IfNotExists bool
	Value       *ast.Ident
	Position    AlterTypeAddValuePosition
}

func (a *AlterTypeAddValue) exprNode()             {}
func (a *AlterTypeAddValue) expr()                 {}
func (a *AlterTypeAddValue) IsExpr()               {}
func (a *AlterTypeAddValue) Span() token.Span      { return token.Span{} }
func (a *AlterTypeAddValue) IsAlterTypeOperation() {}

func (a *AlterTypeAddValue) String() string {
	var f strings.Builder
	f.WriteString("ADD VALUE")
	if a.IfNotExists {
		f.WriteString(" IF NOT EXISTS")
	}
	f.WriteString(" ")
	f.WriteString(a.Value.String())
	if a.Position != nil {
		f.WriteString(" ")
		f.WriteString(a.Position.String())
	}
	return f.String()
}

// AlterTypeRenameValue represents ALTER TYPE RENAME VALUE operation
type AlterTypeRenameValue struct {
	From *ast.Ident
	To   *ast.Ident
}

func (a *AlterTypeRenameValue) exprNode()             {}
func (a *AlterTypeRenameValue) expr()                 {}
func (a *AlterTypeRenameValue) IsExpr()               {}
func (a *AlterTypeRenameValue) Span() token.Span      { return token.Span{} }
func (a *AlterTypeRenameValue) IsAlterTypeOperation() {}

func (a *AlterTypeRenameValue) String() string {
	return "RENAME VALUE " + a.From.String() + " TO " + a.To.String()
}

// Password represents a password value for role options.
// Can be either a password expression or NULL.
type Password struct {
	// Value is the password expression (nil for NullPassword)
	Value Expr
	// IsNull is true for PASSWORD NULL
	IsNull bool
}

func (p *Password) String() string {
	if p.IsNull {
		return "PASSWORD NULL"
	}
	if p.Value != nil {
		return "PASSWORD " + p.Value.String()
	}
	return "PASSWORD"
}

// RoleOption represents an option in a CREATE/ALTER ROLE statement.
// Reference: src/ast/dcl.rs RoleOption
// See: https://www.postgresql.org/docs/current/sql-createrole.html
type RoleOption struct {
	// Option type
	Type RoleOptionType
	// Value is used for ConnectionLimit and ValidUntil
	Value Expr
	// BoolValue is used for boolean options
	BoolValue bool
	// Password is used for PASSWORD option
	Password *Password
}

// RoleOptionType represents the type of role option
type RoleOptionType int

const (
	RoleOptionBypassRLS RoleOptionType = iota
	RoleOptionNoBypassRLS
	RoleOptionConnectionLimit
	RoleOptionCreateDB
	RoleOptionNoCreateDB
	RoleOptionCreateRole
	RoleOptionNoCreateRole
	RoleOptionInherit
	RoleOptionNoInherit
	RoleOptionLogin
	RoleOptionNoLogin
	RoleOptionPassword
	RoleOptionReplication
	RoleOptionNoReplication
	RoleOptionSuperUser
	RoleOptionNoSuperUser
	RoleOptionValidUntil
)

func (r *RoleOption) String() string {
	switch r.Type {
	case RoleOptionBypassRLS:
		return "BYPASSRLS"
	case RoleOptionNoBypassRLS:
		return "NOBYPASSRLS"
	case RoleOptionConnectionLimit:
		if r.Value != nil {
			return "CONNECTION LIMIT " + r.Value.String()
		}
		return "CONNECTION LIMIT"
	case RoleOptionCreateDB:
		return "CREATEDB"
	case RoleOptionNoCreateDB:
		return "NOCREATEDB"
	case RoleOptionCreateRole:
		return "CREATEROLE"
	case RoleOptionNoCreateRole:
		return "NOCREATEROLE"
	case RoleOptionInherit:
		return "INHERIT"
	case RoleOptionNoInherit:
		return "NOINHERIT"
	case RoleOptionLogin:
		return "LOGIN"
	case RoleOptionNoLogin:
		return "NOLOGIN"
	case RoleOptionPassword:
		if r.Password != nil {
			return r.Password.String()
		}
		return "PASSWORD"
	case RoleOptionReplication:
		return "REPLICATION"
	case RoleOptionNoReplication:
		return "NOREPLICATION"
	case RoleOptionSuperUser:
		return "SUPERUSER"
	case RoleOptionNoSuperUser:
		return "NOSUPERUSER"
	case RoleOptionValidUntil:
		if r.Value != nil {
			return "VALID UNTIL " + r.Value.String()
		}
		return "VALID UNTIL"
	}
	return ""
}

// SetConfigValue represents a value for SET configuration_parameter.
// Reference: src/ast/dcl.rs SetConfigValue
// SET config_name { TO | = } { value | DEFAULT }
// SET config_name FROM CURRENT
type SetConfigValue struct {
	// Type determines the kind of value
	Type SetConfigValueType
	// Value is the expression value (for Value type)
	Value Expr
}

// SetConfigValueType represents the type of SET config value
type SetConfigValueType int

const (
	SetConfigValueDefault SetConfigValueType = iota
	SetConfigValueFromCurrent
	SetConfigValueExpr
)

func (s *SetConfigValue) String() string {
	switch s.Type {
	case SetConfigValueDefault:
		return "DEFAULT"
	case SetConfigValueFromCurrent:
		return "FROM CURRENT"
	case SetConfigValueExpr:
		if s.Value != nil {
			return s.Value.String()
		}
	}
	return ""
}

// ResetConfig represents a RESET configuration option.
// Reference: src/ast/dcl.rs ResetConfig
// RESET config_name | RESET ALL
type ResetConfig struct {
	// IsAll is true for RESET ALL
	IsAll bool
	// ConfigName is the name to reset (for non-ALL case)
	ConfigName *ast.ObjectName
}

func (r *ResetConfig) String() string {
	if r.IsAll {
		return "ALL"
	}
	if r.ConfigName != nil {
		return r.ConfigName.String()
	}
	return ""
}

// AlterRoleOperation is an interface for ALTER ROLE operations.
// Reference: src/ast/dcl.rs AlterRoleOperation
type AlterRoleOperation interface {
	Expr
	isAlterRoleOperation()
}

// AlterRoleOperationRenameRole represents RENAME TO new_role_name.
type AlterRoleOperationRenameRole struct {
	RoleName *ast.Ident
}

func (a *AlterRoleOperationRenameRole) exprNode()             {}
func (a *AlterRoleOperationRenameRole) expr()                 {}
func (a *AlterRoleOperationRenameRole) IsExpr()               {}
func (a *AlterRoleOperationRenameRole) Span() token.Span      { return token.Span{} }
func (a *AlterRoleOperationRenameRole) String() string        { return "RENAME TO " + a.RoleName.String() }
func (a *AlterRoleOperationRenameRole) isAlterRoleOperation() {}

// AlterRoleOperationAddMember represents ADD MEMBER member_name (MSSQL).
type AlterRoleOperationAddMember struct {
	MemberName *ast.Ident
}

func (a *AlterRoleOperationAddMember) exprNode()             {}
func (a *AlterRoleOperationAddMember) expr()                 {}
func (a *AlterRoleOperationAddMember) IsExpr()               {}
func (a *AlterRoleOperationAddMember) Span() token.Span      { return token.Span{} }
func (a *AlterRoleOperationAddMember) String() string        { return "ADD MEMBER " + a.MemberName.String() }
func (a *AlterRoleOperationAddMember) isAlterRoleOperation() {}

// AlterRoleOperationDropMember represents DROP MEMBER member_name (MSSQL).
type AlterRoleOperationDropMember struct {
	MemberName *ast.Ident
}

func (a *AlterRoleOperationDropMember) exprNode()             {}
func (a *AlterRoleOperationDropMember) expr()                 {}
func (a *AlterRoleOperationDropMember) IsExpr()               {}
func (a *AlterRoleOperationDropMember) Span() token.Span      { return token.Span{} }
func (a *AlterRoleOperationDropMember) String() string        { return "DROP MEMBER " + a.MemberName.String() }
func (a *AlterRoleOperationDropMember) isAlterRoleOperation() {}

// AlterRoleOperationWithOptions represents WITH option [ ... ] (PostgreSQL).
type AlterRoleOperationWithOptions struct {
	Options []*RoleOption
}

func (a *AlterRoleOperationWithOptions) exprNode()        {}
func (a *AlterRoleOperationWithOptions) expr()            {}
func (a *AlterRoleOperationWithOptions) IsExpr()          {}
func (a *AlterRoleOperationWithOptions) Span() token.Span { return token.Span{} }
func (a *AlterRoleOperationWithOptions) String() string {
	var parts []string
	for _, opt := range a.Options {
		parts = append(parts, opt.String())
	}
	return "WITH " + strings.Join(parts, " ")
}
func (a *AlterRoleOperationWithOptions) isAlterRoleOperation() {}

// AlterRoleOperationSet represents SET configuration_parameter (PostgreSQL).
// SET config_name { TO | = } { value | DEFAULT }
// SET config_name FROM CURRENT
type AlterRoleOperationSet struct {
	// InDatabase is the optional database name for IN DATABASE clause
	InDatabase *ast.ObjectName
	// ConfigName is the configuration parameter name
	ConfigName *ast.ObjectName
	// ConfigValue is the value to set
	ConfigValue *SetConfigValue
	// UseEqual is true if = was used instead of TO (affects canonical form)
	UseEqual bool
}

func (a *AlterRoleOperationSet) exprNode()        {}
func (a *AlterRoleOperationSet) expr()            {}
func (a *AlterRoleOperationSet) IsExpr()          {}
func (a *AlterRoleOperationSet) Span() token.Span { return token.Span{} }
func (a *AlterRoleOperationSet) String() string {
	var f strings.Builder
	if a.InDatabase != nil {
		f.WriteString("IN DATABASE ")
		f.WriteString(a.InDatabase.String())
		f.WriteString(" ")
	}
	f.WriteString("SET ")
	f.WriteString(a.ConfigName.String())
	// The format depends on the value type:
	// FROM CURRENT: SET config_name FROM CURRENT
	// DEFAULT: SET config_name TO DEFAULT
	// Expression: SET config_name {TO|=} value
	switch a.ConfigValue.Type {
	case SetConfigValueFromCurrent:
		f.WriteString(" FROM CURRENT")
	case SetConfigValueDefault:
		f.WriteString(" TO DEFAULT")
	case SetConfigValueExpr:
		if a.UseEqual {
			f.WriteString(" = ")
		} else {
			f.WriteString(" TO ")
		}
		f.WriteString(a.ConfigValue.Value.String())
	}
	return f.String()
}
func (a *AlterRoleOperationSet) isAlterRoleOperation() {}

// AlterRoleOperationReset represents RESET configuration_parameter (PostgreSQL).
// RESET config_name | RESET ALL
type AlterRoleOperationReset struct {
	// InDatabase is the optional database name for IN DATABASE clause
	InDatabase *ast.ObjectName
	// ConfigName is the configuration to reset
	ConfigName *ResetConfig
}

func (a *AlterRoleOperationReset) exprNode()        {}
func (a *AlterRoleOperationReset) expr()            {}
func (a *AlterRoleOperationReset) IsExpr()          {}
func (a *AlterRoleOperationReset) Span() token.Span { return token.Span{} }
func (a *AlterRoleOperationReset) String() string {
	var f strings.Builder
	if a.InDatabase != nil {
		f.WriteString("IN DATABASE ")
		f.WriteString(a.InDatabase.String())
		f.WriteString(" ")
	}
	f.WriteString("RESET ")
	f.WriteString(a.ConfigName.String())
	return f.String()
}
func (a *AlterRoleOperationReset) isAlterRoleOperation() {}

// ObjectType represents object type for DROP statements.
// Reference: src/ast/mod.rs ObjectType
type ObjectType int

const (
	ObjectTypeNone ObjectType = iota
	ObjectTypeTable
	ObjectTypeView
	ObjectTypeMaterializedView
	ObjectTypeIndex
	ObjectTypeSchema
	ObjectTypeDatabase
	ObjectTypeRole
	ObjectTypeSequence
	ObjectTypeStage
	ObjectTypeType
	ObjectTypeUser
	ObjectTypeStream
)

func (o ObjectType) String() string {
	switch o {
	case ObjectTypeTable:
		return "TABLE"
	case ObjectTypeView:
		return "VIEW"
	case ObjectTypeMaterializedView:
		return "MATERIALIZED VIEW"
	case ObjectTypeIndex:
		return "INDEX"
	case ObjectTypeSchema:
		return "SCHEMA"
	case ObjectTypeDatabase:
		return "DATABASE"
	case ObjectTypeRole:
		return "ROLE"
	case ObjectTypeSequence:
		return "SEQUENCE"
	case ObjectTypeStage:
		return "STAGE"
	case ObjectTypeType:
		return "TYPE"
	case ObjectTypeUser:
		return "USER"
	case ObjectTypeStream:
		return "STREAM"
	default:
		return ""
	}
}

// SchemaName represents schema name variants.
// Supports PostgreSQL-style AUTHORIZATION syntax.
type SchemaName struct {
	// For Simple variant: just the schema name
	Name *ast.ObjectName
	// For UnnamedAuthorization variant: AUTHORIZATION <user>
	Authorization *ast.Ident
	// For NamedAuthorization variant: both name and authorization
	HasAuthorization bool
}

func (s *SchemaName) exprNode()        {}
func (s *SchemaName) expr()            {}
func (s *SchemaName) IsExpr()          {}
func (s *SchemaName) Span() token.Span { return token.Span{} }

func (s *SchemaName) String() string {
	if s.Authorization != nil {
		if s.Name != nil {
			// NamedAuthorization: <name> AUTHORIZATION <user>
			return s.Name.String() + " AUTHORIZATION " + s.Authorization.String()
		}
		// UnnamedAuthorization: AUTHORIZATION <user>
		return "AUTHORIZATION " + s.Authorization.String()
	}
	// Simple: just the name
	if s.Name != nil {
		return s.Name.String()
	}
	return ""
}

// CatalogSyncNamespaceMode represents catalog sync namespace mode.
type CatalogSyncNamespaceMode int

const (
	CatalogSyncNamespaceModeNone CatalogSyncNamespaceMode = iota
)

func (c CatalogSyncNamespaceMode) String() string { return "" }

// ContactEntry represents contact entry.
type ContactEntry struct{}

func (c *ContactEntry) exprNode()        {}
func (c *ContactEntry) expr()            {}
func (c *ContactEntry) IsExpr()          {}
func (c *ContactEntry) Span() token.Span { return token.Span{} }
func (c *ContactEntry) String() string   { return "" }

// SequenceOptionsType represents the type of sequence option.
type SequenceOptionsType int

const (
	SeqOptIncrementBy SequenceOptionsType = iota
	SeqOptMinValue
	SeqOptMaxValue
	SeqOptStartWith
	SeqOptCache
	SeqOptCycle
)

// SequenceOptions represents sequence options.
type SequenceOptions struct {
	SpanVal token.Span
	Type    SequenceOptionsType
	// For IncrementBy, MinValue, MaxValue, StartWith, Cache: the expression value
	Expr Expr
	// For IncrementBy and StartWith: whether the BY/WITH keyword was used
	HasByOrWith bool
	// For Cycle: true means NO CYCLE, false means CYCLE
	NoCycle bool
	// For MinValue and MaxValue: true means NO MINVALUE/NO MAXVALUE
	NoValue bool
}

func (s *SequenceOptions) exprNode() {}
func (s *SequenceOptions) expr()     {}

// String returns the string representation without leading space
// This is used for sequence options in various contexts (CREATE SEQUENCE, GENERATED AS IDENTITY)
func (s *SequenceOptions) String() string {
	switch s.Type {
	case SeqOptIncrementBy:
		if s.HasByOrWith {
			return fmt.Sprintf("INCREMENT BY %s", s.Expr.String())
		}
		return fmt.Sprintf("INCREMENT %s", s.Expr.String())
	case SeqOptMinValue:
		if s.NoValue {
			return "NO MINVALUE"
		}
		return fmt.Sprintf("MINVALUE %s", s.Expr.String())
	case SeqOptMaxValue:
		if s.NoValue {
			return "NO MAXVALUE"
		}
		return fmt.Sprintf("MAXVALUE %s", s.Expr.String())
	case SeqOptStartWith:
		if s.HasByOrWith {
			return fmt.Sprintf("START WITH %s", s.Expr.String())
		}
		return fmt.Sprintf("START %s", s.Expr.String())
	case SeqOptCache:
		return fmt.Sprintf("CACHE %s", s.Expr.String())
	case SeqOptCycle:
		if s.NoCycle {
			return "NO CYCLE"
		}
		return "CYCLE"
	}
	return ""
}

// StringWithLeadingSpace returns the string with a leading space for use in contexts
// where the option needs to be separated from the previous text
func (s *SequenceOptions) StringWithLeadingSpace() string {
	str := s.String()
	if str != "" {
		return " " + str
	}
	return ""
}

// DomainConstraint represents a domain constraint (NOT NULL, CHECK, etc.)
type DomainConstraint struct {
	SpanVal   token.Span
	Name      *ast.Ident // Optional constraint name
	Type      DomainConstraintType
	CheckExpr Expr            // For CHECK constraints
	NullFlag  bool            // For NOT NULL (true = NOT NULL, false = NULL)
	Collation *ast.ObjectName // For COLLATE
}

// DomainConstraintType represents the type of domain constraint
type DomainConstraintType int

const (
	DomainConstraintNone DomainConstraintType = iota
	DomainConstraintNotNull
	DomainConstraintNull
	DomainConstraintCheck
	DomainConstraintCollation
)

func (d *DomainConstraint) exprNode()        {}
func (d *DomainConstraint) expr()            {}
func (d *DomainConstraint) IsExpr()          {}
func (d *DomainConstraint) Span() token.Span { return d.SpanVal }
func (d *DomainConstraint) String() string {
	var sb strings.Builder
	if d.Name != nil {
		sb.WriteString("CONSTRAINT ")
		sb.WriteString(d.Name.String())
		sb.WriteString(" ")
	}
	switch d.Type {
	case DomainConstraintNotNull:
		sb.WriteString("NOT NULL")
	case DomainConstraintNull:
		sb.WriteString("NULL")
	case DomainConstraintCheck:
		sb.WriteString("CHECK (")
		if d.CheckExpr != nil {
			sb.WriteString(d.CheckExpr.String())
		}
		sb.WriteString(")")
	case DomainConstraintCollation:
		sb.WriteString("COLLATE ")
		if d.Collation != nil {
			sb.WriteString(d.Collation.String())
		}
	}
	return sb.String()
}

// ============================================================================
// Identity Property Types (for IDENTITY/AUTOINCREMENT columns)
// Reference: src/ast/ddl.rs - IdentityProperty, IdentityPropertyKind, etc.
// ============================================================================

// IdentityPropertyKind represents the kind of identity property (IDENTITY or AUTOINCREMENT)
type IdentityPropertyKind int

const (
	IdentityPropertyKindIdentity IdentityPropertyKind = iota
	IdentityPropertyKindAutoincrement
)

func (i IdentityPropertyKind) String() string {
	switch i {
	case IdentityPropertyKindIdentity:
		return "IDENTITY"
	case IdentityPropertyKindAutoincrement:
		// Rust canonical form is AUTOINCREMENT (no underscore) for Snowflake/SQLite
		// MySQL's AUTO_INCREMENT is handled as dialect-specific option
		return "AUTOINCREMENT"
	}
	return ""
}

// IdentityPropertyFormatKind represents the format of identity parameters
type IdentityPropertyFormatKind int

const (
	IdentityFormatNone              IdentityPropertyFormatKind = iota
	IdentityFormatFunctionCall                                 // (seed, increment)
	IdentityFormatStartAndIncrement                            // START num INCREMENT num
)

// IdentityParameters represents seed and increment for identity columns
type IdentityParameters struct {
	SpanVal   token.Span
	Seed      Expr
	Increment Expr
}

func (i *IdentityParameters) exprNode()        {}
func (i *IdentityParameters) expr()            {}
func (i *IdentityParameters) IsExpr()          {}
func (i *IdentityParameters) Span() token.Span { return i.SpanVal }
func (i *IdentityParameters) String() string {
	return fmt.Sprintf("%s, %s", i.Seed.String(), i.Increment.String())
}

// IdentityPropertyOrder represents ORDER or NOORDER for identity columns
type IdentityPropertyOrder int

const (
	IdentityOrderNone IdentityPropertyOrder = iota
	IdentityOrderOrder
	IdentityOrderNoOrder
)

func (i IdentityPropertyOrder) String() string {
	switch i {
	case IdentityOrderOrder:
		return " ORDER"
	case IdentityOrderNoOrder:
		return " NOORDER"
	}
	return ""
}

// IdentityProperty represents the IDENTITY/AUTOINCREMENT property for a column
type IdentityProperty struct {
	SpanVal    token.Span
	Parameters *IdentityParameters
	Order      IdentityPropertyOrder
	Format     IdentityPropertyFormatKind // How parameters are formatted
}

func (i *IdentityProperty) exprNode()        {}
func (i *IdentityProperty) expr()            {}
func (i *IdentityProperty) IsExpr()          {}
func (i *IdentityProperty) Span() token.Span { return i.SpanVal }
func (i *IdentityProperty) String() string {
	var sb strings.Builder
	if i.Parameters != nil {
		switch i.Format {
		case IdentityFormatFunctionCall:
			sb.WriteString("(")
			sb.WriteString(i.Parameters.String())
			sb.WriteString(")")
		case IdentityFormatStartAndIncrement:
			sb.WriteString(" START ")
			sb.WriteString(i.Parameters.Seed.String())
			sb.WriteString(" INCREMENT ")
			sb.WriteString(i.Parameters.Increment.String())
		}
	}
	sb.WriteString(i.Order.String())
	return sb.String()
}

// ColumnIdentity represents a complete IDENTITY or AUTOINCREMENT column option
type ColumnIdentity struct {
	SpanVal          token.Span
	Kind             IdentityPropertyKind
	Property         *IdentityProperty
	HasAutoIncrement bool // true if original keyword was AUTO_INCREMENT (with underscore, MySQL style)
}

func (c *ColumnIdentity) exprNode()        {}
func (c *ColumnIdentity) expr()            {}
func (c *ColumnIdentity) IsExpr()          {}
func (c *ColumnIdentity) Span() token.Span { return c.SpanVal }
func (c *ColumnIdentity) String() string {
	var sb strings.Builder
	// Use AUTO_INCREMENT (with underscore) for MySQL style, otherwise use Kind.String()
	if c.HasAutoIncrement {
		sb.WriteString("AUTO_INCREMENT")
	} else {
		sb.WriteString(c.Kind.String())
	}
	if c.Property != nil {
		sb.WriteString(c.Property.String())
	}
	return sb.String()
}

// ============================================================================
// GeneratedAs enum for PostgreSQL GENERATED ALWAYS/BY DEFAULT AS IDENTITY
// Reference: src/ast/ddl.rs - GeneratedAs
// ============================================================================

// GeneratedAs represents the type of generation for identity columns
type GeneratedAs int

const (
	GeneratedAsNone      GeneratedAs = iota
	GeneratedAsAlways                // GENERATED ALWAYS
	GeneratedAsByDefault             // GENERATED BY DEFAULT
	GeneratedAsExpStored             // Expression-based (STORED)
)

func (g GeneratedAs) String() string {
	switch g {
	case GeneratedAsAlways:
		return "ALWAYS"
	case GeneratedAsByDefault:
		return "BY DEFAULT"
	case GeneratedAsExpStored:
		return "EXP STORED"
	}
	return ""
}

// GeneratedIdentity represents GENERATED { ALWAYS | BY DEFAULT } AS IDENTITY [ (sequence_options) ]
type GeneratedIdentity struct {
	SpanVal         token.Span
	GeneratedAs     GeneratedAs
	SequenceOptions []*SequenceOptions
}

func (g *GeneratedIdentity) exprNode()        {}
func (g *GeneratedIdentity) expr()            {}
func (g *GeneratedIdentity) IsExpr()          {}
func (g *GeneratedIdentity) Span() token.Span { return g.SpanVal }
func (g *GeneratedIdentity) String() string {
	var sb strings.Builder
	sb.WriteString("GENERATED ")
	sb.WriteString(g.GeneratedAs.String())
	sb.WriteString(" AS IDENTITY")
	if len(g.SequenceOptions) > 0 {
		sb.WriteString(" (")
		for i, opt := range g.SequenceOptions {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(strings.TrimSpace(opt.String()))
		}
		sb.WriteString(")")
	}
	return sb.String()
}

// ============================================================================
// User-Defined Type Representation
// ============================================================================

// UserDefinedTypeRepresentation is an interface for different type representations
type UserDefinedTypeRepresentation interface {
	Expr
	isUserDefinedTypeRepresentation()
}

// UserDefinedTypeEnum represents CREATE TYPE ... AS ENUM (...)
type UserDefinedTypeEnum struct {
	SpanVal token.Span
	Labels  []*ast.Ident
}

func (u *UserDefinedTypeEnum) exprNode()                        {}
func (u *UserDefinedTypeEnum) expr()                            {}
func (u *UserDefinedTypeEnum) IsExpr()                          {}
func (u *UserDefinedTypeEnum) Span() token.Span                 { return u.SpanVal }
func (u *UserDefinedTypeEnum) isUserDefinedTypeRepresentation() {}
func (u *UserDefinedTypeEnum) String() string {
	var sb strings.Builder
	sb.WriteString("AS ENUM (")
	for i, label := range u.Labels {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(label.String())
	}
	sb.WriteString(")")
	return sb.String()
}

// UserDefinedTypeCompositeAttributeDef represents an attribute in a composite type
type UserDefinedTypeCompositeAttributeDef struct {
	SpanVal   token.Span
	Name      *ast.Ident
	DataType  interface{} // datatype.DataType
	Collation *ast.ObjectName
}

func (u *UserDefinedTypeCompositeAttributeDef) exprNode()        {}
func (u *UserDefinedTypeCompositeAttributeDef) expr()            {}
func (u *UserDefinedTypeCompositeAttributeDef) IsExpr()          {}
func (u *UserDefinedTypeCompositeAttributeDef) Span() token.Span { return u.SpanVal }
func (u *UserDefinedTypeCompositeAttributeDef) String() string {
	var sb strings.Builder
	sb.WriteString(u.Name.String())
	sb.WriteString(" ")
	if u.DataType != nil {
		if dt, ok := u.DataType.(fmt.Stringer); ok {
			sb.WriteString(dt.String())
		}
	}
	if u.Collation != nil {
		sb.WriteString(" COLLATE ")
		sb.WriteString(u.Collation.String())
	}
	return sb.String()
}

// UserDefinedTypeComposite represents CREATE TYPE ... AS (...)
type UserDefinedTypeComposite struct {
	SpanVal    token.Span
	Attributes []*UserDefinedTypeCompositeAttributeDef
}

func (u *UserDefinedTypeComposite) exprNode()                        {}
func (u *UserDefinedTypeComposite) expr()                            {}
func (u *UserDefinedTypeComposite) IsExpr()                          {}
func (u *UserDefinedTypeComposite) Span() token.Span                 { return u.SpanVal }
func (u *UserDefinedTypeComposite) isUserDefinedTypeRepresentation() {}
func (u *UserDefinedTypeComposite) String() string {
	var sb strings.Builder
	sb.WriteString("AS (")
	for i, attr := range u.Attributes {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(attr.String())
	}
	sb.WriteString(")")
	return sb.String()
}

// UserDefinedTypeRange represents CREATE TYPE ... AS RANGE (...)
type UserDefinedTypeRange struct {
	SpanVal token.Span
	// Simplified - just store as string for now
	Options []string
}

func (u *UserDefinedTypeRange) exprNode()                        {}
func (u *UserDefinedTypeRange) expr()                            {}
func (u *UserDefinedTypeRange) IsExpr()                          {}
func (u *UserDefinedTypeRange) Span() token.Span                 { return u.SpanVal }
func (u *UserDefinedTypeRange) isUserDefinedTypeRepresentation() {}
func (u *UserDefinedTypeRange) String() string {
	var sb strings.Builder
	sb.WriteString("AS RANGE (")
	for i, opt := range u.Options {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(opt)
	}
	sb.WriteString(")")
	return sb.String()
}

// UserDefinedTypeSqlDefinition represents CREATE TYPE ... (...)
type UserDefinedTypeSqlDefinition struct {
	SpanVal token.Span
	// Simplified - just store as string for now
	Options []string
}

func (u *UserDefinedTypeSqlDefinition) exprNode()                        {}
func (u *UserDefinedTypeSqlDefinition) expr()                            {}
func (u *UserDefinedTypeSqlDefinition) IsExpr()                          {}
func (u *UserDefinedTypeSqlDefinition) Span() token.Span                 { return u.SpanVal }
func (u *UserDefinedTypeSqlDefinition) isUserDefinedTypeRepresentation() {}
func (u *UserDefinedTypeSqlDefinition) String() string {
	var sb strings.Builder
	sb.WriteString("(")
	for i, opt := range u.Options {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(opt)
	}
	sb.WriteString(")")
	return sb.String()
}

// TriggerPeriod represents the trigger timing (BEFORE, AFTER, INSTEAD OF, FOR).
type TriggerPeriod int

const (
	TriggerPeriodNone TriggerPeriod = iota
	TriggerPeriodFor
	TriggerPeriodAfter
	TriggerPeriodBefore
	TriggerPeriodInsteadOf
)

func (t TriggerPeriod) String() string {
	switch t {
	case TriggerPeriodFor:
		return "FOR"
	case TriggerPeriodAfter:
		return "AFTER"
	case TriggerPeriodBefore:
		return "BEFORE"
	case TriggerPeriodInsteadOf:
		return "INSTEAD OF"
	default:
		return ""
	}
}

// TriggerEvent represents trigger event (INSERT, UPDATE, DELETE, TRUNCATE).
type TriggerEvent int

const (
	TriggerEventNone TriggerEvent = iota
	TriggerEventInsert
	TriggerEventUpdate
	TriggerEventDelete
	TriggerEventTruncate
)

// TriggerEventWithColumns represents UPDATE event with column list.
type TriggerEventWithColumns struct {
	Event   TriggerEvent
	Columns []*ast.Ident
}

func (t TriggerEvent) String() string {
	switch t {
	case TriggerEventInsert:
		return "INSERT"
	case TriggerEventUpdate:
		return "UPDATE"
	case TriggerEventDelete:
		return "DELETE"
	case TriggerEventTruncate:
		return "TRUNCATE"
	default:
		return ""
	}
}

// TriggerReferencingType represents the type of trigger referencing (OLD TABLE, NEW TABLE).
type TriggerReferencingType int

const (
	TriggerReferencingTypeNone TriggerReferencingType = iota
	TriggerReferencingTypeOldTable
	TriggerReferencingTypeNewTable
)

func (t TriggerReferencingType) String() string {
	switch t {
	case TriggerReferencingTypeOldTable:
		return "OLD TABLE"
	case TriggerReferencingTypeNewTable:
		return "NEW TABLE"
	default:
		return ""
	}
}

// TriggerReferencing represents trigger referencing clause.
type TriggerReferencing struct {
	ReferType              TriggerReferencingType
	IsAs                   bool
	TransitionRelationName *ast.ObjectName
}

func (t *TriggerReferencing) exprNode()        {}
func (t *TriggerReferencing) expr()            {}
func (t *TriggerReferencing) IsExpr()          {}
func (t *TriggerReferencing) Span() token.Span { return token.Span{} }
func (t *TriggerReferencing) String() string {
	var sb strings.Builder
	sb.WriteString(t.ReferType.String())
	if t.IsAs {
		sb.WriteString(" AS")
	}
	if t.TransitionRelationName != nil {
		sb.WriteString(" ")
		sb.WriteString(t.TransitionRelationName.String())
	}
	return sb.String()
}

// TriggerObject represents whether trigger fires per row or per statement.
type TriggerObject int

const (
	TriggerObjectNone TriggerObject = iota
	TriggerObjectRow
	TriggerObjectStatement
)

func (t TriggerObject) String() string {
	switch t {
	case TriggerObjectRow:
		return "ROW"
	case TriggerObjectStatement:
		return "STATEMENT"
	default:
		return ""
	}
}

// TriggerObjectKind represents FOR ROW/STATEMENT or FOR EACH ROW/STATEMENT.
type TriggerObjectKind int

const (
	TriggerObjectKindNone TriggerObjectKind = iota
	TriggerObjectKindFor
	TriggerObjectKindForEach
)

type TriggerObjectKindWithObject struct {
	Kind   TriggerObjectKind
	Object TriggerObject
}

func (t TriggerObjectKindWithObject) String() string {
	var sb strings.Builder
	if t.Kind == TriggerObjectKindForEach {
		sb.WriteString("FOR EACH ")
	} else {
		sb.WriteString("FOR ")
	}
	sb.WriteString(t.Object.String())
	return sb.String()
}

// TriggerExecBodyType represents the type of trigger execution body.
type TriggerExecBodyType int

const (
	TriggerExecBodyTypeNone TriggerExecBodyType = iota
	TriggerExecBodyTypeFunction
	TriggerExecBodyTypeProcedure
)

func (t TriggerExecBodyType) String() string {
	switch t {
	case TriggerExecBodyTypeFunction:
		return "FUNCTION"
	case TriggerExecBodyTypeProcedure:
		return "PROCEDURE"
	default:
		return ""
	}
}

// FunctionDesc represents a function description.
type FunctionDesc struct {
	Name *ast.ObjectName
	Args []Expr
}

func (f *FunctionDesc) exprNode()        {}
func (f *FunctionDesc) expr()            {}
func (f *FunctionDesc) IsExpr()          {}
func (f *FunctionDesc) Span() token.Span { return token.Span{} }
func (f *FunctionDesc) String() string {
	var sb strings.Builder
	sb.WriteString(f.Name.String())
	// Only add () if there are arguments
	if len(f.Args) > 0 {
		sb.WriteString("(")
		for i, arg := range f.Args {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(arg.String())
		}
		sb.WriteString(")")
	}
	return sb.String()
}

// TriggerExecBody represents trigger execution body.
type TriggerExecBody struct {
	ExecType TriggerExecBodyType
	FuncDesc *FunctionDesc
}

func (t *TriggerExecBody) exprNode()        {}
func (t *TriggerExecBody) expr()            {}
func (t *TriggerExecBody) IsExpr()          {}
func (t *TriggerExecBody) Span() token.Span { return token.Span{} }
func (t *TriggerExecBody) String() string {
	var sb strings.Builder
	sb.WriteString(t.ExecType.String())
	sb.WriteString(" ")
	sb.WriteString(t.FuncDesc.String())
	return sb.String()
}

// ConditionalStatements represents conditional statements (either a BEGIN/END block or a sequence).
type ConditionalStatements struct {
	IsBeginEnd bool            // true if wrapped in BEGIN/END
	Statements []ast.Statement // the statements inside
}

func (c *ConditionalStatements) exprNode()        {}
func (c *ConditionalStatements) expr()            {}
func (c *ConditionalStatements) IsExpr()          {}
func (c *ConditionalStatements) Span() token.Span { return token.Span{} }

func (c *ConditionalStatements) String() string {
	if c == nil || len(c.Statements) == 0 {
		return ""
	}
	var sb strings.Builder
	if c.IsBeginEnd {
		sb.WriteString("BEGIN ")
	}
	for i, stmt := range c.Statements {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(stmt.String())
		// Add semicolon after each statement in a compound block
		if c.IsBeginEnd {
			sb.WriteString(";")
		}
	}
	if c.IsBeginEnd {
		sb.WriteString(" END")
	}
	return sb.String()
}

// MacroArg represents macro argument.
type MacroArg struct{}

func (m *MacroArg) exprNode()        {}
func (m *MacroArg) expr()            {}
func (m *MacroArg) IsExpr()          {}
func (m *MacroArg) Span() token.Span { return token.Span{} }
func (m *MacroArg) String() string   { return "" }

// MacroDefinition represents macro definition.
type MacroDefinition struct{}

func (m *MacroDefinition) exprNode()        {}
func (m *MacroDefinition) expr()            {}
func (m *MacroDefinition) IsExpr()          {}
func (m *MacroDefinition) Span() token.Span { return token.Span{} }
func (m *MacroDefinition) String() string   { return "" }

// StageParamsObject represents stage parameters object.
type StageParamsObject struct {
	Url                *string
	Encryption         *KeyValueOptions
	Endpoint           *string
	StorageIntegration *string
	Credentials        *KeyValueOptions
}

func (s *StageParamsObject) exprNode()        {}
func (s *StageParamsObject) expr()            {}
func (s *StageParamsObject) IsExpr()          {}
func (s *StageParamsObject) Span() token.Span { return token.Span{} }
func (s *StageParamsObject) String() string {
	var parts []string
	if s.Url != nil {
		parts = append(parts, fmt.Sprintf("URL='%s'", escapeSingleQuote(*s.Url)))
	}
	if s.StorageIntegration != nil {
		parts = append(parts, fmt.Sprintf("STORAGE_INTEGRATION=%s", *s.StorageIntegration))
	}
	if s.Endpoint != nil {
		parts = append(parts, fmt.Sprintf("ENDPOINT='%s'", escapeSingleQuote(*s.Endpoint)))
	}
	if s.Credentials != nil && len(s.Credentials.Options) > 0 {
		parts = append(parts, fmt.Sprintf("CREDENTIALS=(%s)", s.Credentials.String()))
	}
	if s.Encryption != nil && len(s.Encryption.Options) > 0 {
		parts = append(parts, fmt.Sprintf("ENCRYPTION=(%s)", s.Encryption.String()))
	}
	return strings.Join(parts, " ")
}

// KeyValueOptionKind represents the kind of value for a key-value option.
type KeyValueOptionKind int

const (
	KeyValueOptionKindSingle KeyValueOptionKind = iota
	KeyValueOptionKindMulti
	KeyValueOptionKindNested
)

// KeyValueOption represents a single key-value option.
type KeyValueOption struct {
	OptionName  string
	OptionValue interface{} // Can be ast.Value, []ast.Value, or *KeyValueOptions
	Kind        KeyValueOptionKind
	Quoted      bool // Whether the value is a quoted string (needs quotes in output)
}

func (k *KeyValueOption) String() string {
	switch k.Kind {
	case KeyValueOptionKindSingle:
		// Handle nil value (for UNSET statements)
		if k.OptionValue == nil {
			return k.OptionName
		}
		if val, ok := k.OptionValue.(fmt.Stringer); ok {
			return fmt.Sprintf("%s=%s", k.OptionName, val.String())
		}
		if val, ok := k.OptionValue.(string); ok {
			if k.Quoted {
				return fmt.Sprintf("%s='%s'", k.OptionName, escapeSingleQuote(val))
			}
			return fmt.Sprintf("%s=%s", k.OptionName, val)
		}
	case KeyValueOptionKindMulti:
		if vals, ok := k.OptionValue.([]string); ok {
			return fmt.Sprintf("%s=(%s)", k.OptionName, strings.Join(vals, ", "))
		}
	case KeyValueOptionKindNested:
		if opts, ok := k.OptionValue.(*KeyValueOptions); ok {
			return fmt.Sprintf("%s=(%s)", k.OptionName, opts.String())
		}
	}
	if k.OptionValue == nil {
		return k.OptionName
	}
	return fmt.Sprintf("%s=%v", k.OptionName, k.OptionValue)
}

// KeyValueOptionsDelimiter represents the delimiter used between key-value options.
type KeyValueOptionsDelimiter int

const (
	KeyValueOptionsDelimiterSpace KeyValueOptionsDelimiter = iota
	KeyValueOptionsDelimiterComma
)

// KeyValueOptions represents key-value options.
type KeyValueOptions struct {
	Options   []*KeyValueOption
	Delimiter KeyValueOptionsDelimiter
}

func (k *KeyValueOptions) exprNode()        {}
func (k *KeyValueOptions) expr()            {}
func (k *KeyValueOptions) IsExpr()          {}
func (k *KeyValueOptions) Span() token.Span { return token.Span{} }
func (k *KeyValueOptions) String() string {
	var parts []string
	for _, opt := range k.Options {
		parts = append(parts, opt.String())
	}
	sep := " "
	if k.Delimiter == KeyValueOptionsDelimiterComma {
		sep = ", "
	}
	return strings.Join(parts, sep)
}

func escapeSingleQuote(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// SecretOption represents secret option.
type SecretOption struct{}

func (s *SecretOption) exprNode()        {}
func (s *SecretOption) expr()            {}
func (s *SecretOption) IsExpr()          {}
func (s *SecretOption) Span() token.Span { return token.Span{} }
func (s *SecretOption) String() string   { return "" }

// CreatePolicyType represents CREATE POLICY type (PERMISSIVE or RESTRICTIVE).
type CreatePolicyType int

const (
	CreatePolicyTypeNone CreatePolicyType = iota
	CreatePolicyTypePermissive
	CreatePolicyTypeRestrictive
)

func (c CreatePolicyType) String() string {
	switch c {
	case CreatePolicyTypePermissive:
		return "PERMISSIVE"
	case CreatePolicyTypeRestrictive:
		return "RESTRICTIVE"
	default:
		return ""
	}
}

// CreatePolicyCommand represents CREATE POLICY command (FOR clause).
type CreatePolicyCommand int

const (
	CreatePolicyCommandNone CreatePolicyCommand = iota
	CreatePolicyCommandAll
	CreatePolicyCommandSelect
	CreatePolicyCommandInsert
	CreatePolicyCommandUpdate
	CreatePolicyCommandDelete
)

func (c CreatePolicyCommand) String() string {
	switch c {
	case CreatePolicyCommandAll:
		return "ALL"
	case CreatePolicyCommandSelect:
		return "SELECT"
	case CreatePolicyCommandInsert:
		return "INSERT"
	case CreatePolicyCommandUpdate:
		return "UPDATE"
	case CreatePolicyCommandDelete:
		return "DELETE"
	default:
		return ""
	}
}

// RoleName represents role name.
type RoleName struct {
	Name string
}

func (r *RoleName) exprNode()        {}
func (r *RoleName) expr()            {}
func (r *RoleName) IsExpr()          {}
func (r *RoleName) Span() token.Span { return token.Span{} }
func (r *RoleName) String() string   { return r.Name }

// OwnerKind represents the type of owner for ALTER TABLE ... OWNER TO
type OwnerKind int

const (
	OwnerKindIdent OwnerKind = iota
	OwnerKindCurrentRole
	OwnerKindCurrentUser
	OwnerKindSessionUser
)

// Owner represents a new owner specification for ALTER TABLE ... OWNER TO
type Owner struct {
	Kind  OwnerKind
	Ident *ast.Ident // Only set when Kind is OwnerKindIdent
}

func (o *Owner) exprNode()        {}
func (o *Owner) expr()            {}
func (o *Owner) IsExpr()          {}
func (o *Owner) Span() token.Span { return token.Span{} }

func (o *Owner) String() string {
	switch o.Kind {
	case OwnerKindCurrentRole:
		return "CURRENT_ROLE"
	case OwnerKindCurrentUser:
		return "CURRENT_USER"
	case OwnerKindSessionUser:
		return "SESSION_USER"
	default:
		if o.Ident != nil {
			return o.Ident.String()
		}
		return ""
	}
}

// OperatorPurpose represents the purpose of an operator in an operator class.
type OperatorPurpose int

const (
	OperatorPurposeNone OperatorPurpose = iota
	OperatorPurposeForSearch
	OperatorPurposeForOrderBy
)

type OperatorPurposeWithFamily struct {
	Purpose    OperatorPurpose
	SortFamily *ast.ObjectName
}

func (o OperatorPurpose) String() string {
	switch o {
	case OperatorPurposeForSearch:
		return "FOR SEARCH"
	case OperatorPurposeForOrderBy:
		return "FOR ORDER BY"
	default:
		return ""
	}
}

// OperatorOptionKind represents the kind of operator option.
type OperatorOptionKind int

const (
	OperatorOptionKindHashes OperatorOptionKind = iota
	OperatorOptionKindMerges
	OperatorOptionKindCommutator
	OperatorOptionKindNegator
	OperatorOptionKindRestrict
	OperatorOptionKindJoin
)

// OperatorOption represents operator option (COMMUTATOR, NEGATOR, RESTRICT, JOIN, HASHES, MERGES).
type OperatorOption struct {
	Kind OperatorOptionKind
	Name *ast.ObjectName // For Commutator, Negator, Restrict, Join
}

func (o *OperatorOption) exprNode()        {}
func (o *OperatorOption) expr()            {}
func (o *OperatorOption) IsExpr()          {}
func (o *OperatorOption) Span() token.Span { return token.Span{} }
func (o *OperatorOption) String() string {
	switch o.Kind {
	case OperatorOptionKindHashes:
		return "HASHES"
	case OperatorOptionKindMerges:
		return "MERGES"
	case OperatorOptionKindCommutator:
		if o.Name != nil {
			return fmt.Sprintf("COMMUTATOR = %s", o.Name.String())
		}
		return "COMMUTATOR"
	case OperatorOptionKindNegator:
		if o.Name != nil {
			return fmt.Sprintf("NEGATOR = %s", o.Name.String())
		}
		return "NEGATOR"
	case OperatorOptionKindRestrict:
		if o.Name != nil {
			return fmt.Sprintf("RESTRICT = %s", o.Name.String())
		}
		return "RESTRICT = NONE"
	case OperatorOptionKindJoin:
		if o.Name != nil {
			return fmt.Sprintf("JOIN = %s", o.Name.String())
		}
		return "JOIN = NONE"
	default:
		return ""
	}
}

// OperatorArgTypes represents operator argument types for CREATE OPERATOR CLASS.
type OperatorArgTypes struct {
	Left  interface{} // DataType
	Right interface{} // DataType
}

func (o *OperatorArgTypes) exprNode()        {}
func (o *OperatorArgTypes) expr()            {}
func (o *OperatorArgTypes) IsExpr()          {}
func (o *OperatorArgTypes) Span() token.Span { return token.Span{} }
func (o *OperatorArgTypes) String() string {
	return fmt.Sprintf("%v, %v", o.Left, o.Right)
}

// OperatorClassItem represents an item in a CREATE OPERATOR CLASS statement.
type OperatorClassItem struct {
	IsOperator     bool
	IsFunction     bool
	IsStorage      bool
	StrategyNumber uint64
	OperatorName   *ast.ObjectName
	OpTypes        *OperatorArgTypes
	Purpose        *OperatorPurposeWithFamily
	SupportNumber  uint64
	FuncOpTypes    []interface{} // []DataType
	FunctionName   *ast.ObjectName
	ArgumentTypes  []interface{} // []DataType
	StorageType    interface{}   // DataType
}

func (o *OperatorClassItem) exprNode()        {}
func (o *OperatorClassItem) expr()            {}
func (o *OperatorClassItem) IsExpr()          {}
func (o *OperatorClassItem) Span() token.Span { return token.Span{} }
func (o *OperatorClassItem) String() string {
	if o.IsOperator {
		var sb strings.Builder
		sb.WriteString("OPERATOR ")
		sb.WriteString(fmt.Sprintf("%d", o.StrategyNumber))
		sb.WriteString(" ")
		sb.WriteString(o.OperatorName.String())
		if o.OpTypes != nil {
			sb.WriteString(" (")
			sb.WriteString(o.OpTypes.String())
			sb.WriteString(")")
		}
		if o.Purpose != nil {
			sb.WriteString(" ")
			sb.WriteString(o.Purpose.Purpose.String())
			if o.Purpose.Purpose == OperatorPurposeForOrderBy && o.Purpose.SortFamily != nil {
				sb.WriteString(" ")
				sb.WriteString(o.Purpose.SortFamily.String())
			}
		}
		return sb.String()
	}
	if o.IsFunction {
		var sb strings.Builder
		sb.WriteString("FUNCTION ")
		sb.WriteString(fmt.Sprintf("%d", o.SupportNumber))
		sb.WriteString(" ")
		sb.WriteString(o.FunctionName.String())
		if len(o.ArgumentTypes) > 0 {
			sb.WriteString("(")
			for i, t := range o.ArgumentTypes {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%v", t))
			}
			sb.WriteString(")")
		}
		return sb.String()
	}
	if o.IsStorage {
		return fmt.Sprintf("STORAGE %v", o.StorageType)
	}
	return ""
}

// AlterPolicyOperation represents ALTER POLICY operation.
type AlterPolicyOperation struct {
	RenameTo  *ast.Ident // If set, this is a RENAME operation
	To        []*Owner   // Role list for APPLY operation
	Using     Expr       // USING expression for APPLY operation
	WithCheck Expr       // WITH CHECK expression for APPLY operation
}

func (a *AlterPolicyOperation) exprNode()        {}
func (a *AlterPolicyOperation) expr()            {}
func (a *AlterPolicyOperation) IsExpr()          {}
func (a *AlterPolicyOperation) Span() token.Span { return token.Span{} }

func (a *AlterPolicyOperation) String() string {
	if a.RenameTo != nil {
		return "RENAME TO " + a.RenameTo.String()
	}
	var parts []string
	if len(a.To) > 0 {
		var toParts []string
		for _, role := range a.To {
			toParts = append(toParts, role.String())
		}
		parts = append(parts, "TO "+strings.Join(toParts, ", "))
	}
	if a.Using != nil {
		parts = append(parts, "USING ("+a.Using.String()+")")
	}
	if a.WithCheck != nil {
		parts = append(parts, "WITH CHECK ("+a.WithCheck.String()+")")
	}
	return strings.Join(parts, " ")
}

// AlterConnectorOwnerKind represents the kind of ALTER CONNECTOR owner
type AlterConnectorOwnerKind int

const (
	AlterConnectorOwnerKindUser AlterConnectorOwnerKind = iota
	AlterConnectorOwnerKindRole
)

// AlterConnectorOwner represents ALTER CONNECTOR owner.
type AlterConnectorOwner struct {
	Kind AlterConnectorOwnerKind
	Name *ast.Ident
}

func (a *AlterConnectorOwner) exprNode()        {}
func (a *AlterConnectorOwner) expr()            {}
func (a *AlterConnectorOwner) IsExpr()          {}
func (a *AlterConnectorOwner) Span() token.Span { return token.Span{} }

func (a *AlterConnectorOwner) String() string {
	kind := "USER"
	if a.Kind == AlterConnectorOwnerKindRole {
		kind = "ROLE"
	}
	return "OWNER " + kind + " " + a.Name.String()
}

// AttachDuckDBDatabaseOption represents ATTACH DuckDB database option.
type AttachDuckDBDatabaseOption struct{}

func (a *AttachDuckDBDatabaseOption) exprNode()        {}
func (a *AttachDuckDBDatabaseOption) expr()            {}
func (a *AttachDuckDBDatabaseOption) IsExpr()          {}
func (a *AttachDuckDBDatabaseOption) Span() token.Span { return token.Span{} }
func (a *AttachDuckDBDatabaseOption) String() string   { return "" }

// DropOperatorSignature represents DROP OPERATOR signature.
type DropOperatorSignature struct {
	Name     *ast.ObjectName
	ArgTypes []string // Type names as strings
}

func (d *DropOperatorSignature) exprNode()        {}
func (d *DropOperatorSignature) expr()            {}
func (d *DropOperatorSignature) IsExpr()          {}
func (d *DropOperatorSignature) Span() token.Span { return token.Span{} }
func (d *DropOperatorSignature) String() string {
	var f strings.Builder
	f.WriteString(d.Name.String())
	f.WriteString(" (")
	for i, t := range d.ArgTypes {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(t)
	}
	f.WriteString(")")
	return f.String()
}

// OperatorSignature represents operator signature for ALTER OPERATOR.
// Format: name (left_type, right_type) where left_type can be NONE for unary operators
type OperatorSignature struct {
	Name      *ast.ObjectName
	LeftType  string // "NONE" for unary operators, or type name
	RightType string // Type name
}

func (o *OperatorSignature) exprNode()        {}
func (o *OperatorSignature) expr()            {}
func (o *OperatorSignature) IsExpr()          {}
func (o *OperatorSignature) Span() token.Span { return token.Span{} }

func (o *OperatorSignature) String() string {
	var f strings.Builder
	f.WriteString(o.Name.String())
	f.WriteString(" (")
	if o.LeftType != "" {
		f.WriteString(o.LeftType)
	}
	f.WriteString(", ")
	f.WriteString(o.RightType)
	f.WriteString(")")
	return f.String()
}

// AlterOperatorOperation represents ALTER OPERATOR operation.
// This is an interface that can be one of:
// - *AlterOperatorOwnerTo
// - *AlterOperatorSetSchema
// - *AlterOperatorSet
type AlterOperatorOperation interface {
	Expr
	IsAlterOperatorOperation()
}

// AlterOperatorOwnerTo represents ALTER OPERATOR OWNER TO operation
type AlterOperatorOwnerTo struct {
	Owner Owner
}

func (a *AlterOperatorOwnerTo) exprNode()                 {}
func (a *AlterOperatorOwnerTo) expr()                     {}
func (a *AlterOperatorOwnerTo) IsExpr()                   {}
func (a *AlterOperatorOwnerTo) Span() token.Span          { return token.Span{} }
func (a *AlterOperatorOwnerTo) IsAlterOperatorOperation() {}

func (a *AlterOperatorOwnerTo) String() string {
	return "OWNER TO " + a.Owner.String()
}

// AlterOperatorSetSchema represents ALTER OPERATOR SET SCHEMA operation
type AlterOperatorSetSchema struct {
	SchemaName *ast.ObjectName
}

func (a *AlterOperatorSetSchema) exprNode()                 {}
func (a *AlterOperatorSetSchema) expr()                     {}
func (a *AlterOperatorSetSchema) IsExpr()                   {}
func (a *AlterOperatorSetSchema) Span() token.Span          { return token.Span{} }
func (a *AlterOperatorSetSchema) IsAlterOperatorOperation() {}

func (a *AlterOperatorSetSchema) String() string {
	return "SET SCHEMA " + a.SchemaName.String()
}

// AlterOperatorSet represents ALTER OPERATOR SET (options) operation
type AlterOperatorSet struct {
	Options []*OperatorOption
}

func (a *AlterOperatorSet) exprNode()                 {}
func (a *AlterOperatorSet) expr()                     {}
func (a *AlterOperatorSet) IsExpr()                   {}
func (a *AlterOperatorSet) Span() token.Span          { return token.Span{} }
func (a *AlterOperatorSet) IsAlterOperatorOperation() {}

func (a *AlterOperatorSet) String() string {
	var f strings.Builder
	f.WriteString("SET (")
	for i, opt := range a.Options {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(opt.String())
	}
	f.WriteString(")")
	return f.String()
}

// OperatorFamilyOperation represents ALTER OPERATOR FAMILY operation.
// This is an interface that can be one of:
// - *OpFamilyAdd
// - *OpFamilyDrop
// - *OpFamilyRenameTo
// - *OpFamilyOwnerTo
// - *OpFamilySetSchema
type OperatorFamilyOperation interface {
	Expr
	IsOperatorFamilyOperation()
}

// OpFamilyAddItem represents an item to add to operator family
type OpFamilyAddItem struct {
	IsOperator bool // true for OPERATOR, false for FUNCTION
	Number     int  // Operator/function number
	Name       *ast.ObjectName
	ArgTypes   []string // Type names for operator/function
	ForSearch  bool     // FOR SEARCH for operators
	ForOrderBy string   // FOR ORDER BY name for operators
}

// OpFamilyAdd represents ADD { OPERATOR | FUNCTION } items
type OpFamilyAdd struct {
	Items []*OpFamilyAddItem
}

func (o *OpFamilyAdd) exprNode()                  {}
func (o *OpFamilyAdd) expr()                      {}
func (o *OpFamilyAdd) IsExpr()                    {}
func (o *OpFamilyAdd) Span() token.Span           { return token.Span{} }
func (o *OpFamilyAdd) IsOperatorFamilyOperation() {}

func (o *OpFamilyAdd) String() string {
	var f strings.Builder
	f.WriteString("ADD ")
	for i, item := range o.Items {
		if i > 0 {
			f.WriteString(", ")
		}
		if item.IsOperator {
			f.WriteString("OPERATOR ")
		} else {
			f.WriteString("FUNCTION ")
		}
		f.WriteString(fmt.Sprintf("%d ", item.Number))
		f.WriteString(item.Name.String())
		if len(item.ArgTypes) > 0 {
			if item.IsOperator {
				// Space before parenthesis for operators per Rust canonical form
				f.WriteString(" (")
			} else {
				// No space for functions
				f.WriteString("(")
			}
			f.WriteString(strings.Join(item.ArgTypes, ", "))
			f.WriteString(")")
		}
		if item.ForSearch {
			f.WriteString(" FOR SEARCH")
		}
		if item.ForOrderBy != "" {
			f.WriteString(" FOR ORDER BY ")
			f.WriteString(item.ForOrderBy)
		}
	}
	return f.String()
}

// OpFamilyDrop represents DROP { OPERATOR | FUNCTION } items
type OpFamilyDrop struct {
	Items []*OpFamilyAddItem
}

func (o *OpFamilyDrop) exprNode()                  {}
func (o *OpFamilyDrop) expr()                      {}
func (o *OpFamilyDrop) IsExpr()                    {}
func (o *OpFamilyDrop) Span() token.Span           { return token.Span{} }
func (o *OpFamilyDrop) IsOperatorFamilyOperation() {}

func (o *OpFamilyDrop) String() string {
	var f strings.Builder
	f.WriteString("DROP ")
	for i, item := range o.Items {
		if i > 0 {
			f.WriteString(", ")
		}
		if item.IsOperator {
			f.WriteString("OPERATOR ")
		} else {
			f.WriteString("FUNCTION ")
		}
		f.WriteString(fmt.Sprintf("%d ", item.Number))
		if len(item.ArgTypes) > 0 {
			f.WriteString("(")
			f.WriteString(strings.Join(item.ArgTypes, ", "))
			f.WriteString(")")
		}
	}
	return f.String()
}

// OpFamilyRenameTo represents RENAME TO operation
type OpFamilyRenameTo struct {
	NewName *ast.ObjectName
}

func (o *OpFamilyRenameTo) exprNode()                  {}
func (o *OpFamilyRenameTo) expr()                      {}
func (o *OpFamilyRenameTo) IsExpr()                    {}
func (o *OpFamilyRenameTo) Span() token.Span           { return token.Span{} }
func (o *OpFamilyRenameTo) IsOperatorFamilyOperation() {}

func (o *OpFamilyRenameTo) String() string {
	return "RENAME TO " + o.NewName.String()
}

// OpFamilyOwnerTo represents OWNER TO operation
type OpFamilyOwnerTo struct {
	Owner Owner
}

func (o *OpFamilyOwnerTo) exprNode()                  {}
func (o *OpFamilyOwnerTo) expr()                      {}
func (o *OpFamilyOwnerTo) IsExpr()                    {}
func (o *OpFamilyOwnerTo) Span() token.Span           { return token.Span{} }
func (o *OpFamilyOwnerTo) IsOperatorFamilyOperation() {}

func (o *OpFamilyOwnerTo) String() string {
	return "OWNER TO " + o.Owner.String()
}

// OpFamilySetSchema represents SET SCHEMA operation
type OpFamilySetSchema struct {
	SchemaName *ast.ObjectName
}

func (o *OpFamilySetSchema) exprNode()                  {}
func (o *OpFamilySetSchema) expr()                      {}
func (o *OpFamilySetSchema) IsExpr()                    {}
func (o *OpFamilySetSchema) Span() token.Span           { return token.Span{} }
func (o *OpFamilySetSchema) IsOperatorFamilyOperation() {}

func (o *OpFamilySetSchema) String() string {
	return "SET SCHEMA " + o.SchemaName.String()
}

// OperatorClassOperation represents ALTER OPERATOR CLASS operation.
// This is an interface that can be one of:
// - *OpClassRenameTo
// - *OpClassOwnerTo
// - *OpClassSetSchema
type OperatorClassOperation interface {
	Expr
	IsOperatorClassOperation()
}

// OpClassRenameTo represents RENAME TO operation for operator class
type OpClassRenameTo struct {
	NewName *ast.ObjectName
}

func (o *OpClassRenameTo) exprNode()                 {}
func (o *OpClassRenameTo) expr()                     {}
func (o *OpClassRenameTo) IsExpr()                   {}
func (o *OpClassRenameTo) Span() token.Span          { return token.Span{} }
func (o *OpClassRenameTo) IsOperatorClassOperation() {}

func (o *OpClassRenameTo) String() string {
	return "RENAME TO " + o.NewName.String()
}

// OpClassOwnerTo represents OWNER TO operation for operator class
type OpClassOwnerTo struct {
	Owner Owner
}

func (o *OpClassOwnerTo) exprNode()                 {}
func (o *OpClassOwnerTo) expr()                     {}
func (o *OpClassOwnerTo) IsExpr()                   {}
func (o *OpClassOwnerTo) Span() token.Span          { return token.Span{} }
func (o *OpClassOwnerTo) IsOperatorClassOperation() {}

func (o *OpClassOwnerTo) String() string {
	return "OWNER TO " + o.Owner.String()
}

// OpClassSetSchema represents SET SCHEMA operation for operator class
type OpClassSetSchema struct {
	SchemaName *ast.ObjectName
}

func (o *OpClassSetSchema) exprNode()                 {}
func (o *OpClassSetSchema) expr()                     {}
func (o *OpClassSetSchema) IsExpr()                   {}
func (o *OpClassSetSchema) Span() token.Span          { return token.Span{} }
func (o *OpClassSetSchema) IsOperatorClassOperation() {}

func (o *OpClassSetSchema) String() string {
	return "SET SCHEMA " + o.SchemaName.String()
}

// OptimizerHintStyle represents the style of an optimizer hint.
type OptimizerHintStyle int

const (
	// OptimizerHintStyleMultiLine is for /*+ ... */ style hints.
	OptimizerHintStyleMultiLine OptimizerHintStyle = iota
	// OptimizerHintStyleSingleLine is for --+ ... style hints.
	OptimizerHintStyleSingleLine
)

// OptimizerHint represents an optimizer hint (e.g., /*+ SET_VAR(...) */).
type OptimizerHint struct {
	Prefix  string             // Optional prefix before the + (e.g., "abc" in /*abc+...*/)
	Text    string             // The hint content without markers
	Style   OptimizerHintStyle // The style of the hint (multiline or single line)
	SpanVal token.Span
}

func (o *OptimizerHint) exprNode()        {}
func (o *OptimizerHint) expr()            {}
func (o *OptimizerHint) IsExpr()          {}
func (o *OptimizerHint) Span() token.Span { return o.SpanVal }

func (o *OptimizerHint) String() string {
	switch o.Style {
	case OptimizerHintStyleSingleLine:
		return "--" + o.Prefix + "+" + o.Text
	case OptimizerHintStyleMultiLine:
		return "/*" + o.Prefix + "+" + o.Text + "*/"
	default:
		return "/*+" + o.Text + "*/"
	}
}

// SqliteOnConflict represents SQLite ON CONFLICT clause.
type SqliteOnConflict int

const (
	SqliteOnConflictNone SqliteOnConflict = iota
	SqliteOnConflictReplace
	SqliteOnConflictRollback
	SqliteOnConflictAbort
	SqliteOnConflictFail
	SqliteOnConflictIgnore
)

func (s SqliteOnConflict) String() string {
	switch s {
	case SqliteOnConflictReplace:
		return "OR REPLACE"
	case SqliteOnConflictRollback:
		return "OR ROLLBACK"
	case SqliteOnConflictAbort:
		return "OR ABORT"
	case SqliteOnConflictFail:
		return "OR FAIL"
	case SqliteOnConflictIgnore:
		return "OR IGNORE"
	default:
		return ""
	}
}

// Assignment represents assignment expression.
type Assignment struct {
	Column *ast.Ident
	Value  Expr
}

func (a *Assignment) exprNode()        {}
func (a *Assignment) expr()            {}
func (a *Assignment) IsExpr()          {}
func (a *Assignment) Span() token.Span { return token.Span{} }
func (a *Assignment) String() string {
	if a.Column != nil && a.Value != nil {
		return fmt.Sprintf("%s = %s", a.Column.String(), a.Value.String())
	}
	return ""
}

// OnInsert represents ON INSERT clause (ON CONFLICT or ON DUPLICATE KEY UPDATE).
type OnInsert struct {
	// One of:
	OnConflict         *OnConflict
	DuplicateKeyUpdate []*Assignment
}

func (o *OnInsert) exprNode()        {}
func (o *OnInsert) expr()            {}
func (o *OnInsert) IsExpr()          {}
func (o *OnInsert) Span() token.Span { return token.Span{} }
func (o *OnInsert) String() string {
	if o.OnConflict != nil {
		return o.OnConflict.String()
	}
	if len(o.DuplicateKeyUpdate) > 0 {
		// MySQL style: ON DUPLICATE KEY UPDATE
		assignments := make([]string, len(o.DuplicateKeyUpdate))
		for i, a := range o.DuplicateKeyUpdate {
			assignments[i] = a.String()
		}
		return "ON DUPLICATE KEY UPDATE " + strings.Join(assignments, ", ")
	}
	return ""
}

// OnConflict represents ON CONFLICT clause (PostgreSQL/SQLite).
type OnConflict struct {
	ConflictTarget *ConflictTarget
	Action         OnConflictAction
}

func (o *OnConflict) String() string {
	var result strings.Builder
	result.WriteString("ON CONFLICT")
	if o.ConflictTarget != nil {
		// ON CONFLICT(col1, col2) - no space before columns
		// ON CONFLICT ON CONSTRAINT name - space before ON CONSTRAINT
		if o.ConflictTarget.OnConstraint != nil {
			result.WriteString(" ")
		}
		result.WriteString(o.ConflictTarget.String())
	}
	result.WriteString(" DO ")
	result.WriteString(o.Action.String())
	return result.String()
}

// ConflictTarget represents the target for ON CONFLICT (columns or constraint).
type ConflictTarget struct {
	// One of:
	Columns      []*ast.Ident
	OnConstraint *ast.ObjectName
}

func (c *ConflictTarget) String() string {
	if c.Columns != nil && len(c.Columns) > 0 {
		cols := make([]string, len(c.Columns))
		for i, col := range c.Columns {
			cols[i] = col.String()
		}
		return "(" + strings.Join(cols, ", ") + ")"
	}
	if c.OnConstraint != nil {
		return "ON CONSTRAINT " + c.OnConstraint.String()
	}
	return ""
}

// OnConflictAction represents the action to take on conflict.
type OnConflictAction struct {
	// One of:
	DoNothing bool
	DoUpdate  *DoUpdate
}

func (o OnConflictAction) String() string {
	if o.DoNothing {
		return "NOTHING"
	}
	if o.DoUpdate != nil {
		return o.DoUpdate.String()
	}
	return ""
}

// DoUpdate represents the DO UPDATE action for ON CONFLICT.
type DoUpdate struct {
	Assignments []*Assignment
	Selection   Expr // WHERE clause
}

func (d *DoUpdate) String() string {
	parts := []string{"UPDATE SET"}
	assignments := make([]string, len(d.Assignments))
	for i, a := range d.Assignments {
		assignments[i] = a.String()
	}
	parts = append(parts, strings.Join(assignments, ", "))
	if d.Selection != nil {
		parts = append(parts, "WHERE")
		parts = append(parts, d.Selection.String())
	}
	return strings.Join(parts, " ")
}

// OutputClause represents OUTPUT clause for MERGE, INSERT, UPDATE, or DELETE (MSSQL).
// Example: OUTPUT $action, deleted.* INTO dbo.temp_products
type OutputClause struct {
	OutputToken    *token.Token
	ReturningToken *token.Token // For RETURNING variant
	SelectItems    []query.SelectItem
	IntoTable      *query.SelectInto
}

func (o *OutputClause) exprNode()        {}
func (o *OutputClause) expr()            {}
func (o *OutputClause) IsExpr()          {}
func (o *OutputClause) Span() token.Span { return token.Span{} }
func (o *OutputClause) String() string {
	var f strings.Builder
	if o.OutputToken != nil {
		f.WriteString("OUTPUT ")
	} else if o.ReturningToken != nil {
		f.WriteString("RETURNING ")
	}
	items := make([]string, len(o.SelectItems))
	for i, item := range o.SelectItems {
		items[i] = item.String()
	}
	f.WriteString(strings.Join(items, ", "))
	if o.IntoTable != nil {
		f.WriteString(" ")
		f.WriteString(o.IntoTable.String())
	}
	return f.String()
}

// MysqlInsertPriority represents MySQL INSERT priority.
type MysqlInsertPriority int

const (
	MysqlInsertPriorityNone MysqlInsertPriority = iota
	MysqlInsertPriorityLowPriority
	MysqlInsertPriorityDelayed
	MysqlInsertPriorityHighPriority
)

func (m MysqlInsertPriority) String() string {
	switch m {
	case MysqlInsertPriorityLowPriority:
		return "LOW_PRIORITY"
	case MysqlInsertPriorityDelayed:
		return "DELAYED"
	case MysqlInsertPriorityHighPriority:
		return "HIGH_PRIORITY"
	default:
		return ""
	}
}

// InsertAliases represents INSERT aliases (MySQL: INSERT ... VALUES (1) AS alias (col1, col2)).
type InsertAliases struct {
	RowAlias   *ast.ObjectName
	ColAliases []*ast.Ident
}

func (i *InsertAliases) exprNode()        {}
func (i *InsertAliases) expr()            {}
func (i *InsertAliases) IsExpr()          {}
func (i *InsertAliases) Span() token.Span { return token.Span{} }
func (i *InsertAliases) String() string {
	if i.RowAlias == nil {
		return ""
	}
	var parts []string
	parts = append(parts, "AS")
	parts = append(parts, i.RowAlias.String())
	if len(i.ColAliases) > 0 {
		cols := make([]string, len(i.ColAliases))
		for i, col := range i.ColAliases {
			cols[i] = col.String()
		}
		parts = append(parts, "("+strings.Join(cols, ", ")+")")
	}
	return strings.Join(parts, " ")
}

// Setting represents a setting clause.
type Setting struct {
	Name  string
	Value Expr
}

func (s *Setting) exprNode()        {}
func (s *Setting) expr()            {}
func (s *Setting) IsExpr()          {}
func (s *Setting) Span() token.Span { return token.Span{} }
func (s *Setting) String() string {
	return fmt.Sprintf("%s = %s", s.Name, s.Value.String())
}

// InputFormatClause represents input format clause.
type InputFormatClause struct{}

func (i *InputFormatClause) exprNode()        {}
func (i *InputFormatClause) expr()            {}
func (i *InputFormatClause) IsExpr()          {}
func (i *InputFormatClause) Span() token.Span { return token.Span{} }
func (i *InputFormatClause) String() string   { return "" }

// MultiTableInsertType represents multi-table insert type.
type MultiTableInsertType int

const (
	MultiTableInsertTypeNone MultiTableInsertType = iota
	MultiTableInsertTypeAll
	MultiTableInsertTypeFirst
)

func (m MultiTableInsertType) String() string {
	switch m {
	case MultiTableInsertTypeAll:
		return "ALL"
	case MultiTableInsertTypeFirst:
		return "FIRST"
	default:
		return ""
	}
}

// MultiTableInsertValues represents the VALUES clause in a multi-table INSERT INTO clause.
type MultiTableInsertValues struct {
	Values []MultiTableInsertValue
}

func (m *MultiTableInsertValues) exprNode()        {}
func (m *MultiTableInsertValues) expr()            {}
func (m *MultiTableInsertValues) IsExpr()          {}
func (m *MultiTableInsertValues) Span() token.Span { return token.Span{} }

func (m *MultiTableInsertValues) String() string {
	var parts []string
	for _, v := range m.Values {
		parts = append(parts, v.String())
	}
	return strings.Join(parts, ", ")
}

// MultiTableInsertValue represents a value in a multi-table INSERT VALUES clause.
type MultiTableInsertValue struct {
	IsDefault bool
	Expr      interface{} // Can be expr.Expr or ast.Expr - use interface{} for compatibility
}

func (m *MultiTableInsertValue) exprNode()        {}
func (m *MultiTableInsertValue) expr()            {}
func (m *MultiTableInsertValue) IsExpr()          {}
func (m *MultiTableInsertValue) Span() token.Span { return token.Span{} }

func (m *MultiTableInsertValue) String() string {
	if m.IsDefault {
		return "DEFAULT"
	}
	if m.Expr != nil {
		// Handle both expr.Expr and ast.Expr via interface{} type
		switch e := m.Expr.(type) {
		case Expr:
			return e.String()
		case fmt.Stringer:
			return e.String()
		default:
			return ""
		}
	}
	return ""
}

// MultiTableInsertIntoClause represents multi-table INSERT INTO clause.
type MultiTableInsertIntoClause struct {
	TableName *ObjectName
	Columns   []*Ident
	Values    *MultiTableInsertValues
}

func (m *MultiTableInsertIntoClause) exprNode()        {}
func (m *MultiTableInsertIntoClause) expr()            {}
func (m *MultiTableInsertIntoClause) IsExpr()          {}
func (m *MultiTableInsertIntoClause) Span() token.Span { return token.Span{} }

func (m *MultiTableInsertIntoClause) String() string {
	var sb strings.Builder
	sb.WriteString("INTO ")
	if m.TableName != nil {
		sb.WriteString(m.TableName.String())
	}
	if len(m.Columns) > 0 {
		sb.WriteString(" (")
		for i, col := range m.Columns {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(col.String())
		}
		sb.WriteString(")")
	}
	if m.Values != nil {
		sb.WriteString(" VALUES (")
		sb.WriteString(m.Values.String())
		sb.WriteString(")")
	}
	return sb.String()
}

// MultiTableInsertWhenClause represents multi-table INSERT WHEN clause.
type MultiTableInsertWhenClause struct {
	Condition   interface{} // Can be expr.Expr or ast.Expr - use interface{} for compatibility
	IntoClauses []*MultiTableInsertIntoClause
}

func (m *MultiTableInsertWhenClause) exprNode()        {}
func (m *MultiTableInsertWhenClause) expr()            {}
func (m *MultiTableInsertWhenClause) IsExpr()          {}
func (m *MultiTableInsertWhenClause) Span() token.Span { return token.Span{} }

func (m *MultiTableInsertWhenClause) String() string {
	var sb strings.Builder
	sb.WriteString("WHEN ")
	if m.Condition != nil {
		// Handle both expr.Expr and ast.Expr via interface{} type
		switch c := m.Condition.(type) {
		case Expr:
			sb.WriteString(c.String())
		case fmt.Stringer:
			sb.WriteString(c.String())
		}
	}
	sb.WriteString(" THEN")
	for _, into := range m.IntoClauses {
		sb.WriteString(" ")
		sb.WriteString(into.String())
	}
	return sb.String()
}

// MergeClause represents a WHEN clause within a MERGE statement.
// Example: WHEN NOT MATCHED BY SOURCE AND product LIKE '%washer%' THEN DELETE
type MergeClause struct {
	WhenToken  *token.Token
	ClauseKind MergeClauseKind
	Predicate  Expr
	Action     *MergeAction
}

func (m *MergeClause) exprNode()        {}
func (m *MergeClause) expr()            {}
func (m *MergeClause) IsExpr()          {}
func (m *MergeClause) Span() token.Span { return token.Span{} }
func (m *MergeClause) String() string {
	var f strings.Builder
	f.WriteString("WHEN ")
	f.WriteString(m.ClauseKind.String())
	if m.Predicate != nil {
		f.WriteString(" AND ")
		f.WriteString(m.Predicate.String())
	}
	f.WriteString(" THEN ")
	f.WriteString(m.Action.String())
	return f.String()
}

// SetScope represents SET scope.
type SetScope int

const (
	SetScopeNone SetScope = iota
)

func (s SetScope) String() string { return "" }

// CaseStatementWhen represents CASE statement WHEN clause.
type CaseStatementWhen struct{}

func (c *CaseStatementWhen) exprNode()        {}
func (c *CaseStatementWhen) expr()            {}
func (c *CaseStatementWhen) IsExpr()          {}
func (c *CaseStatementWhen) Span() token.Span { return token.Span{} }
func (c *CaseStatementWhen) String() string   { return "" }

// CaseStatementElse represents CASE statement ELSE clause.
type CaseStatementElse struct{}

func (c *CaseStatementElse) exprNode()        {}
func (c *CaseStatementElse) expr()            {}
func (c *CaseStatementElse) IsExpr()          {}
func (c *CaseStatementElse) Span() token.Span { return token.Span{} }
func (c *CaseStatementElse) String() string   { return "" }

// IfStatementCondition represents IF statement IF/ELSEIF condition block.
// Reference: src/ast/mod.rs:2701 ConditionalStatementBlock
type IfStatementCondition struct {
	Condition  Expr            // The boolean condition expression
	Statements []ast.Statement // Statements to execute when condition is true
}

func (i *IfStatementCondition) exprNode()        {}
func (i *IfStatementCondition) expr()            {}
func (i *IfStatementCondition) IsExpr()          {}
func (i *IfStatementCondition) Span() token.Span { return token.Span{} }

func (i *IfStatementCondition) String() string {
	var sb strings.Builder
	if i.Condition != nil {
		sb.WriteString(i.Condition.String())
	}
	sb.WriteString(" THEN")
	for _, stmt := range i.Statements {
		sb.WriteString(" ")
		sb.WriteString(stmt.String())
		// Add semicolon after each statement
		sb.WriteString(";")
	}
	return sb.String()
}

// IfStatementElse represents IF statement ELSE clause.
// Reference: src/ast/mod.rs:2701 ConditionalStatementBlock
type IfStatementElse struct {
	Statements []ast.Statement
}

func (i *IfStatementElse) exprNode()        {}
func (i *IfStatementElse) expr()            {}
func (i *IfStatementElse) IsExpr()          {}
func (i *IfStatementElse) Span() token.Span { return token.Span{} }

func (i *IfStatementElse) String() string {
	var sb strings.Builder
	sb.WriteString(" ELSE")
	for _, stmt := range i.Statements {
		sb.WriteString(" ")
		sb.WriteString(stmt.String())
		// Add semicolon after each statement
		sb.WriteString(";")
	}
	return sb.String()
}

// RaiseLevel represents RAISE level.
type RaiseLevel int

const (
	RaiseLevelNone RaiseLevel = iota
)

func (r RaiseLevel) String() string { return "" }

// RaiseUsing represents RAISE USING clause.
type RaiseUsing struct{}

func (r *RaiseUsing) exprNode()        {}
func (r *RaiseUsing) expr()            {}
func (r *RaiseUsing) IsExpr()          {}
func (r *RaiseUsing) Span() token.Span { return token.Span{} }
func (r *RaiseUsing) String() string   { return "" }

// CopySource represents the source for a COPY command: a table or a query.
type CopySource struct {
	TableName *ast.ObjectName
	Columns   []*ast.Ident
	Query     interface{} // *query.Query - using interface{} to avoid import cycle
	SpanVal   token.Span
}

func (c *CopySource) exprNode() {}
func (c *CopySource) expr()     {}
func (c *CopySource) IsExpr()   {}

// Span returns the source span for this expression.
func (c *CopySource) Span() token.Span { return c.SpanVal }

// String returns the SQL representation.
func (c *CopySource) String() string {
	if c.Query != nil {
		// Use fmt.Sprintf with %v to call String() method if available
		return fmt.Sprintf("(%v)", c.Query)
	}
	if c.TableName != nil {
		return c.TableName.String()
	}
	return ""
}

// CopyTargetKind represents the kind of COPY target.
type CopyTargetKind int

const (
	CopyTargetKindStdin CopyTargetKind = iota
	CopyTargetKindStdout
	CopyTargetKindFile
	CopyTargetKindProgram
)

// CopyTarget represents the target for a COPY command: STDIN, STDOUT, a file, or a program.
type CopyTarget struct {
	Kind     CopyTargetKind
	Filename string // For File kind
	Command  string // For Program kind
	SpanVal  token.Span
}

func (c *CopyTarget) exprNode() {}
func (c *CopyTarget) expr()     {}
func (c *CopyTarget) IsExpr()   {}

// Span returns the source span for this expression.
func (c *CopyTarget) Span() token.Span { return c.SpanVal }

// String returns the SQL representation.
func (c *CopyTarget) String() string {
	switch c.Kind {
	case CopyTargetKindStdin:
		return "STDIN"
	case CopyTargetKindStdout:
		return "STDOUT"
	case CopyTargetKindFile:
		return fmt.Sprintf("'%s'", escapeSingleQuote(c.Filename))
	case CopyTargetKindProgram:
		return fmt.Sprintf("PROGRAM '%s'", escapeSingleQuote(c.Command))
	}
	return ""
}

// CopyOption represents an option in a COPY statement (PostgreSQL 9.0+).
type CopyOption struct {
	OptionType CopyOptionType
	Value      interface{} // Can be Ident, bool, char, string, or []Ident
}

// CopyOptionType represents the type of COPY option.
type CopyOptionType int

const (
	CopyOptionFormat CopyOptionType = iota
	CopyOptionFreeze
	CopyOptionDelimiter
	CopyOptionNull
	CopyOptionHeader
	CopyOptionQuote
	CopyOptionEscape
	CopyOptionForceQuote
	CopyOptionForceNotNull
	CopyOptionForceNull
	CopyOptionEncoding
)

func (c *CopyOption) exprNode() {}
func (c *CopyOption) expr()     {}
func (c *CopyOption) IsExpr()   {}

// Span returns the source span for this expression.
func (c *CopyOption) Span() token.Span { return token.Span{} }

// String returns the SQL representation.
func (c *CopyOption) String() string {
	switch c.OptionType {
	case CopyOptionFormat:
		if ident, ok := c.Value.(*ast.Ident); ok {
			return fmt.Sprintf("FORMAT %s", ident.String())
		}
	case CopyOptionFreeze:
		if val, ok := c.Value.(bool); ok && !val {
			return "FREEZE FALSE"
		}
		return "FREEZE"
	case CopyOptionDelimiter:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("DELIMITER '%s'", val)
		}
	case CopyOptionNull:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("NULL '%s'", escapeSingleQuote(val))
		}
	case CopyOptionHeader:
		if val, ok := c.Value.(bool); ok && !val {
			return "HEADER FALSE"
		}
		return "HEADER"
	case CopyOptionQuote:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("QUOTE '%s'", val)
		}
	case CopyOptionEscape:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("ESCAPE '%s'", val)
		}
	case CopyOptionForceQuote:
		if cols, ok := c.Value.([]*ast.Ident); ok {
			colStrs := make([]string, len(cols))
			for i, col := range cols {
				colStrs[i] = col.String()
			}
			return fmt.Sprintf("FORCE_QUOTE (%s)", strings.Join(colStrs, ", "))
		}
	case CopyOptionForceNotNull:
		if cols, ok := c.Value.([]*ast.Ident); ok {
			colStrs := make([]string, len(cols))
			for i, col := range cols {
				colStrs[i] = col.String()
			}
			return fmt.Sprintf("FORCE_NOT_NULL (%s)", strings.Join(colStrs, ", "))
		}
	case CopyOptionForceNull:
		if cols, ok := c.Value.([]*ast.Ident); ok {
			colStrs := make([]string, len(cols))
			for i, col := range cols {
				colStrs[i] = col.String()
			}
			return fmt.Sprintf("FORCE_NULL (%s)", strings.Join(colStrs, ", "))
		}
	case CopyOptionEncoding:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("ENCODING '%s'", escapeSingleQuote(val))
		}
	}
	return ""
}

// CopyLegacyOptionType represents the type of COPY legacy option.
type CopyLegacyOptionType int

const (
	CopyLegacyOptionAcceptAnyDate CopyLegacyOptionType = iota
	CopyLegacyOptionAcceptInvChars
	CopyLegacyOptionAddQuotes
	CopyLegacyOptionAllowOverwrite
	CopyLegacyOptionBinary
	CopyLegacyOptionBlankAsNull
	CopyLegacyOptionBzip2
	CopyLegacyOptionCleanPath
	CopyLegacyOptionCompUpdate
	CopyLegacyOptionCsv
	CopyLegacyOptionDateFormat
	CopyLegacyOptionDelimiter
	CopyLegacyOptionEmptyAsNull
	CopyLegacyOptionEncrypted
	CopyLegacyOptionEscape
	CopyLegacyOptionExtension
	CopyLegacyOptionFixedWidth
	CopyLegacyOptionForceNotNull
	CopyLegacyOptionForceNull
	CopyLegacyOptionForceQuote
	CopyLegacyOptionGzip
	CopyLegacyOptionHeader
	CopyLegacyOptionIamRole
	CopyLegacyOptionIgnoreHeader
	CopyLegacyOptionJson
	CopyLegacyOptionManifest
	CopyLegacyOptionMaxFileSize
	CopyLegacyOptionNull
	CopyLegacyOptionParallel
	CopyLegacyOptionParquet
	CopyLegacyOptionPartitionBy
	CopyLegacyOptionQuote
	CopyLegacyOptionRegion
	CopyLegacyOptionRemoveQuotes
	CopyLegacyOptionRowGroupSize
	CopyLegacyOptionStatUpdate
	CopyLegacyOptionTimeFormat
	CopyLegacyOptionTruncateColumns
	CopyLegacyOptionZstd
	CopyLegacyOptionCredentials
)

// CopyLegacyOption represents a legacy option in a COPY statement (pre-PostgreSQL 9.0, Redshift).
type CopyLegacyOption struct {
	OptionType CopyLegacyOptionType
	Value      interface{} // Type depends on OptionType
}

func (c *CopyLegacyOption) exprNode() {}
func (c *CopyLegacyOption) expr()     {}
func (c *CopyLegacyOption) IsExpr()   {}

// Span returns the source span for this expression.
func (c *CopyLegacyOption) Span() token.Span { return token.Span{} }

// String returns the SQL representation.
func (c *CopyLegacyOption) String() string {
	switch c.OptionType {
	case CopyLegacyOptionAcceptAnyDate:
		return "ACCEPTANYDATE"
	case CopyLegacyOptionAcceptInvChars:
		if val, ok := c.Value.(string); ok && val != "" {
			return fmt.Sprintf("ACCEPTINVCHARS '%s'", escapeSingleQuote(val))
		}
		return "ACCEPTINVCHARS"
	case CopyLegacyOptionAddQuotes:
		return "ADDQUOTES"
	case CopyLegacyOptionAllowOverwrite:
		return "ALLOWOVERWRITE"
	case CopyLegacyOptionBinary:
		return "BINARY"
	case CopyLegacyOptionBlankAsNull:
		return "BLANKSASNULL"
	case CopyLegacyOptionBzip2:
		return "BZIP2"
	case CopyLegacyOptionCleanPath:
		return "CLEANPATH"
	case CopyLegacyOptionCompUpdate:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("COMPUPDATE %s", val)
		}
		return "COMPUPDATE"
	case CopyLegacyOptionCsv:
		return "CSV"
	case CopyLegacyOptionDateFormat:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("DATEFORMAT '%s'", escapeSingleQuote(val))
		}
		return "DATEFORMAT"
	case CopyLegacyOptionDelimiter:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("DELIMITER '%s'", val)
		}
		return "DELIMITER"
	case CopyLegacyOptionEmptyAsNull:
		return "EMPTYASNULL"
	case CopyLegacyOptionEncrypted:
		return "ENCRYPTED"
	case CopyLegacyOptionEscape:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("ESCAPE '%s'", escapeSingleQuote(val))
		}
		return "ESCAPE"
	case CopyLegacyOptionExtension:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("EXTENSION '%s'", escapeSingleQuote(val))
		}
		return "EXTENSION"
	case CopyLegacyOptionFixedWidth:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("FIXEDWIDTH '%s'", escapeSingleQuote(val))
		}
		return "FIXEDWIDTH"
	case CopyLegacyOptionForceNotNull:
		if val, ok := c.Value.(*ast.Ident); ok {
			return fmt.Sprintf("FORCE NOT NULL %s", val.String())
		}
		return "FORCE NOT NULL"
	case CopyLegacyOptionForceNull:
		if val, ok := c.Value.(*ast.Ident); ok {
			return fmt.Sprintf("FORCE NULL %s", val.String())
		}
		return "FORCE NULL"
	case CopyLegacyOptionForceQuote:
		if val, ok := c.Value.(*ast.Ident); ok {
			return fmt.Sprintf("FORCE QUOTE %s", val.String())
		}
		return "FORCE QUOTE"
	case CopyLegacyOptionGzip:
		return "GZIP"
	case CopyLegacyOptionHeader:
		return "HEADER"
	case CopyLegacyOptionIamRole:
		if val, ok := c.Value.(IamRoleKind); ok {
			return fmt.Sprintf("IAM_ROLE %s", val.String())
		}
		return "IAM_ROLE"
	case CopyLegacyOptionIgnoreHeader:
		if val, ok := c.Value.(int); ok {
			return fmt.Sprintf("IGNOREHEADER %d", val)
		}
		return "IGNOREHEADER"
	case CopyLegacyOptionJson:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("JSON '%s'", escapeSingleQuote(val))
		}
		return "JSON"
	case CopyLegacyOptionManifest:
		return "MANIFEST"
	case CopyLegacyOptionMaxFileSize:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("MAXFILESIZE %s", val)
		}
		return "MAXFILESIZE"
	case CopyLegacyOptionNull:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("NULL '%s'", escapeSingleQuote(val))
		}
		return "NULL"
	case CopyLegacyOptionParallel:
		if val, ok := c.Value.(bool); ok {
			if val {
				return "PARALLEL TRUE"
			}
			return "PARALLEL FALSE"
		}
		return "PARALLEL"
	case CopyLegacyOptionParquet:
		return "PARQUET"
	case CopyLegacyOptionQuote:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("QUOTE '%s'", escapeSingleQuote(val))
		}
		return "QUOTE"
	case CopyLegacyOptionRegion:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("REGION '%s'", escapeSingleQuote(val))
		}
		return "REGION"
	case CopyLegacyOptionRemoveQuotes:
		return "REMOVEQUOTES"
	case CopyLegacyOptionRowGroupSize:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("ROWGROUPSIZE %s", val)
		}
		return "ROWGROUPSIZE"
	case CopyLegacyOptionStatUpdate:
		if val, ok := c.Value.(bool); ok {
			if val {
				return "STATUPDATE TRUE"
			}
			return "STATUPDATE FALSE"
		}
		return "STATUPDATE"
	case CopyLegacyOptionTimeFormat:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("TIMEFORMAT '%s'", escapeSingleQuote(val))
		}
		return "TIMEFORMAT"
	case CopyLegacyOptionTruncateColumns:
		return "TRUNCATECOLUMNS"
	case CopyLegacyOptionZstd:
		return "ZSTD"
	case CopyLegacyOptionCredentials:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("CREDENTIALS '%s'", escapeSingleQuote(val))
		}
		return "CREDENTIALS"
	}
	return ""
}

// CopyIntoSnowflakeKind represents COPY INTO Snowflake kind.
type CopyIntoSnowflakeKind int

const (
	CopyIntoSnowflakeKindNone CopyIntoSnowflakeKind = iota
	CopyIntoSnowflakeKindTable
	CopyIntoSnowflakeKindLocation
)

func (c CopyIntoSnowflakeKind) String() string {
	switch c {
	case CopyIntoSnowflakeKindTable:
		return "TABLE"
	case CopyIntoSnowflakeKindLocation:
		return "LOCATION"
	default:
		return ""
	}
}

// StageLoadSelectItem represents a single item in the SELECT list for data loading from staged files.
type StageLoadSelectItem struct {
	Alias      *ast.Ident
	FileColNum int32
	Element    *ast.Ident
	ItemAs     *ast.Ident
}

func (s *StageLoadSelectItem) exprNode()        {}
func (s *StageLoadSelectItem) expr()            {}
func (s *StageLoadSelectItem) IsExpr()          {}
func (s *StageLoadSelectItem) Span() token.Span { return token.Span{} }
func (s *StageLoadSelectItem) String() string {
	var parts []string
	if s.Alias != nil {
		parts = append(parts, s.Alias.String()+".")
	}
	parts = append(parts, fmt.Sprintf("$%d", s.FileColNum))
	if s.Element != nil {
		parts = append(parts, ":"+s.Element.String())
	}
	if s.ItemAs != nil {
		parts = append(parts, " AS "+s.ItemAs.String())
	}
	return strings.Join(parts, "")
}

// StageLoadSelectItemKind represents stage load SELECT item kind.
type StageLoadSelectItemKind int

const (
	StageLoadSelectItemKindNone StageLoadSelectItemKind = iota
	StageLoadSelectItemKindSelectItem
	StageLoadSelectItemKindStageLoad
)

func (s StageLoadSelectItemKind) String() string { return "" }

// StageLoadSelectItemWrapper wraps a stage load select item.
type StageLoadSelectItemWrapper struct {
	Kind StageLoadSelectItemKind
	Item interface{} // Can be *StageLoadSelectItem or ast.SelectItem
}

func (s *StageLoadSelectItemWrapper) exprNode()        {}
func (s *StageLoadSelectItemWrapper) expr()            {}
func (s *StageLoadSelectItemWrapper) IsExpr()          {}
func (s *StageLoadSelectItemWrapper) Span() token.Span { return token.Span{} }
func (s *StageLoadSelectItemWrapper) String() string {
	if s.Item != nil {
		if str, ok := s.Item.(fmt.Stringer); ok {
			return str.String()
		}
	}
	return ""
}

// CloseCursorKind represents which cursor(s) to close.
type CloseCursorKind int

const (
	// CloseCursorAll closes all cursors.
	CloseCursorAll CloseCursorKind = iota
	// CloseCursorSpecific closes a specific cursor by name.
	CloseCursorSpecific
)

// CloseCursor represents which cursor(s) to close.
type CloseCursor struct {
	Kind CloseCursorKind
	Name *Ident // Only set when Kind is CloseCursorSpecific
}

func (c *CloseCursor) exprNode() {}
func (c *CloseCursor) expr()     {}
func (c *CloseCursor) IsExpr()   {}

// Span returns the source span for this expression.
func (c *CloseCursor) Span() token.Span {
	if c.Name != nil {
		return c.Name.Span()
	}
	return token.Span{}
}

// String returns the SQL representation.
func (c *CloseCursor) String() string {
	switch c.Kind {
	case CloseCursorAll:
		return "ALL"
	case CloseCursorSpecific:
		if c.Name != nil {
			return c.Name.String()
		}
	}
	return ""
}

// DeclareType represents the type of a DECLARE statement.
type DeclareType int

const (
	// DeclareTypeCursor is for cursor variable type. e.g. Snowflake, PostgreSQL, MsSql
	DeclareTypeCursor DeclareType = iota
	// DeclareTypeResultSet is for result set variable type (Snowflake)
	DeclareTypeResultSet
	// DeclareTypeException is for exception declaration syntax (Snowflake)
	DeclareTypeException
)

func (d DeclareType) String() string {
	switch d {
	case DeclareTypeCursor:
		return "CURSOR"
	case DeclareTypeResultSet:
		return "RESULTSET"
	case DeclareTypeException:
		return "EXCEPTION"
	}
	return ""
}

// DeclareAssignment represents the assignment type for DECLARE variables.
type DeclareAssignment int

const (
	// DeclareAssignmentExpr is a plain expression specified.
	DeclareAssignmentExpr DeclareAssignment = iota
	// DeclareAssignmentDefault is expression assigned via the DEFAULT keyword.
	DeclareAssignmentDefault
	// DeclareAssignmentDuckAssignment is expression assigned via the := syntax.
	DeclareAssignmentDuckAssignment
	// DeclareAssignmentFor is expression via the FOR keyword.
	DeclareAssignmentFor
	// DeclareAssignmentMsSqlAssignment is expression via the = syntax (MSSQL).
	DeclareAssignmentMsSqlAssignment
)

func (d DeclareAssignment) String() string {
	switch d {
	case DeclareAssignmentExpr:
		return ""
	case DeclareAssignmentDefault:
		return "DEFAULT"
	case DeclareAssignmentDuckAssignment:
		return ":="
	case DeclareAssignmentFor:
		return "FOR"
	case DeclareAssignmentMsSqlAssignment:
		return "="
	}
	return ""
}

// Declare represents a single DECLARE statement item.
// A DECLARE statement can contain multiple declarations (e.g. Snowflake).
type Declare struct {
	// Names being declared. Can be multiple for BigQuery style: DECLARE a, b, c DEFAULT 42;
	Names []*Ident
	// DataType assigned to the declared variable (optional)
	// Using interface{} to avoid import cycle - actually stores datatype.DataType
	DataType interface{}
	// Assignment is the expression being assigned to the declared variable
	Assignment Expr
	// AssignmentType indicates how the assignment was made (DEFAULT, :=, =, etc.)
	AssignmentType DeclareAssignment
	// DeclareType represents the type of the declared variable (Cursor, ResultSet, Exception)
	DeclareType *DeclareType
	// Binary causes the cursor to return data in binary rather than in text format
	Binary *bool
	// Sensitive: None = Not specified, Some(true) = INSENSITIVE, Some(false) = ASENSITIVE
	Sensitive *bool
	// Scroll: None = Not specified, Some(true) = SCROLL, Some(false) = NO SCROLL
	Scroll *bool
	// Hold: None = Not specified, Some(true) = WITH HOLD, Some(false) = WITHOUT HOLD
	Hold *bool
	// ForQuery is the FOR <query> clause in a CURSOR declaration
	ForQuery *query.Query
	// ExceptionParams stores the (code, message) tuple for EXCEPTION declarations
	// Only used when DeclareType is DeclareTypeException
	ExceptionParams []Expr
	// SpanVal is the source span
	SpanVal token.Span
}

func (d *Declare) exprNode() {}
func (d *Declare) expr()     {}
func (d *Declare) IsExpr()   {}

// Span returns the source span for this expression.
func (d *Declare) Span() token.Span {
	return d.SpanVal
}

// String returns the SQL representation.
func (d *Declare) String() string {
	if d == nil {
		return ""
	}
	var sb strings.Builder

	// Write names (comma-separated for multiple)
	for i, name := range d.Names {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(name.String())
	}

	// Write binary option
	if d.Binary != nil {
		if *d.Binary {
			sb.WriteString(" BINARY")
		}
	}

	// Write sensitive option
	if d.Sensitive != nil {
		if *d.Sensitive {
			sb.WriteString(" INSENSITIVE")
		} else {
			sb.WriteString(" ASENSITIVE")
		}
	}

	// Write scroll option
	if d.Scroll != nil {
		if *d.Scroll {
			sb.WriteString(" SCROLL")
		} else {
			sb.WriteString(" NO SCROLL")
		}
	}

	// Write declare type
	if d.DeclareType != nil {
		sb.WriteString(" ")
		sb.WriteString(d.DeclareType.String())
	}

	// Write hold option
	if d.Hold != nil {
		if *d.Hold {
			sb.WriteString(" WITH HOLD")
		} else {
			sb.WriteString(" WITHOUT HOLD")
		}
	}

	// Write data type
	if d.DataType != nil {
		sb.WriteString(" ")
		if s, ok := d.DataType.(fmt.Stringer); ok {
			sb.WriteString(s.String())
		}
	}

	// Write assignment - special handling for EXCEPTION and RESULTSET
	if d.DeclareType != nil && *d.DeclareType == DeclareTypeException && len(d.ExceptionParams) > 0 {
		// EXCEPTION uses parentheses: EXCEPTION (code, 'message')
		sb.WriteString(" (")
		for i, param := range d.ExceptionParams {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(param.String())
		}
		sb.WriteString(")")
	} else if d.Assignment != nil {
		if d.DeclareType != nil && *d.DeclareType == DeclareTypeResultSet && d.ForQuery != nil {
			// RESULTSET with query uses: RESULTSET DEFAULT (query)
			sb.WriteString(" ")
			sb.WriteString(d.AssignmentType.String())
			sb.WriteString(" (")
			sb.WriteString(d.ForQuery.String())
			sb.WriteString(")")
		} else {
			// Standard assignment
			if d.AssignmentType != DeclareAssignmentExpr {
				sb.WriteString(" ")
				sb.WriteString(d.AssignmentType.String())
			}
			sb.WriteString(" ")
			sb.WriteString(d.Assignment.String())
		}
	} else if d.DeclareType != nil && *d.DeclareType == DeclareTypeResultSet && d.ForQuery != nil {
		// RESULTSET with query but no explicit assignment (should have DEFAULT)
		sb.WriteString(" DEFAULT (")
		sb.WriteString(d.ForQuery.String())
		sb.WriteString(")")
	}

	// Write FOR query (only for CURSOR type, not RESULTSET which is handled above)
	if d.ForQuery != nil && (d.DeclareType == nil || *d.DeclareType == DeclareTypeCursor) {
		sb.WriteString(" FOR ")
		sb.WriteString(d.ForQuery.String())
	}

	return sb.String()
}

// FetchDirection represents FETCH direction.
type FetchDirection struct {
	Kind  FetchDirectionKind
	Limit *Expr // Optional limit value for Count, Absolute, Relative, Forward, Backward
}

// FetchDirectionKind represents the kind of fetch direction.
type FetchDirectionKind int

const (
	FetchDirectionCount FetchDirectionKind = iota
	FetchDirectionNext
	FetchDirectionPrior
	FetchDirectionFirst
	FetchDirectionLast
	FetchDirectionAbsolute
	FetchDirectionRelative
	FetchDirectionAll
	FetchDirectionForward
	FetchDirectionForwardAll
	FetchDirectionBackward
	FetchDirectionBackwardAll
)

func (f *FetchDirection) String() string {
	if f == nil {
		return ""
	}
	var sb strings.Builder
	switch f.Kind {
	case FetchDirectionCount:
		if f.Limit != nil {
			sb.WriteString((*f.Limit).String())
		}
	case FetchDirectionNext:
		sb.WriteString("NEXT")
	case FetchDirectionPrior:
		sb.WriteString("PRIOR")
	case FetchDirectionFirst:
		sb.WriteString("FIRST")
	case FetchDirectionLast:
		sb.WriteString("LAST")
	case FetchDirectionAbsolute:
		sb.WriteString("ABSOLUTE ")
		if f.Limit != nil {
			sb.WriteString((*f.Limit).String())
		}
	case FetchDirectionRelative:
		sb.WriteString("RELATIVE ")
		if f.Limit != nil {
			sb.WriteString((*f.Limit).String())
		}
	case FetchDirectionAll:
		sb.WriteString("ALL")
	case FetchDirectionForward:
		sb.WriteString("FORWARD")
		if f.Limit != nil {
			sb.WriteString(" ")
			sb.WriteString((*f.Limit).String())
		}
	case FetchDirectionForwardAll:
		sb.WriteString("FORWARD ALL")
	case FetchDirectionBackward:
		sb.WriteString("BACKWARD")
		if f.Limit != nil {
			sb.WriteString(" ")
			sb.WriteString((*f.Limit).String())
		}
	case FetchDirectionBackwardAll:
		sb.WriteString("BACKWARD ALL")
	}
	return sb.String()
}

// FetchPosition represents FETCH position.
type FetchPosition int

const (
	FetchPositionFrom FetchPosition = iota
	FetchPositionIn
)

func (f FetchPosition) String() string {
	switch f {
	case FetchPositionFrom:
		return "FROM"
	case FetchPositionIn:
		return "IN"
	}
	return ""
}

// FlushType represents FLUSH type.
type FlushType int

const (
	FlushTypeNone FlushType = iota
	FlushTypeOptimizerCosts
	FlushTypeBinaryLogs
	FlushTypeEngineLogs
	FlushTypeErrorLogs
	FlushTypeGeneralLogs
	FlushTypeRelayLogs
	FlushTypeSlowLogs
	FlushTypeTables
	FlushTypeHosts
	FlushTypePrivileges
	FlushTypeLogs
	FlushTypeStatus
)

func (f FlushType) String() string {
	switch f {
	case FlushTypeOptimizerCosts:
		return "OPTIMIZER_COSTS"
	case FlushTypeBinaryLogs:
		return "BINARY LOGS"
	case FlushTypeEngineLogs:
		return "ENGINE LOGS"
	case FlushTypeErrorLogs:
		return "ERROR LOGS"
	case FlushTypeGeneralLogs:
		return "GENERAL LOGS"
	case FlushTypeRelayLogs:
		return "RELAY LOGS"
	case FlushTypeSlowLogs:
		return "SLOW LOGS"
	case FlushTypeTables:
		return "TABLES"
	case FlushTypeHosts:
		return "HOSTS"
	case FlushTypePrivileges:
		return "PRIVILEGES"
	case FlushTypeLogs:
		return "LOGS"
	case FlushTypeStatus:
		return "STATUS"
	default:
		return ""
	}
}

// FlushLocation represents FLUSH location.
type FlushLocation int

const (
	FlushLocationNone FlushLocation = iota
	FlushLocationLocal
	FlushLocationNoWriteToBinlog
)

func (f FlushLocation) String() string {
	switch f {
	case FlushLocationLocal:
		return "LOCAL"
	case FlushLocationNoWriteToBinlog:
		return "NO_WRITE_TO_BINLOG"
	default:
		return ""
	}
}

// DiscardObject represents DISCARD object.
type DiscardObject int

const (
	DiscardObjectNone DiscardObject = iota
	DiscardAll
	DiscardPlans
	DiscardSequences
	DiscardTemp
)

func (d DiscardObject) String() string {
	switch d {
	case DiscardAll:
		return "ALL"
	case DiscardPlans:
		return "PLANS"
	case DiscardSequences:
		return "SEQUENCES"
	case DiscardTemp:
		return "TEMP"
	default:
		return ""
	}
}

// ShowStatementInClause represents the clause type for SHOW ... IN/FROM
type ShowStatementInClause int

const (
	ShowStatementInClauseNone ShowStatementInClause = iota
	ShowStatementInClauseFrom
	ShowStatementInClauseIn
)

func (s ShowStatementInClause) String() string {
	switch s {
	case ShowStatementInClauseFrom:
		return "FROM"
	case ShowStatementInClauseIn:
		return "IN"
	default:
		return ""
	}
}

// ShowStatementIn represents SHOW statement IN clause.
type ShowStatementIn struct {
	Clause     ShowStatementInClause
	ParentType *ast.Ident
	ParentName *ast.ObjectName
}

func (s *ShowStatementIn) exprNode()        {}
func (s *ShowStatementIn) expr()            {}
func (s *ShowStatementIn) IsExpr()          {}
func (s *ShowStatementIn) Span() token.Span { return token.Span{} }
func (s *ShowStatementIn) String() string {
	clause := "IN"
	if s.Clause == ShowStatementInClauseFrom {
		clause = "FROM"
	}

	// If we have a parent type (like ACCOUNT, DATABASE, SCHEMA), include it
	if s.ParentType != nil {
		if s.ParentName != nil {
			return fmt.Sprintf("%s %s %s", clause, s.ParentType.Value, s.ParentName.String())
		}
		return fmt.Sprintf("%s %s", clause, s.ParentType.Value)
	}

	// Just the name
	if s.ParentName != nil {
		return fmt.Sprintf("%s %s", clause, s.ParentName.String())
	}

	return ""
}

// ShowStatementFilterPosition represents SHOW statement filter position.
type ShowStatementFilterPosition int

const (
	ShowStatementFilterPositionNone ShowStatementFilterPosition = iota
	ShowStatementFilterPositionSuffix
	ShowStatementFilterPositionInfix
)

func (s ShowStatementFilterPosition) String() string { return "" }

// ShowStatementFilter represents SHOW statement filter.
type ShowStatementFilter struct {
	Like         *string
	Where        Expr
	SuffixString *string // For Snowflake-style suffix string literal (e.g., SHOW TABLES IN db1 'abc')
}

func (s *ShowStatementFilter) exprNode()        {}
func (s *ShowStatementFilter) expr()            {}
func (s *ShowStatementFilter) IsExpr()          {}
func (s *ShowStatementFilter) Span() token.Span { return token.Span{} }
func (s *ShowStatementFilter) String() string {
	if s.Like != nil {
		return fmt.Sprintf("LIKE '%s'", *s.Like)
	}
	if s.Where != nil {
		return fmt.Sprintf("WHERE %s", s.Where.String())
	}
	if s.SuffixString != nil {
		return fmt.Sprintf("'%s'", *s.SuffixString)
	}
	return ""
}

// ShowCreateObject represents SHOW CREATE object.
type ShowCreateObject int

const (
	ShowCreateObjectNone ShowCreateObject = iota
	ShowCreateObjectTable
	ShowCreateObjectTrigger
	ShowCreateObjectEvent
	ShowCreateObjectFunction
	ShowCreateObjectProcedure
	ShowCreateObjectView
)

func (s ShowCreateObject) String() string {
	switch s {
	case ShowCreateObjectTable:
		return "TABLE"
	case ShowCreateObjectTrigger:
		return "TRIGGER"
	case ShowCreateObjectEvent:
		return "EVENT"
	case ShowCreateObjectFunction:
		return "FUNCTION"
	case ShowCreateObjectProcedure:
		return "PROCEDURE"
	case ShowCreateObjectView:
		return "VIEW"
	default:
		return ""
	}
}

// ShowStatementOptions represents SHOW statement options.
type ShowStatementOptions struct {
	SpanVal        token.Span
	ShowIn         *ShowStatementIn
	Filter         *ShowStatementFilter
	FilterPosition ShowStatementFilterPosition
	LimitFrom      *string
	Limit          *int
	StartsWith     *string
}

func (s *ShowStatementOptions) exprNode()        {}
func (s *ShowStatementOptions) expr()            {}
func (s *ShowStatementOptions) IsExpr()          {}
func (s *ShowStatementOptions) Span() token.Span { return s.SpanVal }
func (s *ShowStatementOptions) String() string {
	var parts []string

	// For infix filter position (Snowflake style), output filter before ShowIn
	if s.FilterPosition == ShowStatementFilterPositionInfix && s.Filter != nil {
		parts = append(parts, s.Filter.String())
	}

	if s.ShowIn != nil {
		parts = append(parts, s.ShowIn.String())
	}

	// For suffix filter position (MySQL style), output filter after ShowIn
	if s.FilterPosition != ShowStatementFilterPositionInfix && s.Filter != nil {
		parts = append(parts, s.Filter.String())
	}

	if s.StartsWith != nil {
		parts = append(parts, fmt.Sprintf("STARTS WITH '%s'", *s.StartsWith))
	}

	if s.Limit != nil {
		parts = append(parts, fmt.Sprintf("LIMIT %d", *s.Limit))
	}

	if s.LimitFrom != nil {
		parts = append(parts, fmt.Sprintf("FROM '%s'", *s.LimitFrom))
	}

	return strings.Join(parts, " ")
}

// TransactionMode represents transaction mode.
type TransactionMode int

const (
	TransactionModeNone TransactionMode = iota
	TransactionModeReadUncommitted
	TransactionModeReadCommitted
	TransactionModeRepeatableRead
	TransactionModeSerializable
	TransactionModeSnapshot
	TransactionModeReadOnly
	TransactionModeReadWrite
	TransactionModeDeferrable
	TransactionModeNotDeferrable
)

func (t TransactionMode) String() string {
	switch t {
	case TransactionModeReadUncommitted:
		return "ISOLATION LEVEL READ UNCOMMITTED"
	case TransactionModeReadCommitted:
		return "ISOLATION LEVEL READ COMMITTED"
	case TransactionModeRepeatableRead:
		return "ISOLATION LEVEL REPEATABLE READ"
	case TransactionModeSerializable:
		return "ISOLATION LEVEL SERIALIZABLE"
	case TransactionModeSnapshot:
		return "ISOLATION LEVEL SNAPSHOT"
	case TransactionModeReadOnly:
		return "READ ONLY"
	case TransactionModeReadWrite:
		return "READ WRITE"
	case TransactionModeDeferrable:
		return "DEFERRABLE"
	case TransactionModeNotDeferrable:
		return "NOT DEFERRABLE"
	default:
		return ""
	}
}

// BeginTransactionKind represents BEGIN TRANSACTION kind.
type BeginTransactionKind int

const (
	BeginTransactionKindNone BeginTransactionKind = iota
	BeginTransactionKindTransaction
	BeginTransactionKindWork
	BeginTransactionKindTran
)

func (b BeginTransactionKind) String() string {
	switch b {
	case BeginTransactionKindTransaction:
		return "TRANSACTION"
	case BeginTransactionKindWork:
		return "WORK"
	case BeginTransactionKindTran:
		return "TRAN"
	default:
		return ""
	}
}

// TransactionModifier represents transaction modifier.
type TransactionModifier int

const (
	TransactionModifierNone TransactionModifier = iota
	TransactionModifierDeferred
	TransactionModifierImmediate
	TransactionModifierExclusive
	TransactionModifierTry
	TransactionModifierCatch
)

func (t TransactionModifier) String() string {
	switch t {
	case TransactionModifierDeferred:
		return "DEFERRED"
	case TransactionModifierImmediate:
		return "IMMEDIATE"
	case TransactionModifierExclusive:
		return "EXCLUSIVE"
	case TransactionModifierTry:
		return "TRY"
	case TransactionModifierCatch:
		return "CATCH"
	default:
		return ""
	}
}

// ExceptionWhen represents EXCEPTION WHEN clause.
type ExceptionWhen struct{}

func (e *ExceptionWhen) exprNode()        {}
func (e *ExceptionWhen) expr()            {}
func (e *ExceptionWhen) IsExpr()          {}
func (e *ExceptionWhen) Span() token.Span { return token.Span{} }
func (e *ExceptionWhen) String() string   { return "" }

// CommentObject represents COMMENT object.
type CommentObject int

const (
	CommentObjectNone CommentObject = iota
	CommentColumn
	CommentDatabase
	CommentDomain
	CommentExtension
	CommentFunction
	CommentIndex
	CommentMaterializedView
	CommentProcedure
	CommentRole
	CommentSchema
	CommentSequence
	CommentTable
	CommentType
	CommentUser
	CommentView
)

func (c CommentObject) String() string {
	switch c {
	case CommentColumn:
		return "COLUMN"
	case CommentDatabase:
		return "DATABASE"
	case CommentDomain:
		return "DOMAIN"
	case CommentExtension:
		return "EXTENSION"
	case CommentFunction:
		return "FUNCTION"
	case CommentIndex:
		return "INDEX"
	case CommentMaterializedView:
		return "MATERIALIZED VIEW"
	case CommentProcedure:
		return "PROCEDURE"
	case CommentRole:
		return "ROLE"
	case CommentSchema:
		return "SCHEMA"
	case CommentSequence:
		return "SEQUENCE"
	case CommentTable:
		return "TABLE"
	case CommentType:
		return "TYPE"
	case CommentUser:
		return "USER"
	case CommentView:
		return "VIEW"
	default:
		return ""
	}
}

// ExprWithAlias represents expression with alias.
type ExprWithAlias struct {
	Expr  Expr
	Alias *ast.Ident
}

func (e *ExprWithAlias) exprNode()        {}
func (e *ExprWithAlias) expr()            {}
func (e *ExprWithAlias) IsExpr()          {}
func (e *ExprWithAlias) Span() token.Span { return token.Span{} }
func (e *ExprWithAlias) String() string {
	if e.Alias != nil {
		return fmt.Sprintf("%s AS %s", e.Expr.String(), e.Alias.String())
	}
	return e.Expr.String()
}

// KillType represents KILL type.
type KillType int

const (
	KillTypeNone KillType = iota
	KillTypeConnection
	KillTypeQuery
	KillTypeMutation
	KillTypeHard
)

func (k KillType) String() string {
	switch k {
	case KillTypeConnection:
		return "CONNECTION"
	case KillTypeQuery:
		return "QUERY"
	case KillTypeMutation:
		return "MUTATION"
	case KillTypeHard:
		return "HARD"
	default:
		return ""
	}
}

// DescribeAlias represents DESCRIBE alias.
type DescribeAlias int

const (
	DescribeAliasNone DescribeAlias = iota
	DescribeAliasDescribe
	DescribeAliasExplain
	DescribeAliasDesc
)

func (d DescribeAlias) String() string {
	switch d {
	case DescribeAliasDescribe:
		return "DESCRIBE"
	case DescribeAliasExplain:
		return "EXPLAIN"
	case DescribeAliasDesc:
		return "DESC"
	default:
		return ""
	}
}

// HiveDescribeFormat represents Hive DESCRIBE format.
type HiveDescribeFormat int

const (
	HiveDescribeFormatNone HiveDescribeFormat = iota
	HiveDescribeFormatExtended
	HiveDescribeFormatFormatted
)

func (h HiveDescribeFormat) String() string {
	switch h {
	case HiveDescribeFormatExtended:
		return "EXTENDED"
	case HiveDescribeFormatFormatted:
		return "FORMATTED"
	default:
		return ""
	}
}

// AnalyzeFormatKind represents ANALYZE format kind.
type AnalyzeFormatKind int

const (
	AnalyzeFormatKindNone AnalyzeFormatKind = iota
)

func (a AnalyzeFormatKind) String() string { return "" }

// UtilityOption represents utility option.
type UtilityOption struct{}

func (u *UtilityOption) exprNode()        {}
func (u *UtilityOption) expr()            {}
func (u *UtilityOption) IsExpr()          {}
func (u *UtilityOption) Span() token.Span { return token.Span{} }
func (u *UtilityOption) String() string   { return "" }

// ValueWithSpan represents value with span.
type ValueWithSpan struct {
	Value string
}

func (v *ValueWithSpan) exprNode()        {}
func (v *ValueWithSpan) expr()            {}
func (v *ValueWithSpan) IsExpr()          {}
func (v *ValueWithSpan) Span() token.Span { return token.Span{} }
func (v *ValueWithSpan) String() string   { return v.Value }

// LockTable represents a table to lock.
type LockTable struct {
	Table       *ast.ObjectName
	Alias       *ast.Ident
	LockType    LockTableType
	Only        bool // PostgreSQL: ONLY table_name
	HasAsterisk bool // PostgreSQL: table_name* (include descendants)
}

// LockTableType represents the type of lock for LOCK TABLES.
type LockTableType int

const (
	LockTableTypeRead LockTableType = iota
	LockTableTypeReadLocal
	LockTableTypeWrite
	LockTableTypeWriteLowPriority
)

func (l LockTableType) String() string {
	switch l {
	case LockTableTypeRead:
		return "READ"
	case LockTableTypeReadLocal:
		return "READ LOCAL"
	case LockTableTypeWrite:
		return "WRITE"
	case LockTableTypeWriteLowPriority:
		return "LOW_PRIORITY WRITE"
	default:
		return ""
	}
}

func (l *LockTable) exprNode()        {}
func (l *LockTable) expr()            {}
func (l *LockTable) IsExpr()          {}
func (l *LockTable) Span() token.Span { return token.Span{} }
func (l *LockTable) String() string {
	var sb strings.Builder
	if l.Only {
		sb.WriteString("ONLY ")
	}
	sb.WriteString(l.Table.String())
	if l.HasAsterisk {
		sb.WriteString(" *")
	}
	return sb.String()
}

// LockMode represents PostgreSQL lock table modes.
type LockMode int

const (
	LockModeNone LockMode = iota
	LockModeAccessShare
	LockModeAccessExclusive
	LockModeRowShare
	LockModeRowExclusive
	LockModeShareUpdateExclusive
	LockModeShareRowExclusive
	LockModeShare
	LockModeExclusive
)

func (l LockMode) String() string {
	switch l {
	case LockModeAccessShare:
		return "ACCESS SHARE"
	case LockModeAccessExclusive:
		return "ACCESS EXCLUSIVE"
	case LockModeRowShare:
		return "ROW SHARE"
	case LockModeRowExclusive:
		return "ROW EXCLUSIVE"
	case LockModeShareUpdateExclusive:
		return "SHARE UPDATE EXCLUSIVE"
	case LockModeShareRowExclusive:
		return "SHARE ROW EXCLUSIVE"
	case LockModeShare:
		return "SHARE"
	case LockModeExclusive:
		return "EXCLUSIVE"
	default:
		return ""
	}
}

// IamRoleKind represents IAM role kind.
type IamRoleKind struct {
	Kind IamRoleKindType
	Arn  string // Only set when Kind is IamRoleKindArn
}

// IamRoleKindType represents the type of IAM role.
type IamRoleKindType int

const (
	IamRoleKindNone IamRoleKindType = iota
	IamRoleKindDefault
	IamRoleKindArn
)

func (i IamRoleKind) String() string {
	switch i.Kind {
	case IamRoleKindDefault:
		return "DEFAULT"
	case IamRoleKindArn:
		return fmt.Sprintf("'%s'", escapeSingleQuote(i.Arn))
	default:
		return ""
	}
}

func (i IamRoleKind) exprNode()        {}
func (i IamRoleKind) expr()            {}
func (i IamRoleKind) IsExpr()          {}
func (i IamRoleKind) Span() token.Span { return token.Span{} }

// Partition represents partition.
type Partition struct{}

func (p *Partition) exprNode()        {}
func (p *Partition) expr()            {}
func (p *Partition) IsExpr()          {}
func (p *Partition) Span() token.Span { return token.Span{} }
func (p *Partition) String() string   { return "" }

// Deduplicate represents deduplicate clause.
type Deduplicate struct{}

func (d *Deduplicate) exprNode()        {}
func (d *Deduplicate) expr()            {}
func (d *Deduplicate) IsExpr()          {}
func (d *Deduplicate) Span() token.Span { return token.Span{} }
func (d *Deduplicate) String() string   { return "" }

// HiveLoadDataFormat represents Hive LOAD DATA format.
type HiveLoadDataFormat struct{}

func (h *HiveLoadDataFormat) exprNode()        {}
func (h *HiveLoadDataFormat) expr()            {}
func (h *HiveLoadDataFormat) IsExpr()          {}
func (h *HiveLoadDataFormat) Span() token.Span { return token.Span{} }
func (h *HiveLoadDataFormat) String() string   { return "" }

// FileStagingCommand represents a file staging command (LIST/REMOVE/LS/RM).
type FileStagingCommand struct {
	SpanVal token.Span
	Stage   *ast.ObjectName
	Pattern *string
}

func (f *FileStagingCommand) exprNode()        {}
func (f *FileStagingCommand) expr()            {}
func (f *FileStagingCommand) IsExpr()          {}
func (f *FileStagingCommand) Span() token.Span { return f.SpanVal }
func (f *FileStagingCommand) String() string {
	var sb strings.Builder
	sb.WriteString(f.Stage.String())
	if f.Pattern != nil {
		sb.WriteString(" PATTERN='")
		sb.WriteString(*f.Pattern)
		sb.WriteString("'")
	}
	return sb.String()
}

// PrintStatement represents PRINT statement.
type PrintStatement struct {
	Message string
}

func (p *PrintStatement) exprNode()        {}
func (p *PrintStatement) expr()            {}
func (p *PrintStatement) IsExpr()          {}
func (p *PrintStatement) Span() token.Span { return token.Span{} }
func (p *PrintStatement) String() string   { return p.Message }

// RaisErrorOption represents RAISERROR option.
type RaisErrorOption int

const (
	RaisErrorOptionNone RaisErrorOption = iota
	RaisErrorOptionLog
	RaisErrorOptionNoWait
	RaisErrorOptionSetError
)

func (r RaisErrorOption) String() string {
	switch r {
	case RaisErrorOptionLog:
		return "LOG"
	case RaisErrorOptionNoWait:
		return "NOWAIT"
	case RaisErrorOptionSetError:
		return "SETERROR"
	default:
		return ""
	}
}

// RenameTable represents RENAME TABLE.
type RenameTable struct {
	OldName *ast.ObjectName
	NewName *ast.ObjectName
}

func (r *RenameTable) exprNode()        {}
func (r *RenameTable) expr()            {}
func (r *RenameTable) IsExpr()          {}
func (r *RenameTable) Span() token.Span { return token.Span{} }
func (r *RenameTable) String() string {
	if r.OldName != nil && r.NewName != nil {
		return fmt.Sprintf("%s TO %s", r.OldName.String(), r.NewName.String())
	}
	return ""
}

// ResetStatement represents RESET statement.
type ResetStatement struct {
	ConfigName string
}

func (r *ResetStatement) exprNode()        {}
func (r *ResetStatement) expr()            {}
func (r *ResetStatement) IsExpr()          {}
func (r *ResetStatement) Span() token.Span { return token.Span{} }
func (r *ResetStatement) String() string   { return r.ConfigName }

// ReturnStatement represents RETURN statement.
type ReturnStatement struct {
	Value Expr
}

func (r *ReturnStatement) exprNode()        {}
func (r *ReturnStatement) expr()            {}
func (r *ReturnStatement) IsExpr()          {}
func (r *ReturnStatement) Span() token.Span { return token.Span{} }
func (r *ReturnStatement) String() string {
	if r.Value != nil {
		return "RETURN " + r.Value.String()
	}
	return "RETURN"
}

// ThrowStatement represents THROW statement.
type ThrowStatement struct {
	ErrorNumber int64
	Message     string
	State       int64
}

func (t *ThrowStatement) exprNode()        {}
func (t *ThrowStatement) expr()            {}
func (t *ThrowStatement) IsExpr()          {}
func (t *ThrowStatement) Span() token.Span { return token.Span{} }
func (t *ThrowStatement) String() string {
	return fmt.Sprintf("%d, '%s', %d", t.ErrorNumber, t.Message, t.State)
}

// VacuumStatement represents VACUUM statement.
type VacuumStatement struct {
	TableName *ast.ObjectName
}

func (v *VacuumStatement) exprNode()        {}
func (v *VacuumStatement) expr()            {}
func (v *VacuumStatement) IsExpr()          {}
func (v *VacuumStatement) Span() token.Span { return token.Span{} }
func (v *VacuumStatement) String() string {
	if v.TableName != nil {
		return v.TableName.String()
	}
	return ""
}

// WaitForStatement represents WAITFOR statement.
type WaitForStatement struct {
	Delay     *string
	Time      *string
	Statement ast.Statement
}

func (w *WaitForStatement) exprNode()        {}
func (w *WaitForStatement) expr()            {}
func (w *WaitForStatement) IsExpr()          {}
func (w *WaitForStatement) Span() token.Span { return token.Span{} }
func (w *WaitForStatement) String() string   { return "WAITFOR" }

// ReferentialAction represents referential action (e.g., CASCADE, RESTRICT).
type ReferentialAction int

const (
	ReferentialActionNone ReferentialAction = iota
	ReferentialActionRestrict
	ReferentialActionCascade
	ReferentialActionSetNull
	ReferentialActionSetDefault
	ReferentialActionNoAction
)

func (r ReferentialAction) String() string {
	switch r {
	case ReferentialActionRestrict:
		return "RESTRICT"
	case ReferentialActionCascade:
		return "CASCADE"
	case ReferentialActionSetNull:
		return "SET NULL"
	case ReferentialActionSetDefault:
		return "SET DEFAULT"
	case ReferentialActionNoAction:
		return "NO ACTION"
	default:
		return ""
	}
}

// AssignmentTarget represents assignment target.
type AssignmentTarget struct {
	Column *ast.Ident
}

func (a *AssignmentTarget) exprNode()        {}
func (a *AssignmentTarget) expr()            {}
func (a *AssignmentTarget) IsExpr()          {}
func (a *AssignmentTarget) Span() token.Span { return token.Span{} }
func (a *AssignmentTarget) String() string {
	if a.Column != nil {
		return a.Column.String()
	}
	return ""
}

// MergeInsertKind represents the type of expression used to insert rows within a MERGE statement.
type MergeInsertKind int

const (
	MergeInsertKindNone MergeInsertKind = iota
	MergeInsertKindValues
	MergeInsertKindRow
)

func (m MergeInsertKind) String() string {
	switch m {
	case MergeInsertKindValues:
		return "VALUES"
	case MergeInsertKindRow:
		return "ROW"
	default:
		return ""
	}
}

// MergeInsertExpr represents the expression used to insert rows within a MERGE statement.
// Example: INSERT (product, quantity) VALUES(product, quantity)
type MergeInsertExpr struct {
	InsertToken     *token.Token
	Columns         []*ast.ObjectName
	KindToken       *token.Token
	Kind            MergeInsertKind
	Values          []Expr // For VALUES kind, stores the value expressions
	InsertPredicate Expr
}

func (m *MergeInsertExpr) exprNode()        {}
func (m *MergeInsertExpr) expr()            {}
func (m *MergeInsertExpr) IsExpr()          {}
func (m *MergeInsertExpr) Span() token.Span { return token.Span{} }
func (m *MergeInsertExpr) String() string {
	var f strings.Builder
	if len(m.Columns) > 0 {
		cols := make([]string, len(m.Columns))
		for i, col := range m.Columns {
			cols[i] = col.String()
		}
		f.WriteString("(")
		f.WriteString(strings.Join(cols, ", "))
		f.WriteString(") ")
	}
	f.WriteString(m.Kind.String())
	if m.Kind == MergeInsertKindValues && len(m.Values) > 0 {
		f.WriteString(" (")
		vals := make([]string, len(m.Values))
		for i, v := range m.Values {
			vals[i] = v.String()
		}
		f.WriteString(strings.Join(vals, ", "))
		f.WriteString(")")
	}
	if m.InsertPredicate != nil {
		f.WriteString(" WHERE ")
		f.WriteString(m.InsertPredicate.String())
	}
	return f.String()
}

// MergeUpdateExpr represents the expression used to update rows within a MERGE statement.
// Example: UPDATE SET quantity = T.quantity + S.quantity
type MergeUpdateExpr struct {
	UpdateToken     *token.Token
	Assignments     []*Assignment
	UpdatePredicate Expr
	DeletePredicate Expr
}

func (m *MergeUpdateExpr) exprNode()        {}
func (m *MergeUpdateExpr) expr()            {}
func (m *MergeUpdateExpr) IsExpr()          {}
func (m *MergeUpdateExpr) Span() token.Span { return token.Span{} }
func (m *MergeUpdateExpr) String() string {
	var f strings.Builder
	f.WriteString("SET ")
	assigns := make([]string, len(m.Assignments))
	for i, a := range m.Assignments {
		assigns[i] = a.String()
	}
	f.WriteString(strings.Join(assigns, ", "))
	if m.UpdatePredicate != nil {
		f.WriteString(" WHERE ")
		f.WriteString(m.UpdatePredicate.String())
	}
	if m.DeletePredicate != nil {
		f.WriteString(" DELETE WHERE ")
		f.WriteString(m.DeletePredicate.String())
	}
	return f.String()
}

// MergeAction represents MERGE action as a union type.
type MergeAction struct {
	Insert *MergeInsertExpr
	Update *MergeUpdateExpr
	Delete *token.Token // non-nil if delete action
}

func (m *MergeAction) exprNode()        {}
func (m *MergeAction) expr()            {}
func (m *MergeAction) IsExpr()          {}
func (m *MergeAction) Span() token.Span { return token.Span{} }
func (m *MergeAction) String() string {
	if m.Insert != nil {
		return "INSERT " + m.Insert.String()
	}
	if m.Update != nil {
		return "UPDATE " + m.Update.String()
	}
	if m.Delete != nil {
		return "DELETE"
	}
	return ""
}

// MergeClauseKind represents MERGE clause kind.
type MergeClauseKind int

const (
	MergeClauseKindNone MergeClauseKind = iota
	MergeClauseKindMatched
	MergeClauseKindNotMatched
	MergeClauseKindNotMatchedByTarget
	MergeClauseKindNotMatchedBySource
)

func (m MergeClauseKind) String() string {
	switch m {
	case MergeClauseKindMatched:
		return "MATCHED"
	case MergeClauseKindNotMatched:
		return "NOT MATCHED"
	case MergeClauseKindNotMatchedByTarget:
		return "NOT MATCHED BY TARGET"
	case MergeClauseKindNotMatchedBySource:
		return "NOT MATCHED BY SOURCE"
	default:
		return ""
	}
}

// IndexType represents index type.
type IndexType int

const (
	IndexTypeNone IndexType = iota
	IndexTypeBTree
	IndexTypeHash
)

func (i IndexType) String() string {
	switch i {
	case IndexTypeBTree:
		return "BTREE"
	case IndexTypeHash:
		return "HASH"
	default:
		return ""
	}
}

// Privilege represents a privilege for GRANT/REVOKE.
type Privilege struct {
	Name string
}

func (p *Privilege) exprNode()        {}
func (p *Privilege) expr()            {}
func (p *Privilege) IsExpr()          {}
func (p *Privilege) Span() token.Span { return token.Span{} }
func (p *Privilege) String() string   { return p.Name }

// ============================================================================
// CREATE TABLE Extensions
// ============================================================================

// CreateTableOnCommit represents ON COMMIT clause for CREATE TABLE
type CreateTableOnCommit struct {
	Action string // PRESERVE ROWS, DELETE ROWS, DROP
}

func (c *CreateTableOnCommit) exprNode()        {}
func (c *CreateTableOnCommit) expr()            {}
func (c *CreateTableOnCommit) IsExpr()          {}
func (c *CreateTableOnCommit) Span() token.Span { return token.Span{} }
func (c *CreateTableOnCommit) String() string {
	if c.Action != "" {
		return "ON COMMIT " + c.Action
	}
	return ""
}

// DistStyle represents Redshift DISTSTYLE clause
type DistStyle struct {
	Style string // ALL, EVEN, KEY, AUTO
}

func (d *DistStyle) exprNode()        {}
func (d *DistStyle) expr()            {}
func (d *DistStyle) IsExpr()          {}
func (d *DistStyle) Span() token.Span { return token.Span{} }
func (d *DistStyle) String() string {
	return "DISTSTYLE " + d.Style
}
