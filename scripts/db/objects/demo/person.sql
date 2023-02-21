create table if not exists person
(
    person_id        uuid                     not null,
    person_extl_id   varchar                  not null,
    create_app_id    uuid                     not null,
    create_user_id   uuid,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null,
    update_user_id   uuid,
    update_timestamp timestamp with time zone not null
);

comment on column person.person_id is 'The user ID is the unique ID for user (pk for table)';

comment on column person.person_extl_id is 'The unique user external ID to be given to outside callers.';

comment on column person.create_app_id is 'The application which created this record.';

comment on column person.create_user_id is 'The user which created this record.';

comment on column person.create_timestamp is 'The timestamp when this record was created.';

comment on column person.update_app_id is 'The application which performed the most recent update to this record.';

comment on column person.update_user_id is 'The user which performed the most recent update to this record.';

comment on column person.update_timestamp is 'The timestamp when the record was updated most recently.';

create unique index if not exists person_extl_id_uindex
    on person (person_extl_id);

alter table person
    add constraint person_pk
        primary key (person_id);

alter table person
    add constraint person_user_create_user_id_fk
        foreign key (create_user_id) references users
            deferrable initially deferred;

alter table person
    add constraint person_user_update_user_id_fk
        foreign key (update_user_id) references users
            deferrable initially deferred;

