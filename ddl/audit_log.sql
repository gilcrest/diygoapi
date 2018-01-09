drop table api.audit_log;

-- auto-generated definition
CREATE TABLE api.audit_log
(
  request_id         VARCHAR(100) NOT NULL
    CONSTRAINT audit_log_pkey
    PRIMARY KEY,
  request_timestamp  TIMESTAMP,
  response_code      INTEGER,
  response_timestamp TIMESTAMP,
  duration_in_millis BIGINT,
  protocol           VARCHAR(20)  NOT NULL,
  protocol_major     INTEGER,
  protocol_minor     INTEGER,
  request_method     VARCHAR(10)  NOT NULL,
  scheme             VARCHAR(100),
  host               VARCHAR(100) NOT NULL,
  port               VARCHAR(100) NOT NULL,
  path               VARCHAR(4000),
  remote_address     VARCHAR(100),
  request_header     JSONB,
  request_content_length     BIGINT,
  request_body       TEXT,
  response_header    JSONB,
  response_body      TEXT

);

create or REPLACE function api.log_request(p_request_id character varying,
                                       p_request_timestamp TIMESTAMP,
                                       p_response_code integer,
                                       p_response_timestamp TIMESTAMP,
                                       p_duration_in_millis BIGINT,
                                       p_protocol character varying, 
                                       p_protocol_major integer, 
                                       p_protocol_minor integer, 
                                       p_request_method character varying, 
                                       p_scheme character varying, 
                                       p_host character varying, 
                                       p_port character varying, 
                                       p_path character varying,
                                       p_remote_address character varying,
                                       p_request_header jsonb, 
                                       p_request_content_length bigint,
                                       p_request_body TEXT,
                                       p_response_header jsonb,
                                       p_response_body TEXT 
                                       ) returns integer
LANGUAGE plpgsql
AS $$
DECLARE
  v_rows_inserted INTEGER;
BEGIN
 INSERT INTO api.audit_log (request_id, 
                            request_timestamp, 
                            response_code,
                            response_timestamp,
                            duration_in_millis,
                            protocol, 
                            protocol_major, 
                            protocol_minor, 
                            request_method, 
                            scheme, 
                            host, 
                            port, 
                            path, 
                            remote_address,
                            request_header,
                            request_content_length,
                            request_body,
                            response_header,
                            response_body
                            )
	  VALUES (p_request_id, 
            p_request_timestamp,
            p_response_code,
            p_response_timestamp,
            p_duration_in_millis,
            p_protocol, 
            p_protocol_major, 
            p_protocol_minor, 
            p_request_method, 
            p_scheme, 
            p_host, 
            p_port, 
            p_path,
            p_remote_address,
            p_request_header,
            p_request_content_length,
            p_request_body, 
            p_response_header,
            p_response_body
            );
  GET DIAGNOSTICS v_rows_inserted = ROW_COUNT;
  return v_rows_inserted;
END;
$$;
