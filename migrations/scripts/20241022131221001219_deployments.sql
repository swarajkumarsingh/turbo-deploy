CREATE TYPE status_enum AS ENUM ('QUEUE', 'PROG', 'READY', 'FAIL');

CREATE TABLE IF NOT EXISTS deployments (
    id SERIAL PRIMARY KEY,                          -- Auto-incrementing unique deployment ID
    user_id INT NOT NULL,                           -- Foreign key referencing the users table
    project_id INT NOT NULL,                        -- Foreign key referencing the projects table
    duration INT,                                   -- Duration of the deployment (in seconds, minutes, etc.)
    ready_url VARCHAR,                              -- URL for the ready deployment
    last_log VARCHAR,                               -- Last log for the deployment
    status status_enum DEFAULT 'QUEUE' NOT NULL,    -- Status of the deployment, defaults to 'QUEUE'
    created_at TIMESTAMP DEFAULT NOW(),             -- Timestamp of deployment creation, defaults to current time
    updated_at TIMESTAMP DEFAULT NOW(),             -- Timestamp of the last update
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,   -- Foreign key to users table
    CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE -- Foreign key to projects table
);
