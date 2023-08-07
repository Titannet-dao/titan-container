CREATE TABLE IF NOT EXISTS properties(
    id INT UNSIGNED AUTO_INCREMENT,
    provider_id VARCHAR(128) NOT NULL UNIQUE,
    app_id VARCHAR(128) NOT NULL,
    app_type INT DEFAULT 0,
    created_at DATETIME     DEFAULT NULL,
    updated_at DATETIME     DEFAULT NULL,
    PRIMARY KEY (id),
    KEY idx_provider_id (provider_id)
)ENGINE=InnoDB COMMENT='properties';