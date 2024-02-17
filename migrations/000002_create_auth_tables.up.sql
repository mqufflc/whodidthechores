CREATE TABLE IF NOT EXISTS users (
	id CHAR(10) PRIMARY KEY CHECK (id != '') NOT NULL,
   	name TEXT NOT NULL UNIQUE CHECK (name != ''),
    hash VARCHAR(255) NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	modified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_sessions (
	id VARCHAR(255) PRIMARY KEY CHECK (id != '') NOT NULL,
    user_id CHAR(10) REFERENCES users,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);