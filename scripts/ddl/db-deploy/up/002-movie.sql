create table demo.movie
(
    movie_id         uuid          not null,
    extl_id          varchar(250)  not null,
    title            varchar(1000) not null,
    rated            varchar(10),
    released         date,
    run_time         integer,
    director         varchar(1000),
    writer           varchar(1000),
    create_username  varchar,
    create_timestamp timestamp with time zone,
    update_username  varchar,
    update_timestamp timestamp with time zone,
    constraint movie_pk
        primary key (movie_id)
);

alter table demo.movie
    owner to demo_user;

create unique index movie_extl_id_uindex
    on demo.movie (extl_id);

