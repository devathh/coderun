CREATE TABLE IF NOT EXISTS user_services (
    timestamp DateTime64(3) DEFAULT now64(),
    user_id String,
    language String,

    INDEX idx_user_id user_id TYPE bloom_filter GRANULARITY 1
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, language)
TTL timestamp + INTERVAL 30 DAY;