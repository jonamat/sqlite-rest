package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jonamat/sqlite-rest/pkg/db"
	"github.com/julienschmidt/httprouter"
)

type Filter struct {
	Column   string `json:"column"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

func GetAll(dbPath string) httprouter.Handle {
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

		// Parse columns from params or use all
		var columnsSelect string
		columnsParam := r.URL.Query().Get("cols")
		if columnsParam == "" {
			columnsSelect = "*"
		}

		// Parse filters_raw from query string and build WHERE clause
		var whereClause string
		filtersParam := r.URL.Query().Get("filters_raw")

		if filtersParam != "" {
			unescapedFilters, err := url.QueryUnescape(filtersParam)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			whereClause = "WHERE " + unescapedFilters
		}

		filtersStruct := r.URL.Query().Get("filters")
		if whereClause != "" {
			if filtersParam != "" {
				http.Error(w, "Cannot use both filters and filters_raw", http.StatusBadRequest)
				return
			}

			filterArr := []Filter{}

			unescapedFilters, err := url.QueryUnescape(filtersStruct)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = json.Unmarshal([]byte(unescapedFilters), &filterArr)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var filterArrStr []string
			for _, filter := range filterArr {
				filterArrStr = append(filterArrStr, fmt.Sprintf("%s %s '%s'", filter.Column, filter.Operator, filter.Value))
			}

			whereClause = "WHERE " + strings.Join(filterArrStr, " AND ")
		}

		// Parse limitClause from query string
		var limitClause string
		limitParam := r.URL.Query().Get("limit")
		if limitParam != "" {
			limitClause = "LIMIT " + limitParam
		}

		// Parse offsetClause from query string
		var offsetClause string
		offsetParam := r.URL.Query().Get("offset")
		if offsetParam != "" && limitParam == "" {
			http.Error(w, "Cannot use offset without limit", http.StatusBadRequest)
			return
		}
		if offsetParam != "" {
			offsetClause = "OFFSET " + offsetParam
		}

		// Parse order by from query string
		var orderByClause string
		orderByParam := r.URL.Query().Get("order_by")
		if orderByParam != "" {
			orderByClause = "ORDER BY " + orderByParam
		}

		// Parse order direction from query string
		orderDir := r.URL.Query().Get("order_dir")
		if orderDir != "" && orderByParam == "" {
			http.Error(w, "Cannot use order_dir without order_by", http.StatusBadRequest)
			return
		}

		// Execute query
		rows, err := db.Query(fmt.Sprintf("SELECT %s FROM %s %s %s %s %s %s", columnsSelect, tableSelect, whereClause, orderByClause, orderDir, limitClause, offsetClause))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Get column names
		var columnNames []string
		columnNames, err = rows.Columns()
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

		// Scan rows
		var data []map[string]interface{}
		for rows.Next() {
			// Create slice of pointers to scan into
			columnPtrs := make([]interface{}, len(columnNames))

			// Infer type from column type
			for i := range columnNames {
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

			// Scan row into column pointers
			err = rows.Scan(columnPtrs...)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Compose row data map
			rowData := make(map[string]interface{})
			for i, columnKey := range columnNames {

				// Preserve null values from db
				switch strings.ToUpper(columnTypes[i].DatabaseTypeName()) {
				case "PRIMARY_KEY", "INTEGER", "INT", "TINYINT", "SMALLINT", "MEDIUMINT", "BIGINT", "UNSIGNED BIG INT", "INT2", "INT8", "DECIMAL":
					if columnPtrs[i].(*sql.NullInt64).Valid {
						rowData[columnKey] = columnPtrs[i].(*sql.NullInt64).Int64
					} else {
						rowData[columnKey] = nil
					}
				case "REAL", "DOUBLE", "DOUBLE PRECISION", "FLOAT", "NUMERIC":
					if columnPtrs[i].(*sql.NullFloat64).Valid {
						rowData[columnKey] = columnPtrs[i].(*sql.NullFloat64).Float64
					} else {
						rowData[columnKey] = nil
					}
				case "BLOB":
					if columnPtrs[i].(*[]byte) != nil {
						rowData[columnKey] = columnPtrs[i].(*[]byte)
					} else {
						rowData[columnKey] = nil
					}
				case "TEXT", "CHARACTER", "VARCHAR", "VARYING CHARACTER", "NCHAR", "NATIVE CHARACTER", "NVARCHAR", "CLOB", "DATE", "DATETIME":
					if columnPtrs[i].(*sql.NullString).Valid {
						rowData[columnKey] = columnPtrs[i].(*sql.NullString).String
					} else {
						rowData[columnKey] = nil
					}
				case "BOOLEAN", "BOOL":
					if columnPtrs[i].(*sql.NullBool).Valid {
						rowData[columnKey] = columnPtrs[i].(*sql.NullBool).Bool
					} else {
						rowData[columnKey] = nil
					}
				default:
					if columnPtrs[i].(*sql.NullString).Valid {
						rowData[columnKey] = columnPtrs[i].(*sql.NullString).String
					} else {
						rowData[columnKey] = nil
					}
				}
			}
			data = append(data, rowData)
		}

		// Compose response and return data
		response := map[string]interface{}{
			"total_rows": len(data),
			"data":       data,
		}

		if offsetParam != "" {
			offset, _ := strconv.ParseInt(offsetParam, 10, 64)
			response["offset"] = offset
		} else {
			response["offset"] = nil
		}

		if limitParam != "" {
			limit, _ := strconv.ParseInt(limitParam, 10, 64)
			response["limit"] = limit
		} else {
			response["limit"] = nil
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
