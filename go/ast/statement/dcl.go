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
	"strings"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/expr"
)

// ============================================================================
// GRANT
// ============================================================================

// Grant represents a GRANT statement
// Reference: src/ast/dcl.rs Grant
type Grant struct {
	BaseStatement
	// Privileges is the set of privileges being granted
	Privileges *Privileges
	// Objects is the set of objects on which privileges are granted
	Objects *GrantObjects
	// Grantees is the list of grantees
	Grantees []*Grantee
	// WithGrantOption indicates if WITH GRANT OPTION was specified
	WithGrantOption bool
	// AsGrantor is the optional AS clause for the grantor
	AsGrantor *ast.Ident
	// GrantedBy is the optional GRANTED BY clause
	GrantedBy *ast.Ident
	// CurrentGrants is the type of current grants (COPY/REVOKE)
	CurrentGrants *CurrentGrantsKind
}

func (g *Grant) statementNode() {}

func (g *Grant) String() string {
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

// ============================================================================
// REVOKE
// ============================================================================

// Revoke represents a REVOKE statement
// Reference: src/ast/dcl.rs Revoke
type Revoke struct {
	BaseStatement
	// Privileges is the set of privileges being revoked
	Privileges *Privileges
	// Objects is the set of objects on which privileges are revoked
	Objects *GrantObjects
	// Grantees is the list of grantees
	Grantees []*Grantee
	// GrantedBy is the optional GRANTED BY clause
	GrantedBy *ast.Ident
	// Cascade indicates if CASCADE was specified
	Cascade bool
	// Restrict indicates if RESTRICT was specified
	Restrict bool
	// CurrentGrants is the type of current grants
	CurrentGrants *CurrentGrantsKind
}

func (r *Revoke) statementNode() {}

func (r *Revoke) String() string {
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

// ============================================================================
// CASCADE OPTION
// ============================================================================

// CascadeOption represents CASCADE or RESTRICT options
type CascadeOption int

const (
	// CascadeNone - no cascade option specified
	CascadeNone CascadeOption = iota
	// Cascade - apply cascading action
	Cascade
	// Restrict - restrict the action
	Restrict
)

func (c CascadeOption) String() string {
	switch c {
	case Cascade:
		return "CASCADE"
	case Restrict:
		return "RESTRICT"
	default:
		return ""
	}
}

// ============================================================================
// DENY
// ============================================================================

// DenyStatement represents a DENY statement (MSSQL)
type DenyStatement struct {
	BaseStatement
	Privileges *Privileges
	Objects    *GrantObjects
	Grantees   []*Grantee
	GrantedBy  *ast.Ident
	Cascade    CascadeOption
}

func (d *DenyStatement) statementNode() {}

func (d *DenyStatement) String() string {
	var f strings.Builder
	f.WriteString("DENY ")
	f.WriteString(d.Privileges.String())
	f.WriteString(" ON ")
	f.WriteString(d.Objects.String())
	f.WriteString(" TO ")
	for i, grantee := range d.Grantees {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(grantee.String())
	}
	if d.Cascade == Cascade {
		f.WriteString(" CASCADE")
	}
	if d.GrantedBy != nil {
		f.WriteString(" AS ")
		f.WriteString(d.GrantedBy.String())
	}
	return f.String()
}

// ============================================================================
// CREATE USER
// ============================================================================

// CreateUser represents a CREATE USER statement
type CreateUser struct {
	BaseStatement
	IfNotExists  bool
	Name         *ast.Ident
	IdentifiedBy *ast.Ident
	Password     expr.Expr
	LoginName    *ast.Ident
	Options      []*expr.SqlOption
}

func (c *CreateUser) statementNode() {}

func (c *CreateUser) String() string {
	var f strings.Builder
	f.WriteString("CREATE USER ")
	if c.IfNotExists {
		f.WriteString("IF NOT EXISTS ")
	}
	f.WriteString(c.Name.String())
	return f.String()
}

// ============================================================================
// ALTER USER
// ============================================================================

// AlterUser represents an ALTER USER statement
type AlterUser struct {
	BaseStatement
	IfExists       bool
	Name           *ast.Ident
	RenameTo       *ast.Ident
	ResetPassword  bool
	RequireProfile *ast.Ident
	Profile        *ast.Ident
	Options        []*expr.SqlOption
}

func (a *AlterUser) statementNode() {}

func (a *AlterUser) String() string {
	var f strings.Builder
	f.WriteString("ALTER USER ")
	if a.IfExists {
		f.WriteString("IF EXISTS ")
	}
	f.WriteString(a.Name.String())
	return f.String()
}

// ============================================================================
// USE
// ============================================================================

// Use represents a USE statement
type Use struct {
	BaseStatement
	DbName *ast.Ident
}

func (u *Use) statementNode() {}

func (u *Use) String() string {
	var f strings.Builder
	f.WriteString("USE ")
	f.WriteString(u.DbName.String())
	return f.String()
}

// ============================================================================
// LOCK TABLES / UNLOCK TABLES (MySQL-specific)
// ============================================================================

// LockTablesStmt represents a MySQL LOCK TABLES statement
// See https://dev.mysql.com/doc/refman/8.0/en/lock-tables.html
type LockTablesStmt struct {
	BaseStatement
	Tables []*LockTable
}

func (l *LockTablesStmt) statementNode() {}

func (l *LockTablesStmt) String() string {
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

// UnlockTablesStmt represents a MySQL UNLOCK TABLES statement
type UnlockTablesStmt struct {
	BaseStatement
}

func (u *UnlockTablesStmt) statementNode() {}

func (u *UnlockTablesStmt) String() string {
	return "UNLOCK TABLES"
}

// LockTable represents a table to lock with its lock type
type LockTable struct {
	Table    *ast.Ident
	Alias    *ast.Ident
	LockType LockTableType
}

func (l *LockTable) String() string {
	var f strings.Builder
	f.WriteString(l.Table.String())
	if l.Alias != nil {
		f.WriteString(" AS ")
		f.WriteString(l.Alias.String())
	}
	f.WriteString(" ")
	f.WriteString(l.LockType.String())
	return f.String()
}

// LockTableType represents the type of lock for LOCK TABLES
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

// ============================================================================
// DCL Helper Types
// ============================================================================

// Privileges represents a list of privileges in a GRANT/REVOKE statement
// Reference: src/ast/mod.rs Privileges
type Privileges struct {
	// All indicates if ALL PRIVILEGES was specified
	All bool
	// WithPrivilegesKeyword indicates if the PRIVILEGES keyword was used after ALL
	WithPrivilegesKeyword bool
	// Actions contains specific privilege actions (if All is false)
	Actions []*Action
}

func (p *Privileges) String() string {
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

// GrantObjectType represents the type of objects being granted
type GrantObjectType int

const (
	// GrantObjectTypeTables - Tables
	GrantObjectTypeTables GrantObjectType = iota
	// GrantObjectTypeSequences - Sequences (ON SEQUENCE ...)
	GrantObjectTypeSequences
	// GrantObjectTypeDatabases - Databases (ON DATABASE ...)
	GrantObjectTypeDatabases
	// GrantObjectTypeSchemas - Schemas (ON SCHEMA ...)
	GrantObjectTypeSchemas
	// GrantObjectTypeViews - Views (ON VIEW ...)
	GrantObjectTypeViews
	// GrantObjectTypeAllSequencesInSchema - ALL SEQUENCES IN SCHEMA
	GrantObjectTypeAllSequencesInSchema
	// GrantObjectTypeAllTablesInSchema - ALL TABLES IN SCHEMA
	GrantObjectTypeAllTablesInSchema
	// GrantObjectTypeAllViewsInSchema - ALL VIEWS IN SCHEMA
	GrantObjectTypeAllViewsInSchema
	// GrantObjectTypeFutureTablesInSchema - FUTURE TABLES IN SCHEMA
	GrantObjectTypeFutureTablesInSchema
)

// GrantObjects represents the objects on which privileges are granted
// Reference: src/ast/mod.rs GrantObjects
type GrantObjects struct {
	ObjectType GrantObjectType
	// Tables is the list of table names (for GrantObjectTypeTables)
	Tables []*ast.ObjectName
	// Schemas is the list of schema names (for IN SCHEMA variants)
	Schemas []*ast.ObjectName
}

func (g *GrantObjects) String() string {
	var f strings.Builder
	switch g.ObjectType {
	case GrantObjectTypeTables:
		if len(g.Tables) > 0 {
			for i, table := range g.Tables {
				if i > 0 {
					f.WriteString(", ")
				}
				f.WriteString(table.String())
			}
		}
	case GrantObjectTypeSequences:
		f.WriteString("SEQUENCE ")
		for i, table := range g.Tables {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(table.String())
		}
	case GrantObjectTypeDatabases:
		f.WriteString("DATABASE ")
		for i, table := range g.Tables {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(table.String())
		}
	case GrantObjectTypeSchemas:
		f.WriteString("SCHEMA ")
		for i, table := range g.Tables {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(table.String())
		}
	case GrantObjectTypeViews:
		f.WriteString("VIEW ")
		for i, table := range g.Tables {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(table.String())
		}
	case GrantObjectTypeAllTablesInSchema:
		f.WriteString("ALL TABLES IN SCHEMA ")
		for i, schema := range g.Schemas {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(schema.String())
		}
	case GrantObjectTypeAllSequencesInSchema:
		f.WriteString("ALL SEQUENCES IN SCHEMA ")
		for i, schema := range g.Schemas {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(schema.String())
		}
	case GrantObjectTypeAllViewsInSchema:
		f.WriteString("ALL VIEWS IN SCHEMA ")
		for i, schema := range g.Schemas {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(schema.String())
		}
	}
	return f.String()
}

// GranteesType represents the type of grantee
// Reference: src/ast/mod.rs GranteesType
type GranteesType int

const (
	// GranteesTypeNone - No type specified
	GranteesTypeNone GranteesType = iota
	// GranteesTypeRole - ROLE
	GranteesTypeRole
	// GranteesTypeUser - USER
	GranteesTypeUser
	// GranteesTypeShare - SHARE
	GranteesTypeShare
	// GranteesTypeGroup - GROUP
	GranteesTypeGroup
	// GranteesTypePublic - PUBLIC
	GranteesTypePublic
	// GranteesTypeDatabaseRole - DATABASE ROLE
	GranteesTypeDatabaseRole
	// GranteesTypeApplication - APPLICATION
	GranteesTypeApplication
	// GranteesTypeApplicationRole - APPLICATION ROLE
	GranteesTypeApplicationRole
)

func (g GranteesType) String() string {
	switch g {
	case GranteesTypeRole:
		return "ROLE "
	case GranteesTypeUser:
		return "USER "
	case GranteesTypeShare:
		return "SHARE "
	case GranteesTypeGroup:
		return "GROUP "
	case GranteesTypePublic:
		return "PUBLIC"
	case GranteesTypeDatabaseRole:
		return "DATABASE ROLE "
	case GranteesTypeApplication:
		return "APPLICATION "
	case GranteesTypeApplicationRole:
		return "APPLICATION ROLE "
	default:
		return ""
	}
}

// GranteeName represents the name of a grantee
// Reference: src/ast/mod.rs GranteeName
type GranteeName struct {
	// User is the user part (for 'user'@'host' syntax)
	User *ast.Ident
	// Host is the host part (for 'user'@'host' syntax)
	Host *ast.Ident
	// ObjectName is a simple object name (if not using user@host)
	ObjectName *ast.ObjectName
}

func (g *GranteeName) String() string {
	if g.User != nil && g.Host != nil {
		return g.User.String() + "@" + g.Host.String()
	}
	if g.ObjectName != nil {
		return g.ObjectName.String()
	}
	return ""
}

// Grantee represents a grantee (user or role)
// Reference: src/ast/mod.rs Grantee
type Grantee struct {
	// GranteeType is the type of grantee (role, user, public, etc.)
	GranteeType GranteesType
	// Name is the name of the grantee (nil for PUBLIC)
	Name *GranteeName
}

func (g *Grantee) String() string {
	var f strings.Builder
	f.WriteString(g.GranteeType.String())
	if g.Name != nil {
		f.WriteString(g.Name.String())
	}
	return f.String()
}

// CurrentGrantsKind represents the type of current grants
type CurrentGrantsKind int

const (
	// CurrentGrantsNone - No current grants clause
	CurrentGrantsNone CurrentGrantsKind = iota
	// CurrentGrantsCopy - COPY CURRENT GRANTS
	CurrentGrantsCopy
	// CurrentGrantsRevoke - REVOKE CURRENT GRANTS
	CurrentGrantsRevoke
)
