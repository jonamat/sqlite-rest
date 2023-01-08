package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jonamat/sqlite-rest/pkg/db"
	"github.com/julienschmidt/httprouter"
)

func Delete(dbPath string) httprouter.Handle {
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

		// Execute query
		_, err = db.Exec("DELETE FROM " + tableSelect + " WHERE id = " + idParam)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int64{"id": id})
	}
}
