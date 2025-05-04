# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v1.1.0

Released on 2025-05-04

### Added
- New metadata API endpoints:
  - `/api/tables` - List all tables in the database
  - `/api/tables/:table` - Get the schema of a specific table
  - `/api/tables/:table/foreign-keys` - Get foreign key relationships for a specific table
  - `/api/db` - Get general database information
- New utility API endpoints:
  - `/api/health` - Health check endpoint
  - `/api/version` - API version endpoint
- Custom router to handle both API and data routes
- Improved build process with support for more platforms
- SHA256 checksums for release artifacts

### Changed
- Moved SQL execution endpoint to `/api/exec` for better organization
- Updated documentation with examples for all new endpoints
- Improved error handling in API responses
- Updated Go dependencies

### Fixed
- Fixed routing conflicts between API and data endpoints
- Improved error messages for better debugging

## v1.0.0

Released on 2020-04-05

### Added
- Initial release with basic CRUD operations
- Support for SQLite database
- Basic authentication
- Filtering and pagination
- Custom SQL execution endpoint
