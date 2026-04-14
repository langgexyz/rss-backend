CREATE TABLE IF NOT EXISTS feeds (
    id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    url              VARCHAR(500)  NOT NULL UNIQUE,
    title            VARCHAR(512)  NOT NULL DEFAULT '',
    description      TEXT,
    site_url         VARCHAR(2048),
    fetch_status     ENUM('pending','fetching','success','failed') NOT NULL DEFAULT 'pending',
    fetch_error      VARCHAR(1024),
    last_fetched_at  DATETIME,
    source_updated_at DATETIME,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS articles (
    id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    feed_id          BIGINT UNSIGNED NOT NULL,
    guid_hash        CHAR(64)      NOT NULL,
    title            VARCHAR(1024) NOT NULL DEFAULT '',
    link             VARCHAR(2048),
    content          MEDIUMTEXT,
    author           VARCHAR(512),
    published_at     DATETIME,
    is_read          TINYINT(1)    NOT NULL DEFAULT 0,
    is_starred       TINYINT(1)    NOT NULL DEFAULT 0,
    is_full_content  TINYINT(1)    NOT NULL DEFAULT 0,
    created_at       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_feed_guid_hash (feed_id, guid_hash),
    INDEX idx_feed_id (feed_id),
    INDEX idx_published_at (published_at),
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
