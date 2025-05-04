package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestGetTables(t *testing.T) {
	// Create a temporary database file
	tmpFile, err := os.CreateTemp("", "test-db-*.sqlite")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create a test router
	router := httprouter.New()
	router.GET("/__/tables", GetTables(tmpFile.Name()))

	// Create a test request
	req, err := http.NewRequest("GET", "/__/tables", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check the response fields
	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got %v", response["status"])
	}

	if _, ok := response["tables"]; !ok {
		t.Errorf("Expected 'tables' field in response")
	}

	if _, ok := response["count"]; !ok {
		t.Errorf("Expected 'count' field in response")
	}
}

func TestHealthCheck(t *testing.T) {
	// Create a temporary database file
	tmpFile, err := os.CreateTemp("", "test-db-*.sqlite")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create a test router
	router := httprouter.New()
	router.GET("/__/health", HealthCheck(tmpFile.Name()))

	// Create a test request
	req, err := http.NewRequest("GET", "/__/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check the response fields
	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got %v", response["status"])
	}

	if response["message"] != "API is healthy" {
		t.Errorf("Expected message 'API is healthy', got %v", response["message"])
	}
}

func TestGetApiVersion(t *testing.T) {
	// Create a test router
	router := httprouter.New()
	router.GET("/__/version", GetApiVersion())

	// Create a test request
	req, err := http.NewRequest("GET", "/__/version", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check the response fields
	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got %v", response["status"])
	}

	if _, ok := response["version"]; !ok {
		t.Errorf("Expected 'version' field in response")
	}
}
