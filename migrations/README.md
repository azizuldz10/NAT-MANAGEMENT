# 🚀 Migrasi ke Neon PostgreSQL - Setup Guide

## 📋 Yang Sudah Dibuat

✅ **Database Schema** (`migrations/001_init_schema.sql`)
- Table: `users` - User accounts dengan authentication
- Table: `routers` - Router configurations
- Table: `router_access_control` - Role-based access control
- Table: `pppoe_search_history` - History pencarian PPPoE
- Table: `audit_logs` - Audit trail untuk security

✅ **Seed Data** (`migrations/002_seed_data.sql`)
- Default users (admin, head1, head2, head3)
- Role-based access control rules

## 🎯 Langkah Setup Neon PostgreSQL

### 1. **Buat Database di Neon.tech**

1. Buka https://neon.tech
2. Sign up / Login
3. Klik "Create Project"
4. Pilih region terdekat (Singapore/Tokyo untuk Indonesia)
5. Copy **Connection String** yang diberikan

Format connection string:
```
postgresql://username:password@ep-xxx-xxx.region.aws.neon.tech/neondb?sslmode=require
```

### 2. **Jalankan Migration SQL**

**Opsi A: Via Neon Console**
1. Buka project di Neon Dashboard
2. Klik "SQL Editor"
3. Copy-paste isi file `migrations/001_init_schema.sql`
4. Klik "Run"
5. Copy-paste isi file `migrations/002_seed_data.sql`
6. Klik "Run"

**Opsi B: Via psql command line**
```bash
# Set connection string
export DATABASE_URL="postgresql://username:password@ep-xxx.neon.tech/neondb?sslmode=require"

# Run migrations
psql $DATABASE_URL -f migrations/001_init_schema.sql
psql $DATABASE_URL -f migrations/002_seed_data.sql
```

### 3. **Update .env File**

Buat file `.env` di root project:
```env
# Neon PostgreSQL Connection
DATABASE_URL=postgresql://username:password@ep-xxx-xxx.region.aws.neon.tech/neondb?sslmode=require

# Alternative format (akan di-parse oleh aplikasi)
DB_HOST=ep-xxx-xxx.region.aws.neon.tech
DB_PORT=5432
DB_USER=username
DB_PASSWORD=password
DB_NAME=neondb
DB_SSLMODE=require

# JWT Keys (existing)
JWT_PRIVATE_KEY_PATH=private.key
JWT_PUBLIC_KEY_PATH=public.key

# Server Config (existing)
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
```

### 4. **Install Go Dependencies**

```bash
go get github.com/lib/pq
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool
go get github.com/joho/godotenv
```

## 📊 Database Schema Overview

### Users Table
```sql
- id (SERIAL PRIMARY KEY)
- username (VARCHAR, UNIQUE)
- password (VARCHAR, bcrypt hashed)
- full_name (VARCHAR)
- email (VARCHAR, UNIQUE)
- role (VARCHAR) -- Administrator, Head Branch 1, 2, 3
- is_active (BOOLEAN)
- created_at, updated_at, last_login_at
```

### Routers Table
```sql
- id (VARCHAR PRIMARY KEY) -- Format: "routername-uuid"
- name (VARCHAR, UNIQUE)
- host, port, username, password
- tunnel_endpoint, public_ont_url
- enabled (BOOLEAN)
- description (TEXT)
- created_at, updated_at
```

### Router Access Control Table
```sql
- id (SERIAL PRIMARY KEY)
- role (VARCHAR)
- router_name (VARCHAR) -- Can be "*" for wildcard
- permissions (VARCHAR[]) -- Array: read, write, delete, manage
- description (TEXT)
```

## 🔐 Default Login Credentials

**⚠️ GANTI PASSWORD SETELAH SETUP!**

| Username | Password | Role | Access |
|----------|----------|------|--------|
| admin | admin123 | Administrator | All routers |
| head1 | head123 | Head Branch 1 | SAMSAT, LANE1 |
| head2 | head123 | Head Branch 2 | LANE2, LANE4 |
| head3 | head123 | Head Branch 3 | BT JAYA/PK JAYA, SUKAWANGI |

## 🧪 Testing Database Connection

Setelah migration, test koneksi dengan query ini di Neon SQL Editor:

```sql
-- Check users
SELECT username, role, is_active FROM users;

-- Check router access control
SELECT role, router_name, permissions FROM router_access_control;

-- Check routers (should be empty initially)
SELECT id, name, enabled FROM routers;
```

## 📝 Next Steps

Setelah database ready:
1. ✅ Migration files created
2. ✅ Run migrations di Neon
3. 🔄 Implement database connection layer (in progress)
4. 🔄 Update RouterService to use PostgreSQL
5. 🔄 Update AuthService to use PostgreSQL
6. 🔄 Test full flow: Add router → DB → PPPoE/NAT sync

## 🎉 Keuntungan Neon PostgreSQL

✅ **Serverless** - Auto-scale, bayar sesuai usage
✅ **Fast** - Connection pooling built-in
✅ **Branching** - Database branches untuk testing
✅ **Free Tier** - 0.5GB storage gratis
✅ **Auto-backup** - Point-in-time recovery
✅ **Global** - Low latency dari berbagai region

## 📞 Support

Jika ada error saat migration:
1. Check connection string format
2. Pastikan SSL mode = require
3. Check Neon dashboard untuk error logs
4. Verify user permissions di Neon
