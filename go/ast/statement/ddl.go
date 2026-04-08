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
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/query"
)

// ============================================================================
// CREATE TABLE
// ============================================================================

// CreateTable represents a CREATE TABLE statement
type CreateTable struct {
	BaseStatement
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
	Name                       *ast.ObjectName
	Columns                    []*expr.ColumnDef
	Constraints                []*expr.TableConstraint
	HiveDistribution           *expr.HiveDistributionStyle
	HiveFormats                *expr.HiveFormat
	TableOptions               *expr.CreateTableOptions
	FileFormat                 *expr.FileFormat
	Location                   *string
	Query                      *query.Query
	WithoutRowid               bool
	Like                       *expr.CreateTableLikeKind
	AsTable                    *ast.ObjectName // PostgreSQL: CREATE TABLE x AS TABLE y
	Clone                      *ast.ObjectName
	Version                    *expr.TableVersion
	Comment                    *expr.CommentDef
	OnCommit                   *expr.OnCommit
	OnCluster                  *ast.Ident
	PrimaryKey                 expr.Expr
	OrderBy                    *expr.OneOrManyWithParens
	PartitionBy                expr.Expr
	ClusterBy                  *expr.WrappedCollection
	ClusteredBy                *expr.ClusteredBy
	Inherits                   []*ast.ObjectName
	PartitionOf                *ast.ObjectName
	ForValues                  *expr.ForValues
	Strict                     bool
	CopyGrants                 bool
	EnableSchemaEvolution      *bool
	ChangeTracking             *bool
	DataRetentionTimeInDays    *uint64
	MaxDataExtensionTimeInDays *uint64
	DefaultDdlCollation        *string
	WithAggregationPolicy      *ast.ObjectName
	WithRowAccessPolicy        *expr.RowAccessPolicy
	WithStorageLifecyclePolicy *expr.StorageLifecyclePolicy
	WithTags                   []*expr.Tag
	ExternalVolume             *string
	BaseLocation               *string
	Catalog                    *string
	CatalogSync                *string
	StorageSerializationPolicy *expr.StorageSerializationPolicy
	TargetLag                  *string
	Warehouse                  *ast.Ident
	RefreshMode                *expr.RefreshModeKind
	Initialize                 *expr.InitializeKind
	RequireUser                bool
	// Redshift-specific fields
	Backup    *bool
	Diststyle *expr.DistStyle
	Distkey   expr.Expr
	Sortkey   []expr.Expr
}

func (c *CreateTable) statementNode() {}

func (c *CreateTable) String() string {
	var f strings.Builder
	f.WriteString("CREATE ")
	if c.OrReplace {
		f.WriteString("OR REPLACE ")
	}
	// LOCAL/GLOBAL comes before TEMPORARY (per SQL standard and Snowflake)
	if c.Global != nil {
		if *c.Global {
			f.WriteString("GLOBAL ")
		} else {
			f.WriteString("LOCAL ")
		}
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
	f.WriteString("TABLE ")
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	f.WriteString(c.Name.String())

	// ClickHouse ON CLUSTER clause
	if c.OnCluster != nil {
		f.WriteString(" ON CLUSTER ")
		f.WriteString(c.OnCluster.String())
	}

	// LIKE clause (Snowflake/BigQuery style: CREATE TABLE new LIKE old)
	if c.Like != nil {
		f.WriteString(" ")
		f.WriteString(c.Like.String())
	}

	// PostgreSQL PARTITION OF
	if c.PartitionOf != nil {
		f.WriteString(" PARTITION OF ")
		f.WriteString(c.PartitionOf.String())
		if c.ForValues != nil {
			f.WriteString(" ")
			f.WriteString(c.ForValues.String())
		}
	}

	if len(c.Columns) > 0 {
		f.WriteString(" (")
		for i, col := range c.Columns {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(col.String())
		}
		for _, constraint := range c.Constraints {
			f.WriteString(", ")
			f.WriteString(constraint.String())
		}
		f.WriteString(")")
	}

	// PostgreSQL INHERITS clause
	if len(c.Inherits) > 0 {
		f.WriteString(" INHERITS (")
		for i, parent := range c.Inherits {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(parent.String())
		}
		f.WriteString(")")
	}

	if c.Comment != nil {
		f.WriteString(" COMMENT '")
		f.WriteString(c.Comment.String())
		f.WriteString("'")
	}

	// Output table options (MySQL-specific: ENGINE, CHARSET, COLLATE, etc.)
	if c.TableOptions != nil && c.TableOptions.Type == expr.CreateTableOptionsPlain {
		for _, opt := range c.TableOptions.Options {
			f.WriteString(" ")
			f.WriteString(opt.String())
		}
	}

	// Hive formats: STORED AS and LOCATION
	if c.HiveFormats != nil {
		if c.HiveFormats.Storage != nil {
			if c.HiveFormats.Storage.InputFormat == c.HiveFormats.Storage.OutputFormat {
				// Simple format like TEXTFILE, PARQUET - use uppercase
				f.WriteString(fmt.Sprintf(" STORED AS %s", strings.ToUpper(c.HiveFormats.Storage.InputFormat)))
			} else {
				// Complex format with INPUTFORMAT/OUTPUTFORMAT
				f.WriteString(fmt.Sprintf(" STORED AS INPUTFORMAT '%s' OUTPUTFORMAT '%s'",
					c.HiveFormats.Storage.InputFormat, c.HiveFormats.Storage.OutputFormat))
			}
		}
		if c.HiveFormats.Location != nil {
			f.WriteString(fmt.Sprintf(" LOCATION '%s'", *c.HiveFormats.Location))
		}
	}

	// Snowflake-specific: COPY GRANTS
	if c.CopyGrants {
		f.WriteString(" COPY GRANTS")
	}

	// Snowflake-specific: CLUSTER BY
	if c.ClusterBy != nil && len(c.ClusterBy.Items) > 0 {
		f.WriteString(" CLUSTER BY (")
		for i, item := range c.ClusterBy.Items {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(item.String())
		}
		f.WriteString(")")
	}

	// Snowflake-specific: ENABLE_SCHEMA_EVOLUTION
	if c.EnableSchemaEvolution != nil {
		f.WriteString(" ENABLE_SCHEMA_EVOLUTION=")
		if *c.EnableSchemaEvolution {
			f.WriteString("TRUE")
		} else {
			f.WriteString("FALSE")
		}
	}

	// Snowflake-specific: CHANGE_TRACKING
	if c.ChangeTracking != nil {
		f.WriteString(" CHANGE_TRACKING=")
		if *c.ChangeTracking {
			f.WriteString("TRUE")
		} else {
			f.WriteString("FALSE")
		}
	}

	// Snowflake-specific: DATA_RETENTION_TIME_IN_DAYS
	if c.DataRetentionTimeInDays != nil {
		f.WriteString(fmt.Sprintf(" DATA_RETENTION_TIME_IN_DAYS=%d", *c.DataRetentionTimeInDays))
	}

	// Snowflake-specific: ON COMMIT
	if c.OnCommit != nil {
		f.WriteString(" ")
		f.WriteString(c.OnCommit.String())
	}

	// Snowflake ICEBERG options
	if c.ExternalVolume != nil {
		f.WriteString(fmt.Sprintf(" EXTERNAL_VOLUME='%s'", *c.ExternalVolume))
	}
	if c.Catalog != nil {
		f.WriteString(fmt.Sprintf(" CATALOG='%s'", *c.Catalog))
	}
	if c.BaseLocation != nil {
		f.WriteString(fmt.Sprintf(" BASE_LOCATION='%s'", *c.BaseLocation))
	}
	if c.CatalogSync != nil {
		f.WriteString(fmt.Sprintf(" CATALOG_SYNC='%s'", *c.CatalogSync))
	}
	if c.StorageSerializationPolicy != nil {
		f.WriteString(fmt.Sprintf(" STORAGE_SERIALIZATION_POLICY=%s", c.StorageSerializationPolicy.String()))
	}

	// Snowflake DYNAMIC table options
	if c.TargetLag != nil {
		f.WriteString(fmt.Sprintf(" TARGET_LAG='%s'", *c.TargetLag))
	}
	if c.Warehouse != nil {
		f.WriteString(fmt.Sprintf(" WAREHOUSE=%s", c.Warehouse.String()))
	}
	if c.RefreshMode != nil {
		f.WriteString(fmt.Sprintf(" REFRESH_MODE=%s", c.RefreshMode.String()))
	}
	if c.Initialize != nil {
		f.WriteString(fmt.Sprintf(" INITIALIZE=%s", c.Initialize.String()))
	}
	if c.RequireUser {
		f.WriteString(" REQUIRE USER")
	}

	// CLONE clause (Snowflake)
	if c.Clone != nil {
		f.WriteString(" CLONE ")
		f.WriteString(c.Clone.String())
	}

	// WITH options (PostgreSQL/Hive style)
	if c.TableOptions != nil && c.TableOptions.Type == expr.CreateTableOptionsWith {
		f.WriteString(" WITH (")
		for i, opt := range c.TableOptions.Options {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(opt.String())
		}
		f.WriteString(")")
	}

	if c.AsTable != nil {
		f.WriteString(" AS TABLE ")
		f.WriteString(c.AsTable.String())
	}

	if c.Query != nil {
		f.WriteString(" AS ")
		f.WriteString(c.Query.String())
	}

	return f.String()
}

// ============================================================================
// CREATE VIEW
// ============================================================================

// CreateViewAlgorithm represents MySQL CREATE VIEW ALGORITHM options.
// Reference: src/ast/mod.rs CreateViewAlgorithm
type CreateViewAlgorithm int

const (
	CreateViewAlgorithmUndefined CreateViewAlgorithm = iota
	CreateViewAlgorithmMerge
	CreateViewAlgorithmTempTable
)

func (a CreateViewAlgorithm) String() string {
	switch a {
	case CreateViewAlgorithmUndefined:
		return "UNDEFINED"
	case CreateViewAlgorithmMerge:
		return "MERGE"
	case CreateViewAlgorithmTempTable:
		return "TEMPTABLE"
	default:
		return ""
	}
}

// CreateViewSecurity represents MySQL CREATE VIEW SQL SECURITY options.
// Reference: src/ast/mod.rs CreateViewSecurity
type CreateViewSecurity int

const (
	CreateViewSecurityDefiner CreateViewSecurity = iota
	CreateViewSecurityInvoker
)

func (s CreateViewSecurity) String() string {
	switch s {
	case CreateViewSecurityDefiner:
		return "DEFINER"
	case CreateViewSecurityInvoker:
		return "INVOKER"
	default:
		return ""
	}
}

// CreateViewParams represents MySQL CREATE VIEW parameters.
// Reference: src/ast/mod.rs CreateViewParams
type CreateViewParams struct {
	Algorithm *CreateViewAlgorithm
	Definer   *GranteeName
	Security  *CreateViewSecurity
}

func (c *CreateViewParams) String() string {
	var f strings.Builder
	if c.Algorithm != nil {
		f.WriteString("ALGORITHM = ")
		f.WriteString(c.Algorithm.String())
		f.WriteString(" ")
	}
	if c.Definer != nil {
		f.WriteString("DEFINER = ")
		f.WriteString(c.Definer.String())
		f.WriteString(" ")
	}
	if c.Security != nil {
		f.WriteString("SQL SECURITY ")
		f.WriteString(c.Security.String())
		f.WriteString(" ")
	}
	return f.String()
}

// CreateView represents a CREATE VIEW statement
type CreateView struct {
	BaseStatement
	OrReplace           bool
	OrAlter             bool
	Materialized        bool
	Secure              bool
	IfNotExists         bool
	NameBeforeNotExists bool // True when name comes before IF NOT EXISTS (e.g., CREATE VIEW v IF NOT EXISTS...)
	Temporary           bool
	Name                *ast.ObjectName
	Columns             []*expr.ViewColumnDef // View columns with optional options (TAG, POLICY, etc.)
	Query               *query.Query
	QueryIsParens       bool // True when query is wrapped in parentheses: AS (SELECT ...)
	Options             []*expr.SqlOption
	ClusterBy           []expr.Expr
	WithNoSchemaBinding bool
	WithNoData          bool
	Comment             *expr.CommentDef
	Versioned           bool
	Envelope            *expr.ViewEnvelope
	QueryArena          bool
	Params              *CreateViewParams
	Backup              bool
	Watermark           expr.Expr
	To                  ast.Statement
	CopyGrants          bool
}

func (c *CreateView) statementNode() {}

func (c *CreateView) String() string {
	var f strings.Builder
	f.WriteString("CREATE ")
	if c.OrReplace {
		f.WriteString("OR REPLACE ")
	}
	if c.Temporary {
		f.WriteString("TEMPORARY ")
	}
	if c.Secure {
		f.WriteString("SECURE ")
	}
	if c.Materialized {
		f.WriteString("MATERIALIZED ")
	}

	// Output MySQL-style params before VIEW keyword
	if c.Params != nil {
		f.WriteString(c.Params.String())
	}

	f.WriteString("VIEW ")
	if c.IfNotExists {
		if c.NameBeforeNotExists {
			// Name comes before IF NOT EXISTS: CREATE VIEW v IF NOT EXISTS ...
			f.WriteString(c.Name.String())
			f.WriteString(" IF NOT EXISTS")
		} else {
			// IF NOT EXISTS comes before name: CREATE VIEW IF NOT EXISTS v ...
			f.WriteString("IF NOT EXISTS ")
			f.WriteString(c.Name.String())
		}
	} else {
		f.WriteString(c.Name.String())
	}

	if c.CopyGrants {
		f.WriteString(" COPY GRANTS")
	}

	if len(c.Columns) > 0 {
		f.WriteString(" (")
		for i, col := range c.Columns {
			if i > 0 {
				f.WriteString(", ")
			}
			if col != nil {
				f.WriteString(col.String())
			}
		}
		f.WriteString(")")
	}

	// WITH options (if any)
	if len(c.Options) > 0 {
		f.WriteString(" WITH (")
		for i, opt := range c.Options {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(opt.String())
		}
		f.WriteString(")")
	}

	// CLUSTER BY
	if len(c.ClusterBy) > 0 {
		f.WriteString(" CLUSTER BY (")
		f.WriteString(formatExprs(c.ClusterBy, ", "))
		f.WriteString(")")
	}

	// TO clause
	if c.To != nil {
		f.WriteString(" TO ")
		f.WriteString(c.To.String())
	}

	// COMMENT
	if c.Comment != nil {
		f.WriteString(" COMMENT = ")
		f.WriteString(fmt.Sprintf("'%s'", c.Comment.Comment))
	}

	f.WriteString(" AS ")
	if c.QueryIsParens {
		f.WriteString("(")
		f.WriteString(c.Query.String())
		f.WriteString(")")
	} else {
		f.WriteString(c.Query.String())
	}

	// WITH NO SCHEMA BINDING (Redshift)
	if c.WithNoSchemaBinding {
		f.WriteString(" WITH NO SCHEMA BINDING")
	}

	return f.String()
}

// ============================================================================
// CREATE INDEX
// ============================================================================

// CreateIndex represents a CREATE INDEX statement
type CreateIndex struct {
	BaseStatement
	OrReplace      bool
	Unique         bool
	IfNotExists    bool
	Concurrently   bool
	Name           *ast.Ident
	TableName      *ast.ObjectName
	Using          *ast.Ident
	UsingAfterCols bool // true if USING comes after columns (MySQL style)
	Columns        []*expr.IndexColumn
	Include        []*ast.Ident
	NullsDistinct  *bool // nil = not specified, true = NULLS DISTINCT, false = NULLS NOT DISTINCT
	Predicate      expr.Expr
	With           []*expr.SqlOption
	TableSpace     *ast.Ident
	SortedBy       []*expr.OrderByExpr
	IgnoreOrRevert *string
	MySQLOptions   []*expr.SqlOption // MySQL-specific options like LOCK, KEY_BLOCK_SIZE
}

func (c *CreateIndex) statementNode() {}

func (c *CreateIndex) String() string {
	var f strings.Builder
	f.WriteString("CREATE ")
	if c.OrReplace {
		f.WriteString("OR REPLACE ")
	}
	if c.Unique {
		f.WriteString("UNIQUE ")
	}
	f.WriteString("INDEX ")
	if c.Concurrently {
		f.WriteString("CONCURRENTLY ")
	}
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	if c.Name != nil {
		f.WriteString(c.Name.String())
		f.WriteString(" ")
	}
	f.WriteString("ON ")
	f.WriteString(c.TableName.String())

	// Serialize USING before columns (PostgreSQL style) unless it was parsed after columns
	if c.Using != nil && !c.UsingAfterCols {
		f.WriteString(" USING ")
		f.WriteString(c.Using.String())
	}

	f.WriteString(" (")
	for i, col := range c.Columns {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(col.String())
	}
	f.WriteString(")")

	// Serialize USING after columns (MySQL style) if it was parsed after columns
	if c.Using != nil && c.UsingAfterCols {
		f.WriteString(" USING ")
		f.WriteString(c.Using.String())
	}

	if len(c.Include) > 0 {
		f.WriteString(" INCLUDE (")
		f.WriteString(formatIdents(c.Include, ", "))
		f.WriteString(")")
	}

	if c.NullsDistinct != nil {
		if *c.NullsDistinct {
			f.WriteString(" NULLS DISTINCT")
		} else {
			f.WriteString(" NULLS NOT DISTINCT")
		}
	}

	if len(c.With) > 0 {
		f.WriteString(" WITH (")
		for i, opt := range c.With {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(opt.String())
		}
		f.WriteString(")")
	}

	// Serialize MySQL-specific index options
	for _, opt := range c.MySQLOptions {
		f.WriteString(" ")
		f.WriteString(opt.String())
	}

	if c.TableSpace != nil {
		f.WriteString(" TABLESPACE ")
		f.WriteString(c.TableSpace.String())
	}

	if c.Predicate != nil {
		f.WriteString(" WHERE ")
		f.WriteString(c.Predicate.String())
	}

	return f.String()
}

// ============================================================================
// CREATE FUNCTION
// ============================================================================

// CreateFunction represents a CREATE FUNCTION statement
type CreateFunction struct {
	BaseStatement
	OrReplace                  bool
	Temporary                  bool
	IfNotExists                bool
	Name                       *ast.ObjectName
	Args                       []*expr.OperateFunctionArg
	ReturnType                 *expr.FunctionReturnType
	Language                   *ast.Ident
	Behavior                   *expr.FunctionBehavior
	CalledOnNull               *expr.FunctionCalledOnNull
	Parallel                   *expr.FunctionParallel
	Security                   *expr.FunctionSecurity
	Determinism                *expr.FunctionDeterminismSpecifier
	Cost                       *expr.Expr
	Rows                       *expr.Expr
	Body                       *expr.CreateFunctionBody
	Comment                    *string
	Attributes                 []*expr.SqlOption
	Set                        []*expr.FunctionDefinitionSetParam
	ReturnNullWhenCalledOnNull bool
	ReturnTypeConstraint       *ast.DataType
	SqlSecurity                *expr.SqlSecurity
	RemoteProperty             *expr.RemoteProperty
	Params                     []*expr.ProcedureParam
	External                   bool
	Definer                    expr.Expr
	Aggregate                  bool
	Window                     bool
	Support                    *ast.ObjectName
	LocateIn                   *ast.ObjectName
	ExecuteAs                  *expr.ExecuteAs
}

func (c *CreateFunction) statementNode() {}

func (c *CreateFunction) String() string {
	var f strings.Builder
	f.WriteString("CREATE ")
	if c.OrReplace {
		f.WriteString("OR REPLACE ")
	}
	if c.Temporary {
		f.WriteString("TEMPORARY ")
	}
	f.WriteString("FUNCTION ")
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	f.WriteString(c.Name.String())

	if len(c.Args) > 0 {
		f.WriteString("(")
		for i, arg := range c.Args {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(arg.String())
		}
		f.WriteString(")")
	} else {
		f.WriteString("()")
	}

	if c.ReturnType != nil {
		f.WriteString(" RETURNS ")
		f.WriteString(c.ReturnType.String())
	}

	if c.Language != nil {
		f.WriteString(" LANGUAGE ")
		f.WriteString(c.Language.String())
	}

	if c.Behavior != nil && *c.Behavior != expr.FunctionBehaviorNone {
		f.WriteString(" ")
		f.WriteString(c.Behavior.String())
	}

	if c.CalledOnNull != nil && *c.CalledOnNull != expr.FunctionCalledOnNullNone {
		f.WriteString(" ")
		f.WriteString(c.CalledOnNull.String())
	}

	if c.Parallel != nil && *c.Parallel != expr.FunctionParallelNone {
		f.WriteString(" ")
		f.WriteString(c.Parallel.String())
	}

	if c.Security != nil && *c.Security != expr.FunctionSecurityNone {
		f.WriteString(" ")
		f.WriteString(c.Security.String())
	}

	for _, setParam := range c.Set {
		f.WriteString(" ")
		f.WriteString(setParam.String())
	}

	if c.Body != nil {
		// Only add AS for non-RETURN bodies
		if c.Body.ReturnExpr == nil {
			f.WriteString(" AS ")
		} else {
			f.WriteString(" ")
		}
		f.WriteString(c.Body.String())
	}

	return f.String()
}

// ============================================================================
// CREATE ROLE
// ============================================================================

// CreateRole represents a CREATE ROLE statement
type CreateRole struct {
	BaseStatement
	IfNotExists bool
	Names       []*ast.Ident
	Options     []*expr.RoleOption
}

func (c *CreateRole) statementNode() {}

func (c *CreateRole) String() string {
	var f strings.Builder
	f.WriteString("CREATE ROLE ")
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	f.WriteString(formatIdents(c.Names, ", "))
	for _, opt := range c.Options {
		f.WriteString(" ")
		f.WriteString(opt.String())
	}
	return f.String()
}

// ============================================================================
// ALTER TABLE
// ============================================================================

// AlterTable represents an ALTER TABLE statement
type AlterTable struct {
	BaseStatement
	Name       *ast.ObjectName
	IfExists   bool
	Only       bool
	OnCluster  *ast.Ident // ClickHouse: ON CLUSTER cluster_name
	Operations []*expr.AlterTableOperation
	Location   *expr.HiveSetLocation
	TableType  expr.AlterTableType // Iceberg, Dynamic, External, or None (regular)
}

func (a *AlterTable) statementNode() {}

func (a *AlterTable) String() string {
	var f strings.Builder
	if a.TableType == expr.AlterTableTypeIceberg {
		f.WriteString("ALTER ICEBERG TABLE ")
	} else if a.TableType == expr.AlterTableTypeDynamic {
		f.WriteString("ALTER DYNAMIC TABLE ")
	} else if a.TableType == expr.AlterTableTypeExternal {
		f.WriteString("ALTER EXTERNAL TABLE ")
	} else {
		f.WriteString("ALTER TABLE ")
	}
	if a.IfExists {
		f.WriteString("IF EXISTS ")
	}
	if a.Only {
		f.WriteString("ONLY ")
	}
	f.WriteString(a.Name.String())
	if a.OnCluster != nil {
		f.WriteString(" ON CLUSTER ")
		f.WriteString(a.OnCluster.String())
	}

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
// ALTER SCHEMA
// ============================================================================

// AlterSchema represents an ALTER SCHEMA statement
type AlterSchema struct {
	BaseStatement
	Name      *ast.ObjectName
	IfExists  bool
	Operation *expr.AlterSchemaOperation
}

func (a *AlterSchema) statementNode() {}

func (a *AlterSchema) String() string {
	var f strings.Builder
	f.WriteString("ALTER SCHEMA ")
	if a.IfExists {
		f.WriteString("IF EXISTS ")
	}
	f.WriteString(a.Name.String())
	f.WriteString(" ")
	f.WriteString(a.Operation.String())
	return f.String()
}

// ============================================================================
// ALTER INDEX
// ============================================================================

// AlterIndex represents an ALTER INDEX statement
type AlterIndex struct {
	BaseStatement
	Name      *ast.ObjectName
	Operation *expr.AlterIndexOperation
}

func (a *AlterIndex) statementNode() {}

func (a *AlterIndex) String() string {
	var f strings.Builder
	f.WriteString("ALTER INDEX ")
	f.WriteString(a.Name.String())
	f.WriteString(" ")
	f.WriteString(a.Operation.String())
	return f.String()
}

// ============================================================================
// ALTER VIEW
// ============================================================================

// AlterView represents an ALTER VIEW statement
type AlterView struct {
	BaseStatement
	Name        *ast.ObjectName
	Columns     []*ast.Ident
	Query       *query.Query
	WithOptions []*expr.SqlOption
}

func (a *AlterView) statementNode() {}

func (a *AlterView) String() string {
	var f strings.Builder
	f.WriteString("ALTER VIEW ")
	f.WriteString(a.Name.String())

	if len(a.Columns) > 0 {
		f.WriteString(" (")
		f.WriteString(formatIdents(a.Columns, ", "))
		f.WriteString(")")
	}

	if len(a.WithOptions) > 0 {
		f.WriteString(" WITH (")
		for i, opt := range a.WithOptions {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(opt.String())
		}
		f.WriteString(")")
	}

	f.WriteString(" AS ")
	f.WriteString(a.Query.String())

	return f.String()
}

// ============================================================================
// ALTER TYPE
// ============================================================================

// AlterType represents an ALTER TYPE statement
type AlterType struct {
	BaseStatement
	Name       *ast.ObjectName
	Operations []*expr.AlterTypeOperation
}

func (a *AlterType) statementNode() {}

func (a *AlterType) String() string {
	var f strings.Builder
	f.WriteString("ALTER TYPE ")
	f.WriteString(a.Name.String())

	for i, op := range a.Operations {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(" ")
		f.WriteString(op.String())
	}

	return f.String()
}

// ============================================================================
// ALTER ROLE
// ============================================================================

// AlterRole represents an ALTER ROLE statement
type AlterRole struct {
	BaseStatement
	Name      *ast.Ident
	Operation *expr.AlterRoleOperation
}

func (a *AlterRole) statementNode() {}

func (a *AlterRole) String() string {
	var f strings.Builder
	f.WriteString("ALTER ROLE ")
	f.WriteString(a.Name.String())
	f.WriteString(" ")
	f.WriteString(a.Operation.String())
	return f.String()
}

// ============================================================================
// DROP
// ============================================================================

// Drop represents a DROP statement
type Drop struct {
	BaseStatement
	ObjectType expr.ObjectType
	IfExists   bool
	Names      []*ast.ObjectName
	Cascade    bool
	Restrict   bool
	Purge      bool
	Temporary  bool
	Table      *ast.ObjectName
}

func (d *Drop) statementNode() {}

func (d *Drop) String() string {
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

// ============================================================================
// TRUNCATE
// ============================================================================

// TruncateIdentityOption represents PostgreSQL identity options for TRUNCATE
type TruncateIdentityOption int

const (
	TruncateIdentityNone TruncateIdentityOption = iota
	TruncateIdentityRestart
	TruncateIdentityContinue
)

// TruncateCascadeOption represents PostgreSQL cascade options for TRUNCATE
type TruncateCascadeOption int

const (
	TruncateCascadeNone TruncateCascadeOption = iota
	TruncateCascadeCascade
	TruncateCascadeRestrict
)

// TruncateTableTarget represents a table target in TRUNCATE with optional ONLY and asterisk
type TruncateTableTarget struct {
	Name        *ast.ObjectName
	Only        bool // ONLY keyword (PostgreSQL)
	HasAsterisk bool // * after table name (PostgreSQL)
}

func (t *TruncateTableTarget) String() string {
	var f strings.Builder
	if t.Only {
		f.WriteString("ONLY ")
	}
	f.WriteString(t.Name.String())
	if t.HasAsterisk {
		f.WriteString(" *")
	}
	return f.String()
}

// Truncate represents a TRUNCATE statement
type Truncate struct {
	BaseStatement
	TableNames []*TruncateTableTarget
	Partitions []expr.Expr
	OnCluster  *ast.Ident
	Table      bool                   // Whether TABLE keyword is present
	IfExists   bool                   // Snowflake/Redshift: IF EXISTS option
	Identity   TruncateIdentityOption // PostgreSQL: RESTART IDENTITY | CONTINUE IDENTITY
	Cascade    TruncateCascadeOption  // PostgreSQL: CASCADE | RESTRICT
}

func (t *Truncate) statementNode() {}

func (t *Truncate) String() string {
	var f strings.Builder
	f.WriteString("TRUNCATE ")
	if t.Table {
		f.WriteString("TABLE ")
	}
	if t.IfExists {
		f.WriteString("IF EXISTS ")
	}

	// Format table names
	var tableStrs []string
	for _, tn := range t.TableNames {
		tableStrs = append(tableStrs, tn.String())
	}
	f.WriteString(strings.Join(tableStrs, ", "))

	// PostgreSQL identity option
	switch t.Identity {
	case TruncateIdentityRestart:
		f.WriteString(" RESTART IDENTITY")
	case TruncateIdentityContinue:
		f.WriteString(" CONTINUE IDENTITY")
	}

	// PostgreSQL cascade option
	switch t.Cascade {
	case TruncateCascadeCascade:
		f.WriteString(" CASCADE")
	case TruncateCascadeRestrict:
		f.WriteString(" RESTRICT")
	}

	return f.String()
}

// ============================================================================
// CREATE SCHEMA
// ============================================================================

// CreateSchema represents a CREATE SCHEMA statement
type CreateSchema struct {
	BaseStatement
	SchemaName         *expr.SchemaName
	IfNotExists        bool
	With               *[]*expr.SqlOption // Pointer to distinguish nil (not present) from empty
	Options            *[]*expr.SqlOption // Pointer to distinguish nil (not present) from empty
	DefaultCollateSpec expr.Expr
	Clone              *ast.ObjectName
}

func (c *CreateSchema) statementNode() {}

func (c *CreateSchema) String() string {
	var f strings.Builder
	f.WriteString("CREATE SCHEMA ")
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	f.WriteString(c.SchemaName.String())

	if c.DefaultCollateSpec != nil {
		f.WriteString(" DEFAULT COLLATE ")
		f.WriteString(c.DefaultCollateSpec.String())
	}

	if c.With != nil {
		f.WriteString(" WITH (")
		for i, opt := range *c.With {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(opt.String())
		}
		f.WriteString(")")
	}

	if c.Options != nil {
		f.WriteString(" OPTIONS(")
		for i, opt := range *c.Options {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(opt.String())
		}
		f.WriteString(")")
	}

	if c.Clone != nil {
		f.WriteString(" CLONE ")
		f.WriteString(c.Clone.String())
	}

	return f.String()
}

// ============================================================================
// CREATE DATABASE
// ============================================================================

// CreateDatabase represents a CREATE DATABASE statement
type CreateDatabase struct {
	BaseStatement
	DbName                               *ast.ObjectName
	IfNotExists                          bool
	Location                             *string
	ManagedLocation                      *string
	OrReplace                            bool
	Transient                            bool
	Clone                                *ast.ObjectName
	DataRetentionTimeInDays              *uint64
	MaxDataExtensionTimeInDays           *uint64
	ExternalVolume                       *string
	Catalog                              *string
	ReplaceInvalidCharacters             *bool
	DefaultDdlCollation                  *string
	StorageSerializationPolicy           *expr.StorageSerializationPolicy
	Comment                              *string
	DefaultCharset                       *string
	DefaultCollation                     *string
	CatalogSync                          *string
	CatalogSyncNamespaceMode             *expr.CatalogSyncNamespaceMode
	CatalogSyncNamespaceFlattenDelimiter *string
	WithTags                             []*expr.Tag
	WithContacts                         []*expr.ContactEntry
}

func (c *CreateDatabase) statementNode() {}

func (c *CreateDatabase) String() string {
	var f strings.Builder
	f.WriteString("CREATE ")
	if c.OrReplace {
		f.WriteString("OR REPLACE ")
	}
	if c.Transient {
		f.WriteString("TRANSIENT ")
	}
	f.WriteString("DATABASE ")
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	f.WriteString(c.DbName.String())

	// CLONE clause (Snowflake)
	if c.Clone != nil {
		f.WriteString(" CLONE ")
		f.WriteString(c.Clone.String())
	}

	// Snowflake-specific options
	if c.DataRetentionTimeInDays != nil {
		f.WriteString(" DATA_RETENTION_TIME_IN_DAYS = ")
		f.WriteString(fmt.Sprintf("%d", *c.DataRetentionTimeInDays))
	}
	if c.MaxDataExtensionTimeInDays != nil {
		f.WriteString(" MAX_DATA_EXTENSION_TIME_IN_DAYS = ")
		f.WriteString(fmt.Sprintf("%d", *c.MaxDataExtensionTimeInDays))
	}
	if c.Comment != nil {
		f.WriteString(" COMMENT = '")
		f.WriteString(*c.Comment)
		f.WriteString("'")
	}

	// Output MySQL-style CHARACTER SET and COLLATE options
	// Note: We always output DEFAULT CHARACTER SET and DEFAULT COLLATE as the normalized form
	if c.DefaultCharset != nil {
		f.WriteString(" DEFAULT CHARACTER SET ")
		f.WriteString(*c.DefaultCharset)
	}
	if c.DefaultCollation != nil {
		f.WriteString(" DEFAULT COLLATE ")
		f.WriteString(*c.DefaultCollation)
	}

	return f.String()
}

// ============================================================================
// CREATE SEQUENCE
// ============================================================================

// CreateSequence represents a CREATE SEQUENCE statement
type CreateSequence struct {
	BaseStatement
	Temporary       bool
	IfNotExists     bool
	Name            *ast.ObjectName
	DataType        string // Optional AS data type (e.g., "BIGINT")
	SequenceOptions []*expr.SequenceOptions
	OwnedBy         *ast.ObjectName
}

func (c *CreateSequence) statementNode() {}

func (c *CreateSequence) String() string {
	var f strings.Builder
	f.WriteString("CREATE ")
	if c.Temporary {
		f.WriteString("TEMPORARY ")
	}
	f.WriteString("SEQUENCE ")
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	f.WriteString(c.Name.String())

	// Add AS data_type if present
	if c.DataType != "" {
		f.WriteString(" AS ")
		f.WriteString(c.DataType)
	}

	// Add sequence options
	for _, opt := range c.SequenceOptions {
		f.WriteString(" ")
		f.WriteString(opt.String())
	}

	// Add OWNED BY if present
	if c.OwnedBy != nil {
		f.WriteString(" OWNED BY ")
		f.WriteString(c.OwnedBy.String())
	}

	return f.String()
}

// ============================================================================
// CREATE DOMAIN
// ============================================================================

// CreateDomain represents a CREATE DOMAIN statement
type CreateDomain struct {
	BaseStatement
	Name         *ast.ObjectName
	DataType     interface{} // datatype.DataType
	Collation    *ast.ObjectName
	DefaultValue expr.Expr
	Constraints  []*expr.DomainConstraint
}

func (c *CreateDomain) statementNode() {}

func (c *CreateDomain) String() string {
	var f strings.Builder
	f.WriteString("CREATE DOMAIN ")
	f.WriteString(c.Name.String())
	f.WriteString(" AS ")
	if c.DataType != nil {
		if dt, ok := c.DataType.(fmt.Stringer); ok {
			f.WriteString(dt.String())
		}
	}
	if c.Collation != nil {
		f.WriteString(" COLLATE ")
		f.WriteString(c.Collation.String())
	}
	if c.DefaultValue != nil {
		f.WriteString(" DEFAULT ")
		f.WriteString(c.DefaultValue.String())
	}
	for _, constraint := range c.Constraints {
		f.WriteString(" ")
		f.WriteString(constraint.String())
	}
	return f.String()
}

// ============================================================================
// CREATE TYPE
// ============================================================================

// CreateType represents a CREATE TYPE statement
type CreateType struct {
	BaseStatement
	Name           *ast.ObjectName
	Representation *expr.UserDefinedTypeRepresentation
}

func (c *CreateType) statementNode() {}

func (c *CreateType) String() string {
	var f strings.Builder
	f.WriteString("CREATE TYPE ")
	f.WriteString(c.Name.String())
	if c.Representation != nil {
		f.WriteString(" ")
		f.WriteString((*c.Representation).String())
	}
	return f.String()
}

// ============================================================================
// CREATE VIRTUAL TABLE
// ============================================================================

// CreateVirtualTable represents a CREATE VIRTUAL TABLE statement (SQLite)
type CreateVirtualTable struct {
	BaseStatement
	Name        *ast.ObjectName
	IfNotExists bool
	ModuleName  *ast.Ident
	ModuleArgs  []*ast.Ident
}

func (c *CreateVirtualTable) statementNode() {}

func (c *CreateVirtualTable) String() string {
	var f strings.Builder
	f.WriteString("CREATE VIRTUAL TABLE ")
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	f.WriteString(c.Name.String())
	f.WriteString(" USING ")
	f.WriteString(c.ModuleName.String())

	if len(c.ModuleArgs) > 0 {
		f.WriteString("(")
		f.WriteString(formatIdents(c.ModuleArgs, ", "))
		f.WriteString(")")
	}

	return f.String()
}

// ============================================================================
// CREATE TRIGGER
// ============================================================================

// CreateTrigger represents a CREATE TRIGGER statement
type CreateTrigger struct {
	BaseStatement
	OrAlter             bool
	Temporary           bool
	OrReplace           bool
	IsConstraint        bool
	Name                *ast.ObjectName
	Period              *expr.TriggerPeriod
	PeriodBeforeTable   bool
	Events              []*expr.TriggerEventWithColumns
	TableName           *ast.ObjectName
	ReferencedTableName *ast.ObjectName
	Referencing         []*expr.TriggerReferencing
	TriggerObject       *expr.TriggerObjectKindWithObject
	Condition           expr.Expr
	ExecBody            *expr.TriggerExecBody
	StatementsAs        bool
	Statements          *expr.ConditionalStatements
}

func (c *CreateTrigger) statementNode() {}

func (c *CreateTrigger) String() string {
	var f strings.Builder
	f.WriteString("CREATE ")
	if c.Temporary {
		f.WriteString("TEMPORARY ")
	}
	if c.OrAlter {
		f.WriteString("OR ALTER ")
	}
	if c.OrReplace {
		f.WriteString("OR REPLACE ")
	}
	if c.IsConstraint {
		f.WriteString("CONSTRAINT ")
	}
	f.WriteString("TRIGGER ")
	f.WriteString(c.Name.String())
	f.WriteString(" ")

	if c.PeriodBeforeTable {
		if c.Period != nil {
			f.WriteString(c.Period.String())
			f.WriteString(" ")
		}
		// Write events
		for i, event := range c.Events {
			if i > 0 {
				f.WriteString(" OR ")
			}
			f.WriteString(event.Event.String())
			if len(event.Columns) > 0 {
				f.WriteString(" OF ")
				for j, col := range event.Columns {
					if j > 0 {
						f.WriteString(", ")
					}
					f.WriteString(col.String())
				}
			}
		}
		f.WriteString(" ON ")
		f.WriteString(c.TableName.String())
	} else {
		f.WriteString("ON ")
		f.WriteString(c.TableName.String())
		f.WriteString(" ")
		if c.Period != nil {
			f.WriteString(c.Period.String())
			f.WriteString(" ")
		}
		// Write events
		for i, event := range c.Events {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(event.Event.String())
			if len(event.Columns) > 0 {
				f.WriteString(" OF ")
				for j, col := range event.Columns {
					if j > 0 {
						f.WriteString(", ")
					}
					f.WriteString(col.String())
				}
			}
		}
	}

	if c.ReferencedTableName != nil {
		f.WriteString(" FROM ")
		f.WriteString(c.ReferencedTableName.String())
	}

	if len(c.Referencing) > 0 {
		f.WriteString(" REFERENCING ")
		for i, ref := range c.Referencing {
			if i > 0 {
				f.WriteString(" ")
			}
			f.WriteString(ref.String())
		}
	}

	if c.TriggerObject != nil {
		f.WriteString(" ")
		f.WriteString(c.TriggerObject.String())
	}

	if c.Condition != nil {
		f.WriteString(" WHEN ")
		f.WriteString(c.Condition.String())
	}

	if c.ExecBody != nil {
		f.WriteString(" EXECUTE ")
		f.WriteString(c.ExecBody.String())
	}

	if c.Statements != nil {
		if c.StatementsAs {
			f.WriteString(" AS")
		}
		f.WriteString(" ")
		f.WriteString(c.Statements.String())
	}

	return f.String()
}

// ============================================================================
// DROP TRIGGER
// ============================================================================

// DropTrigger represents a DROP TRIGGER statement
type DropTrigger struct {
	BaseStatement
	IfExists  bool
	Name      *ast.ObjectName
	TableName *ast.ObjectName // Optional ON table_name
}

func (d *DropTrigger) statementNode() {}

func (d *DropTrigger) String() string {
	var f strings.Builder
	f.WriteString("DROP TRIGGER ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	f.WriteString(d.Name.String())
	if d.TableName != nil {
		f.WriteString(" ON ")
		f.WriteString(d.TableName.String())
	}
	return f.String()
}

// ============================================================================
// CREATE PROCEDURE
// ============================================================================

// CreateProcedure represents a CREATE PROCEDURE statement
type CreateProcedure struct {
	BaseStatement
	OrAlter  bool
	Name     *ast.ObjectName
	Params   []*expr.ProcedureParam
	Language *ast.Ident
	Body     *expr.ConditionalStatements
	BodyStr  string // Raw body text for round-trip serialization
}

func (c *CreateProcedure) statementNode() {}

func (c *CreateProcedure) String() string {
	var f strings.Builder
	f.WriteString("CREATE ")
	if c.OrAlter {
		f.WriteString("OR ALTER ")
	}
	f.WriteString("PROCEDURE ")
	f.WriteString(c.Name.String())

	if len(c.Params) > 0 {
		f.WriteString("(")
		for i, param := range c.Params {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(param.String())
		}
		f.WriteString(")")
	}

	if c.Language != nil {
		f.WriteString(" LANGUAGE ")
		f.WriteString(c.Language.String())
	}

	if c.BodyStr != "" {
		f.WriteString(" ")
		f.WriteString(c.BodyStr)
	}

	return f.String()
}

// ============================================================================
// CREATE MACRO (DuckDB)
// ============================================================================

// CreateMacro represents a CREATE MACRO statement (DuckDB)
type CreateMacro struct {
	BaseStatement
	OrReplace  bool
	Temporary  bool
	Name       *ast.ObjectName
	Args       []*expr.MacroArg
	Definition *expr.MacroDefinition
}

func (c *CreateMacro) statementNode() {}

func (c *CreateMacro) String() string {
	var f strings.Builder
	f.WriteString("CREATE ")
	if c.OrReplace {
		f.WriteString("OR REPLACE ")
	}
	if c.Temporary {
		f.WriteString("TEMPORARY ")
	}
	f.WriteString("MACRO ")
	f.WriteString(c.Name.String())

	if len(c.Args) > 0 {
		f.WriteString("(")
		for i, arg := range c.Args {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(arg.String())
		}
		f.WriteString(")")
	}

	if c.Definition != nil {
		f.WriteString(" AS ")
		f.WriteString(c.Definition.String())
	}

	return f.String()
}

// ============================================================================
// CREATE STAGE (Snowflake)
// ============================================================================

// CreateStage represents a CREATE STAGE statement (Snowflake)
type CreateStage struct {
	BaseStatement
	OrReplace            bool
	Temporary            bool
	IfNotExists          bool
	Name                 *ast.ObjectName
	StageParams          *expr.StageParamsObject
	DirectoryTableParams *expr.KeyValueOptions
	FileFormat           *expr.KeyValueOptions
	CopyOptions          *expr.KeyValueOptions
	Comment              *string
}

func (c *CreateStage) statementNode() {}

func (c *CreateStage) String() string {
	var f strings.Builder
	f.WriteString("CREATE ")
	if c.OrReplace {
		f.WriteString("OR REPLACE ")
	}
	if c.Temporary {
		f.WriteString("TEMPORARY ")
	}
	f.WriteString("STAGE ")
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	f.WriteString(c.Name.String())

	// Stage params
	if c.StageParams != nil {
		if str := c.StageParams.String(); str != "" {
			f.WriteString(" ")
			f.WriteString(str)
		}
	}

	// Directory table params
	if c.DirectoryTableParams != nil && len(c.DirectoryTableParams.Options) > 0 {
		f.WriteString(" DIRECTORY=(")
		f.WriteString(c.DirectoryTableParams.String())
		f.WriteString(")")
	}

	// File format
	if c.FileFormat != nil && len(c.FileFormat.Options) > 0 {
		f.WriteString(" FILE_FORMAT=(")
		f.WriteString(c.FileFormat.String())
		f.WriteString(")")
	}

	// Copy options
	if c.CopyOptions != nil && len(c.CopyOptions.Options) > 0 {
		f.WriteString(" COPY_OPTIONS=(")
		f.WriteString(c.CopyOptions.String())
		f.WriteString(")")
	}

	// Comment
	if c.Comment != nil {
		f.WriteString(fmt.Sprintf(" COMMENT='%s'", *c.Comment))
	}

	return f.String()
}

// ============================================================================
// CREATE SECRET (DuckDB)
// ============================================================================

// CreateSecret represents a CREATE SECRET statement (DuckDB)
type CreateSecret struct {
	BaseStatement
	OrReplace        bool
	Temporary        *bool
	IfNotExists      bool
	Name             *ast.Ident
	StorageSpecifier *ast.Ident
	SecretType       *ast.Ident
	Options          []*expr.SecretOption
}

func (c *CreateSecret) statementNode() {}

func (c *CreateSecret) String() string {
	var f strings.Builder
	f.WriteString("CREATE ")
	if c.OrReplace {
		f.WriteString("OR REPLACE ")
	}
	if c.Temporary != nil && *c.Temporary {
		f.WriteString("TEMPORARY ")
	}
	f.WriteString("SECRET ")
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	if c.Name != nil {
		f.WriteString(c.Name.String())
	}
	return f.String()
}

// ============================================================================
// CREATE SERVER
// ============================================================================

// CreateServerStatement represents a CREATE SERVER statement
type CreateServerStatement struct {
	BaseStatement
	OrReplace          bool
	IfNotExists        bool
	Name               *ast.ObjectName
	ServerType         *string
	Version            *string
	ForeignDataWrapper *ast.Ident
	Options            []*expr.SqlOption
}

func (c *CreateServerStatement) statementNode() {}

func (c *CreateServerStatement) String() string {
	var f strings.Builder
	f.WriteString("CREATE SERVER ")
	if c.OrReplace {
		f.WriteString("OR REPLACE ")
	}
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	f.WriteString(c.Name.String())
	if c.ServerType != nil {
		f.WriteString(" TYPE '")
		f.WriteString(*c.ServerType)
		f.WriteString("'")
	}
	if c.Version != nil {
		f.WriteString(" VERSION '")
		f.WriteString(*c.Version)
		f.WriteString("'")
	}
	if c.ForeignDataWrapper != nil {
		f.WriteString(" FOREIGN DATA WRAPPER ")
		f.WriteString(c.ForeignDataWrapper.String())
	}
	if len(c.Options) > 0 {
		f.WriteString(" OPTIONS (")
		for i, opt := range c.Options {
			if i > 0 {
				f.WriteString(", ")
			}
			// For CREATE SERVER, format is key 'value' without = sign
			f.WriteString(opt.Name.String())
			f.WriteString(" ")
			valueStr := opt.Value.String()
			// Add quotes for string literal values
			if valExpr, ok := opt.Value.(*expr.ValueExpr); ok {
				if _, isString := valExpr.Value.(string); isString {
					valueStr = fmt.Sprintf("'%s'", valueStr)
				}
			}
			f.WriteString(valueStr)
		}
		f.WriteString(")")
	}
	return f.String()
}

// ============================================================================
// CREATE POLICY
// ============================================================================

// CreatePolicy represents a CREATE POLICY statement
type CreatePolicy struct {
	BaseStatement
	OrReplace  bool
	Name       *ast.Ident
	TableName  *ast.ObjectName
	PolicyType *expr.CreatePolicyType
	Command    *expr.CreatePolicyCommand
	To         []*expr.Owner
	Using      expr.Expr
	WithCheck  expr.Expr
}

func (c *CreatePolicy) statementNode() {}

func (c *CreatePolicy) String() string {
	var f strings.Builder
	f.WriteString("CREATE POLICY ")
	if c.OrReplace {
		f.WriteString("OR REPLACE ")
	}
	f.WriteString(c.Name.String())
	f.WriteString(" ON ")
	f.WriteString(c.TableName.String())
	if c.PolicyType != nil {
		f.WriteString(" AS ")
		f.WriteString(c.PolicyType.String())
	}
	if c.Command != nil {
		f.WriteString(" FOR ")
		f.WriteString(c.Command.String())
	}
	if len(c.To) > 0 {
		f.WriteString(" TO ")
		for i, role := range c.To {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(role.String())
		}
	}
	if c.Using != nil {
		f.WriteString(" USING (")
		f.WriteString(c.Using.String())
		f.WriteString(")")
	}
	if c.WithCheck != nil {
		f.WriteString(" WITH CHECK (")
		f.WriteString(c.WithCheck.String())
		f.WriteString(")")
	}
	return f.String()
}

// ============================================================================
// CREATE CONNECTOR
// ============================================================================

// CreateConnector represents a CREATE CONNECTOR statement (Hive)
type CreateConnector struct {
	BaseStatement
	IfNotExists      bool
	Name             *ast.Ident
	ConnectorType    *string
	URL              *string
	Comment          *string
	WithDCProperties []*expr.SqlOption
}

func (c *CreateConnector) statementNode() {}

func (c *CreateConnector) String() string {
	var f strings.Builder
	f.WriteString("CREATE CONNECTOR ")
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	f.WriteString(c.Name.String())
	if c.ConnectorType != nil {
		f.WriteString(" TYPE '")
		f.WriteString(*c.ConnectorType)
		f.WriteString("'")
	}
	if c.URL != nil {
		f.WriteString(" URL '")
		f.WriteString(*c.URL)
		f.WriteString("'")
	}
	if c.Comment != nil {
		f.WriteString(" COMMENT '")
		f.WriteString(*c.Comment)
		f.WriteString("'")
	}
	if len(c.WithDCProperties) > 0 {
		f.WriteString(" WITH DCPROPERTIES(")
		for i, opt := range c.WithDCProperties {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(opt.Name.String())
			f.WriteString("=")
			f.WriteString(opt.Value.String())
		}
		f.WriteString(")")
	}
	return f.String()
}

// ============================================================================
// CREATE OPERATOR
// ============================================================================

// CreateOperator represents a CREATE OPERATOR statement
type CreateOperator struct {
	BaseStatement
	Name        *ast.ObjectName
	Function    *ast.ObjectName
	IsProcedure bool
	LeftArg     interface{} // DataType
	RightArg    interface{} // DataType
	Options     []*expr.OperatorOption
}

func (c *CreateOperator) statementNode() {}

func (c *CreateOperator) String() string {
	var f strings.Builder
	f.WriteString("CREATE OPERATOR ")
	f.WriteString(c.Name.String())
	f.WriteString(" (")

	var params []string
	if c.IsProcedure {
		params = append(params, fmt.Sprintf("PROCEDURE = %s", c.Function.String()))
	} else {
		params = append(params, fmt.Sprintf("FUNCTION = %s", c.Function.String()))
	}

	if c.LeftArg != nil {
		params = append(params, fmt.Sprintf("LEFTARG = %v", c.LeftArg))
	}
	if c.RightArg != nil {
		params = append(params, fmt.Sprintf("RIGHTARG = %v", c.RightArg))
	}

	for _, opt := range c.Options {
		params = append(params, opt.String())
	}

	f.WriteString(strings.Join(params, ", "))
	f.WriteString(")")
	return f.String()
}

// ============================================================================
// CREATE OPERATOR FAMILY
// ============================================================================

// CreateOperatorFamily represents a CREATE OPERATOR FAMILY statement
type CreateOperatorFamily struct {
	BaseStatement
	OrReplace   bool
	Name        *ast.ObjectName
	IndexMethod *ast.Ident
	Options     []*expr.SqlOption
}

func (c *CreateOperatorFamily) statementNode() {}

func (c *CreateOperatorFamily) String() string {
	var f strings.Builder
	f.WriteString("CREATE OPERATOR FAMILY ")
	if c.OrReplace {
		f.WriteString("OR REPLACE ")
	}
	f.WriteString(c.Name.String())
	f.WriteString(" USING ")
	f.WriteString(c.IndexMethod.String())
	return f.String()
}

// ============================================================================
// CREATE OPERATOR CLASS
// ============================================================================

// CreateOperatorClass represents a CREATE OPERATOR CLASS statement
type CreateOperatorClass struct {
	BaseStatement
	IsDefault   bool
	Name        *ast.ObjectName
	IndexMethod *ast.Ident
	DataType    interface{} // datatype.DataType
	Family      *ast.ObjectName
	Items       []*expr.OperatorClassItem
}

func (c *CreateOperatorClass) statementNode() {}

func (c *CreateOperatorClass) String() string {
	var f strings.Builder
	f.WriteString("CREATE OPERATOR CLASS ")
	f.WriteString(c.Name.String())
	if c.IsDefault {
		f.WriteString(" DEFAULT")
	}
	f.WriteString(" FOR TYPE ")
	if c.DataType != nil {
		if dt, ok := c.DataType.(fmt.Stringer); ok {
			f.WriteString(dt.String())
		}
	}
	f.WriteString(" USING ")
	f.WriteString(c.IndexMethod.String())
	if c.Family != nil {
		f.WriteString(" FAMILY ")
		f.WriteString(c.Family.String())
	}
	if len(c.Items) > 0 {
		f.WriteString(" AS ")
		for i, item := range c.Items {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(item.String())
		}
	}
	return f.String()
}

// ============================================================================
// ALTER POLICY
// ============================================================================

// AlterPolicy represents an ALTER POLICY statement
type AlterPolicy struct {
	BaseStatement
	Name      *ast.Ident
	TableName *ast.ObjectName
	Operation *expr.AlterPolicyOperation
}

func (a *AlterPolicy) statementNode() {}

func (a *AlterPolicy) String() string {
	var f strings.Builder
	f.WriteString("ALTER POLICY ")
	f.WriteString(a.Name.String())
	f.WriteString(" ON ")
	f.WriteString(a.TableName.String())
	if a.Operation != nil {
		opStr := a.Operation.String()
		if opStr != "" {
			f.WriteString(" ")
			f.WriteString(opStr)
		}
	}
	return f.String()
}

// ============================================================================
// ALTER CONNECTOR
// ============================================================================

// AlterConnector represents an ALTER CONNECTOR statement (Hive)
type AlterConnector struct {
	BaseStatement
	Name       *ast.Ident
	Properties []*expr.SqlOption
	URL        *string
	Owner      *expr.AlterConnectorOwner
}

func (a *AlterConnector) statementNode() {}

func (a *AlterConnector) String() string {
	var f strings.Builder
	f.WriteString("ALTER CONNECTOR ")
	f.WriteString(a.Name.String())
	if len(a.Properties) > 0 {
		f.WriteString(" SET DCPROPERTIES(")
		for i, prop := range a.Properties {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(prop.String())
		}
		f.WriteString(")")
	}
	if a.URL != nil {
		f.WriteString(" SET URL '")
		f.WriteString(*a.URL)
		f.WriteString("'")
	}
	if a.Owner != nil {
		f.WriteString(" SET ")
		f.WriteString(a.Owner.String())
	}
	return f.String()
}

// ============================================================================
// ALTER SESSION
// ============================================================================

// AlterSession represents an ALTER SESSION statement
type AlterSession struct {
	BaseStatement
	Set           bool
	SessionParams *expr.KeyValueOptions
}

func (a *AlterSession) statementNode() {}

func (a *AlterSession) String() string {
	var f strings.Builder
	f.WriteString("ALTER SESSION ")
	if a.Set {
		f.WriteString("SET ")
	} else {
		f.WriteString("UNSET ")
	}
	if a.SessionParams != nil {
		f.WriteString(a.SessionParams.String())
	}
	return f.String()
}

// ============================================================================
// ATTACH DATABASE
// ============================================================================

// AttachDatabase represents an ATTACH DATABASE statement (SQLite)
type AttachDatabase struct {
	BaseStatement
	SchemaName       *ast.Ident
	DatabaseFileName expr.Expr
	Database         bool
}

func (a *AttachDatabase) statementNode() {}

func (a *AttachDatabase) String() string {
	var f strings.Builder
	f.WriteString("ATTACH ")
	if a.Database {
		f.WriteString("DATABASE ")
	}
	f.WriteString(a.DatabaseFileName.String())
	f.WriteString(" AS ")
	f.WriteString(a.SchemaName.String())
	return f.String()
}

// ============================================================================
// ATTACH DUCKDB DATABASE
// ============================================================================

// AttachDuckDBDatabase represents an ATTACH DATABASE statement (DuckDB)
type AttachDuckDBDatabase struct {
	BaseStatement
	IfNotExists   bool
	Database      bool
	DatabasePath  *ast.Ident
	DatabaseAlias *ast.Ident
	AttachOptions []*expr.AttachDuckDBDatabaseOption
}

func (a *AttachDuckDBDatabase) statementNode() {}

func (a *AttachDuckDBDatabase) String() string {
	var f strings.Builder
	f.WriteString("ATTACH ")
	if a.Database {
		f.WriteString("DATABASE ")
	}
	if a.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	f.WriteString(a.DatabasePath.String())
	return f.String()
}

// ============================================================================
// DETACH DUCKDB DATABASE
// ============================================================================

// DetachDuckDBDatabase represents a DETACH DATABASE statement (DuckDB)
type DetachDuckDBDatabase struct {
	BaseStatement
	IfExists      bool
	Database      bool
	DatabaseAlias *ast.Ident
}

func (d *DetachDuckDBDatabase) statementNode() {}

func (d *DetachDuckDBDatabase) String() string {
	var f strings.Builder
	f.WriteString("DETACH ")
	if d.Database {
		f.WriteString("DATABASE ")
	}
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	f.WriteString(d.DatabaseAlias.String())
	return f.String()
}

// ============================================================================
// DROP FUNCTION
// ============================================================================

// DropFunction represents a DROP FUNCTION statement
type DropFunction struct {
	BaseStatement
	IfExists     bool
	FuncDesc     []*expr.FunctionDesc
	DropBehavior *expr.DropBehavior
}

func (d *DropFunction) statementNode() {}

func (d *DropFunction) String() string {
	var f strings.Builder
	f.WriteString("DROP FUNCTION ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	for i, desc := range d.FuncDesc {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(desc.String())
	}
	if d.DropBehavior != nil {
		f.WriteString(" ")
		f.WriteString(d.DropBehavior.String())
	}
	return f.String()
}

// ============================================================================
// DROP DOMAIN
// ============================================================================

// DropDomain represents a DROP DOMAIN statement
type DropDomain struct {
	BaseStatement
	IfExists     bool
	Names        []*ast.ObjectName
	DropBehavior *expr.DropBehavior
}

func (d *DropDomain) statementNode() {}

func (d *DropDomain) String() string {
	var f strings.Builder
	f.WriteString("DROP DOMAIN ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	f.WriteString(formatObjectNames(d.Names, ", "))
	return f.String()
}

// ============================================================================
// DROP PROCEDURE
// ============================================================================

// DropProcedure represents a DROP PROCEDURE statement
type DropProcedure struct {
	BaseStatement
	IfExists     bool
	ProcDesc     []*expr.FunctionDesc
	DropBehavior *expr.DropBehavior
}

func (d *DropProcedure) statementNode() {}

func (d *DropProcedure) String() string {
	var f strings.Builder
	f.WriteString("DROP PROCEDURE ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	for i, desc := range d.ProcDesc {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(desc.String())
	}
	return f.String()
}

// ============================================================================
// DROP SECRET
// ============================================================================

// DropSecret represents a DROP SECRET statement
type DropSecret struct {
	BaseStatement
	IfExists         bool
	Temporary        *bool
	Name             *ast.Ident
	StorageSpecifier *ast.Ident
}

func (d *DropSecret) statementNode() {}

func (d *DropSecret) String() string {
	var f strings.Builder
	f.WriteString("DROP ")
	if d.Temporary != nil && *d.Temporary {
		f.WriteString("TEMPORARY ")
	}
	f.WriteString("SECRET ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	f.WriteString(d.Name.String())
	return f.String()
}

// ============================================================================
// DROP POLICY
// ============================================================================

// DropPolicy represents a DROP POLICY statement
type DropPolicy struct {
	BaseStatement
	IfExists     bool
	Name         *ast.Ident
	TableName    *ast.ObjectName
	DropBehavior *expr.DropBehavior
}

func (d *DropPolicy) statementNode() {}

func (d *DropPolicy) String() string {
	var f strings.Builder
	f.WriteString("DROP POLICY ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	f.WriteString(d.Name.String())
	f.WriteString(" ON ")
	f.WriteString(d.TableName.String())
	if d.DropBehavior != nil {
		f.WriteString(" ")
		f.WriteString(d.DropBehavior.String())
	}
	return f.String()
}

// ============================================================================
// DROP CONNECTOR
// ============================================================================

// DropConnector represents a DROP CONNECTOR statement
type DropConnector struct {
	BaseStatement
	IfExists bool
	Name     *ast.Ident
}

func (d *DropConnector) statementNode() {}

func (d *DropConnector) String() string {
	var f strings.Builder
	f.WriteString("DROP CONNECTOR ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	f.WriteString(d.Name.String())
	return f.String()
}

// ============================================================================
// DROP EXTENSION
// ============================================================================

// DropExtension represents a DROP EXTENSION statement
type DropExtension struct {
	BaseStatement
	IfExists     bool
	Names        []*ast.Ident
	DropBehavior *expr.DropBehavior
}

func (d *DropExtension) statementNode() {}

func (d *DropExtension) String() string {
	var f strings.Builder
	f.WriteString("DROP EXTENSION ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	f.WriteString(formatIdents(d.Names, ", "))
	if d.DropBehavior != nil {
		f.WriteString(" ")
		f.WriteString(d.DropBehavior.String())
	}
	return f.String()
}

// ============================================================================
// DROP OPERATOR
// ============================================================================

// DropOperator represents a DROP OPERATOR statement
type DropOperator struct {
	BaseStatement
	IfExists     bool
	Names        []*expr.DropOperatorSignature
	DropBehavior *expr.DropBehavior
}

func (d *DropOperator) statementNode() {}

func (d *DropOperator) String() string {
	var f strings.Builder
	f.WriteString("DROP OPERATOR ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	for i, sig := range d.Names {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(sig.String())
	}
	if d.DropBehavior != nil {
		f.WriteString(" ")
		f.WriteString(d.DropBehavior.String())
	}
	return f.String()
}

// ============================================================================
// DROP OPERATOR FAMILY
// ============================================================================

// DropOperatorFamily represents a DROP OPERATOR FAMILY statement
type DropOperatorFamily struct {
	BaseStatement
	IfExists     bool
	Names        []*ast.ObjectName
	IndexMethod  *ast.Ident
	DropBehavior *expr.DropBehavior
}

func (d *DropOperatorFamily) statementNode() {}

func (d *DropOperatorFamily) String() string {
	var f strings.Builder
	f.WriteString("DROP OPERATOR FAMILY ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	for i, name := range d.Names {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(name.String())
	}
	f.WriteString(" USING ")
	f.WriteString(d.IndexMethod.String())
	if d.DropBehavior != nil {
		f.WriteString(" ")
		f.WriteString(d.DropBehavior.String())
	}
	return f.String()
}

// ============================================================================
// DROP OPERATOR CLASS
// ============================================================================

// DropOperatorClass represents a DROP OPERATOR CLASS statement
type DropOperatorClass struct {
	BaseStatement
	IfExists     bool
	Names        []*ast.ObjectName
	IndexMethod  *ast.Ident
	DropBehavior *expr.DropBehavior
}

func (d *DropOperatorClass) statementNode() {}

func (d *DropOperatorClass) String() string {
	var f strings.Builder
	f.WriteString("DROP OPERATOR CLASS ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	for i, name := range d.Names {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(name.String())
	}
	f.WriteString(" USING ")
	f.WriteString(d.IndexMethod.String())
	if d.DropBehavior != nil {
		f.WriteString(" ")
		f.WriteString(d.DropBehavior.String())
	}
	return f.String()
}

// ============================================================================
// DROP STAGE (Snowflake)
// ============================================================================

// DropStage represents a DROP STAGE statement (Snowflake-specific)
type DropStage struct {
	BaseStatement
	IfExists bool
	Name     *ast.ObjectName
}

func (d *DropStage) statementNode() {}

func (d *DropStage) String() string {
	var f strings.Builder
	f.WriteString("DROP STAGE ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	f.WriteString(d.Name.String())
	return f.String()
}

// ============================================================================
// CREATE EXTENSION
// ============================================================================

// CreateExtension represents a CREATE EXTENSION statement
type CreateExtension struct {
	BaseStatement
	OrReplace   bool
	IfNotExists bool
	Name        *ast.Ident
	Schema      *ast.ObjectName
	Version     *string
	Cascade     bool
}

func (c *CreateExtension) statementNode() {}

func (c *CreateExtension) String() string {
	var f strings.Builder
	f.WriteString("CREATE EXTENSION ")
	if c.OrReplace {
		f.WriteString("OR REPLACE ")
	}
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	f.WriteString(c.Name.String())

	// Add WITH clause if any option is present
	if c.Cascade || c.Schema != nil || c.Version != nil {
		f.WriteString(" WITH")
		if c.Schema != nil {
			f.WriteString(" SCHEMA ")
			f.WriteString(c.Schema.String())
		}
		if c.Version != nil {
			f.WriteString(" VERSION ")
			f.WriteString(*c.Version)
		}
		if c.Cascade {
			f.WriteString(" CASCADE")
		}
	}

	return f.String()
}

// ============================================================================
// ALTER OPERATOR
// ============================================================================

// AlterOperator represents an ALTER OPERATOR statement
type AlterOperator struct {
	BaseStatement
	Name      *ast.ObjectName
	Signature *expr.OperatorSignature
	Operation *expr.AlterOperatorOperation
}

func (a *AlterOperator) statementNode() {}

func (a *AlterOperator) String() string {
	var f strings.Builder
	f.WriteString("ALTER OPERATOR ")
	f.WriteString(a.Name.String())
	return f.String()
}

// ============================================================================
// ALTER OPERATOR FAMILY
// ============================================================================

// AlterOperatorFamily represents an ALTER OPERATOR FAMILY statement
type AlterOperatorFamily struct {
	BaseStatement
	Name        *ast.ObjectName
	IndexMethod *ast.Ident
	Operations  []*expr.OperatorFamilyOperation
}

func (a *AlterOperatorFamily) statementNode() {}

func (a *AlterOperatorFamily) String() string {
	var f strings.Builder
	f.WriteString("ALTER OPERATOR FAMILY ")
	f.WriteString(a.Name.String())
	f.WriteString(" USING ")
	f.WriteString(a.IndexMethod.String())
	return f.String()
}

// ============================================================================
// ALTER OPERATOR CLASS
// ============================================================================

// AlterOperatorClass represents an ALTER OPERATOR CLASS statement
type AlterOperatorClass struct {
	BaseStatement
	Name        *ast.ObjectName
	IndexMethod *ast.Ident
	Operations  []*expr.OperatorClassOperation
}

func (a *AlterOperatorClass) statementNode() {}

func (a *AlterOperatorClass) String() string {
	var f strings.Builder
	f.WriteString("ALTER OPERATOR CLASS ")
	f.WriteString(a.Name.String())
	f.WriteString(" USING ")
	f.WriteString(a.IndexMethod.String())
	return f.String()
}

// AlterUserOperation represents ALTER USER operation.
type AlterUserOperation interface {
	alterUserOperation()
	fmt.Stringer
}

// TableConstraint represents table constraint for CREATE TABLE.
type TableConstraint struct {
	BaseStatement
	Name       *ast.Ident
	Constraint interface{} // TODO: Add specific constraint types
}

func (t *TableConstraint) statementNode() {}

func (t *TableConstraint) String() string {
	return "CONSTRAINT"
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

// AtBeforeClause represents AT/BEFORE clause for time travel queries.
type AtBeforeClause struct {
	Timestamp *string
	Offset    *int64
	Statement *string
}

func (a *AtBeforeClause) statementNode() {}

func (a *AtBeforeClause) String() string {
	return "AT/BEFORE"
}

// SetAssignment represents SET assignment.
type SetAssignment struct {
	Variable *ast.Ident
	Value    expr.Expr
}

func (s *SetAssignment) statementNode() {}

func (s *SetAssignment) String() string {
	return fmt.Sprintf("%s = %s", s.Variable.String(), s.Value.String())
}

// CopyOption represents COPY option.
type CopyOption struct {
	Name  string
	Value expr.Expr
}

func (c *CopyOption) statementNode() {}

func (c *CopyOption) String() string {
	return fmt.Sprintf("%s = %s", c.Name, c.Value.String())
}

// ============================================================================
// DROP Statements
// ============================================================================

// DropTable represents a DROP TABLE statement
type DropTable struct {
	BaseStatement
	Temporary bool
	IfExists  bool
	Names     []*ast.ObjectName
	Cascade   bool
	Restrict  bool
	Purge     bool
}

func (d *DropTable) statementNode() {}

func (d *DropTable) String() string {
	var f strings.Builder
	f.WriteString("DROP ")
	if d.Temporary {
		f.WriteString("TEMPORARY ")
	}
	f.WriteString("TABLE ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	for i, name := range d.Names {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(name.String())
	}
	if d.Cascade {
		f.WriteString(" CASCADE")
	}
	if d.Restrict {
		f.WriteString(" RESTRICT")
	}
	if d.Purge {
		f.WriteString(" PURGE")
	}
	return f.String()
}

// DropView represents a DROP VIEW statement
type DropView struct {
	BaseStatement
	IfExists bool
	Names    []*ast.ObjectName
	Cascade  bool
	Restrict bool
}

func (d *DropView) statementNode() {}

func (d *DropView) String() string {
	var f strings.Builder
	f.WriteString("DROP VIEW ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	for i, name := range d.Names {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(name.String())
	}
	if d.Cascade {
		f.WriteString(" CASCADE")
	}
	if d.Restrict {
		f.WriteString(" RESTRICT")
	}
	return f.String()
}

// DropIndex represents a DROP INDEX statement
type DropIndex struct {
	BaseStatement
	IfExists     bool
	Names        []*ast.ObjectName
	Name         *ast.ObjectName // For single index with ON table syntax
	OnTable      *ast.ObjectName // Optional table for DROP INDEX name ON table syntax
	Cascade      bool
	Restrict     bool
	Concurrently bool
}

func (d *DropIndex) statementNode() {}

func (d *DropIndex) String() string {
	var f strings.Builder
	f.WriteString("DROP INDEX ")
	if d.Concurrently {
		f.WriteString("CONCURRENTLY ")
	}
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	for i, name := range d.Names {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(name.String())
	}
	if d.OnTable != nil {
		f.WriteString(" ON ")
		f.WriteString(d.OnTable.String())
	}
	if d.Cascade {
		f.WriteString(" CASCADE")
	}
	if d.Restrict {
		f.WriteString(" RESTRICT")
	}
	return f.String()
}

// DropRole represents a DROP ROLE statement
type DropRole struct {
	BaseStatement
	IfExists bool
	Names    []*ast.Ident
	Cascade  bool
	Restrict bool
}

func (d *DropRole) statementNode() {}

func (d *DropRole) String() string {
	var f strings.Builder
	f.WriteString("DROP ROLE ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	for i, name := range d.Names {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(name.String())
	}
	return f.String()
}

// DropDatabase represents a DROP DATABASE statement
type DropDatabase struct {
	BaseStatement
	IfExists bool
	Names    []*ast.ObjectName
	Cascade  bool
	Restrict bool
}

func (d *DropDatabase) statementNode() {}

func (d *DropDatabase) String() string {
	var f strings.Builder
	f.WriteString("DROP DATABASE ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	for i, name := range d.Names {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(name.String())
	}
	return f.String()
}

// DropSchema represents a DROP SCHEMA statement
type DropSchema struct {
	BaseStatement
	IfExists bool
	Names    []*ast.ObjectName
	Cascade  bool
	Restrict bool
}

func (d *DropSchema) statementNode() {}

func (d *DropSchema) String() string {
	var f strings.Builder
	f.WriteString("DROP SCHEMA ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	for i, name := range d.Names {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(name.String())
	}
	if d.Cascade {
		f.WriteString(" CASCADE")
	}
	if d.Restrict {
		f.WriteString(" RESTRICT")
	}
	return f.String()
}

// DropSequence represents a DROP SEQUENCE statement
type DropSequence struct {
	BaseStatement
	IfExists bool
	Names    []*ast.ObjectName
	Cascade  bool
	Restrict bool
}

func (d *DropSequence) statementNode() {}

func (d *DropSequence) String() string {
	var f strings.Builder
	f.WriteString("DROP SEQUENCE ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	for i, name := range d.Names {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(name.String())
	}
	if d.Cascade {
		f.WriteString(" CASCADE")
	}
	if d.Restrict {
		f.WriteString(" RESTRICT")
	}
	return f.String()
}

// DropType represents a DROP TYPE statement
type DropType struct {
	BaseStatement
	IfExists bool
	Names    []*ast.ObjectName
	Cascade  bool
	Restrict bool
}

func (d *DropType) statementNode() {}

func (d *DropType) String() string {
	var f strings.Builder
	f.WriteString("DROP TYPE ")
	if d.IfExists {
		f.WriteString("IF EXISTS ")
	}
	for i, name := range d.Names {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(name.String())
	}
	return f.String()
}
