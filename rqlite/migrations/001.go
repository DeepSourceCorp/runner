package migrations

const (
	Up001   = `CREATE TABLE IF NOT EXISTS code (code TEXT PRIMARY KEY, user TEXT) WITHOUT ROWID;`
	Down001 = `DROP TABLE code`
)
