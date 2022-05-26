create table if not exists role_user
(
    role_id          uuid                     not null,
    user_id          uuid                     not null,
    create_app_id    uuid                     not null,
    create_user_id   uuid,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null,
    update_user_id   uuid,
    update_timestamp timestamp with time zone not null,
    constraint role_user_pk
        primary key (role_id, user_id),
    constraint role_user_role_id_fk
        foreign key (role_id) references role,
    constraint role_user_user_id_fk
        foreign key (user_id) references org_user,
    constraint role_user_create_app_fk
        foreign key (create_app_id) references app
            deferrable initially deferred,
    constraint role_user_update_app_fk
        foreign key (update_app_id) references app
            deferrable initially deferred,
    constraint role_user_create_user_fk
        foreign key (create_user_id) references org_user
            deferrable initially deferred,
    constraint role_user_update_user_fk
        foreign key (update_user_id) references org_user
            deferrable initially deferred
);

comment on table role_user is 'The role_user table stores which roles have which users.';

comment on column role_user.role_id is 'The unique role which can have one to many users set in this table.';

comment on column role_user.user_id is 'The unique user that is being given the role.';

comment on column role_user.create_app_id is 'The application which created this record.';

comment on column role_user.create_user_id is 'The user which created this record.';

comment on column role_user.create_timestamp is 'The timestamp when this record was created.';

comment on column role_user.update_app_id is 'The application which performed the most recent update to this record.';

comment on column role_user.update_user_id is 'The user which performed the most recent update to this record.';

comment on column role_user.update_timestamp is 'The timestamp when the record was updated most recently.';

