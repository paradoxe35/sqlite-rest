# SQLite REST

[![GitHub release (latest by date)](https://img.shields.io/github/v/release/paradoxe35/sqlite-rest)](https://github.com/paradoxe35/sqlite-rest/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/paradoxe35/sqlite-rest)](https://goreportcard.com/report/github.com/paradoxe35/sqlite-rest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lightweight, high-performance REST API for SQLite databases. Expose CRUD operations and database metadata over HTTP with minimal configuration.

## Features

- **Full CRUD Operations**: Create, read, update, and delete records via REST endpoints
- **Metadata API**: Explore database structure, table schemas, and relationships
- **SQL Execution**: Run arbitrary SQL queries with security controls
- **Filtering & Pagination**: Filter records with SQL or JSON syntax, with pagination support
- **Authentication**: Basic auth support for securing your API
- **Cross-Platform**: Available for Windows, macOS, and Linux (including ARM support)
- **Docker Support**: Easy deployment with Docker
- **Minimal Footprint**: Small binary size and low memory usage

## Quick Start

```bash
# Download and run (Linux/macOS)
curl -L https://github.com/paradoxe35/sqlite-rest/releases/latest/download/sqlite-rest_v1.1.0_linux-amd64.tar.gz | tar xz
./sqlite-rest

# Or with Docker
docker run -p 8080:8080 -v "$(pwd)"/data.sqlite:/app/data/data.sqlite:rw ghcr.io/paradoxe35/sqlite-rest
```

## Installation

### From releases page

Download the binary for your platform from the [releases page](https://github.com/paradoxe35/sqlite-rest/releases/latest).

```bash
# Linux/macOS
curl -L https://github.com/paradoxe35/sqlite-rest/releases/latest/download/sqlite-rest_v1.1.0_linux-amd64.tar.gz | tar xz
chmod +x sqlite-rest
./sqlite-rest

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/paradoxe35/sqlite-rest/releases/latest/download/sqlite-rest_v1.1.0_windows-amd64.zip" -OutFile "sqlite-rest.zip"
Expand-Archive -Path "sqlite-rest.zip" -DestinationPath "."
.\sqlite-rest.exe
```

### From source

```bash
# Clone the repository
git clone https://github.com/paradoxe35/sqlite-rest.git
cd sqlite-rest

# Build the binary
go build -o sqlite-rest ./cmd/sqlite-rest.go

# Run the server
./sqlite-rest
```

### From Docker

```bash
# Pull and run the image
docker run -p 8080:8080 -v "$(pwd)"/data.sqlite:/app/data/data.sqlite:rw ghcr.io/paradoxe35/sqlite-rest

# Or with docker-compose
docker-compose up -d
```

## CLI Usage

```bash
# Show help
sqlite-rest --help

# Show version
sqlite-rest --version

# Run with custom port and database path
sqlite-rest -p 3000 -f ./path/to/database.sqlite

# Run with default settings (port 8080, database at ./data/data.sqlite)
sqlite-rest
```

## Authentication

SQLite REST supports Basic Authentication. To enable it, set the following environment variables:

- `SQLITE_REST_USERNAME`: The username for Basic Authentication
- `SQLITE_REST_PASSWORD`: The password for Basic Authentication

If both variables are set, Basic Authentication will be enabled. If either variable is not set, authentication will be disabled.

Example with Docker:

```bash
$ docker run -p 8080:8080 -v "$(pwd)"/data.sqlite:/app/data/data.sqlite:rw \
  -e SQLITE_REST_USERNAME=admin \
  -e SQLITE_REST_PASSWORD=secret \
  ghcr.io/paradoxe35/sqlite-rest
```

Example with Docker Compose:

Create a `docker-compose.yml` file:

```yaml
version: "3.7"

services:
  sqlite-rest:
    image: ghcr.io/paradoxe35/sqlite-rest
    container_name: sqlite-rest
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data:rw  # Mount a directory for the database
    environment:
      - SQLITE_REST_USERNAME=admin
      - SQLITE_REST_PASSWORD=secret
      # Optional: Customize dangerous operations
      - SQLITE_REST_DANGEROUS_OPS=DROP TABLE,DELETE FROM
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/__/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 5s

volumes:
  sqlite-data:
    driver: local
```

Then run:

```bash
mkdir -p data
docker-compose up -d
```

## API

# Core API

[Search all records](#search-all-records) - `GET /:table` <br>
[Get record by id](#get-record-by-id) - `GET /:table/:id` <br>
[Create record](#create-record) - `POST /:table` <br>
[Update record by id](#update-record) - `PATCH /:table/:id` <br>
[Delete record by id](#delete-record) - `DELETE /:table/:id` <br>
[Execute arbitrary query](#execute-arbitrary-query) - `OPTIONS /__/exec` <br>

# Metadata API

[List all tables](#list-all-tables) - `GET /__/tables` <br>
[Get table schema](#get-table-schema) - `GET /__/tables/:table` <br>
[Get foreign keys](#get-foreign-keys) - `GET /__/tables/:table/foreign-keys` <br>
[Get database info](#get-database-info) - `GET /__/db` <br>

# Utility API

[Health check](#health-check) - `GET /__/health` <br>
[API version](#api-version) - `GET /__/version` <br>

### Search all records

Get all record in a table.<br>

Request: `GET /:table`<br>

Basic example:<br>

```bash
$ curl localhost:8080/cats

{
  "data": [
    { "id": 1, "name": "Tequila", "paw": 4 },
    { "id": 2, "name": "Whisky", "paw": 3 }
  ],
  "limit": null,
  "offset": null,
  "total_rows": 2
}

```

**Optional parameters:**<br>

- `offset`: Offset the number of records returned. Default: `0`
- `limit`: Limit the number of records returned. Default: not set
- `order_by`: Order the records by a column. Default: `id`
- `order_dir`: Order the records by a column. Default: `asc`
- `columns`: Select only the specified columns. Default: `*`
- `filters_raw`: Filter the records by a raw SQL query. Must be URIescaped.
- `filters`: Filter the records by a JSON object. Must be URIescaped.

**Filters:**<br>

Can be passed as a JSON object or as a raw WHERE clause. The JSON object is more convenient to use, the raw query is more flexible. Both must be URIescaped. Cannot be used together. Filters provided by `filters` param are joined with `AND` operator.

Example with `filters_raw` parameter in cURL:<br>

```bash
$ curl "localhost:8080/cats?offset=10&limit=2&order_by=name&order_dir=desc&filters_raw=paw%20%3D%204%20OR%20name%20LIKE%20'%25Tequila%25'"

{
  "data": [
    { "id": 10, "name": "Tequila", "paw": 4 },
    { "id": 11, "name": "Cognac", "paw": 4 }
  ],
  "limit": 2,
  "offset": 10,
  "total_rows": 2
}
```

Example with `filters_raw` parameter in JS:<br>

```js
const filters = "paw = 4 OR name LIKE '%Tequila%'"

fetch(`http://localhost:8080/cats?filters_raw=${encodeURIComponent(filters)}`)
```

Example with `filters` parameter in JS:<br>

```js
const filters = [
  {
    "column": "paw",
    "operator": "=",
    "value": 4
  },
  {
    "column": "name",
    "operator": "LIKE",
    "value": "%Tequila%"
  }
]

fetch(`http://localhost:8080/cats?filters=${encodeURIComponent(JSON.stringify(filters))}`)
```

### Get record by id

Get a record by its id in a table.<br>

Request: `GET /:table/:id`<br>

Example:<br>

```bash
$ curl localhost:8080/cats/1

{
  "id": 1,
  "name": "Tequila",
  "paw": 4
}
```

**Optional parameters:**<br>

- `columns`: Select only the specified columns. Default: `*`

Example with parameters:<br>

```bash
$ curl localhost:8080/cats/1?columns=name,paw

{
  "name": "Tequila",
  "paw": 4
}
```

### Create record

Create a record in a table.<br>

Request: `POST /:table`<br>

Example:<br>

```bash
$ curl -X POST -H "Content-Type: application/json" -d '{"name": "Tequila", "paw": 4}' localhost:8080/cats

{
  "id": 1,
}
```

### Update record

Update a record in a table.<br>
⚠️ The update is a PATCH, not a PUT. Only the fields passed in the body will be updated. The other fields will be left untouched.

Request: `PATCH /:table/:id`<br>

Example:<br>

```bash
$ curl -X PATCH -H "Content-Type: application/json" -d '{"name": "Tequila", "paw": 4}' localhost:8080/cats/1

{
  "id": 1,
}
```

### Delete record

Delete a record in a table.<br>

Request: `DELETE /:table/:id`<br>

Example:<br>

```bash
$ curl -X DELETE localhost:8080/cats/1

{
  "id": 1,
}
```

### Execute arbitrary query

Execute an arbitrary query. ⚠️ Experimental<br>

Request: `OPTIONS /__/exec`<br>

This endpoint is protected by authentication when enabled. It allows executing SQL queries and returns the results.

For security reasons, the following operations are blocked by default:
- DROP TABLE
- DROP DATABASE
- DELETE FROM
- TRUNCATE TABLE
- ALTER TABLE
- PRAGMA
- ATTACH DATABASE
- DETACH DATABASE

You can customize the list of dangerous operations by setting the `SQLITE_REST_DANGEROUS_OPS` environment variable. This should be a comma-separated list of SQL operations to block. For example:

```
SQLITE_REST_DANGEROUS_OPS="DROP TABLE,DELETE FROM"
```

To allow all operations (use with caution), set the variable to an empty string:

```
SQLITE_REST_DANGEROUS_OPS=""
```

Example of creating a table:<br>

```bash
$ curl -X OPTIONS -H "Content-Type: application/json" -d '{"query": "CREATE TABLE cats (id INTEGER PRIMARY KEY, name TEXT, paw INTEGER)"}' localhost:8080/__/exec

{
  "status": "success",
  "type": "create",
  "rows_affected": 0
}
```

Example of inserting data:<br>

```bash
$ curl -X OPTIONS -H "Content-Type: application/json" -d '{"query": "INSERT INTO cats (name, paw) VALUES (\"Tequila\", 4)"}' localhost:8080/__/exec

{
  "status": "success",
  "type": "insert",
  "rows_affected": 1
}
```

Example of selecting data:<br>

```bash
$ curl -X OPTIONS -H "Content-Type: application/json" -d '{"query": "SELECT * FROM cats"}' localhost:8080/__/exec

{
  "status": "success",
  "type": "select",
  "rows": [
    {
      "id": 1,
      "name": "Tequila",
      "paw": 4
    }
  ],
  "count": 1
}
```

Example of listing tables:<br>

```bash
$ curl -X OPTIONS -H "Content-Type: application/json" -d '{"query": "SHOW TABLES"}' localhost:8080/__/exec

{
  "status": "success",
  "type": "show_tables",
  "tables": ["cats", "dogs", "birds"],
  "rows": [
    {"table_name": "cats"},
    {"table_name": "dogs"},
    {"table_name": "birds"}
  ],
  "count": 3
}
```

You can also use `LIST TABLES` as an alternative to `SHOW TABLES`.

### List all tables

Get a list of all tables in the database.

Request: `GET /__/tables`

Example:

```bash
$ curl localhost:8080/__/tables

{
  "status": "success",
  "tables": ["cats", "dogs", "birds"],
  "count": 3
}
```

### Get table schema

Get the schema of a specific table.

Request: `GET /__/tables/:table`

Example:

```bash
$ curl localhost:8080/__/tables/cats

{
  "status": "success",
  "table": "cats",
  "schema": [
    {
      "cid": 0,
      "name": "id",
      "type": "INTEGER",
      "notnull": false,
      "default_val": null,
      "pk": 1
    },
    {
      "cid": 1,
      "name": "name",
      "type": "TEXT",
      "notnull": false,
      "default_val": null,
      "pk": 0
    },
    {
      "cid": 2,
      "name": "paw",
      "type": "INTEGER",
      "notnull": false,
      "default_val": null,
      "pk": 0
    }
  ]
}
```

### Get foreign keys

Get the foreign key relationships for a specific table.

Request: `GET /__/tables/:table/foreign-keys`

Example:

```bash
$ curl localhost:8080/__/tables/cats/foreign-keys

{
  "status": "success",
  "table": "cats",
  "foreign_keys": [
    {
      "id": 0,
      "seq": 0,
      "table": "owners",
      "from": "owner_id",
      "to": "id",
      "on_update": "NO ACTION",
      "on_delete": "NO ACTION",
      "match": "NONE"
    }
  ]
}
```

### Get database info

Get general information about the database.

Request: `GET /__/db`

Example:

```bash
$ curl localhost:8080/__/db

{
  "status": "success",
  "sqlite_version": "3.36.0",
  "table_count": 3,
  "tables": ["cats", "dogs", "birds"],
  "database_size": 16384,
  "database_path": "./data/data.sqlite"
}
```

### Health check

Check if the API is healthy.

Request: `GET /__/health`

Example:

```bash
$ curl localhost:8080/__/health

{
  "status": "success",
  "message": "API is healthy"
}
```

### API version

Get the API version.

Request: `GET /__/version`

Example:

```bash
$ curl localhost:8080/__/version

{
  "status": "success",
  "version": "1.1.0"
}
```

## Credits

This project was inspired by and builds upon [jonamat/sqlite-rest](https://github.com/jonamat/sqlite-rest). See [CREDITS.md](CREDITS.md) for more details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.