-- Reverse of 003_add_notification_read_at.up.sql.
-- LOSSY: the exact read_at timestamp is collapsed back to a boolean and cannot
-- be recovered if up is re-applied later. Down is for emergency rollback only.
ALTER TABLE notifications ADD COLUMN is_read BOOLEAN NOT NULL DEFAULT false;

-- Preserve read state: any row that had a read_at timestamp is considered read.
UPDATE notifications SET is_read = true WHERE read_at IS NOT NULL;

DROP INDEX IF EXISTS idx_notifications_user_unread;

ALTER TABLE notifications DROP COLUMN read_at;

CREATE INDEX idx_notifications_user_unread
    ON notifications (user_id)
    WHERE is_read = false;
