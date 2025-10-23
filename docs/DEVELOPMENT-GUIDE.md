# 🛠️ Development Guide - NAT Management System

## Table of Contents

- [Development Environment Setup](#development-environment-setup)
- [Project Structure](#project-structure)
- [Coding Standards](#coding-standards)
- [Adding New Features](#adding-new-features)
- [Testing](#testing)
- [Debugging](#debugging)
- [Database Migrations](#database-migrations)
- [Common Tasks](#common-tasks)

---

## Development Environment Setup

### Prerequisites

- **Go**: 1.24.0 or higher
- **PostgreSQL**: 15+ (or Supabase account)
- **Git**: For version control
- **Code Editor**: VS Code (recommended) or any Go-compatible IDE
- **MikroTik Router**: For testing (or use emulator)

### Step 1: Install Go

```bash
# Download from https://go.dev/dl/
# Verify installation
go version

# Setup GOPATH (if not set)
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

### Step 2: Clone Repository

```bash
git clone <repository-url>
cd nat-management-app
```

### Step 3: Install Dependencies

```bash
# Download all dependencies
go mod download

# Verify dependencies
go mod verify

# Tidy up if needed
go mod tidy
```

### Step 4: Setup Database

#### Using Supabase (Recommended for Development)

1. Sign up at [neon.tech](https://neon.tech)
2. Create new project
3. Copy connection string
4. Create `.env` file (see below)

#### Using Local PostgreSQL

```bash
# Install PostgreSQL
# Create database
createdb nat_management

# Run migrations (automatic on first run)
# Or manually:
psql -U postgres -d nat_management -f migrations/init.sql
```

### Step 5: Configure Environment

Create `.env` file in root directory:

```env
# Development Configuration
SERVER_HOST=localhost
SERVER_PORT=8080
DEBUG=true

# Database
DATABASE_URL=postgresql://user:password@host/nat_management?sslmode=require

# JWT (use strong random strings for production)
JWT_SECRET=dev-secret-key-min-32-chars-long-change-in-production
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h

# Session
SESSION_SECRET=dev-session-secret-change-in-production
SESSION_MAX_AGE=86400

# CORS (allow localhost for development)
ALLOWED_ORIGINS=http://localhost:8080,http://127.0.0.1:8080

# Rate Limiting (relaxed for development)
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_DURATION=60s
```

### Step 6: Run Application

```bash
# Run directly
go run ./cmd/main.go

# Or build first
go build -o nat-supabase.exe ./cmd
./nat-supabase.exe
```

Application will start on http://localhost:8080

### Step 7: VS Code Setup (Optional)

Install recommended extensions:

```json
{
  "recommendations": [
    "golang.go",
    "eamodio.gitlens",
    "ms-vscode.vscode-typescript-next",
    "esbenp.prettier-vscode"
  ]
}
```

Configure `.vscode/settings.json`:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "editor.formatOnSave": true,
  "go.testFlags": ["-v"]
}
```

---

## Project Structure

### Directory Organization

```
nat-management-app/
├── cmd/                    # Application entry points
├── config/                 # Configuration management
├── internal/               # Private application code
│   ├── api/               # HTTP handlers (controllers)
│   ├── database/          # Database connection & pooling
│   ├── middleware/        # HTTP middleware
│   ├── models/            # Data models & DTOs
│   ├── services/          # Business logic layer
│   └── utils/             # Helper functions
├── web/                   # Frontend assets
│   ├── static/           # CSS, JS, images
│   └── templates/        # HTML templates
├── tools/                 # CLI tools
├── docs/                  # Documentation
└── migrations/            # Database migrations
```

### Layer Architecture

```
┌─────────────────────────────────────────┐
│  Presentation Layer (web/templates)     │
├─────────────────────────────────────────┤
│  API Layer (internal/api)               │
│  - Handlers                             │
│  - Request validation                   │
│  - Response formatting                  │
├─────────────────────────────────────────┤
│  Middleware Layer (internal/middleware) │
│  - Authentication                       │
│  - Authorization                        │
│  - Rate limiting                        │
│  - Logging                              │
├─────────────────────────────────────────┤
│  Business Logic Layer (internal/services)│
│  - Services                             │
│  - Business rules                       │
│  - Data transformation                  │
├─────────────────────────────────────────┤
│  Data Layer (internal/database)         │
│  - Database connection                  │
│  - Query execution                      │
│  - Transaction management               │
└─────────────────────────────────────────┘
```

---

## Coding Standards

### Go Code Style

Follow [Effective Go](https://go.dev/doc/effective_go) and these conventions:

#### Naming Conventions

```go
// Types: PascalCase
type RouterService struct {}

// Interfaces: PascalCase with -er suffix
type RouterManager interface {}

// Functions/Methods: camelCase for private, PascalCase for exported
func privateFunction() {}
func PublicFunction() {}

// Variables: camelCase for private, PascalCase for exported
var privateVar = "value"
var PublicVar = "value"

// Constants: PascalCase or SCREAMING_SNAKE_CASE
const MaxRetries = 3
const API_VERSION = "4.1"
```

#### Error Handling

```go
// Always check errors
result, err := someFunction()
if err != nil {
    return fmt.Errorf("context: %w", err)  // Wrap errors
}

// Use specific error types when needed
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
```

#### Logging

```go
// Use logrus for structured logging
logger.WithFields(logrus.Fields{
    "user_id": userID,
    "action":  "NAT_UPDATE",
}).Info("NAT rule updated successfully")

// Log levels
logger.Debug("Detailed debug information")
logger.Info("General informational messages")
logger.Warn("Warning messages")
logger.Error("Error messages")
logger.Fatal("Fatal errors (exits app)")
```

#### Comments

```go
// Package comment
// Package nat provides NAT management functionality for MikroTik routers.
package nat

// Function comment
// UpdateNATRule updates the destination address of a NAT rule identified by PPPoE username.
// It returns an error if the rule is not found or the update fails.
func UpdateNATRule(router, username, ip string, port int) error {
    // Implementation
}
```

### Frontend Code Style

#### JavaScript

```javascript
// Use ES6+ syntax
const apiCall = async (endpoint, options = {}) => {
    try {
        const response = await fetch(endpoint, {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${getToken()}`,
                ...options.headers
            }
        });

        if (!response.ok) {
            throw new Error(`API error: ${response.status}`);
        }

        return await response.json();
    } catch (error) {
        console.error('API call failed:', error);
        throw error;
    }
};
```

#### CSS

```css
/* Use BEM naming convention */
.router-card {}
.router-card__header {}
.router-card__body {}
.router-card--disabled {}

/* Use CSS variables for theming */
:root {
    --primary-color: #3498db;
    --danger-color: #e74c3c;
    --spacing-unit: 8px;
}
```

---

## Adding New Features

### Example: Adding New API Endpoint

Let's add a feature to export router statistics.

#### Step 1: Define Model (if needed)

`internal/models/router.go`:
```go
type RouterStatistics struct {
    RouterID         string    `json:"router_id"`
    TotalNATRules    int       `json:"total_nat_rules"`
    ActiveConnections int      `json:"active_connections"`
    Uptime           string    `json:"uptime"`
    CPULoad          int       `json:"cpu_load"`
    MemoryUsed       int       `json:"memory_used"`
    GeneratedAt      time.Time `json:"generated_at"`
}
```

#### Step 2: Implement Service Logic

`internal/services/router_service_db.go`:
```go
// GetRouterStatistics retrieves comprehensive statistics for a router
func (rs *RouterServiceDB) GetRouterStatistics(routerName string) (*models.RouterStatistics, error) {
    // 1. Get router configuration
    router, err := rs.GetRouterByName(routerName)
    if err != nil {
        return nil, fmt.Errorf("failed to get router: %w", err)
    }

    // 2. Connect to router
    client, err := rs.connectToRouter(router)
    if err != nil {
        return nil, fmt.Errorf("failed to connect: %w", err)
    }
    defer client.Close()

    // 3. Get NAT rules count
    natReply, err := client.Run("/ip/firewall/nat/print", "?comment~REMOTE")
    if err != nil {
        return nil, fmt.Errorf("failed to get NAT rules: %w", err)
    }

    // 4. Get system resources
    resourceReply, err := client.Run("/system/resource/print")
    if err != nil {
        return nil, fmt.Errorf("failed to get system resources: %w", err)
    }

    // 5. Parse and construct statistics
    stats := &models.RouterStatistics{
        RouterID:          router.ID,
        TotalNATRules:     len(natReply.Re),
        ActiveConnections: 0, // Parse from connection tracking
        Uptime:            resourceReply.Re[0].Map["uptime"],
        CPULoad:           parseInt(resourceReply.Re[0].Map["cpu-load"]),
        MemoryUsed:        parseInt(resourceReply.Re[0].Map["free-memory"]),
        GeneratedAt:       time.Now(),
    }

    return stats, nil
}
```

#### Step 3: Create Handler

`internal/api/router_handler.go`:
```go
// GetRouterStatistics returns comprehensive statistics for a router
func (h *RouterHandler) GetRouterStatistics(c *gin.Context) {
    routerName := c.Param("name")

    // Get authenticated user
    user := c.MustGet("user").(models.User)

    // Check if user has access to this router
    if !h.routerService.UserHasAccessToRouter(user, routerName) {
        c.JSON(http.StatusForbidden, gin.H{
            "error": "Access denied to this router",
        })
        return
    }

    // Get statistics
    stats, err := h.routerService.GetRouterStatistics(routerName)
    if err != nil {
        h.logger.WithError(err).Error("Failed to get router statistics")
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to retrieve statistics",
        })
        return
    }

    // Log activity
    h.activityLog.Log(user, "ROUTER_STATS_VIEW", "ROUTER", routerName,
        fmt.Sprintf("Viewed statistics for %s", routerName))

    // Return response
    c.JSON(http.StatusOK, gin.H{
        "router": routerName,
        "statistics": stats,
    })
}
```

#### Step 4: Register Route

`cmd/main.go`:
```go
routerGroup := apiGroup.Group("/routers")
{
    // ... existing routes
    routerGroup.GET("/:name/statistics", routerHandler.GetRouterStatistics)
}
```

#### Step 5: Add Frontend (Optional)

`web/static/js/routers.js`:
```javascript
async function getRouterStatistics(routerName) {
    try {
        const response = await fetch(`/api/routers/${routerName}/statistics`, {
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });

        if (!response.ok) {
            throw new Error('Failed to fetch statistics');
        }

        const data = await response.json();
        displayStatistics(data.statistics);
    } catch (error) {
        console.error('Error fetching statistics:', error);
        showError('Failed to load router statistics');
    }
}
```

#### Step 6: Test

```bash
# Test API endpoint
curl -X GET "http://localhost:8080/api/routers/JAKARTA-01/statistics" \
     -H "Authorization: Bearer <token>"

# Expected response:
{
  "router": "JAKARTA-01",
  "statistics": {
    "router_id": "uuid",
    "total_nat_rules": 15,
    "active_connections": 42,
    "uptime": "5d3h15m",
    "cpu_load": 45,
    "memory_used": 60,
    "generated_at": "2025-10-16T10:30:00Z"
  }
}
```

---

## Testing

### Unit Tests

#### Writing Tests

`internal/services/nat_service_test.go`:
```go
package services

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/sirupsen/logrus"
)

func TestNATService_ParseNATRule(t *testing.T) {
    logger := logrus.New()
    service := NewNATService(logger, nil)

    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid rule",
            input:    "dst-address=10.10.10.100:80",
            expected: "10.10.10.100:80",
            wantErr:  false,
        },
        {
            name:     "invalid rule",
            input:    "invalid",
            expected: "",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := service.ParseNATRule(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

#### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -v -run TestNATService_ParseNATRule ./internal/services

# Run tests with race detection
go test -race ./...
```

### Integration Tests

```go
func TestRouterService_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup test database
    db := setupTestDB(t)
    defer db.Close()

    // Create service
    logger := logrus.New()
    service := NewRouterServiceDB(logger, db)

    // Test create router
    router := &models.Router{
        Name:     "TEST-ROUTER",
        Host:     "192.168.1.1",
        Port:     8728,
        Username: "admin",
        Password: "test123",
    }

    err := service.CreateRouter(router)
    assert.NoError(t, err)

    // Test get router
    retrieved, err := service.GetRouterByName("TEST-ROUTER")
    assert.NoError(t, err)
    assert.Equal(t, router.Name, retrieved.Name)

    // Cleanup
    service.DeleteRouter(router.ID)
}
```

### Manual Testing

Use tools like:

- **curl** for API testing
- **Postman** for API collection
- **Browser DevTools** for frontend debugging

---

## Debugging

### Logging

Enable debug mode in `.env`:
```env
DEBUG=true
```

This will:
- Set log level to DEBUG
- Show detailed SQL queries
- Display RouterOS API commands
- Verbose error messages

### VS Code Debugging

`.vscode/launch.json`:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Application",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd/main.go",
            "env": {
                "DEBUG": "true"
            },
            "args": []
        }
    ]
}
```

### Common Debug Scenarios

#### Debug RouterOS API Calls

```go
// Enable verbose logging
logger.SetLevel(logrus.DebugLevel)

// Log RouterOS commands
logger.Debugf("Executing RouterOS command: %s", command)
reply, err := client.Run(command)
logger.Debugf("Reply: %+v", reply)
```

#### Debug Database Queries

```go
// Log SQL queries
logger.Debugf("Executing query: %s with args: %v", query, args)
rows, err := db.Query(query, args...)
```

#### Debug JWT Issues

```go
// Log token validation
logger.WithFields(logrus.Fields{
    "token": token[:20] + "...",  // Log partial token only
    "claims": claims,
}).Debug("Validating JWT token")
```

---

## Database Migrations

### Creating Migration

1. Create new SQL file in `migrations/` directory:

`migrations/002_add_router_health.sql`:
```sql
-- Add router health monitoring table
CREATE TABLE router_health (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    router_id UUID NOT NULL REFERENCES routers(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,
    response_time_ms INTEGER,
    cpu_load INTEGER,
    memory_used INTEGER,
    uptime VARCHAR(50),
    last_check TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_router_health_router ON router_health(router_id);
CREATE INDEX idx_router_health_created ON router_health(created_at DESC);
```

2. Update migration runner (if using custom tool).

### Running Migrations

Currently migrations run automatically on application start.

For manual execution:
```bash
psql -U postgres -d nat_management -f migrations/002_add_router_health.sql
```

---

## Common Tasks

### Adding New User Role

1. Update models:
```go
const (
    RoleAdministrator = "Administrator"
    RoleHeadBranch1   = "Head Branch 1"
    RoleHeadBranch2   = "Head Branch 2"
    RoleHeadBranch3   = "Head Branch 3"
    RoleNewRole       = "New Role Name"  // Add new role
)
```

2. Update authorization logic in middleware
3. Update frontend role selection
4. Update database seed data

### Adding New Activity Log Type

1. Define constant:
```go
const (
    ActionLogin      = "LOGIN"
    ActionNATUpdate  = "NAT_UPDATE"
    ActionNewAction  = "NEW_ACTION"  // Add new action
)
```

2. Use in handlers:
```go
h.activityLog.Log(user, ActionNewAction, "RESOURCE", ...)
```

### Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Live profiling
import _ "net/http/pprof"
# Access http://localhost:8080/debug/pprof/
```

---

## Best Practices

### Security

- Never commit sensitive data (passwords, keys) to git
- Always use parameterized queries to prevent SQL injection
- Validate all user input
- Use HTTPS in production
- Rotate secrets regularly
- Follow principle of least privilege

### Code Quality

- Write tests for critical functionality
- Use meaningful variable and function names
- Keep functions small and focused
- Comment complex logic
- Use consistent error handling
- Review code before committing

### Git Workflow

```bash
# Create feature branch
git checkout -b feature/new-feature

# Make changes and commit
git add .
git commit -m "Add new feature: description"

# Push to remote
git push origin feature/new-feature

# Create pull request for review
```

---

**Version:** 4.1
**Last Updated:** 2025-10-16
**For questions, see:** [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
