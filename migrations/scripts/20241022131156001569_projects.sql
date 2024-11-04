CREATE TYPE language_enum AS ENUM ('javascript');
CREATE TYPE source_code_enum AS ENUM ('github');

CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    source_code_url VARCHAR(255),
    subdomain VARCHAR(100) UNIQUE,
    custom_domain VARCHAR(255),
    source_code source_code_enum DEFAULT 'github' NOT NULL,
    language language_enum DEFAULT 'javascript' NOT NULL,
    is_dockerized BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
