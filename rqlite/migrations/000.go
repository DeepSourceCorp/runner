package migrations

const (
	Up000   = `CREATE TABLE IF NOT EXISTS migrations (name VARCHAR(255) PRIMARY KEY, time DATETIME) WITHOUT ROWID;`
	Down000 = `DROP TABLE migrations;`
)
