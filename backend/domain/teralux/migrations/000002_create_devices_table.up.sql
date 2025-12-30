-- Create devices table
CREATE TABLE IF NOT EXISTS devices (
    id CHAR(36) PRIMARY KEY,
    teralux_id CHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    -- Foreign key constraint with CASCADE delete
    CONSTRAINT fk_teralux
        FOREIGN KEY (teralux_id)
        REFERENCES teralux(id)
        ON DELETE CASCADE
);

-- Create index on teralux_id
CREATE INDEX idx_devices_teralux_id ON devices(teralux_id);

-- Create index on deleted_at for soft delete queries
CREATE INDEX idx_devices_deleted_at ON devices(deleted_at);
