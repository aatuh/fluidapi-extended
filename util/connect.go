package util

import (
	"time"

	"github.com/pakkasys/fluidapi/database"
)

// NewDefaultTCPConfig returns a connection Config with default
// settings for TCP MySQL connections.
func NewDefaultTCPConfig(
	user string,
	password string,
	db string,
	driver string,
) *database.ConnectConfig {
	return &database.ConnectConfig{
		User:            user,
		Password:        password,
		Database:        db,
		ConnectionType:  database.TCP,
		Host:            "localhost",
		Port:            3306,
		ConnMaxLifetime: 10 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		Driver:          driver,
		DSNFormat:       "", // If empty, DSN() uses default building for MySQL
	}
}

// NewDefaultUnixConfig returns a connection Config with default
// settings for Unix socket MySQL connections.
func NewDefaultUnixConfig(
	user string,
	password string,
	db string,
	socketDirectory string,
	socketName string,
	driver string,
) *database.ConnectConfig {
	return &database.ConnectConfig{
		User:            user,
		Password:        password,
		Database:        db,
		ConnectionType:  database.Unix,
		SocketDirectory: socketDirectory,
		SocketName:      socketName,
		ConnMaxLifetime: 10 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		Driver:          driver,
		DSNFormat:       "", // If empty, DSN() uses default building
	}
}

// NewDefaultSQLiteConfig returns a connection Config with default
// settings for SQLite. The "db" argument can be a file path or
// ":memory:" if you prefer an in-memory database.
func NewDefaultSQLiteConfig(db string) *database.ConnectConfig {
	return &database.ConnectConfig{
		Database:        db,
		Driver:          database.SQLite3,
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 0,
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		// DSNFormat empty -> DSN() will do "database?parameters"
	}
}
