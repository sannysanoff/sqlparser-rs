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

package statement

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/datatype"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/query"
)

// ============================================================================
// ANALYZE
// ============================================================================

// Analyze represents an ANALYZE statement
type Analyze struct {
	BaseStatement
	HasTableKeyword   bool
	TableName         *ast.ObjectName
	Partitions        []expr.Expr
	Columns           []*ast.Ident
	ForColumns        bool
	CacheMetadata     bool
	Noscan            bool
	ComputeStatistics bool
}

func (a *Analyze) statementNode() {}

func (a *Analyze) String() string {
	var f strings.Builder
	f.WriteString("ANALYZE ")
	if a.HasTableKeyword {
		f.WriteString("TABLE ")
	}
	if a.TableName != nil {
		f.WriteString(a.TableName.String())
	}
	return f.String()
}

// ============================================================================
// SET
// ============================================================================

// Set represents a SET statement
type Set struct {
	BaseStatement
	Variable *ast.ObjectName
	Values   []expr.Expr
	Local    bool
	Session  bool
	Global   bool
	TimeZone bool
	HiveVar  bool
	Scope    *expr.SetScope
}

func (s *Set) statementNode() {}

func (s *Set) String() string {
	var f strings.Builder
	f.WriteString("SET ")
	if s.Local {
		f.WriteString("LOCAL ")
	}
	if s.Global {
		f.WriteString("GLOBAL ")
	}
	if s.Session {
		f.WriteString("SESSION ")
	}
	if s.TimeZone {
		f.WriteString("TIME ZONE ")
	} else if s.Variable != nil {
		f.WriteString(s.Variable.String())
		f.WriteString(" = ")
	}

	f.WriteString(formatExprs(s.Values, ", "))

	return f.String()
}

// ============================================================================
// SetNames
// ============================================================================

// SetNames represents a SET NAMES statement (MySQL specific).
// Syntax: SET NAMES charset_name [COLLATE collation_name]
type SetNames struct {
	BaseStatement
	CharsetName   string
	CollationName *string
}

func (s *SetNames) statementNode() {}

func (s *SetNames) String() string {
	var f strings.Builder
	f.WriteString("SET NAMES ")
	f.WriteString(s.CharsetName)
	if s.CollationName != nil {
		f.WriteString(" COLLATE ")
		f.WriteString(*s.CollationName)
	}
	return f.String()
}

// ============================================================================
// MSCK
// ============================================================================

// Msck represents an MSCK (repair table) statement (Hive)
type Msck struct {
	BaseStatement
	TableName        *ast.ObjectName
	RepairPartitions bool
	AddPartitions    bool
	DropPartitions   bool
	SyncPartitions   bool
	PartitionSpec    []expr.Expr
}

func (m *Msck) statementNode() {}

func (m *Msck) String() string {
	var f strings.Builder
	f.WriteString("MSCK ")
	if m.RepairPartitions {
		f.WriteString("REPAIR ")
	}
	if m.AddPartitions {
		f.WriteString("ADD ")
	}
	if m.DropPartitions {
		f.WriteString("DROP ")
	}
	if m.SyncPartitions {
		f.WriteString("SYNC ")
	}
	f.WriteString("TABLE ")
	f.WriteString(m.TableName.String())
	return f.String()
}

// ============================================================================
// INSTALL
// ============================================================================

// Install represents an INSTALL statement (DuckDB)
type Install struct {
	BaseStatement
	ExtensionName *ast.Ident
}

func (i *Install) statementNode() {}

func (i *Install) String() string {
	var f strings.Builder
	f.WriteString("INSTALL ")
	f.WriteString(i.ExtensionName.String())
	return f.String()
}

// ============================================================================
// LOAD
// ============================================================================

// Load represents a LOAD statement (DuckDB)
type Load struct {
	BaseStatement
	ExtensionName *ast.Ident
}

func (l *Load) statementNode() {}

func (l *Load) String() string {
	var f strings.Builder
	f.WriteString("LOAD ")
	f.WriteString(l.ExtensionName.String())
	return f.String()
}

// ============================================================================
// DIRECTORY
// ============================================================================

// Directory represents a DIRECTORY statement
type Directory struct {
	BaseStatement
	Overwrite  bool
	Local      bool
	Path       string
	FileFormat *expr.FileFormat
	Source     *query.Query
}

func (d *Directory) statementNode() {}

func (d *Directory) String() string {
	var f strings.Builder
	f.WriteString("LOAD DATA ")
	if d.Local {
		f.WriteString("LOCAL ")
	}
	f.WriteString("INPATH '")
	f.WriteString(d.Path)
	f.WriteString("' ")
	if d.Overwrite {
		f.WriteString("OVERWRITE ")
	}
	f.WriteString("INTO TABLE ...")
	return f.String()
}

// ============================================================================
// CASE STATEMENT
// ============================================================================

// CaseStatement represents a CASE statement
type CaseStatement struct {
	BaseStatement
	Expression expr.Expr
	Whens      []*expr.CaseStatementWhen
	Else       *expr.CaseStatementElse
}

func (c *CaseStatement) statementNode() {}

func (c *CaseStatement) String() string {
	var f strings.Builder
	f.WriteString("CASE ")
	if c.Expression != nil {
		f.WriteString(c.Expression.String())
		f.WriteString(" ")
	}
	for _, when := range c.Whens {
		f.WriteString(when.String())
		f.WriteString(" ")
	}
	if c.Else != nil {
		f.WriteString(c.Else.String())
		f.WriteString(" ")
	}
	f.WriteString("END CASE")
	return f.String()
}

// ============================================================================
// IF STATEMENT
// ============================================================================

// IfStatement represents an IF statement
// Reference: src/ast/mod.rs:2609
type IfStatement struct {
	BaseStatement
	Conditions []*expr.IfStatementCondition // IF and ELSEIF conditions
	Else       *expr.IfStatementElse        // Optional ELSE clause
}

func (i *IfStatement) statementNode() {}

func (i *IfStatement) String() string {
	var sb strings.Builder
	for idx, cond := range i.Conditions {
		if idx == 0 {
			sb.WriteString("IF ")
		} else {
			sb.WriteString(" ELSEIF ")
		}
		sb.WriteString(cond.String())
	}
	if i.Else != nil {
		sb.WriteString(i.Else.String())
	}
	sb.WriteString(" END IF")
	return sb.String()
}

// ============================================================================
// WHILE STATEMENT
// ============================================================================

// WhileStatement represents a WHILE statement
type WhileStatement struct {
	BaseStatement
	Condition  expr.Expr
	Statements []ast.Statement
	Label      *string
}

func (w *WhileStatement) statementNode() {}

func (w *WhileStatement) String() string {
	var f strings.Builder
	if w.Label != nil {
		f.WriteString(*w.Label)
		f.WriteString(": ")
	}
	f.WriteString("WHILE ")
	f.WriteString(w.Condition.String())
	f.WriteString(" DO ... END WHILE")
	return f.String()
}

// ============================================================================
// RAISE STATEMENT
// ============================================================================

// RaiseStatement represents a RAISE statement
// Reference: src/ast/mod.rs:2840
type RaiseStatement struct {
	BaseStatement
	// UsingMessage indicates whether the USING MESSAGE = syntax was used
	UsingMessage bool
	// Message is the expression provided to RAISE
	Message expr.Expr
}

func (r *RaiseStatement) statementNode() {}

func (r *RaiseStatement) String() string {
	var f strings.Builder
	f.WriteString("RAISE")
	if r.Message != nil {
		if r.UsingMessage {
			f.WriteString(" USING MESSAGE = ")
			f.WriteString(r.Message.String())
		} else {
			f.WriteString(" ")
			f.WriteString(r.Message.String())
		}
	}
	return f.String()
}

// ============================================================================
// CALL
// ============================================================================

// Call represents a CALL statement
type Call struct {
	BaseStatement
	Function *expr.FunctionExpr
}

func (c *Call) statementNode() {}

func (c *Call) String() string {
	return "CALL " + c.Function.String()
}

// ============================================================================
// COPY
// ============================================================================

// Copy represents a COPY statement
type Copy struct {
	BaseStatement
	Source        *expr.CopySource
	To            bool
	Target        *expr.CopyTarget
	Options       []*expr.CopyOption
	LegacyOptions []*expr.CopyLegacyOption
	Values        []*string
}

func (c *Copy) statementNode() {}

func (c *Copy) String() string {
	var f strings.Builder
	f.WriteString("COPY")
	if c.Source != nil {
		f.WriteString(c.Source.String())
	}
	if c.To {
		f.WriteString(" TO ")
	} else {
		f.WriteString(" FROM ")
	}
	if c.Target != nil {
		f.WriteString(c.Target.String())
	}
	if len(c.Options) > 0 {
		f.WriteString(" (")
		optStrs := make([]string, len(c.Options))
		for i, opt := range c.Options {
			optStrs[i] = opt.String()
		}
		f.WriteString(strings.Join(optStrs, ", "))
		f.WriteString(")")
	}
	if len(c.LegacyOptions) > 0 {
		for _, opt := range c.LegacyOptions {
			f.WriteString(" ")
			f.WriteString(opt.String())
		}
	}
	return f.String()
}

// ============================================================================
// COPY INTO SNOWFLAKE
// ============================================================================

// CopyIntoSnowflake represents a COPY INTO statement (Snowflake)
type CopyIntoSnowflake struct {
	BaseStatement
	Kind                *expr.CopyIntoSnowflakeKind
	Into                *ast.ObjectName
	IntoColumns         []*ast.Ident
	FromObj             *ast.ObjectName
	FromObjAlias        *ast.Ident
	StageParams         *expr.StageParamsObject
	FromTransformations []*expr.StageLoadSelectItemWrapper
	FromQuery           *query.Query
	Files               []string
	Pattern             *string
	FileFormat          *expr.KeyValueOptions
	CopyOptions         *expr.KeyValueOptions
	ValidationMode      *string
	Partition           ast.Expr
}

func (c *CopyIntoSnowflake) statementNode() {}

func (c *CopyIntoSnowflake) String() string {
	var f strings.Builder
	f.WriteString("COPY INTO ")
	f.WriteString(c.Into.String())

	// Add column list if present
	if len(c.IntoColumns) > 0 {
		cols := make([]string, len(c.IntoColumns))
		for i, col := range c.IntoColumns {
			cols[i] = col.String()
		}
		f.WriteString(" (")
		f.WriteString(strings.Join(cols, ", "))
		f.WriteString(")")
	}

	// Add FROM clause
	if c.FromObj != nil {
		f.WriteString(" FROM ")
		f.WriteString(c.FromObj.String())
		if c.FromObjAlias != nil {
			f.WriteString(" AS ")
			f.WriteString(c.FromObjAlias.String())
		}
	}

	// Add stage params
	if c.StageParams != nil {
		stageParamsStr := c.StageParams.String()
		if stageParamsStr != "" {
			f.WriteString(" ")
			f.WriteString(stageParamsStr)
		}
	}

	// Add FILE_FORMAT
	if c.FileFormat != nil && len(c.FileFormat.Options) > 0 {
		f.WriteString(" FILE_FORMAT = (")
		f.WriteString(c.FileFormat.String())
		f.WriteString(")")
	}

	// Add FILES
	if len(c.Files) > 0 {
		f.WriteString(" FILES = (")
		quotedFiles := make([]string, len(c.Files))
		for i, file := range c.Files {
			quotedFiles[i] = fmt.Sprintf("'%s'", file)
		}
		f.WriteString(strings.Join(quotedFiles, ", "))
		f.WriteString(")")
	}

	// Add PATTERN
	if c.Pattern != nil {
		f.WriteString(" PATTERN = '")
		f.WriteString(*c.Pattern)
		f.WriteString("'")
	}

	// Add VALIDATION_MODE
	if c.ValidationMode != nil {
		f.WriteString(" VALIDATION_MODE = ")
		f.WriteString(*c.ValidationMode)
	}

	// Add COPY_OPTIONS
	if c.CopyOptions != nil && len(c.CopyOptions.Options) > 0 {
		f.WriteString(" COPY_OPTIONS = (")
		f.WriteString(c.CopyOptions.String())
		f.WriteString(")")
	}

	// Add PARTITION BY
	if c.Partition != nil {
		f.WriteString(" PARTITION BY ")
		f.WriteString(c.Partition.String())
	}

	return f.String()
}

// ============================================================================
// OPEN
// ============================================================================

// Open represents an OPEN cursor statement
type Open struct {
	BaseStatement
	Cursor *ast.Ident
}

func (o *Open) statementNode() {}

func (o *Open) String() string {
	var f strings.Builder
	f.WriteString("OPEN ")
	f.WriteString(o.Cursor.String())
	return f.String()
}

// ============================================================================
// CLOSE
// ============================================================================

// Close represents a CLOSE cursor statement
type Close struct {
	BaseStatement
	Cursor *expr.CloseCursor
}

func (c *Close) statementNode() {}

func (c *Close) String() string {
	var f strings.Builder
	f.WriteString("CLOSE ")
	f.WriteString(c.Cursor.String())
	return f.String()
}

// ============================================================================
// DECLARE
// ============================================================================

// Declare represents a DECLARE statement
type Declare struct {
	BaseStatement
	Stmts []*expr.Declare
}

func (d *Declare) statementNode() {}

func (d *Declare) String() string {
	var f strings.Builder
	f.WriteString("DECLARE ")
	for i, stmt := range d.Stmts {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(stmt.String())
	}
	return f.String()
}

// ============================================================================
// FETCH
// ============================================================================

// Fetch represents a FETCH statement
type Fetch struct {
	BaseStatement
	Name      *ast.Ident
	Direction *expr.FetchDirection
	Position  *expr.FetchPosition
	Into      *ast.ObjectName
}

func (f *Fetch) statementNode() {}

func (f *Fetch) String() string {
	var sb strings.Builder
	sb.WriteString("FETCH ")
	if f.Direction != nil {
		sb.WriteString(f.Direction.String())
		sb.WriteString(" ")
	}
	if f.Position != nil {
		sb.WriteString(f.Position.String())
		sb.WriteString(" ")
	}
	sb.WriteString(f.Name.String())
	return sb.String()
}

// ============================================================================
// FLUSH
// ============================================================================

// Flush represents a FLUSH statement
type Flush struct {
	BaseStatement
	ObjectType expr.FlushType
	Location   *expr.FlushLocation
	Channel    *string
	ReadLock   bool
	Export     bool
	Tables     []*ast.ObjectName
}

func (f *Flush) statementNode() {}

func (f *Flush) String() string {
	var sb strings.Builder
	sb.WriteString("FLUSH ")
	sb.WriteString(f.ObjectType.String())
	return sb.String()
}

// ============================================================================
// DISCARD
// ============================================================================

// Discard represents a DISCARD statement
type Discard struct {
	BaseStatement
	ObjectType expr.DiscardObject
}

func (d *Discard) statementNode() {}

func (d *Discard) String() string {
	var sb strings.Builder
	sb.WriteString("DISCARD ")
	sb.WriteString(d.ObjectType.String())
	return sb.String()
}

// ============================================================================
// SHOW FUNCTIONS
// ============================================================================

// ShowFunctions represents a SHOW FUNCTIONS statement
type ShowFunctions struct {
	BaseStatement
	Filter *expr.ShowStatementFilter
}

func (s *ShowFunctions) statementNode() {}

func (s *ShowFunctions) String() string {
	var f strings.Builder
	f.WriteString("SHOW FUNCTIONS")
	if s.Filter != nil {
		f.WriteString(" ")
		f.WriteString(s.Filter.String())
	}
	return f.String()
}

// ============================================================================
// SHOW VARIABLE
// ============================================================================

// ShowVariable represents a SHOW VARIABLE statement
type ShowVariable struct {
	BaseStatement
	Variable []*ast.Ident
}

func (s *ShowVariable) statementNode() {}

func (s *ShowVariable) String() string {
	var f strings.Builder
	f.WriteString("SHOW ")
	for i, v := range s.Variable {
		if i > 0 {
			f.WriteString(".")
		}
		f.WriteString(v.String())
	}
	return f.String()
}

// ============================================================================
// SHOW STATUS
// ============================================================================

// ShowStatus represents a SHOW STATUS statement
type ShowStatus struct {
	BaseStatement
	Filter  *expr.ShowStatementFilter
	Global  bool
	Session bool
}

func (s *ShowStatus) statementNode() {}

func (s *ShowStatus) String() string {
	var f strings.Builder
	f.WriteString("SHOW ")
	if s.Global {
		f.WriteString("GLOBAL ")
	}
	if s.Session {
		f.WriteString("SESSION ")
	}
	f.WriteString("STATUS")
	if s.Filter != nil {
		f.WriteString(" ")
		f.WriteString(s.Filter.String())
	}
	return f.String()
}

// ============================================================================
// SHOW VARIABLES
// ============================================================================

// ShowVariables represents a SHOW VARIABLES statement
type ShowVariables struct {
	BaseStatement
	Filter  *expr.ShowStatementFilter
	Global  bool
	Session bool
}

func (s *ShowVariables) statementNode() {}

func (s *ShowVariables) String() string {
	var f strings.Builder
	f.WriteString("SHOW ")
	if s.Global {
		f.WriteString("GLOBAL ")
	}
	if s.Session {
		f.WriteString("SESSION ")
	}
	f.WriteString("VARIABLES")
	if s.Filter != nil {
		f.WriteString(" ")
		f.WriteString(s.Filter.String())
	}
	return f.String()
}

// ============================================================================
// SHOW CREATE
// ============================================================================

// ShowCreate represents a SHOW CREATE statement
type ShowCreate struct {
	BaseStatement
	ObjType expr.ShowCreateObject
	ObjName *ast.ObjectName
}

func (s *ShowCreate) statementNode() {}

func (s *ShowCreate) String() string {
	var f strings.Builder
	f.WriteString("SHOW CREATE ")
	f.WriteString(s.ObjType.String())
	f.WriteString(" ")
	f.WriteString(s.ObjName.String())
	return f.String()
}

// ============================================================================
// SHOW COLUMNS
// ============================================================================

// ShowColumns represents a SHOW COLUMNS statement
type ShowColumns struct {
	BaseStatement
	Extended      bool
	Full          bool
	ExtendedFirst bool // true if EXTENDED appeared before FULL (to preserve order)
	ShowOptions   *expr.ShowStatementOptions
}

func (s *ShowColumns) statementNode() {}

func (s *ShowColumns) String() string {
	var f strings.Builder
	f.WriteString("SHOW ")
	// Preserve order of EXTENDED and FULL as they appeared
	if s.Full && s.Extended {
		if s.ExtendedFirst {
			f.WriteString("EXTENDED FULL ")
		} else {
			f.WriteString("FULL EXTENDED ")
		}
	} else {
		if s.Full {
			f.WriteString("FULL ")
		}
		if s.Extended {
			f.WriteString("EXTENDED ")
		}
	}
	f.WriteString("COLUMNS")
	if s.ShowOptions != nil {
		optStr := s.ShowOptions.String()
		if optStr != "" {
			f.WriteString(" ")
			f.WriteString(optStr)
		}
	}
	return f.String()
}

// ============================================================================
// SHOW DATABASES
// ============================================================================

// ShowDatabases represents a SHOW DATABASES statement
type ShowDatabases struct {
	BaseStatement
	Terse       bool
	History     bool
	ShowOptions *expr.ShowStatementOptions
}

func (s *ShowDatabases) statementNode() {}

func (s *ShowDatabases) String() string {
	var f strings.Builder
	f.WriteString("SHOW ")
	if s.Terse {
		f.WriteString("TERSE ")
	}
	f.WriteString("DATABASES")
	if s.History {
		f.WriteString(" HISTORY")
	}
	if s.ShowOptions != nil {
		optStr := s.ShowOptions.String()
		if optStr != "" {
			f.WriteString(" ")
			f.WriteString(optStr)
		}
	}
	return f.String()
}

// ============================================================================
// SHOW SCHEMAS
// ============================================================================

// ShowSchemas represents a SHOW SCHEMAS statement
type ShowSchemas struct {
	BaseStatement
	Terse       bool
	History     bool
	ShowOptions *expr.ShowStatementOptions
}

func (s *ShowSchemas) statementNode() {}

func (s *ShowSchemas) String() string {
	var f strings.Builder
	f.WriteString("SHOW ")
	if s.Terse {
		f.WriteString("TERSE ")
	}
	f.WriteString("SCHEMAS")
	if s.History {
		f.WriteString(" HISTORY")
	}
	if s.ShowOptions != nil {
		optStr := s.ShowOptions.String()
		if optStr != "" {
			f.WriteString(" ")
			f.WriteString(optStr)
		}
	}
	return f.String()
}

// ============================================================================
// SHOW CHARSET
// ============================================================================

// ShowCharset represents a SHOW CHARSET statement
type ShowCharset struct {
	BaseStatement
	Filter          *expr.ShowStatementFilter
	UseCharacterSet bool // true if "CHARACTER SET" was used instead of "CHARSET" // true if "CHARACTER SET" was used instead of "CHARSET"
}

func (s *ShowCharset) statementNode() {}

func (s *ShowCharset) String() string {
	var f strings.Builder
	if s.UseCharacterSet {
		f.WriteString("SHOW CHARACTER SET")
	} else {
		f.WriteString("SHOW CHARSET")
	}
	if s.Filter != nil {
		f.WriteString(" ")
		f.WriteString(s.Filter.String())
	}
	return f.String()
}

// ============================================================================
// SHOW OBJECTS
// ============================================================================

// ShowObjects represents a SHOW OBJECTS statement (Snowflake)
type ShowObjects struct {
	BaseStatement
	Terse       bool
	History     bool
	ShowOptions *expr.ShowStatementOptions
}

func (s *ShowObjects) statementNode() {}

func (s *ShowObjects) String() string {
	var f strings.Builder
	f.WriteString("SHOW ")
	if s.Terse {
		f.WriteString("TERSE ")
	}
	f.WriteString("OBJECTS")
	if s.History {
		f.WriteString(" HISTORY")
	}
	if s.ShowOptions != nil {
		optStr := s.ShowOptions.String()
		if optStr != "" {
			f.WriteString(" ")
			f.WriteString(optStr)
		}
	}
	return f.String()
}

// ============================================================================
// SHOW TABLES
// ============================================================================

// ShowTables represents a SHOW TABLES statement
type ShowTables struct {
	BaseStatement
	Terse         bool
	History       bool
	Extended      bool
	Full          bool
	ExtendedFirst bool // true if EXTENDED appeared before FULL (to preserve order)
	External      bool
	ShowOptions   *expr.ShowStatementOptions
}

func (s *ShowTables) statementNode() {}

func (s *ShowTables) String() string {
	var f strings.Builder
	f.WriteString("SHOW ")
	if s.Terse {
		f.WriteString("TERSE ")
	}
	// Preserve order of EXTENDED and FULL as they appeared
	if s.Full && s.Extended {
		if s.ExtendedFirst {
			f.WriteString("EXTENDED FULL ")
		} else {
			f.WriteString("FULL EXTENDED ")
		}
	} else {
		if s.Full {
			f.WriteString("FULL ")
		}
		if s.Extended {
			f.WriteString("EXTENDED ")
		}
	}
	if s.External {
		f.WriteString("EXTERNAL ")
	}
	f.WriteString("TABLES")
	if s.History {
		f.WriteString(" HISTORY")
	}
	if s.ShowOptions != nil {
		optStr := s.ShowOptions.String()
		if optStr != "" {
			f.WriteString(" ")
			f.WriteString(optStr)
		}
	}
	return f.String()
}

// ============================================================================
// SHOW VIEWS
// ============================================================================

// ShowViews represents a SHOW VIEWS statement
type ShowViews struct {
	BaseStatement
	Terse        bool
	Materialized bool
	ShowOptions  *expr.ShowStatementOptions
}

func (s *ShowViews) statementNode() {}

func (s *ShowViews) String() string {
	var f strings.Builder
	f.WriteString("SHOW ")
	if s.Terse {
		f.WriteString("TERSE ")
	}
	if s.Materialized {
		f.WriteString("MATERIALIZED ")
	}
	f.WriteString("VIEWS")
	if s.ShowOptions != nil {
		optStr := s.ShowOptions.String()
		if optStr != "" {
			f.WriteString(" ")
			f.WriteString(optStr)
		}
	}
	return f.String()
}

// ============================================================================
// SHOW COLLATION
// ============================================================================

// ShowCollation represents a SHOW COLLATION statement
type ShowCollation struct {
	BaseStatement
	Filter *expr.ShowStatementFilter
}

func (s *ShowCollation) statementNode() {}

func (s *ShowCollation) String() string {
	var f strings.Builder
	f.WriteString("SHOW COLLATION")
	if s.Filter != nil {
		f.WriteString(" ")
		f.WriteString(s.Filter.String())
	}
	return f.String()
}

// ============================================================================
// START TRANSACTION
// ============================================================================

// StartTransaction represents a START TRANSACTION / BEGIN statement
type StartTransaction struct {
	BaseStatement
	Modes         []*expr.TransactionMode
	Begin         bool
	Transaction   *expr.BeginTransactionKind
	Modifier      *expr.TransactionModifier
	Statements    []ast.Statement
	Exception     []*expr.ExceptionWhen
	HasEndKeyword bool
}

func (s *StartTransaction) statementNode() {}

func (s *StartTransaction) String() string {
	var f strings.Builder
	if s.Begin {
		f.WriteString("BEGIN")
		if s.Modifier != nil && *s.Modifier != expr.TransactionModifierNone {
			f.WriteString(" ")
			f.WriteString(s.Modifier.String())
		}
		if s.Transaction != nil && *s.Transaction != expr.BeginTransactionKindNone {
			f.WriteString(" ")
			f.WriteString(s.Transaction.String())
		}
	} else {
		f.WriteString("START TRANSACTION")
	}
	for _, mode := range s.Modes {
		if *mode != expr.TransactionModeNone {
			f.WriteString(" ")
			f.WriteString(mode.String())
		}
	}
	return f.String()
}

// ============================================================================
// COMMENT
// ============================================================================

// Comment represents a COMMENT statement
type Comment struct {
	BaseStatement
	ObjectType expr.CommentObject
	ObjectName *ast.ObjectName
	Comment    *string
	IfExists   bool
}

func (c *Comment) statementNode() {}

func (c *Comment) String() string {
	var f strings.Builder
	f.WriteString("COMMENT ON ")
	f.WriteString(c.ObjectType.String())
	f.WriteString(" ")
	f.WriteString(c.ObjectName.String())
	f.WriteString(" IS ")
	if c.Comment != nil {
		f.WriteString("'")
		f.WriteString(*c.Comment)
		f.WriteString("'")
	} else {
		f.WriteString("NULL")
	}
	return f.String()
}

// ============================================================================
// COMMIT
// ============================================================================

// Commit represents a COMMIT statement
type Commit struct {
	BaseStatement
	Chain    bool
	End      bool
	Modifier *expr.TransactionModifier
}

func (c *Commit) statementNode() {}

func (c *Commit) String() string {
	var f strings.Builder
	if c.End {
		f.WriteString("END")
	} else {
		f.WriteString("COMMIT")
	}
	if c.Chain {
		f.WriteString(" AND CHAIN")
	}
	return f.String()
}

// ============================================================================
// ROLLBACK
// ============================================================================

// Rollback represents a ROLLBACK statement
type Rollback struct {
	BaseStatement
	Chain     bool
	Savepoint *ast.Ident
}

func (r *Rollback) statementNode() {}

func (r *Rollback) String() string {
	var f strings.Builder
	f.WriteString("ROLLBACK")
	if r.Chain {
		f.WriteString(" AND CHAIN")
	}
	if r.Savepoint != nil {
		f.WriteString(" TO SAVEPOINT ")
		f.WriteString(r.Savepoint.String())
	}
	return f.String()
}

// ============================================================================
// SAVEPOINT
// ============================================================================

// Savepoint represents a SAVEPOINT statement
type Savepoint struct {
	BaseStatement
	Name *ast.Ident
}

func (s *Savepoint) statementNode() {}

func (s *Savepoint) String() string {
	var f strings.Builder
	f.WriteString("SAVEPOINT ")
	f.WriteString(s.Name.String())
	return f.String()
}

// ============================================================================
// RELEASE SAVEPOINT
// ============================================================================

// ReleaseSavepoint represents a RELEASE SAVEPOINT statement
type ReleaseSavepoint struct {
	BaseStatement
	Name *ast.Ident
}

func (r *ReleaseSavepoint) statementNode() {}

func (r *ReleaseSavepoint) String() string {
	var f strings.Builder
	f.WriteString("RELEASE SAVEPOINT ")
	f.WriteString(r.Name.String())
	return f.String()
}

// ============================================================================
// ASSERT
// ============================================================================

// Assert represents an ASSERT statement
type Assert struct {
	BaseStatement
	Condition expr.Expr
	Message   expr.Expr
}

func (a *Assert) statementNode() {}

func (a *Assert) String() string {
	var f strings.Builder
	f.WriteString("ASSERT ")
	f.WriteString(a.Condition.String())
	if a.Message != nil {
		f.WriteString(" AS ")
		f.WriteString(a.Message.String())
	}
	return f.String()
}

// ============================================================================
// DEALLOCATE
// ============================================================================

// Deallocate represents a DEALLOCATE statement
type Deallocate struct {
	BaseStatement
	Name    *ast.Ident
	Prepare bool
}

func (d *Deallocate) statementNode() {}

func (d *Deallocate) String() string {
	var f strings.Builder
	f.WriteString("DEALLOCATE ")
	if d.Prepare {
		f.WriteString("PREPARE ")
	}
	f.WriteString(d.Name.String())
	return f.String()
}

// ============================================================================
// EXECUTE
// ============================================================================

// Execute represents an EXECUTE statement
type Execute struct {
	BaseStatement
	Name           *ast.ObjectName
	Parameters     []expr.Expr
	HasParentheses bool
	Immediate      bool
	Into           []*ast.Ident
	Using          []*expr.ExprWithAlias
	Output         bool
	Default        bool
}

func (e *Execute) statementNode() {}

func (e *Execute) String() string {
	var f strings.Builder
	if e.Immediate {
		f.WriteString("EXECUTE IMMEDIATE ")
	} else {
		f.WriteString("EXECUTE ")
	}
	if e.Name != nil {
		f.WriteString(e.Name.String())
	}
	if e.HasParentheses && len(e.Parameters) > 0 {
		f.WriteString("(")
		for i, param := range e.Parameters {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(param.String())
		}
		f.WriteString(")")
	}
	if len(e.Using) > 0 {
		f.WriteString(" USING ")
		for i, u := range e.Using {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(u.String())
		}
	}
	return f.String()
}

// ============================================================================
// PREPARE
// ============================================================================

// Prepare represents a PREPARE statement
type Prepare struct {
	BaseStatement
	Name      *ast.Ident
	DataTypes []datatype.DataType
	Statement ast.Statement
}

func (p *Prepare) statementNode() {}

func (p *Prepare) String() string {
	var f strings.Builder
	f.WriteString("PREPARE ")
	f.WriteString(p.Name.String())
	if len(p.DataTypes) > 0 {
		f.WriteString(" (")
		for i, dt := range p.DataTypes {
			if i > 0 {
				f.WriteString(", ")
			}
			if dt != nil {
				f.WriteString(dt.String())
			}
		}
		f.WriteString(")")
	}
	f.WriteString(" AS ")
	f.WriteString(p.Statement.String())
	return f.String()
}

// ============================================================================
// KILL
// ============================================================================

// Kill represents a KILL statement
type Kill struct {
	BaseStatement
	Modifier *expr.KillType
	Hard     bool // For KILL HARD [QUERY|CONNECTION] syntax
	ID       uint64
}

func (k *Kill) statementNode() {}

func (k *Kill) String() string {
	var f strings.Builder
	f.WriteString("KILL ")
	if k.Hard {
		f.WriteString("HARD ")
	}
	if k.Modifier != nil {
		f.WriteString(k.Modifier.String())
		f.WriteString(" ")
	}
	f.WriteString(fmt.Sprintf("%d", k.ID))
	return f.String()
}

// ============================================================================
// EXPLAIN TABLE
// ============================================================================

// ExplainTable represents an EXPLAIN TABLE statement
type ExplainTable struct {
	BaseStatement
	DescribeAlias   expr.DescribeAlias
	HiveFormat      *expr.HiveDescribeFormat
	HasTableKeyword bool
	TableName       *ast.ObjectName
}

func (e *ExplainTable) statementNode() {}

func (e *ExplainTable) String() string {
	var f strings.Builder
	f.WriteString(e.DescribeAlias.String())
	if e.HasTableKeyword {
		f.WriteString(" TABLE")
	}
	f.WriteString(" ")
	f.WriteString(e.TableName.String())
	return f.String()
}

// ============================================================================
// EXPLAIN
// ============================================================================

// Explain represents an EXPLAIN statement
type Explain struct {
	BaseStatement
	DescribeAlias expr.DescribeAlias
	Analyze       bool
	Verbose       bool
	QueryPlan     bool
	Estimate      bool
	Statement     ast.Statement
	Format        *expr.AnalyzeFormatKind
	Options       []*expr.UtilityOption
}

func (e *Explain) statementNode() {}

func (e *Explain) String() string {
	var f strings.Builder
	f.WriteString(e.DescribeAlias.String())
	if e.Analyze {
		f.WriteString(" ANALYZE")
	}
	if e.Verbose {
		f.WriteString(" VERBOSE")
	}
	if e.QueryPlan {
		f.WriteString(" QUERY PLAN")
	}
	f.WriteString(" ")
	f.WriteString(e.Statement.String())
	return f.String()
}

// ============================================================================
// CACHE
// ============================================================================

// Cache represents a CACHE statement (Spark)
type Cache struct {
	BaseStatement
	TableFlag *ast.ObjectName
	TableName *ast.ObjectName
	HasAs     bool
	Options   []*expr.SqlOption
	Query     *query.Query
}

func (c *Cache) statementNode() {}

func (c *Cache) String() string {
	var f strings.Builder
	f.WriteString("CACHE TABLE ")
	f.WriteString(c.TableName.String())
	if c.Query != nil {
		f.WriteString(" AS ")
		f.WriteString(c.Query.String())
	}
	return f.String()
}

// ============================================================================
// UNCACHE
// ============================================================================

// Uncache represents an UNCACHE statement (Spark)
type Uncache struct {
	BaseStatement
	TableName *ast.ObjectName
	IfExists  bool
}

func (u *Uncache) statementNode() {}

func (u *Uncache) String() string {
	var f strings.Builder
	f.WriteString("UNCACHE TABLE ")
	if u.IfExists {
		f.WriteString("IF EXISTS ")
	}
	f.WriteString(u.TableName.String())
	return f.String()
}

// ============================================================================
// PRAGMA
// ============================================================================

// Pragma represents a PRAGMA statement (SQLite)
type Pragma struct {
	BaseStatement
	Name  *ast.ObjectName
	Value expr.Expr
	IsEq  bool
}

func (p *Pragma) statementNode() {}

func (p *Pragma) String() string {
	var f strings.Builder
	f.WriteString("PRAGMA ")
	f.WriteString(p.Name.String())
	if p.Value != nil {
		if p.IsEq {
			f.WriteString(" = ")
		} else {
			f.WriteString("(")
			f.WriteString(p.Value.String())
			f.WriteString(")")
		}
	}
	return f.String()
}

// ============================================================================
// LOCK
// ============================================================================

// Lock represents a LOCK statement
type Lock struct {
	BaseStatement
	Tables []*expr.LockTable
	Mode   *expr.LockMode
	NoWait bool
}

func (l *Lock) statementNode() {}

func (l *Lock) String() string {
	var f strings.Builder
	f.WriteString("LOCK TABLE ")
	for i, table := range l.Tables {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(table.String())
	}
	return f.String()
}

// ============================================================================
// LOCK TABLES
// ============================================================================

// LockTables represents a LOCK TABLES statement (MySQL)
type LockTables struct {
	BaseStatement
	Tables []*expr.LockTable
}

func (l *LockTables) statementNode() {}

func (l *LockTables) String() string {
	var f strings.Builder
	f.WriteString("LOCK TABLES ")
	for i, table := range l.Tables {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(table.String())
	}
	return f.String()
}

// ============================================================================
// UNLOCK TABLES
// ============================================================================

// UnlockTables represents an UNLOCK TABLES statement
type UnlockTables struct {
	BaseStatement
}

func (u *UnlockTables) statementNode() {}

func (u *UnlockTables) String() string {
	return "UNLOCK TABLES"
}

// ============================================================================
// UNLOAD
// ============================================================================

// Unload represents an UNLOAD statement (Athena/Redshift)
type Unload struct {
	BaseStatement
	Query     *query.Query
	QueryText *string
	To        *ast.Ident
	Auth      *expr.IamRoleKind
	With      []*expr.SqlOption
	Options   []*expr.CopyLegacyOption
}

func (u *Unload) statementNode() {}

func (u *Unload) String() string {
	var f strings.Builder
	f.WriteString("UNLOAD")
	if u.Query != nil {
		f.WriteString("(")
		f.WriteString(u.Query.String())
		f.WriteString(")")
	}
	f.WriteString(" TO ")
	f.WriteString(u.To.String())
	return f.String()
}

// ============================================================================
// OPTIMIZE TABLE
// ============================================================================

// OptimizeTable represents an OPTIMIZE TABLE statement
type OptimizeTable struct {
	BaseStatement
	Name            *ast.ObjectName
	HasTableKeyword bool
	OnCluster       *ast.Ident
	Partition       *expr.Partition
	IncludeFinal    bool
	Deduplicate     *expr.Deduplicate
	Predicate       expr.Expr
	Zorder          []expr.Expr
}

func (o *OptimizeTable) statementNode() {}

func (o *OptimizeTable) String() string {
	var f strings.Builder
	f.WriteString("OPTIMIZE ")
	if o.HasTableKeyword {
		f.WriteString("TABLE ")
	}
	f.WriteString(o.Name.String())
	return f.String()
}

// ============================================================================
// LISTEN
// ============================================================================

// Listen represents a LISTEN statement (PostgreSQL)
type Listen struct {
	BaseStatement
	Channel *ast.Ident
}

func (l *Listen) statementNode() {}

func (l *Listen) String() string {
	var f strings.Builder
	f.WriteString("LISTEN ")
	f.WriteString(l.Channel.String())
	return f.String()
}

// ============================================================================
// UNLISTEN
// ============================================================================

// Unlisten represents an UNLISTEN statement (PostgreSQL)
type Unlisten struct {
	BaseStatement
	Channel *ast.Ident
}

func (u *Unlisten) statementNode() {}

func (u *Unlisten) String() string {
	var f strings.Builder
	f.WriteString("UNLISTEN ")
	f.WriteString(u.Channel.String())
	return f.String()
}

// ============================================================================
// NOTIFY
// ============================================================================

// Notify represents a NOTIFY statement (PostgreSQL)
type Notify struct {
	BaseStatement
	Channel *ast.Ident
	Payload *string
}

func (n *Notify) statementNode() {}

func (n *Notify) String() string {
	var f strings.Builder
	f.WriteString("NOTIFY ")
	f.WriteString(n.Channel.String())
	if n.Payload != nil {
		f.WriteString(", '")
		f.WriteString(*n.Payload)
		f.WriteString("'")
	}
	return f.String()
}

// ============================================================================
// LOAD DATA
// ============================================================================

// LoadData represents a LOAD DATA statement (Hive)
type LoadData struct {
	BaseStatement
	Local       bool
	Inpath      string
	Overwrite   bool
	TableName   *ast.ObjectName
	Partitioned []expr.Expr
	TableFormat *expr.HiveLoadDataFormat
}

func (l *LoadData) statementNode() {}

func (l *LoadData) String() string {
	var f strings.Builder
	f.WriteString("LOAD DATA ")
	if l.Local {
		f.WriteString("LOCAL ")
	}
	f.WriteString("INPATH '")
	f.WriteString(l.Inpath)
	f.WriteString("' ")
	if l.Overwrite {
		f.WriteString("OVERWRITE ")
	}
	f.WriteString("INTO TABLE ")
	f.WriteString(l.TableName.String())
	return f.String()
}

// ============================================================================
// RENAME TABLE
// ============================================================================

// RenameTable represents a RENAME TABLE statement
type RenameTable struct {
	BaseStatement
	Renames []*expr.RenameTable
}

func (r *RenameTable) statementNode() {}

func (r *RenameTable) String() string {
	var f strings.Builder
	f.WriteString("RENAME TABLE ")
	for i, rename := range r.Renames {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(rename.String())
	}
	return f.String()
}

// ============================================================================
// LIST
// ============================================================================

// List represents a LIST statement (Snowflake)
type List struct {
	BaseStatement
	Command *expr.FileStagingCommand
}

func (l *List) statementNode() {}

func (l *List) String() string {
	var f strings.Builder
	f.WriteString("LIST ")
	f.WriteString(l.Command.String())
	return f.String()
}

// ============================================================================
// REMOVE
// ============================================================================

// Remove represents a REMOVE statement (Snowflake)
type Remove struct {
	BaseStatement
	Command *expr.FileStagingCommand
}

func (r *Remove) statementNode() {}

func (r *Remove) String() string {
	var f strings.Builder
	f.WriteString("REMOVE ")
	f.WriteString(r.Command.String())
	return f.String()
}

// ============================================================================
// RAISERROR
// ============================================================================

// RaisError represents a RAISERROR statement (MSSQL)
type RaisError struct {
	BaseStatement
	Message   expr.Expr
	Severity  expr.Expr
	State     expr.Expr
	Arguments []expr.Expr
	Options   []*expr.RaisErrorOption
}

func (r *RaisError) statementNode() {}

func (r *RaisError) String() string {
	var f strings.Builder
	f.WriteString("RAISERROR (")
	f.WriteString(r.Message.String())
	f.WriteString(", ")
	f.WriteString(r.Severity.String())
	f.WriteString(", ")
	f.WriteString(r.State.String())
	f.WriteString(")")
	return f.String()
}

// ============================================================================
// THROW
// ============================================================================

// Throw represents a THROW statement (MSSQL)
type Throw struct {
	BaseStatement
	Statement *expr.ThrowStatement
}

func (t *Throw) statementNode() {}

func (t *Throw) String() string {
	var f strings.Builder
	f.WriteString("THROW ")
	f.WriteString(t.Statement.String())
	return f.String()
}

// ============================================================================
// PRINT
// ============================================================================

// Print represents a PRINT statement (MSSQL)
type Print struct {
	BaseStatement
	Statement *expr.PrintStatement
}

func (p *Print) statementNode() {}

func (p *Print) String() string {
	var f strings.Builder
	f.WriteString("PRINT ")
	f.WriteString(p.Statement.String())
	return f.String()
}

// ============================================================================
// WAITFOR
// ============================================================================

// WaitFor represents a WAITFOR statement (MSSQL)
type WaitFor struct {
	BaseStatement
	Statement *expr.WaitForStatement
}

func (w *WaitFor) statementNode() {}

func (w *WaitFor) String() string {
	var f strings.Builder
	f.WriteString("WAITFOR ")
	f.WriteString(w.Statement.String())
	return f.String()
}

// ============================================================================
// RETURN
// ============================================================================

// Return represents a RETURN statement
type Return struct {
	BaseStatement
	Statement *expr.ReturnStatement
}

func (r *Return) statementNode() {}

func (r *Return) String() string {
	var f strings.Builder
	f.WriteString("RETURN")
	if r.Statement != nil {
		f.WriteString(" ")
		f.WriteString(r.Statement.String())
	}
	return f.String()
}

// ============================================================================
// EXPORT DATA
// ============================================================================

// ExportData represents an EXPORT DATA statement (BigQuery)
type ExportData struct {
	BaseStatement
	Options *expr.KeyValueOptions
	Query   *query.Query
}

func (e *ExportData) statementNode() {}

func (e *ExportData) String() string {
	var f strings.Builder
	f.WriteString("EXPORT DATA ")
	if e.Options != nil {
		f.WriteString("OPTIONS(")
		f.WriteString(e.Options.String())
		f.WriteString(") ")
	}
	f.WriteString("AS ")
	f.WriteString(e.Query.String())
	return f.String()
}

// ============================================================================
// VACUUM
// ============================================================================

// Vacuum represents a VACUUM statement (Redshift)
type Vacuum struct {
	BaseStatement
	Statement *expr.VacuumStatement
}

func (v *Vacuum) statementNode() {}

func (v *Vacuum) String() string {
	var f strings.Builder
	f.WriteString("VACUUM ")
	f.WriteString(v.Statement.String())
	return f.String()
}

// ============================================================================
// RESET
// ============================================================================

// Reset represents a RESET statement
type Reset struct {
	BaseStatement
	Statement *expr.ResetStatement
}

func (r *Reset) statementNode() {}

func (r *Reset) String() string {
	var f strings.Builder
	f.WriteString("RESET ")
	f.WriteString(r.Statement.String())
	return f.String()
}

// ============================================================================
// SET TRANSACTION
// ============================================================================

// SetTransaction represents a SET TRANSACTION statement.
// Syntax: SET [ SESSION | LOCAL ] TRANSACTION [ modes ] [ SNAPSHOT value ]
type SetTransaction struct {
	BaseStatement
	Modes    []expr.TransactionMode
	Snapshot expr.Expr
	Session  bool
	Local    bool
}

func (s *SetTransaction) statementNode() {}

func (s *SetTransaction) String() string {
	var f strings.Builder
	f.WriteString("SET ")
	if s.Session {
		f.WriteString("SESSION ")
	}
	if s.Local {
		f.WriteString("LOCAL ")
	}
	f.WriteString("TRANSACTION")
	if len(s.Modes) > 0 {
		f.WriteString(" ")
		for i, mode := range s.Modes {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(mode.String())
		}
	}
	if s.Snapshot != nil {
		f.WriteString(" SNAPSHOT ")
		f.WriteString(s.Snapshot.String())
	}
	return f.String()
}

// ============================================================================
// SET SESSION AUTHORIZATION
// ============================================================================

// SetSessionAuthorization represents a SET SESSION AUTHORIZATION statement.
// Syntax: SET { SESSION | LOCAL } AUTHORIZATION { user_name | DEFAULT }
type SetSessionAuthorization struct {
	BaseStatement
	Local   bool
	Session bool
	User    *ast.Ident
	Default bool
}

func (s *SetSessionAuthorization) statementNode() {}

func (s *SetSessionAuthorization) String() string {
	var f strings.Builder
	f.WriteString("SET ")
	if s.Session {
		f.WriteString("SESSION ")
	}
	if s.Local {
		f.WriteString("LOCAL ")
	}
	f.WriteString("AUTHORIZATION ")
	if s.Default {
		f.WriteString("DEFAULT")
	} else if s.User != nil {
		f.WriteString(s.User.String())
	}
	return f.String()
}

// ============================================================================
// SET ROLE
// ============================================================================

// SetRole represents a SET ROLE statement.
// Syntax: SET [ SESSION | LOCAL ] ROLE { role_name | NONE }
type SetRole struct {
	BaseStatement
	Local   bool
	Session bool
	Role    *ast.Ident
	None    bool
}

func (s *SetRole) statementNode() {}

func (s *SetRole) String() string {
	var f strings.Builder
	f.WriteString("SET ")
	if s.Session {
		f.WriteString("SESSION ")
	}
	if s.Local {
		f.WriteString("LOCAL ")
	}
	f.WriteString("ROLE ")
	if s.None {
		f.WriteString("NONE")
	} else if s.Role != nil {
		f.WriteString(s.Role.String())
	}
	return f.String()
}
