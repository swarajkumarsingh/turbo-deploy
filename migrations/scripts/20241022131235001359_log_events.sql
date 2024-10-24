CREATE TABLE IF NOT EXISTS log_events (
    id SERIAL PRIMARY KEY,                              -- Auto-incrementing unique log event ID
    deployment_id INT NOT NULL,                         -- Foreign key referencing the deployments table
    project_id INT NOT NULL,                            -- Project ID related to the log event
    data VARCHAR,                                       -- Log event data
    metadata VARCHAR,                                   -- Metadata associated with the log event
    data_length INT,                                    -- Length of the log data
    created_at TIMESTAMP DEFAULT NOW(),                 -- Timestamp of log event creation
    CONSTRAINT fk_deployment FOREIGN KEY (deployment_id) REFERENCES deployments(id) ON DELETE CASCADE -- Foreign key to deployments table
);
