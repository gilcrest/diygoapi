create table demo.person_profile
(
    person_profile_id uuid                     not null,
    person_id         uuid                     not null,
    name_prefix       varchar,
    first_name        varchar                  not null,
    middle_name       varchar,
    last_name         varchar                  not null,
    name_suffix       varchar,
    nickname          varchar,
    company_name      varchar,
    company_dept      varchar,
    job_title         varchar,
    birth_date        date,
    birth_year        bigint,
    birth_month       bigint,
    birth_day         bigint,
    language_id       uuid,
    create_app_id     uuid,
    create_user_id    uuid,
    create_timestamp  timestamp with time zone not null,
    update_app_id     uuid,
    update_user_id    uuid,
    update_timestamp  timestamp with time zone not null,
    constraint person_profile_pk
        primary key (person_profile_id),
    constraint person_profile_person_fk
        foreign key (person_id) references demo.person
            deferrable initially deferred,
    constraint person_profile_create_app_fk
        foreign key (create_app_id) references demo.app
            deferrable initially deferred,
    constraint person_profile_create_user_fk
        foreign key (create_user_id) references demo.app_user
            deferrable initially deferred,
    constraint person_profile_update_app_fk
        foreign key (update_app_id) references demo.app
            deferrable initially deferred,
    constraint person_profile_update_user_fk
        foreign key (update_user_id) references demo.app_user
            deferrable initially deferred
);

alter table demo.person_profile
    owner to postgres;

