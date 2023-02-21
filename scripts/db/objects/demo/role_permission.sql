create table if not exists role_permission
(
    role_id          uuid                     not null,
    permission_id    uuid                     not null,
    create_app_id    uuid                     not null,
    create_user_id   uuid,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null,
    update_user_id   uuid,
    update_timestamp timestamp with time zone not null
);

comment on table role_permission is 'The role_permission table stores which roles have which permissions.';

comment on column role_permission.role_id is 'The unique role which can have 1 to many permissions set in this table.';

comment on column role_permission.permission_id is 'The unique permission that is being given to the role.';

comment on column role_permission.create_app_id is 'The application which created this record.';

comment on column role_permission.create_user_id is 'The user which created this record.';

comment on column role_permission.create_timestamp is 'The timestamp when this record was created.';

comment on column role_permission.update_app_id is 'The application which performed the most recent update to this record.';

comment on column role_permission.update_user_id is 'The user which performed the most recent update to this record.';

comment on column role_permission.update_timestamp is 'The timestamp when the record was updated most recently.';

alter table role_permission
    add constraint role_permission_pk
        primary key (role_id, permission_id);

alter table role_permission
    add constraint role_permission_role_id_fk
        foreign key (role_id) references role;

alter table role_permission
    add constraint role_permission_permission_id_fk
        foreign key (permission_id) references permission;

alter table role_permission
    add constraint role_permission_create_app_fk
        foreign key (create_app_id) references app
            deferrable initially deferred;

alter table role_permission
    add constraint role_permission_update_app_fk
        foreign key (update_app_id) references app
            deferrable initially deferred;

alter table role_permission
    add constraint role_permission_create_user_fk
        foreign key (create_user_id) references users
            deferrable initially deferred;

alter table role_permission
    add constraint role_permission_update_user_fk
        foreign key (update_user_id) references users
            deferrable initially deferred;

