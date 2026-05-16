-- Replace boolean is_read with nullable timestamp read_at so the API can
-- expose WHEN a notification was read (matching lib-proto/dto NotificationDTO.ReadAt).
ALTER TABLE notifications ADD COLUMN read_at TIMESTAMPTZ NULL;

-- Preserve existing read state: previously-read rows get their updated_at as a
-- best-effort approximation of when they were marked read.
UPDATE notifications SET read_at = updated_at WHERE is_read = true;

-- Drop the old partial index before dropping the column it references.
DROP INDEX IF EXISTS idx_notifications_user_unread;

ALTER TABLE notifications DROP COLUMN is_read;

-- Recreate the unread partial index against the new column.
-- Widened from (user_id) to (user_id, created_at DESC) to support the
-- paginated unread inbox added by Phase 5 (Task 5.2).
CREATE INDEX idx_notifications_user_unread
    ON notifications (user_id, created_at DESC)
    WHERE read_at IS NULL;
