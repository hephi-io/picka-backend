package db

import (
	"database/sql"
	"log"
)

func RunMigrations(db *sql.DB) {
	_, err := db.Exec(`
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";

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
	log.Println("✅ Rider DB migrations completed successfully.")
}
