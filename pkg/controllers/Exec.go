package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/paradoxe35/sqlite-rest/pkg/db"
)

type ExecBody struct {
	Query string `json:"query"`
}

// List of dangerous SQL operations that should be blocked
var dangerousOperations = []string{
	"DROP TABLE",
	"DROP DATABASE",
	"DELETE FROM",
	"TRUNCATE TABLE",
	"ALTER TABLE",
	"PRAGMA",
	"ATTACH DATABASE",
	"DETACH DATABASE",
}

// isQuerySafe checks if the query contains any dangerous operations
func isQuerySafe(query string) bool {
	upperQuery := strings.ToUpper(query)

	// Check for dangerous operations
	for _, op := range dangerousOperations {
		if strings.Contains(upperQuery, op) {
			return false
		}
	}

	return true
}

// determineQueryType identifies if the query is a SELECT or non-SELECT query
func determineQueryType(query string) string {
	trimmedQuery := strings.TrimSpace(query)
	upperQuery := strings.ToUpper(trimmedQuery)

	if strings.HasPrefix(upperQuery, "SELECT") {
		return "SELECT"
	} else if strings.HasPrefix(upperQuery, "INSERT") {
		return "INSERT"
	} else if strings.HasPrefix(upperQuery, "UPDATE") {
		return "UPDATE"
	} else if strings.HasPrefix(upperQuery, "CREATE") {
		return "CREATE"
	}

	return "OTHER"
}

// executeSelect handles SELECT queries and returns the results
func executeSelect(db *sql.DB, query string) ([]map[string]interface{}, error) {
	// Execute query
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get column names
	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Get column types
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	// Prepare result
	var result []map[string]interface{}

	// Process rows
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
			return nil, err
		}

		// Compose data map
		rowData := make(map[string]interface{})
		for i, columnKey := range columnNames {
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
				rowData[columnKey] = string(*columnPtrs[i].(*[]byte))
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

		result = append(result, rowData)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// executeNonSelect handles non-SELECT queries and returns affected rows
func executeNonSelect(db *sql.DB, query string) (int64, error) {
	// Execute query
	result, err := db.Exec(query)
	if err != nil {
		return 0, err
	}

	// Get affected rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func Exec(dbPath string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Create sql.DB instance
		db, err := db.Open(dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Parse body data
		data := ExecBody{}
		err = json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if data.Query == "" {
			http.Error(w, "Missing query", http.StatusBadRequest)
			return
		}

		// Check if query is safe
		if !isQuerySafe(data.Query) {
			http.Error(w, "Query contains dangerous operations", http.StatusForbidden)
			return
		}

		// Determine query type
		queryType := determineQueryType(data.Query)

		// Execute query based on type
		var result interface{}

		if queryType == "SELECT" {
			// Handle SELECT query
			rows, err := executeSelect(db, data.Query)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			result = map[string]interface{}{
				"status": "success",
				"type":   "select",
				"rows":   rows,
				"count":  len(rows),
			}
		} else {
			// Handle non-SELECT query
			rowsAffected, err := executeNonSelect(db, data.Query)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			result = map[string]interface{}{
				"status":        "success",
				"type":          strings.ToLower(queryType),
				"rows_affected": rowsAffected,
			}
		}

		// Return result
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
