-- Create scenes table
CREATE TABLE IF NOT EXISTS scenes (
    id CHAR(36) PRIMARY KEY,
    terminal_id CHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    actions TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    FOREIGN KEY (terminal_id) REFERENCES terminal(id) ON DELETE CASCADE
);

-- Create index on terminal_id
CREATE INDEX idx_scenes_terminal_id ON scenes(terminal_id);

-- Create index on deleted_at for soft delete queries
CREATE INDEX idx_scenes_deleted_at ON scenes(deleted_at);
