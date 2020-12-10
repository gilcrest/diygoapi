-- Execute this first, then switch connection to newly created
-- go_api_basic database and run remainder of script
create database go_api_basic
    with owner postgres;

create schema demo;

create table demo.movie
(
    movie_id uuid not null
        constraint movie_pk
            primary key,
    extl_id varchar(250) not null,
    title varchar(1000) not null,
    rated varchar(10),
    released date,
    run_time integer,
    director varchar(1000),
    writer varchar(1000),
    create_username varchar,
    create_timestamp timestamp with time zone,
    update_username varchar,
    update_timestamp timestamp with time zone
);

alter table demo.movie owner to postgres;

create unique index movie_extl_id_uindex
    on demo.movie (extl_id);

create function demo.create_movie(p_id uuid, p_extl_id character varying, p_title character varying, p_rated character varying, p_released date, p_run_time integer, p_director character varying, p_writer character varying, p_create_client_id uuid, p_create_username character varying)
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

alter function demo.create_movie(uuid, varchar, varchar, varchar, date, integer, varchar, varchar, uuid, varchar) owner to postgres;
