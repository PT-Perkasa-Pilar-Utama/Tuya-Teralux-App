-- Create devices table
CREATE TABLE IF NOT EXISTS devices (
    id CHAR(36) PRIMARY KEY,
    teralux_id CHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    online BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    -- Tuya Fields
    tuya_id VARCHAR(255),
    remote_id VARCHAR(255),
    category VARCHAR(255),
    remote_category VARCHAR(255),
    product_name VARCHAR(255),
    remote_product_name VARCHAR(255),
    local_key VARCHAR(255),
    gateway_id VARCHAR(255),
    ip VARCHAR(255),
    model VARCHAR(255),
    icon VARCHAR(255),
    
    -- Foreign key constraint with CASCADE delete
    CONSTRAINT fk_teralux
        FOREIGN KEY (teralux_id)
        REFERENCES teralux(id)
        ON DELETE CASCADE
);

-- Create index on teralux_id
CREATE INDEX idx_devices_teralux_id ON devices(teralux_id);

-- Create index on tuya_id for device lookup
CREATE INDEX idx_devices_tuya_id ON devices(tuya_id);

-- Create index on remote_id for IR device lookup
CREATE INDEX idx_devices_remote_id ON devices(remote_id);

-- Create index on deleted_at for soft delete queries
CREATE INDEX idx_devices_deleted_at ON devices(deleted_at);
