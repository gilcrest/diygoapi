create table if not exists role
(
    role_id          uuid                     not null,
    role_extl_id     varchar                  not null,
    role_cd          varchar                  not null,
    role_description varchar                  not null,
    active           boolean                  not null,
    create_app_id    uuid                     not null,
    create_user_id   uuid,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null,
    update_user_id   uuid,
    update_timestamp timestamp with time zone not null
);

comment on table role is 'The role table stores a job function or title which defines an authority level.';

comment on column role.role_id is 'The unique ID for the table.';

comment on column role.role_extl_id is 'Unique External ID to be given to outside callers.';

comment on column role.role_cd is 'A human-readable code which represents the role.';

comment on column role.role_description is 'A longer description of the role.';

comment on column role.active is 'A boolean denoting whether the role is active (true) or not (false).';

comment on column role.create_app_id is 'The application which created this record.';

comment on column role.create_user_id is 'The user which created this record.';

comment on column role.create_timestamp is 'The timestamp when this record was created.';

comment on column role.update_app_id is 'The application which performed the most recent update to this record.';

comment on column role.update_user_id is 'The user which performed the most recent update to this record.';

comment on column role.update_timestamp is 'The timestamp when the record was updated most recently.';

alter table role
    add constraint role_pk
        primary key (role_id);

alter table role
    add constraint role_cd_ui
        unique (role_cd);

alter table role
    add constraint role_create_app_fk
        foreign key (create_app_id) references app
            deferrable initially deferred;

alter table role
    add constraint role_update_app_fk
        foreign key (update_app_id) references app
            deferrable initially deferred;

alter table role
    add constraint role_create_user_fk
        foreign key (create_user_id) references users
            deferrable initially deferred;

alter table role
    add constraint role_update_user_fk
        foreign key (update_user_id) references users
            deferrable initially deferred;

