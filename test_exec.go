package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type ExecBody struct {
	Query string `json:"query"`
}

func main() {
	// Create a temporary database
	err := os.MkdirAll("./data", 0755)
	if err != nil {
		fmt.Printf("Error creating data directory: %v\n", err)
		return
	}

	// Start the server in the background
	go func() {
		cmd := exec.Command("./sqlite-rest -p 8082")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error starting server: %v\n", err)
		}
	}()

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	// Test queries
	testQueries := []struct {
		name  string
		query string
	}{
		{"CREATE TABLE", "CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT, value INTEGER)"},
		{"INSERT DATA", "INSERT INTO test_table (name, value) VALUES ('test1', 100), ('test2', 200)"},
		{"SELECT", "SELECT * FROM test_table"},
		{"PRAGMA", "PRAGMA table_info(test_table)"},
		{"EXPLAIN", "EXPLAIN SELECT * FROM test_table"},
		{"ANALYZE", "ANALYZE test_table"},
	}

	for _, test := range testQueries {
		fmt.Printf("\nTesting %s query...\n", test.name)

		// Create request body
		body := ExecBody{
			Query: test.query,
		}
		bodyJSON, _ := json.Marshal(body)

		// Send request
		req, err := http.NewRequest("OPTIONS", "http://localhost:8082/__/exec", bytes.NewBuffer(bodyJSON))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			continue
		}

		// Read response
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// Print response
		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Printf("Response: %s\n", string(respBody))
	}
}
