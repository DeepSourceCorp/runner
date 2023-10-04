package migrations

const (
	Up002 = `CREATE TABLE oauth_session (
		id TEXT PRIMARY KEY,
		backend_access_token TEXT,
		backend_access_token_expires_at INTEGER,
		backend_refresh_token TEXT,
		code TEXT,
		runner_access_token TEXT,
		runner_access_token_expires_at INTEGER,
		runner_refresh_token TEXT
	);
	`
	Down002 = `DROP TABLE session`
)
