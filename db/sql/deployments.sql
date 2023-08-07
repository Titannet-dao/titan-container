CREATE TABLE IF NOT EXISTS deployments(
    id VARCHAR(128) NOT NULL UNIQUE,
    owner VARCHAR(128) NOT NULL,
    name VARCHAR(128) NOT NULL DEFAULT '',
    state INT DEFAULT 0,
    type INT DEFAULT 0,
    authority TINYINT(1) DEFAULT 0,
    version VARCHAR(128) DEFAULT '',
    balance FLOAT        DEFAULT 0,
    cost FLOAT        DEFAULT 0,
    provider_id VARCHAR(128) NOT NULL,
    expiration DATETIME     DEFAULT NULL,
    created_at DATETIME     DEFAULT NULL,
    updated_at DATETIME     DEFAULT NULL
    )ENGINE=InnoDB COMMENT='deployments';