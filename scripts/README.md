# Backend Scripts

This directory contains utility scripts for database seeding, geocoding, and key generation.

## Why Build Tags?

Each script has `//go:build ignore` at the top to prevent compilation conflicts. This allows multiple standalone scripts with `func main()` to coexist in the same directory.

## Available Scripts

### 1. Generate Encryption Key

Generates a secure AES-256 encryption key for encrypting sensitive data (IBAN, CNP, etc.).

```bash
go run scripts/generate_encryption_key.go
```

**Output**: A base64-encoded key to add to your `.env` file.

**Important**:
- Never commit encryption keys to version control
- Store in GCP Secret Manager in production
- If lost, encrypted data cannot be recovered

---

### 2. Seed Database

Seeds the development database with test data (users, cleaners, bookings, etc.).

```bash
# Using Makefile
make db-seed

# Or directly
go run scripts/seed.go

# Or with custom database URL
DATABASE_URL="postgresql://..." go run scripts/seed.go
```

**Creates**:
- 10 test clients
- 10 test cleaners (with various tiers and ratings)
- 5 test bookings (various statuses)
- Sample addresses, reviews, and disputes

**Default DB URL**: `postgresql://cleanbuddy:devpassword@localhost:5432/cleanbuddy?sslmode=disable`

---

### 3. Backfill Geocoding

Backfills latitude/longitude coordinates for existing addresses and cleaners using OpenStreetMap Nominatim API.

```bash
# With Google Maps API (recommended)
GOOGLE_MAPS_API_KEY="your-key" go run scripts/backfill_geocoding.go

# Or with Nominatim (free, rate-limited)
go run scripts/backfill_geocoding.go
```

**Features**:
- Geocodes addresses without coordinates
- Geocodes cleaner locations without coordinates
- Rate limiting (1 request/second for Nominatim)
- Caches results in database
- Resume-safe (skips already geocoded records)

**APIs**:
- Google Maps Geocoding API (requires API key, 0.005$/request)
- Nominatim (free, 1 req/sec limit, no API key)

---

## Troubleshooting

### "main redeclared" Error

If you see this error, it means the build tags are missing or incorrect. Each script should start with:

```go
//go:build ignore
// +build ignore

package main
```

### Database Connection Failed

Ensure PostgreSQL is running and credentials are correct:

```bash
# Check if database is accessible
psql postgresql://cleanbuddy:devpassword@localhost:5432/cleanbuddy

# Or use Docker
docker-compose up -d postgres
```

### Geocoding Rate Limit

Nominatim has strict rate limits (1 req/sec). For large datasets, use Google Maps API or run the script slowly.

---

## Development Workflow

1. **Setup database**: `docker-compose up -d postgres`
2. **Run migrations**: `make db-migrate-up`
3. **Seed test data**: `make db-seed`
4. **Generate encryption key**: `go run scripts/generate_encryption_key.go`
5. **Backfill geocoding** (optional): `go run scripts/backfill_geocoding.go`

---

## Notes

- Scripts use `//go:build ignore` to prevent inclusion in main build
- Each script is standalone and can be run independently
- Scripts connect directly to the database (no API)
- Ideal for development, testing, and data migration
