create table demo.movie
(
    movie_id         uuid                     not null
        constraint movie_pk
            primary key,
    extl_id          varchar                  not null,
    title            varchar(1000)            not null,
    rated            varchar,
    released         date,
    run_time         integer,
    director         varchar(1000),
    writer           varchar(1000),
    create_app_id    uuid                     not null
        constraint movie_create_app_fk
            references demo.app
            deferrable initially deferred,
    create_user_id   uuid
        constraint movie_create_user_fk
            references demo.users
            deferrable initially deferred,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null
        constraint movie_update_app_fk
            references demo.app
            deferrable initially deferred,
    update_user_id   uuid
        constraint movie_update_user_fk
            references demo.users
            deferrable initially deferred,
    update_timestamp timestamp with time zone not null
);

comment on table demo.movie is 'The movie table stores details about a movie.';

comment on column demo.movie.movie_id is 'The unique ID given to the movie.';

comment on column demo.movie.extl_id is 'A unique ID given to the movie which can be used externally.';

comment on column demo.movie.title is 'The title of the movie.';

comment on column demo.movie.rated is 'The movie rating (PG, PG-13, R, etc.)';

comment on column demo.movie.released is 'The date the movie was released.';

comment on column demo.movie.run_time is 'The movie run time in minutes.';

comment on column demo.movie.director is 'The movie director.';

comment on column demo.movie.writer is 'The movie writer.';

comment on column demo.movie.create_app_id is 'The application which created this record.';

comment on column demo.movie.create_user_id is 'The user which created this record.';

comment on column demo.movie.create_timestamp is 'The timestamp when this record was created.';

comment on column demo.movie.update_app_id is 'The application which performed the most recent update to this record.';

comment on column demo.movie.update_user_id is 'The user which performed the most recent update to this record.';

comment on column demo.movie.update_timestamp is 'The timestamp when the record was updated most recently.';

alter table demo.movie
    owner to demo_user;

create unique index movie_extl_id_uindex
    on demo.movie (extl_id);
