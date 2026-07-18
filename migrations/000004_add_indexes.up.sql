CREATE INDEX IF NOT EXISTS idx_messages_private_id ON messages(private_id);
CREATE INDEX IF NOT EXISTS idx_messages_from_id ON messages(from_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_privates_user1_id ON privates(user1_id);
CREATE INDEX IF NOT EXISTS idx_privates_user2_id ON privates(user2_id);
