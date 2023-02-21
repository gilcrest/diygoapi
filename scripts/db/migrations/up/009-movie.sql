create table if not exists movie
(
    movie_id         uuid                     not null,
    extl_id          varchar                  not null,
    title            varchar(1000)            not null,
    rated            varchar,
    released         date,
    run_time         integer,
    director         varchar(1000),
    writer           varchar(1000),
    create_app_id    uuid                     not null,
    create_user_id   uuid,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null,
    update_user_id   uuid,
    update_timestamp timestamp with time zone not null
);

comment on table movie is 'The movie table stores details about a movie.';

comment on column movie.movie_id is 'The unique ID given to the movie.';

comment on column movie.extl_id is 'A unique ID given to the movie which can be used externally.';

comment on column movie.title is 'The title of the movie.';

comment on column movie.rated is 'The movie rating (PG, PG-13, R, etc.)';

comment on column movie.released is 'The date the movie was released.';

comment on column movie.run_time is 'The movie run time in minutes.';

comment on column movie.director is 'The movie director.';

comment on column movie.writer is 'The movie writer.';

comment on column movie.create_app_id is 'The application which created this record.';

comment on column movie.create_user_id is 'The user which created this record.';

comment on column movie.create_timestamp is 'The timestamp when this record was created.';

comment on column movie.update_app_id is 'The application which performed the most recent update to this record.';

comment on column movie.update_user_id is 'The user which performed the most recent update to this record.';

comment on column movie.update_timestamp is 'The timestamp when the record was updated most recently.';

create unique index if not exists movie_extl_id_uindex
    on movie (extl_id);

alter table movie
    add constraint movie_pk
        primary key (movie_id);

alter table movie
    add constraint movie_create_app_fk
        foreign key (create_app_id) references app
            deferrable initially deferred;

alter table movie
    add constraint movie_create_user_fk
        foreign key (create_user_id) references users
            deferrable initially deferred;

alter table movie
    add constraint movie_update_app_fk
        foreign key (update_app_id) references app
            deferrable initially deferred;

alter table movie
    add constraint movie_update_user_fk
        foreign key (update_user_id) references users
            deferrable initially deferred;

