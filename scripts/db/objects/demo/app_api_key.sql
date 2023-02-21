create table if not exists app_api_key
(
    api_key          varchar                  not null,
    app_id           uuid                     not null,
    deactv_date      date                     not null,
    create_app_id    uuid                     not null,
    create_user_id   uuid,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null,
    update_user_id   uuid,
    update_timestamp timestamp with time zone not null
);

comment on column app_api_key.api_key is 'app_key is a hash of a key given to an person for an app';

comment on column app_api_key.app_id is 'foreign key to app table';

alter table app_api_key
    add constraint app_key_pk
        primary key (api_key);

alter table app_api_key
    add constraint app_key_app_app_id_fk
        foreign key (app_id) references app;

alter table app_api_key
    add constraint app_api_key_app_fk1
        foreign key (create_app_id) references app
            deferrable initially deferred;

alter table app_api_key
    add constraint app_api_key_users_fk1
        foreign key (create_user_id) references users
            deferrable initially deferred;

alter table app_api_key
    add constraint app_api_key_app_fk2
        foreign key (update_app_id) references app
            deferrable initially deferred;

alter table app_api_key
    add constraint app_api_key_users_fk2
        foreign key (update_user_id) references users
            deferrable initially deferred;

