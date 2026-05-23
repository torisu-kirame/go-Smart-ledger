package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
)

// OpenAndMigrate 解析 DSN、创建库（若不存在）、检测并补齐表/字段/索引/外键后返回连接。
func OpenAndMigrate(dsn string) (*sql.DB, error) {
	dbName, adminDSN, err := splitDSN(dsn)
	if err != nil {
		return nil, err
	}
	admin, err := sql.Open("mysql", adminDSN)
	if err != nil {
		return nil, fmt.Errorf("mysql admin open: %w", err)
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		return nil, fmt.Errorf("mysql admin ping: %w", err)
	}
	if err := ensureDatabase(admin, dbName); err != nil {
		return nil, err
	}

	db, err := Open(dsn)
	if err != nil {
		return nil, err
	}
	if err := Migrate(db, dbName); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// Migrate 在已选定的库上检测并创建表、字段、索引与外键。
func Migrate(db *sql.DB, dbName string) error {
	if dbName == "" {
		var current string
		if err := db.QueryRow("SELECT DATABASE()").Scan(&current); err != nil {
			return fmt.Errorf("resolve database name: %w", err)
		}
		dbName = current
	}
	if dbName == "" {
		return fmt.Errorf("no database selected")
	}

	for _, tbl := range schemaTables {
		if err := ensureTable(db, dbName, tbl); err != nil {
			return fmt.Errorf("table %s: %w", tbl.name, err)
		}
	}
	for _, tbl := range schemaTables {
		for _, fk := range tbl.foreignKeys {
			if err := ensureForeignKey(db, dbName, tbl.name, fk); err != nil {
				return fmt.Errorf("fk %s on %s: %w", fk.name, tbl.name, err)
			}
		}
	}
	return nil
}

func splitDSN(dsn string) (dbName, adminDSN string, err error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "", "", fmt.Errorf("parse dsn: %w", err)
	}
	dbName = cfg.DBName
	if dbName == "" {
		return "", "", fmt.Errorf("dsn missing database name (e.g. /smart_ledger)")
	}
	cfg.DBName = ""
	adminDSN = cfg.FormatDSN()
	return dbName, adminDSN, nil
}

func ensureDatabase(admin *sql.DB, name string) error {
	exists, err := databaseExists(admin, name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	q := fmt.Sprintf(
		"CREATE DATABASE `%s` DEFAULT CHARACTER SET %s COLLATE %s",
		escapeIdent(name), defaultCharset, defaultCollation,
	)
	if _, err := admin.Exec(q); err != nil {
		return fmt.Errorf("create database %s: %w", name, err)
	}
	return nil
}

func databaseExists(admin *sql.DB, name string) (bool, error) {
	var n int
	err := admin.QueryRow(
		`SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?`,
		name,
	).Scan(&n)
	return n > 0, err
}

func ensureTable(db *sql.DB, dbName string, tbl tableSchema) error {
	exists, err := tableExists(db, dbName, tbl.name)
	if err != nil {
		return err
	}
	if !exists {
		if _, err := db.Exec(tbl.createSQL); err != nil {
			return fmt.Errorf("create: %w", err)
		}
	}
	for _, col := range tbl.columns {
		if err := ensureColumn(db, dbName, tbl.name, col); err != nil {
			return err
		}
	}
	for _, idx := range tbl.indexes {
		if err := ensureIndex(db, dbName, tbl.name, idx); err != nil {
			return err
		}
	}
	return nil
}

func tableExists(db *sql.DB, dbName, table string) (bool, error) {
	var n int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM information_schema.TABLES
		 WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`,
		dbName, table,
	).Scan(&n)
	return n > 0, err
}

func ensureColumn(db *sql.DB, dbName, table string, col columnSchema) error {
	exists, err := columnExists(db, dbName, table, col.name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	q := fmt.Sprintf("ALTER TABLE `%s` %s", escapeIdent(table), col.addSQL)
	if _, err := db.Exec(q); err != nil {
		return fmt.Errorf("add column %s: %w", col.name, err)
	}
	return nil
}

func columnExists(db *sql.DB, dbName, table, column string) (bool, error) {
	var n int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM information_schema.COLUMNS
		 WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND COLUMN_NAME = ?`,
		dbName, table, column,
	).Scan(&n)
	return n > 0, err
}

func ensureIndex(db *sql.DB, dbName, table string, idx indexSchema) error {
	exists, err := indexExists(db, dbName, table, idx.name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if _, err := db.Exec(idx.createSQL); err != nil {
		if isDuplicate(err) {
			return nil
		}
		return fmt.Errorf("create index %s: %w", idx.name, err)
	}
	return nil
}

func indexExists(db *sql.DB, dbName, table, indexName string) (bool, error) {
	var n int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM information_schema.STATISTICS
		 WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND INDEX_NAME = ?`,
		dbName, table, indexName,
	).Scan(&n)
	return n > 0, err
}

func ensureForeignKey(db *sql.DB, dbName, table string, fk fkSchema) error {
	exists, err := foreignKeyExists(db, dbName, table, fk.name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	refExists, err := tableExists(db, dbName, fk.referenced)
	if err != nil {
		return err
	}
	if !refExists {
		return fmt.Errorf("referenced table %s not found", fk.referenced)
	}
	if _, err := db.Exec(fk.createSQL); err != nil {
		if isDuplicate(err) {
			return nil
		}
		return fmt.Errorf("create fk %s: %w", fk.name, err)
	}
	return nil
}

func foreignKeyExists(db *sql.DB, dbName, table, fkName string) (bool, error) {
	var n int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS
		 WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND CONSTRAINT_NAME = ? AND CONSTRAINT_TYPE = 'FOREIGN KEY'`,
		dbName, table, fkName,
	).Scan(&n)
	return n > 0, err
}

func escapeIdent(name string) string {
	return strings.ReplaceAll(name, "`", "``")
}

func isDuplicate(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Duplicate") || strings.Contains(msg, "1061") || strings.Contains(msg, "1826")
}
