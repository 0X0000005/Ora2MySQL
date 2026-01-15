CREATE TABLE test_naming (
  id BIGINT NOT NULL,
  email VARCHAR(100),
  username VARCHAR(50),
  PRIMARY KEY (id),
  CONSTRAINT pk_test_pk PRIMARY KEY (id),
  CONSTRAINT uk_email_uk UNIQUE (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE INDEX idx_username_idx ON test_naming (username);

CREATE UNIQUE INDEX uidx_email_uk ON test_naming (email);

