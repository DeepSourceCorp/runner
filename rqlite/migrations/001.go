package migrations

const (
	Up001 = `CREATE TABLE IF NOT EXISTS oauth_sessions (
		id VARCHAR(27) PRIMARY KEY,
		access_code VARCHAR(255),
		access_code_expiry DATETIME,
		backend_access_token VARCHAR(255),
		backend_refresh_token VARCHAR(255),
		backend_token_expiry DATETIME,
		runner_access_token VARCHAR(255),
		runner_refresh_token VARCHAR(255),
		runner_token_expiry DATETIME
		) WITHOUT ROWID;
		`
	Down001 = `DROP TABLE oauth_sessions`
)
