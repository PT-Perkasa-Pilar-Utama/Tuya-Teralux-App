ALTER TABLE terminal ADD COLUMN ai_engine_profile VARCHAR(50);
CREATE INDEX idx_terminal_ai_engine_profile ON terminal(ai_engine_profile);
