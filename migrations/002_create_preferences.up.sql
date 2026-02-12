CREATE TABLE notification_preferences (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id           UUID NOT NULL UNIQUE,
    enable_push       BOOLEAN NOT NULL DEFAULT true,
    enable_sms        BOOLEAN NOT NULL DEFAULT true,
    enable_email      BOOLEAN NOT NULL DEFAULT true,
    fcm_token         VARCHAR(255),
    phone_number      VARCHAR(20),
    email             VARCHAR(255),
    quiet_hours_start TIME,
    quiet_hours_end   TIME,
    version           BIGINT NOT NULL DEFAULT 1,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notification_preferences_user_id ON notification_preferences(user_id);
