package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/paradoxe35/sqlite-rest/pkg/db"
)

func Delete(dbPath string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Create sql.DB instance
		db, err := db.Open(dbPath)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Database error: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Parse table name from params
		tableSelect := params.ByName("table")
		if tableSelect == "" {
			sendJSONError(w, "Missing table parameter", http.StatusBadRequest)
			return
		}

		// Parse id from params
		idParam := params.ByName("id")
		if idParam == "" {
			sendJSONError(w, "Missing ID parameter", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Invalid ID format: %s", err.Error()), http.StatusBadRequest)
			return
		}

		// Execute query
		result, err := db.Exec("DELETE FROM " + tableSelect + " WHERE id = " + idParam)
		if err != nil {
			// Check if this is a syntax error (client error) or a server error
			errMsg := err.Error()
			if strings.Contains(errMsg, "no such table") {
				sendJSONError(w, fmt.Sprintf("Table not found: %s", tableSelect), http.StatusBadRequest)
			} else {
				sendJSONError(w, fmt.Sprintf("Error deleting record: %s", errMsg), http.StatusInternalServerError)
			}
			return
		}

		// Check if any rows were affected
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			sendJSONError(w, fmt.Sprintf("Record with ID %d not found", id), http.StatusNotFound)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"id":     id,
		})
	}
}
