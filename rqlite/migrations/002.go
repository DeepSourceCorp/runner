package migrations

const (
	Up002 = `CREATE TABLE session (
		id TEXT PRIMARY KEY,
		app_id TEXT,
		code TEXT,

		backend_token TEXT,
		
		runner_access_token TEXT,
		runner_token_expiry INTEGER,
		runner_refresh_token TEXT
	);
	`
	Down002 = `DROP TABLE session`
)
