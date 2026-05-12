CREATE TABLE notification_dismissals (
    log_id TEXT NOT NULL,
    dismissed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (log_id)
);

CREATE INDEX idx_notification_dismissals_dismissed_at ON notification_dismissals(dismissed_at);