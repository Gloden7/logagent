package task

import (
	"database/sql"
	"fmt"
	"strings"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"

	// postgresql driver
	_ "github.com/blusewang/pg"
	// sqlite driver
	_ "github.com/mattn/go-sqlite3"
)

func initDatabase(driverName, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(10)
	return db, nil
}

func createTable(db *sql.DB, driverName string, table string, fields []string) error {
	sql := "CREATE TABLE IF NOT EXISTS " + table + " (\n"
	columnSQL := strings.Join(fields, ",\n")
	switch driverName {
	case "mysql":
		sql += columnSQL + "\n)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;"
	default:
		sql += columnSQL + "\n);"
	}
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func genInsertSQL(dn string, tableName string, columns []string) string {
	var sql string
	n := len(columns)
	var b strings.Builder
	switch dn {
	case "pg":
		for i := 1; i <= n; i++ {
			sql += fmt.Sprintf("$%d", i)
			if i == n {
				break
			}
			sql += ","
		}

		for i := 0; i < len(columns); i++ {
			n += len(columns[i])
		}

		b.Grow(n)
		b.WriteString(fmt.Sprintf("\"%s\"", strings.ToLower(columns[0])))
		for _, s := range columns[1:] {
			b.WriteByte(',')
			b.WriteString(fmt.Sprintf("\"%s\"", strings.ToLower(s)))
		}
	default:
		for i := 1; i <= n; i++ {
			sql += "?"
			if i == n {
				break
			}
			sql += ","
		}

		b.Grow(n)
		b.WriteString(fmt.Sprintf("`%s`", strings.ToLower(columns[0])))
		for _, s := range columns[1:] {
			b.WriteByte(',')
			b.WriteString(fmt.Sprintf("`%s`", strings.ToLower(s)))
		}
	}

	return fmt.Sprintf("INSERT INTO %s(%s)values(%s)", tableName,
		b.String(), sql)
}

func genSortFunc(columns []string) func(data message) []interface{} {
	return func(data message) []interface{} {
		var insertData []interface{}
		for _, k := range columns {
			if x, ok := data[k]; ok {
				insertData = append(insertData, x)
			} else {
				insertData = append(insertData, nil)
			}
		}
		return insertData
	}
}
