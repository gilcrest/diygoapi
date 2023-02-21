create table if not exists users_role
(
    user_id          uuid                     not null,
    role_id          uuid                     not null,
    org_id           uuid                     not null,
    create_app_id    uuid                     not null,
    create_user_id   uuid,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null,
    update_user_id   uuid,
    update_timestamp timestamp with time zone not null
);

comment on table users_role is 'The users_role table stores which users have which role(s) in which organization(s).';

comment on column users_role.user_id is 'The user which is being given a role (within an organization).';

comment on column users_role.role_id is 'The role which can have one to many users set in this table.';

comment on column users_role.org_id is 'The organization to which the role and user are associated.';

comment on column users_role.create_app_id is 'The application which created this record.';

comment on column users_role.create_user_id is 'The user which created this record.';

comment on column users_role.create_timestamp is 'The timestamp when this record was created.';

comment on column users_role.update_app_id is 'The application which performed the most recent update to this record.';

comment on column users_role.update_user_id is 'The user which performed the most recent update to this record.';

comment on column users_role.update_timestamp is 'The timestamp when the record was updated most recently.';

alter table users_role
    add constraint users_role_pk
        primary key (user_id, role_id, org_id);

alter table users_role
    add constraint users_role_user_id_fk
        foreign key (user_id) references users;

alter table users_role
    add constraint users_role_role_id_fk
        foreign key (role_id) references role;

alter table users_role
    add constraint users_role_org_id_fk
        foreign key (org_id) references org;

alter table users_role
    add constraint users_role_create_app_fk
        foreign key (create_app_id) references app
            deferrable initially deferred;

alter table users_role
    add constraint users_role_update_app_fk
        foreign key (update_app_id) references app
            deferrable initially deferred;

alter table users_role
    add constraint users_role_create_user_fk
        foreign key (create_user_id) references users
            deferrable initially deferred;

alter table users_role
    add constraint users_role_update_user_fk
        foreign key (update_user_id) references users
            deferrable initially deferred;

