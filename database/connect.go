package database

import (
	"time"

	"github.com/pakkasys/fluidapi/database"
)

// NewDefaultMySQLTCPConfig returns a connection Config with default
// settings for TCP MySQL connections.
//
// Parameters:
//   - user: The MySQL username.
//   - password: The MySQL password.
//   - db: The MySQL database name.
//
// Returns:
//   - A ConnectConfig with default settings for MySQL TCP connections.
func NewDefaultMySQLTCPConfig(
	user string, password string, db string,
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
		Driver:          database.MySQL,
		DSNFormat:       "", // If empty, DSN() uses default string.
	}
}

// NewDefaultMySQLUnixConfig returns a connection Config with default
// settings for Unix socket MySQL connections.
//
// Parameters:
//   - user: The MySQL username.
//   - password: The MySQL password.
//   - db: The MySQL database name.
//   - socketDirectory: The directory where the socket file is located.
//   - socketName: The name of the socket file.
//
// Returns:
//   - A ConnectConfig with default settings for MySQL Unix socket connections.
func NewDefaultMySQLUnixConfig(
	user string,
	password string,
	db string,
	socketDirectory string,
	socketName string,
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
		Driver:          database.MySQL,
		DSNFormat:       "", // If empty, DSN() uses default string.
	}
}

// NewDefaultSQLiteConfig returns a connection Config with default
// settings for SQLite. The "db" argument can be a file path or
// ":memory:" if you prefer an in-memory database.
//
// Parameters:
//   - db: The SQLite database name.
//
// Returns:
//   - A ConnectConfig with default settings for SQLite.
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
