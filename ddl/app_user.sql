create table demo.app_user
(
	username varchar(100) not null,
	mobile_id varchar(100),
	email_address varchar(254) not null,
	first_name varchar(100) not null,
	last_name varchar(100) not null,
	create_user_id varchar(100) not null,
	create_date timestamp not null,
	update_user_id varchar(100) not null,
	update_date timestamp not null
);

create function demo.create_app_user(p_username character varying, p_mobile_id character varying, p_email_address character varying, p_first_name character varying, p_last_name character varying, p_create_user_id character varying)
  returns timestamp with time zone
LANGUAGE plpgsql
AS $$
DECLARE
  v_create_date   TIMESTAMP;
BEGIN
  INSERT INTO demo.app_user (username, mobile_id, email_address, first_name, last_name, create_user_id, create_date, update_user_id, update_date)
	  VALUES (p_username, p_mobile_id, p_email_address, p_first_name, p_last_name, p_create_user_id, current_timestamp, p_create_user_id, current_timestamp)
  RETURNING create_date into v_create_date;
  return v_create_date;
END;

$$;
