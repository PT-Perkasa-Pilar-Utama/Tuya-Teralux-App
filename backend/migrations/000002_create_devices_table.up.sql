-- Create devices table
CREATE TABLE IF NOT EXISTS devices (
    id CHAR(36) PRIMARY KEY,
    terminal_id CHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    -- Tuya Fields (matching Tuya API response, except 'status' and 'online')
    remote_id VARCHAR(255),
    category VARCHAR(255),
    remote_category VARCHAR(255),
    product_name VARCHAR(255),
    remote_product_name VARCHAR(255),
    icon VARCHAR(255),
    custom_name VARCHAR(255),
    model VARCHAR(255),
    ip VARCHAR(255),
    local_key VARCHAR(255),
    gateway_id VARCHAR(255),
    create_time BIGINT,
    update_time BIGINT,
    collections TEXT,
    
    -- Foreign key constraint with CASCADE delete
    CONSTRAINT fk_terminal
        FOREIGN KEY (terminal_id)
        REFERENCES terminal(id)
        ON DELETE CASCADE
);

-- Create index on terminal_id
CREATE INDEX idx_devices_terminal_id ON devices(terminal_id);

-- Create index on remote_id for IR device lookup
CREATE INDEX idx_devices_remote_id ON devices(remote_id);

-- Create index on deleted_at for soft delete queries
CREATE INDEX idx_devices_deleted_at ON devices(deleted_at);
