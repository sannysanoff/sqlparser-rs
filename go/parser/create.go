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
	"reflect"
	"strings"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/query"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/token"
)

// parseCreate parses a CREATE statement
// Reference: src/parser/mod.rs:5095
func parseCreate(p *Parser) (ast.Statement, error) {
	// Check for OR REPLACE / OR ALTER
	orReplace := p.ParseKeywords([]string{"OR", "REPLACE"})
	_ = p.ParseKeywords([]string{"OR", "ALTER"}) // orAlter not used yet

	// Check for LOCAL/GLOBAL
	local := p.ParseKeyword("LOCAL")
	global := p.ParseKeyword("GLOBAL")
	var globalOpt *bool
	if global {
		globalOpt = &[]bool{true}[0]
	} else if local {
		globalOpt = &[]bool{false}[0]
	}

	// Check for TRANSIENT
	transient := p.ParseKeyword("TRANSIENT")

	// Check for TEMPORARY/TEMP
	temporary := p.ParseKeyword("TEMPORARY") || p.ParseKeyword("TEMP")

	// Check for VOLATILE (Snowflake)
	volatile := p.ParseKeyword("VOLATILE")

	// Check for PERSISTENT (DuckDB)
	// Note: Persistent is not stored separately, it's just a modifier
	p.ParseKeyword("PERSISTENT")

	// Check for DYNAMIC (Snowflake)
	dynamic := p.ParseKeyword("DYNAMIC")

	// Check for ICEBERG (Snowflake) - can appear after DYNAMIC
	iceberg := p.ParseKeyword("ICEBERG")

	// Try various CREATE targets
	switch {
	case p.PeekKeyword("TABLE"):
		return parseCreateTable(p, orReplace, temporary, globalOpt, transient, volatile, iceberg, dynamic)
	case p.PeekKeyword("VIEW"):
		return parseCreateView(p, orReplace, temporary, false)
	case p.PeekKeyword("ALGORITHM"), p.PeekKeyword("DEFINER"), p.PeekKeyword("SQL"):
		// MySQL CREATE VIEW parameters before VIEW keyword
		return parseCreateView(p, orReplace, temporary, false)
	case p.PeekKeyword("INDEX"):
		return parseCreateIndex(p, false)
	case p.PeekKeyword("UNIQUE"):
		p.NextToken()
		return parseCreateIndex(p, true)
	case p.PeekKeyword("ROLE"):
		return parseCreateRole(p, orReplace)
	case p.PeekKeyword("DATABASE"):
		return parseCreateDatabase(p)
	case p.PeekKeyword("SCHEMA"):
		return parseCreateSchema(p)
	case p.PeekKeyword("SEQUENCE"):
		return parseCreateSequence(p, orReplace, temporary)
	case p.PeekKeyword("TYPE"):
		return parseCreateType(p)
	case p.PeekKeyword("DOMAIN"):
		return parseCreateDomain(p)
	case p.PeekKeyword("EXTENSION"):
		p.NextToken() // Consume EXTENSION keyword
		return parseCreateExtension(p)
	case p.PeekKeyword("TRIGGER"):
		return parseCreateTrigger(p, orReplace)
	case p.PeekKeyword("POLICY"):
		p.NextToken() // Consume POLICY keyword
		return parseCreatePolicy(p, orReplace)
	case p.PeekKeyword("FUNCTION"):
		return parseCreateFunction(p, orReplace, temporary)
	case p.PeekKeyword("VIRTUAL"):
		p.NextToken()
		return parseCreateVirtualTable(p)
	case p.PeekKeyword("MACRO"):
		return parseCreateMacro(p)
	case p.PeekKeyword("SECRET"):
		return parseCreateSecret(p, orReplace, temporary)
	case p.PeekKeyword("CONNECTOR"):
		return parseCreateConnector(p, orReplace)
	case p.PeekKeyword("OPERATOR"):
		return parseCreateOperator(p)
	case p.PeekKeyword("USER"):
		return parseCreateUser(p, orReplace)
	case p.PeekKeyword("PROCEDURE"):
		return parseCreateProcedure(p, false)
	case p.PeekKeyword("STAGE"):
		return parseCreateStage(p, orReplace, temporary, transient)
	case p.PeekKeyword("SECURE"):
		// CREATE SECURE VIEW or CREATE SECURE MATERIALIZED VIEW
		p.NextToken() // consume SECURE
		// Check if MATERIALIZED follows
		isMaterialized := p.ParseKeyword("MATERIALIZED")
		if !p.PeekKeyword("VIEW") {
			return nil, p.ExpectedRef("VIEW after SECURE", p.PeekTokenRef())
		}
		return parseCreateView(p, orReplace, temporary, true, isMaterialized) // secure=true, isMaterialized
	case p.PeekKeyword("MATERIALIZED"):
		// CREATE MATERIALIZED VIEW
		p.NextToken() // consume MATERIALIZED
		if !p.PeekKeyword("VIEW") {
			return nil, p.ExpectedRef("VIEW after MATERIALIZED", p.PeekTokenRef())
		}
		return parseCreateView(p, orReplace, temporary, false, true) // secure=false, isMaterialized=true
	default:
		return nil, p.ExpectedRef("TABLE, VIEW, INDEX, FUNCTION, ROLE, or other CREATE target", p.PeekTokenRef())
	}
}

// parseCreateTable parses CREATE TABLE
// Reference: src/parser/mod.rs:8339
func parseCreateTable(p *Parser, orReplace, temporary bool, global *bool, transient bool, volatile bool, iceberg bool, dynamic bool) (ast.Statement, error) {
	// Consume TABLE keyword
	if _, err := p.ExpectKeyword("TABLE"); err != nil {
		return nil, err
	}

	// Check for IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Snowflake-specific options that can appear before or after columns
	var copyGrants bool
	var enableSchemaEvolution *bool
	var changeTracking *bool
	var dataRetentionTimeInDays *uint64
	var clusterBy *expr.WrappedCollection
	var comment *expr.CommentDef
	var onCommit *expr.OnCommit
	// Additional Snowflake options for ICEBERG and DYNAMIC tables
	var externalVolume *string
	var baseLocation *string
	var catalog *string
	var catalogSync *string
	var storageSerializationPolicy *expr.StorageSerializationPolicy
	var targetLag *string
	var warehouse *ast.Ident
	var refreshMode *expr.RefreshModeKind
	var initialize *expr.InitializeKind
	var requireUser bool

	// For Snowflake, parse options that can appear before the column list
	if p.GetDialect().Dialect() == "snowflake" {
		snowflakeOpts := parseSnowflakeCreateTableOptionsBeforeColumns(p)
		copyGrants = snowflakeOpts.CopyGrants
		enableSchemaEvolution = snowflakeOpts.EnableSchemaEvolution
		changeTracking = snowflakeOpts.ChangeTracking
		dataRetentionTimeInDays = snowflakeOpts.DataRetentionTimeInDays
		clusterBy = snowflakeOpts.ClusterBy
		comment = snowflakeOpts.Comment
		onCommit = snowflakeOpts.OnCommit
		requireUser = snowflakeOpts.RequireUser
		// Additional options
		if snowflakeOpts.ExternalVolume != nil {
			externalVolume = snowflakeOpts.ExternalVolume
		}
		if snowflakeOpts.BaseLocation != nil {
			baseLocation = snowflakeOpts.BaseLocation
		}
		if snowflakeOpts.Catalog != nil {
			catalog = snowflakeOpts.Catalog
		}
		if snowflakeOpts.CatalogSync != nil {
			catalogSync = snowflakeOpts.CatalogSync
		}
		if snowflakeOpts.StorageSerializationPolicy != nil {
			storageSerializationPolicy = snowflakeOpts.StorageSerializationPolicy
		}
		if snowflakeOpts.TargetLag != nil {
			targetLag = snowflakeOpts.TargetLag
		}
		if snowflakeOpts.Warehouse != nil {
			warehouse = snowflakeOpts.Warehouse
		}
		if snowflakeOpts.RefreshMode != nil {
			refreshMode = snowflakeOpts.RefreshMode
		}
		if snowflakeOpts.Initialize != nil {
			initialize = snowflakeOpts.Initialize
		}
	}

	// PostgreSQL PARTITION OF for child partition tables
	// Note: This is a PostgreSQL-specific feature, but the dialect check was intentionally
	// removed to allow GenericDialect and other dialects to parse this syntax.
	var partitionOf *ast.ObjectName
	if p.ParseKeywords([]string{"PARTITION", "OF"}) {
		partitionOf, _ = p.ParseObjectName()
	}

	// ClickHouse ON CLUSTER - must be parsed before columns
	var onCluster *ast.Ident
	if p.ParseKeywords([]string{"ON", "CLUSTER"}) {
		tok := p.NextToken()
		if word, ok := tok.Token.(token.TokenWord); ok {
			onCluster = &ast.Ident{Value: word.Word.Value}
		} else if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			onCluster = &ast.Ident{Value: str.Value}
		} else if str, ok := tok.Token.(token.TokenDoubleQuotedString); ok {
			onCluster = &ast.Ident{Value: str.Value}
		} else {
			return nil, p.Expected("identifier or string", tok)
		}
	}

	// Parse LIKE clause (before column list, parenthesized or plain)
	var like *expr.CreateTableLikeKind
	like, err = parseCreateTableLike(p)
	if err != nil {
		return nil, err
	}

	// Parse CLONE clause (Snowflake)
	var clone *ast.ObjectName
	if p.ParseKeyword("CLONE") {
		clone, err = p.ParseObjectName()
		if err != nil {
			return nil, err
		}
	}

	// Parse optional column list and constraints
	var columns []*expr.ColumnDef
	var constraints []*expr.TableConstraint

	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen && like == nil {
		cols, cons, err := parseCreateTableColumns(p)
		if err != nil {
			return nil, err
		}
		columns = cols
		constraints = cons
	}

	// PostgreSQL PARTITION OF: parse FOR VALUES clause if PARTITION OF was specified
	var forValues *expr.ForValues
	if partitionOf != nil {
		forValues = parsePartitionForValues(p)
		if forValues == nil && !p.PeekKeyword("FOR") && !p.PeekKeyword("DEFAULT") {
			// FOR VALUES or DEFAULT is required after PARTITION OF
			return nil, fmt.Errorf("Expected: FOR VALUES or DEFAULT after PARTITION OF, found: %s", p.PeekToken().Token.String())
		}
	}

	// Parse optional MySQL table options (ENGINE, CHARSET, COLLATE, COMMENT, etc.)
	var tableOptions *expr.CreateTableOptions
	if len(columns) > 0 || len(constraints) > 0 {
		tableOptions, err = parseCreateTableOptions(p)
		if err != nil {
			return nil, err
		}
	}

	// Parse AS (CREATE TABLE ... AS SELECT)
	var asQuery *query.Query
	if p.PeekKeyword("AS") {
		p.AdvanceToken()
		innerQuery, err := p.ParseQuery()
		if err != nil {
			return nil, err
		}
		asQuery = extractQueryFromStatement(innerQuery)
	}

	// Parse CREATE TABLE ... SELECT (MySQL style without AS)
	if asQuery == nil && p.GetDialect().SupportsCreateTableSelect() && p.PeekKeyword("SELECT") {
		innerQuery, err := p.ParseQuery()
		if err != nil {
			return nil, err
		}
		asQuery = extractQueryFromStatement(innerQuery)
	}

	// SQLite WITHOUT ROWID
	withoutRowid := p.ParseKeywords([]string{"WITHOUT", "ROWID"})

	// Hive distribution - using existing HiveDistributionStyle type
	hiveDistribution := parseHiveDistributionStyle(p)

	// CLUSTERED BY (Hive) - using existing ClusteredBy type
	clusteredBy := parseClusteredByClause(p)

	// Hive formats - using existing HiveFormat type
	hiveFormats := parseHiveFormatClause(p)

	// ClickHouse PRIMARY KEY
	var primaryKey expr.Expr
	if p.GetDialect().Dialect() == "clickhouse" || p.GetDialect().Dialect() == "generic" {
		if p.ParseKeywords([]string{"PRIMARY", "KEY"}) {
			ep := NewExpressionParser(p)
			primaryKey, _ = ep.ParseExpr()
		}
	}

	// ORDER BY (ClickHouse) - using existing OneOrManyWithParens type
	var orderBy *expr.OneOrManyWithParens
	if p.ParseKeywords([]string{"ORDER", "BY"}) {
		orderBy = parseOrderByClauseForCreateTable(p)
	}

	// ON COMMIT (PostgreSQL) - using existing OnCommit type
	// Note: For Snowflake, this is already parsed above
	if onCommit == nil && p.ParseKeywords([]string{"ON", "COMMIT"}) {
		onCommit = parseOnCommitClause(p)
	}

	// STRICT (SQLite 3.37+)
	strict := p.ParseKeyword("STRICT")

	// Redshift BACKUP
	var backup *bool
	if p.ParseKeyword("BACKUP") {
		if p.PeekKeyword("YES") {
			p.AdvanceToken()
			backupVal := true
			backup = &backupVal
		} else if p.PeekKeyword("NO") {
			p.AdvanceToken()
			backupVal := false
			backup = &backupVal
		}
	}

	// Redshift DISTSTYLE
	var diststyle *expr.DistStyle
	if p.ParseKeyword("DISTSTYLE") {
		diststyle = parseDistStyle(p)
	}

	// Redshift DISTKEY
	var distkey expr.Expr
	if p.ParseKeyword("DISTKEY") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err == nil {
			ep := NewExpressionParser(p)
			distkey, _ = ep.ParseExpr()
			p.ExpectToken(token.TokenRParen{})
		}
	}

	// Redshift SORTKEY
	var sortkey []expr.Expr
	if p.ParseKeyword("SORTKEY") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err == nil {
			ep := NewExpressionParser(p)
			for {
				expr, _ := ep.ParseExpr()
				if expr == nil {
					break
				}
				sortkey = append(sortkey, expr)
				if !p.ConsumeToken(token.TokenComma{}) {
					break
				}
			}
			p.ExpectToken(token.TokenRParen{})
		}
	}

	// For Snowflake, parse additional options that can appear after columns
	// Snowflake allows options in any order, so we use a loop
	if p.GetDialect().Dialect() == "snowflake" {
		snowflakeOpts := parseSnowflakeCreateTableOptions(p)
		// Merge Snowflake options with existing values (later values override earlier ones)
		if snowflakeOpts.CopyGrants {
			copyGrants = true
		}
		if snowflakeOpts.EnableSchemaEvolution != nil {
			enableSchemaEvolution = snowflakeOpts.EnableSchemaEvolution
		}
		if snowflakeOpts.ChangeTracking != nil {
			changeTracking = snowflakeOpts.ChangeTracking
		}
		if snowflakeOpts.DataRetentionTimeInDays != nil {
			dataRetentionTimeInDays = snowflakeOpts.DataRetentionTimeInDays
		}
		if snowflakeOpts.ClusterBy != nil {
			clusterBy = snowflakeOpts.ClusterBy
		}
		if snowflakeOpts.Comment != nil {
			comment = snowflakeOpts.Comment
		}
		if snowflakeOpts.OnCommit != nil {
			onCommit = snowflakeOpts.OnCommit
		}
		// Additional options
		if snowflakeOpts.ExternalVolume != nil {
			externalVolume = snowflakeOpts.ExternalVolume
		}
		if snowflakeOpts.BaseLocation != nil {
			baseLocation = snowflakeOpts.BaseLocation
		}
		if snowflakeOpts.Catalog != nil {
			catalog = snowflakeOpts.Catalog
		}
		if snowflakeOpts.CatalogSync != nil {
			catalogSync = snowflakeOpts.CatalogSync
		}
		if snowflakeOpts.StorageSerializationPolicy != nil {
			storageSerializationPolicy = snowflakeOpts.StorageSerializationPolicy
		}
		if snowflakeOpts.TargetLag != nil {
			targetLag = snowflakeOpts.TargetLag
		}
		if snowflakeOpts.Warehouse != nil {
			warehouse = snowflakeOpts.Warehouse
		}
		if snowflakeOpts.RefreshMode != nil {
			refreshMode = snowflakeOpts.RefreshMode
		}
		if snowflakeOpts.Initialize != nil {
			initialize = snowflakeOpts.Initialize
		}
		if snowflakeOpts.RequireUser {
			requireUser = true
		}
	}

	// For Snowflake DYNAMIC tables, AS query comes after all options
	// Check if we haven't parsed AS query yet and there's one now
	if asQuery == nil && p.PeekKeyword("AS") {
		p.AdvanceToken()
		innerQuery, err := p.ParseQuery()
		if err != nil {
			return nil, err
		}
		asQuery = extractQueryFromStatement(innerQuery)
	}

	return &statement.CreateTable{
		OrReplace:        orReplace,
		Temporary:        temporary,
		Global:           global,
		Transient:        transient,
		Volatile:         volatile,
		Iceberg:          iceberg,
		Dynamic:          dynamic,
		IfNotExists:      ifNotExists,
		Name:             tableName,
		Columns:          columns,
		Constraints:      constraints,
		TableOptions:     tableOptions,
		Query:            asQuery,
		Like:             like,
		Clone:            clone,
		OnCluster:        onCluster,
		PartitionOf:      partitionOf,
		ForValues:        forValues,
		WithoutRowid:     withoutRowid,
		HiveDistribution: hiveDistribution,
		HiveFormats:      hiveFormats,
		ClusteredBy:      clusteredBy,
		PrimaryKey:       primaryKey,
		OrderBy:          orderBy,
		OnCommit:         onCommit,
		Strict:           strict,
		Backup:           backup,
		Diststyle:        diststyle,
		Distkey:          distkey,
		Sortkey:          sortkey,
		// Snowflake-specific fields
		CopyGrants:              copyGrants,
		EnableSchemaEvolution:   enableSchemaEvolution,
		ChangeTracking:          changeTracking,
		DataRetentionTimeInDays: dataRetentionTimeInDays,
		ClusterBy:               clusterBy,
		Comment:                 comment,
		// Additional Snowflake fields
		ExternalVolume:             externalVolume,
		BaseLocation:               baseLocation,
		Catalog:                    catalog,
		CatalogSync:                catalogSync,
		StorageSerializationPolicy: storageSerializationPolicy,
		TargetLag:                  targetLag,
		Warehouse:                  warehouse,
		RefreshMode:                refreshMode,
		Initialize:                 initialize,
		RequireUser:                requireUser,
	}, nil
}

// snowflakeCreateTableOptions holds Snowflake-specific CREATE TABLE options
type snowflakeCreateTableOptions struct {
	CopyGrants              bool
	EnableSchemaEvolution   *bool
	ChangeTracking          *bool
	DataRetentionTimeInDays *uint64
	ClusterBy               *expr.WrappedCollection
	Comment                 *expr.CommentDef
	OnCommit                *expr.OnCommit
	// Additional Snowflake options for ICEBERG and DYNAMIC tables
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
}

// parseSnowflakeCreateTableOptionsBeforeColumns parses Snowflake options that appear before columns
// Snowflake allows these options in any order, before or after the column list
func parseSnowflakeCreateTableOptionsBeforeColumns(p *Parser) snowflakeCreateTableOptions {
	return parseSnowflakeCreateTableOptionsWithTerminator(p, token.TokenLParen{})
}

// parseSnowflakeCreateTableOptions parses Snowflake options that appear after columns
func parseSnowflakeCreateTableOptions(p *Parser) snowflakeCreateTableOptions {
	return parseSnowflakeCreateTableOptionsWithTerminator(p, nil)
}

// parseSnowflakeCreateTableOptionsWithTerminator parses Snowflake CREATE TABLE options in a loop
// until a terminator token is encountered or no more options are available
func parseSnowflakeCreateTableOptionsWithTerminator(p *Parser, terminator token.Token) snowflakeCreateTableOptions {
	var opts snowflakeCreateTableOptions

	for {
		// Check for terminator - if the next token matches the terminator type, stop
		if terminator != nil {
			switch terminator.(type) {
			case token.TokenLParen:
				if _, isTerminator := p.PeekToken().Token.(token.TokenLParen); isTerminator {
					break
				}
			}
		}

		tok := p.PeekToken()
		word, ok := tok.Token.(token.TokenWord)
		if !ok {
			break
		}

		kw := string(word.Word.Keyword)
		if kw == "" {
			break
		}

		switch kw {
		case "COPY":
			p.AdvanceToken()
			if p.ParseKeyword("GRANTS") {
				opts.CopyGrants = true
			} else {
				// Not COPY GRANTS, backtrack
				p.PrevToken()
				return opts
			}

		case "COMMENT":
			p.AdvanceToken()
			commentTok := p.NextToken()
			if str, ok := commentTok.Token.(token.TokenSingleQuotedString); ok {
				opts.Comment = &expr.CommentDef{Comment: str.Value}
			} else {
				// Backtrack
				p.PrevToken()
				p.PrevToken()
				return opts
			}

		case "CLUSTER":
			p.AdvanceToken()
			if p.ParseKeyword("BY") {
				if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
					p.AdvanceToken() // consume (
					ep := NewExpressionParser(p)
					var clusterExprs []expr.Expr
					for {
						expr, _ := ep.ParseExpr()
						if expr == nil {
							break
						}
						clusterExprs = append(clusterExprs, expr)
						if !p.ConsumeToken(token.TokenComma{}) {
							break
						}
					}
					if _, ok := p.PeekToken().Token.(token.TokenRParen); ok {
						p.AdvanceToken() // consume )
					}
					wrapped := &expr.WrappedCollection{Items: clusterExprs}
					opts.ClusterBy = wrapped
				}
			} else {
				// Not CLUSTER BY, backtrack
				p.PrevToken()
				return opts
			}

		case "ENABLE_SCHEMA_EVOLUTION":
			p.AdvanceToken()
			if p.ConsumeToken(token.TokenEq{}) {
				val := p.ParseKeyword("TRUE")
				if !val {
					val = p.ParseKeyword("FALSE")
				}
				opts.EnableSchemaEvolution = &val
			}

		case "CHANGE_TRACKING":
			p.AdvanceToken()
			if p.ConsumeToken(token.TokenEq{}) {
				val := p.ParseKeyword("TRUE")
				if !val {
					val = p.ParseKeyword("FALSE")
				}
				opts.ChangeTracking = &val
			}

		case "DATA_RETENTION_TIME_IN_DAYS":
			p.AdvanceToken()
			if p.ConsumeToken(token.TokenEq{}) {
				valTok := p.NextToken()
				if num, ok := valTok.Token.(token.TokenNumber); ok {
					var val uint64
					fmt.Sscanf(num.Value, "%d", &val)
					opts.DataRetentionTimeInDays = &val
				}
			}

		case "ON":
			p.AdvanceToken()
			if p.ParseKeyword("COMMIT") {
				opts.OnCommit = parseOnCommitClause(p)
			} else {
				// Not ON COMMIT, backtrack
				p.PrevToken()
				return opts
			}

		// ICEBERG table options
		case "EXTERNAL_VOLUME":
			p.AdvanceToken()
			if p.ConsumeToken(token.TokenEq{}) {
				valTok := p.NextToken()
				if str, ok := valTok.Token.(token.TokenSingleQuotedString); ok {
					opts.ExternalVolume = &str.Value
				}
			}

		case "CATALOG":
			p.AdvanceToken()
			if p.ConsumeToken(token.TokenEq{}) {
				valTok := p.NextToken()
				if str, ok := valTok.Token.(token.TokenSingleQuotedString); ok {
					opts.Catalog = &str.Value
				}
			}

		case "BASE_LOCATION":
			p.AdvanceToken()
			if p.ConsumeToken(token.TokenEq{}) {
				valTok := p.NextToken()
				if str, ok := valTok.Token.(token.TokenSingleQuotedString); ok {
					opts.BaseLocation = &str.Value
				}
			}

		case "CATALOG_SYNC":
			p.AdvanceToken()
			if p.ConsumeToken(token.TokenEq{}) {
				valTok := p.NextToken()
				if str, ok := valTok.Token.(token.TokenSingleQuotedString); ok {
					opts.CatalogSync = &str.Value
				}
			}

		// DYNAMIC table options
		case "TARGET_LAG":
			p.AdvanceToken()
			if p.ConsumeToken(token.TokenEq{}) {
				valTok := p.NextToken()
				if str, ok := valTok.Token.(token.TokenSingleQuotedString); ok {
					opts.TargetLag = &str.Value
				}
			}

		case "WAREHOUSE":
			p.AdvanceToken()
			if p.ConsumeToken(token.TokenEq{}) {
				ident, _ := p.ParseIdentifier()
				if ident != nil {
					opts.Warehouse = ident
				}
			}

		case "INITIALIZE":
			p.AdvanceToken()
			if p.ConsumeToken(token.TokenEq{}) {
				if p.ParseKeyword("ON_CREATE") {
					kind := expr.InitializeKindOnCreate
					opts.Initialize = &kind
				} else if p.ParseKeyword("ON_SCHEDULE") {
					kind := expr.InitializeKindOnSchedule
					opts.Initialize = &kind
				}
			}

		case "REQUIRE":
			p.AdvanceToken()
			if p.ParseKeyword("USER") {
				opts.RequireUser = true
			}

		default:
			// Not a Snowflake option, return what we have
			return opts
		}
	}

	return opts
}

// Syntax: LIKE table_name [INCLUDING DEFAULTS | EXCLUDING DEFAULTS]
// Reference: src/parser/mod.rs
func parseCreateTableLike(p *Parser) (*expr.CreateTableLikeKind, error) {
	if !p.ParseKeyword("LIKE") {
		return nil, nil
	}

	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	like := &expr.CreateTableLikeKind{
		Kind: expr.CreateTableLikePlain,
		Name: tableName,
	}

	// Check for INCLUDING DEFAULTS or EXCLUDING DEFAULTS
	if p.ParseKeywords([]string{"INCLUDING", "DEFAULTS"}) {
		defaults := expr.CreateTableLikeDefaultsIncluding
		like.Defaults = &defaults
	} else if p.ParseKeywords([]string{"EXCLUDING", "DEFAULTS"}) {
		defaults := expr.CreateTableLikeDefaultsExcluding
		like.Defaults = &defaults
	}

	return like, nil
}

// parseCreateTableColumns parses the parenthesized column list in CREATE TABLE
// Format: (col_def [, col_def ...] [, table_constraint ...])
func parseCreateTableColumns(p *Parser) ([]*expr.ColumnDef, []*expr.TableConstraint, error) {
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, nil, err
	}

	var columns []*expr.ColumnDef
	var constraints []*expr.TableConstraint

	for {
		// Check for end of list
		if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
			p.NextToken() // consume )
			break
		}

		// Check if this is a table constraint (starts with CONSTRAINT or a constraint keyword)
		if isTableConstraint(p) {
			constraint, err := parseTableConstraint(p)
			if err != nil {
				return nil, nil, err
			}
			constraints = append(constraints, constraint)
		} else {
			// Parse column definition
			col, err := parseColumnDef(p)
			if err != nil {
				return nil, nil, err
			}
			columns = append(columns, col)
		}

		// Check for comma
		if !p.ConsumeToken(token.TokenComma{}) {
			// No comma, expect end of list
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, nil, err
			}
			break
		}

		// Handle trailing comma (DuckDB style)
		if p.GetDialect().SupportsTrailingCommas() {
			if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
				p.NextToken() // consume )
				break
			}
		}
	}

	return columns, constraints, nil
}

// isTableConstraint checks if the next token indicates a table constraint
func isTableConstraint(p *Parser) bool {
	// Check for CONSTRAINT keyword
	if p.PeekKeyword("CONSTRAINT") {
		return true
	}

	// Check for table constraint keywords
	tableConstraintKeywords := []string{
		"PRIMARY", "FOREIGN", "UNIQUE", "CHECK",
	}

	for _, kw := range tableConstraintKeywords {
		if p.PeekKeyword(kw) {
			return true
		}
	}

	// MySQL-specific: INDEX/KEY/FULLTEXT/SPATIAL inline index constraints
	// Reference: src/parser/mod.rs:9732-9760
	if p.GetDialect().SupportsIndexHints() {
		mysqlConstraintKeywords := []string{"INDEX", "KEY", "FULLTEXT", "SPATIAL"}
		for _, kw := range mysqlConstraintKeywords {
			if p.PeekKeyword(kw) {
				return true
			}
		}
	}

	return false
}

// parseCreateTableOptions parses optional MySQL table options after the column list.
// Reference: src/parser/mod.rs:8690-8693, parse_plain_options
// Format: [ENGINE=InnoDB] [DEFAULT CHARSET=utf8] [COLLATE=utf8mb4_unicode_ci] [COMMENT='text'] [AUTO_INCREMENT=123] ...
// Options can be space-separated or comma-separated.
func parseCreateTableOptions(p *Parser) (*expr.CreateTableOptions, error) {
	var options []*expr.SqlOption

	for {
		opt, err := parsePlainTableOption(p)
		if err != nil {
			return nil, err
		}
		if opt == nil {
			break
		}
		options = append(options, opt)
		// Consume optional comma between options
		p.ConsumeToken(token.TokenComma{})
	}

	if len(options) == 0 {
		return nil, nil
	}

	return &expr.CreateTableOptions{
		Type:    expr.CreateTableOptionsPlain,
		Options: options,
	}, nil
}

// parsePlainTableOption parses a single MySQL table option.
// Returns nil if the next token is not a recognized option keyword.
// Reference: src/parser/mod.rs:parse_plain_option
func parsePlainTableOption(p *Parser) (*expr.SqlOption, error) {
	// COMMENT option: COMMENT [=] 'string'
	if p.PeekKeyword("COMMENT") {
		p.NextToken()
		p.ConsumeToken(token.TokenEq{})
		val, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		name := &expr.Ident{Value: "COMMENT"}
		return &expr.SqlOption{
			Name:  name,
			Value: &expr.ValueExpr{Value: val},
		}, nil
	}

	// ENGINE option: ENGINE [=] name [(param, ...)]
	if p.PeekKeyword("ENGINE") {
		p.NextToken()
		p.ConsumeToken(token.TokenEq{})
		tok := p.NextToken()
		word, ok := tok.Token.(token.TokenWord)
		if !ok {
			return nil, p.ExpectedRef("engine name (identifier)", &tok)
		}
		engineName := word.Value

		// Check for optional parenthesized parameter list
		var engineParams []*expr.Ident
		if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
			p.NextToken() // consume (
			for {
				if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
					p.NextToken() // consume )
					break
				}
				param, err := p.ParseIdentifier()
				if err != nil {
					return nil, err
				}
				engineParams = append(engineParams, &expr.Ident{Value: param.Value})
				if !p.ConsumeToken(token.TokenComma{}) {
					// No comma, expect )
					if _, isRParen := p.PeekToken().Token.(token.TokenRParen); !isRParen {
						return nil, p.ExpectedRef(") or ,", p.PeekTokenRef())
					}
				}
			}
		}

		return &expr.SqlOption{
			Name:         &expr.Ident{Value: "ENGINE"},
			Value:        &expr.Ident{Value: engineName},
			EngineParams: engineParams,
		}, nil
	}

	// TABLESPACE option: TABLESPACE [=] name [STORAGE [=] {DISK|MEMORY}]
	if p.PeekKeyword("TABLESPACE") {
		p.NextToken()
		p.ConsumeToken(token.TokenEq{})
		tok := p.NextToken()
		var tsName string
		isString := false
		switch t := tok.Token.(type) {
		case token.TokenWord:
			tsName = t.Value
		case token.TokenSingleQuotedString:
			tsName = t.Value
			isString = true
		default:
			return nil, p.ExpectedRef("tablespace name", &tok)
		}

		// Check for optional STORAGE clause
		var storage string
		if p.PeekKeyword("STORAGE") {
			p.NextToken()
			p.ConsumeToken(token.TokenEq{})
			storageTok := p.NextToken()
			if word, ok := storageTok.Token.(token.TokenWord); ok {
				storage = strings.ToUpper(word.Value)
			}
		}

		var value expr.Expr
		if isString {
			value = &expr.ValueExpr{Value: tsName}
		} else {
			value = &expr.Ident{Value: tsName}
		}
		if storage != "" {
			value = &expr.Ident{Value: tsName + " STORAGE " + storage}
		}

		return &expr.SqlOption{
			Name:  &expr.Ident{Value: "TABLESPACE"},
			Value: value,
		}, nil
	}

	// UNION option: UNION [=] (table1, table2, ...)
	if p.PeekKeyword("UNION") {
		p.NextToken()
		p.ConsumeToken(token.TokenEq{})
		if _, isLParen := p.PeekToken().Token.(token.TokenLParen); !isLParen {
			return nil, p.ExpectedRef("\"(\" after UNION", p.PeekTokenRef())
		}
		p.NextToken() // consume (
		var tables []string
		for {
			if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
				p.NextToken() // consume )
				break
			}
			tbl, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			tables = append(tables, tbl.Value)
			p.ConsumeToken(token.TokenComma{})
		}
		return &expr.SqlOption{
			Name:  &expr.Ident{Value: "UNION"},
			Value: &expr.ValueExpr{Value: "(" + strings.Join(tables, ", ") + ")"},
		}, nil
	}

	// Multi-word keys must be checked before single-word keys
	// DEFAULT CHARSET / DEFAULT CHARACTER SET
	if p.PeekKeyword("DEFAULT") {
		// Save position in case this is not a table option
		restore := p.SavePosition()
		p.NextToken() // consume DEFAULT

		var keyName string
		if p.PeekKeyword("CHARSET") {
			p.NextToken()
			keyName = "DEFAULT CHARSET"
		} else if p.PeekKeyword("CHARACTER") {
			p.NextToken()
			if !p.ParseKeyword("SET") {
				restore()
				return nil, nil
			}
			keyName = "DEFAULT CHARACTER SET"
		} else if p.PeekKeyword("COLLATE") {
			p.NextToken()
			keyName = "DEFAULT COLLATE"
		} else {
			restore()
			return nil, nil
		}

		p.ConsumeToken(token.TokenEq{})
		val, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		return &expr.SqlOption{
			Name:  &expr.Ident{Value: keyName},
			Value: &expr.Ident{Value: val.Value},
		}, nil
	}

	// CHARACTER SET (without DEFAULT)
	if p.PeekKeyword("CHARACTER") {
		restore := p.SavePosition()
		p.NextToken()
		if !p.ParseKeyword("SET") {
			restore()
			return nil, nil
		}
		p.ConsumeToken(token.TokenEq{})
		val, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		return &expr.SqlOption{
			Name:  &expr.Ident{Value: "CHARACTER SET"},
			Value: &expr.Ident{Value: val.Value},
		}, nil
	}

	// DATA DIRECTORY / INDEX DIRECTORY
	if p.PeekKeyword("DATA") {
		restore := p.SavePosition()
		p.NextToken()
		if !p.ParseKeyword("DIRECTORY") {
			restore()
			return nil, nil
		}
		p.ConsumeToken(token.TokenEq{})
		val, err := p.ParseStringLiteral()
		if err != nil {
			// Fall back to identifier
			id, err2 := p.ParseIdentifier()
			if err2 != nil {
				return nil, err
			}
			return &expr.SqlOption{
				Name:  &expr.Ident{Value: "DATA DIRECTORY"},
				Value: &expr.Ident{Value: id.Value},
			}, nil
		}
		return &expr.SqlOption{
			Name:  &expr.Ident{Value: "DATA DIRECTORY"},
			Value: &expr.ValueExpr{Value: val},
		}, nil
	}

	if p.PeekKeyword("INDEX") {
		restore := p.SavePosition()
		p.NextToken()
		if !p.ParseKeyword("DIRECTORY") {
			restore()
			return nil, nil
		}
		p.ConsumeToken(token.TokenEq{})
		val, err := p.ParseStringLiteral()
		if err != nil {
			id, err2 := p.ParseIdentifier()
			if err2 != nil {
				return nil, err
			}
			return &expr.SqlOption{
				Name:  &expr.Ident{Value: "INDEX DIRECTORY"},
				Value: &expr.Ident{Value: id.Value},
			}, nil
		}
		return &expr.SqlOption{
			Name:  &expr.Ident{Value: "INDEX DIRECTORY"},
			Value: &expr.ValueExpr{Value: val},
		}, nil
	}

	// Single-word keys
	singleWordKeys := []string{
		"CHARSET", "COLLATE",
		"KEY_BLOCK_SIZE", "ROW_FORMAT", "PACK_KEYS",
		"STATS_AUTO_RECALC", "STATS_PERSISTENT", "STATS_SAMPLE_PAGES",
		"DELAY_KEY_WRITE", "COMPRESSION", "ENCRYPTION",
		"MAX_ROWS", "MIN_ROWS", "AUTOEXTEND_SIZE", "AVG_ROW_LENGTH",
		"CHECKSUM", "CONNECTION", "ENGINE_ATTRIBUTE", "PASSWORD",
		"SECONDARY_ENGINE_ATTRIBUTE", "INSERT_METHOD", "AUTO_INCREMENT",
	}

	for _, kw := range singleWordKeys {
		if p.PeekKeyword(kw) {
			p.NextToken()
			p.ConsumeToken(token.TokenEq{})

			// Try to parse as value, fall back to identifier
			exprParser := NewExpressionParser(p)
			val, err := exprParser.ParseExpr()
			if err != nil {
				// Fall back to identifier
				id, err2 := p.ParseIdentifier()
				if err2 != nil {
					return nil, err
				}
				val = &expr.ValueExpr{Value: id.Value}
			}

			return &expr.SqlOption{
				Name:  &expr.Ident{Value: kw},
				Value: val,
			}, nil
		}
	}

	return nil, nil
}

// parseCreateViewParams parses optional MySQL CREATE VIEW parameters.
// Reference: src/parser/mod.rs parse_create_view_params
// These parameters appear BEFORE the VIEW keyword and can be in any order:
//   - ALGORITHM = UNDEFINED | MERGE | TEMPTABLE
//   - DEFINER = user_specification
//   - SQL SECURITY DEFINER | INVOKER
func parseCreateViewParams(p *Parser) (*statement.CreateViewParams, error) {
	var algorithm *statement.CreateViewAlgorithm
	var definer *statement.GranteeName
	var security *statement.CreateViewSecurity

	for {
		// Try parsing ALGORITHM = ...
		if algorithm == nil && p.PeekKeyword("ALGORITHM") {
			p.NextToken() // consume ALGORITHM
			if !p.ConsumeToken(token.TokenEq{}) {
				return nil, p.ExpectedRef("= after ALGORITHM", p.PeekTokenRef())
			}
			if p.PeekKeyword("UNDEFINED") {
				p.NextToken()
				v := statement.CreateViewAlgorithmUndefined
				algorithm = &v
			} else if p.PeekKeyword("MERGE") {
				p.NextToken()
				v := statement.CreateViewAlgorithmMerge
				algorithm = &v
			} else if p.PeekKeyword("TEMPTABLE") {
				p.NextToken()
				v := statement.CreateViewAlgorithmTempTable
				algorithm = &v
			} else {
				return nil, p.ExpectedRef("UNDEFINED, MERGE, or TEMPTABLE after ALGORITHM =", p.PeekTokenRef())
			}
			continue
		}

		// Try parsing DEFINER = ...
		if definer == nil && p.PeekKeyword("DEFINER") {
			p.NextToken() // consume DEFINER
			if !p.ConsumeToken(token.TokenEq{}) {
				return nil, p.ExpectedRef("= after DEFINER", p.PeekTokenRef())
			}
			d, err := parseGranteeName(p)
			if err != nil {
				return nil, err
			}
			definer = d
			continue
		}

		// Try parsing SQL SECURITY ...
		if security == nil && p.PeekKeyword("SQL") {
			restore := p.SavePosition()
			p.NextToken() // consume SQL
			if p.PeekKeyword("SECURITY") {
				p.NextToken() // consume SECURITY
				if p.PeekKeyword("DEFINER") {
					p.NextToken()
					v := statement.CreateViewSecurityDefiner
					security = &v
				} else if p.PeekKeyword("INVOKER") {
					p.NextToken()
					v := statement.CreateViewSecurityInvoker
					security = &v
				} else {
					return nil, p.ExpectedRef("DEFINER or INVOKER after SQL SECURITY", p.PeekTokenRef())
				}
			} else {
				restore()
			}
			continue
		}

		// No more recognized parameters
		break
	}

	if algorithm != nil || definer != nil || security != nil {
		return &statement.CreateViewParams{
			Algorithm: algorithm,
			Definer:   definer,
			Security:  security,
		}, nil
	}
	return nil, nil
}

// parseCreateView parses CREATE VIEW
// Reference: src/parser/mod.rs parse_create_view (lines 6417-6509)
func parseCreateView(p *Parser, orReplace, temporary bool, secure bool, materialized ...bool) (ast.Statement, error) {
	isMaterialized := len(materialized) > 0 && materialized[0]

	// Parse optional MySQL CREATE VIEW parameters BEFORE the VIEW keyword
	// Reference: src/parser/mod.rs parse_create_view_params
	// These can appear in any order: ALGORITHM = ..., DEFINER = ..., SQL SECURITY ...
	params, err := parseCreateViewParams(p)
	if err != nil {
		return nil, err
	}

	// VIEW keyword is expected (already checked by caller)
	if _, err := p.ExpectKeyword("VIEW"); err != nil {
		return nil, err
	}

	// Check for IF NOT EXISTS (before name)
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse view name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Check for IF NOT EXISTS after name (Snowflake style)
	if !ifNotExists {
		ifNotExists = p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})
	}

	// Parse COPY GRANTS (Snowflake)
	copyGrants := p.ParseKeywords([]string{"COPY", "GRANTS"})

	// Parse optional column list: (col1, col2, ...) or (col1 WITH TAG (...), ...)
	var columns []*expr.ViewColumnDef
	if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
		columns, err = p.ParseViewColumns()
		if err != nil {
			return nil, err
		}
	}

	// Parse WITH options (various dialects)
	var options []*expr.SqlOption
	if p.ParseKeyword("WITH") {
		if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
			opts, err := parseSqlOptions(p)
			if err != nil {
				return nil, err
			}
			options = opts
		}
	}

	// Parse CLUSTER BY (BigQuery/Snowflake)
	var clusterBy []expr.Expr
	if p.ParseKeyword("CLUSTER") {
		if _, err := p.ExpectKeyword("BY"); err != nil {
			return nil, err
		}
		clusterCols, err := p.ParseParenthesizedColumnList()
		if err != nil {
			return nil, err
		}
		// Convert []*ast.Ident to []expr.Expr
		for _, col := range clusterCols {
			clusterBy = append(clusterBy, &expr.Ident{Value: col.Value})
		}
	}

	// Parse BigQuery OPTIONS - check by dialect name
	dialectName := p.GetDialect().Dialect()
	if dialectName == "bigquery" || dialectName == "generic" {
		if p.ParseKeyword("OPTIONS") {
			if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
				opts, err := parseSqlOptions(p)
				if err != nil {
					return nil, err
				}
				options = append(options, opts...)
			}
		}
	}

	// Parse TO clause (ClickHouse) - currently not stored due to type mismatch
	// TODO: Fix AST To field type to accept ObjectName
	if (dialectName == "clickhouse" || dialectName == "generic") && p.ParseKeyword("TO") {
		_, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		// to = toName - cannot assign due to type mismatch
	}

	// Parse COMMENT (various dialects)
	var comment *expr.CommentDef
	if p.GetDialect().SupportsCreateViewCommentSyntax() && p.ParseKeyword("COMMENT") {
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return nil, err
		}
		commentStr, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		comment = &expr.CommentDef{
			Type:    expr.CommentDefWithEq,
			Comment: commentStr,
		}
	}

	// Expect AS keyword
	if _, err := p.ExpectKeyword("AS"); err != nil {
		return nil, err
	}

	// Parse the query
	stmt, err := p.ParseQuery()
	if err != nil {
		return nil, err
	}

	// Convert statement to *query.Query
	var q *query.Query
	switch s := stmt.(type) {
	case *SelectStatement:
		selectCopy := s.Select
		q = &query.Query{
			Body: &query.SelectSetExpr{Select: &selectCopy},
		}
	case *QueryStatement:
		q = s.Query
	case *statement.Query:
		q = s.Query
	default:
		return nil, fmt.Errorf("expected SELECT query in CREATE VIEW, got %T", stmt)
	}

	// Parse WITH NO SCHEMA BINDING (Redshift)
	withNoSchemaBinding := false
	if dialectName == "redshift" || dialectName == "generic" {
		if p.ParseKeywords([]string{"WITH", "NO", "SCHEMA", "BINDING"}) {
			withNoSchemaBinding = true
		}
	}

	return &statement.CreateView{
		OrReplace:           orReplace,
		Temporary:           temporary,
		Materialized:        isMaterialized,
		Secure:              secure,
		IfNotExists:         ifNotExists,
		Name:                name,
		Columns:             columns,
		Query:               q,
		Options:             options,
		ClusterBy:           clusterBy,
		WithNoSchemaBinding: withNoSchemaBinding,
		Comment:             comment,
		CopyGrants:          copyGrants,
		Params:              params,
	}, nil
}

// parseCreateIndex parses CREATE INDEX
// Reference: src/parser/mod.rs parse_create_index
func parseCreateIndex(p *Parser, unique bool) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("INDEX"); err != nil {
		return nil, err
	}

	// Parse CONCURRENTLY (PostgreSQL specific)
	concurrently := p.ParseKeyword("CONCURRENTLY")

	// Parse IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Check if index name is provided
	// In PostgreSQL, the index name is optional: CREATE INDEX ON table_name (col)
	// MySQL requires the index name: CREATE INDEX name ON table_name (col)
	var indexName *ast.Ident
	var using *ast.Ident

	// Check if we have ON keyword (meaning no index name)
	if !p.PeekKeyword("ON") {
		// Parse index name
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		indexName = name

		// Check for USING after index name (MySQL style: CREATE INDEX name USING btree ON ...)
		if p.ParseKeyword("USING") {
			using, err = p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
		}
	}

	// Expect ON keyword (whether we had an index name or not)
	if _, err := p.ExpectKeyword("ON"); err != nil {
		return nil, err
	}

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Check for USING after table name (PostgreSQL style: CREATE INDEX ON table USING btree ...)
	if using == nil && p.ParseKeyword("USING") {
		using, err = p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
	}

	// Parse column list: (col1, col2, ...)
	columns, err := parseIndexColumnList(p)
	if err != nil {
		return nil, err
	}

	// Parse INCLUDE clause (PostgreSQL 11+)
	var include []*ast.Ident
	if p.ParseKeyword("INCLUDE") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		include, err = parseCommaSeparatedIdents(p)
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	// Parse NULLS DISTINCT / NULLS NOT DISTINCT (PostgreSQL 15+)
	var nullsDistinct *bool
	if p.ParseKeyword("NULLS") {
		notDistinct := p.ParseKeyword("NOT")
		if !notDistinct && !p.ParseKeyword("DISTINCT") {
			return nil, p.ExpectedRef("NOT DISTINCT or DISTINCT", p.PeekTokenRef())
		}
		if notDistinct {
			if !p.ParseKeyword("DISTINCT") {
				return nil, p.ExpectedRef("DISTINCT after NULLS NOT", p.PeekTokenRef())
			}
			val := false
			nullsDistinct = &val
		} else {
			val := true
			nullsDistinct = &val
		}
	}

	// Parse WITH (storage_parameters) clause
	var withOpts []*expr.SqlOption
	if p.ParseKeyword("WITH") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		withOpts, err = parseSqlOptions(p)
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	// Parse TABLESPACE clause
	var tablespace *ast.Ident
	if p.ParseKeyword("TABLESPACE") {
		tablespace, err = p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
	}

	// Parse WHERE clause (partial index)
	var predicate expr.Expr
	if p.ParseKeyword("WHERE") {
		exprParser := NewExpressionParser(p)
		predicate, err = exprParser.ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	return &statement.CreateIndex{
		Unique:        unique,
		Concurrently:  concurrently,
		IfNotExists:   ifNotExists,
		Name:          indexName,
		TableName:     tableName,
		Using:         using,
		Columns:       columns,
		Include:       include,
		NullsDistinct: nullsDistinct,
		With:          withOpts,
		TableSpace:    tablespace,
		Predicate:     predicate,
	}, nil
}

// parseIndexColumnList parses a parenthesized list of index columns
func parseIndexColumnList(p *Parser) ([]*expr.IndexColumn, error) {
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	var columns []*expr.IndexColumn
	if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
		// Empty list
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		return columns, nil
	}

	for {
		col, err := parseIndexColumn(p)
		if err != nil {
			return nil, err
		}
		columns = append(columns, col)

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return columns, nil
}

// parseIndexColumn parses a single index column expression
// Format: expression [ASC|DESC] [NULLS FIRST|LAST] or expression opclass [ASC|DESC]
func parseIndexColumn(p *Parser) (*expr.IndexColumn, error) {
	exprParser := NewExpressionParser(p)

	// Parse the expression
	colExpr, err := exprParser.ParseExpr()
	if err != nil {
		return nil, err
	}

	col := &expr.IndexColumn{
		Expr: colExpr,
	}

	// Check for operator class (only if next token is not a keyword like ASC, DESC, NULLS)
	if !p.PeekKeyword("ASC") && !p.PeekKeyword("DESC") && !p.PeekKeyword("NULLS") {
		// Try to parse as operator class
		if opclass, err := tryParseOpclass(p); err == nil && opclass != nil {
			col.Opclass = opclass
		}
	}

	// Check for ASC/DESC
	if p.ParseKeyword("ASC") {
		asc := true
		col.Asc = &asc
	} else if p.ParseKeyword("DESC") {
		asc := false
		col.Asc = &asc
	}

	// Check for NULLS FIRST/LAST
	if p.ParseKeyword("NULLS") {
		if p.ParseKeyword("FIRST") {
			nullsFirst := true
			col.NullsFirst = &nullsFirst
		} else if p.ParseKeyword("LAST") {
			nullsFirst := false
			col.NullsFirst = &nullsFirst
		}
	}

	return col, nil
}

// tryParseOpclass attempts to parse an operator class
func tryParseOpclass(p *Parser) (*ast.ObjectName, error) {
	// Save position
	restore := p.SavePosition()

	// Try to parse as object name
	name, err := p.ParseObjectName()
	if err != nil {
		restore()
		return nil, err
	}

	// Check if the next token is one of the keywords that would indicate
	// this is not an operator class but part of the expression
	if p.PeekKeyword("ASC") || p.PeekKeyword("DESC") || p.PeekKeyword("NULLS") || p.PeekKeyword(",") {
		// This is an operator class
		return name, nil
	}

	// Check for right parenthesis
	if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
		// This is an operator class
		return name, nil
	}

	// Not followed by expected tokens, restore position
	restore()
	return nil, nil
}

// parseSqlOptions parses a comma-separated list of SQL options like (key = value, ...)
func parseSqlOptions(p *Parser) ([]*expr.SqlOption, error) {
	var options []*expr.SqlOption
	if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
		return options, nil
	}

	for {
		// Parse option name
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}

		// Expect =
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return nil, err
		}

		// Parse value
		exprParser := NewExpressionParser(p)
		val, err := exprParser.ParseExpr()
		if err != nil {
			return nil, err
		}

		// Convert ast.Ident to expr.Ident
		exprName := &expr.Ident{
			SpanVal:    name.Span(),
			Value:      name.Value,
			QuoteStyle: name.QuoteStyle,
		}

		options = append(options, &expr.SqlOption{
			Name:  exprName,
			Value: val,
		})

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return options, nil
}

func parseCreateRole(p *Parser, orReplace bool) (ast.Statement, error) {
	// ROLE keyword is expected (already checked by caller)
	if _, err := p.ExpectKeyword("ROLE"); err != nil {
		return nil, err
	}

	// Parse IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse comma-separated role names (identifiers, not full object names)
	names, err := parseCommaSeparatedIdents(p)
	if err != nil {
		return nil, err
	}

	return &statement.CreateRole{
		IfNotExists: ifNotExists,
		Names:       names,
		Options:     nil, // Role options not yet implemented (Postgres/MSSQL specific)
	}, nil
}

func parseCreateDatabase(p *Parser) (ast.Statement, error) {
	// Consume DATABASE keyword
	if _, err := p.ExpectKeyword("DATABASE"); err != nil {
		return nil, err
	}

	// Check for IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse database name
	dbName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse optional LOCATION/MANAGEDLOCATION (Hive/Databricks style)
	var location, managedLocation *string
	for {
		if p.PeekKeyword("LOCATION") {
			p.NextToken()
			loc, err := p.ParseStringLiteral()
			if err != nil {
				return nil, err
			}
			location = &loc
		} else if p.PeekKeyword("MANAGEDLOCATION") {
			p.NextToken()
			loc, err := p.ParseStringLiteral()
			if err != nil {
				return nil, err
			}
			managedLocation = &loc
		} else {
			break
		}
	}

	// Parse optional CLONE (Snowflake style)
	var clone *ast.ObjectName
	if p.PeekKeyword("CLONE") {
		p.NextToken()
		clone, err = p.ParseObjectName()
		if err != nil {
			return nil, err
		}
	}

	// Parse MySQL-style [DEFAULT] CHARACTER SET and [DEFAULT] COLLATE options
	var defaultCharset, defaultCollation *string
	for {
		hasDefault := p.PeekKeyword("DEFAULT")
		if hasDefault {
			p.NextToken()
		}

		if p.PeekKeyword("CHARACTER") {
			p.NextToken()
			if !p.ParseKeyword("SET") {
				// Not CHARACTER SET, put back and break
				break
			}
			p.ConsumeToken(token.TokenEq{}) // Optional =
			charset, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			defaultCharset = &charset.Value
		} else if p.PeekKeyword("CHARSET") {
			p.NextToken()
			p.ConsumeToken(token.TokenEq{}) // Optional =
			charset, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			defaultCharset = &charset.Value
		} else if p.PeekKeyword("COLLATE") {
			p.NextToken()
			p.ConsumeToken(token.TokenEq{}) // Optional =
			collation, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			defaultCollation = &collation.Value
		} else if hasDefault {
			// DEFAULT keyword not followed by CHARACTER SET, CHARSET, or COLLATE
			// Put it back and break
			break
		} else {
			break
		}
	}

	return &statement.CreateDatabase{
		DbName:           dbName,
		IfNotExists:      ifNotExists,
		Location:         location,
		ManagedLocation:  managedLocation,
		Clone:            clone,
		DefaultCharset:   defaultCharset,
		DefaultCollation: defaultCollation,
	}, nil
}

func parseCreateSchema(p *Parser) (ast.Statement, error) {
	// SCHEMA keyword is expected (already checked by caller)
	if _, err := p.ExpectKeyword("SCHEMA"); err != nil {
		return nil, err
	}

	// Check for IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse schema name (handles AUTHORIZATION variants)
	schemaName, err := parseSchemaName(p)
	if err != nil {
		return nil, err
	}

	// Parse optional DEFAULT COLLATE (BigQuery)
	var defaultCollateSpec expr.Expr
	if p.ParseKeywords([]string{"DEFAULT", "COLLATE"}) {
		exprParser := NewExpressionParser(p)
		defaultCollateSpec, err = exprParser.ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	// Parse optional WITH options (Trino)
	var withOpts *[]*expr.SqlOption
	if p.PeekKeyword("WITH") {
		p.NextToken() // consume WITH
		opts, err := parseOptions(p)
		if err != nil {
			return nil, err
		}
		withOpts = &opts
	}

	// Parse optional OPTIONS (BigQuery)
	var options *[]*expr.SqlOption
	if p.PeekKeyword("OPTIONS") {
		p.NextToken() // consume OPTIONS
		opts, err := parseOptions(p)
		if err != nil {
			return nil, err
		}
		options = &opts
	}

	// Parse optional CLONE (Snowflake)
	var clone *ast.ObjectName
	if p.ParseKeyword("CLONE") {
		clone, err = p.ParseObjectName()
		if err != nil {
			return nil, err
		}
	}

	return &statement.CreateSchema{
		SchemaName:         schemaName,
		IfNotExists:        ifNotExists,
		With:               withOpts,
		Options:            options,
		DefaultCollateSpec: defaultCollateSpec,
		Clone:              clone,
	}, nil
}

// parseSchemaName parses schema name with optional AUTHORIZATION clause
// Reference: src/parser/mod.rs parse_schema_name
// Supports:
//   - Simple: <schema_name>
//   - UnnamedAuthorization: AUTHORIZATION <user>
//   - NamedAuthorization: <schema_name> AUTHORIZATION <user>
func parseSchemaName(p *Parser) (*expr.SchemaName, error) {
	// Check for AUTHORIZATION first (UnnamedAuthorization case)
	if p.ParseKeyword("AUTHORIZATION") {
		auth, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected authorization identifier after AUTHORIZATION: %w", err)
		}
		return &expr.SchemaName{
			Authorization:    auth,
			HasAuthorization: true,
		}, nil
	}

	// Parse the schema name (could be simple identifier or object name)
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected schema name: %w", err)
	}

	// Check for AUTHORIZATION after name (NamedAuthorization case)
	if p.ParseKeyword("AUTHORIZATION") {
		auth, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected authorization identifier after AUTHORIZATION: %w", err)
		}
		return &expr.SchemaName{
			Name:             name,
			Authorization:    auth,
			HasAuthorization: true,
		}, nil
	}

	// Simple case: just the name
	return &expr.SchemaName{
		Name: name,
	}, nil
}

func parseCreateSequence(p *Parser, orReplace, temporary bool) (ast.Statement, error) {
	// Consume SEQUENCE keyword (already checked by caller)
	if _, err := p.ExpectKeyword("SEQUENCE"); err != nil {
		return nil, err
	}

	// Parse IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse sequence name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse AS data_type (optional)
	var dataType string
	if p.ParseKeyword("AS") {
		dt, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		// Store the data type name as a string
		dataType = dt.Value
	}

	// Parse sequence options
	sequenceOptions, err := parseCreateSequenceOptions(p)
	if err != nil {
		return nil, err
	}

	// Parse OWNED BY clause
	var ownedBy *ast.ObjectName
	if p.ParseKeywords([]string{"OWNED", "BY"}) {
		if p.ParseKeyword("NONE") {
			// OWNED BY NONE - represented as a special identifier
			ownedBy = &ast.ObjectName{
				Parts: []ast.ObjectNamePart{&ast.ObjectNamePartIdentifier{Ident: &ast.Ident{Value: "NONE"}}},
			}
		} else {
			ownedBy, err = p.ParseObjectName()
			if err != nil {
				return nil, err
			}
		}
	}

	return &statement.CreateSequence{
		Temporary:       temporary,
		IfNotExists:     ifNotExists,
		Name:            name,
		DataType:        dataType,
		SequenceOptions: sequenceOptions,
		OwnedBy:         ownedBy,
	}, nil
}

// parseCreateSequenceOptions parses the various options for CREATE SEQUENCE
func parseCreateSequenceOptions(p *Parser) ([]*expr.SequenceOptions, error) {
	var sequenceOptions []*expr.SequenceOptions
	exprParser := NewExpressionParser(p)

	for {
		// INCREMENT [BY] increment
		if p.ParseKeyword("INCREMENT") {
			hasBy := p.ParseKeyword("BY")
			incExpr, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:        expr.SeqOptIncrementBy,
				Expr:        incExpr,
				HasByOrWith: hasBy,
			})
			continue
		}

		// MINVALUE minvalue | NO MINVALUE
		if p.ParseKeyword("MINVALUE") {
			minExpr, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptMinValue,
				Expr:    minExpr,
				NoValue: false,
			})
			continue
		} else if p.ParseKeywords([]string{"NO", "MINVALUE"}) {
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptMinValue,
				NoValue: true,
			})
			continue
		}

		// MAXVALUE maxvalue | NO MAXVALUE
		if p.ParseKeyword("MAXVALUE") {
			maxExpr, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptMaxValue,
				Expr:    maxExpr,
				NoValue: false,
			})
			continue
		} else if p.ParseKeywords([]string{"NO", "MAXVALUE"}) {
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptMaxValue,
				NoValue: true,
			})
			continue
		}

		// START [WITH] start
		if p.ParseKeyword("START") {
			hasWith := p.ParseKeyword("WITH")
			startExpr, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:        expr.SeqOptStartWith,
				Expr:        startExpr,
				HasByOrWith: hasWith,
			})
			continue
		}

		// CACHE cache
		if p.ParseKeyword("CACHE") {
			cacheExpr, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type: expr.SeqOptCache,
				Expr: cacheExpr,
			})
			continue
		}

		// [NO] CYCLE
		if p.ParseKeywords([]string{"NO", "CYCLE"}) {
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptCycle,
				NoCycle: true,
			})
			continue
		} else if p.ParseKeyword("CYCLE") {
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptCycle,
				NoCycle: false,
			})
			continue
		}

		// No more options
		break
	}

	return sequenceOptions, nil
}

func parseCreateType(p *Parser) (ast.Statement, error) {
	// Consume TYPE keyword (already checked by caller)
	if _, err := p.ExpectKeyword("TYPE"); err != nil {
		return nil, err
	}

	// Parse type name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Check for AS keyword
	hasAs := p.ParseKeyword("AS")

	if !hasAs {
		// Check for CREATE TYPE name (options) - SQL definition without AS
		if p.ConsumeToken(token.TokenLParen{}) {
			// Parse SQL definition options
			var options []string
			for {
				if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
					break
				}
				// Parse option as identifier = value
				optTok := p.PeekToken()
				p.AdvanceToken()
				optStr := optTok.Token.String()
				if p.ConsumeToken(token.TokenEq{}) {
					valTok := p.PeekToken()
					p.AdvanceToken()
					optStr = optStr + " = " + valTok.Token.String()
				}
				options = append(options, optStr)
				if !p.ConsumeToken(token.TokenComma{}) {
					break
				}
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
			repr := expr.UserDefinedTypeSqlDefinition{
				Options: options,
			}
			var r expr.UserDefinedTypeRepresentation = &repr
			return &statement.CreateType{
				Name:           name,
				Representation: &r,
			}, nil
		}

		// Simple CREATE TYPE name;
		return &statement.CreateType{
			Name: name,
		}, nil
	}

	// Parse AS variant
	if p.ParseKeyword("ENUM") {
		// CREATE TYPE name AS ENUM (...)
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		// Parse enum labels (comma-separated identifiers)
		var labels []*ast.Ident
		for {
			if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
				break
			}
			label, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			labels = append(labels, label)
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		enumRepr := &expr.UserDefinedTypeEnum{
			Labels: labels,
		}
		var r expr.UserDefinedTypeRepresentation = enumRepr
		return &statement.CreateType{
			Name:           name,
			Representation: &r,
		}, nil
	}

	if p.ParseKeyword("RANGE") {
		// CREATE TYPE name AS RANGE (...)
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		// Parse range options - simplified
		var options []string
		for {
			if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
				break
			}
			optTok := p.PeekToken()
			p.AdvanceToken()
			optStr := optTok.Token.String()
			if p.ConsumeToken(token.TokenEq{}) {
				valTok := p.PeekToken()
				p.AdvanceToken()
				optStr = optStr + " = " + valTok.Token.String()
			}
			options = append(options, optStr)
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		rangeRepr := &expr.UserDefinedTypeRange{
			Options: options,
		}
		var r expr.UserDefinedTypeRepresentation = rangeRepr
		return &statement.CreateType{
			Name:           name,
			Representation: &r,
		}, nil
	}

	// Try composite type: CREATE TYPE name AS (attr1 type1, attr2 type2, ...)
	if p.ConsumeToken(token.TokenLParen{}) {
		// Parse composite attributes
		var attributes []*expr.UserDefinedTypeCompositeAttributeDef
		for {
			if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
				break
			}
			// Parse attribute name
			attrName, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			// Parse data type
			attrDataType, err := p.ParseDataType()
			if err != nil {
				return nil, err
			}
			var attrCollation *ast.ObjectName
			// Optional COLLATE
			if p.ParseKeyword("COLLATE") {
				attrCollation, _ = p.ParseObjectName()
			}
			attributes = append(attributes, &expr.UserDefinedTypeCompositeAttributeDef{
				Name:      attrName,
				DataType:  attrDataType,
				Collation: attrCollation,
			})
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		compositeRepr := &expr.UserDefinedTypeComposite{
			Attributes: attributes,
		}
		var r expr.UserDefinedTypeRepresentation = compositeRepr
		return &statement.CreateType{
			Name:           name,
			Representation: &r,
		}, nil
	}

	return nil, p.expectedRef("ENUM, RANGE, or '(' after AS", p.PeekTokenRef())
}

func parseCreateDomain(p *Parser) (ast.Statement, error) {
	// Consume DOMAIN keyword (already checked by caller)
	if _, err := p.ExpectKeyword("DOMAIN"); err != nil {
		return nil, err
	}

	// Parse domain name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse AS keyword
	if _, err := p.ExpectKeyword("AS"); err != nil {
		return nil, err
	}

	// Parse data type
	dataType, err := p.ParseDataType()
	if err != nil {
		return nil, err
	}

	// Parse optional COLLATE
	var collation *ast.ObjectName
	if p.ParseKeyword("COLLATE") {
		collation, _ = p.ParseObjectName()
	}

	// Parse optional DEFAULT
	var defaultValue expr.Expr
	if p.ParseKeyword("DEFAULT") {
		defaultValue, _ = NewExpressionParser(p).ParseExpr()
	}

	// Parse optional constraints
	var constraints []*expr.DomainConstraint
	for {
		// Check for EOF or end of statement
		tok := p.PeekToken()
		if _, isEOF := tok.Token.(token.EOF); isEOF {
			break
		}
		// Check for semicolon
		if word, ok := tok.Token.(token.TokenWord); ok && word.Value == ";" {
			break
		}

		// Parse constraint name if present
		var constraintName *ast.Ident
		if p.PeekKeyword("CONSTRAINT") {
			p.AdvanceToken() // consume CONSTRAINT
			constraintName, _ = p.ParseIdentifier()
		}

		// Check for constraint types
		if p.PeekKeyword("NOT") {
			p.AdvanceToken() // consume NOT
			if p.PeekKeyword("NULL") {
				p.AdvanceToken() // consume NULL
				constraints = append(constraints, &expr.DomainConstraint{
					Name: constraintName,
					Type: expr.DomainConstraintNotNull,
				})
				continue
			}
		} else if p.PeekKeyword("NULL") {
			p.AdvanceToken() // consume NULL
			constraints = append(constraints, &expr.DomainConstraint{
				Name: constraintName,
				Type: expr.DomainConstraintNull,
			})
			continue
		} else if p.PeekKeyword("CHECK") {
			p.AdvanceToken() // consume CHECK
			// Parse CHECK expression
			if p.ConsumeToken(token.TokenLParen{}) {
				checkExpr, _ := NewExpressionParser(p).ParseExpr()
				p.ConsumeToken(token.TokenRParen{})
				constraints = append(constraints, &expr.DomainConstraint{
					Name:      constraintName,
					Type:      expr.DomainConstraintCheck,
					CheckExpr: checkExpr,
				})
			}
			continue
		} else {
			// If we consumed a constraint name but no constraint, put it back
			if constraintName != nil {
				// Can't put back, so just continue
			}
			break
		}
	}

	return &statement.CreateDomain{
		Name:         name,
		DataType:     dataType,
		Collation:    collation,
		DefaultValue: defaultValue,
		Constraints:  constraints,
	}, nil
}

func parseCreateExtension(p *Parser) (ast.Statement, error) {
	// Reference: src/parser/mod.rs:8018-8050
	// Parse optional IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse extension name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Parse optional WITH clause
	var schema *ast.ObjectName
	var version *string
	cascade := false

	if p.ParseKeyword("WITH") {
		// Parse optional SCHEMA
		if p.ParseKeyword("SCHEMA") {
			schemaIdent, err := p.ParseIdentifier()
			if err == nil {
				schema = ast.NewObjectNameFromIdents(schemaIdent)
			}
		}

		// Parse optional VERSION
		if p.ParseKeyword("VERSION") {
			verIdent, err := p.ParseIdentifier()
			if err == nil {
				v := verIdent.Value
				version = &v
			}
		}

		// Parse optional CASCADE
		cascade = p.ParseKeyword("CASCADE")
	}

	return &statement.CreateExtension{
		IfNotExists: ifNotExists,
		Name:        name,
		Schema:      schema,
		Version:     version,
		Cascade:     cascade,
	}, nil
}

func parseCreateTrigger(p *Parser, orReplace bool) (ast.Statement, error) {
	// Consume TRIGGER keyword
	if _, err := p.ExpectKeyword("TRIGGER"); err != nil {
		return nil, err
	}

	// Parse trigger name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	trigger := &statement.CreateTrigger{
		Name:              name,
		OrReplace:         orReplace,
		PeriodBeforeTable: true, // Default to PostgreSQL/MySQL style (period before ON)
	}

	// Parse optional period (BEFORE, AFTER, INSTEAD OF, FOR)
	if p.PeekKeyword("BEFORE") {
		p.NextToken()
		period := expr.TriggerPeriodBefore
		trigger.Period = &period
	} else if p.PeekKeyword("AFTER") {
		p.NextToken()
		period := expr.TriggerPeriodAfter
		trigger.Period = &period
	} else if p.PeekKeyword("INSTEAD") {
		p.NextToken()
		if _, err := p.ExpectKeyword("OF"); err != nil {
			return nil, err
		}
		period := expr.TriggerPeriodInsteadOf
		trigger.Period = &period
	} else if p.PeekKeyword("FOR") {
		p.NextToken()
		period := expr.TriggerPeriodFor
		trigger.Period = &period
	}

	// Parse trigger events (can be OR-separated: INSERT OR UPDATE OR DELETE)
	for {
		event, eventCols, err := parseTriggerEvent(p)
		if err != nil {
			return nil, err
		}
		trigger.Events = append(trigger.Events, &expr.TriggerEventWithColumns{
			Event:   event,
			Columns: eventCols,
		})

		// Check for OR keyword to parse more events
		if !p.PeekKeyword("OR") {
			break
		}
		p.NextToken() // consume OR
	}

	// Expect ON keyword
	if _, err := p.ExpectKeyword("ON"); err != nil {
		return nil, err
	}

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}
	trigger.TableName = tableName

	// Parse optional FROM clause (referenced table)
	if p.PeekKeyword("FROM") {
		p.NextToken()
		refTable, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		trigger.ReferencedTableName = refTable
	}

	// Parse optional constraint characteristics (DEFERRABLE, etc.)
	// For now, just parse and discard - full implementation TODO
	parseConstraintCharacteristics(p)

	// Parse optional REFERENCING clause
	if p.PeekKeyword("REFERENCING") {
		p.NextToken()
		for {
			referencing, err := parseTriggerReferencing(p)
			if err != nil {
				return nil, err
			}
			if referencing == nil {
				break
			}
			trigger.Referencing = append(trigger.Referencing, referencing)
			// Check if next token could be another referencing clause
			if !p.PeekKeyword("OLD") && !p.PeekKeyword("NEW") {
				break
			}
		}
	}

	// Parse optional FOR [EACH] ROW/STATEMENT
	if p.PeekKeyword("FOR") {
		p.NextToken()
		kind := expr.TriggerObjectKindFor
		if p.PeekKeyword("EACH") {
			p.NextToken()
			kind = expr.TriggerObjectKindForEach
		}

		var obj expr.TriggerObject
		if p.PeekKeyword("ROW") {
			p.NextToken()
			obj = expr.TriggerObjectRow
		} else if p.PeekKeyword("STATEMENT") {
			p.NextToken()
			obj = expr.TriggerObjectStatement
		}
		trigger.TriggerObject = &expr.TriggerObjectKindWithObject{
			Kind:   kind,
			Object: obj,
		}
	}

	// Parse optional WHEN clause
	if p.PeekKeyword("WHEN") {
		p.NextToken()
		condition, err := NewExpressionParser(p).ParseExpr()
		if err != nil {
			return nil, err
		}
		trigger.Condition = condition
	}

	// Parse EXECUTE clause (FUNCTION or PROCEDURE)
	if p.PeekKeyword("EXECUTE") {
		p.NextToken()
		execBody, err := parseTriggerExecBody(p)
		if err != nil {
			return nil, err
		}
		trigger.ExecBody = execBody
	} else {
		// Parse statement body (for T-SQL style triggers)
		// For now, skip this - it's complex conditional statement parsing
	}

	return trigger, nil
}

func parseTriggerEvent(p *Parser) (expr.TriggerEvent, []*ast.Ident, error) {
	switch {
	case p.PeekKeyword("INSERT"):
		p.NextToken()
		return expr.TriggerEventInsert, nil, nil
	case p.PeekKeyword("UPDATE"):
		p.NextToken()
		var cols []*ast.Ident
		if p.PeekKeyword("OF") {
			p.NextToken()
			// Parse column list
			for {
				col, err := p.ParseIdentifier()
				if err != nil {
					return expr.TriggerEventNone, nil, err
				}
				cols = append(cols, col)
				if !p.ConsumeToken(token.TokenComma{}) {
					break
				}
			}
		}
		return expr.TriggerEventUpdate, cols, nil
	case p.PeekKeyword("DELETE"):
		p.NextToken()
		return expr.TriggerEventDelete, nil, nil
	case p.PeekKeyword("TRUNCATE"):
		p.NextToken()
		return expr.TriggerEventTruncate, nil, nil
	default:
		return expr.TriggerEventNone, nil, fmt.Errorf("expected INSERT, UPDATE, DELETE, or TRUNCATE, found %v", p.PeekToken())
	}
}

func parseTriggerReferencing(p *Parser) (*expr.TriggerReferencing, error) {
	var referType expr.TriggerReferencingType

	if p.PeekKeyword("OLD") {
		p.NextToken()
		if p.PeekKeyword("TABLE") {
			p.NextToken()
			referType = expr.TriggerReferencingTypeOldTable
		} else {
			// Not a valid referencing clause
			return nil, nil
		}
	} else if p.PeekKeyword("NEW") {
		p.NextToken()
		if p.PeekKeyword("TABLE") {
			p.NextToken()
			referType = expr.TriggerReferencingTypeNewTable
		} else {
			// Not a valid referencing clause
			return nil, nil
		}
	} else {
		return nil, nil
	}

	isAs := p.PeekKeyword("AS")
	if isAs {
		p.NextToken()
	}

	transitionName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	return &expr.TriggerReferencing{
		ReferType:              referType,
		IsAs:                   isAs,
		TransitionRelationName: transitionName,
	}, nil
}

func parseTriggerExecBody(p *Parser) (*expr.TriggerExecBody, error) {
	var execType expr.TriggerExecBodyType

	if p.PeekKeyword("FUNCTION") {
		p.NextToken()
		execType = expr.TriggerExecBodyTypeFunction
	} else if p.PeekKeyword("PROCEDURE") {
		p.NextToken()
		execType = expr.TriggerExecBodyTypeProcedure
	} else {
		return nil, fmt.Errorf("expected FUNCTION or PROCEDURE after EXECUTE")
	}

	// Parse function/procedure name and optional arguments
	funcName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	var args []expr.Expr
	if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
		p.NextToken() // consume (
		if _, ok := p.PeekToken().Token.(token.TokenRParen); !ok {
			for {
				arg, err := NewExpressionParser(p).ParseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				if !p.ConsumeToken(token.TokenComma{}) {
					break
				}
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	return &expr.TriggerExecBody{
		ExecType: execType,
		FuncDesc: &expr.FunctionDesc{
			Name: funcName,
			Args: args,
		},
	}, nil
}

func parseCreatePolicy(p *Parser, orReplace bool) (ast.Statement, error) {
	// Parse policy name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Parse ON keyword and table name
	if _, err := p.ExpectKeyword("ON"); err != nil {
		return nil, err
	}
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse optional AS PERMISSIVE/RESTRICTIVE
	var policyType *expr.CreatePolicyType
	if p.ParseKeyword("AS") {
		if p.ParseKeyword("PERMISSIVE") {
			pt := expr.CreatePolicyTypePermissive
			policyType = &pt
		} else if p.ParseKeyword("RESTRICTIVE") {
			pt := expr.CreatePolicyTypeRestrictive
			policyType = &pt
		} else {
			return nil, p.Expected("PERMISSIVE or RESTRICTIVE after AS", p.PeekToken())
		}
	}

	// Parse optional FOR ALL|SELECT|INSERT|UPDATE|DELETE
	var command *expr.CreatePolicyCommand
	if p.ParseKeyword("FOR") {
		if p.ParseKeyword("ALL") {
			cmd := expr.CreatePolicyCommandAll
			command = &cmd
		} else if p.ParseKeyword("SELECT") {
			cmd := expr.CreatePolicyCommandSelect
			command = &cmd
		} else if p.ParseKeyword("INSERT") {
			cmd := expr.CreatePolicyCommandInsert
			command = &cmd
		} else if p.ParseKeyword("UPDATE") {
			cmd := expr.CreatePolicyCommandUpdate
			command = &cmd
		} else if p.ParseKeyword("DELETE") {
			cmd := expr.CreatePolicyCommandDelete
			command = &cmd
		} else {
			return nil, p.Expected("ALL, SELECT, INSERT, UPDATE, or DELETE after FOR", p.PeekToken())
		}
	}

	// Parse optional TO clause (role names)
	var to []*expr.Owner
	if p.ParseKeyword("TO") {
		for {
			owner, err := parseOwner(p)
			if err != nil {
				return nil, err
			}
			to = append(to, owner)
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
	}

	// Parse optional USING (expression)
	var usingExpr expr.Expr
	if p.ParseKeyword("USING") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		ep := NewExpressionParser(p)
		usingExpr, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	// Parse optional WITH CHECK (expression)
	var withCheckExpr expr.Expr
	if p.ParseKeywords([]string{"WITH", "CHECK"}) {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		ep := NewExpressionParser(p)
		withCheckExpr, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	return &statement.CreatePolicy{
		OrReplace:  orReplace,
		Name:       name,
		TableName:  tableName,
		PolicyType: policyType,
		Command:    command,
		To:         to,
		Using:      usingExpr,
		WithCheck:  withCheckExpr,
	}, nil
}

// parseOwner parses an owner specification (CURRENT_USER, CURRENT_ROLE, SESSION_USER, or identifier)
// Reference: src/parser/mod.rs:6795
func parseOwner(p *Parser) (*expr.Owner, error) {
	if p.ParseKeyword("CURRENT_USER") {
		return &expr.Owner{Kind: expr.OwnerKindCurrentUser}, nil
	}
	if p.ParseKeyword("CURRENT_ROLE") {
		return &expr.Owner{Kind: expr.OwnerKindCurrentRole}, nil
	}
	if p.ParseKeyword("SESSION_USER") {
		return &expr.Owner{Kind: expr.OwnerKindSessionUser}, nil
	}

	// Otherwise, parse as identifier
	ident, err := p.ParseIdentifier()
	if err != nil {
		return nil, p.Expected("CURRENT_USER, CURRENT_ROLE, SESSION_USER, or identifier", p.PeekToken())
	}
	return &expr.Owner{Kind: expr.OwnerKindIdent, Ident: ident}, nil
}

func parseCreateFunction(p *Parser, orReplace, temporary bool) (ast.Statement, error) {
	// Consume FUNCTION keyword
	if _, err := p.ExpectKeyword("FUNCTION"); err != nil {
		return nil, err
	}

	// Parse function name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse function arguments: (arg1 TYPE, arg2 TYPE, ...)
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	var args []*expr.OperateFunctionArg
	if _, ok := p.PeekToken().Token.(token.TokenRParen); !ok {
		// Parse comma-separated arguments
		for {
			arg, err := parseFunctionArg(p)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)

			if p.ConsumeToken(token.TokenComma{}) {
				continue
			}
			break
		}
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// Parse optional RETURNS clause
	var returnType *expr.FunctionReturnType
	if p.PeekKeyword("RETURNS") {
		p.NextToken()
		returnType, err = parseFunctionReturnType(p)
		if err != nil {
			return nil, err
		}
	}

	// Parse function attributes (LANGUAGE, AS, IMMUTABLE, etc.)
	var language *ast.Ident
	var behavior *expr.FunctionBehavior
	var calledOnNull *expr.FunctionCalledOnNull
	var parallel *expr.FunctionParallel
	var security *expr.FunctionSecurity
	var body *expr.CreateFunctionBody
	var setParams []*expr.FunctionDefinitionSetParam

	for {
		if p.PeekKeyword("LANGUAGE") {
			p.NextToken()
			lang, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			language = lang
		} else if p.PeekKeyword("AS") {
			p.NextToken()
			// Parse function body as string literal or dollar-quoted string
			tok := p.PeekToken()
			bodyStr, err := p.ParseStringLiteral()
			if err != nil {
				return nil, err
			}
			// Check if it was a dollar-quoted string
			_, isDollarQuoted := tok.Token.(token.TokenDollarQuotedString)
			body = &expr.CreateFunctionBody{Value: bodyStr, IsDollarQuoted: isDollarQuoted}
		} else if p.PeekKeyword("IMMUTABLE") {
			p.NextToken()
			b := expr.FunctionBehaviorImmutable
			behavior = &b
		} else if p.PeekKeyword("STABLE") {
			p.NextToken()
			b := expr.FunctionBehaviorStable
			behavior = &b
		} else if p.PeekKeyword("VOLATILE") {
			p.NextToken()
			b := expr.FunctionBehaviorVolatile
			behavior = &b
		} else if p.ParseKeywords([]string{"CALLED", "ON", "NULL", "INPUT"}) {
			c := expr.FunctionCalledOnNullCalledOnNullInput
			calledOnNull = &c
		} else if p.ParseKeywords([]string{"RETURNS", "NULL", "ON", "NULL", "INPUT"}) {
			c := expr.FunctionCalledOnNullReturnsNullOnNullInput
			calledOnNull = &c
		} else if p.PeekKeyword("STRICT") {
			p.NextToken()
			c := expr.FunctionCalledOnNullStrict
			calledOnNull = &c
		} else if p.PeekKeyword("PARALLEL") {
			p.NextToken()
			if p.PeekKeyword("UNSAFE") {
				p.NextToken()
				par := expr.FunctionParallelUnsafe
				parallel = &par
			} else if p.PeekKeyword("RESTRICTED") {
				p.NextToken()
				par := expr.FunctionParallelRestricted
				parallel = &par
			} else if p.PeekKeyword("SAFE") {
				p.NextToken()
				par := expr.FunctionParallelSafe
				parallel = &par
			}
		} else if p.ParseKeywords([]string{"SECURITY", "DEFINER"}) {
			s := expr.FunctionSecurityDefiner
			security = &s
		} else if p.ParseKeywords([]string{"SECURITY", "INVOKER"}) {
			s := expr.FunctionSecurityInvoker
			security = &s
		} else if p.PeekKeyword("SET") {
			p.NextToken()
			// Parse SET param_name = value or SET param_name FROM CURRENT
			paramName, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			var paramValue expr.FunctionSetValue
			if p.ParseKeywords([]string{"FROM", "CURRENT"}) {
				paramValue = expr.FunctionSetValue{Kind: expr.FunctionSetValueFromCurrent}
			} else {
				// Parse = or TO followed by values
				if !p.ConsumeToken(token.TokenEq{}) && !p.PeekKeyword("TO") {
					return nil, fmt.Errorf("expected = or TO after SET parameter name")
				}
				if p.PeekKeyword("TO") {
					p.NextToken()
				}
				// For simplicity, parse a single expression value
				exprParser := NewExpressionParser(p)
				value, err := exprParser.ParseExpr()
				if err != nil {
					return nil, err
				}
				paramValue = expr.FunctionSetValue{Kind: expr.FunctionSetValueExpr, Expr: value}
			}
			setParams = append(setParams, &expr.FunctionDefinitionSetParam{
				Name:  paramName,
				Value: paramValue,
			})
		} else if p.PeekKeyword("RETURN") {
			p.NextToken()
			exprParser := NewExpressionParser(p)
			retExpr, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			body = &expr.CreateFunctionBody{ReturnExpr: retExpr}
		} else {
			break
		}
	}

	return &statement.CreateFunction{
		OrReplace:    orReplace,
		Temporary:    temporary,
		Name:         name,
		Args:         args,
		ReturnType:   returnType,
		Language:     language,
		Behavior:     behavior,
		CalledOnNull: calledOnNull,
		Parallel:     parallel,
		Security:     security,
		Body:         body,
		Set:          setParams,
	}, nil
}

// parseFunctionArg parses a function argument like "name TYPE" or "IN name TYPE"
// Reference: src/parser/mod.rs:5972
func parseFunctionArg(p *Parser) (*expr.OperateFunctionArg, error) {
	// Check for IN/OUT/INOUT mode
	var mode *expr.ArgMode
	if p.PeekKeyword("IN") {
		p.NextToken()
		m := expr.ArgModeIn
		mode = &m
	} else if p.PeekKeyword("OUT") {
		p.NextToken()
		m := expr.ArgModeOut
		mode = &m
	} else if p.PeekKeyword("INOUT") {
		p.NextToken()
		m := expr.ArgModeInOut
		mode = &m
	}

	// Try to parse the first token - it could be either:
	// 1. A parameter name followed by a data type (e.g., "str1 VARCHAR")
	// 2. Just a data type (e.g., "INTEGER")

	// Save current position for potential backtracking
	savedIdx := p.index

	// Try to parse as data type first
	firstDataType, err := p.ParseDataType()
	if err != nil {
		// Failed to parse as data type, try as identifier then data type
		p.index = savedIdx
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		dataType, err := p.ParseDataType()
		if err != nil {
			return nil, err
		}

		// Check for DEFAULT or = value
		var defaultExpr expr.Expr
		var defaultOp string
		if p.PeekKeyword("DEFAULT") {
			p.NextToken()
			defaultOp = "DEFAULT"
			exprParser := NewExpressionParser(p)
			defaultExpr, err = exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
		} else if p.ConsumeToken(token.TokenEq{}) {
			defaultOp = "="
			exprParser := NewExpressionParser(p)
			defaultExpr, err = exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
		}

		return &expr.OperateFunctionArg{
			Mode:        mode,
			Name:        name,
			DataType:    dataType,
			DefaultExpr: defaultExpr,
			DefaultOp:   defaultOp,
		}, nil
	}

	// We successfully parsed a data type. Now check if the next token could also be
	// a data type keyword (which would mean the first token was actually a parameter name).
	// This is the Rust approach: try to parse another data type, and if it succeeds,
	// treat the first as a name.

	// For simplicity, we check if next token looks like a type keyword
	// Common SQL type keywords: VARCHAR, INTEGER, INT, TEXT, BOOLEAN, etc.
	// If next token is one of these, the first token was likely a name.

	// Save position again
	secondIdx := p.index

	// Try to peek at the next token to see if it looks like a type
	nextTok := p.PeekToken()
	if word, ok := nextTok.Token.(token.TokenWord); ok {
		// Check if next token is a common SQL type
		typeKeywords := map[string]bool{
			"VARCHAR": true, "CHAR": true, "TEXT": true, "INTEGER": true,
			"INT": true, "BIGINT": true, "SMALLINT": true, "BOOLEAN": true,
			"BOOL": true, "REAL": true, "DOUBLE": true, "FLOAT": true,
			"DECIMAL": true, "NUMERIC": true, "DATE": true, "TIME": true,
			"TIMESTAMP": true, "INTERVAL": true, "ARRAY": true, "JSON": true,
			"JSONB": true, "BYTEA": true, "UUID": true, "SERIAL": true,
			"BIGSERIAL": true, "SMALLSERIAL": true, "MONEY": true,
		}
		upperWord := strings.ToUpper(word.Word.Value)
		if typeKeywords[upperWord] {
			// The next token is a type keyword, so first token was a name
			// We need to parse the second data type
			p.index = secondIdx
			secondDataType, err := p.ParseDataType()
			if err != nil {
				// If we fail to parse second, just use first as data type
				p.index = secondIdx

				// Check for DEFAULT or = value
				var defaultExpr expr.Expr
				var defaultOp string
				if p.PeekKeyword("DEFAULT") {
					p.NextToken()
					defaultOp = "DEFAULT"
					exprParser := NewExpressionParser(p)
					defaultExpr, err = exprParser.ParseExpr()
					if err != nil {
						return nil, err
					}
				} else if p.ConsumeToken(token.TokenEq{}) {
					defaultOp = "="
					exprParser := NewExpressionParser(p)
					defaultExpr, err = exprParser.ParseExpr()
					if err != nil {
						return nil, err
					}
				}

				return &expr.OperateFunctionArg{
					Mode:        mode,
					DataType:    firstDataType,
					DefaultExpr: defaultExpr,
					DefaultOp:   defaultOp,
				}, nil
			}

			// Create identifier from first "data type" (which was actually a name)
			name := &ast.Ident{Value: firstDataType.(fmt.Stringer).String()}

			// Check for DEFAULT or = value
			var defaultExpr expr.Expr
			var defaultOp string
			if p.PeekKeyword("DEFAULT") {
				p.NextToken()
				defaultOp = "DEFAULT"
				exprParser := NewExpressionParser(p)
				defaultExpr, err = exprParser.ParseExpr()
				if err != nil {
					return nil, err
				}
			} else if p.ConsumeToken(token.TokenEq{}) {
				defaultOp = "="
				exprParser := NewExpressionParser(p)
				defaultExpr, err = exprParser.ParseExpr()
				if err != nil {
					return nil, err
				}
			}

			return &expr.OperateFunctionArg{
				Mode:        mode,
				Name:        name,
				DataType:    secondDataType,
				DefaultExpr: defaultExpr,
				DefaultOp:   defaultOp,
			}, nil
		}
	}

	// Next token is not a type keyword, so firstDataType is the actual data type
	// Check for DEFAULT or = value
	var defaultExpr expr.Expr
	var defaultOp string
	if p.PeekKeyword("DEFAULT") {
		p.NextToken()
		defaultOp = "DEFAULT"
		exprParser := NewExpressionParser(p)
		defaultExpr, err = exprParser.ParseExpr()
		if err != nil {
			return nil, err
		}
	} else if p.ConsumeToken(token.TokenEq{}) {
		defaultOp = "="
		exprParser := NewExpressionParser(p)
		defaultExpr, err = exprParser.ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	return &expr.OperateFunctionArg{
		Mode:        mode,
		DataType:    firstDataType,
		DefaultExpr: defaultExpr,
		DefaultOp:   defaultOp,
	}, nil
}

// parseFunctionReturnType parses a return type like "INTEGER" or "SETOF INTEGER"
func parseFunctionReturnType(p *Parser) (*expr.FunctionReturnType, error) {
	if p.PeekKeyword("SETOF") {
		p.NextToken()
		dataType, err := p.ParseDataType()
		if err != nil {
			return nil, err
		}
		return &expr.FunctionReturnType{Kind: expr.FunctionReturnTypeSetOf, DataType: dataType}, nil
	}

	dataType, err := p.ParseDataType()
	if err != nil {
		return nil, err
	}
	return &expr.FunctionReturnType{Kind: expr.FunctionReturnTypeDataType, DataType: dataType}, nil
}

func parseCreateVirtualTable(p *Parser) (ast.Statement, error) {
	// SQLite: CREATE VIRTUAL TABLE table_name USING module_name (args...)
	if _, err := p.ExpectKeyword("VIRTUAL"); err != nil {
		return nil, err
	}
	if _, err := p.ExpectKeyword("TABLE"); err != nil {
		return nil, err
	}

	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectKeyword("USING"); err != nil {
		return nil, err
	}

	moduleName, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Parse optional arguments
	var args []*ast.Ident
	if p.ConsumeToken(token.TokenLParen{}) {
		for {
			arg, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	return &statement.CreateVirtualTable{
		IfNotExists: ifNotExists,
		Name:        tableName,
		ModuleName:  moduleName,
		ModuleArgs:  args,
	}, nil
}

func parseCreateMacro(p *Parser) (ast.Statement, error) {
	// DuckDB: CREATE [OR REPLACE] MACRO name AS ...
	orReplace := p.ParseKeywords([]string{"OR", "REPLACE"})

	macroName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse optional parameters
	var params []*ast.Ident
	if p.ConsumeToken(token.TokenLParen{}) {
		for {
			param, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			params = append(params, param)
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	if _, err := p.ExpectKeyword("AS"); err != nil {
		return nil, err
	}

	// Parse the macro definition (query or expression)
	// Skip full query parsing for now
	return &statement.CreateMacro{
		OrReplace: orReplace,
		Name:      macroName,
	}, nil
}

func parseCreateSecret(p *Parser, orReplace, temporary bool) (ast.Statement, error) {
	// DuckDB: CREATE [OR REPLACE] [TEMPORARY] SECRET [IF NOT EXISTS] name TYPE type ...
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	secretName, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Parse TYPE
	var secretType *ast.Ident
	if p.ParseKeyword("TYPE") {
		secretType, err = p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
	}

	tempPtr := temporary
	return &statement.CreateSecret{
		OrReplace:   orReplace,
		Temporary:   &tempPtr,
		IfNotExists: ifNotExists,
		Name:        secretName,
		SecretType:  secretType,
	}, nil
}

func parseCreateConnector(p *Parser, orReplace bool) (ast.Statement, error) {
	// Check for IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse connector name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Parse optional TYPE 'datasource_type'
	var connectorType *string
	if p.ParseKeyword("TYPE") {
		tok := p.PeekToken()
		if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			p.AdvanceToken()
			connectorType = &str.Value
		} else {
			return nil, p.Expected("string literal after TYPE", tok)
		}
	}

	// Parse optional URL 'datasource_url'
	var url *string
	if p.ParseKeyword("URL") {
		tok := p.PeekToken()
		if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			p.AdvanceToken()
			url = &str.Value
		} else {
			return nil, p.Expected("string literal after URL", tok)
		}
	}

	// Parse optional COMMENT 'comment'
	var comment *string
	if p.ParseKeyword("COMMENT") {
		// Optional = sign
		p.ParseKeyword("=")
		tok := p.PeekToken()
		if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			p.AdvanceToken()
			comment = &str.Value
		} else {
			return nil, p.Expected("string literal after COMMENT", tok)
		}
	}

	// Parse optional WITH DCPROPERTIES (property_name=property_value, ...)
	var dcProperties []*expr.SqlOption
	if p.ParseKeywords([]string{"WITH", "DCPROPERTIES"}) {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		for {
			// Parse property name
			propName, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			// Expect =
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			// Parse property value (identifier or string)
			var propValue expr.Expr
			tok := p.PeekToken()
			if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
				p.AdvanceToken()
				propValue = &expr.ValueExpr{
					Value: str.Value,
				}
			} else {
				propVal, err := p.ParseIdentifier()
				if err != nil {
					return nil, err
				}
				propValue = &expr.Ident{Value: propVal.Value}
			}
			dcProperties = append(dcProperties, &expr.SqlOption{
				Name:  &expr.Ident{Value: propName.Value},
				Value: propValue,
			})
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	return &statement.CreateConnector{
		IfNotExists:      ifNotExists,
		Name:             name,
		ConnectorType:    connectorType,
		URL:              url,
		Comment:          comment,
		WithDCProperties: dcProperties,
	}, nil
}

func parseCreateOperator(p *Parser) (ast.Statement, error) {
	// Consume OPERATOR keyword
	if _, err := p.ExpectKeyword("OPERATOR"); err != nil {
		return nil, err
	}

	// Check for FAMILY or CLASS
	if p.PeekKeyword("FAMILY") {
		return parseCreateOperatorFamily(p)
	}
	if p.PeekKeyword("CLASS") {
		return parseCreateOperatorClass(p)
	}

	// Parse operator name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Expect opening parenthesis
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	createOp := &statement.CreateOperator{
		Name: name,
	}

	// Parse operator parameters
	for {
		done := false
		switch {
		case p.PeekKeyword("FUNCTION"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			funcName, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			createOp.Function = funcName
			createOp.IsProcedure = false

		case p.PeekKeyword("PROCEDURE"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			funcName, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			createOp.Function = funcName
			createOp.IsProcedure = true

		case p.PeekKeyword("LEFTARG"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			// Parse data type
			dataType, err := p.ParseDataType()
			if err != nil {
				return nil, err
			}
			createOp.LeftArg = dataType

		case p.PeekKeyword("RIGHTARG"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			// Parse data type
			dataType, err := p.ParseDataType()
			if err != nil {
				return nil, err
			}
			createOp.RightArg = dataType

		case p.PeekKeyword("HASHES"):
			p.NextToken()
			createOp.Options = append(createOp.Options, &expr.OperatorOption{
				Kind: expr.OperatorOptionKindHashes,
			})

		case p.PeekKeyword("MERGES"):
			p.NextToken()
			createOp.Options = append(createOp.Options, &expr.OperatorOption{
				Kind: expr.OperatorOptionKindMerges,
			})

		case p.PeekKeyword("COMMUTATOR"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			var opName *ast.ObjectName
			if p.PeekKeyword("OPERATOR") {
				p.NextToken()
				if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
					return nil, err
				}
				opName, err = p.ParseObjectName()
				if err != nil {
					return nil, err
				}
				if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
					return nil, err
				}
			} else {
				opName, err = p.ParseObjectName()
				if err != nil {
					return nil, err
				}
			}
			createOp.Options = append(createOp.Options, &expr.OperatorOption{
				Kind: expr.OperatorOptionKindCommutator,
				Name: opName,
			})

		case p.PeekKeyword("NEGATOR"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			var opName *ast.ObjectName
			if p.PeekKeyword("OPERATOR") {
				p.NextToken()
				if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
					return nil, err
				}
				opName, err = p.ParseObjectName()
				if err != nil {
					return nil, err
				}
				if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
					return nil, err
				}
			} else {
				opName, err = p.ParseObjectName()
				if err != nil {
					return nil, err
				}
			}
			createOp.Options = append(createOp.Options, &expr.OperatorOption{
				Kind: expr.OperatorOptionKindNegator,
				Name: opName,
			})

		case p.PeekKeyword("RESTRICT"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			funcName, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			createOp.Options = append(createOp.Options, &expr.OperatorOption{
				Kind: expr.OperatorOptionKindRestrict,
				Name: funcName,
			})

		case p.PeekKeyword("JOIN"):
			p.NextToken()
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			funcName, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			createOp.Options = append(createOp.Options, &expr.OperatorOption{
				Kind: expr.OperatorOptionKindJoin,
				Name: funcName,
			})

		default:
			done = true
		}

		if done {
			break
		}

		// Check for comma separator
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	// Expect closing parenthesis
	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// Validate that FUNCTION was specified
	if createOp.Function == nil {
		return nil, fmt.Errorf("CREATE OPERATOR requires FUNCTION parameter")
	}

	return createOp, nil
}

func parseCreateOperatorFamily(p *Parser) (ast.Statement, error) {
	// Consume FAMILY keyword
	if _, err := p.ExpectKeyword("FAMILY"); err != nil {
		return nil, err
	}

	// Parse family name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Expect USING keyword
	if _, err := p.ExpectKeyword("USING"); err != nil {
		return nil, err
	}

	// Parse index method
	indexMethod, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	return &statement.CreateOperatorFamily{
		Name:        name,
		IndexMethod: indexMethod,
	}, nil
}

func parseCreateOperatorClass(p *Parser) (ast.Statement, error) {
	// Consume CLASS keyword
	if _, err := p.ExpectKeyword("CLASS"); err != nil {
		return nil, err
	}

	// Parse class name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	createOpClass := &statement.CreateOperatorClass{
		Name: name,
	}

	// Check for DEFAULT
	if p.PeekKeyword("DEFAULT") {
		p.NextToken()
		createOpClass.IsDefault = true
	}

	// Expect FOR TYPE keywords
	if err := p.ExpectKeywords([]string{"FOR", "TYPE"}); err != nil {
		return nil, err
	}

	// Parse data type
	dataType, err := p.ParseDataType()
	if err != nil {
		return nil, err
	}
	createOpClass.DataType = dataType

	// Expect USING keyword
	if _, err := p.ExpectKeyword("USING"); err != nil {
		return nil, err
	}

	// Parse index method
	indexMethod, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	createOpClass.IndexMethod = indexMethod

	// Check for FAMILY clause
	if p.PeekKeyword("FAMILY") {
		p.NextToken()
		family, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		createOpClass.Family = family
	}

	// Expect AS keyword
	if _, err := p.ExpectKeyword("AS"); err != nil {
		return nil, err
	}

	// Parse operator class items
	for {
		item, err := parseOperatorClassItem(p)
		if err != nil {
			return nil, err
		}
		if item == nil {
			break
		}
		createOpClass.Items = append(createOpClass.Items, item)

		// Check for comma separator
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return createOpClass, nil
}

func parseOperatorClassItem(p *Parser) (*expr.OperatorClassItem, error) {
	if p.PeekKeyword("OPERATOR") {
		p.NextToken()

		// Parse strategy number
		stratNum, err := parseLiteralUint(p)
		if err != nil {
			return nil, err
		}

		// Parse operator name
		opName, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}

		item := &expr.OperatorClassItem{
			IsOperator:     true,
			StrategyNumber: stratNum,
			OperatorName:   opName,
		}

		// Check for optional argument types
		if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
			p.NextToken()
			leftType, err := p.ParseDataType()
			if err != nil {
				return nil, err
			}
			if _, err := p.ExpectToken(token.TokenComma{}); err != nil {
				return nil, err
			}
			rightType, err := p.ParseDataType()
			if err != nil {
				return nil, err
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
			item.OpTypes = &expr.OperatorArgTypes{
				Left:  leftType,
				Right: rightType,
			}
		}

		// Check for optional purpose (FOR SEARCH or FOR ORDER BY)
		if p.PeekKeyword("FOR") {
			p.NextToken()
			if p.PeekKeyword("SEARCH") {
				p.NextToken()
				item.Purpose = &expr.OperatorPurposeWithFamily{
					Purpose: expr.OperatorPurposeForSearch,
				}
			} else if p.PeekKeyword("ORDER") {
				p.NextToken()
				if _, err := p.ExpectKeyword("BY"); err != nil {
					return nil, err
				}
				sortFamily, err := p.ParseObjectName()
				if err != nil {
					return nil, err
				}
				item.Purpose = &expr.OperatorPurposeWithFamily{
					Purpose:    expr.OperatorPurposeForOrderBy,
					SortFamily: sortFamily,
				}
			}
		}

		return item, nil
	}

	if p.PeekKeyword("FUNCTION") {
		p.NextToken()

		// Parse support number
		supportNum, err := parseLiteralUint(p)
		if err != nil {
			return nil, err
		}

		// Parse function name
		funcName, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}

		item := &expr.OperatorClassItem{
			IsFunction:    true,
			SupportNumber: supportNum,
			FunctionName:  funcName,
		}

		// Parse argument types
		if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
			p.NextToken()
			if _, ok := p.PeekToken().Token.(token.TokenRParen); !ok {
				for {
					argType, err := p.ParseDataType()
					if err != nil {
						return nil, err
					}
					item.ArgumentTypes = append(item.ArgumentTypes, argType)
					if !p.ConsumeToken(token.TokenComma{}) {
						break
					}
				}
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		}

		return item, nil
	}

	if p.PeekKeyword("STORAGE") {
		p.NextToken()
		storageType, err := p.ParseDataType()
		if err != nil {
			return nil, err
		}
		return &expr.OperatorClassItem{
			IsStorage:   true,
			StorageType: storageType,
		}, nil
	}

	// No more items
	return nil, nil
}

func parseLiteralUint(p *Parser) (uint64, error) {
	tok := p.PeekToken()
	if numTok, ok := tok.Token.(token.TokenNumber); ok {
		var val uint64
		_, err := fmt.Sscanf(numTok.Value, "%d", &val)
		if err != nil {
			return 0, fmt.Errorf("expected unsigned integer, got %s", numTok.Value)
		}
		p.NextToken()
		return val, nil
	}
	// Also try parsing a word that represents a number
	if wordTok, ok := tok.Token.(token.TokenWord); ok {
		var val uint64
		_, err := fmt.Sscanf(wordTok.Word.Value, "%d", &val)
		if err == nil {
			p.NextToken()
			return val, nil
		}
	}
	return 0, fmt.Errorf("expected unsigned integer, got %v", tok)
}

func parseCreateUser(p *Parser, orReplace bool) (ast.Statement, error) {
	// Parse optional IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse user name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Parse optional WITH options (simplified - just look for key=value patterns)
	var options []*expr.SqlOption
	for {
		if p.PeekKeyword("WITH") || p.PeekKeyword("TAG") {
			break
		}
		// Try to parse key=value option
		key, err := p.ParseIdentifier()
		if err != nil {
			p.PrevToken()
			break
		}
		if !p.ConsumeToken(token.TokenEq{}) {
			p.PrevToken()
			break
		}
		value, err := NewExpressionParser(p).ParseExpr()
		if err != nil {
			p.PrevToken()
			p.PrevToken()
			break
		}
		// Convert ast.Ident to expr.Ident
		exprKey := &expr.Ident{
			SpanVal:    key.Span(),
			Value:      key.Value,
			QuoteStyle: key.QuoteStyle,
		}
		options = append(options, &expr.SqlOption{
			Name:  exprKey,
			Value: value,
		})
	}

	// Parse optional WITH TAG
	if p.ParseKeyword("WITH") {
		p.ParseKeyword("TAG")
	}

	return &statement.CreateUser{
		IfNotExists: ifNotExists,
		Name:        name,
		Options:     options,
	}, nil
}

// parseCreateProcedure parses CREATE PROCEDURE
// Reference: src/parser/mod.rs:19319 parse_create_procedure
func parseCreateProcedure(p *Parser, orAlter bool) (ast.Statement, error) {
	// Consume PROCEDURE keyword
	if _, err := p.ExpectKeyword("PROCEDURE"); err != nil {
		return nil, err
	}

	// Parse procedure name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse optional parameters
	params, err := parseProcedureParams(p)
	if err != nil {
		return nil, err
	}

	// Parse optional LANGUAGE clause
	var language *ast.Ident
	if p.ParseKeyword("LANGUAGE") {
		lang, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		language = lang
	}

	// Expect AS keyword
	if _, err := p.ExpectKeyword("AS"); err != nil {
		return nil, err
	}

	// Parse the procedure body
	// For now, just skip tokens until we hit END
	// Full implementation would parse actual procedure statements
	for {
		if p.PeekKeyword("END") {
			p.NextToken() // consume END
			break
		}
		p.NextToken()
	}

	return &statement.CreateProcedure{
		OrAlter:  orAlter,
		Name:     name,
		Params:   params,
		Language: language,
		Body:     &expr.ConditionalStatements{}, // TODO: properly parse body
	}, nil
}

// parseProcedureParams parses procedure parameters: (name TYPE, ...)
func parseProcedureParams(p *Parser) ([]*expr.ProcedureParam, error) {
	// Check for opening paren
	if !p.ConsumeToken(token.TokenLParen{}) {
		// No parameters
		return nil, nil
	}

	// Check for empty params: ()
	if p.ConsumeToken(token.TokenRParen{}) {
		return []*expr.ProcedureParam{}, nil
	}

	var params []*expr.ProcedureParam

	for {
		// Parse parameter
		param, err := parseProcedureParam(p)
		if err != nil {
			return nil, err
		}
		params = append(params, param)

		// Check for comma or closing paren
		if p.ConsumeToken(token.TokenComma{}) {
			// Continue to next parameter
			// Check for trailing comma with closing paren
			if p.ConsumeToken(token.TokenRParen{}) {
				break
			}
		} else if p.ConsumeToken(token.TokenRParen{}) {
			break
		} else {
			return nil, fmt.Errorf("expected ',' or ')' after parameter definition")
		}
	}

	return params, nil
}

// parseProcedureParam parses a single procedure parameter
func parseProcedureParam(p *Parser) (*expr.ProcedureParam, error) {
	spanStart := p.GetCurrentToken().Span

	// Check for parameter mode: IN, OUT, INOUT
	var mode *expr.ArgMode
	if p.PeekKeyword("IN") {
		p.NextToken()
		// Check for INOUT after IN
		if p.PeekKeyword("OUT") {
			p.NextToken()
			m := expr.ArgModeInOut
			mode = &m
		} else {
			m := expr.ArgModeIn
			mode = &m
		}
	} else if p.PeekKeyword("OUT") {
		p.NextToken()
		m := expr.ArgModeOut
		mode = &m
	}

	// Parse parameter name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// Convert to expr.Ident
	exprName := &expr.Ident{
		SpanVal:    name.Span(),
		Value:      name.Value,
		QuoteStyle: name.QuoteStyle,
	}

	// Parse data type
	dtype, err := p.ParseDataType()
	if err != nil {
		return nil, err
	}

	// Check for default value: = expression
	var defaultVal expr.Expr
	if p.ConsumeToken(token.TokenEq{}) {
		defaultVal, err = NewExpressionParser(p).ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	return &expr.ProcedureParam{
		SpanVal:  token.Span{Start: spanStart.Start, End: p.GetCurrentToken().Span.End},
		Name:     exprName,
		DataType: dtype,
		Mode:     mode,
		Default:  defaultVal,
	}, nil
}

// parseHiveDistributionStyle parses Hive DISTRIBUTED BY / SORTED BY clauses
// Reference: src/parser/mod.rs:parse_hive_distribution
func parseHiveDistributionStyle(p *Parser) *expr.HiveDistributionStyle {
	if !p.ParseKeyword("DISTRIBUTED") {
		return nil
	}
	if !p.ParseKeyword("BY") {
		return nil
	}

	ep := NewExpressionParser(p)
	var columns []*expr.ColumnDef
	for {
		col, err := ep.ParseExpr()
		if err != nil || col == nil {
			break
		}
		// Convert expr to ColumnDef if needed
		columns = append(columns, &expr.ColumnDef{})
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return &expr.HiveDistributionStyle{
		Type: expr.HiveDistributionPARTITIONED,
	}
}

// parseClusteredByClause parses Hive CLUSTERED BY clause
// Reference: src/parser/mod.rs:parse_optional_clustered_by
func parseClusteredByClause(p *Parser) *expr.ClusteredBy {
	if !p.ParseKeyword("CLUSTERED") {
		return nil
	}
	if !p.ParseKeyword("BY") {
		return nil
	}

	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil
	}

	var columns []*ast.Ident
	for {
		ident, err := p.ParseIdentifier()
		if err != nil {
			break
		}
		columns = append(columns, ident)
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}
	p.ExpectToken(token.TokenRParen{})

	return &expr.ClusteredBy{
		Columns: columns,
	}
}

// parseHiveFormatClause parses Hive row format and storage format clauses
// Reference: src/parser/mod.rs:parse_hive_formats
func parseHiveFormatClause(p *Parser) *expr.HiveFormat {
	formats := &expr.HiveFormat{}

	// Parse ROW FORMAT
	if p.ParseKeywords([]string{"ROW", "FORMAT"}) {
		if p.ParseKeyword("DELIMITED") {
			formats.RowFormat = &expr.HiveRowFormat{Delimited: true}
		} else if p.ParseKeyword("SERDE") {
			serde, _ := p.ParseStringLiteral()
			formats.RowFormat = &expr.HiveRowFormat{SerdeClass: &serde}
		}
	}

	// Parse STORED AS
	if p.ParseKeywords([]string{"STORED", "AS"}) {
		if p.ParseKeyword("INPUTFORMAT") {
			inputFormat, _ := p.ParseStringLiteral()
			storage := &expr.HiveIOFormat{InputFormat: inputFormat}
			if p.ParseKeyword("OUTPUTFORMAT") {
				outputFormat, _ := p.ParseStringLiteral()
				storage.OutputFormat = outputFormat
			}
			formats.Storage = storage
		} else {
			format, _ := p.ParseIdentifier()
			if format != nil {
				formats.Storage = &expr.HiveIOFormat{
					InputFormat:  format.Value,
					OutputFormat: format.Value,
				}
			}
		}
	}

	// Parse LOCATION
	if p.ParseKeyword("LOCATION") {
		location, _ := p.ParseStringLiteral()
		formats.Location = &location
	}

	if formats.RowFormat != nil || formats.Storage != nil || formats.Location != nil {
		return formats
	}
	return nil
}

// parseOrderByClauseForCreateTable parses ORDER BY clause for CREATE TABLE (ClickHouse)
func parseOrderByClauseForCreateTable(p *Parser) *expr.OneOrManyWithParens {
	if _, err := p.ExpectToken(token.TokenLParen{}); err == nil {
		ep := NewExpressionParser(p)
		var items []expr.Expr
		for {
			item, err := ep.ParseExpr()
			if err != nil || item == nil {
				break
			}
			items = append(items, item)
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		p.ExpectToken(token.TokenRParen{})
		return &expr.OneOrManyWithParens{Items: items}
	}

	// Single expression without parens
	ep := NewExpressionParser(p)
	item, _ := ep.ParseExpr()
	if item != nil {
		return &expr.OneOrManyWithParens{Items: []expr.Expr{item}}
	}
	return nil
}

// parseOnCommitClause parses ON COMMIT { PRESERVE ROWS | DELETE ROWS | DROP } clause
// Reference: src/parser/mod.rs:parse_create_table_on_commit
func parseOnCommitClause(p *Parser) *expr.OnCommit {
	if p.ParseKeyword("PRESERVE") {
		p.ParseKeyword("ROWS")
		commitType := expr.OnCommitPreserveRows
		return &commitType
	} else if p.ParseKeyword("DELETE") {
		p.ParseKeyword("ROWS")
		commitType := expr.OnCommitDeleteRows
		return &commitType
	} else if p.ParseKeyword("DROP") {
		commitType := expr.OnCommitDrop
		return &commitType
	}
	return nil
}

// parseDistStyle parses Redshift DISTSTYLE { ALL | EVEN | KEY | AUTO } clause
// Reference: src/parser/mod.rs:parse_dist_style
func parseDistStyle(p *Parser) *expr.DistStyle {
	if p.PeekKeyword("ALL") {
		p.AdvanceToken()
		return &expr.DistStyle{Style: "ALL"}
	} else if p.PeekKeyword("EVEN") {
		p.AdvanceToken()
		return &expr.DistStyle{Style: "EVEN"}
	} else if p.PeekKeyword("KEY") {
		p.AdvanceToken()
		return &expr.DistStyle{Style: "KEY"}
	} else if p.PeekKeyword("AUTO") {
		p.AdvanceToken()
		return &expr.DistStyle{Style: "AUTO"}
	}
	return nil
}

// parsePartitionForValues parses FOR VALUES clause for PostgreSQL PARTITION OF.
// Reference: src/parser/mod.rs:parse_partition_for_values
//
// Syntax: FOR VALUES { IN (expr, ...) | FROM (...) TO (...) | WITH (MODULUS n, REMAINDER r) } | DEFAULT
func parsePartitionForValues(p *Parser) *expr.ForValues {
	// Check for DEFAULT first
	if p.ParseKeyword("DEFAULT") {
		return &expr.ForValues{Kind: expr.ForValuesKindDefault}
	}

	// Must start with FOR VALUES
	if !p.ParseKeywords([]string{"FOR", "VALUES"}) {
		return nil
	}

	// FOR VALUES IN (expr, ...)
	if p.ParseKeyword("IN") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil
		}
		var values []expr.Expr
		ep := NewExpressionParser(p)
		for {
			if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
				break
			}
			val, err := ep.ParseExpr()
			if err != nil {
				break
			}
			if val != nil {
				values = append(values, val)
			}
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		p.ExpectToken(token.TokenRParen{})
		return &expr.ForValues{Kind: expr.ForValuesKindIn, Values: values}
	}

	// FOR VALUES FROM (...) TO (...)
	if p.ParseKeyword("FROM") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil
		}
		var fromValues []expr.PartitionBoundValue
		for {
			if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
				break
			}
			boundVal := parsePartitionBoundValue(p)
			if boundVal != nil {
				fromValues = append(fromValues, *boundVal)
			}
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		p.ExpectToken(token.TokenRParen{})

		// Expect TO
		if !p.ParseKeyword("TO") {
			return nil
		}

		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil
		}
		var toValues []expr.PartitionBoundValue
		for {
			if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
				break
			}
			boundVal := parsePartitionBoundValue(p)
			if boundVal != nil {
				toValues = append(toValues, *boundVal)
			}
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		p.ExpectToken(token.TokenRParen{})
		return &expr.ForValues{Kind: expr.ForValuesKindFrom, From: fromValues, To: toValues}
	}

	// FOR VALUES WITH (MODULUS n, REMAINDER r)
	if p.ParseKeyword("WITH") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil
		}
		// Expect MODULUS
		if !p.ParseKeyword("MODULUS") {
			return nil
		}
		modulusTok := p.NextToken()
		var modulus uint64
		if intTok, ok := modulusTok.Token.(token.TokenNumber); ok {
			fmt.Sscanf(intTok.Value, "%d", &modulus)
		}

		if !p.ConsumeToken(token.TokenComma{}) {
			return nil
		}

		// Expect REMAINDER
		if !p.ParseKeyword("REMAINDER") {
			return nil
		}
		remainderTok := p.NextToken()
		var remainder uint64
		if intTok, ok := remainderTok.Token.(token.TokenNumber); ok {
			fmt.Sscanf(intTok.Value, "%d", &remainder)
		}

		p.ExpectToken(token.TokenRParen{})
		return &expr.ForValues{Kind: expr.ForValuesKindWith, Modulus: modulus, Remainder: remainder}
	}

	return nil
}

// parsePartitionBoundValue parses a partition bound value (MINVALUE, MAXVALUE, or expression).
// Reference: src/parser/mod.rs:parse_partition_bound_value
func parsePartitionBoundValue(p *Parser) *expr.PartitionBoundValue {
	if p.ParseKeyword("MINVALUE") {
		return &expr.PartitionBoundValue{IsMinValue: true}
	}
	if p.ParseKeyword("MAXVALUE") {
		return &expr.PartitionBoundValue{IsMaxValue: true}
	}
	ep := NewExpressionParser(p)
	exprVal, err := ep.ParseExpr()
	if err != nil || exprVal == nil {
		return nil
	}
	return &expr.PartitionBoundValue{Expr: exprVal}
}

// parseCreateStage parses a CREATE STAGE statement (Snowflake)
// Reference: src/dialect/snowflake.rs:parse_create_stage
func parseCreateStage(p *Parser, orReplace bool, temporary bool, transient bool) (ast.Statement, error) {
	// Consume STAGE keyword
	if _, err := p.ExpectKeyword("STAGE"); err != nil {
		return nil, err
	}

	// Parse IF NOT EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})

	// Parse stage name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse stage parameters (URL, STORAGE_INTEGRATION, ENDPOINT, CREDENTIALS, ENCRYPTION)
	stageParams := &expr.StageParamsObject{}

	for {
		if p.ParseKeyword("URL") {
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			urlStr, err := p.ParseStringLiteral()
			if err != nil {
				return nil, err
			}
			stageParams.Url = &urlStr
		} else if p.ParseKeyword("STORAGE_INTEGRATION") {
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			siStr, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			siVal := siStr.String()
			stageParams.StorageIntegration = &siVal
		} else if p.ParseKeyword("ENDPOINT") {
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			endpointStr, err := p.ParseStringLiteral()
			if err != nil {
				return nil, err
			}
			stageParams.Endpoint = &endpointStr
		} else if p.ParseKeyword("CREDENTIALS") {
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
				return nil, err
			}
			creds, err := parseKeyValueOptions(p, token.TokenRParen{})
			if err != nil {
				return nil, err
			}
			stageParams.Credentials = creds
		} else if p.ParseKeyword("ENCRYPTION") {
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
				return nil, err
			}
			encryption, err := parseKeyValueOptions(p, token.TokenRParen{})
			if err != nil {
				return nil, err
			}
			stageParams.Encryption = encryption
		} else {
			break
		}
	}

	// Parse DIRECTORY TABLE options
	var directoryTableParams *expr.KeyValueOptions
	if p.ParseKeyword("DIRECTORY") {
		if p.ParseKeyword("ENABLE") {
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			// Parse boolean or string value
			tok := p.PeekTokenRef()
			var val string
			if word, ok := tok.Token.(token.TokenWord); ok {
				val = strings.ToUpper(word.Word.Value)
				p.AdvanceToken()
			} else if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
				val = str.Value
				p.AdvanceToken()
			}
			directoryTableParams = &expr.KeyValueOptions{
				Options: []*expr.KeyValueOption{
					{OptionName: "ENABLE", OptionValue: val, Kind: expr.KeyValueOptionKindSingle},
				},
			}
		}
	}

	// Parse FILE_FORMAT options
	var fileFormat *expr.KeyValueOptions
	if p.ParseKeyword("FILE_FORMAT") {
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return nil, err
		}
		// Can be a format name or (type = ...)
		if p.ConsumeToken(token.TokenLParen{}) {
			fmt, err := parseKeyValueOptions(p, token.TokenRParen{})
			if err != nil {
				return nil, err
			}
			fileFormat = fmt
		} else {
			// Single format name
			fmtName, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			fileFormat = &expr.KeyValueOptions{
				Options: []*expr.KeyValueOption{
					{OptionName: "FORMAT_NAME", OptionValue: fmtName.String(), Kind: expr.KeyValueOptionKindSingle},
				},
			}
		}
	}

	// Parse COPY_OPTIONS
	var copyOptions *expr.KeyValueOptions
	if p.ParseKeyword("COPY_OPTIONS") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		opts, err := parseKeyValueOptions(p, token.TokenRParen{})
		if err != nil {
			return nil, err
		}
		copyOptions = opts
	}

	// Parse COMMENT
	var comment *string
	if p.ParseKeyword("COMMENT") {
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return nil, err
		}
		commentStr, err := p.ParseStringLiteral()
		if err != nil {
			return nil, err
		}
		comment = &commentStr
	}

	return &statement.CreateStage{
		OrReplace:            orReplace,
		Temporary:            temporary,
		IfNotExists:          ifNotExists,
		Name:                 name,
		StageParams:          stageParams,
		DirectoryTableParams: directoryTableParams,
		FileFormat:           fileFormat,
		CopyOptions:          copyOptions,
		Comment:              comment,
	}, nil
}

// parseKeyValueOptions parses key=value options until a closing token is found
func parseKeyValueOptions(p *Parser, closeTok token.Token) (*expr.KeyValueOptions, error) {
	opts := &expr.KeyValueOptions{
		Options:   []*expr.KeyValueOption{},
		Delimiter: expr.KeyValueOptionsDelimiterSpace,
	}

	for {
		// Check for closing token at the start of each iteration
		if p.ConsumeToken(closeTok) {
			break
		}

		// Check for closing token without consuming (for peek)
		tok := p.PeekTokenRef()
		if reflect.TypeOf(tok.Token) == reflect.TypeOf(closeTok) {
			p.AdvanceToken() // consume and break
			break
		}

		// Parse key
		keyTok := p.PeekTokenRef()
		var key string
		if word, ok := keyTok.Token.(token.TokenWord); ok {
			key = word.Word.Value
			p.AdvanceToken()
		} else {
			return nil, fmt.Errorf("expected identifier in key-value option, found %s", keyTok.Token.String())
		}

		// Expect =
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return nil, err
		}

		// Parse value - can be string, identifier, or number
		valTok := p.PeekTokenRef()
		var val interface{}
		if str, ok := valTok.Token.(token.TokenSingleQuotedString); ok {
			val = str.Value
			p.AdvanceToken()
		} else if word, ok := valTok.Token.(token.TokenWord); ok {
			val = word.Word.Value
			p.AdvanceToken()
		} else if num, ok := valTok.Token.(token.TokenNumber); ok {
			val = num.Value
			p.AdvanceToken()
		} else if p.ConsumeToken(token.TokenLParen{}) {
			// Nested options
			nested, err := parseKeyValueOptions(p, token.TokenRParen{})
			if err != nil {
				return nil, err
			}
			opts.Options = append(opts.Options, &expr.KeyValueOption{
				OptionName:  key,
				OptionValue: nested,
				Kind:        expr.KeyValueOptionKindNested,
			})
			// After nested, continue to next option
			continue
		} else {
			return nil, fmt.Errorf("expected value in key-value option, found %s", valTok.Token.String())
		}

		opts.Options = append(opts.Options, &expr.KeyValueOption{
			OptionName:  key,
			OptionValue: val,
			Kind:        expr.KeyValueOptionKindSingle,
		})

		// Check for comma (consume it if present and continue)
		if p.ConsumeToken(token.TokenComma{}) {
			continue
		}
	}

	return opts, nil
}
