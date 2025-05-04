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

func Update(dbPath string) httprouter.Handle {
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

		// Parse body data
		data := make(map[string]interface{})
		err = json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Invalid request body: %s", err.Error()), http.StatusBadRequest)
			return
		}

		if len(data) == 0 {
			sendJSONError(w, "Missing data in request body", http.StatusBadRequest)
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
		result, err := db.Exec(fmt.Sprintf("UPDATE %s SET %s WHERE id = %d", tableSelect, columnValuesString, id))
		if err != nil {
			// Check if this is a syntax error (client error) or a server error
			errMsg := err.Error()
			if strings.Contains(errMsg, "no such table") {
				sendJSONError(w, fmt.Sprintf("Table not found: %s", tableSelect), http.StatusBadRequest)
			} else if strings.Contains(errMsg, "no such column") {
				sendJSONError(w, fmt.Sprintf("Invalid column in update: %s", errMsg), http.StatusBadRequest)
			} else if strings.Contains(errMsg, "constraint failed") || strings.Contains(errMsg, "UNIQUE constraint") {
				sendJSONError(w, fmt.Sprintf("Constraint violation: %s", errMsg), http.StatusBadRequest)
			} else {
				sendJSONError(w, fmt.Sprintf("Error updating record: %s", errMsg), http.StatusInternalServerError)
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
