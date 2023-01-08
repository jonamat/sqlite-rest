package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jonamat/sqlite-rest/pkg/db"
	"github.com/julienschmidt/httprouter"
)

func Get(dbPath string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Create sql.DB instance
		db, err := db.Open(dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Parse table name from params
		tableSelect := params.ByName("table")
		if tableSelect == "" {
			http.Error(w, "Missing table", http.StatusBadRequest)
			return
		}

		// Parse id from params
		idParam := params.ByName("id")
		if idParam == "" {
			http.Error(w, "Missing ID", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Parse columns from params or use all
		columnsSelect := r.URL.Query().Get("columns")
		if columnsSelect == "" {
			columnsSelect = "*"
		}

		// Execute query
		rows, err := db.Query(fmt.Sprintf("SELECT %s FROM %s WHERE id = %d", columnsSelect, tableSelect, id))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Get column names
		columnNames, err := rows.Columns()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get column types
		columnTypes, err := rows.ColumnTypes()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create slice of pointers to scan into
		columnPtrs := make([]interface{}, len(columnNames))

		// Infer type from column type
		for i := range columnNames {
			// Refer to https://www.sqlite.org/datatype3.html index 3.1.1
			switch strings.ToUpper(columnTypes[i].DatabaseTypeName()) {
			case "PRIMARY_KEY", "INTEGER", "INT", "TINYINT", "SMALLINT", "MEDIUMINT", "BIGINT", "UNSIGNED BIG INT", "INT2", "INT8", "DECIMAL":
				columnPtrs[i] = new(sql.NullInt64)
			case "REAL", "DOUBLE", "DOUBLE PRECISION", "FLOAT", "NUMERIC":
				columnPtrs[i] = new(sql.NullFloat64)
			case "BLOB":
				columnPtrs[i] = new([]byte)
			case "TEXT", "CHARACTER", "VARCHAR", "VARYING CHARACTER", "NCHAR", "NATIVE CHARACTER", "NVARCHAR", "CLOB", "DATE", "DATETIME":
				columnPtrs[i] = new(sql.NullString)
			case "BOOLEAN", "BOOL":
				columnPtrs[i] = new(sql.NullBool)
			default:
				columnPtrs[i] = new(sql.NullString)
			}
		}

		// Check if row exists
		next := rows.Next()
		if !next {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		// Scan row into column pointers
		err = rows.Scan(columnPtrs...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Compose data map
		data := make(map[string]interface{})
		for i, columnKey := range columnNames {

			// Preserve null values from db
			switch strings.ToUpper(columnTypes[i].DatabaseTypeName()) {
			case "PRIMARY_KEY", "INTEGER", "INT", "TINYINT", "SMALLINT", "MEDIUMINT", "BIGINT", "UNSIGNED BIG INT", "INT2", "INT8", "DECIMAL":
				if columnPtrs[i].(*sql.NullInt64).Valid {
					data[columnKey] = columnPtrs[i].(*sql.NullInt64).Int64
				} else {
					data[columnKey] = nil
				}
			case "REAL", "DOUBLE", "DOUBLE PRECISION", "FLOAT", "NUMERIC":
				if columnPtrs[i].(*sql.NullFloat64).Valid {
					data[columnKey] = columnPtrs[i].(*sql.NullFloat64).Float64
				} else {
					data[columnKey] = nil
				}
			case "BLOB":
				if columnPtrs[i].(*[]byte) != nil {
					data[columnKey] = columnPtrs[i].(*[]byte)
				} else {
					data[columnKey] = nil
				}
			case "TEXT", "CHARACTER", "VARCHAR", "VARYING CHARACTER", "NCHAR", "NATIVE CHARACTER", "NVARCHAR", "CLOB", "DATE", "DATETIME":
				if columnPtrs[i].(*sql.NullString).Valid {
					data[columnKey] = columnPtrs[i].(*sql.NullString).String
				} else {
					data[columnKey] = nil
				}
			case "BOOLEAN", "BOOL":
				if columnPtrs[i].(*sql.NullBool).Valid {
					data[columnKey] = columnPtrs[i].(*sql.NullBool).Bool
				} else {
					data[columnKey] = nil
				}
			default:
				if columnPtrs[i].(*sql.NullString).Valid {
					data[columnKey] = columnPtrs[i].(*sql.NullString).String
				} else {
					data[columnKey] = nil
				}
			}

		}

		// Return data
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
