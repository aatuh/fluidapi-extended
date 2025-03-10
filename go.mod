module github.com/pakkasys/fluidapi-extended

go 1.23.0

require (
	github.com/go-sql-driver/mysql v1.8.1
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pakkasys/fluidapi v0.6.0
)

require filippo.io/edwards25519 v1.1.0 // indirect

replace github.com/pakkasys/fluidapi => ../fluidapi
