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

// Package ast provides consolidated SQL statement AST types.
// This file contains statement types migrated from the statement/ subpackage.
//
// Naming convention: S-prefix types (e.g., SGrant, SCreateTable) are the
// consolidated versions that use the main ast package interfaces.
package ast

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/token"
)

// ============================================================================
// Base Types and Helpers
// ============================================================================

// SBaseStatement provides common fields for all consolidated statements
type SBaseStatement struct {
	BaseStatement
}

// DisplaySeparated is a helper for formatting slices with separators
type DisplaySeparated struct {
	Slice []fmt.Stringer
	Sep   string
}

func (d DisplaySeparated) String() string {
	var parts []string
	for _, s := range d.Slice {
		parts = append(parts, s.String())
	}
	return strings.Join(parts, d.Sep)
}

// DisplayCommaSeparated returns a comma-separated display helper
func DisplayCommaSeparated(slice []fmt.Stringer) DisplaySeparated {
	return DisplaySeparated{Slice: slice, Sep: ", "}
}

// helper functions for SQL generation
func writeIfNotEmpty(f *strings.Builder, prefix, value string) {
	if value != "" {
		f.WriteString(prefix)
		f.WriteString(value)
	}
}

func writeIfTrue(f *strings.Builder, condition bool, value string) {
	if condition {
		f.WriteString(value)
	}
}

func writeOptional(f *strings.Builder, prefix string, value fmt.Stringer) {
	if value != nil {
		f.WriteString(prefix)
		f.WriteString(value.String())
	}
}

// Helper function to format idents slice
func formatIdents(idents []*Ident, sep string) string {
	if len(idents) == 0 {
		return ""
	}
	parts := make([]string, len(idents))
	for i, ident := range idents {
		parts[i] = ident.String()
	}
	return strings.Join(parts, sep)
}

// Helper function to format object names slice
func formatObjectNames(names []*ObjectName, sep string) string {
	if len(names) == 0 {
		return ""
	}
	parts := make([]string, len(names))
	for i, name := range names {
		parts[i] = name.String()
	}
	return strings.Join(parts, sep)
}

// ============================================================================
// DCL Statements - GRANT/REVOKE
// ============================================================================

// SGrant represents a GRANT statement
type SGrant struct {
	SBaseStatement
	Privileges      *SPrivileges
	Objects         *SGrantObjects
	Grantees        []*SGrantee
	WithGrantOption bool
	AsGrantor       *Ident
	GrantedBy       *Ident
	CurrentGrants   *SCurrentGrantsKind
}

func (g *SGrant) statementNode() {}

func (g *SGrant) String() string {
	var f strings.Builder
	f.WriteString("GRANT ")
	if g.Privileges != nil {
		f.WriteString(g.Privileges.String())
	}
	if g.Objects != nil {
		f.WriteString(" ON ")
		f.WriteString(g.Objects.String())
	}
	f.WriteString(" TO ")
	for i, grantee := range g.Grantees {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(grantee.String())
	}
	if g.WithGrantOption {
		f.WriteString(" WITH GRANT OPTION")
	}
	if g.AsGrantor != nil {
		f.WriteString(" AS ")
		f.WriteString(g.AsGrantor.String())
	}
	if g.GrantedBy != nil {
		f.WriteString(" GRANTED BY ")
		f.WriteString(g.GrantedBy.String())
	}
	return f.String()
}

// SRevoke represents a REVOKE statement
type SRevoke struct {
	SBaseStatement
	Privileges    *SPrivileges
	Objects       *SGrantObjects
	Grantees      []*SGrantee
	GrantedBy     *Ident
	Cascade       bool
	Restrict      bool
	CurrentGrants *SCurrentGrantsKind
}

func (r *SRevoke) statementNode() {}

func (r *SRevoke) String() string {
	var f strings.Builder
	f.WriteString("REVOKE ")
	if r.Privileges != nil {
		f.WriteString(r.Privileges.String())
	}
	if r.Objects != nil {
		f.WriteString(" ON ")
		f.WriteString(r.Objects.String())
	}
	f.WriteString(" FROM ")
	for i, grantee := range r.Grantees {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(grantee.String())
	}
	if r.GrantedBy != nil {
		f.WriteString(" GRANTED BY ")
		f.WriteString(r.GrantedBy.String())
	}
	if r.Cascade {
		f.WriteString(" CASCADE")
	}
	if r.Restrict {
		f.WriteString(" RESTRICT")
	}
	return f.String()
}

// SCurrentGrantsKind represents the type of current grants
type SCurrentGrantsKind int

const (
	SCurrentGrantsNone SCurrentGrantsKind = iota
	SCurrentGrantsCopy
	SCurrentGrantsRevoke
)

// SPrivileges represents a list of privileges in a GRANT/REVOKE statement
type SPrivileges struct {
	All                   bool
	WithPrivilegesKeyword bool
	Actions               []*SAction
}

func (p *SPrivileges) String() string {
	var f strings.Builder
	if p.All {
		f.WriteString("ALL")
		if p.WithPrivilegesKeyword {
			f.WriteString(" PRIVILEGES")
		}
	} else if len(p.Actions) > 0 {
		for i, action := range p.Actions {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(action.String())
		}
	}
	return f.String()
}

// SAction represents a specific privilege action in a GRANT/REVOKE statement
type SAction struct {
	ActionType ActionType
	RawKeyword string
	Columns    []*Ident
}

func (a *SAction) String() string {
	var f strings.Builder
	if a.RawKeyword != "" {
		f.WriteString(a.RawKeyword)
	} else {
		f.WriteString(a.ActionType.String())
	}
	if len(a.Columns) > 0 {
		f.WriteString(" (")
		f.WriteString(formatIdents(a.Columns, ", "))
		f.WriteString(")")
	}
	return f.String()
}

// ActionType represents the type of privilege
type ActionType int

const (
	ActionTypeConnect ActionType = iota
	ActionTypeCreate
	ActionTypeDelete
	ActionTypeDrop
	ActionTypeExecute
	ActionTypeInsert
	ActionTypeReferences
	ActionTypeSelect
	ActionTypeTemporary
	ActionTypeTrigger
	ActionTypeTruncate
	ActionTypeUpdate
	ActionTypeUsage
)

func (a ActionType) String() string {
	switch a {
	case ActionTypeConnect:
		return "CONNECT"
	case ActionTypeCreate:
		return "CREATE"
	case ActionTypeDelete:
		return "DELETE"
	case ActionTypeDrop:
		return "DROP"
	case ActionTypeExecute:
		return "EXECUTE"
	case ActionTypeInsert:
		return "INSERT"
	case ActionTypeReferences:
		return "REFERENCES"
	case ActionTypeSelect:
		return "SELECT"
	case ActionTypeTemporary:
		return "TEMPORARY"
	case ActionTypeTrigger:
		return "TRIGGER"
	case ActionTypeTruncate:
		return "TRUNCATE"
	case ActionTypeUpdate:
		return "UPDATE"
	case ActionTypeUsage:
		return "USAGE"
	default:
		return ""
	}
}

// SGrantObjects represents the objects on which privileges are granted
type SGrantObjects struct {
	ObjectType SGrantObjectType
	Tables     []*ObjectName
	Schemas    []*ObjectName
}

type SGrantObjectType int

const (
	SGrantObjectTypeTables SGrantObjectType = iota
	SGrantObjectTypeSequences
	SGrantObjectTypeDatabases
	SGrantObjectTypeSchemas
	SGrantObjectTypeViews
	SGrantObjectTypeAllSequencesInSchema
	SGrantObjectTypeAllTablesInSchema
	SGrantObjectTypeAllViewsInSchema
	SGrantObjectTypeFutureTablesInSchema
)

func (g *SGrantObjects) String() string {
	var f strings.Builder
	switch g.ObjectType {
	case SGrantObjectTypeTables:
		if len(g.Tables) > 0 {
			f.WriteString(formatObjectNames(g.Tables, ", "))
		}
	case SGrantObjectTypeSequences:
		f.WriteString("SEQUENCE ")
		f.WriteString(formatObjectNames(g.Tables, ", "))
	case SGrantObjectTypeDatabases:
		f.WriteString("DATABASE ")
		f.WriteString(formatObjectNames(g.Tables, ", "))
	case SGrantObjectTypeSchemas:
		f.WriteString("SCHEMA ")
		f.WriteString(formatObjectNames(g.Tables, ", "))
	case SGrantObjectTypeViews:
		f.WriteString("VIEW ")
		f.WriteString(formatObjectNames(g.Tables, ", "))
	case SGrantObjectTypeAllTablesInSchema:
		f.WriteString("ALL TABLES IN SCHEMA ")
		f.WriteString(formatObjectNames(g.Schemas, ", "))
	case SGrantObjectTypeAllSequencesInSchema:
		f.WriteString("ALL SEQUENCES IN SCHEMA ")
		f.WriteString(formatObjectNames(g.Schemas, ", "))
	case SGrantObjectTypeAllViewsInSchema:
		f.WriteString("ALL VIEWS IN SCHEMA ")
		f.WriteString(formatObjectNames(g.Schemas, ", "))
	}
	return f.String()
}

// SGrantee represents a grantee (user or role)
type SGrantee struct {
	GranteeType SGranteesType
	Name        *SGranteeName
}

type SGranteesType int

const (
	SGranteesTypeNone SGranteesType = iota
	SGranteesTypeRole
	SGranteesTypeUser
	SGranteesTypeShare
	SGranteesTypeGroup
	SGranteesTypePublic
	SGranteesTypeDatabaseRole
	SGranteesTypeApplication
	SGranteesTypeApplicationRole
)

func (g SGranteesType) String() string {
	switch g {
	case SGranteesTypeRole:
		return "ROLE "
	case SGranteesTypeUser:
		return "USER "
	case SGranteesTypeShare:
		return "SHARE "
	case SGranteesTypeGroup:
		return "GROUP "
	case SGranteesTypePublic:
		return "PUBLIC"
	case SGranteesTypeDatabaseRole:
		return "DATABASE ROLE "
	case SGranteesTypeApplication:
		return "APPLICATION "
	case SGranteesTypeApplicationRole:
		return "APPLICATION ROLE "
	default:
		return ""
	}
}

func (g *SGrantee) String() string {
	var f strings.Builder
	f.WriteString(g.GranteeType.String())
	if g.Name != nil {
		f.WriteString(g.Name.String())
	}
	return f.String()
}

// SGranteeName represents the name of a grantee
type SGranteeName struct {
	User       *Ident
	Host       *Ident
	ObjectName *ObjectName
}

func (g *SGranteeName) String() string {
	if g.User != nil && g.Host != nil {
		return g.User.String() + "@" + g.Host.String()
	}
	if g.ObjectName != nil {
		return g.ObjectName.String()
	}
	return ""
}

// ============================================================================
// DCL - Other Statements
// ============================================================================

// SCascadeOption represents CASCADE or RESTRICT options
type SCascadeOption int

const (
	SCascadeNone SCascadeOption = iota
	SCascade
	SRestrict
)

func (c SCascadeOption) String() string {
	switch c {
	case SCascade:
		return "CASCADE"
	case SRestrict:
		return "RESTRICT"
	default:
		return ""
	}
}

// SUse represents a USE statement
type SUse struct {
	SBaseStatement
	DbName *Ident
}

func (u *SUse) statementNode() {}

func (u *SUse) String() string {
	return "USE " + u.DbName.String()
}

// ============================================================================
// DDL - CREATE TABLE
// ============================================================================

// SCreateTable represents a CREATE TABLE statement
type SCreateTable struct {
	SBaseStatement
	OrReplace                  bool
	Temporary                  bool
	External                   bool
	Dynamic                    bool
	Global                     *bool
	IfNotExists                bool
	Transient                  bool
	Volatile                   bool
	Iceberg                    bool
	Snapshot                   bool
	Name                       *ObjectName
	Columns                    []*SColumnDef
	Constraints                []*STableConstraint
	HiveDistribution           *SHiveDistributionStyle
	HiveFormats                *SHiveFormat
	TableOptions               *SCreateTableOptions
	FileFormat                 *SFileFormat
	Location                   *string
	Query                      Query
	WithoutRowid               bool
	Like                       *SCreateTableLikeKind
	Clone                      *ObjectName
	Version                    *STableVersion
	Comment                    *SCommentDef
	OnCommit                   *SOnCommit
	OnCluster                  *Ident
	PrimaryKey                 Expr
	OrderBy                    *SOneOrManyWithParens
	PartitionBy                Expr
	ClusterBy                  *SWrappedCollection
	ClusteredBy                *SClusteredBy
	Inherits                   []*ObjectName
	PartitionOf                *ObjectName
	ForValues                  *SForValues
	Strict                     bool
	CopyGrants                 bool
	EnableSchemaEvolution      *bool
	ChangeTracking             *bool
	DataRetentionTimeInDays    *uint64
	MaxDataExtensionTimeInDays *uint64
	DefaultDdlCollation        *string
	WithAggregationPolicy      *ObjectName
	WithRowAccessPolicy        *SRowAccessPolicy
	WithStorageLifecyclePolicy *SStorageLifecyclePolicy
	WithTags                   []*STag
	ExternalVolume             *string
	BaseLocation               *string
	Catalog                    *string
	CatalogSync                *string
	StorageSerializationPolicy *SStorageSerializationPolicy
	TargetLag                  *string
	Warehouse                  *Ident
	RefreshMode                *SRefreshModeKind
	Initialize                 *SInitializeKind
}

func (c *SCreateTable) statementNode() {}

func (c *SCreateTable) String() string {
	var f strings.Builder
	f.WriteString("CREATE ")
	if c.OrReplace {
		f.WriteString("OR REPLACE ")
	}
	if c.Temporary {
		f.WriteString("TEMPORARY ")
	}
	if c.Transient {
		f.WriteString("TRANSIENT ")
	}
	if c.External {
		f.WriteString("EXTERNAL ")
	}
	if c.Volatile {
		f.WriteString("VOLATILE ")
	}
	if c.Dynamic {
		f.WriteString("DYNAMIC ")
	}
	if c.Iceberg {
		f.WriteString("ICEBERG ")
	}
	if c.Snapshot {
		f.WriteString("SNAPSHOT ")
	}
	if c.Global != nil {
		if *c.Global {
			f.WriteString("GLOBAL ")
		} else {
			f.WriteString("LOCAL ")
		}
	}
	f.WriteString("TABLE ")
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	f.WriteString(c.Name.String())
	if len(c.Columns) > 0 {
		f.WriteString(" (")
		for i, col := range c.Columns {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(col.String())
		}
		f.WriteString(")")
	}
	if c.Comment != nil {
		f.WriteString(" COMMENT '")
		f.WriteString(c.Comment.String())
		f.WriteString("'")
	}
	if c.Query != nil {
		f.WriteString(" AS ")
		f.WriteString(c.Query.String())
	}
	return f.String()
}

// SCreateView represents a CREATE VIEW statement
type SCreateView struct {
	SBaseStatement
	OrReplace           bool
	OrAlter             bool
	Materialized        bool
	IfNotExists         bool
	Temporary           bool
	Name                *ObjectName
	Columns             []*Ident
	Query               Query
	Options             []*SSqlOption
	ClusterBy           []Expr
	WithNoSchemaBinding bool
	WithNoData          bool
	Comment             *SCommentDef
	Versioned           bool
	Envelope            *SViewEnvelope
	QueryArena          bool
	Params              *SCreateViewParams
	Backup              bool
	Watermark           Expr
	To                  Statement
}

func (c *SCreateView) statementNode() {}

func (c *SCreateView) String() string {
	var f strings.Builder
	f.WriteString("CREATE ")
	if c.OrReplace {
		f.WriteString("OR REPLACE ")
	}
	if c.Temporary {
		f.WriteString("TEMPORARY ")
	}
	if c.Materialized {
		f.WriteString("MATERIALIZED ")
	}
	f.WriteString("VIEW ")
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	f.WriteString(c.Name.String())
	if len(c.Columns) > 0 {
		f.WriteString(" (")
		f.WriteString(formatIdents(c.Columns, ", "))
		f.WriteString(")")
	}
	if c.Query != nil {
		f.WriteString(" AS ")
		f.WriteString(c.Query.String())
	}
	return f.String()
}

// ============================================================================
// DDL - DROP Statements
// ============================================================================

// SDrop represents a DROP statement
type SDrop struct {
	SBaseStatement
	ObjectType SObjectType
	IfExists   bool
	Names      []*ObjectName
	Cascade    bool
	Restrict   bool
	Purge      bool
	Temporary  bool
	Table      *ObjectName
}

func (d *SDrop) statementNode() {}

func (d *SDrop) String() string {
	var f strings.Builder
	f.WriteString("DROP ")
	f.WriteString(d.ObjectType.String())
	f.WriteString(" ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	f.WriteString(formatObjectNames(d.Names, ", "))
	if d.Cascade {
		f.WriteString(" CASCADE")
	}
	if d.Restrict {
		f.WriteString(" RESTRICT")
	}
	return f.String()
}

// SObjectType represents the type of object to drop
type SObjectType int

const (
	SObjectTypeTable SObjectType = iota
	SObjectTypeView
	SObjectTypeIndex
	SObjectTypeSchema
	SObjectTypeDatabase
	SObjectTypeRole
	SObjectTypeSequence
	SObjectTypeStage
	SObjectTypeType
	SObjectTypeFunction
	SObjectTypeProcedure
	SObjectTypeTrigger
)

func (o SObjectType) String() string {
	switch o {
	case SObjectTypeTable:
		return "TABLE"
	case SObjectTypeView:
		return "VIEW"
	case SObjectTypeIndex:
		return "INDEX"
	case SObjectTypeSchema:
		return "SCHEMA"
	case SObjectTypeDatabase:
		return "DATABASE"
	case SObjectTypeRole:
		return "ROLE"
	case SObjectTypeSequence:
		return "SEQUENCE"
	case SObjectTypeStage:
		return "STAGE"
	case SObjectTypeType:
		return "TYPE"
	case SObjectTypeFunction:
		return "FUNCTION"
	case SObjectTypeProcedure:
		return "PROCEDURE"
	case SObjectTypeTrigger:
		return "TRIGGER"
	default:
		return ""
	}
}

// ============================================================================
// DDL - ALTER Statements
// ============================================================================

// SAlterTable represents an ALTER TABLE statement
type SAlterTable struct {
	SBaseStatement
	Name       *ObjectName
	IfExists   bool
	Only       bool
	Operations []*SAlterTableOperation
	Location   *SHiveSetLocation
}

func (a *SAlterTable) statementNode() {}

func (a *SAlterTable) String() string {
	var f strings.Builder
	f.WriteString("ALTER TABLE ")
	if a.IfExists {
		f.WriteString("IF EXISTS ")
	}
	if a.Only {
		f.WriteString("ONLY ")
	}
	f.WriteString(a.Name.String())
	for i, op := range a.Operations {
		if i > 0 {
			f.WriteString(", ")
		} else {
			f.WriteString(" ")
		}
		f.WriteString(op.String())
	}
	return f.String()
}

// ============================================================================
// DML - Query, Insert, Update, Delete, Merge
// ============================================================================

// SQuery wraps a SELECT statement as a standalone statement
type SQuery struct {
	SBaseStatement
	Query Query
}

func (q *SQuery) statementNode() {}

func (q *SQuery) String() string {
	if q.Query != nil {
		return q.Query.String()
	}
	return ""
}

// SInsert represents an INSERT statement
type SInsert struct {
	SBaseStatement
	OptimizerHints        []*SOptimizerHint
	Or                    *SSqliteOnConflict
	Ignore                bool
	Into                  bool
	Table                 *ObjectName
	TableAlias            *Ident
	Columns               []*Ident
	Overwrite             bool
	Source                Query
	Assignments           []*SAssignment
	Partitioned           []Expr
	AfterColumns          []*Ident
	HasTableKeyword       bool
	On                    *SOnInsert
	Returning             []Expr
	Output                *SOutputClause
	ReplaceInto           bool
	Priority              *SMysqlInsertPriority
	InsertAlias           *SInsertAliases
	Settings              []*SSetting
	FormatClause          *SInputFormatClause
	MultiTableInsertType  *SMultiTableInsertType
	MultiTableIntoClauses []*SMultiTableInsertIntoClause
	MultiTableWhenClauses []*SMultiTableInsertWhenClause
	MultiTableElseClause  []*SMultiTableInsertIntoClause
	DefaultValues         bool
}

func (i *SInsert) statementNode() {}

func (i *SInsert) String() string {
	var f strings.Builder
	if i.ReplaceInto {
		f.WriteString("REPLACE")
	} else {
		f.WriteString("INSERT")
	}
	for _, hint := range i.OptimizerHints {
		f.WriteString(" ")
		f.WriteString(hint.String())
	}
	if i.Priority != nil {
		f.WriteString(" ")
		f.WriteString(i.Priority.String())
	}
	if i.Ignore {
		f.WriteString(" IGNORE")
	}
	if i.Overwrite {
		f.WriteString(" OVERWRITE")
	}
	if i.MultiTableInsertType != nil {
		f.WriteString(" ")
		f.WriteString(i.MultiTableInsertType.String())
	}
	if i.Into {
		f.WriteString(" INTO")
	}
	if i.HasTableKeyword {
		f.WriteString(" TABLE")
	}
	f.WriteString(" ")
	f.WriteString(i.Table.String())
	if len(i.Columns) > 0 {
		f.WriteString(" (")
		f.WriteString(formatIdents(i.Columns, ", "))
		f.WriteString(")")
	}
	if i.DefaultValues {
		f.WriteString(" DEFAULT VALUES")
	} else if i.Source != nil {
		f.WriteString(" ")
		f.WriteString(i.Source.String())
	}
	return f.String()
}

// SUpdate represents an UPDATE statement
type SUpdate struct {
	SBaseStatement
	Table           *ObjectName
	TableAlias      *Ident
	Assignments     []*SAssignment
	From            TableFactor
	Selection       Expr
	Returning       []Expr
	Output          *SOutputClause
	OrderBy         []*OrderByExpr
	Limit           LimitClause
	IsFromStatement bool
	Setting         []*SSetting
}

func (u *SUpdate) statementNode() {}

func (u *SUpdate) String() string {
	var f strings.Builder
	f.WriteString("UPDATE ")
	if u.Table != nil {
		f.WriteString(u.Table.String())
	}
	if len(u.Assignments) > 0 {
		f.WriteString(" SET ")
		for i, assign := range u.Assignments {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(assign.String())
		}
	}
	if u.Selection != nil {
		f.WriteString(" WHERE ")
		f.WriteString(u.Selection.String())
	}
	return f.String()
}

// SDelete represents a DELETE statement
type SDelete struct {
	SBaseStatement
	Tables    []*ObjectName
	Using     []TableFactor
	Selection Expr
	Returning []Expr
	Output    *SOutputClause
	OrderBy   []Expr
	Limit     LimitClause
}

func (d *SDelete) statementNode() {}

func (d *SDelete) String() string {
	var f strings.Builder
	f.WriteString("DELETE FROM ")
	for i, table := range d.Tables {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(table.String())
	}
	if d.Selection != nil {
		f.WriteString(" WHERE ")
		f.WriteString(d.Selection.String())
	}
	return f.String()
}

// SMerge represents a MERGE statement
type SMerge struct {
	SBaseStatement
	Into       bool
	Table      TableFactor
	TableAlias *Ident
	Source     TableFactor
	On         Expr
	Clauses    []*SMergeClause
	Output     *SOutputClause
}

func (m *SMerge) statementNode() {}

func (m *SMerge) String() string {
	var f strings.Builder
	f.WriteString("MERGE")
	if m.Into {
		f.WriteString(" INTO")
	}
	if m.Table.Kind != 0 || m.Table.Data.Table != nil {
		f.WriteString(" ")
		f.WriteString(m.Table.String())
	}
	f.WriteString(" USING ")
	f.WriteString(m.Source.String())
	f.WriteString(" ON ")
	f.WriteString(m.On.String())
	return f.String()
}

// ============================================================================
// Misc Statements
// ============================================================================

// SSet represents a SET statement
type SSet struct {
	SBaseStatement
	Variable *ObjectName
	Values   []Expr
	Local    bool
	Session  bool
	Global   bool
	TimeZone bool
	HiveVar  bool
	Scope    *SSetScope
}

func (s *SSet) statementNode() {}

func (s *SSet) String() string {
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
	if len(s.Values) > 0 {
		parts := make([]string, len(s.Values))
		for i, v := range s.Values {
			parts[i] = v.String()
		}
		f.WriteString(strings.Join(parts, ", "))
	}
	return f.String()
}

// SExplain represents an EXPLAIN statement
type SExplain struct {
	SBaseStatement
	DescribeAlias SDescribeAlias
	Analyze       bool
	Verbose       bool
	QueryPlan     bool
	Estimate      bool
	Statement     Statement
	Format        *SAnalyzeFormatKind
	Options       []*SUtilityOption
}

func (e *SExplain) statementNode() {}

func (e *SExplain) String() string {
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
	if e.Statement != nil {
		f.WriteString(e.Statement.String())
	}
	return f.String()
}

// SDescribeAlias represents DESCRIBE, EXPLAIN, or DESC
type SDescribeAlias int

const (
	SDescribeAliasDescribe SDescribeAlias = iota
	SDescribeAliasExplain
	SDescribeAliasDesc
)

func (d SDescribeAlias) String() string {
	switch d {
	case SDescribeAliasDescribe:
		return "DESCRIBE"
	case SDescribeAliasExplain:
		return "EXPLAIN"
	case SDescribeAliasDesc:
		return "DESC"
	default:
		return ""
	}
}

// ============================================================================
// Transaction Statements
// ============================================================================

// SStartTransaction represents a START TRANSACTION / BEGIN statement
type SStartTransaction struct {
	SBaseStatement
	Modes         []*STransactionMode
	Begin         bool
	Transaction   *SBeginTransactionKind
	Modifier      *STransactionModifier
	Statements    []Statement
	Exception     []*SExceptionWhen
	HasEndKeyword bool
}

func (s *SStartTransaction) statementNode() {}

func (s *SStartTransaction) String() string {
	var f strings.Builder
	if s.Begin {
		f.WriteString("BEGIN")
	} else {
		f.WriteString("START TRANSACTION")
	}
	return f.String()
}

// SCommit represents a COMMIT statement
type SCommit struct {
	SBaseStatement
	Chain    bool
	End      bool
	Modifier *STransactionModifier
}

func (c *SCommit) statementNode() {}

func (c *SCommit) String() string {
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

// SRollback represents a ROLLBACK statement
type SRollback struct {
	SBaseStatement
	Chain     bool
	Savepoint *Ident
}

func (r *SRollback) statementNode() {}

func (r *SRollback) String() string {
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

// SSavepoint represents a SAVEPOINT statement
type SSavepoint struct {
	SBaseStatement
	Name *Ident
}

func (s *SSavepoint) statementNode() {}

func (s *SSavepoint) String() string {
	return "SAVEPOINT " + s.Name.String()
}

// ============================================================================
// Helper Types (placeholders - full implementations will be in separate files)
// ============================================================================

// SColumnDef represents a column definition (placeholder)
type SColumnDef struct {
	Name     *Ident
	DataType DataType
	span     token.Span
}

func (c *SColumnDef) Span() token.Span { return c.span }
func (c *SColumnDef) String() string   { return c.Name.String() }

// STableConstraint represents a table constraint (placeholder)
type STableConstraint struct {
	Name       *Ident
	Constraint interface{}
}

func (t *STableConstraint) String() string { return "CONSTRAINT" }

// SHiveDistributionStyle, SHiveFormat, SCreateTableOptions, SFileFormat, etc.
// are placeholder types that will be defined in detail in a follow-up

type SHiveDistributionStyle struct{}
type SHiveFormat struct{}
type SCreateTableOptions struct{}
type SFileFormat struct{}
type SCreateTableLikeKind struct{}
type STableVersion struct{}
type SCommentDef struct{}
type SOnCommit struct{}
type SOneOrManyWithParens struct{}
type SWrappedCollection struct{}
type SClusteredBy struct{}
type SForValues struct{}
type SRowAccessPolicy struct{}
type SStorageLifecyclePolicy struct{}
type STag struct{}
type SStorageSerializationPolicy struct{}
type SRefreshModeKind struct{}
type SInitializeKind struct{}
type SViewEnvelope struct{}
type SCreateViewParams struct{}
type SSqlOption struct{}
type SAlterTableOperation struct{}
type SHiveSetLocation struct{}
type SOptimizerHint struct{}
type SSqliteOnConflict struct{}
type SAssignment struct{}
type SOnInsert struct{}
type SOutputClause struct{}
type SMysqlInsertPriority struct{}
type SInsertAliases struct{}
type SSetting struct{}
type SInputFormatClause struct{}
type SMultiTableInsertType struct{}
type SMultiTableInsertIntoClause struct{}
type SMultiTableInsertWhenClause struct{}
type SMergeClause struct{}
type SSetScope struct{}
type SAnalyzeFormatKind struct{}
type SUtilityOption struct{}
type STransactionMode struct{}
type SBeginTransactionKind struct{}
type STransactionModifier struct{}
type SExceptionWhen struct{}

// String methods for placeholder types
func (s *SHiveDistributionStyle) String() string  { return "" }
func (s *SHiveFormat) String() string             { return "" }
func (s *SCreateTableOptions) String() string     { return "" }
func (s *SFileFormat) String() string             { return "" }
func (s *SCreateTableLikeKind) String() string    { return "" }
func (s *STableVersion) String() string           { return "" }
func (s *SCommentDef) String() string             { return "" }
func (s *SOnCommit) String() string               { return "" }
func (s *SOneOrManyWithParens) String() string    { return "" }
func (s *SWrappedCollection) String() string      { return "" }
func (s *SClusteredBy) String() string            { return "" }
func (s *SForValues) String() string              { return "" }
func (s *SRowAccessPolicy) String() string        { return "" }
func (s *SStorageLifecyclePolicy) String() string { return "" }
func (s *STag) String() string                    { return "" }
func (s *SStorageSerializationPolicy) String() string {
	return ""
}
func (s *SRefreshModeKind) String() string      { return "" }
func (s *SInitializeKind) String() string       { return "" }
func (s *SViewEnvelope) String() string         { return "" }
func (s *SCreateViewParams) String() string     { return "" }
func (s *SSqlOption) String() string            { return "" }
func (s *SAlterTableOperation) String() string  { return "" }
func (s *SHiveSetLocation) String() string      { return "" }
func (s *SOptimizerHint) String() string        { return "" }
func (s *SSqliteOnConflict) String() string     { return "" }
func (s *SAssignment) String() string           { return "" }
func (s *SOnInsert) String() string             { return "" }
func (s *SOutputClause) String() string         { return "" }
func (s *SMysqlInsertPriority) String() string  { return "" }
func (s *SInsertAliases) String() string        { return "" }
func (s *SSetting) String() string              { return "" }
func (s *SInputFormatClause) String() string    { return "" }
func (s *SMultiTableInsertType) String() string { return "" }
func (s *SMultiTableInsertIntoClause) String() string {
	return ""
}
func (s *SMultiTableInsertWhenClause) String() string { return "" }
func (s *SMergeClause) String() string                { return "" }
func (s *SSetScope) String() string                   { return "" }
func (s *SAnalyzeFormatKind) String() string          { return "" }
func (s *SUtilityOption) String() string              { return "" }
func (s *STransactionMode) String() string            { return "" }
func (s *SBeginTransactionKind) String() string       { return "" }
func (s *STransactionModifier) String() string        { return "" }
func (s *SExceptionWhen) String() string              { return "" }

// Ensure S-prefix statement types implement Statement interface
var _ Statement = (*SGrant)(nil)
var _ Statement = (*SRevoke)(nil)
var _ Statement = (*SUse)(nil)
var _ Statement = (*SCreateTable)(nil)
var _ Statement = (*SCreateView)(nil)
var _ Statement = (*SDrop)(nil)
var _ Statement = (*SAlterTable)(nil)
var _ Statement = (*SQuery)(nil)
var _ Statement = (*SInsert)(nil)
var _ Statement = (*SUpdate)(nil)
var _ Statement = (*SDelete)(nil)
var _ Statement = (*SMerge)(nil)
var _ Statement = (*SSet)(nil)
var _ Statement = (*SExplain)(nil)
var _ Statement = (*SStartTransaction)(nil)
var _ Statement = (*SCommit)(nil)
var _ Statement = (*SRollback)(nil)
var _ Statement = (*SSavepoint)(nil)
