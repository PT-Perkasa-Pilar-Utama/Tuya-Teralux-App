-- Create scenes table
CREATE TABLE IF NOT EXISTS scenes (
    id CHAR(36) PRIMARY KEY,
    teralux_id CHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    actions TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    FOREIGN KEY (teralux_id) REFERENCES teralux(id) ON DELETE CASCADE
);

-- Create index on teralux_id
CREATE INDEX idx_scenes_teralux_id ON scenes(teralux_id);

-- Create index on deleted_at for soft delete queries
CREATE INDEX idx_scenes_deleted_at ON scenes(deleted_at);
