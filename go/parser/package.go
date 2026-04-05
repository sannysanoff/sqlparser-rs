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

// Package statements provides SQL statement parsing implementations.
// This is the Go port of the sqlparser-rs statement parsing modules.
//
// The following statement parsers are implemented:
//
// Query Statements:
//   - ParseQuery: SELECT with CTEs, ORDER BY, LIMIT
//   - ParseSelect: Core SELECT parsing
//   - ParseWith: CTE definitions
//
// DML Statements:
//   - ParseInsert: INSERT with VALUES or subquery
//   - ParseUpdate: UPDATE with SET and WHERE
//   - ParseDelete: DELETE with WHERE
//   - parseMerge: MERGE statement
//
// DDL Statements:
//   - parseCreate: CREATE TABLE, VIEW, INDEX, FUNCTION, ROLE, etc.
//   - parseAlter: ALTER TABLE, VIEW, INDEX, ROLE, USER, etc.
//   - ParseDrop: DROP statements (internal)
//   - ParseTruncate: TRUNCATE TABLE (internal)
//
// Transaction Statements:
//   - ParseStartTransaction: BEGIN, START TRANSACTION
//   - ParseCommit: COMMIT
//   - ParseRollback: ROLLBACK
//   - ParseSavepoint: SAVEPOINT
//   - ParseRelease: RELEASE
//
// Other Statements:
//   - ParseAnalyze: ANALYZE
//   - parseCopy: COPY (PostgreSQL)
//   - ParseExplain: EXPLAIN
//   - ParseSet: SET variables
//   - ParseShow: SHOW
//   - ParseGrant: GRANT
//   - ParseRevoke: REVOKE
//   - ParseDeclare: DECLARE (cursor)
//   - ParseFetch: FETCH
//   - ParseClose: CLOSE
//   - ParseUse: USE
//   - ParseCall: CALL
package parser
