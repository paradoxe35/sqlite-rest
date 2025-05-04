package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/paradoxe35/sqlite-rest/pkg/db"
)

type ExecBody struct {
	Query string `json:"query"`
}

// Default list of dangerous SQL operations that should be blocked
var defaultDangerousOperations = []string{
	"DROP TABLE",
	"DROP DATABASE",
	"DELETE FROM",
	"TRUNCATE TABLE",
	"ALTER TABLE",
	// "PRAGMA" is removed to allow PRAGMA queries that return data
	"ATTACH DATABASE",
	"DETACH DATABASE",
}

// getDangerousOperations returns the list of dangerous operations
// It checks if the user has defined custom dangerous operations via environment variables
func getDangerousOperations() []string {
	// Check if the environment variable exists
	customOperations, exists := os.LookupEnv("SQLITE_REST_DANGEROUS_OPS")

	// If the environment variable exists (even if empty), use it
	if exists {
		// If it's empty, return an empty list (all operations allowed)
		if customOperations == "" {
			return []string{}
		}

		// Parse comma-separated list
		operations := strings.Split(customOperations, ",")
		// Trim whitespace
		for i, op := range operations {
			operations[i] = strings.ToUpper(strings.TrimSpace(op))
		}
		return operations
	}

	// Return default list if no custom operations defined
	return defaultDangerousOperations
}

// isQuerySafe checks if the query contains any dangerous operations
func isQuerySafe(query string) bool {
	upperQuery := strings.ToUpper(query)

	// Get the list of dangerous operations (either default or custom)
	dangerousOperations := getDangerousOperations()

	// If the list is empty, all operations are allowed
	if len(dangerousOperations) == 0 {
		return true
	}

	// Check for dangerous operations
	for _, op := range dangerousOperations {
		if op != "" && strings.Contains(upperQuery, op) {
			return false
		}
	}

	return true
}

// determineQueryType identifies the type of SQL query
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
	} else if strings.HasPrefix(upperQuery, "SHOW TABLES") ||
		strings.HasPrefix(upperQuery, "LIST TABLES") {
		return "SHOW_TABLES"
	} else if strings.HasPrefix(upperQuery, "PRAGMA") {
		return "PRAGMA"
	} else if strings.HasPrefix(upperQuery, "EXPLAIN") {
		return "EXPLAIN"
	} else if strings.HasPrefix(upperQuery, "ANALYZE") {
		return "ANALYZE"
	}

	return "OTHER"
}

// isDataReturningQuery checks if the query is expected to return data
func isDataReturningQuery(queryType string) bool {
	// These query types typically return data
	dataReturningTypes := map[string]bool{
		"SELECT":      true,
		"PRAGMA":      true,
		"EXPLAIN":     true,
		"ANALYZE":     true,
		"SHOW_TABLES": true,
	}

	return dataReturningTypes[queryType]
}

// listTables returns a list of all tables in the database
func listTables(db *sql.DB) ([]string, error) {
	// In SQLite, we can query the sqlite_master table to get a list of all tables
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		// Skip internal SQLite tables
		if !strings.HasPrefix(name, "sqlite_") {
			tables = append(tables, name)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
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
			sendJSONError(w, fmt.Sprintf("Database error: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Parse body data
		data := ExecBody{}
		err = json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Invalid request body: %s", err.Error()), http.StatusBadRequest)
			return
		}

		if data.Query == "" {
			sendJSONError(w, "Missing query parameter", http.StatusBadRequest)
			return
		}

		// Check if query is safe
		if !isQuerySafe(data.Query) {
			sendJSONError(w, "Query contains dangerous operations that are not allowed", http.StatusForbidden)
			return
		}

		// Determine query type
		queryType := determineQueryType(data.Query)

		// Execute query based on type
		var result interface{}

		if queryType == "SHOW_TABLES" {
			// Handle SHOW TABLES command
			tables, err := listTables(db)
			if err != nil {
				// For listing tables, most errors would be server-side issues
				// since this is a simple query on sqlite_master
				sendJSONError(w, fmt.Sprintf("Error listing tables: %s", err.Error()), http.StatusInternalServerError)
				return
			}

			// Convert to rows format for consistency
			var rows []map[string]interface{}
			for _, table := range tables {
				rows = append(rows, map[string]interface{}{
					"table_name": table,
				})
			}

			result = map[string]interface{}{
				"status": "success",
				"type":   "show_tables",
				"tables": tables,
				"rows":   rows,
				"count":  len(tables),
			}
		} else if isDataReturningQuery(queryType) {
			// Handle queries that return data (SELECT, PRAGMA, EXPLAIN, etc.)
			rows, err := executeSelect(db, data.Query)
			if err != nil {
				// Check if this is a syntax error (client error) or a server error
				errMsg := err.Error()
				if strings.Contains(errMsg, "syntax error") ||
					strings.Contains(errMsg, "no such table") ||
					strings.Contains(errMsg, "no such column") {
					// This is likely a client error - bad SQL syntax or referencing non-existent tables/columns
					sendJSONError(w, fmt.Sprintf("Invalid SQL query: %s", errMsg), http.StatusBadRequest)
				} else {
					// This is likely a server error - database issues, etc.
					sendJSONError(w, fmt.Sprintf("Error executing %s query: %s", strings.ToLower(queryType), errMsg), http.StatusInternalServerError)
				}
				return
			}
			result = map[string]interface{}{
				"status": "success",
				"type":   strings.ToLower(queryType),
				"rows":   rows,
				"count":  len(rows),
			}
		} else {
			// Handle non-data-returning queries (INSERT, UPDATE, DELETE, etc.)
			rowsAffected, err := executeNonSelect(db, data.Query)
			if err != nil {
				// Check if this is a syntax error (client error) or a server error
				errMsg := err.Error()
				if strings.Contains(errMsg, "syntax error") ||
					strings.Contains(errMsg, "no such table") ||
					strings.Contains(errMsg, "no such column") ||
					strings.Contains(errMsg, "constraint failed") {
					// This is likely a client error - bad SQL syntax or constraint violations
					sendJSONError(w, fmt.Sprintf("Invalid SQL query: %s", errMsg), http.StatusBadRequest)
				} else {
					// This is likely a server error - database issues, etc.
					sendJSONError(w, fmt.Sprintf("Error executing %s query: %s", strings.ToLower(queryType), errMsg), http.StatusInternalServerError)
				}
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
