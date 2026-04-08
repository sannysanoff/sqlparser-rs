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
)

// ============================================================================
// Privilege Actions for GRANT/REVOKE
// ============================================================================

// Action represents a specific privilege action in a GRANT/REVOKE statement
// Reference: src/ast/mod.rs Action
type Action struct {
	// ActionType is the type of privilege (SELECT, INSERT, etc.)
	ActionType ActionType
	// RawKeyword is the original keyword as it appeared in the SQL (e.g., "EXEC" vs "EXECUTE")
	// If empty, the canonical form from ActionType.String() will be used
	RawKeyword string
	// Columns is an optional list of columns for column-level privileges
	Columns []*ast.Ident
	// Role is the role name for ROLE action type (GRANT ROLE name TO ...)
	Role *ast.ObjectName
	// CreateObjectType is the optional object type for CREATE action (e.g., SCHEMA, DATABASE, etc.)
	CreateObjectType string
}

func (a *Action) String() string {
	var f strings.Builder
	// Use RawKeyword if available, otherwise use canonical form
	if a.RawKeyword != "" {
		f.WriteString(a.RawKeyword)
	} else {
		f.WriteString(a.ActionType.String())
	}
	// For CREATE action, include the object type if present
	if a.ActionType == ActionTypeCreate && a.CreateObjectType != "" {
		f.WriteString(" ")
		f.WriteString(a.CreateObjectType)
	}
	// For ROLE action, include the role name
	if a.ActionType == ActionTypeRole && a.Role != nil {
		f.WriteString(" ")
		f.WriteString(a.Role.String())
	}
	if len(a.Columns) > 0 {
		f.WriteString(" (")
		for i, col := range a.Columns {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(col.String())
		}
		f.WriteString(")")
	}
	return f.String()
}

// ActionType represents the type of privilege
type ActionType int

const (
	// ActionTypeConnect - CONNECT
	ActionTypeConnect ActionType = iota
	// ActionTypeCreate - CREATE
	ActionTypeCreate
	// ActionTypeDelete - DELETE
	ActionTypeDelete
	// ActionTypeDrop - DROP
	ActionTypeDrop
	// ActionTypeExecute - EXECUTE
	ActionTypeExecute
	// ActionTypeInsert - INSERT
	ActionTypeInsert
	// ActionTypeReferences - REFERENCES
	ActionTypeReferences
	// ActionTypeSelect - SELECT
	ActionTypeSelect
	// ActionTypeTemporary - TEMPORARY
	ActionTypeTemporary
	// ActionTypeTrigger - TRIGGER
	ActionTypeTrigger
	// ActionTypeTruncate - TRUNCATE
	ActionTypeTruncate
	// ActionTypeUpdate - UPDATE
	ActionTypeUpdate
	// ActionTypeUsage - USAGE
	ActionTypeUsage
	// ActionTypeOwnership - OWNERSHIP (Snowflake)
	ActionTypeOwnership
	// ActionTypeRead - READ
	ActionTypeRead
	// ActionTypeWrite - WRITE
	ActionTypeWrite
	// ActionTypeOperate - OPERATE
	ActionTypeOperate
	// ActionTypeApply - APPLY
	ActionTypeApply
	// ActionTypeAudit - AUDIT
	ActionTypeAudit
	// ActionTypeFailover - FAILOVER
	ActionTypeFailover
	// ActionTypeReplicate - REPLICATE
	ActionTypeReplicate
	// ActionTypeRole - ROLE (for GRANT ROLE name TO ...)
	ActionTypeRole
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
	case ActionTypeOwnership:
		return "OWNERSHIP"
	case ActionTypeRead:
		return "READ"
	case ActionTypeWrite:
		return "WRITE"
	case ActionTypeOperate:
		return "OPERATE"
	case ActionTypeApply:
		return "APPLY"
	case ActionTypeAudit:
		return "AUDIT"
	case ActionTypeFailover:
		return "FAILOVER"
	case ActionTypeReplicate:
		return "REPLICATE"
	case ActionTypeRole:
		return "ROLE"
	default:
		return ""
	}
}

// ParseActionType parses a string into an ActionType
func ParseActionType(s string) (ActionType, bool) {
	switch strings.ToUpper(s) {
	case "CONNECT":
		return ActionTypeConnect, true
	case "CREATE":
		return ActionTypeCreate, true
	case "DELETE":
		return ActionTypeDelete, true
	case "DROP":
		return ActionTypeDrop, true
	case "EXECUTE", "EXEC":
		return ActionTypeExecute, true
	case "INSERT":
		return ActionTypeInsert, true
	case "REFERENCES":
		return ActionTypeReferences, true
	case "SELECT":
		return ActionTypeSelect, true
	case "TEMPORARY":
		return ActionTypeTemporary, true
	case "TRIGGER":
		return ActionTypeTrigger, true
	case "TRUNCATE":
		return ActionTypeTruncate, true
	case "UPDATE":
		return ActionTypeUpdate, true
	case "USAGE":
		return ActionTypeUsage, true
	case "OWNERSHIP":
		return ActionTypeOwnership, true
	case "READ":
		return ActionTypeRead, true
	case "WRITE":
		return ActionTypeWrite, true
	case "OPERATE":
		return ActionTypeOperate, true
	case "APPLY":
		return ActionTypeApply, true
	case "AUDIT":
		return ActionTypeAudit, true
	case "FAILOVER":
		return ActionTypeFailover, true
	case "REPLICATE":
		return ActionTypeReplicate, true
	case "ROLE":
		return ActionTypeRole, true
	default:
		return ActionTypeConnect, false
	}
}
