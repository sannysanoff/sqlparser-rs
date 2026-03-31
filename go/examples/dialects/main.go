package main

import (
	"fmt"
	"log"

	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/snowflake"
	"github.com/user/sqlparser/parser"
)

func main() {
	fmt.Println("Dialect-Specific Examples")
	fmt.Println("=========================\n")

	// PostgreSQL Example
	fmt.Println("PostgreSQL - Arrays and JSON:")
	pg := postgresql.NewPostgreSqlDialect()
	pgSQL := "SELECT ARRAY[1, 2, 3], '{\"key\": \"value\"}'::JSONB FROM users"

	statements, err := parser.ParseSQL(pg, pgSQL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Input:  %s\n", pgSQL)
	fmt.Printf("  Output: %s\n\n", statements[0].String())

	// MySQL Example
	fmt.Println("MySQL - Backtick identifiers and ON DUPLICATE KEY:")
	mysql := mysql.NewMySqlDialect()
	mysqlSQL := "INSERT INTO `my-table` (`col1`, `col2`) VALUES (1, 2) ON DUPLICATE KEY UPDATE `col2` = VALUES(`col2`)"

	statements, err = parser.ParseSQL(mysql, mysqlSQL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Input:  %s\n", mysqlSQL)
	fmt.Printf("  Output: %s\n\n", statements[0].String())

	// BigQuery Example
	fmt.Println("BigQuery - STRUCT and ARRAY:")
	bq := bigquery.NewBigQueryDialect()
	bqSQL := "SELECT STRUCT(1 AS x, 'a' AS y), [1, 2, 3] AS arr FROM `project.dataset.table`"

	statements, err = parser.ParseSQL(bq, bqSQL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Input:  %s\n", bqSQL)
	fmt.Printf("  Output: %s\n\n", statements[0].String())

	// Snowflake Example
	fmt.Println("Snowflake - Semi-structured data:")
	sf := snowflake.NewSnowflakeDialect()
	sfSQL := "SELECT data:key::STRING FROM table1"

	statements, err = parser.ParseSQL(sf, sfSQL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Input:  %s\n", sfSQL)
	fmt.Printf("  Output: %s\n\n", statements[0].String())

	// More PostgreSQL - Dollar-quoted strings
	fmt.Println("PostgreSQL - Dollar-quoted strings:")
	pgSQL2 := "SELECT $$This is a 'quoted' string$$"

	statements, err = parser.ParseSQL(pg, pgSQL2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Input:  %s\n", pgSQL2)
	fmt.Printf("  Output: %s\n\n", statements[0].String())

	// MySQL - LIMIT with comma
	fmt.Println("MySQL - LIMIT with comma:")
	mysqlSQL2 := "SELECT * FROM users LIMIT 10, 20"

	statements, err = parser.ParseSQL(mysql, mysqlSQL2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Input:  %s\n", mysqlSQL2)
	fmt.Printf("  Output: %s\n\n", statements[0].String())

	fmt.Println("All dialect-specific examples completed successfully!")
}
