# Snippetbox

A web application for sharing code snippets built with Go. Users can create, view, and manage code snippets with expiration dates and user authentication.

## Features

- 📝 Create and share code snippets with configurable expiration (1, 7, or 365 days)
- 👤 User registration and authentication with secure password hashing
- 🔒 Session-based authentication with CSRF protection
- 🎨 Clean, responsive web interface with HTML templates
- 🔐 HTTPS-only with TLS configuration
- 💾 MySQL database storage with connection pooling

## Quick Start

### Using Docker (Recommended)

```bash
# Start the application with MySQL
docker-compose up

# The app will be available at https://localhost:4001
```

### Manual Setup

1. **Set up MySQL database:**
   ```bash
   mysql -u root -p < db/create.sql
   ```

2. **Set environment variable:**
   ```bash
   export SNIP_DB_DSN="username:password@tcp(localhost:3306)/snipdb?parseTime=true"
   ```

3. **Build and run:**
   ```bash
   make build
   ./bin/web -addr=:4001
   ```

## Development

### Building
```bash
make build
```

### Testing
```bash
go test ./...
go test ./cmd/web -v  # Run web tests with verbose output
```

### API Testing
```bash
make test_get_view      # Test snippet viewing
make httpie_get_create  # Test snippet creation form
make httpie_post_create # Test snippet creation
```

## Project Structure

- `cmd/web/` - HTTP server, handlers, middleware, and routing
- `internal/models/` - Data models and database layer
- `internal/validator/` - Form validation utilities
- `ui/` - HTML templates and static assets (embedded)
- `db/` - Database schema and migrations
- `tls/` - TLS certificates for HTTPS

## Technology Stack

- **Backend:** Go with standard library + minimal dependencies
- **Database:** MySQL 8.0
- **Authentication:** Session-based with bcrypt password hashing
- **Security:** CSRF protection, secure headers, TLS
- **Frontend:** HTML templates with embedded CSS/JS
