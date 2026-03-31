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
type Grant struct {
	BaseStatement
	Privileges      *Privileges
	Objects         *GrantObjects
	Grantees        []*Grantee
	WithGrantOption bool
	GrantedBy       *Grantee
	CurrentGrants   *CurrentGrantsKind
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

	return f.String()
}

// ============================================================================
// REVOKE
// ============================================================================

// Revoke represents a REVOKE statement
type Revoke struct {
	BaseStatement
	Privileges    *Privileges
	Objects       *GrantObjects
	Grantees      []*Grantee
	GrantedBy     *Grantee
	Cascade       bool
	Restrict      bool
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

	return f.String()
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
type Privileges struct {
	All bool
	// TODO: Add specific privilege types
}

func (p *Privileges) String() string {
	if p.All {
		return "ALL PRIVILEGES"
	}
	return ""
}

// GrantObjects represents the objects on which privileges are granted
type GrantObjects struct {
	// TODO: Add object types
}

func (g *GrantObjects) String() string {
	return ""
}

// Grantee represents a grantee (user or role)
type Grantee struct {
	Name string
}

func (g *Grantee) String() string {
	return g.Name
}

// CurrentGrantsKind represents the type of current grants
type CurrentGrantsKind int

const (
	CurrentGrantsRole CurrentGrantsKind = iota
	CurrentGrantsSystem
)
