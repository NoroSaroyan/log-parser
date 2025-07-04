package config

import (
	"fmt"
	_ "github.com/lib/pq"
)

// DSN builds a Postgres connection string from DatabaseConfig
func (db *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode,
	)
}
