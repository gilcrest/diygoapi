create or replace function create_movie(p_id uuid, p_extl_id character varying, p_title character varying, p_rated character varying, p_released date, p_run_time integer, p_director character varying, p_writer character varying, p_create_client_id uuid, p_create_username character varying)
    returns TABLE(o_create_timestamp timestamp without time zone, o_update_timestamp timestamp without time zone)
    language plpgsql
as
$$
DECLARE
  v_dml_timestamp TIMESTAMP;
  v_create_timestamp timestamp;
  v_update_timestamp timestamp;
BEGIN

  v_dml_timestamp := now() at time zone 'utc';

  INSERT INTO demo.movie (movie_id,
                          extl_id,
                          title,
                          rated,
                          released,
                          run_time,
                          director,
                          writer,
--                           create_client_id,
                          create_username,
                          create_timestamp,
--                           update_client_id,
                          update_username,
                          update_timestamp)
  VALUES (p_id,
          p_extl_id,
          p_title,
          p_rated,
          p_released,
          p_run_time,
          p_director,
          p_writer,
--           p_create_client_id,
          p_create_username,
          v_dml_timestamp,
--           p_create_client_id,
          p_create_username,
          v_dml_timestamp)
      RETURNING create_timestamp, update_timestamp
        into v_create_timestamp, v_update_timestamp;

      o_create_timestamp := v_create_timestamp;
      o_update_timestamp := v_update_timestamp;

      RETURN NEXT;

END;

$$;

