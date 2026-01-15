-- 测试索引命名规范
CREATE TABLE test_naming (
  id NUMBER(10) PRIMARY KEY,
  email VARCHAR2(100),
  username VARCHAR2(50),
  CONSTRAINT pk_test PRIMARY KEY (id),
  CONSTRAINT uk_email UNIQUE (email)
);

CREATE INDEX idx_username ON test_naming(username);
CREATE UNIQUE INDEX uidx_email ON test_naming(email);
