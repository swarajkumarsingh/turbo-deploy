CREATE TYPE log_type_enum AS ENUM ('INFO', 'WARN', 'ERROR');

CREATE TABLE IF NOT EXISTS deployment_logs (
    id SERIAL PRIMARY KEY,
    deployment_id INT NOT NULL,
    project_id INT NOT NULL,
    environment VARCHAR,
    message VARCHAR,
    cause VARCHAR,
    stack VARCHAR,
    name VARCHAR,
    host VARCHAR,
    log_type log_type_enum DEFAULT 'INFO' NOT NULL,
    timestamp VARCHAR,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_deployment FOREIGN KEY (deployment_id) REFERENCES deployments(id) ON DELETE CASCADE
);
