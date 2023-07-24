package migrations

const (
	Up002 = `CREATE TABLE IF NOT EXISTS saml_sessions (
		id VARCHAR(27) PRIMARY KEY,
		access_code VARCHAR(255),
		access_code_expiry DATETIME,
		backend_email VARCHAR(255),
		backend_first_name VARCHAR(255),
		backend_last_name VARCHAR(255),
		backend_raw VARCHAR(255),
		backend_expiry DATETIME,
		runner_access_token VARCHAR(255),
		runner_refresh_token VARCHAR(255),
		runner_token_expiry DATETIME
		) WITHOUT ROWID;
		`
	Down002 = `DROP TABLE saml_sessions`
)
