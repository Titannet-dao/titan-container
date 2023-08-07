CREATE TABLE IF NOT EXISTS services(
    id INT UNSIGNED AUTO_INCREMENT,
    name VARCHAR(128) NOT NULL,
    image VARCHAR(128) NOT NULL,
    ports VARCHAR(256),
    expose_port INT DEFAULT 0,
    state INT DEFAULT 0,
    cpu FLOAT        DEFAULT 0,
    memory FLOAT        DEFAULT 0,
    storage FLOAT        DEFAULT 0,
    env VARCHAR(128) DEFAULT NULL,
    arguments VARCHAR(128) DEFAULT NULL,
    deployment_id VARCHAR(128) NOT NULL,
    error_message VARCHAR(128) DEFAULT NULL,
    created_at DATETIME     DEFAULT NULL,
    updated_at DATETIME     DEFAULT NULL,
    PRIMARY KEY (id)
    )ENGINE=InnoDB COMMENT='services';