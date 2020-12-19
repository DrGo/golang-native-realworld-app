package main

import (
	"log"

	sql "crawshaw.io/sqlite"
	sqlx "crawshaw.io/sqlite/sqlitex"
	"github.com/drgo/realworld/errors"
)

const poolSize = 10

// A flags value of 0 defaults to:
//	SQLITE_OPEN_READWRITE
//	SQLITE_OPEN_CREATE
//	SQLITE_OPEN_WAL
//	SQLITE_OPEN_URI
//	SQLITE_OPEN_NOMUTEX
const poolFlags = 0

type sqlite struct {
	pool *sqlx.Pool
}

// Guarantee that sqlite implements the DB interface
var _ DB = (*sqlite)(nil)

// NewDB opens a fixed-size pool of SQLite connections
func NewDB(dsn string) (DB, error) {
	db := sqlite{}
	var err error
	db.pool, err = sqlx.Open(dsn, poolFlags, poolSize)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func (db *sqlite) Close() error {
	return db.pool.Close()
}

// bindQuery bind stmt to args based on type
func bindQuery(stmt *sql.Stmt, args Args) error {
	for k, v := range args {
		switch tv := v.(type) {
		case string:
			stmt.SetText(k, tv)
		case int:
			stmt.SetInt64(k, int64(tv))
		case int64:
			stmt.SetInt64(k, tv)
		case float64:
			stmt.SetFloat(k, tv)
		case bool:
			stmt.SetBool(k, tv)
		case []byte:
			stmt.SetBytes(k, tv)
		default:
			return errors.Errorf("%s has unsupported type", k)
		}
		//TODO: SetNull needed?
	}
	return nil
}

func (db *sqlite) Query(query string, args Args) (rows []Row, rowCount int, err error) {
	conn := db.pool.Get(nil)
	defer db.pool.Put(conn)
	log.Println("Query:", query)
	//compile (and cache) query; no need to finalize it
	stmt := conn.Prep(query)
	// bind args to compiled statement
	if err := bindQuery(stmt, args); err != nil {
		return nil, 0, err
	}
	defer func() {
		if err := stmt.Reset(); err != nil {
			panic(err)
		}
	}()
	// rows := []interface{}{}
	// TODO: not clear if needed https://stackoverflow.com/questions/35741175/sqlite3-reset-when-is-it-needed
	// step through returned records and retrieve info
	for {
		if hasRow, err := stmt.Step(); err != nil {
			return nil, 0, err
		} else if !hasRow {
			break
		}
		// loop over returned columns and determine type (starting at 0)
		colCount := stmt.ColumnCount()
		// construct a map with field names and values
		row := Row{}
		for i := 0; i < colCount; i++ {
			name := stmt.ColumnName(i)
			// ColumnType returns the datatype code for the initial data type of the result column.
			switch stmt.ColumnType(i) {
			case sql.SQLITE_INTEGER:
				row[name] = stmt.ColumnInt64(i)
			case sql.SQLITE_FLOAT:
				row[name] = stmt.ColumnFloat(i)
			case sql.SQLITE_TEXT:
				row[name] = stmt.ColumnText(i)
			case sql.SQLITE_BLOB:
				//TODO:
			case sql.SQLITE_NULL:
				row[name] = ""
			}
		}
		rows = append(rows, row)
	}
	return rows, len(rows), nil
}

// Exec
// returns the number of rows modified, inserted or deleted by the most recently completed INSERT, UPDATE or DELETE; usually returns the rowid of the most recent successful INSERT or error
func (db *sqlite) Exec(query string, args Args) (rowsAffected int, lastRowID int64, err error) {
	conn := db.pool.Get(nil)
	defer db.pool.Put(conn)
	log.Println("Exec:", query)
	//compile (and cache) query; no need to finalize it
	stmt := conn.Prep(query)
	// bind args to compiled statement
	if err = bindQuery(stmt, args); err != nil {
		return 0, 0, err
	}
	defer func() {
		if err = stmt.Reset(); err != nil {
			panic(err)
		}
	}()
	if _, err = stmt.Step(); err != nil {
		return 0, 0, err
	}
	return conn.Changes(), conn.LastInsertRowID(), nil
}

func (db *sqlite) JSONQuery(query string, args Args) (result string, rowCount int, err error) {
	conn := db.pool.Get(nil)
	defer db.pool.Put(conn)
	log.Println("JSONQuery:", query)
	//compile (and cache) query; no need to finalize it
	stmt := conn.Prep(query)
	// bind args to compiled statement
	if err := bindQuery(stmt, args); err != nil {
		return "", 0, err
	}
	defer func() {
		if err := stmt.Reset(); err != nil {
			panic(err)
		}
	}()
	if hasRow, err := stmt.Step(); err != nil {
		return "", 0, err
	} else if !hasRow {
		return "", 0, nil
	}
	// there is only column (index=0) containing the json payload
	result = stmt.ColumnText(0)
	return result, 1, nil
}

// TODO: cleanup
// func (db *sqliteDb) cleanup() {
// 	conn := db.pool.Get(context.TODO())
// 	defer db.pool.Put(conn)
// 	sqlite_exec(conn, "DELETE FROM sessions WHERE unlisted!=0 OR last_active < DATETIME('now', '-1 day')")
// }

// func sqliteCleanupTask(db *sqliteDb) {
// 	for {
// 		time.Sleep(24 * time.Hour)
// 		db.cleanup()
// 	}
// }
