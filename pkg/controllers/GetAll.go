package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/paradoxe35/sqlite-rest/pkg/db"
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

		// Parse columns from params or use all
		var columnsSelect string
		columnsParam := r.URL.Query().Get("cols")
		if columnsParam == "" {
			columnsSelect = "*"
		} else {
			columnsSelect = columnsParam
		}

		// Parse filters_raw from query string and build WHERE clause
		var whereClause string
		filtersParam := r.URL.Query().Get("filters_raw")

		if filtersParam != "" {
			unescapedFilters, err := url.QueryUnescape(filtersParam)
			if err != nil {
				sendJSONError(w, fmt.Sprintf("Error unescaping filters_raw: %s", err.Error()), http.StatusBadRequest)
				return
			}
			whereClause = "WHERE " + unescapedFilters
		}

		filtersStruct := r.URL.Query().Get("filters")
		if filtersStruct != "" {
			if filtersParam != "" {
				sendJSONError(w, "Cannot use both filters and filters_raw parameters", http.StatusBadRequest)
				return
			}

			filterArr := []Filter{}

			unescapedFilters, err := url.QueryUnescape(filtersStruct)
			if err != nil {
				sendJSONError(w, fmt.Sprintf("Error unescaping filters: %s", err.Error()), http.StatusBadRequest)
				return
			}
			err = json.Unmarshal([]byte(unescapedFilters), &filterArr)
			if err != nil {
				sendJSONError(w, fmt.Sprintf("Invalid filters format: %s", err.Error()), http.StatusBadRequest)
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
			sendJSONError(w, "Cannot use offset parameter without limit parameter", http.StatusBadRequest)
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
			sendJSONError(w, "Cannot use order_dir parameter without order_by parameter", http.StatusBadRequest)
			return
		}

		// Execute query
		query := fmt.Sprintf("SELECT %s FROM %s %s %s %s %s %s", columnsSelect, tableSelect, whereClause, orderByClause, orderDir, limitClause, offsetClause)
		rows, err := db.Query(query)
		if err != nil {
			// Check if this is a syntax error (client error) or a server error
			errMsg := err.Error()
			if strings.Contains(errMsg, "no such table") {
				sendJSONError(w, fmt.Sprintf("Table not found: %s", tableSelect), http.StatusBadRequest)
			} else if strings.Contains(errMsg, "no such column") {
				sendJSONError(w, fmt.Sprintf("Invalid column in query: %s", errMsg), http.StatusBadRequest)
			} else if strings.Contains(errMsg, "syntax error") {
				sendJSONError(w, fmt.Sprintf("SQL syntax error: %s", errMsg), http.StatusBadRequest)
			} else {
				sendJSONError(w, fmt.Sprintf("Error executing query: %s", errMsg), http.StatusInternalServerError)
			}
			return
		}
		defer rows.Close()

		// Get column names
		var columnNames []string
		columnNames, err = rows.Columns()
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Error retrieving column names: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// Get column types
		columnTypes, err := rows.ColumnTypes()
		if err != nil {
			sendJSONError(w, fmt.Sprintf("Error retrieving column types: %s", err.Error()), http.StatusInternalServerError)
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
				sendJSONError(w, fmt.Sprintf("Error scanning row data: %s", err.Error()), http.StatusInternalServerError)
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

		// Check for errors from iterating over rows
		if err = rows.Err(); err != nil {
			sendJSONError(w, fmt.Sprintf("Error iterating over rows: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// Compose response and return data
		response := map[string]interface{}{
			"status":     "success",
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
			sendJSONError(w, fmt.Sprintf("Error encoding response: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	}
}
