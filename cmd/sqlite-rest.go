package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/paradoxe35/sqlite-rest/pkg/controllers"
	"github.com/paradoxe35/sqlite-rest/pkg/middleware"
)

const (
	VERSION         = "1.0.0"
	DEFAULT_PORT    = "8080"
	DEFAULT_DB_PATH = "./data/data.sqlite"
)

var help = flag.Bool("help", false, "Show help")
var port = flag.String("p", DEFAULT_PORT, "Port to listen on")
var dbPath = flag.String("f", DEFAULT_DB_PATH, "Path to sqlite database file")

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Ensure the directory exists
	dbDir := filepath.Dir(*dbPath)
	if dbDir != "." {
		err := os.MkdirAll(dbDir, 0755)
		if err != nil {
			log.Fatal("Error creating directory for database: " + err.Error())
		}
	}

	_, err := os.Stat(*dbPath)
	if err != nil {
		// Create db if not exits
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("Database not found. Creating new one in %s\n", *dbPath)
			_, err := os.Create(*dbPath)
			if err != nil {
				log.Fatal("Error creating sqlite file. " + err.Error())
			}
		} else {
			log.Fatal("Error reading sqlite file. " + err.Error())
		}
	}
	log.Printf("Using database in %s\n", *dbPath)

	// Create a custom router that can handle both API and data routes
	router := middleware.NewCustomRouter()

	// Metadata endpoints
	router.GET("/api/tables", controllers.GetTables(*dbPath))
	router.GET("/api/tables/:table", controllers.GetTableSchema(*dbPath))
	router.GET("/api/tables/:table/foreign-keys", controllers.GetForeignKeys(*dbPath))
	router.GET("/api/db", controllers.GetDatabaseInfo(*dbPath))

	// Utility endpoints
	router.GET("/api/health", controllers.HealthCheck(*dbPath))
	router.GET("/api/version", controllers.GetApiVersion())

	// SQL execution endpoint
	router.OPTIONS("/api/exec", controllers.Exec(*dbPath))

	// Core CRUD endpoints
	router.GET("/:table", controllers.GetAll(*dbPath))
	router.GET("/:table/:id", controllers.Get(*dbPath))
	router.POST("/:table", controllers.Create(*dbPath))
	router.PATCH("/:table/:id", controllers.Update(*dbPath))
	// router.PUT("/:table/:id", controllers.Update(*dbPath))
	router.DELETE("/:table/:id", controllers.Delete(*dbPath))

	// Check if authentication is enabled
	username := os.Getenv("SQLITE_REST_USERNAME")
	password := os.Getenv("SQLITE_REST_PASSWORD")
	if username != "" && password != "" {
		log.Println("Basic Authentication enabled")
	}

	// Create a handler with the router
	handler := middleware.BasicAuth(router)

	log.Println("Listening on port " + *port)
	log.Fatal(http.ListenAndServe(":"+*port, handler))
}
