create table if not exists auth_provider
(
    auth_provider_id   bigint                   not null,
    auth_provider_cd   varchar                  not null,
    auth_provider_desc varchar                  not null,
    create_app_id      uuid                     not null,
    create_user_id     uuid,
    create_timestamp   timestamp with time zone not null,
    update_app_id      uuid                     not null,
    update_user_id     uuid,
    update_timestamp   timestamp with time zone not null
);

comment on table auth_provider is 'Authentication Provider (e.g. Google, Github, Apple, Facebook, etc.)';

comment on column auth_provider.auth_provider_id is 'Unique ID representing the authentication provider.';

comment on column auth_provider.auth_provider_cd is 'Short code representing the authentication provider (e.g., google, github, apple, etc.)';

comment on column auth_provider.auth_provider_desc is 'Longer description of authentication provider';

comment on column auth_provider.create_app_id is 'The application which created this record.';

comment on column auth_provider.create_user_id is 'The user which created this record.';

comment on column auth_provider.create_timestamp is 'The timestamp when this record was created.';

comment on column auth_provider.update_app_id is 'The application which performed the most recent update to this record.';

comment on column auth_provider.update_user_id is 'The user which performed the most recent update to this record.';

comment on column auth_provider.update_timestamp is 'The timestamp when the record was updated most recently.';

create unique index if not exists auth_provider_cd_ui
    on auth_provider (auth_provider_cd);

alter table auth_provider
    add constraint auth_provider_pk
        primary key (auth_provider_id);

alter table auth_provider
    add constraint auth_provider_create_app_fk
        foreign key (create_app_id) references app
            deferrable initially deferred;

alter table auth_provider
    add constraint auth_provider_create_user_fk
        foreign key (create_user_id) references users
            deferrable initially deferred;

alter table auth_provider
    add constraint auth_provider_update_app_fk
        foreign key (update_app_id) references app
            deferrable initially deferred;

alter table auth_provider
    add constraint auth_provider_update_user_fk
        foreign key (update_user_id) references users
            deferrable initially deferred;

