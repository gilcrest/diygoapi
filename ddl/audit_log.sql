drop table api.audit_log;

-- auto-generated definition
CREATE TABLE api.audit_log
(
  request_id     VARCHAR(100) NOT NULL
    CONSTRAINT audit_log_pkey
    PRIMARY KEY,
  protocol       VARCHAR(20)  NOT NULL,
  protocol_major INTEGER,
  protocol_minor INTEGER,
  request_method VARCHAR(10)  NOT NULL,
  scheme         VARCHAR(100),
  host           VARCHAR(100) NOT NULL,
  port           VARCHAR(100) NOT NULL,
  path           VARCHAR(4000),
  header         jsonb,
  content_length BIGINT,
  remote_address VARCHAR(100)
);
