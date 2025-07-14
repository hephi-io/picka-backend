package db

import (
	"database/sql"
	"log"
)

func RunMigrations(db *sql.DB) {
	_, err := db.Exec(`
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";

	CREATE TABLE IF NOT EXISTS shipments (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		vendor_id UUID NOT NULL,
		product_id UUID NOT NULL,
		quantity INT NOT NULL,
		destination TEXT NOT NULL,
		notes TEXT,
		status TEXT DEFAULT 'pending',
		pickup_time TIMESTAMP,
		created_at TIMESTAMP DEFAULT now()
	);
	`)
	if err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("✅ Database migrations completed successfully.")
}
