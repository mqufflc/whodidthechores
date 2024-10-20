CREATE TABLE IF NOT EXISTS chores (
	id SERIAL PRIMARY KEY,
   	name TEXT NOT NULL UNIQUE CHECK (name != ''),
    description TEXT NOT NULL,
	default_duration_mn INT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
   	name TEXT NOT NULL UNIQUE CHECK (name != '')
);

CREATE TABLE IF NOT EXISTS tasks (
	id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
	user_id INT REFERENCES users (id) ON DELETE RESTRICT NOT NULL,
	chore_id INT REFERENCES chores (id) ON DELETE RESTRICT NOT NULL,
	started_at TIMESTAMPTZ NOT NULL,
	duration_mn INT NOT NULL,
	description TEXT NOT NULL
);