#!/bin/bash

# Database setup script for Mafia Wars: Criminal Empire

# Configuration variables
DB_USER="postgres"
DB_PASSWORD="postgres"
DB_NAME="mafia_wars"
MIGRATIONS_DIR="./migrations"

# Check for psql command
if ! command -v psql &> /dev/null; then
    echo "Error: PostgreSQL client (psql) is not installed or not in your PATH."
    exit 1
fi

# Create database if it doesn't exist
echo "Creating database $DB_NAME if it doesn't exist..."
PGPASSWORD=$DB_PASSWORD psql -U $DB_USER -c "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || PGPASSWORD=$DB_PASSWORD psql -U $DB_USER -c "CREATE DATABASE $DB_NAME"

# Apply migrations
echo "Applying database migrations from $MIGRATIONS_DIR..."
for migration in $(find "$MIGRATIONS_DIR" -name "*.sql" | sort); do
    echo "Applying migration: $migration"
    PGPASSWORD=$DB_PASSWORD psql -U $DB_USER -d $DB_NAME -f $migration
done

# Check if we have a territory.yaml file to seed territory data
TERRITORY_FILE="./configs/territory.yaml"
if [ -f "$TERRITORY_FILE" ]; then
    echo "Seeding territory data from $TERRITORY_FILE..."
    # In a real scenario, we would have a Go command to seed territory data
    # For now, we'll just display a message
    echo "NOTE: To seed territory data, run the following command:"
    echo "go run cmd/seed/main.go --config=./configs/app.yaml --territory=./configs/territory.yaml"
fi

echo "Database setup complete."