CREATE TABLE IF NOT EXISTS keys (
    id SERIAL PRIMARY KEY,
    key_value VARCHAR(255) UNIQUE NOT NULL,
    group_name VARCHAR(100) NOT NULL,
    pattern VARCHAR(100) NOT NULL,
    status BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_key_value UNIQUE (key_value)
);

CREATE INDEX idx_keys_group ON keys(group_name);
CREATE INDEX idx_keys_status ON keys(status);
CREATE INDEX idx_keys_created ON keys(created_at);
