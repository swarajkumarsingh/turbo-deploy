CREATE TYPE language_enum AS ENUM ('javascript');
CREATE TYPE source_code_enum AS ENUM ('github');

CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,                         -- Auto-incrementing unique project ID
    user_id INT NOT NULL,                          -- Foreign key referencing the users table
    name VARCHAR(100) NOT NULL,                    -- Project name, required
    source_code_url TEXT,                          -- URL for the source code
    subdomain VARCHAR(100),                        -- Subdomain for the project
    custom_domain VARCHAR(255),                    -- Custom domain for the project
    source_code source_code_enum DEFAULT 'github' NOT NULL, -- ENUM for source code type
    language language_enum DEFAULT 'javascript' NOT NULL,   -- Programming language ENUM
    is_dockerized BOOLEAN DEFAULT FALSE NOT NULL,  -- Dockerized flag, defaults to false
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Timestamp of project creation
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE -- Foreign key to users table
);
