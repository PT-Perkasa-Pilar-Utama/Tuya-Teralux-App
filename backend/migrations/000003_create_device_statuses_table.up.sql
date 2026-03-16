-- Create device_statuses table
CREATE TABLE IF NOT EXISTS device_statuses (
    device_id CHAR(36) NOT NULL,
    code VARCHAR(255) NOT NULL,
    value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    -- Composite primary key
    PRIMARY KEY (device_id, code),
    
    -- Foreign key constraint to devices table
    CONSTRAINT fk_device
        FOREIGN KEY (device_id)
        REFERENCES devices(id)
        ON DELETE CASCADE
);

-- Create index on device_id
CREATE INDEX idx_device_statuses_device_id ON device_statuses(device_id);

-- Create index on code
CREATE INDEX idx_device_statuses_code ON device_statuses(code);

-- Create index on deleted_at for soft delete queries
CREATE INDEX idx_device_statuses_deleted_at ON device_statuses(deleted_at);
