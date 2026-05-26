DROP INDEX IF EXISTS idx_notification_preferences_scope;
ALTER TABLE notification_preferences DROP COLUMN IF EXISTS scope_id;
ALTER TABLE notification_preferences DROP COLUMN IF EXISTS scope_type;
