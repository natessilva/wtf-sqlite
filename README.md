# SQLite Web Application

A Go web application using SQLite for data storage and Templ for HTML templating. It provides a practical example for a very simple and fast DX and UX.

## Project Overview

This application provides user authentication and a "dials" interface where users can create and modify personal dial objects. It uses:

- **SQLite** for data storage
- **Templ** for HTML templating
- **SQLC** for type-safe SQL queries
- **Go HTTP Router** for routing
- **Prometheus** for metrics

## Prerequisites

- Go 1.24.6 or higher
- SQLite 3
- [sqlc](https://sqlc.dev/) - For generating Go code from SQL
- [templ](https://templ.guide) - For HTML templating

## Development

### Setup

Clone the repository and install dependencies:

```bash
git clone <repository-url>
cd sqlite
go mod download
```

### Build and Run

To build and run the application:

```bash
make run
```

This will:

1. Generate Go code from SQL queries using sqlc
2. Generate Go code from templ templates
3. Run the application on port 8000

### Testing

Run tests with:

```bash
make test
```

### Database Migrations

To create a new database migration:

```bash
./scripts/new-migration.sh migration_name
```

This creates a timestamped SQL file in the `db/migrations` directory that you can edit to define your schema changes.

## Project Structure

- `assets/` - Static assets like CSS
- `cmd/server/` - Application entry point
- `db/` - Database related files
  - `migrations/` - SQL migration files
  - `queries/` - SQLC query definitions
- `model/` - Generated Go types and query methods
- `scripts/` - Utility scripts
- `templates/` - Templ template files

## Key Components

### Database Schema

The application has three main tables:

- `user` - User authentication data
- `dial` - User-specific dial objects
- `session` - Authentication sessions

### Code Generation

The project uses two code generation tools:

1. **sqlc** - Generates type-safe Go code from SQL in `db/queries/`
2. **templ** - Generates Go code from HTML templates in `templates/`

### Authentication

The application provides user signup, login, and session management.

## Deployment

To deploy the application:

```bash
make deploy
```

This script:

1. Builds a statically-linked Linux binary
2. Copies it to the remote server (`silva.world`)
3. Stops the current service
4. Replaces the binary
5. Restarts the service

## Environment Variables

- `ENV=prod` - Set to "prod" for production mode with TLS support

## Production Setup

In production mode, the application:

- Serves on ports 80 (redirects to HTTPS) and 443 (HTTPS)
- Uses Let's Encrypt for automatic TLS certificate management
- Requires the domain `silva.world`

## Development Mode

In development mode, the application:

- Serves on port 8000
- Does not use TLS

## Metrics

Prometheus metrics are available at `/metrics` on port 6060.
