package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jonamat/sqlite-rest/pkg/db"
	"github.com/julienschmidt/httprouter"
)

func Create(dbPath string) httprouter.Handle {
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

		// Parse body data
		data := make(map[string]interface{})
		err = json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(data) == 0 {
			http.Error(w, "Missing data", http.StatusBadRequest)
			return
		}

		// Extract keys and values from data
		columnNames := make([]string, 0, len(data))
		var columnValuesString string
		for k, v := range data {
			columnNames = append(columnNames, k)

			if v == nil {
				columnValuesString += "NULL,"
				continue
			}

			switch v.(type) {
			case string:
				columnValuesString += fmt.Sprintf("\"%s\",", v)
			case int:
				columnValuesString += fmt.Sprintf("%d,", v)
			case float64:
				columnValuesString += fmt.Sprintf("%f,", v)
			case bool:
				columnValuesString += fmt.Sprintf("%t,", v)
			default:
				columnValuesString += fmt.Sprintf("%v,", v)
			}
		}
		// Remove last comma
		columnValuesString = columnValuesString[:len(columnValuesString)-1]

		// Execute query
		res, err := db.Exec(fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableSelect, strings.Join(columnNames, ", "), columnValuesString))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get ID
		id, err := res.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return ID
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int64{"id": id})
	}
}
