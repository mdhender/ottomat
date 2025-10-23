# OttoMat

A minimal web server built with Go backend and HTMX frontend, featuring session management and role-based authorization.

## Features

- **Role-Based Access Control**: Three user roles (guest, chief, admin) with appropriate dashboards
- **Session Management**: Secure session handling with HTTP-only cookies
- **Graceful Shutdown**: Proper signal handling (SIGINT, SIGTERM) with database flush time
- **Database Management**: Simple CLI commands for database initialization, migrations, and seeding
- **Modern Frontend**: HTMX for dynamic interactions, TailwindCSS for styling, Alpine.js for client-side behavior

## Technology Stack

- **Backend**: Go with standard library
- **CLI**: Cobra for command-line interface
- **ORM**: Ent with Atlas for migrations
- **Database**: SQLite (modernc.org/sqlite)
- **Frontend**: HTMX, Alpine.js, TailwindCSS
- **Authentication**: bcrypt for password hashing
- **Versioning**: github.com/maloquacious/semver

## Installation

### Prerequisites

- Go 1.21 or higher

### Build

Using Make:

```bash
make build
```

Or manually:

```bash
go build -o dist/local/ottomat ./cmd/ottomat
```

For Linux:

```bash
VERSION=$(./dist/local/ottomat version)
GOOS=linux GOARCH=amd64 go build -o dist/linux/ottomat-${VERSION} ./cmd/ottomat
```

### Quick Start with Test Database

To quickly set up a test database with a default admin user:

```bash
make new-database
```

This will:
1. Build the binary
2. Remove any existing test database
3. Create and migrate a new database at `testdata/ottomat.db`
4. Seed with admin user (username: `admin`, password: `happy.cat.happy.nap`)

### Makefile Targets

- `make help` - Show all available targets
- `make build` - Build the ottomat binary
- `make new-database` - Initialize a fresh test database with default admin
- `make clean` - Remove build artifacts and test database

## Database Setup

### Initialize Database

Create a new database file:

```bash
./dist/local/ottomat db init
```

### Run Migrations

Apply schema migrations:

```bash
./dist/local/ottomat db migrate
```

### Seed Database

Create default admin user with username `admin`:

```bash
# With specific password
./dist/local/ottomat db seed --password mypassword

# Generate random 6-word passphrase (will be logged)
./dist/local/ottomat db seed
```

### Create User

Create a new user with a username and optional password, role, and clan ID:

```bash
# Create user with auto-generated password (role defaults to 'guest')
./dist/local/ottomat db create user alice

# Create user with specific password
./dist/local/ottomat db create user bob --password secret123

# Create chief with clan ID
./dist/local/ottomat db create user charlie --role chief --clan-id 42

# Create admin user
./dist/local/ottomat db create user boss --role admin --password adminpass
```

### Update User

Update user fields by username. At least one field flag must be provided:

```bash
# Update password with specific value
./dist/local/ottomat db update user admin --password newpassword

# Generate random password
./dist/local/ottomat db update user admin --password ""

# Update role
./dist/local/ottomat db update user alice --role chief

# Update clan ID
./dist/local/ottomat db update user alice --clan-id 42

# Update multiple fields at once
./dist/local/ottomat db update user bob --password secret123 --role chief --clan-id 99

# Clear clan ID (set to NULL)
./dist/local/ottomat db update user alice --clan-id 0
```

### Database Options

All database commands accept a `--db` flag to specify the database file path:

```bash
./dist/local/ottomat db init --db /path/to/database.db
./dist/local/ottomat db migrate --db /path/to/database.db
./dist/local/ottomat db seed --db /path/to/database.db
./dist/local/ottomat db create user alice --db /path/to/database.db
./dist/local/ottomat db update user admin --db /path/to/database.db
```

Default: `ottomat.db`

## Running the Server

### Start Server

```bash
./dist/local/ottomat server
```

Server will start on port 8080 by default. Access at http://localhost:8080

### Server Options

```bash
./dist/local/ottomat server --port 3000              # Custom port
./dist/local/ottomat server --db custom.db           # Custom database
./dist/local/ottomat server --timeout 5m             # Auto-shutdown after 5 minutes (testing)
./dist/local/ottomat server --dev                    # Development mode (disables password managers)
```

**Development Mode**: When `--dev` is enabled:
- HTTP request logging is enabled, showing method, path, status code, and response time
- Example: `2025/10/23 16:12:26 [GET] /login 200 107.167µs`

**Visible Passwords**: Use `--visible-passwords` with `--dev` to show passwords as plain text:
```bash
./dist/local/ottomat server --dev --visible-passwords
```
This prevents password managers from interfering during testing. Cannot be used without `--dev`.

## User Roles

### Guest
- Not logged in
- Always redirected to login page

### Chief
- Standard user role
- Can view personal dashboard showing:
  - Username
  - Clan number (if assigned)
  - Logout option

### Admin
- Full administrative access
- Can view admin dashboard showing:
  - List of all users
  - Add new users (with username, password, role, optional clan ID)
  - Delete existing users

## API Endpoints

### Public
- `GET /login` - Login page
- `POST /login` - Process login credentials

### Authenticated
- `GET /` - Dashboard (redirects based on role)
- `POST /logout` - Logout and clear session

### Admin Only
- `GET /admin` - Admin dashboard
- `POST /admin/users` - Create new user
- `DELETE /admin/users/{id}` - Delete user

## Development

### Project Structure

```
ottomat/
├── cmd/                        # Cobra commands
│   ├── root.go                # Root command
│   ├── version.go             # Version command
│   ├── version_info.go        # Version constants
│   ├── server.go              # Server command
│   └── db.go                  # Database commands
├── ent/                        # Ent ORM generated code
│   └── schema/                # Schema definitions
│       ├── user.go            # User entity
│       └── session.go         # Session entity
├── internal/
│   ├── auth/                  # Authentication utilities
│   │   └── auth.go            # Session token generation
│   ├── database/              # Database utilities
│   │   └── database.go        # DB connection and migration
│   └── server/                # HTTP server
│       ├── server.go          # Server setup and routing
│       ├── handlers/          # HTTP handlers
│       │   ├── auth.go        # Login/logout handlers
│       │   ├── dashboard.go   # Chief dashboard
│       │   └── admin.go       # Admin dashboard
│       └── middleware/        # HTTP middleware
│           ├── session.go     # Session validation
│           └── auth.go        # Authorization
├── main.go                     # Application entry point
└── README.md                   # This file
```

### Database Schema

#### User Table
- `id` - Auto-incrementing primary key
- `username` - Unique username
- `password_hash` - bcrypt hashed password
- `role` - Enum: guest, chief, admin
- `clan_id` - Optional integer for chiefs
- `created_at` - Timestamp
- `updated_at` - Timestamp

#### Session Table
- `id` - Auto-incrementing primary key
- `token` - Unique session token (base64 encoded, 32 bytes)
- `user_id` - Foreign key to users table
- `expires_at` - Session expiration timestamp
- `created_at` - Timestamp

### Commands

```bash
# Format code
go fmt ./...

# Run tests
go test ./...

# Build
make build

# Version info
./dist/local/ottomat version
./dist/local/ottomat version --build-info

# Quick test database setup
make new-database

# Clean build artifacts
make clean
```

## Security Features

- **Password Hashing**: bcrypt with default cost
- **Session Tokens**: 32-byte cryptographically secure random tokens
- **HTTP-Only Cookies**: Session cookies not accessible via JavaScript
- **SameSite Lax**: CSRF protection
- **Session Expiration**: 24-hour session lifetime
- **Role-Based Access**: Middleware enforces authorization

## License

See [LICENSE](LICENSE) file for details.
