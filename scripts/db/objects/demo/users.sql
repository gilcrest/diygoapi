create table if not exists users
(
    user_id          uuid                     not null,
    user_extl_id     varchar                  not null,
    person_id        uuid                     not null,
    name_prefix      varchar,
    first_name       varchar                  not null,
    middle_name      varchar,
    last_name        varchar                  not null,
    name_suffix      varchar,
    nickname         varchar,
    email            varchar,
    company_name     varchar,
    company_dept     varchar,
    job_title        varchar,
    birth_date       date,
    birth_year       bigint,
    birth_month      bigint,
    birth_day        bigint,
    create_app_id    uuid                     not null,
    create_user_id   uuid,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null,
    update_user_id   uuid,
    update_timestamp timestamp with time zone not null
);

comment on table users is 'users stores data about users that interact with the system. A user is a person who utilizes a computer or network service." In the context of this project, given that we allow Persons to authenticate with multiple providers, a User is akin to a user (Wikipedia - "The word persona derives from Latin, where it originally referred to a theatrical mask. On the social web, users develop virtual personas as online identities.") and as such, a Person can have one to many Users (for instance, I can have a GitHub user and a Google user, but I am just one Person). As a general, practical matter, most operations are considered at the User level. For instance, roles are assigned at the user level instead of the Person level, which allows for more fine-grained access control. Architecture note: All tables are to be singular, however, because user is a reserved word, the rules are broken here. It is unfortunate, but the alternatives are no better.';

comment on column users.email is 'Primary email for the user';

alter table users
    add constraint users_pk
        primary key (user_id);

alter table users
    add constraint users_extl_id_ui
        unique (user_extl_id);

alter table users
    add constraint users_create_app_fk
        foreign key (create_app_id) references app
            deferrable initially deferred;

alter table users
    add constraint users_create_self_ref_fk
        foreign key (create_user_id) references users
            deferrable initially deferred;

alter table users
    add constraint users_update_app_fk
        foreign key (update_app_id) references app
            deferrable initially deferred;

alter table users
    add constraint users_update_self_ref_fk
        foreign key (update_user_id) references users
            deferrable initially deferred;

alter table users
    add constraint user_person_fk
        foreign key (person_id) references person
            deferrable initially deferred;

