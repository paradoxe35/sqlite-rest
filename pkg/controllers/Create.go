package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/paradoxe35/sqlite-rest/pkg/db"
)

func Create(dbPath string) httprouter.Handle {
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
			// Check if this is a syntax error (client error) or a server error
			errMsg := err.Error()
			if strings.Contains(errMsg, "no such table") {
				sendJSONError(w, fmt.Sprintf("Table not found: %s", tableSelect), http.StatusBadRequest)
			} else if strings.Contains(errMsg, "constraint failed") || strings.Contains(errMsg, "UNIQUE constraint") {
				sendJSONError(w, fmt.Sprintf("Constraint violation: %s", errMsg), http.StatusBadRequest)
			} else {
				sendJSONError(w, fmt.Sprintf("Error creating record: %s", errMsg), http.StatusInternalServerError)
			}
			return
		}

		// Get ID
		id, err := res.LastInsertId()
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Error retrieving inserted ID: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"id":     id,
		})
	}
}
