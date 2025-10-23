# NAT Management Dashboard

![Version](https://img.shields.io/badge/version-4.2-blue)
![Go](https://img.shields.io/badge/go-1.24.0-00ADD8?logo=go)
![PostgreSQL](https://img.shields.io/badge/postgresql-15+-336791?logo=postgresql)
![License](https://img.shields.io/badge/license-MIT-green)

> **Sistem manajemen NAT (Network Address Translation) yang komprehensif untuk MikroTik RouterOS dengan PostgreSQL backend, authentication JWT, dan UI/UX modern.**

## 📋 Daftar Isi

- [Overview](#-overview)
- [Features](#-features)
- [Technology Stack](#-technology-stack)
- [Prerequisites](#-prerequisites)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Usage](#-usage)
- [API Documentation](#-api-documentation)
- [Tools](#-tools)
- [Documentation](#-documentation)
- [Troubleshooting](#-troubleshooting)
- [Contributing](#-contributing)
- [License](#-license)

---

## 🌟 Overview

NAT Management Dashboard adalah aplikasi web modern untuk mengelola konfigurasi NAT pada router MikroTik. Aplikasi ini dirancang untuk mempermudah administrator jaringan dalam mengelola multiple routers, monitoring clients, dan melakukan troubleshooting dengan tools yang komprehensif.

### Keunggulan Utama

- ✅ **Multi-Router Management** - Kelola banyak router dari satu dashboard
- ✅ **Role-Based Access Control** - 4 level user (Admin + 3 Head Branch)
- ✅ **Real-time Monitoring** - Online clients & PPPoE status
- ✅ **Advanced Connection Handling** - Retry logic dengan exponential backoff
- ✅ **Comprehensive Diagnostic Tools** - Built-in troubleshooting utilities
- ✅ **Modern UI/UX** - Responsive design dengan mobile optimization
- ✅ **PostgreSQL Backend** - Reliable & scalable database (Neon Serverless)
- ✅ **JWT Authentication** - Secure API access dengan token management
- ✅ **Activity Logging** - Comprehensive audit trail

---

## ✨ Features

### Core Features

#### 1. **NAT Configuration Management**
- Update NAT rules via RouterOS API
- Search NAT configurations by comment
- Real-time synchronization dengan router
- Support multiple tunnel endpoints
- Public ONT URL management

#### 2. **Online Clients Monitoring**
- Real-time NAT clients list
- Filter by IP address, port, protocol
- Connection status tracking
- Bandwidth monitoring ready

#### 3. **PPPoE Status Checker**
- Search PPPoE active sessions by username
- Fuzzy search support
- Multiple router search
- Connection details (IP, uptime, interface)

#### 4. **Router Management**
- Add/Edit/Delete router configurations
- Connection testing dengan detailed report
- Router health monitoring
- Configuration validation
- Support API & API-SSL

#### 5. **User Management**
- Role-based access control (RBAC)
- Per-router access permissions
- User activity tracking
- Password management
- Account activation/deactivation

#### 6. **Activity Logs**
- Comprehensive audit trail
- Filter by user, action, router
- Export logs to CSV
- Retention policy management

### Security Features

- 🔐 JWT-based authentication
- 🔒 Secure password hashing (bcrypt)
- 🛡️ CORS protection dengan whitelist
- 🚫 Rate limiting on sensitive endpoints
- 📝 Security event logging
- 🔑 Token refresh mechanism
- ⏱️ Session timeout management

### Advanced Features (v4.2)

#### 🔄 1. Smart Auto-Refresh System
- **Intelligent refresh intervals**: 90s normal, 30s fast, 180s slow mode
- **Pause on user interaction**: Auto-pauses during user activity
- **Resume after inactivity**: Resumes after 5s of inactivity
- **Visual refresh indicator**: Shows countdown and status
- **Configurable intervals**: Customizable per-user preferences

#### 🔍 2. Advanced Search & Filters
- **Real-time search**: Instant filtering with 300ms debounce
- **Multi-field search**: Search across username, IP, caller ID
- **Router filtering**: Filter by specific routers
- **Filter persistence**: Saves filter state across sessions
- **Export filtered results**: Export only filtered data

#### 📊 3. Data Export Functionality
- **Multiple formats**: Excel (.xlsx), CSV, PDF, JSON
- **Formatted exports**: Professional styling with headers
- **Custom filename**: Auto-timestamped filenames
- **Bulk export**: Export all or filtered data
- **PDF with tables**: jsPDF with auto-table plugin

#### ⌨️ 4. Keyboard Shortcuts
- **Global shortcuts**: Ctrl+K (search), Ctrl+R (refresh), Ctrl+E (export)
- **Modal navigation**: ESC to close modals
- **Help dialog**: Press ? to see all shortcuts
- **Accessibility**: Full keyboard navigation support
- **Customizable bindings**: Extend with custom shortcuts

#### ⏳ 5. Skeleton Loading States
- **Visual placeholders**: Prevent layout shift during loading
- **Smooth transitions**: Fade-in animations for loaded content
- **Per-component skeletons**: Cards, tables, lists
- **Improved UX**: Professional loading experience
- **Reduced perceived wait time**: Makes app feel faster

#### 📱 6. Mobile UX Enhancements
- **Pull-to-refresh**: Native mobile refresh gesture
- **Swipe gestures**: Open/close sidebar with swipe
- **Floating Action Button (FAB)**: Quick access menu
- **Touch optimizations**: 44px minimum touch targets
- **Haptic feedback**: Visual feedback for interactions
- **Safe area support**: Notched device compatibility

#### ✅ 7. Quick Actions Toolbar
- **Bulk operations**: Multi-select with checkboxes
- **Bulk disconnect**: Disconnect multiple clients at once
- **Bulk export**: Export selected items
- **Bulk NAT target**: Set single client as NAT target
- **Selection management**: Select all/none, max 100 items
- **Smooth animations**: Slide-up toolbar with bounce effect

#### 🔧 Connection & Reliability
- 🔄 Connection retry logic (3 attempts, exponential backoff)
- ⚡ Enhanced timeout handling (15-45 seconds)
- 🔍 Router diagnostic tool
- 🧙‍♂️ Interactive setup wizard
- 📊 Detailed error reporting
- 📖 Comprehensive documentation

---

## 🛠 Technology Stack

### Backend
- **Language**: Go 1.24.0
- **Web Framework**: Gin (v1.9.1)
- **Database**: PostgreSQL 15+ (Neon Serverless)
- **Database Driver**: pgx/v5, lib/pq
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Password Hashing**: bcrypt (golang.org/x/crypto)
- **RouterOS Client**: go-routeros
- **Logging**: Logrus

### Frontend
- **HTML5** + **CSS3** (Custom styling)
- **JavaScript** (Vanilla ES6+)
- **Chart.js** (untuk visualisasi - ready)
- **Font Awesome** (icons)
- **Responsive Design** (Mobile-first approach)

### Infrastructure
- **Database Hosting**: Neon (Serverless PostgreSQL)
- **Environment Config**: godotenv
- **Session Management**: UUID-based sessions
- **Rate Limiting**: golang.org/x/time

### Development Tools
- **Router Diagnostic Tool** (Go CLI)
- **Router Setup Wizard** (Interactive CLI)
- **Build Scripts** (Batch scripts untuk Windows)

---

## 📦 Prerequisites

### System Requirements

- **OS**: Windows 10/11, Linux, macOS
- **Go**: Version 1.24.0 or higher
- **PostgreSQL**: Version 15+ (or Neon Serverless account)
- **MikroTik Router**: RouterOS v6.0+ dengan API service enabled
- **Network**: Koneksi ke router via API port (8728 atau 8729)

### Required Software

```bash
# Check Go installation
go version  # Should be >= 1.24.0

# Check PostgreSQL (if using local)
psql --version  # Should be >= 15.0
```

### MikroTik Requirements

- RouterOS API service enabled (`/ip service enable api`)
- User dengan API permission
- Firewall rules allow API access dari server

---

## 🚀 Installation

### 1. Clone Repository

```bash
git clone <repository-url>
cd nat-management-app
```

### 2. Install Dependencies

```bash
go mod download
go mod verify
```

### 3. Setup Database

#### Option A: Using Neon Serverless (Recommended)

1. Create account di [neon.tech](https://neon.tech)
2. Create new project
3. Copy connection string
4. Create `.env` file (lihat Configuration section)

#### Option B: Using Local PostgreSQL

```sql
-- Create database
CREATE DATABASE nat_management;

-- Connect to database
\c nat_management

-- Run migration (automatic on first start)
-- Or manually:
psql -U postgres -d nat_management -f migrations/init.sql
```

### 4. Configure Environment

Create `.env` file di root directory:

```env
# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080
DEBUG=false

# Database Configuration (Neon Serverless)
DATABASE_URL=postgresql://user:password@ep-xxx.region.aws.neon.tech/nat_management?sslmode=require

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-min-32-chars
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h

# Session Configuration
SESSION_SECRET=your-session-secret-key
SESSION_MAX_AGE=86400

# CORS Configuration
ALLOWED_ORIGINS=http://localhost:8080,http://127.0.0.1:8080

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_DURATION=60s
```

### 5. Build Application

```bash
# Build main application
go build -o nat-supabase.exe ./cmd

# Build diagnostic tools (optional)
cd tools
go build -o router-diagnostic.exe router-diagnostic.go
go build -o router-setup-wizard.exe router-setup-wizard.go
cd ..

# Or use batch script (Windows)
build-tools.bat
```

### 6. Run Application

```bash
# Run directly
./nat-supabase.exe

# Or with Go
go run ./cmd/main.go
```

### 7. Access Application

Open browser: `http://localhost:8080`

**Default Login:**
- Username: `admin`
- Password: `admin123`

⚠️ **IMPORTANT**: Change default password setelah first login!

---

## ⚙️ Configuration

### Router Configuration

Add router via Web UI or using Setup Wizard:

#### Via Web UI:

1. Login as Administrator
2. Go to **Router Management** page
3. Click **Add Router**
4. Fill form:
   - **Router Name**: Unique identifier (e.g., JAKARTA-01)
   - **Host**: IP address atau hostname
   - **Port**: `8728` (API) atau `8729` (API-SSL)
   - **Username**: MikroTik user dengan API permission
   - **Password**: Router password
   - **Tunnel Endpoint**: Internal IP:port (e.g., 172.22.28.5:80)
   - **Public ONT URL**: Public URL (e.g., http://tunnel3.ebilling.id:19701)
5. Click **Test Connection**
6. If success, click **Save Router**

#### Via Setup Wizard:

```bash
cd tools
router-setup-wizard.exe
# Follow interactive prompts
```

### User Configuration

#### Create New User:

1. Login as Administrator
2. Go to **User Management** page
3. Click **Add User**
4. Fill form:
   - Username
   - Password (min 6 chars)
   - Full Name
   - Email
   - Role (Administrator / Head Branch 1/2/3)
5. Select accessible routers
6. Click **Create User**

### Environment Variables Reference

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SERVER_HOST` | Server bind address | `localhost` | No |
| `SERVER_PORT` | Server port | `8080` | No |
| `DEBUG` | Debug mode | `false` | No |
| `DATABASE_URL` | PostgreSQL connection string | - | **Yes** |
| `JWT_SECRET` | JWT signing key | - | **Yes** |
| `JWT_EXPIRY` | JWT token expiry | `24h` | No |
| `JWT_REFRESH_EXPIRY` | Refresh token expiry | `168h` | No |
| `SESSION_SECRET` | Session encryption key | - | **Yes** |
| `SESSION_MAX_AGE` | Session max age (seconds) | `86400` | No |
| `ALLOWED_ORIGINS` | CORS allowed origins | `*` | No |
| `RATE_LIMIT_REQUESTS` | Rate limit requests | `100` | No |
| `RATE_LIMIT_DURATION` | Rate limit window | `60s` | No |

---

## 📖 Usage

### Quick Start Guide

1. **Login to Application**
   ```
   URL: http://localhost:8080
   Username: admin
   Password: admin123
   ```

2. **Add Your First Router**
   - Go to Router Management
   - Click "Add Router"
   - Enter router details (use port 8728!)
   - Test connection
   - Save

3. **Update NAT Rule**
   - Go to NAT Management
   - Select router
   - Search PPPoE username
   - Update destination IP/port
   - Save changes

4. **Check PPPoE Status**
   - Go to PPPoE Checker
   - Select router(s)
   - Enter username (partial match supported)
   - Click "Check Status"

5. **View Activity Logs**
   - Go to Activity Logs (Admin only)
   - Filter by date, user, action
   - Export to CSV if needed

### Common Operations

#### Update NAT for ONT Remote Access

```
1. Login to application
2. NAT Management → Select Router
3. Search by PPPoE username
4. Update fields:
   - Destination: <new-ont-ip>:80
   - Comment: REMOTE ONT PELANGGAN
5. Click "Update NAT Rule"
6. Verify di MikroTik Winbox
```

#### Troubleshoot Router Connection

```bash
# Using diagnostic tool
cd tools
router-diagnostic.exe <host> 8728 <username> <password>

# Example:
router-diagnostic.exe 192.168.1.1 8728 admin password123
```

#### Manage Multi-Router Setup

```
1. Login as admin
2. Router Management → Add all routers
3. User Management → Create branch users
4. Assign routers per user:
   - head1 → JAKARTA, BANDUNG
   - head2 → SURABAYA, MEDAN
   - head3 → BALI, MAKASSAR
5. Each user can only access their assigned routers
```

---

## 🔌 API Documentation

### Authentication

All API endpoints require JWT authentication.

#### Login

```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}

Response:
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-10-17T10:30:00Z",
  "user": {
    "id": "uuid",
    "username": "admin",
    "role": "Administrator"
  }
}
```

#### Refresh Token

```http
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### API Request with JWT

```http
GET /api/routers
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Router Endpoints

```http
GET    /api/routers              # List all routers
POST   /api/routers              # Create router
GET    /api/routers/:id          # Get router details
PUT    /api/routers/:id          # Update router
DELETE /api/routers/:id          # Delete router
POST   /api/routers/:id/test     # Test connection
GET    /api/routers/stats        # Get statistics
```

### NAT Endpoints

```http
GET    /api/nat/configs          # Get NAT configs
GET    /api/nat/clients          # Get online clients
POST   /api/nat/update           # Update NAT rule
GET    /api/nat/status           # Get NAT status
```

### PPPoE Endpoints

```http
POST   /api/pppoe/check          # Check PPPoE status
GET    /api/pppoe/routers        # Get available routers
POST   /api/pppoe/fuzzy-search   # Fuzzy search PPPoE
```

### User Endpoints

```http
GET    /api/users                # List users (Admin only)
POST   /api/users                # Create user (Admin only)
GET    /api/users/:id            # Get user details
PUT    /api/users/:id            # Update user
DELETE /api/users/:id            # Delete user (Admin only)
```

### Activity Log Endpoints

```http
GET    /api/logs                 # Get logs (Admin only)
GET    /api/logs/:id             # Get log details
GET    /api/logs/stats           # Get log statistics
POST   /api/logs/cleanup         # Delete old logs
```

For detailed API documentation, see: [docs/API-REFERENCE.md](docs/API-REFERENCE.md)

---

## 🧰 Tools

### 1. Router Diagnostic Tool

Comprehensive diagnostic tool untuk troubleshooting router connection.

**Usage:**
```bash
cd tools
router-diagnostic.exe <host> <port> <username> <password>

# Example:
router-diagnostic.exe 192.168.1.1 8728 admin password123
```

**Tests Performed:**
- ✅ DNS Resolution
- ✅ TCP Connection (multiple timeouts: 5s, 15s, 30s)
- ✅ RouterOS API Authentication
- ✅ Router Identity Retrieval
- ✅ System Resources Info
- ✅ Detailed error reporting with suggestions

**Output:**
```
🔍 ========================================
🔍 Router Connection Diagnostic Tool
🔍 ========================================
🎯 Target: 192.168.1.1:8728
👤 Username: admin
🔍 ========================================

📋 Running diagnostic tests...

🔍 Test 1: DNS Resolution for 192.168.1.1
   ✅ Host is already an IP address (0.00s)

🔍 Test 2: TCP Connection (timeout: 5s)
   ✅ TCP connection successful (0.15s)

🔍 Test 3: RouterOS API Connection
   ✅ RouterOS API authentication successful (0.20s)

...
```

### 2. Router Setup Wizard

Interactive CLI wizard untuk setup router dengan guided prompts.

**Usage:**
```bash
cd tools
router-setup-wizard.exe
# Follow interactive prompts
```

**Features:**
- Step-by-step configuration
- Input validation
- Connection testing sebelum save
- Configuration summary
- Next steps guidance

### 3. Build Tools Script

Batch script untuk compile semua tools.

**Usage:**
```bash
build-tools.bat
```

---

## 📚 Documentation

### Available Documentation

1. **[TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)**
   - Common errors & solutions
   - Connection timeout troubleshooting
   - Port configuration guide
   - Firewall setup
   - Detailed checklist

2. **[ROUTER-SETUP.md](docs/ROUTER-SETUP.md)**
   - Complete router setup guide
   - Prerequisites checklist
   - Security best practices
   - Multi-router setup scenarios
   - Production checklist

3. **[PROJECT-OVERVIEW.md](docs/PROJECT-OVERVIEW.md)**
   - Detailed project structure
   - Database schema
   - Architecture diagrams
   - Component descriptions

4. **[API-REFERENCE.md](docs/API-REFERENCE.md)**
   - Complete API documentation
   - Request/response examples
   - Authentication guide
   - Error codes

5. **[DEVELOPMENT-GUIDE.md](docs/DEVELOPMENT-GUIDE.md)**
   - Development setup
   - Code structure
   - Adding features
   - Testing guide

6. **[DEPLOYMENT.md](docs/DEPLOYMENT.md)**
   - Production deployment
   - Security checklist
   - Performance optimization
   - Monitoring setup

---

## 🔧 Troubleshooting

### Common Issues

#### 1. Connection Timeout to Router

**Error:**
```
dial tcp <ip>:<port>: connectex: A connection attempt failed...
```

**Solutions:**
1. Verify port is **8728** (bukan 19699/19701!)
2. Enable API service: `/ip service enable api`
3. Check firewall rules
4. Run diagnostic tool

See: [TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)

#### 2. Authentication Failed

**Error:**
```
invalid user name or password
```

**Solutions:**
1. Verify username/password
2. Check user has API permission
3. Try login via Winbox/SSH untuk verify credentials

#### 3. Database Connection Error

**Error:**
```
Failed to connect to database
```

**Solutions:**
1. Verify DATABASE_URL in `.env`
2. Check Neon project is running
3. Verify SSL mode is `require` for Neon
4. Check internet connection

#### 4. Port Already in Use

**Error:**
```
bind: address already in use
```

**Solutions:**
```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <pid> /F

# Linux/Mac
lsof -ti:8080 | xargs kill -9
```

---

## 🎯 Project Structure

```
nat-management-app/
├── cmd/
│   └── main.go                 # Application entry point
├── config/
│   └── config.go               # Configuration management
├── internal/
│   ├── api/                    # API handlers
│   │   ├── auth_handler.go
│   │   ├── nat_handler.go
│   │   ├── router_handler.go
│   │   ├── user_handler.go
│   │   └── activity_log_handler.go
│   ├── database/               # Database layer
│   │   └── db.go
│   ├── middleware/             # HTTP middleware
│   │   └── auth_middleware.go
│   ├── models/                 # Data models
│   │   ├── user.go
│   │   ├── router.go
│   │   ├── nat.go
│   │   └── activity_log.go
│   ├── services/               # Business logic
│   │   ├── auth_service.go
│   │   ├── nat_service.go
│   │   ├── router_service_db.go
│   │   ├── user_service.go
│   │   └── activity_log_service.go
│   └── utils/                  # Utility functions
│       └── jwt.go
├── web/
│   ├── static/                 # Static files (CSS, JS, images)
│   └── templates/              # HTML templates
│       ├── base.html
│       ├── login.html
│       ├── nat_management.html
│       ├── pppoe_checker.html
│       ├── router_management.html
│       ├── user_management.html
│       └── activity_logs.html
├── tools/                      # Diagnostic & setup tools
│   ├── router-diagnostic.go
│   └── router-setup-wizard.go
├── docs/                       # Documentation
│   ├── TROUBLESHOOTING.md
│   ├── ROUTER-SETUP.md
│   ├── CONNECTION-FIX-SUMMARY.md
│   ├── PROJECT-OVERVIEW.md
│   ├── API-REFERENCE.md
│   ├── DEVELOPMENT-GUIDE.md
│   ├── DEPLOYMENT.md
│   └── CHANGELOG.md
├── migrations/                 # Database migrations
│   └── init.sql
├── .env.example                # Environment template
├── .gitignore
├── go.mod
├── go.sum
├── build-tools.bat            # Build script
└── README.md                   # This file
```

---

## 🚀 Future Enhancements

### Planned Features (Roadmap)

#### Phase 1: Router Health Monitoring ⏳
- Background health monitoring service
- In-memory cache layer (60s TTL)
- Router health dashboard with visual cards
- REST API for health data
- 90% reduction in TCP connections
- Status: **Ready to implement**

#### Phase 2: Advanced Monitoring 📊
- Connection pooling
- WebSocket real-time updates
- Historical health data
- Performance graphs
- Alert system for router down events

#### Phase 3: Enhanced Features 🎯
- Bandwidth monitoring & graphs
- Traffic analysis
- Automatic failover configuration
- Backup/restore router configs
- Bulk operations support

#### Phase 4: Enterprise Features 🏢
- Multi-tenancy support
- API rate limiting per user
- Advanced reporting & analytics
- Integration dengan monitoring tools (Prometheus, Grafana)
- Webhook notifications

---

## 👥 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

### Development Guidelines

- Follow Go best practices
- Add tests for new features
- Update documentation
- Use meaningful commit messages
- Check code with `go vet` dan `golint`

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 👤 Author

**NAT Management Team**

- Version: 4.2
- Last Updated: 2025-10-17

---

## 🙏 Acknowledgments

- [MikroTik](https://mikrotik.com) - RouterOS platform
- [Gin](https://gin-gonic.com) - Web framework
- [Neon](https://neon.tech) - Serverless PostgreSQL
- [go-routeros](https://github.com/go-routeros/routeros) - RouterOS API client
- Community contributors

---

## 📞 Support

For issues, questions, or feature requests:

1. Check [TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)
2. Run diagnostic tool dan collect output
3. Check existing GitHub issues
4. Create new issue dengan detail lengkap

---

## ⚡ Quick Commands Reference

```bash
# Development
go run ./cmd/main.go                    # Run application
go build -o nat-supabase.exe ./cmd      # Build binary
go test ./...                           # Run tests
go vet ./...                            # Static analysis

# Tools
cd tools
router-diagnostic.exe <host> <port> <user> <pass>
router-setup-wizard.exe

# Database (if local)
psql -U postgres -d nat_management      # Connect to DB
psql -U postgres -d nat_management -f migrations/init.sql  # Run migrations

# Production
./nat-supabase.exe                      # Run production binary
nohup ./nat-supabase.exe &              # Run in background (Linux)
```

---

**Made with ❤️ for Network Administrators**

---

*For detailed guides, please refer to the [docs/](docs/) directory.*
