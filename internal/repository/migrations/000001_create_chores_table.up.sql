CREATE TABLE IF NOT EXISTS chores (
	id CHAR(10) PRIMARY KEY CHECK (id != '') NOT NULL,
   	name TEXT NOT NULL UNIQUE CHECK (name != ''),
    description TEXT,
	created_at timestamp with time zone NOT NULL DEFAULT now(),
	modified_at timestamp with time zone NOT NULL DEFAULT now()
);