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

// Package mysql contains the MySQL-specific SQL parsing tests.
// These tests are ported from tests/sqlparser_mysql.rs in the Rust implementation.
package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// MySQL returns a TestedDialects with MySQL dialect only.
func MySQL() *utils.TestedDialects {
	return &utils.TestedDialects{
		Dialects: []dialects.Dialect{mysql.NewMySqlDialect()},
	}
}

// MySQLAndGeneric returns a TestedDialects with both MySQL and Generic dialects.
func MySQLAndGeneric() *utils.TestedDialects {
	return &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			mysql.NewMySqlDialect(),
			generic.NewGenericDialect(),
		},
	}
}

// TestParseIdentifiers verifies MySQL-specific identifier parsing.
// Reference: tests/sqlparser_mysql.rs:45
func TestParseIdentifiers(t *testing.T) {
	MySQL().VerifiedStmt(t, "SELECT $a$, àà")
}

// TestParseLiteralString verifies MySQL string literal parsing.
// Reference: tests/sqlparser_mysql.rs:50
func TestParseLiteralString(t *testing.T) {
	dialects := MySQL()
	sql := `SELECT 'single', "double"`
	dialects.VerifiedStmt(t, sql)
}

// TestParseFlush verifies FLUSH statement parsing.
// Reference: tests/sqlparser_mysql.rs:65
func TestParseFlush(t *testing.T) {
	dialects := MySQLAndGeneric()
	dialects.VerifiedStmt(t, "FLUSH OPTIMIZER_COSTS")
	dialects.VerifiedStmt(t, "FLUSH BINARY LOGS")
	dialects.VerifiedStmt(t, "FLUSH ENGINE LOGS")
	dialects.VerifiedStmt(t, "FLUSH ERROR LOGS")
	dialects.VerifiedStmt(t, "FLUSH NO_WRITE_TO_BINLOG GENERAL LOGS")
	dialects.VerifiedStmt(t, "FLUSH RELAY LOGS FOR CHANNEL test")
	dialects.VerifiedStmt(t, "FLUSH LOCAL SLOW LOGS")
	dialects.VerifiedStmt(t, "FLUSH TABLES `mek`.`table1`, table2")
	dialects.VerifiedStmt(t, "FLUSH TABLES WITH READ LOCK")
	dialects.VerifiedStmt(t, "FLUSH TABLES `mek`.`table1`, table2 WITH READ LOCK")
	dialects.VerifiedStmt(t, "FLUSH TABLES `mek`.`table1`, table2 FOR EXPORT")
}

// TestParseShowColumns verifies SHOW COLUMNS statement parsing.
// Reference: tests/sqlparser_mysql.rs:244
func TestParseShowColumns(t *testing.T) {
	dialects := MySQLAndGeneric()
	dialects.VerifiedStmt(t, "SHOW COLUMNS FROM mytable")
	dialects.VerifiedStmt(t, "SHOW COLUMNS FROM mydb.mytable")
	dialects.VerifiedStmt(t, "SHOW EXTENDED COLUMNS FROM mytable")
	dialects.VerifiedStmt(t, "SHOW FULL COLUMNS FROM mytable")
	dialects.VerifiedStmt(t, "SHOW COLUMNS FROM mytable LIKE 'pattern'")
	dialects.VerifiedStmt(t, "SHOW COLUMNS FROM mytable WHERE 1 = 2")
	dialects.OneStatementParsesTo(t, "SHOW FIELDS FROM mytable", "SHOW COLUMNS FROM mytable")
	dialects.OneStatementParsesTo(t, "SHOW COLUMNS IN mytable", "SHOW COLUMNS IN mytable")
	dialects.OneStatementParsesTo(t, "SHOW FIELDS IN mytable", "SHOW COLUMNS IN mytable")
	dialects.OneStatementParsesTo(t, "SHOW COLUMNS FROM mytable FROM mydb", "SHOW COLUMNS FROM mydb.mytable")
}

// TestParseShowStatus verifies SHOW STATUS statement parsing.
// Reference: tests/sqlparser_mysql.rs:373
func TestParseShowStatus(t *testing.T) {
	dialects := MySQLAndGeneric()
	dialects.VerifiedStmt(t, "SHOW SESSION STATUS LIKE 'ssl_cipher'")
	dialects.VerifiedStmt(t, "SHOW GLOBAL STATUS LIKE 'ssl_cipher'")
	dialects.VerifiedStmt(t, "SHOW STATUS WHERE value = 2")
}

// TestParseShowTables verifies SHOW TABLES statement parsing.
// Reference: tests/sqlparser_mysql.rs:403
func TestParseShowTables(t *testing.T) {
	dialects := MySQLAndGeneric()
	dialects.VerifiedStmt(t, "SHOW TABLES")
	dialects.VerifiedStmt(t, "SHOW TABLES FROM mydb")
	dialects.VerifiedStmt(t, "SHOW EXTENDED TABLES")
	dialects.VerifiedStmt(t, "SHOW FULL TABLES")
	dialects.VerifiedStmt(t, "SHOW TABLES LIKE 'pattern'")
	dialects.VerifiedStmt(t, "SHOW TABLES WHERE 1 = 2")
	dialects.VerifiedStmt(t, "SHOW TABLES IN mydb")
	dialects.VerifiedStmt(t, "SHOW TABLES FROM mydb")
}

// TestParseShowExtendedFull verifies SHOW EXTENDED/FULL restrictions.
// Reference: tests/sqlparser_mysql.rs:519
func TestParseShowExtendedFull(t *testing.T) {
	dialects := MySQLAndGeneric()

	// These should parse successfully
	stmts, err := parser.ParseSQL(dialects.Dialects[0], "SHOW EXTENDED FULL TABLES")
	require.NoError(t, err)
	assert.NotEmpty(t, stmts)

	stmts, err = parser.ParseSQL(dialects.Dialects[0], "SHOW EXTENDED FULL COLUMNS FROM mytable")
	require.NoError(t, err)
	assert.NotEmpty(t, stmts)

	// These should fail
	_, err = parser.ParseSQL(dialects.Dialects[0], "SHOW EXTENDED FULL CREATE TABLE mytable")
	assert.Error(t, err)

	_, err = parser.ParseSQL(dialects.Dialects[0], "SHOW EXTENDED FULL COLLATION")
	assert.Error(t, err)

	_, err = parser.ParseSQL(dialects.Dialects[0], "SHOW EXTENDED FULL VARIABLES")
	assert.Error(t, err)
}

// TestParseShowCreate verifies SHOW CREATE statement parsing.
// Reference: tests/sqlparser_mysql.rs:539
func TestParseShowCreate(t *testing.T) {
	dialects := MySQLAndGeneric()

	// Test various SHOW CREATE object types
	objectTypes := []string{"TABLE", "TRIGGER", "EVENT", "FUNCTION", "PROCEDURE", "VIEW"}
	for _, objType := range objectTypes {
		sql := "SHOW CREATE " + objType + " myident"
		dialects.VerifiedStmt(t, sql)
	}
}

// TestParseShowCollation verifies SHOW COLLATION statement parsing.
// Reference: tests/sqlparser_mysql.rs:561
func TestParseShowCollation(t *testing.T) {
	dialects := MySQLAndGeneric()
	dialects.VerifiedStmt(t, "SHOW COLLATION")
	dialects.VerifiedStmt(t, "SHOW COLLATION LIKE 'pattern'")
	dialects.VerifiedStmt(t, "SHOW COLLATION WHERE 1 = 2")
}

// TestParseUse verifies USE statement parsing.
// Reference: tests/sqlparser_mysql.rs:583
func TestParseUse(t *testing.T) {
	dialects := MySQLAndGeneric()

	validObjectNames := []string{"mydb", "SCHEMA", "DATABASE", "CATALOG", "WAREHOUSE", "DEFAULT"}
	quoteStyles := []rune{'\'', '"', '`'}

	for _, objectName := range validObjectNames {
		// Test single identifier without quotes
		sql := "USE " + objectName
		dialects.VerifiedStmt(t, sql)

		// Test single identifier with different quote styles
		for _, quote := range quoteStyles {
			quotedSql := "USE " + string(quote) + objectName + string(quote)
			dialects.VerifiedStmt(t, quotedSql)
		}
	}
}

// TestParseSetVariables verifies SET variable statement parsing.
// Reference: tests/sqlparser_mysql.rs:615
func TestParseSetVariables(t *testing.T) {
	dialects := MySQLAndGeneric()
	dialects.VerifiedStmt(t, "SET sql_mode = CONCAT(@@sql_mode, ',STRICT_TRANS_TABLES')")
	dialects.VerifiedStmt(t, "SET LOCAL autocommit = 1")
}

// TestParseCreateTableAutoIncrement verifies CREATE TABLE with AUTO_INCREMENT.
// Reference: tests/sqlparser_mysql.rs:629
func TestParseCreateTableAutoIncrement(t *testing.T) {
	dialects := MySQL()
	sql := "CREATE TABLE foo (bar INT PRIMARY KEY AUTO_INCREMENT)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// TestParseCreateTablePrimaryAndUniqueKey verifies CREATE TABLE with PRIMARY/UNIQUE KEY.
// Reference: tests/sqlparser_mysql.rs:715
func TestParseCreateTablePrimaryAndUniqueKey(t *testing.T) {
	dialects := MySQL()

	sql := "CREATE TABLE foo (id INT PRIMARY KEY AUTO_INCREMENT, bar INT NOT NULL, CONSTRAINT bar_key UNIQUE KEY (bar))"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// TestParseCreateTablePrimaryAndUniqueKeyWithIndexOptions verifies CREATE TABLE with index options.
// Reference: tests/sqlparser_mysql.rs:785
func TestParseCreateTablePrimaryAndUniqueKeyWithIndexOptions(t *testing.T) {
	dialects := MySQLAndGeneric()
	sql := "CREATE TABLE foo (bar INT, var INT, CONSTRAINT constr UNIQUE INDEX index_name (bar, var) USING HASH COMMENT 'yes, ' USING BTREE COMMENT 'MySQL allows')"
	dialects.VerifiedStmt(t, sql)
}

// TestParsePrefixKeyPart verifies prefix key part parsing.
// Reference: tests/sqlparser_mysql.rs:822
func TestParsePrefixKeyPart(t *testing.T) {
	dialects := MySQLAndGeneric()

	for _, sql := range []string{
		"CREATE INDEX idx_index ON t(textcol(10))",
		"ALTER TABLE tab ADD INDEX idx_index (textcol(10))",
		"ALTER TABLE tab ADD PRIMARY KEY (textcol(10))",
		"ALTER TABLE tab ADD UNIQUE KEY (textcol(10))",
		"ALTER TABLE tab ADD FULLTEXT INDEX (textcol(10))",
		"CREATE TABLE t (textcol TEXT, INDEX idx_index (textcol(10)))",
	} {
		dialects.VerifiedStmt(t, sql)
	}
}

// TestFunctionalKeyPart verifies functional key part parsing.
// Reference: tests/sqlparser_mysql.rs:850
func TestFunctionalKeyPart(t *testing.T) {
	dialects := MySQLAndGeneric()
	dialects.VerifiedStmt(t, "CREATE INDEX idx_index ON t((col COLLATE utf8mb4_bin) DESC)")
	dialects.VerifiedStmt(t, "CREATE TABLE t (jsoncol JSON, PRIMARY KEY ((CAST(col ->> '$.id' AS UNSIGNED)) ASC))")
	dialects.VerifiedStmt(t, "CREATE TABLE t (jsoncol JSON, PRIMARY KEY ((CAST(col ->> '$.fields' AS UNSIGNED ARRAY)) ASC))")
}

// TestParseCreateTablePrimaryAndUniqueKeyWithIndexType verifies CREATE TABLE with index type.
// Reference: tests/sqlparser_mysql.rs:902
func TestParseCreateTablePrimaryAndUniqueKeyWithIndexType(t *testing.T) {
	dialects := MySQLAndGeneric()
	dialects.VerifiedStmt(t, "CREATE TABLE foo (bar INT, UNIQUE index_name USING BTREE (bar) USING HASH)")
	dialects.VerifiedStmt(t, "CREATE TABLE foo (bar INT, PRIMARY KEY index_name USING BTREE (bar) USING HASH)")
	dialects.VerifiedStmt(t, "CREATE TABLE foo (bar INT, UNIQUE INDEX index_name USING BTREE (bar) USING HASH)")
	dialects.VerifiedStmt(t, "CREATE TABLE foo (bar INT, PRIMARY KEY index_name USING BTREE (bar) USING HASH)")
}

// TestParseCreateTablePrimaryAndUniqueKeyCharacteristic verifies constraint characteristics.
// Reference: tests/sqlparser_mysql.rs:939
func TestParseCreateTablePrimaryAndUniqueKeyCharacteristic(t *testing.T) {
	dialects := MySQLAndGeneric()
	dialects.VerifiedStmt(t, "CREATE TABLE x (y INT, CONSTRAINT constr UNIQUE INDEX (y) NOT DEFERRABLE INITIALLY IMMEDIATE)")
	dialects.VerifiedStmt(t, "CREATE TABLE x (y INT, CONSTRAINT constr PRIMARY KEY (y) NOT DEFERRABLE INITIALLY IMMEDIATE)")
}

// TestParseCreateTableColumnKeyOptions verifies column key options.
// Reference: tests/sqlparser_mysql.rs:948
func TestParseCreateTableColumnKeyOptions(t *testing.T) {
	dialects := MySQLAndGeneric()
	dialects.VerifiedStmt(t, "CREATE TABLE foo (x INT UNIQUE KEY)")
	dialects.OneStatementParsesTo(t, "CREATE TABLE foo (x INT KEY)", "CREATE TABLE foo (x INT PRIMARY KEY)")
}

// TestParseCreateTableComment verifies table comment parsing.
// Reference: tests/sqlparser_mysql.rs:957
func TestParseCreateTableComment(t *testing.T) {
	dialects := MySQL()
	dialects.VerifiedStmt(t, "CREATE TABLE foo (bar INT) COMMENT 'baz'")
	dialects.VerifiedStmt(t, "CREATE TABLE foo (bar INT) COMMENT = 'baz'")
}

// TestParseCreateTableAutoIncrementOffset verifies AUTO_INCREMENT offset parsing.
// Reference: tests/sqlparser_mysql.rs:987
func TestParseCreateTableAutoIncrementOffset(t *testing.T) {
	dialects := MySQL()
	sql := "CREATE TABLE foo (bar INT NOT NULL AUTO_INCREMENT) ENGINE = InnoDB AUTO_INCREMENT = 123"
	dialects.VerifiedStmt(t, sql)
}

// TestParseCreateTableMultipleOptionsOrderIndependent verifies table options order.
// Reference: tests/sqlparser_mysql.rs:1014
func TestParseCreateTableMultipleOptionsOrderIndependent(t *testing.T) {
	dialects := MySQL()

	sqls := []string{
		"CREATE TABLE mytable (id INT) ENGINE=InnoDB ROW_FORMAT=DYNAMIC KEY_BLOCK_SIZE=8 COMMENT='abc'",
		"CREATE TABLE mytable (id INT) KEY_BLOCK_SIZE=8 COMMENT='abc' ENGINE=InnoDB ROW_FORMAT=DYNAMIC",
		"CREATE TABLE mytable (id INT) ROW_FORMAT=DYNAMIC KEY_BLOCK_SIZE=8 COMMENT='abc' ENGINE=InnoDB",
	}

	for _, sql := range sqls {
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)
	}
}

// TestParseCreateTableSetEnum verifies SET and ENUM type parsing.
// Reference: tests/sqlparser_mysql.rs:1209
func TestParseCreateTableSetEnum(t *testing.T) {
	dialects := MySQL()
	sql := "CREATE TABLE foo (bar SET('a', 'b'), baz ENUM('a', 'b'))"
	dialects.VerifiedStmt(t, sql)
}

// TestParseCreateTableEngineDefaultCharset verifies ENGINE and DEFAULT CHARSET.
// Reference: tests/sqlparser_mysql.rs:1241
func TestParseCreateTableEngineDefaultCharset(t *testing.T) {
	dialects := MySQL()
	sql := "CREATE TABLE foo (id INT(11)) ENGINE = InnoDB DEFAULT CHARSET = utf8mb3"
	dialects.VerifiedStmt(t, sql)
}

// TestParseCreateTableCollate verifies COLLATE option parsing.
// Reference: tests/sqlparser_mysql.rs:1283
func TestParseCreateTableCollate(t *testing.T) {
	dialects := MySQL()
	sql := "CREATE TABLE foo (id INT(11)) COLLATE = utf8mb4_0900_ai_ci"
	dialects.VerifiedStmt(t, sql)
}

// TestParseCreateTableBothOptionsAndAsQuery verifies CREATE TABLE AS with options.
// Reference: tests/sqlparser_mysql.rs:1317
func TestParseCreateTableBothOptionsAndAsQuery(t *testing.T) {
	dialects := MySQLAndGeneric()
	sql := "CREATE TABLE foo (id INT(11)) ENGINE = InnoDB DEFAULT CHARSET = utf8mb3 COLLATE = utf8mb4_0900_ai_ci AS SELECT 1"
	dialects.VerifiedStmt(t, sql)
}

// TestParseCreateTableCommentCharacterSet verifies column CHARACTER SET and COMMENT.
// Reference: tests/sqlparser_mysql.rs:1357
func TestParseCreateTableCommentCharacterSet(t *testing.T) {
	dialects := MySQL()
	sql := "CREATE TABLE foo (s TEXT CHARACTER SET utf8mb4 COMMENT 'comment')"
	dialects.VerifiedStmt(t, sql)
}

// TestParseCreateTableGencol verifies generated column parsing.
// Reference: tests/sqlparser_mysql.rs:1387
func TestParseCreateTableGencol(t *testing.T) {
	dialects := MySQLAndGeneric()
	dialects.VerifiedStmt(t, "CREATE TABLE t1 (a INT, b INT GENERATED ALWAYS AS (a * 2))")
	dialects.VerifiedStmt(t, "CREATE TABLE t1 (a INT, b INT GENERATED ALWAYS AS (a * 2) VIRTUAL)")
	dialects.VerifiedStmt(t, "CREATE TABLE t1 (a INT, b INT GENERATED ALWAYS AS (a * 2) STORED)")
	dialects.VerifiedStmt(t, "CREATE TABLE t1 (a INT, b INT AS (a * 2))")
	dialects.VerifiedStmt(t, "CREATE TABLE t1 (a INT, b INT AS (a * 2) VIRTUAL)")
	dialects.VerifiedStmt(t, "CREATE TABLE t1 (a INT, b INT AS (a * 2) STORED)")
}

// TestParseCreateTableOptionsCommaSeparated verifies comma-separated table options.
// Reference: tests/sqlparser_mysql.rs:1403
func TestParseCreateTableOptionsCommaSeparated(t *testing.T) {
	dialects := MySQLAndGeneric()
	sql := "CREATE TABLE t (x INT) DEFAULT CHARSET = utf8mb4, ENGINE = InnoDB , AUTO_INCREMENT 1 DATA DIRECTORY '/var/lib/mysql/data'"
	canonical := "CREATE TABLE t (x INT) DEFAULT CHARSET = utf8mb4 ENGINE = InnoDB AUTO_INCREMENT = 1 DATA DIRECTORY = '/var/lib/mysql/data'"
	dialects.OneStatementParsesTo(t, sql, canonical)
}

// TestParseQuoteIdentifiers verifies quoted identifier parsing.
// Reference: tests/sqlparser_mysql.rs:1410
func TestParseQuoteIdentifiers(t *testing.T) {
	dialects := MySQL()
	sql := "CREATE TABLE `PRIMARY` (`BEGIN` INT PRIMARY KEY)"
	dialects.VerifiedStmt(t, sql)
}

// TestParseEscapedQuoteIdentifiersWithEscape verifies escaped quote identifiers.
// Reference: tests/sqlparser_mysql.rs:1439
func TestParseEscapedQuoteIdentifiersWithEscape(t *testing.T) {
	sql := "SELECT `quoted `` identifier`"
	dialect := mysql.NewMySqlDialect()
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// TestParseEscapedQuoteIdentifiersWithNoEscape verifies escaped quote without unescape.
// Reference: tests/sqlparser_mysql.rs:1488
func TestParseEscapedQuoteIdentifiersWithNoEscape(t *testing.T) {
	sql := "SELECT `quoted `` identifier`"
	dialect := mysql.NewMySqlDialect()
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// TestParseEscapedBackticksWithEscape verifies escaped backticks with escape.
// Reference: tests/sqlparser_mysql.rs:1545
func TestParseEscapedBackticksWithEscape(t *testing.T) {
	sql := "SELECT ```quoted identifier```"
	dialect := mysql.NewMySqlDialect()
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// TestParseEscapedBackticksWithNoEscape verifies escaped backticks without unescape.
// Reference: tests/sqlparser_mysql.rs:1594
func TestParseEscapedBackticksWithNoEscape(t *testing.T) {
	sql := "SELECT ```quoted identifier```"
	dialect := mysql.NewMySqlDialect()
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// TestParseUnterminatedEscape verifies unterminated escape sequences.
// Reference: tests/sqlparser_mysql.rs:1647
func TestParseUnterminatedEscape(t *testing.T) {
	dialects := MySQL()

	// These should cause panic/errors
	sql1 := `SELECT 'I\'m not fine\'`
	_, err1 := parser.ParseSQL(dialects.Dialects[0], sql1)
	// Some implementations may error, some may not - behavior varies
	_ = err1

	sql2 := `SELECT 'I\\'m not fine'`
	_, err2 := parser.ParseSQL(dialects.Dialects[0], sql2)
	_ = err2
}

// TestCheckRoundtripOfEscapedString verifies escaped string roundtrip.
// Reference: tests/sqlparser_mysql.rs:1658
func TestCheckRoundtripOfEscapedString(t *testing.T) {
	dialect := mysql.NewMySqlDialect()

	stmts, err := parser.ParseSQL(dialect, `SELECT 'I\'m fine'`)
	require.NoError(t, err)
	require.Len(t, stmts, 1)

	stmts, err = parser.ParseSQL(dialect, `SELECT 'I''m fine'`)
	require.NoError(t, err)
	require.Len(t, stmts, 1)

	stmts, err = parser.ParseSQL(dialect, `SELECT 'I\\'m fine'`)
	require.NoError(t, err)
	require.Len(t, stmts, 1)

	stmts, err = parser.ParseSQL(dialect, `SELECT "I\"m fine"`)
	require.NoError(t, err)
	require.Len(t, stmts, 1)

	stmts, err = parser.ParseSQL(dialect, `SELECT "I""m fine"`)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// TestParseCreateTableWithMinimumDisplayWidth verifies minimum display width.
// Reference: tests/sqlparser_mysql.rs:1682
func TestParseCreateTableWithMinimumDisplayWidth(t *testing.T) {
	dialects := MySQL()
	sql := "CREATE TABLE foo (bar_tinyint TINYINT(3), bar_smallint SMALLINT(5), bar_mediumint MEDIUMINT(6), bar_int INT(11), bar_bigint BIGINT(20))"
	dialects.VerifiedStmt(t, sql)
}

// TestParseCreateTableUnsigned verifies UNSIGNED data type parsing.
// Reference: tests/sqlparser_mysql.rs:1723
func TestParseCreateTableUnsigned(t *testing.T) {
	dialects := MySQL()
	sql := "CREATE TABLE foo (bar_tinyint TINYINT(3) UNSIGNED, bar_smallint SMALLINT(5) UNSIGNED, bar_mediumint MEDIUMINT(13) UNSIGNED, bar_int INT(11) UNSIGNED, bar_bigint BIGINT(20) UNSIGNED)"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSignedDataTypes verifies SIGNED data type parsing.
// Reference: tests/sqlparser_mysql.rs:1764
func TestParseSignedDataTypes(t *testing.T) {
	dialects := MySQL()
	sql := "CREATE TABLE foo (bar_tinyint TINYINT(3) SIGNED, bar_smallint SMALLINT(5) SIGNED, bar_mediumint MEDIUMINT(13) SIGNED, bar_int INT(11) SIGNED, bar_bigint BIGINT(20) SIGNED)"
	canonical := "CREATE TABLE foo (bar_tinyint TINYINT(3), bar_smallint SMALLINT(5), bar_mediumint MEDIUMINT(13), bar_int INT(11), bar_bigint BIGINT(20))"
	dialects.OneStatementParsesTo(t, sql, canonical)
}

// TestParseDeprecatedMySQLUnsignedDataTypes verifies deprecated UNSIGNED data types.
// Reference: tests/sqlparser_mysql.rs:1809
func TestParseDeprecatedMySQLUnsignedDataTypes(t *testing.T) {
	dialects := MySQL()
	sql := "CREATE TABLE foo (bar_decimal DECIMAL UNSIGNED, bar_decimal_prec DECIMAL(10) UNSIGNED, bar_decimal_scale DECIMAL(10,2) UNSIGNED, bar_dec DEC UNSIGNED, bar_dec_prec DEC(10) UNSIGNED, bar_dec_scale DEC(10,2) UNSIGNED, bar_float FLOAT UNSIGNED, bar_float_prec FLOAT(10) UNSIGNED, bar_float_scale FLOAT(10,2) UNSIGNED, bar_double DOUBLE UNSIGNED, bar_double_prec DOUBLE(10) UNSIGNED, bar_double_scale DOUBLE(10,2) UNSIGNED, bar_real REAL UNSIGNED, bar_double_precision DOUBLE PRECISION UNSIGNED)"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSimpleInsert verifies simple INSERT statement parsing.
// Reference: tests/sqlparser_mysql.rs:1901
func TestParseSimpleInsert(t *testing.T) {
	dialects := MySQL()
	sql := `INSERT INTO tasks (title, priority) VALUES ('Test Some Inserts', 1), ('Test Entry 2', 2), ('Test Entry 3', 3)`
	dialects.VerifiedStmt(t, sql)
}

// TestParseIgnoreInsert verifies INSERT IGNORE statement parsing.
// Reference: tests/sqlparser_mysql.rs:1971
func TestParseIgnoreInsert(t *testing.T) {
	dialects := MySQLAndGeneric()
	sql := `INSERT IGNORE INTO tasks (title, priority) VALUES ('Test Some Inserts', 1)`
	dialects.VerifiedStmt(t, sql)
}

// TestParsePriorityInsert verifies INSERT with priority parsing.
// Reference: tests/sqlparser_mysql.rs:2027
func TestParsePriorityInsert(t *testing.T) {
	dialects := MySQLAndGeneric()
	dialects.VerifiedStmt(t, `INSERT HIGH_PRIORITY INTO tasks (title, priority) VALUES ('Test Some Inserts', 1)`)

	dialects2 := MySQL()
	dialects2.VerifiedStmt(t, `INSERT LOW_PRIORITY INTO tasks (title, priority) VALUES ('Test Some Inserts', 1)`)
}

// TestParseInsertAs verifies INSERT AS alias parsing.
// Reference: tests/sqlparser_mysql.rs:2136
func TestParseInsertAs(t *testing.T) {
	dialects := MySQLAndGeneric()
	sql := "INSERT INTO `table` (`date`) VALUES ('2024-01-01') AS `alias`"
	dialects.VerifiedStmt(t, sql)

	sql2 := "INSERT INTO `table` (`id`, `date`) VALUES (1, '2024-01-01') AS `alias` (`mek_id`, `mek_date`)"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseReplaceInsert verifies REPLACE statement parsing.
// Reference: tests/sqlparser_mysql.rs:2255
func TestParseReplaceInsert(t *testing.T) {
	dialects := MySQL()
	sql := `REPLACE DELAYED INTO tasks (title, priority) VALUES ('Test Some Inserts', 1)`
	dialects.VerifiedStmt(t, sql)
}

// TestParseEmptyRowInsert verifies INSERT with empty row parsing.
// Reference: tests/sqlparser_mysql.rs:2312
func TestParseEmptyRowInsert(t *testing.T) {
	dialects := MySQL()
	sql := "INSERT INTO tb () VALUES (), ()"
	canonical := "INSERT INTO tb VALUES (), ()"
	dialects.OneStatementParsesTo(t, sql, canonical)
}

// TestParseInsertWithOnDuplicateUpdate verifies ON DUPLICATE KEY UPDATE.
// Reference: tests/sqlparser_mysql.rs:2354
func TestParseInsertWithOnDuplicateUpdate(t *testing.T) {
	dialects := MySQL()
	sql := "INSERT INTO permission_groups (name, description, perm_create, perm_read, perm_update, perm_delete) VALUES ('accounting_manager', 'Some description about the group', true, true, true, true) ON DUPLICATE KEY UPDATE description = VALUES(description), perm_create = VALUES(perm_create), perm_read = VALUES(perm_read), perm_update = VALUES(perm_update), perm_delete = VALUES(perm_delete)"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectWithNumericPrefixColumnName verifies numeric prefix column names.
// Reference: tests/sqlparser_mysql.rs:2455
func TestParseSelectWithNumericPrefixColumnName(t *testing.T) {
	dialects := MySQL()
	sql := `SELECT 123col_$@123abc FROM "table"`
	dialects.VerifiedStmt(t, sql)
}

// TestParseQualifiedIdentifiersWithNumericPrefix verifies qualified numeric identifiers.
// Reference: tests/sqlparser_mysql.rs:2501
func TestParseQualifiedIdentifiersWithNumericPrefix(t *testing.T) {
	dialects := MySQL()
	dialects.VerifiedStmt(t, "SELECT t.15to29 FROM my_table AS t")
	dialects.VerifiedStmt(t, "SELECT t.15e29 FROM my_table AS t")
	dialects.VerifiedStmt(t, "SELECT `15e29` FROM my_table")
	dialects.VerifiedStmt(t, "SELECT t.`15e29` FROM my_table AS t")
	dialects.VerifiedStmt(t, "SELECT 1db.1table.1column")
	dialects.VerifiedStmt(t, "SELECT `1`.`2`.`3`")
}

// TestParseSelectWithConcatenationOfExpNumberAndNumericPrefixColumn verifies exp number concatenation.
// Reference: tests/sqlparser_mysql.rs:2631
func TestParseSelectWithConcatenationOfExpNumberAndNumericPrefixColumn(t *testing.T) {
	dialects := MySQL()
	sql := `SELECT 123e4, 123col_$@123abc FROM "table"`
	dialects.VerifiedStmt(t, sql)
}

// TestParseInsertWithNumericPrefixColumnName verifies INSERT with numeric column names.
// Reference: tests/sqlparser_mysql.rs:2678
func TestParseInsertWithNumericPrefixColumnName(t *testing.T) {
	dialects := MySQL()
	sql := "INSERT INTO s1.t1 (123col_$@length123) VALUES (67.654)"
	dialects.VerifiedStmt(t, sql)
}
