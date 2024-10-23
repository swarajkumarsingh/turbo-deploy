CREATE TABLE IF NOT EXISTS log_events(
    id SERIAL PRIMARY KEY,
    deployment_id INT NOT NULL,
    project_id INT NOT NULL,
    data VARCHAR,
    metadata VARCHAR,
    data_length INT,
	created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_deployment FOREIGN KEY (deployment_id) REFERENCES deployments(id) ON DELETE CASCADE
);