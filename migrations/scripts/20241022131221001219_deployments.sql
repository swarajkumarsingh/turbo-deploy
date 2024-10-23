CREATE TYPE status_enum AS ENUM ('QUEUE', 'PROG', 'READY', 'FAIL');

CREATE TABLE IF NOT EXISTS deployments(
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    project_id INT NOT NULL,
    duration INT,
    ready_url VARCHAR,
    last_log VARCHAR,
    status status_enum DEFAULT 'QUEUE' NOT NULL,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NOW()
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);