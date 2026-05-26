ALTER TABLE notification_preferences ADD COLUMN IF NOT EXISTS scope_type VARCHAR(20) DEFAULT 'user';
ALTER TABLE notification_preferences ADD COLUMN IF NOT EXISTS scope_id UUID;

CREATE INDEX IF NOT EXISTS idx_notification_preferences_scope
    ON notification_preferences(scope_type, scope_id);
