package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestExec(t *testing.T) {
	// Create a temporary database file
	tmpFile, err := os.CreateTemp("", "test-db-*.sqlite")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create a test router
	router := httprouter.New()
	router.OPTIONS("/__/exec", Exec(tmpFile.Name()))

	// Create a test database with a table
	createTableBody := ExecBody{
		Query: "CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT, value INTEGER)",
	}
	createTableJSON, _ := json.Marshal(createTableBody)
	req, _ := http.NewRequest("OPTIONS", "/__/exec", bytes.NewBuffer(createTableJSON))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Insert test data
	insertDataBody := ExecBody{
		Query: "INSERT INTO test_table (name, value) VALUES ('test1', 100), ('test2', 200)",
	}
	insertDataJSON, _ := json.Marshal(insertDataBody)
	req, _ = http.NewRequest("OPTIONS", "/__/exec", bytes.NewBuffer(insertDataJSON))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test SELECT query
	selectBody := ExecBody{
		Query: "SELECT * FROM test_table",
	}
	selectJSON, _ := json.Marshal(selectBody)
	req, _ = http.NewRequest("OPTIONS", "/__/exec", bytes.NewBuffer(selectJSON))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse the response
	var selectResponse map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &selectResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check the response
	if selectResponse["status"] != "success" {
		t.Errorf("Expected status 'success', got %v", selectResponse["status"])
	}
	if selectResponse["type"] != "select" {
		t.Errorf("Expected type 'select', got %v", selectResponse["type"])
	}
	rows, ok := selectResponse["rows"].([]interface{})
	if !ok {
		t.Fatalf("Expected rows to be an array")
	}
	if len(rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(rows))
	}

	// Test PRAGMA query
	pragmaBody := ExecBody{
		Query: "PRAGMA table_info(test_table)",
	}
	pragmaJSON, _ := json.Marshal(pragmaBody)
	req, _ = http.NewRequest("OPTIONS", "/__/exec", bytes.NewBuffer(pragmaJSON))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse the response
	var pragmaResponse map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &pragmaResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check the response
	if pragmaResponse["status"] != "success" {
		t.Errorf("Expected status 'success', got %v", pragmaResponse["status"])
	}
	if pragmaResponse["type"] != "pragma" {
		t.Errorf("Expected type 'pragma', got %v", pragmaResponse["type"])
	}
	pragmaRows, ok := pragmaResponse["rows"].([]interface{})
	if !ok {
		t.Fatalf("Expected rows to be an array")
	}
	if len(pragmaRows) != 3 {
		t.Errorf("Expected 3 columns in table_info, got %d", len(pragmaRows))
	}

	// Test EXPLAIN query
	explainBody := ExecBody{
		Query: "EXPLAIN SELECT * FROM test_table",
	}
	explainJSON, _ := json.Marshal(explainBody)
	req, _ = http.NewRequest("OPTIONS", "/__/exec", bytes.NewBuffer(explainJSON))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse the response
	var explainResponse map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &explainResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check the response
	if explainResponse["status"] != "success" {
		t.Errorf("Expected status 'success', got %v", explainResponse["status"])
	}
	if explainResponse["type"] != "explain" {
		t.Errorf("Expected type 'explain', got %v", explainResponse["type"])
	}
	explainRows, ok := explainResponse["rows"].([]interface{})
	if !ok {
		t.Fatalf("Expected rows to be an array")
	}
	if len(explainRows) == 0 {
		t.Errorf("Expected at least one row in explain output")
	}
}
