#!/bin/bash
set -e

DB_PATH="${1:-testdata/ottomat.db}"
ADMIN_PASSWORD="${2:-happy.cat.happy.nap}"

echo "Initializing new database at ${DB_PATH}"

# Remove existing database files
echo "Removing old database files..."
rm -f "${DB_PATH}" "${DB_PATH}-shm" "${DB_PATH}-wal"

# Create database
echo "Creating database..."
./dist/local/ottomat db init --db "${DB_PATH}"

# Run migrations
echo "Running migrations..."
./dist/local/ottomat db migrate --db "${DB_PATH}"

# Seed with admin user
echo "Seeding admin user..."
./dist/local/ottomat db seed --db "${DB_PATH}" --password "${ADMIN_PASSWORD}"

echo ""
echo "Database initialized successfully!"
echo "Database: ${DB_PATH}"
echo "Admin username: admin"
echo "Admin password: ${ADMIN_PASSWORD}"
