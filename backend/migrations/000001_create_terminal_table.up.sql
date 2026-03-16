-- Create terminal table
CREATE TABLE IF NOT EXISTS terminal (
    id CHAR(36) PRIMARY KEY,
    mac_address VARCHAR(255) UNIQUE NOT NULL,
    room_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL
);

-- Create index on mac_address
CREATE INDEX idx_terminal_mac_address ON terminal(mac_address);

-- Create index on deleted_at for soft delete queries
CREATE INDEX idx_terminal_deleted_at ON terminal(deleted_at);
