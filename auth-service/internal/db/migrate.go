package db

import (
	"database/sql"
	"log"
)

func RunMigrations(db *sql.DB) {
	_, err := db.Exec(`
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";

	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email TEXT UNIQUE,
		phone TEXT UNIQUE,
		password TEXT NOT NULL,
		role TEXT NOT NULL CHECK (role IN ('vendor', 'rider', 'admin')),
		first_name TEXT,
		last_name TEXT,
		business_name TEXT,
		is_active BOOLEAN DEFAULT TRUE,
		last_login TIMESTAMP,
		created_at TIMESTAMP DEFAULT NOW(),
		password_reset_token TEXT,
		reset_token_expires TIMESTAMP,
		email_verified BOOLEAN DEFAULT FALSE
	);

	CREATE TABLE IF NOT EXISTS vendor_profiles (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL,
		business_name TEXT NOT NULL,
		business_address TEXT,
		logo_url TEXT,
		created_at TIMESTAMP DEFAULT NOW()
	);
	CREATE UNIQUE INDEX IF NOT EXISTS uniq_vendor_user ON vendor_profiles(user_id);

	CREATE TABLE IF NOT EXISTS rider_profiles (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL,
		phone TEXT,
		vehicle_type TEXT,
		status TEXT DEFAULT 'offline',
		location GEOGRAPHY(Point, 4326),
		created_at TIMESTAMP DEFAULT NOW()
	);
	CREATE UNIQUE INDEX IF NOT EXISTS uniq_rider_user ON rider_profiles(user_id);
	`)
	if err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("✅ Database migrations completed successfully.")
}
