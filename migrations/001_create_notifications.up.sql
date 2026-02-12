CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL,
    booking_id      UUID,
    event_type      VARCHAR(50) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    body            TEXT NOT NULL,
    channels_sent   JSONB NOT NULL DEFAULT '[]',
    channels_failed JSONB NOT NULL DEFAULT '[]',
    retry_count     INT NOT NULL DEFAULT 0,
    max_retries     INT NOT NULL DEFAULT 3,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    is_read         BOOLEAN NOT NULL DEFAULT false,
    metadata        JSONB,
    fcm_sent_at     TIMESTAMPTZ,
    sms_sent_at     TIMESTAMPTZ,
    email_sent_at   TIMESTAMPTZ,
    version         BIGINT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_booking_id ON notifications(booking_id);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX idx_notifications_user_unread ON notifications(user_id) WHERE is_read = false;
