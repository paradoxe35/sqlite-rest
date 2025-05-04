package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/paradoxe35/sqlite-rest/pkg/db"
)

// GetTables returns a list of all tables in the database
func GetTables(dbPath string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Create sql.DB instance
		db, err := db.Open(dbPath)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Database error: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Get tables
		tables, err := listTables(db)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Error listing tables: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// Convert to rows format for consistency
		var rows []map[string]interface{}
		for _, table := range tables {
			rows = append(rows, map[string]interface{}{
				"name": table,
			})
		}

		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Send response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"tables": tables,
			"count":  len(tables),
		})
	}
}

// GetTableSchema returns the schema of a specific table
func GetTableSchema(dbPath string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Create sql.DB instance
		db, err := db.Open(dbPath)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Database error: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Parse table name from params
		tableName := params.ByName("table")
		if tableName == "" {
			sendJSONError(w, "Missing table parameter", http.StatusBadRequest)
			return
		}

		// Check if table exists
		tables, err := listTables(db)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Error listing tables: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		tableExists := false
		for _, t := range tables {
			if t == tableName {
				tableExists = true
				break
			}
		}

		if !tableExists {
			sendJSONError(w, fmt.Sprintf("Table not found: %s", tableName), http.StatusNotFound)
			return
		}

		// Get table schema
		schema, err := getTableSchema(db, tableName)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Error getting table schema: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Send response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"table":  tableName,
			"schema": schema,
		})
	}
}

// GetDatabaseInfo returns general information about the database
func GetDatabaseInfo(dbPath string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Create sql.DB instance
		db, err := db.Open(dbPath)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Database error: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Get tables
		tables, err := listTables(db)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Error listing tables: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// Get SQLite version
		var version string
		err = db.QueryRow("SELECT sqlite_version()").Scan(&version)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Error getting SQLite version: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// Get database size (approximate)
		var pageCount, pageSize int64
		err = db.QueryRow("PRAGMA page_count").Scan(&pageCount)
		if err != nil {
			pageCount = -1
		}

		err = db.QueryRow("PRAGMA page_size").Scan(&pageSize)
		if err != nil {
			pageSize = -1
		}

		dbSize := pageCount * pageSize

		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Send response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":         "success",
			"sqlite_version": version,
			"table_count":    len(tables),
			"tables":         tables,
			"database_size":  dbSize,
			"database_path":  dbPath,
		})
	}
}

// GetForeignKeys returns foreign key relationships for a specific table
func GetForeignKeys(dbPath string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Create sql.DB instance
		db, err := db.Open(dbPath)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Database error: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Parse table name from params
		tableName := params.ByName("table")
		if tableName == "" {
			sendJSONError(w, "Missing table parameter", http.StatusBadRequest)
			return
		}

		// Check if table exists
		tables, err := listTables(db)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Error listing tables: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		tableExists := false
		for _, t := range tables {
			if t == tableName {
				tableExists = true
				break
			}
		}

		if !tableExists {
			sendJSONError(w, fmt.Sprintf("Table not found: %s", tableName), http.StatusNotFound)
			return
		}

		// Get foreign keys
		foreignKeys, err := getForeignKeys(db, tableName)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Error getting foreign keys: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Send response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":       "success",
			"table":        tableName,
			"foreign_keys": foreignKeys,
		})
	}
}

// GetApiVersion returns the API version
func GetApiVersion() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Send response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"version": "1.1.0",
		})
	}
}

// HealthCheck returns a simple health check response
func HealthCheck(dbPath string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Create sql.DB instance
		db, err := db.Open(dbPath)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Database error: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Check if database is accessible
		err = db.Ping()
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Database ping failed: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Send response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "API is healthy",
		})
	}
}

// Helper functions

// getTableSchema returns the schema of a specific table
func getTableSchema(db *sql.DB, tableName string) ([]map[string]interface{}, error) {
	// In SQLite, we can use PRAGMA table_info to get table schema
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schema []map[string]interface{}
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var dfltValue interface{}

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk); err != nil {
			return nil, err
		}

		column := map[string]interface{}{
			"cid":         cid,
			"name":        name,
			"type":        dataType,
			"notnull":     notNull == 1,
			"default_val": dfltValue,
			"pk":          pk == 1,
		}

		schema = append(schema, column)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schema, nil
}

// getForeignKeys returns foreign key relationships for a specific table
func getForeignKeys(db *sql.DB, tableName string) ([]map[string]interface{}, error) {
	// In SQLite, we can use PRAGMA foreign_key_list to get foreign keys
	rows, err := db.Query(fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var foreignKeys []map[string]interface{}
	for rows.Next() {
		var id, seq int
		var table, from, to string
		var onUpdate, onDelete string
		var match string

		if err := rows.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			return nil, err
		}

		foreignKey := map[string]interface{}{
			"id":        id,
			"seq":       seq,
			"table":     table,
			"from":      from,
			"to":        to,
			"on_update": onUpdate,
			"on_delete": onDelete,
			"match":     match,
		}

		foreignKeys = append(foreignKeys, foreignKey)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return foreignKeys, nil
}
