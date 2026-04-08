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

// Package token defines:
//
//  1. A list of constants for every SQL keyword
//
//  2. An ALL_KEYWORDS slice with every keyword in it
//     This is not a list of *reserved* keywords: some of these can be
//     parsed as identifiers if the parser decides so. This means that
//     new keywords can be added here without affecting the parse result.
//
//     As a matter of fact, most of these keywords are not used at all
//     and could be removed.
//
//  3. A RESERVED_FOR_TABLE_ALIAS slice with keywords reserved in a
//     "table alias" context.
package token

// Keyword represents an SQL keyword.
// Keywords are sorted alphabetically to enable binary search matching.
type Keyword string

// Keyword constants
const (
	NoKeyword Keyword = ""

	// ABORT through ZSTD - SQL keyword constants
	ABORT                                    Keyword = "ABORT"
	ABS                                      Keyword = "ABS"
	ABSENT                                   Keyword = "ABSENT"
	ABSOLUTE                                 Keyword = "ABSOLUTE"
	ACCEPTANYDATE                            Keyword = "ACCEPTANYDATE"
	ACCEPTINVCHARS                           Keyword = "ACCEPTINVCHARS"
	ACCESS                                   Keyword = "ACCESS"
	ACCOUNT                                  Keyword = "ACCOUNT"
	ACTION                                   Keyword = "ACTION"
	ADD                                      Keyword = "ADD"
	ADDQUOTES                                Keyword = "ADDQUOTES"
	ADMIN                                    Keyword = "ADMIN"
	AFTER                                    Keyword = "AFTER"
	AGAINST                                  Keyword = "AGAINST"
	AGGREGATE                                Keyword = "AGGREGATE"
	AGGREGATION                              Keyword = "AGGREGATION"
	ALERT                                    Keyword = "ALERT"
	ALGORITHM                                Keyword = "ALGORITHM"
	ALIAS                                    Keyword = "ALIAS"
	ALIGNMENT                                Keyword = "ALIGNMENT"
	ALL                                      Keyword = "ALL"
	ALLOCATE                                 Keyword = "ALLOCATE"
	ALLOWOVERWRITE                           Keyword = "ALLOWOVERWRITE"
	ALTER                                    Keyword = "ALTER"
	ALWAYS                                   Keyword = "ALWAYS"
	ANALYZE                                  Keyword = "ANALYZE"
	AND                                      Keyword = "AND"
	ANTI                                     Keyword = "ANTI"
	ANY                                      Keyword = "ANY"
	APPLICATION                              Keyword = "APPLICATION"
	APPLY                                    Keyword = "APPLY"
	APPLYBUDGET                              Keyword = "APPLYBUDGET"
	ARCHIVE                                  Keyword = "ARCHIVE"
	ARE                                      Keyword = "ARE"
	ARRAY                                    Keyword = "ARRAY"
	ARRAY_MAX_CARDINALITY                    Keyword = "ARRAY_MAX_CARDINALITY"
	AS                                       Keyword = "AS"
	ASC                                      Keyword = "ASC"
	ASENSITIVE                               Keyword = "ASENSITIVE"
	ASOF                                     Keyword = "ASOF"
	ASSERT                                   Keyword = "ASSERT"
	ASYMMETRIC                               Keyword = "ASYMMETRIC"
	AT                                       Keyword = "AT"
	ATOMIC                                   Keyword = "ATOMIC"
	ATTACH                                   Keyword = "ATTACH"
	AUDIT                                    Keyword = "AUDIT"
	AUTHENTICATION                           Keyword = "AUTHENTICATION"
	AUTHORIZATION                            Keyword = "AUTHORIZATION"
	AUTHORIZATIONS                           Keyword = "AUTHORIZATIONS"
	AUTO                                     Keyword = "AUTO"
	AUTOEXTEND_SIZE                          Keyword = "AUTOEXTEND_SIZE"
	AUTOINCREMENT                            Keyword = "AUTOINCREMENT"
	AUTO_INCREMENT                           Keyword = "AUTO_INCREMENT"
	AVG                                      Keyword = "AVG"
	AVG_ROW_LENGTH                           Keyword = "AVG_ROW_LENGTH"
	AVRO                                     Keyword = "AVRO"
	BACKUP                                   Keyword = "BACKUP"
	BACKWARD                                 Keyword = "BACKWARD"
	BASE64                                   Keyword = "BASE64"
	BASE_LOCATION                            Keyword = "BASE_LOCATION"
	BEFORE                                   Keyword = "BEFORE"
	BEGIN                                    Keyword = "BEGIN"
	BEGIN_FRAME                              Keyword = "BEGIN_FRAME"
	BEGIN_PARTITION                          Keyword = "BEGIN_PARTITION"
	BERNOULLI                                Keyword = "BERNOULLI"
	BETWEEN                                  Keyword = "BETWEEN"
	BIGDECIMAL                               Keyword = "BIGDECIMAL"
	BIGINT                                   Keyword = "BIGINT"
	BIGNUMERIC                               Keyword = "BIGNUMERIC"
	BINARY                                   Keyword = "BINARY"
	BIND                                     Keyword = "BIND"
	BINDING                                  Keyword = "BINDING"
	BIT                                      Keyword = "BIT"
	BLANKSASNULL                             Keyword = "BLANKSASNULL"
	BLOB                                     Keyword = "BLOB"
	BLOCK                                    Keyword = "BLOCK"
	BLOOM                                    Keyword = "BLOOM"
	BLOOMFILTER                              Keyword = "BLOOMFILTER"
	BOOL                                     Keyword = "BOOL"
	BOOLEAN                                  Keyword = "BOOLEAN"
	BOOST                                    Keyword = "BOOST"
	BOTH                                     Keyword = "BOTH"
	BOX                                      Keyword = "BOX"
	BRIN                                     Keyword = "BRIN"
	BROWSE                                   Keyword = "BROWSE"
	BTREE                                    Keyword = "BTREE"
	BUCKET                                   Keyword = "BUCKET"
	BUCKETS                                  Keyword = "BUCKETS"
	BY                                       Keyword = "BY"
	BYPASSRLS                                Keyword = "BYPASSRLS"
	BYTEA                                    Keyword = "BYTEA"
	BYTES                                    Keyword = "BYTES"
	BZIP2                                    Keyword = "BZIP2"
	CACHE                                    Keyword = "CACHE"
	CALL                                     Keyword = "CALL"
	CALLED                                   Keyword = "CALLED"
	CANONICAL                                Keyword = "CANONICAL"
	CARDINALITY                              Keyword = "CARDINALITY"
	CASCADE                                  Keyword = "CASCADE"
	CASCADED                                 Keyword = "CASCADED"
	CASE                                     Keyword = "CASE"
	CASES                                    Keyword = "CASES"
	CAST                                     Keyword = "CAST"
	CATALOG                                  Keyword = "CATALOG"
	CATALOG_SYNC                             Keyword = "CATALOG_SYNC"
	CATALOG_SYNC_NAMESPACE_FLATTEN_DELIMITER Keyword = "CATALOG_SYNC_NAMESPACE_FLATTEN_DELIMITER"
	CATALOG_SYNC_NAMESPACE_MODE              Keyword = "CATALOG_SYNC_NAMESPACE_MODE"
	CATCH                                    Keyword = "CATCH"
	CATEGORY                                 Keyword = "CATEGORY"
	CEIL                                     Keyword = "CEIL"
	CEILING                                  Keyword = "CEILING"
	CENTURY                                  Keyword = "CENTURY"
	CHAIN                                    Keyword = "CHAIN"
	CHANGE                                   Keyword = "CHANGE"
	CHANGES                                  Keyword = "CHANGES"
	CHANGE_TRACKING                          Keyword = "CHANGE_TRACKING"
	CHANNEL                                  Keyword = "CHANNEL"
	CHAR                                     Keyword = "CHAR"
	CHARACTER                                Keyword = "CHARACTER"
	CHARACTERISTICS                          Keyword = "CHARACTERISTICS"
	CHARACTERS                               Keyword = "CHARACTERS"
	CHARACTER_LENGTH                         Keyword = "CHARACTER_LENGTH"
	CHARSET                                  Keyword = "CHARSET"
	CHAR_LENGTH                              Keyword = "CHAR_LENGTH"
	CHECK                                    Keyword = "CHECK"
	CHECKSUM                                 Keyword = "CHECKSUM"
	CIRCLE                                   Keyword = "CIRCLE"
	CLASS                                    Keyword = "CLASS"
	CLEANPATH                                Keyword = "CLEANPATH"
	CLEAR                                    Keyword = "CLEAR"
	CLOB                                     Keyword = "CLOB"
	CLONE                                    Keyword = "CLONE"
	CLOSE                                    Keyword = "CLOSE"
	CLUSTER                                  Keyword = "CLUSTER"
	CLUSTERED                                Keyword = "CLUSTERED"
	CLUSTERING                               Keyword = "CLUSTERING"
	COALESCE                                 Keyword = "COALESCE"
	COLLATABLE                               Keyword = "COLLATABLE"
	COLLATE                                  Keyword = "COLLATE"
	COLLATION                                Keyword = "COLLATION"
	COLLECT                                  Keyword = "COLLECT"
	COLLECTION                               Keyword = "COLLECTION"
	COLUMN                                   Keyword = "COLUMN"
	COLUMNS                                  Keyword = "COLUMNS"
	COLUMNSTORE                              Keyword = "COLUMNSTORE"
	COMMENT                                  Keyword = "COMMENT"
	COMMIT                                   Keyword = "COMMIT"
	COMMITTED                                Keyword = "COMMITTED"
	COMMUTATOR                               Keyword = "COMMUTATOR"
	COMPATIBLE                               Keyword = "COMPATIBLE"
	COMPRESS                                 Keyword = "COMPRESS"
	COMPRESSION                              Keyword = "COMPRESSION"
	COMPUPDATE                               Keyword = "COMPUPDATE"
	COMPUTE                                  Keyword = "COMPUTE"
	CONCURRENTLY                             Keyword = "CONCURRENTLY"
	CONDITION                                Keyword = "CONDITION"
	CONFLICT                                 Keyword = "CONFLICT"
	CONNECT                                  Keyword = "CONNECT"
	CONNECTION                               Keyword = "CONNECTION"
	CONNECTOR                                Keyword = "CONNECTOR"
	CONNECT_BY_ROOT                          Keyword = "CONNECT_BY_ROOT"
	CONSTRAINT                               Keyword = "CONSTRAINT"
	CONTACT                                  Keyword = "CONTACT"
	CONTAINS                                 Keyword = "CONTAINS"
	CONTINUE                                 Keyword = "CONTINUE"
	CONVERT                                  Keyword = "CONVERT"
	COPY                                     Keyword = "COPY"
	COPY_OPTIONS                             Keyword = "COPY_OPTIONS"
	CORR                                     Keyword = "CORR"
	CORRESPONDING                            Keyword = "CORRESPONDING"
	COUNT                                    Keyword = "COUNT"
	COVAR_POP                                Keyword = "COVAR_POP"
	COVAR_SAMP                               Keyword = "COVAR_SAMP"
	CREATE                                   Keyword = "CREATE"
	CREATEDB                                 Keyword = "CREATEDB"
	CREATEROLE                               Keyword = "CREATEROLE"
	CREDENTIALS                              Keyword = "CREDENTIALS"
	CROSS                                    Keyword = "CROSS"
	CSV                                      Keyword = "CSV"
	CUBE                                     Keyword = "CUBE"
	CUME_DIST                                Keyword = "CUME_DIST"
	CURRENT                                  Keyword = "CURRENT"
	CURRENT_CATALOG                          Keyword = "CURRENT_CATALOG"
	CURRENT_DATE                             Keyword = "CURRENT_DATE"
	CURRENT_DEFAULT_TRANSFORM_GROUP          Keyword = "CURRENT_DEFAULT_TRANSFORM_GROUP"
	CURRENT_PATH                             Keyword = "CURRENT_PATH"
	CURRENT_ROLE                             Keyword = "CURRENT_ROLE"
	CURRENT_ROW                              Keyword = "CURRENT_ROW"
	CURRENT_SCHEMA                           Keyword = "CURRENT_SCHEMA"
	CURRENT_TIME                             Keyword = "CURRENT_TIME"
	CURRENT_TIMESTAMP                        Keyword = "CURRENT_TIMESTAMP"
	CURRENT_TRANSFORM_GROUP_FOR_TYPE         Keyword = "CURRENT_TRANSFORM_GROUP_FOR_TYPE"
	CURRENT_USER                             Keyword = "CURRENT_USER"
	CURSOR                                   Keyword = "CURSOR"
	CYCLE                                    Keyword = "CYCLE"
	DATA                                     Keyword = "DATA"
	DATABASE                                 Keyword = "DATABASE"
	DATABASES                                Keyword = "DATABASES"
	DATA_RETENTION_TIME_IN_DAYS              Keyword = "DATA_RETENTION_TIME_IN_DAYS"
	DATE                                     Keyword = "DATE"
	DATE32                                   Keyword = "DATE32"
	DATEFORMAT                               Keyword = "DATEFORMAT"
	DATETIME                                 Keyword = "DATETIME"
	DATETIME64                               Keyword = "DATETIME64"
	DAY                                      Keyword = "DAY"
	DAYOFWEEK                                Keyword = "DAYOFWEEK"
	DAYOFYEAR                                Keyword = "DAYOFYEAR"
	DAYS                                     Keyword = "DAYS"
	DCPROPERTIES                             Keyword = "DCPROPERTIES"
	DEALLOCATE                               Keyword = "DEALLOCATE"
	DEC                                      Keyword = "DEC"
	DECADE                                   Keyword = "DECADE"
	DECIMAL                                  Keyword = "DECIMAL"
	DECLARE                                  Keyword = "DECLARE"
	DEDUPLICATE                              Keyword = "DEDUPLICATE"
	DEFAULT                                  Keyword = "DEFAULT"
	DEFAULTS                                 Keyword = "DEFAULTS"
	DEFAULT_DDL_COLLATION                    Keyword = "DEFAULT_DDL_COLLATION"
	DEFAULT_MFA_METHOD                       Keyword = "DEFAULT_MFA_METHOD"
	DEFAULT_SECONDARY_ROLES                  Keyword = "DEFAULT_SECONDARY_ROLES"
	DEFERRABLE                               Keyword = "DEFERRABLE"
	DEFERRED                                 Keyword = "DEFERRED"
	DEFINE                                   Keyword = "DEFINE"
	DEFINED                                  Keyword = "DEFINED"
	DEFINER                                  Keyword = "DEFINER"
	DELAY                                    Keyword = "DELAY"
	DELAYED                                  Keyword = "DELAYED"
	DELAY_KEY_WRITE                          Keyword = "DELAY_KEY_WRITE"
	DELEGATED                                Keyword = "DELEGATED"
	DELETE                                   Keyword = "DELETE"
	DELIMITED                                Keyword = "DELIMITED"
	DELIMITER                                Keyword = "DELIMITER"
	DELTA                                    Keyword = "DELTA"
	DENSE_RANK                               Keyword = "DENSE_RANK"
	DENY                                     Keyword = "DENY"
	DEREF                                    Keyword = "DEREF"
	DESC                                     Keyword = "DESC"
	DESCRIBE                                 Keyword = "DESCRIBE"
	DETACH                                   Keyword = "DETACH"
	DETAIL                                   Keyword = "DETAIL"
	DETERMINISTIC                            Keyword = "DETERMINISTIC"
	DIMENSIONS                               Keyword = "DIMENSIONS"
	DIRECTORY                                Keyword = "DIRECTORY"
	DISABLE                                  Keyword = "DISABLE"
	DISCARD                                  Keyword = "DISCARD"
	DISCONNECT                               Keyword = "DISCONNECT"
	DISTINCT                                 Keyword = "DISTINCT"
	DISTINCTROW                              Keyword = "DISTINCTROW"
	DISTKEY                                  Keyword = "DISTKEY"
	DISTRIBUTE                               Keyword = "DISTRIBUTE"
	DISTSTYLE                                Keyword = "DISTSTYLE"
	DIV                                      Keyword = "DIV"
	DO                                       Keyword = "DO"
	DOMAIN                                   Keyword = "DOMAIN"
	DOUBLE                                   Keyword = "DOUBLE"
	DOW                                      Keyword = "DOW"
	DOWNSTREAM                               Keyword = "DOWNSTREAM"
	DOY                                      Keyword = "DOY"
	DROP                                     Keyword = "DROP"
	DRY                                      Keyword = "DRY"
	DUO                                      Keyword = "DUO"
	DUPLICATE                                Keyword = "DUPLICATE"
	DYNAMIC                                  Keyword = "DYNAMIC"
	EACH                                     Keyword = "EACH"
	ELEMENT                                  Keyword = "ELEMENT"
	ELEMENTS                                 Keyword = "ELEMENTS"
	ELSE                                     Keyword = "ELSE"
	ELSEIF                                   Keyword = "ELSEIF"
	EMPTY                                    Keyword = "EMPTY"
	EMPTYASNULL                              Keyword = "EMPTYASNULL"
	ENABLE                                   Keyword = "ENABLE"
	ENABLE_SCHEMA_EVOLUTION                  Keyword = "ENABLE_SCHEMA_EVOLUTION"
	ENCODING                                 Keyword = "ENCODING"
	ENCRYPTED                                Keyword = "ENCRYPTED"
	ENCRYPTION                               Keyword = "ENCRYPTION"
	END                                      Keyword = "END"
	END_EXEC                                 Keyword = "END-EXEC"
	ENDPOINT                                 Keyword = "ENDPOINT"
	END_FRAME                                Keyword = "END_FRAME"
	END_PARTITION                            Keyword = "END_PARTITION"
	ENFORCED                                 Keyword = "ENFORCED"
	ENGINE                                   Keyword = "ENGINE"
	ENGINE_ATTRIBUTE                         Keyword = "ENGINE_ATTRIBUTE"
	ENROLL                                   Keyword = "ENROLL"
	ENUM                                     Keyword = "ENUM"
	ENUM16                                   Keyword = "ENUM16"
	ENUM8                                    Keyword = "ENUM8"
	EPHEMERAL                                Keyword = "EPHEMERAL"
	EPOCH                                    Keyword = "EPOCH"
	EQUALS                                   Keyword = "EQUALS"
	ERROR                                    Keyword = "ERROR"
	ESCAPE                                   Keyword = "ESCAPE"
	ESCAPED                                  Keyword = "ESCAPED"
	ESTIMATE                                 Keyword = "ESTIMATE"
	EVEN                                     Keyword = "EVEN"
	EVENT                                    Keyword = "EVENT"
	EVERY                                    Keyword = "EVERY"
	EVOLVE                                   Keyword = "EVOLVE"
	EXCEPT                                   Keyword = "EXCEPT"
	EXCEPTION                                Keyword = "EXCEPTION"
	EXCHANGE                                 Keyword = "EXCHANGE"
	EXCLUDE                                  Keyword = "EXCLUDE"
	EXCLUDING                                Keyword = "EXCLUDING"
	EXCLUSIVE                                Keyword = "EXCLUSIVE"
	EXEC                                     Keyword = "EXEC"
	EXECUTE                                  Keyword = "EXECUTE"
	EXECUTION                                Keyword = "EXECUTION"
	EXISTS                                   Keyword = "EXISTS"
	EXP                                      Keyword = "EXP"
	EXPANSION                                Keyword = "EXPANSION"
	EXPLAIN                                  Keyword = "EXPLAIN"
	EXPLICIT                                 Keyword = "EXPLICIT"
	EXPORT                                   Keyword = "EXPORT"
	EXTEND                                   Keyword = "EXTEND"
	EXTENDED                                 Keyword = "EXTENDED"
	EXTENSION                                Keyword = "EXTENSION"
	EXTERNAL                                 Keyword = "EXTERNAL"
	EXTERNAL_VOLUME                          Keyword = "EXTERNAL_VOLUME"
	EXTRACT                                  Keyword = "EXTRACT"
	FACTS                                    Keyword = "FACTS"
	FAIL                                     Keyword = "FAIL"
	FAILOVER                                 Keyword = "FAILOVER"
	FALSE                                    Keyword = "FALSE"
	FAMILY                                   Keyword = "FAMILY"
	FETCH                                    Keyword = "FETCH"
	FIELDS                                   Keyword = "FIELDS"
	FILE                                     Keyword = "FILE"
	FILES                                    Keyword = "FILES"
	FILE_FORMAT                              Keyword = "FILE_FORMAT"
	FILL                                     Keyword = "FILL"
	FILTER                                   Keyword = "FILTER"
	FINAL                                    Keyword = "FINAL"
	FIRST                                    Keyword = "FIRST"
	FIRST_VALUE                              Keyword = "FIRST_VALUE"
	FIXEDSTRING                              Keyword = "FIXEDSTRING"
	FIXEDWIDTH                               Keyword = "FIXEDWIDTH"
	FLATTEN                                  Keyword = "FLATTEN"
	FLOAT                                    Keyword = "FLOAT"
	FLOAT32                                  Keyword = "FLOAT32"
	FLOAT4                                   Keyword = "FLOAT4"
	FLOAT64                                  Keyword = "FLOAT64"
	FLOAT8                                   Keyword = "FLOAT8"
	FLOOR                                    Keyword = "FLOOR"
	FLUSH                                    Keyword = "FLUSH"
	FN                                       Keyword = "FN"
	FOLLOWING                                Keyword = "FOLLOWING"
	FOR                                      Keyword = "FOR"
	FORCE                                    Keyword = "FORCE"
	FORCE_NOT_NULL                           Keyword = "FORCE_NOT_NULL"
	FORCE_NULL                               Keyword = "FORCE_NULL"
	FORCE_QUOTE                              Keyword = "FORCE_QUOTE"
	FOREIGN                                  Keyword = "FOREIGN"
	FORMAT                                   Keyword = "FORMAT"
	FORMATTED                                Keyword = "FORMATTED"
	FORWARD                                  Keyword = "FORWARD"
	FRAME_ROW                                Keyword = "FRAME_ROW"
	FREE                                     Keyword = "FREE"
	FREEZE                                   Keyword = "FREEZE"
	FROM                                     Keyword = "FROM"
	FSCK                                     Keyword = "FSCK"
	FULFILLMENT                              Keyword = "FULFILLMENT"
	FULL                                     Keyword = "FULL"
	FULLTEXT                                 Keyword = "FULLTEXT"
	FUNCTION                                 Keyword = "FUNCTION"
	FUNCTIONS                                Keyword = "FUNCTIONS"
	FUSION                                   Keyword = "FUSION"
	FUTURE                                   Keyword = "FUTURE"
	GB                                       Keyword = "GB"
	GENERAL                                  Keyword = "GENERAL"
	GENERATE                                 Keyword = "GENERATE"
	GENERATED                                Keyword = "GENERATED"
	GEOGRAPHY                                Keyword = "GEOGRAPHY"
	GET                                      Keyword = "GET"
	GIN                                      Keyword = "GIN"
	GIST                                     Keyword = "GIST"
	GLOBAL                                   Keyword = "GLOBAL"
	GRANT                                    Keyword = "GRANT"
	GRANTED                                  Keyword = "GRANTED"
	GRANTS                                   Keyword = "GRANTS"
	GRAPHVIZ                                 Keyword = "GRAPHVIZ"
	GROUP                                    Keyword = "GROUP"
	GROUPING                                 Keyword = "GROUPING"
	GROUPS                                   Keyword = "GROUPS"
	GZIP                                     Keyword = "GZIP"
	HARD                                     Keyword = "HARD"
	HASH                                     Keyword = "HASH"
	HASHES                                   Keyword = "HASHES"
	HAVING                                   Keyword = "HAVING"
	HEADER                                   Keyword = "HEADER"
	HEAP                                     Keyword = "HEAP"
	HIGH_PRIORITY                            Keyword = "HIGH_PRIORITY"
	HISTORY                                  Keyword = "HISTORY"
	HIVEVAR                                  Keyword = "HIVEVAR"
	HOLD                                     Keyword = "HOLD"
	HOSTS                                    Keyword = "HOSTS"
	HOUR                                     Keyword = "HOUR"
	HOURS                                    Keyword = "HOURS"
	HUGEINT                                  Keyword = "HUGEINT"
	IAM_ROLE                                 Keyword = "IAM_ROLE"
	ICEBERG                                  Keyword = "ICEBERG"
	ID                                       Keyword = "ID"
	IDENTIFIED                               Keyword = "IDENTIFIED"
	IDENTITY                                 Keyword = "IDENTITY"
	IDENTITY_INSERT                          Keyword = "IDENTITY_INSERT"
	IF                                       Keyword = "IF"
	IGNORE                                   Keyword = "IGNORE"
	IGNOREHEADER                             Keyword = "IGNOREHEADER"
	ILIKE                                    Keyword = "ILIKE"
	IMMEDIATE                                Keyword = "IMMEDIATE"
	IMMUTABLE                                Keyword = "IMMUTABLE"
	IMPORT                                   Keyword = "IMPORT"
	IMPORTED                                 Keyword = "IMPORTED"
	IN                                       Keyword = "IN"
	INCLUDE                                  Keyword = "INCLUDE"
	INCLUDE_NULL_VALUES                      Keyword = "INCLUDE_NULL_VALUES"
	INCLUDING                                Keyword = "INCLUDING"
	INCREMENT                                Keyword = "INCREMENT"
	INCREMENTAL                              Keyword = "INCREMENTAL"
	INDEX                                    Keyword = "INDEX"
	INDICATOR                                Keyword = "INDICATOR"
	INHERIT                                  Keyword = "INHERIT"
	INHERITS                                 Keyword = "INHERITS"
	INITIALIZE                               Keyword = "INITIALIZE"
	INITIALLY                                Keyword = "INITIALLY"
	INNER                                    Keyword = "INNER"
	INOUT                                    Keyword = "INOUT"
	INPATH                                   Keyword = "INPATH"
	INPLACE                                  Keyword = "INPLACE"
	INPUT                                    Keyword = "INPUT"
	INPUTFORMAT                              Keyword = "INPUTFORMAT"
	INSENSITIVE                              Keyword = "INSENSITIVE"
	INSERT                                   Keyword = "INSERT"
	INSERT_METHOD                            Keyword = "INSERT_METHOD"
	INSTALL                                  Keyword = "INSTALL"
	INSTANT                                  Keyword = "INSTANT"
	INSTEAD                                  Keyword = "INSTEAD"
	INT                                      Keyword = "INT"
	INT128                                   Keyword = "INT128"
	INT16                                    Keyword = "INT16"
	INT2                                     Keyword = "INT2"
	INT256                                   Keyword = "INT256"
	INT32                                    Keyword = "INT32"
	INT4                                     Keyword = "INT4"
	INT64                                    Keyword = "INT64"
	INT8                                     Keyword = "INT8"
	INTEGER                                  Keyword = "INTEGER"
	INTEGRATION                              Keyword = "INTEGRATION"
	INTERNALLENGTH                           Keyword = "INTERNALLENGTH"
	INTERPOLATE                              Keyword = "INTERPOLATE"
	INTERSECT                                Keyword = "INTERSECT"
	INTERSECTION                             Keyword = "INTERSECTION"
	INTERVAL                                 Keyword = "INTERVAL"
	INTO                                     Keyword = "INTO"
	INVISIBLE                                Keyword = "INVISIBLE"
	INVOKER                                  Keyword = "INVOKER"
	IO                                       Keyword = "IO"
	IS                                       Keyword = "IS"
	ISODOW                                   Keyword = "ISODOW"
	ISOLATION                                Keyword = "ISOLATION"
	ISOWEEK                                  Keyword = "ISOWEEK"
	ISOYEAR                                  Keyword = "ISOYEAR"
	ITEMS                                    Keyword = "ITEMS"
	JAR                                      Keyword = "JAR"
	JOIN                                     Keyword = "JOIN"
	JSON                                     Keyword = "JSON"
	JSONB                                    Keyword = "JSONB"
	JSONFILE                                 Keyword = "JSONFILE"
	JSON_TABLE                               Keyword = "JSON_TABLE"
	JULIAN                                   Keyword = "JULIAN"
	KEY                                      Keyword = "KEY"
	KEYS                                     Keyword = "KEYS"
	KEY_BLOCK_SIZE                           Keyword = "KEY_BLOCK_SIZE"
	KILL                                     Keyword = "KILL"
	LAG                                      Keyword = "LAG"
	LAMBDA                                   Keyword = "LAMBDA"
	LANGUAGE                                 Keyword = "LANGUAGE"
	LARGE                                    Keyword = "LARGE"
	LAST                                     Keyword = "LAST"
	LAST_VALUE                               Keyword = "LAST_VALUE"
	LATERAL                                  Keyword = "LATERAL"
	LEAD                                     Keyword = "LEAD"
	LEADING                                  Keyword = "LEADING"
	LEAKPROOF                                Keyword = "LEAKPROOF"
	LEAST                                    Keyword = "LEAST"
	LEFT                                     Keyword = "LEFT"
	LEFTARG                                  Keyword = "LEFTARG"
	LEVEL                                    Keyword = "LEVEL"
	LIFECYCLE                                Keyword = "LIFECYCLE"
	LIKE                                     Keyword = "LIKE"
	LIKE_REGEX                               Keyword = "LIKE_REGEX"
	LIMIT                                    Keyword = "LIMIT"
	LINE                                     Keyword = "LINE"
	LINES                                    Keyword = "LINES"
	LIST                                     Keyword = "LIST"
	LISTEN                                   Keyword = "LISTEN"
	LISTING                                  Keyword = "LISTING"
	LN                                       Keyword = "LN"
	LOAD                                     Keyword = "LOAD"
	LOCAL                                    Keyword = "LOCAL"
	LOCALTIME                                Keyword = "LOCALTIME"
	LOCALTIMESTAMP                           Keyword = "LOCALTIMESTAMP"
	LOCATION                                 Keyword = "LOCATION"
	LOCK                                     Keyword = "LOCK"
	LOCKED                                   Keyword = "LOCKED"
	LOG                                      Keyword = "LOG"
	LOGIN                                    Keyword = "LOGIN"
	LOGS                                     Keyword = "LOGS"
	LONG                                     Keyword = "LONG"
	LONGBLOB                                 Keyword = "LONGBLOB"
	LONGTEXT                                 Keyword = "LONGTEXT"
	LOWCARDINALITY                           Keyword = "LOWCARDINALITY"
	LOWER                                    Keyword = "LOWER"
	LOW_PRIORITY                             Keyword = "LOW_PRIORITY"
	LS                                       Keyword = "LS"
	LSEG                                     Keyword = "LSEG"
	MACRO                                    Keyword = "MACRO"
	MAIN                                     Keyword = "MAIN"
	MANAGE                                   Keyword = "MANAGE"
	MANAGED                                  Keyword = "MANAGED"
	MANAGEDLOCATION                          Keyword = "MANAGEDLOCATION"
	MANIFEST                                 Keyword = "MANIFEST"
	MAP                                      Keyword = "MAP"
	MASKING                                  Keyword = "MASKING"
	MATCH                                    Keyword = "MATCH"
	MATCHED                                  Keyword = "MATCHED"
	MATCHES                                  Keyword = "MATCHES"
	MATCH_CONDITION                          Keyword = "MATCH_CONDITION"
	MATCH_RECOGNIZE                          Keyword = "MATCH_RECOGNIZE"
	MATERIALIZE                              Keyword = "MATERIALIZE"
	MATERIALIZED                             Keyword = "MATERIALIZED"
	MAX                                      Keyword = "MAX"
	MAXFILESIZE                              Keyword = "MAXFILESIZE"
	MAXVALUE                                 Keyword = "MAXVALUE"
	MAX_DATA_EXTENSION_TIME_IN_DAYS          Keyword = "MAX_DATA_EXTENSION_TIME_IN_DAYS"
	MAX_ROWS                                 Keyword = "MAX_ROWS"
	MB                                       Keyword = "MB"
	MEASURES                                 Keyword = "MEASURES"
	MEDIUMBLOB                               Keyword = "MEDIUMBLOB"
	MEDIUMINT                                Keyword = "MEDIUMINT"
	MEDIUMTEXT                               Keyword = "MEDIUMTEXT"
	MEMBER                                   Keyword = "MEMBER"
	MERGE                                    Keyword = "MERGE"
	MERGES                                   Keyword = "MERGES"
	MESSAGE                                  Keyword = "MESSAGE"
	METADATA                                 Keyword = "METADATA"
	METHOD                                   Keyword = "METHOD"
	METRIC                                   Keyword = "METRIC"
	METRICS                                  Keyword = "METRICS"
	MFA                                      Keyword = "MFA"
	MICROSECOND                              Keyword = "MICROSECOND"
	MICROSECONDS                             Keyword = "MICROSECONDS"
	MILLENIUM                                Keyword = "MILLENIUM"
	MILLENNIUM                               Keyword = "MILLENNIUM"
	MILLISECOND                              Keyword = "MILLISECOND"
	MILLISECONDS                             Keyword = "MILLISECONDS"
	MIN                                      Keyword = "MIN"
	MINUS                                    Keyword = "MINUS"
	MINUTE                                   Keyword = "MINUTE"
	MINUTES                                  Keyword = "MINUTES"
	MINVALUE                                 Keyword = "MINVALUE"
	MIN_ROWS                                 Keyword = "MIN_ROWS"
	MOD                                      Keyword = "MOD"
	MODE                                     Keyword = "MODE"
	MODIFIES                                 Keyword = "MODIFIES"
	MODIFY                                   Keyword = "MODIFY"
	MODULE                                   Keyword = "MODULE"
	MODULUS                                  Keyword = "MODULUS"
	MONITOR                                  Keyword = "MONITOR"
	MONTH                                    Keyword = "MONTH"
	MONTHS                                   Keyword = "MONTHS"
	MSCK                                     Keyword = "MSCK"
	MULTIRANGE_TYPE_NAME                     Keyword = "MULTIRANGE_TYPE_NAME"
	MULTISET                                 Keyword = "MULTISET"
	MUTATION                                 Keyword = "MUTATION"
	NAME                                     Keyword = "NAME"
	NAMES                                    Keyword = "NAMES"
	NANOSECOND                               Keyword = "NANOSECOND"
	NANOSECONDS                              Keyword = "NANOSECONDS"
	NATIONAL                                 Keyword = "NATIONAL"
	NATURAL                                  Keyword = "NATURAL"
	NCHAR                                    Keyword = "NCHAR"
	NCLOB                                    Keyword = "NCLOB"
	NEGATOR                                  Keyword = "NEGATOR"
	NEST                                     Keyword = "NEST"
	NESTED                                   Keyword = "NESTED"
	NETWORK                                  Keyword = "NETWORK"
	NEW                                      Keyword = "NEW"
	NEXT                                     Keyword = "NEXT"
	NFC                                      Keyword = "NFC"
	NFD                                      Keyword = "NFD"
	NFKC                                     Keyword = "NFKC"
	NFKD                                     Keyword = "NFKD"
	NO                                       Keyword = "NO"
	NOBYPASSRLS                              Keyword = "NOBYPASSRLS"
	NOCOMPRESS                               Keyword = "NOCOMPRESS"
	NOCREATEDB                               Keyword = "NOCREATEDB"
	NOCREATEROLE                             Keyword = "NOCREATEROLE"
	NOCYCLE                                  Keyword = "NOCYCLE"
	NOINHERIT                                Keyword = "NOINHERIT"
	NOLOGIN                                  Keyword = "NOLOGIN"
	NONE                                     Keyword = "NONE"
	NOORDER                                  Keyword = "NOORDER"
	NOREPLICATION                            Keyword = "NOREPLICATION"
	NORMALIZE                                Keyword = "NORMALIZE"
	NORMALIZED                               Keyword = "NORMALIZED"
	NOSCAN                                   Keyword = "NOSCAN"
	NOSUPERUSER                              Keyword = "NOSUPERUSER"
	NOT                                      Keyword = "NOT"
	NOTHING                                  Keyword = "NOTHING"
	NOTIFY                                   Keyword = "NOTIFY"
	NOTNULL                                  Keyword = "NOTNULL"
	NOWAIT                                   Keyword = "NOWAIT"
	NO_WRITE_TO_BINLOG                       Keyword = "NO_WRITE_TO_BINLOG"
	NTH_VALUE                                Keyword = "NTH_VALUE"
	NTILE                                    Keyword = "NTILE"
	NULL                                     Keyword = "NULL"
	NULLABLE                                 Keyword = "NULLABLE"
	NULLIF                                   Keyword = "NULLIF"
	NULLS                                    Keyword = "NULLS"
	NUMBER                                   Keyword = "NUMBER"
	NUMERIC                                  Keyword = "NUMERIC"
	NVARCHAR                                 Keyword = "NVARCHAR"
	OBJECT                                   Keyword = "OBJECT"
	OBJECTS                                  Keyword = "OBJECTS"
	OCCURRENCES_REGEX                        Keyword = "OCCURRENCES_REGEX"
	OCTETS                                   Keyword = "OCTETS"
	OCTET_LENGTH                             Keyword = "OCTET_LENGTH"
	OF                                       Keyword = "OF"
	OFF                                      Keyword = "OFF"
	OFFSET                                   Keyword = "OFFSET"
	OFFSETS                                  Keyword = "OFFSETS"
	OLD                                      Keyword = "OLD"
	OMIT                                     Keyword = "OMIT"
	ON                                       Keyword = "ON"
	ONE                                      Keyword = "ONE"
	ONLY                                     Keyword = "ONLY"
	ON_CREATE                                Keyword = "ON_CREATE"
	ON_SCHEDULE                              Keyword = "ON_SCHEDULE"
	OPEN                                     Keyword = "OPEN"
	OPENJSON                                 Keyword = "OPENJSON"
	OPERATE                                  Keyword = "OPERATE"
	OPERATOR                                 Keyword = "OPERATOR"
	OPTIMIZATION                             Keyword = "OPTIMIZATION"
	OPTIMIZE                                 Keyword = "OPTIMIZE"
	OPTIMIZED                                Keyword = "OPTIMIZED"
	OPTIMIZER_COSTS                          Keyword = "OPTIMIZER_COSTS"
	OPTION                                   Keyword = "OPTION"
	OPTIONS                                  Keyword = "OPTIONS"
	OR                                       Keyword = "OR"
	ORC                                      Keyword = "ORC"
	ORDER                                    Keyword = "ORDER"
	ORDINALITY                               Keyword = "ORDINALITY"
	ORGANIZATION                             Keyword = "ORGANIZATION"
	OTHER                                    Keyword = "OTHER"
	OTP                                      Keyword = "OTP"
	OUT                                      Keyword = "OUT"
	OUTER                                    Keyword = "OUTER"
	OUTPUT                                   Keyword = "OUTPUT"
	OUTPUTFORMAT                             Keyword = "OUTPUTFORMAT"
	OVER                                     Keyword = "OVER"
	OVERFLOW                                 Keyword = "OVERFLOW"
	OVERLAPS                                 Keyword = "OVERLAPS"
	OVERLAY                                  Keyword = "OVERLAY"
	OVERRIDE                                 Keyword = "OVERRIDE"
	OVERWRITE                                Keyword = "OVERWRITE"
	OWNED                                    Keyword = "OWNED"
	OWNER                                    Keyword = "OWNER"
	OWNERSHIP                                Keyword = "OWNERSHIP"
	PACKAGE                                  Keyword = "PACKAGE"
	PACKAGES                                 Keyword = "PACKAGES"
	PACK_KEYS                                Keyword = "PACK_KEYS"
	PARALLEL                                 Keyword = "PARALLEL"
	PARAMETER                                Keyword = "PARAMETER"
	PARQUET                                  Keyword = "PARQUET"
	PART                                     Keyword = "PART"
	PARTIAL                                  Keyword = "PARTIAL"
	PARTITION                                Keyword = "PARTITION"
	PARTITIONED                              Keyword = "PARTITIONED"
	PARTITIONS                               Keyword = "PARTITIONS"
	PASSEDBYVALUE                            Keyword = "PASSEDBYVALUE"
	PASSING                                  Keyword = "PASSING"
	PASSKEY                                  Keyword = "PASSKEY"
	PASSWORD                                 Keyword = "PASSWORD"
	PAST                                     Keyword = "PAST"
	PATH                                     Keyword = "PATH"
	PATTERN                                  Keyword = "PATTERN"
	PCTFREE                                  Keyword = "PCTFREE"
	PER                                      Keyword = "PER"
	PERCENT                                  Keyword = "PERCENT"
	PERCENTILE_CONT                          Keyword = "PERCENTILE_CONT"
	PERCENTILE_DISC                          Keyword = "PERCENTILE_DISC"
	PERCENT_RANK                             Keyword = "PERCENT_RANK"
	PERIOD                                   Keyword = "PERIOD"
	PERMISSIVE                               Keyword = "PERMISSIVE"
	PERSISTENT                               Keyword = "PERSISTENT"
	PIVOT                                    Keyword = "PIVOT"
	PLACING                                  Keyword = "PLACING"
	PLAIN                                    Keyword = "PLAIN"
	PLAN                                     Keyword = "PLAN"
	PLANS                                    Keyword = "PLANS"
	POINT                                    Keyword = "POINT"
	POLICY                                   Keyword = "POLICY"
	POLYGON                                  Keyword = "POLYGON"
	POOL                                     Keyword = "POOL"
	PORTION                                  Keyword = "PORTION"
	POSITION                                 Keyword = "POSITION"
	POSITION_REGEX                           Keyword = "POSITION_REGEX"
	POWER                                    Keyword = "POWER"
	PRAGMA                                   Keyword = "PRAGMA"
	PRECEDES                                 Keyword = "PRECEDES"
	PRECEDING                                Keyword = "PRECEDING"
	PRECISION                                Keyword = "PRECISION"
	PREFERRED                                Keyword = "PREFERRED"
	PREPARE                                  Keyword = "PREPARE"
	PRESERVE                                 Keyword = "PRESERVE"
	PRESET                                   Keyword = "PRESET"
	PREWHERE                                 Keyword = "PREWHERE"
	PRIMARY                                  Keyword = "PRIMARY"
	PRINT                                    Keyword = "PRINT"
	PRIOR                                    Keyword = "PRIOR"
	PRIVILEGES                               Keyword = "PRIVILEGES"
	PROCEDURE                                Keyword = "PROCEDURE"
	PROFILE                                  Keyword = "PROFILE"
	PROGRAM                                  Keyword = "PROGRAM"
	PROJECTION                               Keyword = "PROJECTION"
	PUBLIC                                   Keyword = "PUBLIC"
	PURCHASE                                 Keyword = "PURCHASE"
	PURGE                                    Keyword = "PURGE"
	QUALIFY                                  Keyword = "QUALIFY"
	QUARTER                                  Keyword = "QUARTER"
	QUERIES                                  Keyword = "QUERIES"
	QUERY                                    Keyword = "QUERY"
	QUOTE                                    Keyword = "QUOTE"
	RAISE                                    Keyword = "RAISE"
	RAISERROR                                Keyword = "RAISERROR"
	RANGE                                    Keyword = "RANGE"
	RANK                                     Keyword = "RANK"
	RAW                                      Keyword = "RAW"
	RCFILE                                   Keyword = "RCFILE"
	READ                                     Keyword = "READ"
	READS                                    Keyword = "READS"
	READ_ONLY                                Keyword = "READ_ONLY"
	REAL                                     Keyword = "REAL"
	RECEIVE                                  Keyword = "RECEIVE"
	RECLUSTER                                Keyword = "RECLUSTER"
	RECURSIVE                                Keyword = "RECURSIVE"
	REF                                      Keyword = "REF"
	REFERENCES                               Keyword = "REFERENCES"
	REFERENCING                              Keyword = "REFERENCING"
	REFRESH                                  Keyword = "REFRESH"
	REFRESH_MODE                             Keyword = "REFRESH_MODE"
	REGCLASS                                 Keyword = "REGCLASS"
	REGEXP                                   Keyword = "REGEXP"
	REGION                                   Keyword = "REGION"
	REGR_AVGX                                Keyword = "REGR_AVGX"
	REGR_AVGY                                Keyword = "REGR_AVGY"
	REGR_COUNT                               Keyword = "REGR_COUNT"
	REGR_INTERCEPT                           Keyword = "REGR_INTERCEPT"
	REGR_R2                                  Keyword = "REGR_R2"
	REGR_SLOPE                               Keyword = "REGR_SLOPE"
	REGR_SXX                                 Keyword = "REGR_SXX"
	REGR_SXY                                 Keyword = "REGR_SXY"
	REGR_SYY                                 Keyword = "REGR_SYY"
	REINDEX                                  Keyword = "REINDEX"
	RELATIVE                                 Keyword = "RELATIVE"
	RELAY                                    Keyword = "RELAY"
	RELEASE                                  Keyword = "RELEASE"
	RELEASES                                 Keyword = "RELEASES"
	REMAINDER                                Keyword = "REMAINDER"
	REMOTE                                   Keyword = "REMOTE"
	REMOVE                                   Keyword = "REMOVE"
	REMOVEQUOTES                             Keyword = "REMOVEQUOTES"
	RENAME                                   Keyword = "RENAME"
	REORG                                    Keyword = "REORG"
	REPAIR                                   Keyword = "REPAIR"
	REPEATABLE                               Keyword = "REPEATABLE"
	REPLACE                                  Keyword = "REPLACE"
	REPLACE_INVALID_CHARACTERS               Keyword = "REPLACE_INVALID_CHARACTERS"
	REPLICA                                  Keyword = "REPLICA"
	REPLICATE                                Keyword = "REPLICATE"
	REPLICATION                              Keyword = "REPLICATION"
	REQUIRE                                  Keyword = "REQUIRE"
	RESET                                    Keyword = "RESET"
	RESOLVE                                  Keyword = "RESOLVE"
	RESOURCE                                 Keyword = "RESOURCE"
	RESPECT                                  Keyword = "RESPECT"
	RESTART                                  Keyword = "RESTART"
	RESTRICT                                 Keyword = "RESTRICT"
	RESTRICTED                               Keyword = "RESTRICTED"
	RESTRICTIONS                             Keyword = "RESTRICTIONS"
	RESTRICTIVE                              Keyword = "RESTRICTIVE"
	RESULT                                   Keyword = "RESULT"
	RESULTSET                                Keyword = "RESULTSET"
	RESUME                                   Keyword = "RESUME"
	RETAIN                                   Keyword = "RETAIN"
	RETURN                                   Keyword = "RETURN"
	RETURNING                                Keyword = "RETURNING"
	RETURNS                                  Keyword = "RETURNS"
	REVOKE                                   Keyword = "REVOKE"
	RIGHT                                    Keyword = "RIGHT"
	RIGHTARG                                 Keyword = "RIGHTARG"
	RLIKE                                    Keyword = "RLIKE"
	RM                                       Keyword = "RM"
	ROLE                                     Keyword = "ROLE"
	ROLES                                    Keyword = "ROLES"
	ROLLBACK                                 Keyword = "ROLLBACK"
	ROLLUP                                   Keyword = "ROLLUP"
	ROOT                                     Keyword = "ROOT"
	ROW                                      Keyword = "ROW"
	ROWGROUPSIZE                             Keyword = "ROWGROUPSIZE"
	ROWID                                    Keyword = "ROWID"
	ROWS                                     Keyword = "ROWS"
	ROW_FORMAT                               Keyword = "ROW_FORMAT"
	ROW_NUMBER                               Keyword = "ROW_NUMBER"
	RULE                                     Keyword = "RULE"
	RUN                                      Keyword = "RUN"
	SAFE                                     Keyword = "SAFE"
	SAFE_CAST                                Keyword = "SAFE_CAST"
	SAMPLE                                   Keyword = "SAMPLE"
	SAVEPOINT                                Keyword = "SAVEPOINT"
	SCHEMA                                   Keyword = "SCHEMA"
	SCHEMAS                                  Keyword = "SCHEMAS"
	SCOPE                                    Keyword = "SCOPE"
	SCROLL                                   Keyword = "SCROLL"
	SEARCH                                   Keyword = "SEARCH"
	SECOND                                   Keyword = "SECOND"
	SECONDARY                                Keyword = "SECONDARY"
	SECONDARY_ENGINE_ATTRIBUTE               Keyword = "SECONDARY_ENGINE_ATTRIBUTE"
	SECONDS                                  Keyword = "SECONDS"
	SECRET                                   Keyword = "SECRET"
	SECURE                                   Keyword = "SECURE"
	SECURITY                                 Keyword = "SECURITY"
	SEED                                     Keyword = "SEED"
	SELECT                                   Keyword = "SELECT"
	SEMANTIC_VIEW                            Keyword = "SEMANTIC_VIEW"
	SEMI                                     Keyword = "SEMI"
	SEND                                     Keyword = "SEND"
	SENSITIVE                                Keyword = "SENSITIVE"
	SEPARATOR                                Keyword = "SEPARATOR"
	SEQUENCE                                 Keyword = "SEQUENCE"
	SEQUENCEFILE                             Keyword = "SEQUENCEFILE"
	SEQUENCES                                Keyword = "SEQUENCES"
	SERDE                                    Keyword = "SERDE"
	SERDEPROPERTIES                          Keyword = "SERDEPROPERTIES"
	SERIALIZABLE                             Keyword = "SERIALIZABLE"
	SERVER                                   Keyword = "SERVER"
	SERVICE                                  Keyword = "SERVICE"
	SESSION                                  Keyword = "SESSION"
	SESSION_USER                             Keyword = "SESSION_USER"
	SET                                      Keyword = "SET"
	SETERROR                                 Keyword = "SETERROR"
	SETOF                                    Keyword = "SETOF"
	SETS                                     Keyword = "SETS"
	SETTINGS                                 Keyword = "SETTINGS"
	SHARE                                    Keyword = "SHARE"
	SHARED                                   Keyword = "SHARED"
	SHARING                                  Keyword = "SHARING"
	SHOW                                     Keyword = "SHOW"
	SIGNED                                   Keyword = "SIGNED"
	SIMILAR                                  Keyword = "SIMILAR"
	SIMPLE                                   Keyword = "SIMPLE"
	SIZE                                     Keyword = "SIZE"
	SKIP                                     Keyword = "SKIP"
	SLOW                                     Keyword = "SLOW"
	SMALLINT                                 Keyword = "SMALLINT"
	SNAPSHOT                                 Keyword = "SNAPSHOT"
	SOME                                     Keyword = "SOME"
	SORT                                     Keyword = "SORT"
	SORTED                                   Keyword = "SORTED"
	SORTKEY                                  Keyword = "SORTKEY"
	SOURCE                                   Keyword = "SOURCE"
	SPATIAL                                  Keyword = "SPATIAL"
	SPECIFIC                                 Keyword = "SPECIFIC"
	SPECIFICTYPE                             Keyword = "SPECIFICTYPE"
	SPGIST                                   Keyword = "SPGIST"
	SQL                                      Keyword = "SQL"
	SQLEXCEPTION                             Keyword = "SQLEXCEPTION"
	SQLSTATE                                 Keyword = "SQLSTATE"
	SQLWARNING                               Keyword = "SQLWARNING"
	SQL_BIG_RESULT                           Keyword = "SQL_BIG_RESULT"
	SQL_BUFFER_RESULT                        Keyword = "SQL_BUFFER_RESULT"
	SQL_CALC_FOUND_ROWS                      Keyword = "SQL_CALC_FOUND_ROWS"
	SQL_NO_CACHE                             Keyword = "SQL_NO_CACHE"
	SQL_SMALL_RESULT                         Keyword = "SQL_SMALL_RESULT"
	SQRT                                     Keyword = "SQRT"
	SRID                                     Keyword = "SRID"
	STABLE                                   Keyword = "STABLE"
	STAGE                                    Keyword = "STAGE"
	START                                    Keyword = "START"
	STARTS                                   Keyword = "STARTS"
	STATEMENT                                Keyword = "STATEMENT"
	STATIC                                   Keyword = "STATIC"
	STATISTICS                               Keyword = "STATISTICS"
	STATS_AUTO_RECALC                        Keyword = "STATS_AUTO_RECALC"
	STATS_PERSISTENT                         Keyword = "STATS_PERSISTENT"
	STATS_SAMPLE_PAGES                       Keyword = "STATS_SAMPLE_PAGES"
	STATUPDATE                               Keyword = "STATUPDATE"
	STATUS                                   Keyword = "STATUS"
	STDDEV_POP                               Keyword = "STDDEV_POP"
	STDDEV_SAMP                              Keyword = "STDDEV_SAMP"
	STDIN                                    Keyword = "STDIN"
	STDOUT                                   Keyword = "STDOUT"
	STEP                                     Keyword = "STEP"
	STORAGE                                  Keyword = "STORAGE"
	STORAGE_INTEGRATION                      Keyword = "STORAGE_INTEGRATION"
	STORAGE_SERIALIZATION_POLICY             Keyword = "STORAGE_SERIALIZATION_POLICY"
	STORED                                   Keyword = "STORED"
	STRAIGHT_JOIN                            Keyword = "STRAIGHT_JOIN"
	STREAM                                   Keyword = "STREAM"
	STRICT                                   Keyword = "STRICT"
	STRING                                   Keyword = "STRING"
	STRUCT                                   Keyword = "STRUCT"
	SUBMULTISET                              Keyword = "SUBMULTISET"
	SUBSCRIPT                                Keyword = "SUBSCRIPT"
	SUBSTR                                   Keyword = "SUBSTR"
	SUBSTRING                                Keyword = "SUBSTRING"
	SUBSTRING_REGEX                          Keyword = "SUBSTRING_REGEX"
	SUBTYPE                                  Keyword = "SUBTYPE"
	SUBTYPE_DIFF                             Keyword = "SUBTYPE_DIFF"
	SUBTYPE_OPCLASS                          Keyword = "SUBTYPE_OPCLASS"
	SUCCEEDS                                 Keyword = "SUCCEEDS"
	SUM                                      Keyword = "SUM"
	SUPER                                    Keyword = "SUPER"
	SUPERUSER                                Keyword = "SUPERUSER"
	SUPPORT                                  Keyword = "SUPPORT"
	SUSPEND                                  Keyword = "SUSPEND"
	SWAP                                     Keyword = "SWAP"
	SYMMETRIC                                Keyword = "SYMMETRIC"
	SYNC                                     Keyword = "SYNC"
	SYNONYM                                  Keyword = "SYNONYM"
	SYSTEM                                   Keyword = "SYSTEM"
	SYSTEM_TIME                              Keyword = "SYSTEM_TIME"
	SYSTEM_USER                              Keyword = "SYSTEM_USER"
	TABLE                                    Keyword = "TABLE"
	TABLES                                   Keyword = "TABLES"
	TABLESAMPLE                              Keyword = "TABLESAMPLE"
	TABLESPACE                               Keyword = "TABLESPACE"
	TAG                                      Keyword = "TAG"
	TARGET                                   Keyword = "TARGET"
	TARGET_LAG                               Keyword = "TARGET_LAG"
	TASK                                     Keyword = "TASK"
	TBLPROPERTIES                            Keyword = "TBLPROPERTIES"
	TEMP                                     Keyword = "TEMP"
	TEMPORARY                                Keyword = "TEMPORARY"
	TEMPTABLE                                Keyword = "TEMPTABLE"
	TERMINATED                               Keyword = "TERMINATED"
	TERSE                                    Keyword = "TERSE"
	TEXT                                     Keyword = "TEXT"
	TEXTFILE                                 Keyword = "TEXTFILE"
	THEN                                     Keyword = "THEN"
	THROW                                    Keyword = "THROW"
	TIES                                     Keyword = "TIES"
	TIME                                     Keyword = "TIME"
	TIMEFORMAT                               Keyword = "TIMEFORMAT"
	TIMESTAMP                                Keyword = "TIMESTAMP"
	TIMESTAMPTZ                              Keyword = "TIMESTAMPTZ"
	TIMESTAMP_NTZ                            Keyword = "TIMESTAMP_NTZ"
	TIMETZ                                   Keyword = "TIMETZ"
	TIMEZONE                                 Keyword = "TIMEZONE"
	TIMEZONE_ABBR                            Keyword = "TIMEZONE_ABBR"
	TIMEZONE_HOUR                            Keyword = "TIMEZONE_HOUR"
	TIMEZONE_MINUTE                          Keyword = "TIMEZONE_MINUTE"
	TIMEZONE_REGION                          Keyword = "TIMEZONE_REGION"
	TINYBLOB                                 Keyword = "TINYBLOB"
	TINYINT                                  Keyword = "TINYINT"
	TINYTEXT                                 Keyword = "TINYTEXT"
	TO                                       Keyword = "TO"
	TOP                                      Keyword = "TOP"
	TOTALS                                   Keyword = "TOTALS"
	TOTP                                     Keyword = "TOTP"
	TRACE                                    Keyword = "TRACE"
	TRAILING                                 Keyword = "TRAILING"
	TRAN                                     Keyword = "TRAN"
	TRANSACTION                              Keyword = "TRANSACTION"
	TRANSIENT                                Keyword = "TRANSIENT"
	TRANSLATE                                Keyword = "TRANSLATE"
	TRANSLATE_REGEX                          Keyword = "TRANSLATE_REGEX"
	TRANSLATION                              Keyword = "TRANSLATION"
	TREAT                                    Keyword = "TREAT"
	TREE                                     Keyword = "TREE"
	TRIGGER                                  Keyword = "TRIGGER"
	TRIM                                     Keyword = "TRIM"
	TRIM_ARRAY                               Keyword = "TRIM_ARRAY"
	TRUE                                     Keyword = "TRUE"
	TRUNCATE                                 Keyword = "TRUNCATE"
	TRUNCATECOLUMNS                          Keyword = "TRUNCATECOLUMNS"
	TRY                                      Keyword = "TRY"
	TRY_CAST                                 Keyword = "TRY_CAST"
	TRY_CONVERT                              Keyword = "TRY_CONVERT"
	TSQUERY                                  Keyword = "TSQUERY"
	TSVECTOR                                 Keyword = "TSVECTOR"
	TUPLE                                    Keyword = "TUPLE"
	TYPE                                     Keyword = "TYPE"
	TYPMOD_IN                                Keyword = "TYPMOD_IN"
	TYPMOD_OUT                               Keyword = "TYPMOD_OUT"
	UBIGINT                                  Keyword = "UBIGINT"
	UESCAPE                                  Keyword = "UESCAPE"
	UHUGEINT                                 Keyword = "UHUGEINT"
	UINT128                                  Keyword = "UINT128"
	UINT16                                   Keyword = "UINT16"
	UINT256                                  Keyword = "UINT256"
	UINT32                                   Keyword = "UINT32"
	UINT64                                   Keyword = "UINT64"
	UINT8                                    Keyword = "UINT8"
	UNBOUNDED                                Keyword = "UNBOUNDED"
	UNCACHE                                  Keyword = "UNCACHE"
	UNCOMMITTED                              Keyword = "UNCOMMITTED"
	UNDEFINED                                Keyword = "UNDEFINED"
	UNFREEZE                                 Keyword = "UNFREEZE"
	UNION                                    Keyword = "UNION"
	UNIQUE                                   Keyword = "UNIQUE"
	UNKNOWN                                  Keyword = "UNKNOWN"
	UNLISTEN                                 Keyword = "UNLISTEN"
	UNLOAD                                   Keyword = "UNLOAD"
	UNLOCK                                   Keyword = "UNLOCK"
	UNLOGGED                                 Keyword = "UNLOGGED"
	UNMATCHED                                Keyword = "UNMATCHED"
	UNNEST                                   Keyword = "UNNEST"
	UNPIVOT                                  Keyword = "UNPIVOT"
	UNSAFE                                   Keyword = "UNSAFE"
	UNSET                                    Keyword = "UNSET"
	UNSIGNED                                 Keyword = "UNSIGNED"
	UNTIL                                    Keyword = "UNTIL"
	UPDATE                                   Keyword = "UPDATE"
	UPPER                                    Keyword = "UPPER"
	URL                                      Keyword = "URL"
	USAGE                                    Keyword = "USAGE"
	USE                                      Keyword = "USE"
	USER                                     Keyword = "USER"
	USER_RESOURCES                           Keyword = "USER_RESOURCES"
	USING                                    Keyword = "USING"
	USMALLINT                                Keyword = "USMALLINT"
	UTINYINT                                 Keyword = "UTINYINT"
	UUID                                     Keyword = "UUID"
	VACUUM                                   Keyword = "VACUUM"
	VALID                                    Keyword = "VALID"
	VALIDATE                                 Keyword = "VALIDATE"
	VALIDATION_MODE                          Keyword = "VALIDATION_MODE"
	VALUE                                    Keyword = "VALUE"
	VALUES                                   Keyword = "VALUES"
	VALUE_OF                                 Keyword = "VALUE_OF"
	VARBINARY                                Keyword = "VARBINARY"
	VARBIT                                   Keyword = "VARBIT"
	VARCHAR                                  Keyword = "VARCHAR"
	VARCHAR2                                 Keyword = "VARCHAR2"
	VARIABLE                                 Keyword = "VARIABLE"
	VARIABLES                                Keyword = "VARIABLES"
	VARYING                                  Keyword = "VARYING"
	VAR_POP                                  Keyword = "VAR_POP"
	VAR_SAMP                                 Keyword = "VAR_SAMP"
	VERBOSE                                  Keyword = "VERBOSE"
	VERSION                                  Keyword = "VERSION"
	VERSIONING                               Keyword = "VERSIONING"
	VERSIONS                                 Keyword = "VERSIONS"
	VIEW                                     Keyword = "VIEW"
	VIEWS                                    Keyword = "VIEWS"
	VIRTUAL                                  Keyword = "VIRTUAL"
	VOLATILE                                 Keyword = "VOLATILE"
	VOLUME                                   Keyword = "VOLUME"
	WAITFOR                                  Keyword = "WAITFOR"
	WAREHOUSE                                Keyword = "WAREHOUSE"
	WAREHOUSES                               Keyword = "WAREHOUSES"
	WEEK                                     Keyword = "WEEK"
	WEEKS                                    Keyword = "WEEKS"
	WHEN                                     Keyword = "WHEN"
	WHENEVER                                 Keyword = "WHENEVER"
	WHERE                                    Keyword = "WHERE"
	WHILE                                    Keyword = "WHILE"
	WIDTH_BUCKET                             Keyword = "WIDTH_BUCKET"
	WINDOW                                   Keyword = "WINDOW"
	WITH                                     Keyword = "WITH"
	WITHIN                                   Keyword = "WITHIN"
	WITHOUT                                  Keyword = "WITHOUT"
	WITHOUT_ARRAY_WRAPPER                    Keyword = "WITHOUT_ARRAY_WRAPPER"
	WORK                                     Keyword = "WORK"
	WORKLOAD_IDENTITY                        Keyword = "WORKLOAD_IDENTITY"
	WRAPPER                                  Keyword = "WRAPPER"
	WRITE                                    Keyword = "WRITE"
	XML                                      Keyword = "XML"
	XMLNAMESPACES                            Keyword = "XMLNAMESPACES"
	XMLTABLE                                 Keyword = "XMLTABLE"
	XOR                                      Keyword = "XOR"
	YEAR                                     Keyword = "YEAR"
	YEARS                                    Keyword = "YEARS"
	YES                                      Keyword = "YES"
	ZONE                                     Keyword = "ZONE"
	ZORDER                                   Keyword = "ZORDER"
	ZSTD                                     Keyword = "ZSTD"
)

// AllKeywords contains all SQL keyword constants for iteration
var AllKeywords = []Keyword{
	ABORT, ABS, ABSENT, ABSOLUTE, ACCEPTANYDATE, ACCEPTINVCHARS, ACCESS, ACCOUNT, ACTION, ADD,
	ADDQUOTES, ADMIN, AFTER, AGAINST, AGGREGATE, AGGREGATION, ALERT, ALGORITHM, ALIAS, ALIGNMENT,
	ALL, ALLOCATE, ALLOWOVERWRITE, ALTER, ALWAYS, ANALYZE, AND, ANTI, ANY, APPLICATION, APPLY,
	APPLYBUDGET, ARCHIVE, ARE, ARRAY, ARRAY_MAX_CARDINALITY, AS, ASC, ASENSITIVE, ASOF, ASSERT,
	ASYMMETRIC, AT, ATOMIC, ATTACH, AUDIT, AUTHENTICATION, AUTHORIZATION, AUTHORIZATIONS, AUTO,
	AUTOEXTEND_SIZE, AUTOINCREMENT, AUTO_INCREMENT, AVG, AVG_ROW_LENGTH, AVRO, BACKUP, BACKWARD,
	BASE64, BASE_LOCATION, BEFORE, BEGIN, BEGIN_FRAME, BEGIN_PARTITION, BERNOULLI, BETWEEN,
	BIGDECIMAL, BIGINT, BIGNUMERIC, BINARY, BIND, BINDING, BIT, BLANKSASNULL, BLOB, BLOCK, BLOOM,
	BLOOMFILTER, BOOL, BOOLEAN, BOOST, BOTH, BOX, BRIN, BROWSE, BTREE, BUCKET, BUCKETS, BY,
	BYPASSRLS, BYTEA, BYTES, BZIP2, CACHE, CALL, CALLED, CANONICAL, CARDINALITY, CASCADE,
	CASCADED, CASE, CASES, CAST, CATALOG, CATALOG_SYNC, CATALOG_SYNC_NAMESPACE_FLATTEN_DELIMITER,
	CATALOG_SYNC_NAMESPACE_MODE, CATCH, CATEGORY, CEIL, CEILING, CENTURY, CHAIN, CHANGE, CHANGES,
	CHANGE_TRACKING, CHANNEL, CHAR, CHARACTER, CHARACTERISTICS, CHARACTERS, CHARACTER_LENGTH,
	CHARSET, CHAR_LENGTH, CHECK, CHECKSUM, CIRCLE, CLASS, CLEANPATH, CLEAR, CLOB, CLONE, CLOSE,
	CLUSTER, CLUSTERED, CLUSTERING, COALESCE, COLLATABLE, COLLATE, COLLATION, COLLECT, COLLECTION,
	COLUMN, COLUMNS, COLUMNSTORE, COMMENT, COMMIT, COMMITTED, COMMUTATOR, COMPATIBLE, COMPRESS,
	COMPRESSION, COMPUPDATE, COMPUTE, CONCURRENTLY, CONDITION, CONFLICT, CONNECT, CONNECTION,
	CONNECTOR, CONNECT_BY_ROOT, CONSTRAINT, CONTACT, CONTAINS, CONTINUE, CONVERT, COPY,
	COPY_OPTIONS, CORR, CORRESPONDING, COUNT, COVAR_POP, COVAR_SAMP, CREATE, CREATEDB, CREATEROLE,
	CREDENTIALS, CROSS, CSV, CUBE, CUME_DIST, CURRENT, CURRENT_CATALOG, CURRENT_DATE,
	CURRENT_DEFAULT_TRANSFORM_GROUP, CURRENT_PATH, CURRENT_ROLE, CURRENT_ROW, CURRENT_SCHEMA,
	CURRENT_TIME, CURRENT_TIMESTAMP, CURRENT_TRANSFORM_GROUP_FOR_TYPE, CURRENT_USER, CURSOR, CYCLE,
	DATA, DATABASE, DATABASES, DATA_RETENTION_TIME_IN_DAYS, DATE, DATE32, DATEFORMAT, DATETIME,
	DATETIME64, DAY, DAYOFWEEK, DAYOFYEAR, DAYS, DCPROPERTIES, DEALLOCATE, DEC, DECADE, DECIMAL,
	DECLARE, DEDUPLICATE, DEFAULT, DEFAULTS, DEFAULT_DDL_COLLATION, DEFAULT_MFA_METHOD,
	DEFAULT_SECONDARY_ROLES, DEFERRABLE, DEFERRED, DEFINE, DEFINED, DEFINER, DELAY, DELAYED,
	DELAY_KEY_WRITE, DELEGATED, DELETE, DELIMITED, DELIMITER, DELTA, DENSE_RANK, DENY, DEREF, DESC,
	DESCRIBE, DETACH, DETAIL, DETERMINISTIC, DIMENSIONS, DIRECTORY, DISABLE, DISCARD, DISCONNECT,
	DISTINCT, DISTINCTROW, DISTKEY, DISTRIBUTE, DISTSTYLE, DIV, DO, DOMAIN, DOUBLE, DOW,
	DOWNSTREAM, DOY, DROP, DRY, DUO, DUPLICATE, DYNAMIC, EACH, ELEMENT, ELEMENTS, ELSE, ELSEIF,
	EMPTY, EMPTYASNULL, ENABLE, ENABLE_SCHEMA_EVOLUTION, ENCODING, ENCRYPTED, ENCRYPTION, END,
	END_EXEC, ENDPOINT, END_FRAME, END_PARTITION, ENFORCED, ENGINE, ENGINE_ATTRIBUTE, ENROLL,
	ENUM, ENUM16, ENUM8, EPHEMERAL, EPOCH, EQUALS, ERROR, ESCAPE, ESCAPED, ESTIMATE, EVEN, EVENT,
	EVERY, EVOLVE, EXCEPT, EXCEPTION, EXCHANGE, EXCLUDE, EXCLUDING, EXCLUSIVE, EXEC, EXECUTE,
	EXECUTION, EXISTS, EXP, EXPANSION, EXPLAIN, EXPLICIT, EXPORT, EXTEND, EXTENDED, EXTENSION,
	EXTERNAL, EXTERNAL_VOLUME, EXTRACT, FACTS, FAIL, FAILOVER, FALSE, FAMILY, FETCH, FIELDS, FILE,
	FILES, FILE_FORMAT, FILL, FILTER, FINAL, FIRST, FIRST_VALUE, FIXEDSTRING, FIXEDWIDTH, FLATTEN,
	FLOAT, FLOAT32, FLOAT4, FLOAT64, FLOAT8, FLOOR, FLUSH, FN, FOLLOWING, FOR, FORCE, FORCE_NOT_NULL,
	FORCE_NULL, FORCE_QUOTE, FOREIGN, FORMAT, FORMATTED, FORWARD, FRAME_ROW, FREE, FREEZE, FROM,
	FSCK, FULFILLMENT, FULL, FULLTEXT, FUNCTION, FUNCTIONS, FUSION, FUTURE, GB, GENERAL, GENERATE,
	GENERATED, GEOGRAPHY, GET, GIN, GIST, GLOBAL, GRANT, GRANTED, GRANTS, GRAPHVIZ, GROUP, GROUPING,
	GROUPS, GZIP, HARD, HASH, HASHES, HAVING, HEADER, HEAP, HIGH_PRIORITY, HISTORY, HIVEVAR, HOLD, HOSTS,
	HOUR, HOURS, HUGEINT, IAM_ROLE, ICEBERG, ID, IDENTIFIED, IDENTITY, IDENTITY_INSERT, IF, IGNORE,
	IGNOREHEADER, ILIKE, IMMEDIATE, IMMUTABLE, IMPORT, IMPORTED, IN, INCLUDE, INCLUDE_NULL_VALUES,
	INCLUDING, INCREMENT, INCREMENTAL, INDEX, INDICATOR, INHERIT, INHERITS, INITIALIZE, INITIALLY,
	INNER, INOUT, INPATH, INPLACE, INPUT, INPUTFORMAT, INSENSITIVE, INSERT, INSERT_METHOD, INSTALL,
	INSTANT, INSTEAD, INT, INT128, INT16, INT2, INT256, INT32, INT4, INT64, INT8, INTEGER,
	INTEGRATION, INTERNALLENGTH, INTERPOLATE, INTERSECT, INTERSECTION, INTERVAL, INTO, INVISIBLE,
	INVOKER, IO, IS, ISODOW, ISOLATION, ISOWEEK, ISOYEAR, ITEMS, JAR, JOIN, JSON, JSONB, JSONFILE,
	JSON_TABLE, JULIAN, KEY, KEYS, KEY_BLOCK_SIZE, KILL, LAG, LAMBDA, LANGUAGE, LARGE, LAST,
	LAST_VALUE, LATERAL, LEAD, LEADING, LEAKPROOF, LEAST, LEFT, LEFTARG, LEVEL, LIFECYCLE, LIKE,
	LIKE_REGEX, LIMIT, LINE, LINES, LIST, LISTEN, LISTING, LN, LOAD, LOCAL, LOCALTIME,
	LOCALTIMESTAMP, LOCATION, LOCK, LOCKED, LOG, LOGIN, LOGS, LONG, LONGBLOB, LONGTEXT,
	LOWCARDINALITY, LOWER, LOW_PRIORITY, LS, LSEG, MACRO, MAIN, MANAGE, MANAGED, MANAGEDLOCATION,
	MANIFEST, MAP, MASKING, MATCH, MATCHED, MATCHES, MATCH_CONDITION, MATCH_RECOGNIZE, MATERIALIZE,
	MATERIALIZED, MAX, MAXFILESIZE, MAXVALUE, MAX_DATA_EXTENSION_TIME_IN_DAYS, MAX_ROWS, MB,
	MEASURES, MEDIUMBLOB, MEDIUMINT, MEDIUMTEXT, MEMBER, MERGE, MERGES, MESSAGE, METADATA, METHOD,
	METRIC, METRICS, MFA, MICROSECOND, MICROSECONDS, MILLENIUM, MILLENNIUM, MILLISECOND,
	MILLISECONDS, MIN, MINUS, MINUTE, MINUTES, MINVALUE, MIN_ROWS, MOD, MODE, MODIFIES, MODIFY,
	MODULE, MODULUS, MONITOR, MONTH, MONTHS, MSCK, MULTIRANGE_TYPE_NAME, MULTISET, MUTATION, NAME,
	NAMES, NANOSECOND, NANOSECONDS, NATIONAL, NATURAL, NCHAR, NCLOB, NEGATOR, NEST, NESTED,
	NETWORK, NEW, NEXT, NFC, NFD, NFKC, NFKD, NO, NOBYPASSRLS, NOCOMPRESS, NOCREATEDB, NOCREATEROLE,
	NOCYCLE, NOINHERIT, NOLOGIN, NONE, NOORDER, NOREPLICATION, NORMALIZE, NORMALIZED, NOSCAN,
	NOSUPERUSER, NOT, NOTHING, NOTIFY, NOTNULL, NOWAIT, NO_WRITE_TO_BINLOG, NTH_VALUE, NTILE, NULL,
	NULLABLE, NULLIF, NULLS, NUMBER, NUMERIC, NVARCHAR, OBJECT, OBJECTS, OCCURRENCES_REGEX, OCTETS,
	OCTET_LENGTH, OF, OFF, OFFSET, OFFSETS, OLD, OMIT, ON, ONE, ONLY, ON_CREATE, ON_SCHEDULE, OPEN,
	OPENJSON, OPERATE, OPERATOR, OPTIMIZATION, OPTIMIZE, OPTIMIZED, OPTIMIZER_COSTS, OPTION,
	OPTIONS, OR, ORC, ORDER, ORDINALITY, ORGANIZATION, OTHER, OTP, OUT, OUTER, OUTPUT, OUTPUTFORMAT,
	OVER, OVERFLOW, OVERLAPS, OVERLAY, OVERRIDE, OVERWRITE, OWNED, OWNER, OWNERSHIP, PACKAGE,
	PACKAGES, PACK_KEYS, PARALLEL, PARAMETER, PARQUET, PART, PARTIAL, PARTITION, PARTITIONED,
	PARTITIONS, PASSEDBYVALUE, PASSING, PASSKEY, PASSWORD, PAST, PATH, PATTERN, PCTFREE, PER,
	PERCENT, PERCENTILE_CONT, PERCENTILE_DISC, PERCENT_RANK, PERIOD, PERMISSIVE, PERSISTENT,
	PIVOT, PLACING, PLAIN, PLAN, PLANS, POINT, POLICY, POLYGON, POOL, PORTION, POSITION,
	POSITION_REGEX, POWER, PRAGMA, PRECEDES, PRECEDING, PRECISION, PREFERRED, PREPARE, PRESERVE,
	PRESET, PREWHERE, PRIMARY, PRINT, PRIOR, PRIVILEGES, PROCEDURE, PROFILE, PROGRAM, PROJECTION,
	PUBLIC, PURCHASE, PURGE, QUALIFY, QUARTER, QUERIES, QUERY, QUOTE, RAISE, RAISERROR, RANGE,
	RANK, RAW, RCFILE, READ, READS, READ_ONLY, REAL, RECEIVE, RECLUSTER, RECURSIVE, REF, REFERENCES,
	REFERENCING, REFRESH, REFRESH_MODE, REGCLASS, REGEXP, REGION, REGR_AVGX, REGR_AVGY, REGR_COUNT,
	REGR_INTERCEPT, REGR_R2, REGR_SLOPE, REGR_SXX, REGR_SXY, REGR_SYY, REINDEX, RELATIVE, RELAY,
	RELEASE, RELEASES, REMAINDER, REMOTE, REMOVE, REMOVEQUOTES, RENAME, REORG, REPAIR, REPEATABLE,
	REPLACE, REPLACE_INVALID_CHARACTERS, REPLICA, REPLICATE, REPLICATION, REQUIRE, RESET, RESOLVE,
	RESOURCE, RESPECT, RESTART, RESTRICT, RESTRICTED, RESTRICTIONS, RESTRICTIVE, RESULT,
	RESULTSET, RESUME, RETAIN, RETURN, RETURNING, RETURNS, REVOKE, RIGHT, RIGHTARG, RLIKE, RM,
	ROLE, ROLES, ROLLBACK, ROLLUP, ROOT, ROW, ROWGROUPSIZE, ROWID, ROWS, ROW_FORMAT, ROW_NUMBER,
	RULE, RUN, SAFE, SAFE_CAST, SAMPLE, SAVEPOINT, SCHEMA, SCHEMAS, SCOPE, SCROLL, SEARCH, SECOND,
	SECONDARY, SECONDARY_ENGINE_ATTRIBUTE, SECONDS, SECRET, SECURE, SECURITY, SEED, SELECT,
	SEMANTIC_VIEW, SEMI, SEND, SENSITIVE, SEPARATOR, SEQUENCE, SEQUENCEFILE, SEQUENCES, SERDE,
	SERDEPROPERTIES, SERIALIZABLE, SERVER, SERVICE, SESSION, SESSION_USER, SET, SETERROR, SETOF,
	SETS, SETTINGS, SHARE, SHARED, SHARING, SHOW, SIGNED, SIMILAR, SIMPLE, SIZE, SKIP, SLOW,
	SMALLINT, SNAPSHOT, SOME, SORT, SORTED, SORTKEY, SOURCE, SPATIAL, SPECIFIC, SPECIFICTYPE,
	SPGIST, SQL, SQLEXCEPTION, SQLSTATE, SQLWARNING, SQL_BIG_RESULT, SQL_BUFFER_RESULT,
	SQL_CALC_FOUND_ROWS, SQL_NO_CACHE, SQL_SMALL_RESULT, SQRT, SRID, STABLE, STAGE, START, STARTS,
	STATEMENT, STATIC, STATISTICS, STATS_AUTO_RECALC, STATS_PERSISTENT, STATS_SAMPLE_PAGES,
	STATUPDATE, STATUS, STDDEV_POP, STDDEV_SAMP, STDIN, STDOUT, STEP, STORAGE, STORAGE_INTEGRATION,
	STORAGE_SERIALIZATION_POLICY, STORED, STRAIGHT_JOIN, STREAM, STRICT, STRING, STRUCT,
	SUBMULTISET, SUBSCRIPT, SUBSTR, SUBSTRING, SUBSTRING_REGEX, SUBTYPE, SUBTYPE_DIFF,
	SUBTYPE_OPCLASS, SUCCEEDS, SUM, SUPER, SUPERUSER, SUPPORT, SUSPEND, SWAP, SYMMETRIC, SYNC,
	SYNONYM, SYSTEM, SYSTEM_TIME, SYSTEM_USER, TABLE, TABLES, TABLESAMPLE, TABLESPACE, TAG, TARGET,
	TARGET_LAG, TASK, TBLPROPERTIES, TEMP, TEMPORARY, TEMPTABLE, TERMINATED, TERSE, TEXT,
	TEXTFILE, THEN, THROW, TIES, TIME, TIMEFORMAT, TIMESTAMP, TIMESTAMPTZ, TIMESTAMP_NTZ, TIMETZ,
	TIMEZONE, TIMEZONE_ABBR, TIMEZONE_HOUR, TIMEZONE_MINUTE, TIMEZONE_REGION, TINYBLOB, TINYINT,
	TINYTEXT, TO, TOP, TOTALS, TOTP, TRACE, TRAILING, TRAN, TRANSACTION, TRANSIENT, TRANSLATE,
	TRANSLATE_REGEX, TRANSLATION, TREAT, TREE, TRIGGER, TRIM, TRIM_ARRAY, TRUE, TRUNCATE,
	TRUNCATECOLUMNS, TRY, TRY_CAST, TRY_CONVERT, TSQUERY, TSVECTOR, TUPLE, TYPE, TYPMOD_IN,
	TYPMOD_OUT, UBIGINT, UESCAPE, UHUGEINT, UINT128, UINT16, UINT256, UINT32, UINT64, UINT8,
	UNBOUNDED, UNCACHE, UNCOMMITTED, UNDEFINED, UNFREEZE, UNION, UNIQUE, UNKNOWN, UNLISTEN,
	UNLOAD, UNLOCK, UNLOGGED, UNMATCHED, UNNEST, UNPIVOT, UNSAFE, UNSET, UNSIGNED, UNTIL, UPDATE,
	UPPER, URL, USAGE, USE, USER, USER_RESOURCES, USING, USMALLINT, UTINYINT, UUID, VACUUM,
	VALID, VALIDATE, VALIDATION_MODE, VALUE, VALUES, VALUE_OF, VARBINARY, VARBIT, VARCHAR,
	VARCHAR2, VARIABLE, VARIABLES, VARYING, VAR_POP, VAR_SAMP, VERBOSE, VERSION, VERSIONING,
	VERSIONS, VIEW, VIEWS, VIRTUAL, VOLATILE, VOLUME, WAITFOR, WAREHOUSE, WAREHOUSES, WEEK,
	WEEKS, WHEN, WHENEVER, WHERE, WHILE, WIDTH_BUCKET, WINDOW, WITH, WITHIN, WITHOUT,
	WITHOUT_ARRAY_WRAPPER, WORK, WORKLOAD_IDENTITY, WRAPPER, WRITE, XML, XMLNAMESPACES, XMLTABLE,
	XOR, YEAR, YEARS, YES, ZONE, ZORDER, ZSTD,
}

// KeywordFromString converts a string to a Keyword constant.
// Returns NoKeyword if the string doesn't match any keyword.
func KeywordFromString(s string) (Keyword, bool) {
	// Binary search through sorted keywords
	low, high := 0, len(AllKeywords)
	for low < high {
		mid := (low + high) / 2
		if AllKeywords[mid] < Keyword(s) {
			low = mid + 1
		} else {
			high = mid
		}
	}
	if low < len(AllKeywords) && AllKeywords[low] == Keyword(s) {
		return AllKeywords[low], true
	}
	return NoKeyword, false
}

// Reserved keyword sets for parsing

// RESERVED_FOR_TABLE_ALIAS contains keywords that can't be used as a table alias,
// so that `FROM table_name alias` can be parsed unambiguously without looking ahead.
var RESERVED_FOR_TABLE_ALIAS = []Keyword{
	// Reserved as both a table and a column alias:
	WITH, EXPLAIN, ANALYZE, SELECT, WHERE, GROUP, SORT, HAVING, ORDER, PIVOT, UNPIVOT, TOP,
	LATERAL, VIEW, LIMIT, OFFSET, FETCH, UNION, EXCEPT, INTERSECT, MINUS,
	// Reserved only as a table alias in the FROM/JOIN clauses:
	ON, JOIN, INNER, CROSS, FULL, LEFT, RIGHT, NATURAL, USING, CLUSTER, DISTRIBUTE, GLOBAL,
	ANTI, SEMI, RETURNING, OUTPUT, ASOF, MATCH_CONDITION,
	// For MSSQL-specific OUTER APPLY (seems reserved in most dialects)
	OUTER, SET, QUALIFY, WINDOW, END, FOR,
	// For MYSQL PARTITION SELECTION
	PARTITION,
	// For Clickhouse PREWHERE
	PREWHERE, SETTINGS, FORMAT,
	// For Snowflake START WITH .. CONNECT BY
	START, CONNECT,
	// Reserved for Snowflake MATCH_RECOGNIZE
	MATCH_RECOGNIZE,
	// Reserved for Snowflake table sample
	SAMPLE, TABLESAMPLE, FROM, OPEN,
}

// RESERVED_FOR_COLUMN_ALIAS contains keywords that can't be used as a column alias,
// so that `SELECT <expr> alias` can be parsed unambiguously without looking ahead.
var RESERVED_FOR_COLUMN_ALIAS = []Keyword{
	// Reserved as both a table and a column alias:
	WITH, EXPLAIN, ANALYZE, SELECT, WHERE, GROUP, SORT, HAVING, ORDER, TOP, LATERAL, VIEW,
	LIMIT, OFFSET, FETCH, UNION, EXCEPT, EXCLUDE, INTERSECT, MINUS, CLUSTER, DISTRIBUTE, RETURNING,
	VALUES,
	// Reserved only as a column alias in the SELECT clause:
	FROM, INTO, END,
}

// RESERVED_FOR_TABLE_FACTOR contains the global list of reserved keywords allowed after FROM.
// Parser should call Dialect.GetReservedKeywordAfterFrom to allow for each dialect to customize the list.
var RESERVED_FOR_TABLE_FACTOR = []Keyword{
	INTO, LIMIT, HAVING, WHERE,
}

// RESERVED_FOR_IDENTIFIER contains the global list of reserved keywords that cannot be parsed
// as identifiers without special handling like quoting. Parser should call Dialect.IsReservedForIdentifier
// to allow for each dialect to customize the list.
var RESERVED_FOR_IDENTIFIER = []Keyword{
	EXISTS, INTERVAL, STRUCT, TRIM,
	// File format keywords - these appear in Snowflake stage paths
	// and should not be treated as identifiers/aliases
	PARQUET, CSV, JSON, ORC, AVRO, XML,
}

// Helper functions for checking reserved keywords

// IsReservedForIdentifier returns true if the keyword is reserved for identifiers
func IsReservedForIdentifier(kw Keyword) bool {
	for _, reserved := range RESERVED_FOR_IDENTIFIER {
		if reserved == kw {
			return true
		}
	}
	return false
}

// IsReservedForColumnAlias returns true if the keyword is reserved for column aliases
func IsReservedForColumnAlias(kw Keyword) bool {
	for _, reserved := range RESERVED_FOR_COLUMN_ALIAS {
		if reserved == kw {
			return true
		}
	}
	return false
}

// IsReservedForTableAlias returns true if the keyword is reserved for table aliases
func IsReservedForTableAlias(kw Keyword) bool {
	for _, reserved := range RESERVED_FOR_TABLE_ALIAS {
		if reserved == kw {
			return true
		}
	}
	return false
}

// IsReservedForTableFactor returns true if the keyword is reserved for table factors
func IsReservedForTableFactor(kw Keyword) bool {
	for _, reserved := range RESERVED_FOR_TABLE_FACTOR {
		if reserved == kw {
			return true
		}
	}
	return false
}
