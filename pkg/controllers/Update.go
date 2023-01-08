package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jonamat/sqlite-rest/pkg/db"
	"github.com/julienschmidt/httprouter"
)

func Update(dbPath string) httprouter.Handle {
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
		var columnValuesString string
		for k, v := range data {

			if v == nil {
				columnValuesString += fmt.Sprintf("%s=NULL,", k)
				continue
			}

			switch v.(type) {
			case string:
				columnValuesString += fmt.Sprintf("%s=\"%s\",", k, v)
			case int:
				columnValuesString += fmt.Sprintf("%s=%d,", k, v)
			case float64:
				columnValuesString += fmt.Sprintf("%s=%f,", k, v)
			case bool:
				columnValuesString += fmt.Sprintf("%s=%t,", k, v)
			default:
				columnValuesString += fmt.Sprintf("%s=%v,", k, v)
			}
		}
		// Remove last comma
		columnValuesString = columnValuesString[:len(columnValuesString)-1]

		// Execute query
		_, err = db.Exec(fmt.Sprintf("UPDATE %s SET %s WHERE id = %d", tableSelect, columnValuesString, id))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int64{"id": id})
	}
}
