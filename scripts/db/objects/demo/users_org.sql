create table if not exists users_org
(
    users_org_id     uuid                     not null,
    org_id           uuid                     not null,
    user_id          uuid                     not null,
    create_app_id    uuid                     not null,
    create_user_id   uuid,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null,
    update_user_id   uuid,
    update_timestamp timestamp with time zone not null
);

comment on column users_org.users_org_id is 'Unique identifier for a user''s association with an organization';

comment on column users_org.org_id is 'Organization ID foreign key to org table';

comment on column users_org.user_id is 'User ID foreign key to users table';

comment on column users_org.create_app_id is 'The application which created this record.';

comment on column users_org.create_user_id is 'The user which created this record.';

comment on column users_org.create_timestamp is 'The timestamp when this record was created.';

comment on column users_org.update_app_id is 'The application which performed the most recent update to this record.';

comment on column users_org.update_user_id is 'The user which performed the most recent update to this record.';

comment on column users_org.update_timestamp is 'The timestamp when the record was updated most recently.';

create unique index if not exists users_org_org_id_user_id_uindex
    on users_org (user_id, org_id);

alter table users_org
    add constraint users_org_pk
        primary key (users_org_id);

alter table users_org
    add constraint users_org_org_id_fk
        foreign key (org_id) references org
            deferrable initially deferred;

alter table users_org
    add constraint users_org_user_id_fk
        foreign key (user_id) references users
            deferrable initially deferred;

alter table users_org
    add constraint users_org_create_app_fk
        foreign key (create_app_id) references app
            deferrable initially deferred;

alter table users_org
    add constraint users_org_create_user_fk
        foreign key (create_user_id) references users
            deferrable initially deferred;

alter table users_org
    add constraint users_org_update_app_fk
        foreign key (update_app_id) references app
            deferrable initially deferred;

alter table users_org
    add constraint users_org_update_user_fk
        foreign key (update_user_id) references users
            deferrable initially deferred;

