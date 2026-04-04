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
	"github.com/user/sqlparser/span"
	"github.com/user/sqlparser/tokenizer"
)

// ============================================================================
// ColumnDef - Column definition for CREATE TABLE
// ============================================================================

// ColumnDef represents a column definition in a CREATE TABLE statement.
type ColumnDef struct {
	Name     *ast.Ident
	DataType interface{} // datatype.DataType - using interface{} to avoid import cycle
	Options  []*ColumnOptionDef
	SpanVal  span.Span
}

func (c *ColumnDef) exprNode() {}

// Span returns the source span for this expression.
func (c *ColumnDef) Span() span.Span {
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
	SpanVal    span.Span
}

func (t *TableConstraint) exprNode() {}

// Span returns the source span for this expression.
func (t *TableConstraint) Span() span.Span {
	return t.SpanVal
}

// String returns the SQL representation.
func (t *TableConstraint) String() string {
	return "CONSTRAINT"
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

// Span returns the source span for this expression.
func (h *HiveDistributionStyle) Span() span.Span {
	return span.Span{}
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

// Span returns the source span for this expression.
func (h *HiveFormat) Span() span.Span {
	return span.Span{}
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

// Span returns the source span for this expression.
func (c *CreateTableOptions) Span() span.Span {
	return span.Span{}
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

// Span returns the source span for this expression.
func (c *CreateTableLikeKind) Span() span.Span {
	return span.Span{}
}

// String returns the SQL representation.
func (c *CreateTableLikeKind) String() string {
	if c.Name != nil {
		return fmt.Sprintf("LIKE %s", c.Name.String())
	}
	return "LIKE"
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

// Span returns the source span for this expression.
func (t *TableVersion) Span() span.Span {
	return span.Span{}
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

// Span returns the source span for this expression.
func (c *CommentDef) Span() span.Span {
	return span.Span{}
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

// ColumnOptionDef represents a column option definition.
type ColumnOptionDef struct {
	Name  string
	Value Expr
}

func (c *ColumnOptionDef) String() string {
	if c.Value != nil {
		return fmt.Sprintf("%s %s", c.Name, c.Value.String())
	}
	return c.Name
}

// ============================================================================
// Additional DDL Types (Stubs for compilation)
// ============================================================================

// OneOrManyWithParens represents one or many items with parentheses.
type OneOrManyWithParens struct {
	Items []Expr
}

func (o *OneOrManyWithParens) exprNode()       {}
func (o *OneOrManyWithParens) Span() span.Span { return span.Span{} }
func (o *OneOrManyWithParens) String() string  { return "(...)" }

// WrappedCollection represents a wrapped collection of items.
type WrappedCollection struct {
	Items []Expr
}

func (w *WrappedCollection) exprNode()       {}
func (w *WrappedCollection) Span() span.Span { return span.Span{} }
func (w *WrappedCollection) String() string  { return "(...)" }

// ClusteredBy represents CLUSTER BY clause.
type ClusteredBy struct {
	Columns []*ast.Ident
}

func (c *ClusteredBy) exprNode()       {}
func (c *ClusteredBy) Span() span.Span { return span.Span{} }
func (c *ClusteredBy) String() string  { return "CLUSTERED BY" }

// ForValues represents FOR VALUES clause.
type ForValues struct {
	Values []Expr
}

func (f *ForValues) exprNode()       {}
func (f *ForValues) Span() span.Span { return span.Span{} }
func (f *ForValues) String() string  { return "FOR VALUES" }

// RowAccessPolicy represents row access policy.
type RowAccessPolicy struct {
	Name *ast.ObjectName
}

func (r *RowAccessPolicy) exprNode()       {}
func (r *RowAccessPolicy) Span() span.Span { return span.Span{} }
func (r *RowAccessPolicy) String() string  { return "ROW ACCESS POLICY" }

// StorageLifecyclePolicy represents storage lifecycle policy.
type StorageLifecyclePolicy struct {
	Name string
}

func (s *StorageLifecyclePolicy) exprNode()       {}
func (s *StorageLifecyclePolicy) Span() span.Span { return span.Span{} }
func (s *StorageLifecyclePolicy) String() string  { return "STORAGE LIFECYCLE POLICY" }

// Tag represents a tag.
type Tag struct {
	Name  string
	Value string
}

func (t *Tag) exprNode()       {}
func (t *Tag) Span() span.Span { return span.Span{} }
func (t *Tag) String() string  { return fmt.Sprintf("%s=%s", t.Name, t.Value) }

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
)

func (i InitializeKind) String() string { return "" }

// ViewEnvelope represents view envelope.
type ViewEnvelope struct{}

func (v *ViewEnvelope) exprNode()       {}
func (v *ViewEnvelope) Span() span.Span { return span.Span{} }
func (v *ViewEnvelope) String() string  { return "" }

// CreateViewParams represents CREATE VIEW parameters.
type CreateViewParams struct{}

func (c *CreateViewParams) exprNode()       {}
func (c *CreateViewParams) Span() span.Span { return span.Span{} }
func (c *CreateViewParams) String() string  { return "" }

// IndexColumn represents an index column.
type IndexColumn struct {
	Expr       Expr
	Opclass    *ast.ObjectName
	Asc        *bool // nil means not specified
	NullsFirst *bool // nil means not specified
}

func (i *IndexColumn) exprNode()       {}
func (i *IndexColumn) Span() span.Span { return span.Span{} }
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

// OperateFunctionArg represents operate function argument.
type OperateFunctionArg struct{}

func (o *OperateFunctionArg) exprNode()       {}
func (o *OperateFunctionArg) Span() span.Span { return span.Span{} }
func (o *OperateFunctionArg) String() string  { return "" }

// FunctionReturnType represents function return type.
type FunctionReturnType struct{}

func (f *FunctionReturnType) exprNode()       {}
func (f *FunctionReturnType) Span() span.Span { return span.Span{} }
func (f *FunctionReturnType) String() string  { return "" }

// FunctionBehavior represents function behavior.
type FunctionBehavior int

const (
	FunctionBehaviorNone FunctionBehavior = iota
)

func (f FunctionBehavior) String() string { return "" }

// FunctionCalledOnNull represents function called on null.
type FunctionCalledOnNull int

const (
	FunctionCalledOnNullNone FunctionCalledOnNull = iota
)

func (f FunctionCalledOnNull) String() string { return "" }

// FunctionParallel represents function parallel.
type FunctionParallel int

const (
	FunctionParallelNone FunctionParallel = iota
)

func (f FunctionParallel) String() string { return "" }

// FunctionSecurity represents function security.
type FunctionSecurity int

const (
	FunctionSecurityNone FunctionSecurity = iota
)

func (f FunctionSecurity) String() string { return "" }

// FunctionDeterminismSpecifier represents function determinism specifier.
type FunctionDeterminismSpecifier int

const (
	FunctionDeterminismSpecifierNone FunctionDeterminismSpecifier = iota
)

func (f FunctionDeterminismSpecifier) String() string { return "" }

// CreateFunctionBody represents function body.
type CreateFunctionBody struct{}

func (c *CreateFunctionBody) exprNode()       {}
func (c *CreateFunctionBody) Span() span.Span { return span.Span{} }
func (c *CreateFunctionBody) String() string  { return "" }

// FunctionDefinitionSetParam represents function definition set parameter.
type FunctionDefinitionSetParam struct{}

func (f *FunctionDefinitionSetParam) exprNode()       {}
func (f *FunctionDefinitionSetParam) Span() span.Span { return span.Span{} }
func (f *FunctionDefinitionSetParam) String() string  { return "" }

// SqlSecurity represents SQL security.
type SqlSecurity int

const (
	SqlSecurityNone SqlSecurity = iota
)

func (s SqlSecurity) String() string { return "" }

// RemoteProperty represents remote property.
type RemoteProperty struct{}

func (r *RemoteProperty) exprNode()       {}
func (r *RemoteProperty) Span() span.Span { return span.Span{} }
func (r *RemoteProperty) String() string  { return "" }

// ProcedureParam represents procedure parameter.
type ProcedureParam struct{}

func (p *ProcedureParam) exprNode()       {}
func (p *ProcedureParam) Span() span.Span { return span.Span{} }
func (p *ProcedureParam) String() string  { return "" }

// ExecuteAs represents EXECUTE AS clause.
type ExecuteAs int

const (
	ExecuteAsNone ExecuteAs = iota
)

func (e ExecuteAs) String() string { return "" }

// RoleOption represents role option.
type RoleOption struct{}

func (r *RoleOption) exprNode()       {}
func (r *RoleOption) Span() span.Span { return span.Span{} }
func (r *RoleOption) String() string  { return "" }

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

	// Fields for RenameColumn
	RenameOldColumn *ast.Ident
	RenameNewColumn *ast.Ident

	// Fields for RenameTable
	NewTableName *ast.ObjectName

	// Fields for AlterColumn
	AlterColumnName *ast.Ident
	AlterColumnOp   AlterColumnOpType
	AlterDefault    Expr
	AlterDataType   interface{} // datatype.DataType

	// Fields for SetTblProperties
	TblProperties []*SqlOption

	// Fields for SetOptions (MySQL: AUTO_INCREMENT, ALGORITHM, LOCK)
	AutoIncrementValue string     // For AUTO_INCREMENT = N
	AlgorithmValue     *ast.Ident // For ALGORITHM = {COPY|INPLACE|INSTANT}
	LockValue          *ast.Ident // For LOCK = {DEFAULT|NONE|SHARED|EXCLUSIVE}

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

	// Span
	SpanVal span.Span
}

// AlterTableOpType represents the type of ALTER TABLE operation
type AlterTableOpType int

const (
	AlterTableOpAddColumn AlterTableOpType = iota
	AlterTableOpDropColumn
	AlterTableOpAddConstraint
	AlterTableOpDropConstraint
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
	AlterTableOpValidateConstraint
)

// AlterColumnOpType represents operations on a column via ALTER COLUMN
type AlterColumnOpType int

const (
	AlterColumnOpSetNotNull AlterColumnOpType = iota
	AlterColumnOpDropNotNull
	AlterColumnOpSetDefault
	AlterColumnOpDropDefault
	AlterColumnOpSetDataType
)

func (a *AlterTableOperation) exprNode()       {}
func (a *AlterTableOperation) Span() span.Span { return a.SpanVal }
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
		buf.WriteString("RENAME TO ")
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
			buf.WriteString(" SET DATA TYPE")
			if a.AlterDataType != nil {
				if dt, ok := a.AlterDataType.(fmt.Stringer); ok {
					buf.WriteString(" ")
					buf.WriteString(dt.String())
				}
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

func (h *HiveSetLocation) exprNode()       {}
func (h *HiveSetLocation) Span() span.Span { return span.Span{} }
func (h *HiveSetLocation) String() string  { return "" }

// AlterIndexOperation represents ALTER INDEX operation.
type AlterIndexOperation struct{}

func (a *AlterIndexOperation) exprNode()       {}
func (a *AlterIndexOperation) Span() span.Span { return span.Span{} }
func (a *AlterIndexOperation) String() string  { return "" }

// AlterSchemaOperation represents ALTER SCHEMA operation.
type AlterSchemaOperation struct{}

func (a *AlterSchemaOperation) exprNode()       {}
func (a *AlterSchemaOperation) Span() span.Span { return span.Span{} }
func (a *AlterSchemaOperation) String() string  { return "" }

// AlterTypeOperation represents ALTER TYPE operation.
type AlterTypeOperation struct{}

func (a *AlterTypeOperation) exprNode()       {}
func (a *AlterTypeOperation) Span() span.Span { return span.Span{} }
func (a *AlterTypeOperation) String() string  { return "" }

// AlterRoleOperation represents ALTER ROLE operation.
type AlterRoleOperation struct{}

func (a *AlterRoleOperation) exprNode()       {}
func (a *AlterRoleOperation) Span() span.Span { return span.Span{} }
func (a *AlterRoleOperation) String() string  { return "" }

// ObjectType represents object type.
type ObjectType int

const (
	ObjectTypeNone ObjectType = iota
)

func (o ObjectType) String() string { return "" }

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

func (s *SchemaName) exprNode()       {}
func (s *SchemaName) Span() span.Span { return span.Span{} }

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

func (c *ContactEntry) exprNode()       {}
func (c *ContactEntry) Span() span.Span { return span.Span{} }
func (c *ContactEntry) String() string  { return "" }

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
	SpanVal span.Span
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

func (s *SequenceOptions) exprNode()       {}
func (s *SequenceOptions) Span() span.Span { return s.SpanVal }
func (s *SequenceOptions) String() string {
	switch s.Type {
	case SeqOptIncrementBy:
		if s.HasByOrWith {
			return fmt.Sprintf(" INCREMENT BY %s", s.Expr.String())
		}
		return fmt.Sprintf(" INCREMENT %s", s.Expr.String())
	case SeqOptMinValue:
		if s.NoValue {
			return " NO MINVALUE"
		}
		return fmt.Sprintf(" MINVALUE %s", s.Expr.String())
	case SeqOptMaxValue:
		if s.NoValue {
			return " NO MAXVALUE"
		}
		return fmt.Sprintf(" MAXVALUE %s", s.Expr.String())
	case SeqOptStartWith:
		if s.HasByOrWith {
			return fmt.Sprintf(" START WITH %s", s.Expr.String())
		}
		return fmt.Sprintf(" START %s", s.Expr.String())
	case SeqOptCache:
		return fmt.Sprintf(" CACHE %s", s.Expr.String())
	case SeqOptCycle:
		if s.NoCycle {
			return " NO CYCLE"
		}
		return " CYCLE"
	}
	return ""
}

// DomainConstraint represents domain constraint.
type DomainConstraint struct{}

func (d *DomainConstraint) exprNode()       {}
func (d *DomainConstraint) Span() span.Span { return span.Span{} }
func (d *DomainConstraint) String() string  { return "" }

// UserDefinedTypeRepresentation represents user-defined type representation.
type UserDefinedTypeRepresentation struct{}

func (u *UserDefinedTypeRepresentation) exprNode()       {}
func (u *UserDefinedTypeRepresentation) Span() span.Span { return span.Span{} }
func (u *UserDefinedTypeRepresentation) String() string  { return "" }

// TriggerPeriod represents trigger period.
type TriggerPeriod int

const (
	TriggerPeriodNone TriggerPeriod = iota
)

func (t TriggerPeriod) String() string { return "" }

// TriggerEvent represents trigger event.
type TriggerEvent int

const (
	TriggerEventNone TriggerEvent = iota
)

func (t TriggerEvent) String() string { return "" }

// TriggerReferencing represents trigger referencing.
type TriggerReferencing struct{}

func (t *TriggerReferencing) exprNode()       {}
func (t *TriggerReferencing) Span() span.Span { return span.Span{} }
func (t *TriggerReferencing) String() string  { return "" }

// TriggerExecBody represents trigger execution body.
type TriggerExecBody struct{}

func (t *TriggerExecBody) exprNode()       {}
func (t *TriggerExecBody) Span() span.Span { return span.Span{} }
func (t *TriggerExecBody) String() string  { return "" }

// ConditionalStatements represents conditional statements.
type ConditionalStatements struct{}

func (c *ConditionalStatements) exprNode()       {}
func (c *ConditionalStatements) Span() span.Span { return span.Span{} }
func (c *ConditionalStatements) String() string  { return "" }

// MacroArg represents macro argument.
type MacroArg struct{}

func (m *MacroArg) exprNode()       {}
func (m *MacroArg) Span() span.Span { return span.Span{} }
func (m *MacroArg) String() string  { return "" }

// MacroDefinition represents macro definition.
type MacroDefinition struct{}

func (m *MacroDefinition) exprNode()       {}
func (m *MacroDefinition) Span() span.Span { return span.Span{} }
func (m *MacroDefinition) String() string  { return "" }

// StageParamsObject represents stage parameters object.
type StageParamsObject struct {
	Url                *string
	Encryption         *KeyValueOptions
	Endpoint           *string
	StorageIntegration *string
	Credentials        *KeyValueOptions
}

func (s *StageParamsObject) exprNode()       {}
func (s *StageParamsObject) Span() span.Span { return span.Span{} }
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
}

func (k *KeyValueOption) String() string {
	switch k.Kind {
	case KeyValueOptionKindSingle:
		if val, ok := k.OptionValue.(fmt.Stringer); ok {
			return fmt.Sprintf("%s=%s", k.OptionName, val.String())
		}
		if val, ok := k.OptionValue.(string); ok {
			return fmt.Sprintf("%s='%s'", k.OptionName, escapeSingleQuote(val))
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

func (k *KeyValueOptions) exprNode()       {}
func (k *KeyValueOptions) Span() span.Span { return span.Span{} }
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

func (s *SecretOption) exprNode()       {}
func (s *SecretOption) Span() span.Span { return span.Span{} }
func (s *SecretOption) String() string  { return "" }

// CreatePolicyCommand represents CREATE POLICY command.
type CreatePolicyCommand int

const (
	CreatePolicyCommandNone CreatePolicyCommand = iota
)

func (c CreatePolicyCommand) String() string { return "" }

// RoleName represents role name.
type RoleName struct {
	Name string
}

func (r *RoleName) exprNode()       {}
func (r *RoleName) Span() span.Span { return span.Span{} }
func (r *RoleName) String() string  { return r.Name }

// OperatorPurpose represents operator purpose.
type OperatorPurpose int

const (
	OperatorPurposeNone OperatorPurpose = iota
)

func (o OperatorPurpose) String() string { return "" }

// OperatorOption represents operator option.
type OperatorOption struct{}

func (o *OperatorOption) exprNode()       {}
func (o *OperatorOption) Span() span.Span { return span.Span{} }
func (o *OperatorOption) String() string  { return "" }

// OperatorClassItem represents operator class item.
type OperatorClassItem struct{}

func (o *OperatorClassItem) exprNode()       {}
func (o *OperatorClassItem) Span() span.Span { return span.Span{} }
func (o *OperatorClassItem) String() string  { return "" }

// AlterPolicyOperation represents ALTER POLICY operation.
type AlterPolicyOperation struct{}

func (a *AlterPolicyOperation) exprNode()       {}
func (a *AlterPolicyOperation) Span() span.Span { return span.Span{} }
func (a *AlterPolicyOperation) String() string  { return "" }

// AlterConnectorOwner represents ALTER CONNECTOR owner.
type AlterConnectorOwner struct{}

func (a *AlterConnectorOwner) exprNode()       {}
func (a *AlterConnectorOwner) Span() span.Span { return span.Span{} }
func (a *AlterConnectorOwner) String() string  { return "" }

// AttachDuckDBDatabaseOption represents ATTACH DuckDB database option.
type AttachDuckDBDatabaseOption struct{}

func (a *AttachDuckDBDatabaseOption) exprNode()       {}
func (a *AttachDuckDBDatabaseOption) Span() span.Span { return span.Span{} }
func (a *AttachDuckDBDatabaseOption) String() string  { return "" }

// FunctionDesc represents function description.
type FunctionDesc struct{}

func (f *FunctionDesc) exprNode()       {}
func (f *FunctionDesc) Span() span.Span { return span.Span{} }
func (f *FunctionDesc) String() string  { return "" }

// DropOperatorSignature represents DROP OPERATOR signature.
type DropOperatorSignature struct{}

func (d *DropOperatorSignature) exprNode()       {}
func (d *DropOperatorSignature) Span() span.Span { return span.Span{} }
func (d *DropOperatorSignature) String() string  { return "" }

// OperatorSignature represents operator signature.
type OperatorSignature struct{}

func (o *OperatorSignature) exprNode()       {}
func (o *OperatorSignature) Span() span.Span { return span.Span{} }
func (o *OperatorSignature) String() string  { return "" }

// AlterOperatorOperation represents ALTER OPERATOR operation.
type AlterOperatorOperation struct{}

func (a *AlterOperatorOperation) exprNode()       {}
func (a *AlterOperatorOperation) Span() span.Span { return span.Span{} }
func (a *AlterOperatorOperation) String() string  { return "" }

// OperatorFamilyOperation represents operator family operation.
type OperatorFamilyOperation struct{}

func (o *OperatorFamilyOperation) exprNode()       {}
func (o *OperatorFamilyOperation) Span() span.Span { return span.Span{} }
func (o *OperatorFamilyOperation) String() string  { return "" }

// OperatorClassOperation represents operator class operation.
type OperatorClassOperation struct{}

func (o *OperatorClassOperation) exprNode()       {}
func (o *OperatorClassOperation) Span() span.Span { return span.Span{} }
func (o *OperatorClassOperation) String() string  { return "" }

// OptimizerHint represents optimizer hint.
type OptimizerHint struct{}

func (o *OptimizerHint) exprNode()       {}
func (o *OptimizerHint) Span() span.Span { return span.Span{} }
func (o *OptimizerHint) String() string  { return "" }

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

func (a *Assignment) exprNode()       {}
func (a *Assignment) Span() span.Span { return span.Span{} }
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

func (o *OnInsert) exprNode()       {}
func (o *OnInsert) Span() span.Span { return span.Span{} }
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
	var parts []string
	parts = append(parts, "ON CONFLICT")
	if o.ConflictTarget != nil {
		parts = append(parts, o.ConflictTarget.String())
	}
	parts = append(parts, "DO")
	parts = append(parts, o.Action.String())
	return strings.Join(parts, " ")
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
	OutputToken    *tokenizer.Token
	ReturningToken *tokenizer.Token // For RETURNING variant
	SelectItems    []query.SelectItem
	IntoTable      *query.SelectInto
}

func (o *OutputClause) exprNode()       {}
func (o *OutputClause) Span() span.Span { return span.Span{} }
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

func (i *InsertAliases) exprNode()       {}
func (i *InsertAliases) Span() span.Span { return span.Span{} }
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

func (s *Setting) exprNode()       {}
func (s *Setting) Span() span.Span { return span.Span{} }
func (s *Setting) String() string {
	return fmt.Sprintf("%s = %s", s.Name, s.Value.String())
}

// InputFormatClause represents input format clause.
type InputFormatClause struct{}

func (i *InputFormatClause) exprNode()       {}
func (i *InputFormatClause) Span() span.Span { return span.Span{} }
func (i *InputFormatClause) String() string  { return "" }

// MultiTableInsertType represents multi-table insert type.
type MultiTableInsertType int

const (
	MultiTableInsertTypeNone MultiTableInsertType = iota
)

func (m MultiTableInsertType) String() string { return "" }

// MultiTableInsertIntoClause represents multi-table INSERT INTO clause.
type MultiTableInsertIntoClause struct{}

func (m *MultiTableInsertIntoClause) exprNode()       {}
func (m *MultiTableInsertIntoClause) Span() span.Span { return span.Span{} }
func (m *MultiTableInsertIntoClause) String() string  { return "" }

// MultiTableInsertWhenClause represents multi-table INSERT WHEN clause.
type MultiTableInsertWhenClause struct{}

func (m *MultiTableInsertWhenClause) exprNode()       {}
func (m *MultiTableInsertWhenClause) Span() span.Span { return span.Span{} }
func (m *MultiTableInsertWhenClause) String() string  { return "" }

// MergeClause represents a WHEN clause within a MERGE statement.
// Example: WHEN NOT MATCHED BY SOURCE AND product LIKE '%washer%' THEN DELETE
type MergeClause struct {
	WhenToken  *tokenizer.Token
	ClauseKind MergeClauseKind
	Predicate  Expr
	Action     *MergeAction
}

func (m *MergeClause) exprNode()       {}
func (m *MergeClause) Span() span.Span { return span.Span{} }
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

func (c *CaseStatementWhen) exprNode()       {}
func (c *CaseStatementWhen) Span() span.Span { return span.Span{} }
func (c *CaseStatementWhen) String() string  { return "" }

// CaseStatementElse represents CASE statement ELSE clause.
type CaseStatementElse struct{}

func (c *CaseStatementElse) exprNode()       {}
func (c *CaseStatementElse) Span() span.Span { return span.Span{} }
func (c *CaseStatementElse) String() string  { return "" }

// IfStatementCondition represents IF statement condition.
type IfStatementCondition struct{}

func (i *IfStatementCondition) exprNode()       {}
func (i *IfStatementCondition) Span() span.Span { return span.Span{} }
func (i *IfStatementCondition) String() string  { return "" }

// IfStatementElse represents IF statement ELSE clause.
type IfStatementElse struct{}

func (i *IfStatementElse) exprNode()       {}
func (i *IfStatementElse) Span() span.Span { return span.Span{} }
func (i *IfStatementElse) String() string  { return "" }

// RaiseLevel represents RAISE level.
type RaiseLevel int

const (
	RaiseLevelNone RaiseLevel = iota
)

func (r RaiseLevel) String() string { return "" }

// RaiseUsing represents RAISE USING clause.
type RaiseUsing struct{}

func (r *RaiseUsing) exprNode()       {}
func (r *RaiseUsing) Span() span.Span { return span.Span{} }
func (r *RaiseUsing) String() string  { return "" }

// CopySource represents the source for a COPY command: a table or a query.
type CopySource struct {
	TableName *ast.ObjectName
	Columns   []*ast.Ident
	Query     interface{} // *query.Query - using interface{} to avoid import cycle
	SpanVal   span.Span
}

func (c *CopySource) exprNode() {}

// Span returns the source span for this expression.
func (c *CopySource) Span() span.Span { return c.SpanVal }

// String returns the SQL representation.
func (c *CopySource) String() string {
	if c.Query != nil {
		// Use fmt.Sprintf with %v to call String() method if available
		return fmt.Sprintf(" (%v)", c.Query)
	}
	if c.TableName != nil {
		var parts []string
		parts = append(parts, c.TableName.String())
		if len(c.Columns) > 0 {
			colStrs := make([]string, len(c.Columns))
			for i, col := range c.Columns {
				colStrs[i] = col.String()
			}
			parts = append(parts, fmt.Sprintf("(%s)", strings.Join(colStrs, ", ")))
		}
		return " " + strings.Join(parts, " ")
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
	SpanVal  span.Span
}

func (c *CopyTarget) exprNode() {}

// Span returns the source span for this expression.
func (c *CopyTarget) Span() span.Span { return c.SpanVal }

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

// Span returns the source span for this expression.
func (c *CopyOption) Span() span.Span { return span.Span{} }

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

// Span returns the source span for this expression.
func (c *CopyLegacyOption) Span() span.Span { return span.Span{} }

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
	case CopyLegacyOptionGzip:
		return "GZIP"
	case CopyLegacyOptionHeader:
		return "HEADER"
	case CopyLegacyOptionIamRole:
		if val, ok := c.Value.(string); ok {
			return fmt.Sprintf("IAM_ROLE '%s'", escapeSingleQuote(val))
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

func (s *StageLoadSelectItem) exprNode()       {}
func (s *StageLoadSelectItem) Span() span.Span { return span.Span{} }
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
		parts = append(parts, "AS "+s.ItemAs.String())
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

func (s *StageLoadSelectItemWrapper) exprNode()       {}
func (s *StageLoadSelectItemWrapper) Span() span.Span { return span.Span{} }
func (s *StageLoadSelectItemWrapper) String() string {
	if s.Item != nil {
		if str, ok := s.Item.(fmt.Stringer); ok {
			return str.String()
		}
	}
	return ""
}

// CloseCursor represents CLOSE CURSOR.
type CloseCursor struct{}

func (c *CloseCursor) exprNode()       {}
func (c *CloseCursor) Span() span.Span { return span.Span{} }
func (c *CloseCursor) String() string  { return "" }

// Declare represents DECLARE statement.
type Declare struct{}

func (d *Declare) exprNode()       {}
func (d *Declare) Span() span.Span { return span.Span{} }
func (d *Declare) String() string  { return "" }

// FetchDirection represents FETCH direction.
type FetchDirection int

const (
	FetchDirectionNone FetchDirection = iota
)

func (f FetchDirection) String() string { return "" }

// FetchPosition represents FETCH position.
type FetchPosition int

const (
	FetchPositionNone FetchPosition = iota
)

func (f FetchPosition) String() string { return "" }

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
)

func (d DiscardObject) String() string { return "" }

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

func (s *ShowStatementIn) exprNode()       {}
func (s *ShowStatementIn) Span() span.Span { return span.Span{} }
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
	Like  *string
	Where Expr
}

func (s *ShowStatementFilter) exprNode()       {}
func (s *ShowStatementFilter) Span() span.Span { return span.Span{} }
func (s *ShowStatementFilter) String() string {
	if s.Like != nil {
		return fmt.Sprintf("LIKE '%s'", *s.Like)
	}
	if s.Where != nil {
		return fmt.Sprintf("WHERE %s", s.Where.String())
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
	SpanVal        span.Span
	ShowIn         *ShowStatementIn
	Filter         *ShowStatementFilter
	FilterPosition ShowStatementFilterPosition
	LimitFrom      *string
	Limit          *int
	StartsWith     *string
}

func (s *ShowStatementOptions) exprNode()       {}
func (s *ShowStatementOptions) Span() span.Span { return s.SpanVal }
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

func (e *ExceptionWhen) exprNode()       {}
func (e *ExceptionWhen) Span() span.Span { return span.Span{} }
func (e *ExceptionWhen) String() string  { return "" }

// CommentObject represents COMMENT object.
type CommentObject int

const (
	CommentObjectNone CommentObject = iota
)

func (c CommentObject) String() string { return "" }

// ExprWithAlias represents expression with alias.
type ExprWithAlias struct {
	Expr  Expr
	Alias *ast.Ident
}

func (e *ExprWithAlias) exprNode()       {}
func (e *ExprWithAlias) Span() span.Span { return span.Span{} }
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

func (u *UtilityOption) exprNode()       {}
func (u *UtilityOption) Span() span.Span { return span.Span{} }
func (u *UtilityOption) String() string  { return "" }

// ValueWithSpan represents value with span.
type ValueWithSpan struct {
	Value string
}

func (v *ValueWithSpan) exprNode()       {}
func (v *ValueWithSpan) Span() span.Span { return span.Span{} }
func (v *ValueWithSpan) String() string  { return v.Value }

// LockTable represents a table to lock.
type LockTable struct {
	Table    *ast.Ident
	Alias    *ast.Ident
	LockType LockTableType
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

func (l *LockTable) exprNode()       {}
func (l *LockTable) Span() span.Span { return span.Span{} }
func (l *LockTable) String() string {
	return fmt.Sprintf("%s", l.Table.String())
}

// LockMode represents lock mode.
type LockMode int

const (
	LockModeNone LockMode = iota
)

func (l LockMode) String() string { return "" }

// IamRoleKind represents IAM role kind.
type IamRoleKind int

const (
	IamRoleKindNone IamRoleKind = iota
)

func (i IamRoleKind) String() string { return "" }

// Partition represents partition.
type Partition struct{}

func (p *Partition) exprNode()       {}
func (p *Partition) Span() span.Span { return span.Span{} }
func (p *Partition) String() string  { return "" }

// Deduplicate represents deduplicate clause.
type Deduplicate struct{}

func (d *Deduplicate) exprNode()       {}
func (d *Deduplicate) Span() span.Span { return span.Span{} }
func (d *Deduplicate) String() string  { return "" }

// HiveLoadDataFormat represents Hive LOAD DATA format.
type HiveLoadDataFormat struct{}

func (h *HiveLoadDataFormat) exprNode()       {}
func (h *HiveLoadDataFormat) Span() span.Span { return span.Span{} }
func (h *HiveLoadDataFormat) String() string  { return "" }

// FileStagingCommand represents file staging command.
type FileStagingCommand struct{}

func (f *FileStagingCommand) exprNode()       {}
func (f *FileStagingCommand) Span() span.Span { return span.Span{} }
func (f *FileStagingCommand) String() string  { return "" }

// PrintStatement represents PRINT statement.
type PrintStatement struct {
	Message string
}

func (p *PrintStatement) exprNode()       {}
func (p *PrintStatement) Span() span.Span { return span.Span{} }
func (p *PrintStatement) String() string  { return p.Message }

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

func (r *RenameTable) exprNode()       {}
func (r *RenameTable) Span() span.Span { return span.Span{} }
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

func (r *ResetStatement) exprNode()       {}
func (r *ResetStatement) Span() span.Span { return span.Span{} }
func (r *ResetStatement) String() string  { return r.ConfigName }

// ReturnStatement represents RETURN statement.
type ReturnStatement struct {
	Value Expr
}

func (r *ReturnStatement) exprNode()       {}
func (r *ReturnStatement) Span() span.Span { return span.Span{} }
func (r *ReturnStatement) String() string {
	if r.Value != nil {
		return r.Value.String()
	}
	return ""
}

// ThrowStatement represents THROW statement.
type ThrowStatement struct {
	ErrorNumber int64
	Message     string
	State       int64
}

func (t *ThrowStatement) exprNode()       {}
func (t *ThrowStatement) Span() span.Span { return span.Span{} }
func (t *ThrowStatement) String() string {
	return fmt.Sprintf("%d, '%s', %d", t.ErrorNumber, t.Message, t.State)
}

// VacuumStatement represents VACUUM statement.
type VacuumStatement struct {
	TableName *ast.ObjectName
}

func (v *VacuumStatement) exprNode()       {}
func (v *VacuumStatement) Span() span.Span { return span.Span{} }
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

func (w *WaitForStatement) exprNode()       {}
func (w *WaitForStatement) Span() span.Span { return span.Span{} }
func (w *WaitForStatement) String() string  { return "WAITFOR" }

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

func (a *AssignmentTarget) exprNode()       {}
func (a *AssignmentTarget) Span() span.Span { return span.Span{} }
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
	InsertToken     *tokenizer.Token
	Columns         []*ast.ObjectName
	KindToken       *tokenizer.Token
	Kind            MergeInsertKind
	Values          []Expr // For VALUES kind, stores the value expressions
	InsertPredicate Expr
}

func (m *MergeInsertExpr) exprNode()       {}
func (m *MergeInsertExpr) Span() span.Span { return span.Span{} }
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
	UpdateToken     *tokenizer.Token
	Assignments     []*Assignment
	UpdatePredicate Expr
	DeletePredicate Expr
}

func (m *MergeUpdateExpr) exprNode()       {}
func (m *MergeUpdateExpr) Span() span.Span { return span.Span{} }
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
	Delete *tokenizer.Token // non-nil if delete action
}

func (m *MergeAction) exprNode()       {}
func (m *MergeAction) Span() span.Span { return span.Span{} }
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

func (p *Privilege) exprNode()       {}
func (p *Privilege) Span() span.Span { return span.Span{} }
func (p *Privilege) String() string  { return p.Name }
