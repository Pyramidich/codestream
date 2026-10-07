CREATE TABLE IF NOT EXISTS change_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID REFERENCES files(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_change_history_file_id ON change_history(file_id);
CREATE INDEX IF NOT EXISTS idx_change_history_project_id ON change_history(project_id);
CREATE INDEX IF NOT EXISTS idx_change_history_user_id ON change_history(user_id);
CREATE INDEX IF NOT EXISTS idx_change_history_created_at ON change_history(created_at);
