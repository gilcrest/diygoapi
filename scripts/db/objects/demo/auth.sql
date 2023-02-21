create table if not exists auth
(
    auth_id                           uuid                     not null,
    user_id                           uuid                     not null,
    auth_provider_id                  bigint                   not null,
    auth_provider_cd                  varchar                  not null,
    auth_provider_client_id           varchar,
    auth_provider_person_id           varchar                  not null,
    auth_provider_access_token        varchar                  not null,
    auth_provider_refresh_token       varchar,
    auth_provider_access_token_expiry timestamp with time zone not null,
    create_app_id                     uuid                     not null,
    create_user_id                    uuid,
    create_timestamp                  timestamp with time zone not null,
    update_app_id                     uuid                     not null,
    update_user_id                    uuid,
    update_timestamp                  timestamp with time zone not null
);

comment on table auth is 'The auth table stores which user has authenticated through an Oauth2 provider.';

comment on column auth.auth_id is 'The unique id given to the authorization.';

comment on column auth.auth_provider_id is 'Unique ID given to an authorization provider.';

comment on column auth.auth_provider_cd is 'Unique code given to an authorization provider (e.g. google).';

comment on column auth.auth_provider_client_id is 'External ID (given by authorization provider) which represents the Oauth2 client which authenticated the user';

comment on column auth.auth_provider_person_id is 'Unique ID given by the authorization provider which represents the person.';

comment on column auth.auth_provider_access_token is 'Oauth2 access token given by the authorization provider.';

comment on column auth.auth_provider_refresh_token is 'OAuth2 refresh token given by the authorization provider.';

comment on column auth.auth_provider_access_token_expiry is 'Expiration of access token given by the authorization provider. Is not a perfect precision instrument as some providers do not give an exact time, but rather seconds until expiration, which means the value is calculated relative to the server time.';

comment on column auth.create_app_id is 'The application which created this record.';

comment on column auth.create_user_id is 'The user which created this record.';

comment on column auth.create_timestamp is 'The timestamp when this record was created.';

comment on column auth.update_app_id is 'The application which performed the most recent update to this record.';

comment on column auth.update_user_id is 'The user which performed the most recent update to this record.';

comment on column auth.update_timestamp is 'The timestamp when the record was updated most recently.';

create unique index if not exists auth_access_token_ui
    on auth (auth_provider_access_token);

comment on index auth_access_token_ui is 'Only one access token per authentication is allowed';

create unique index if not exists auth_auth_provider_person_id_ui
    on auth (auth_provider_id, auth_provider_person_id);

comment on index auth_auth_provider_person_id_ui is 'one auth per provider person';

create unique index if not exists auth_user_provider_ui
    on auth (user_id, auth_provider_id);

comment on index auth_user_provider_ui is 'one provider per user';

alter table auth
    add constraint auth_pk
        primary key (auth_id);

alter table auth
    add constraint auth_user_id_fk
        foreign key (user_id) references users;

alter table auth
    add constraint auth_create_app_fk
        foreign key (create_app_id) references app
            deferrable initially deferred;

alter table auth
    add constraint auth_create_user_fk
        foreign key (create_user_id) references users
            deferrable initially deferred;

alter table auth
    add constraint auth_update_app_fk
        foreign key (update_app_id) references app
            deferrable initially deferred;

alter table auth
    add constraint auth_update_user_fk
        foreign key (update_user_id) references users
            deferrable initially deferred;

alter table auth
    add constraint auth_auth_provider_auth_provider_id_fk
        foreign key (auth_provider_id) references auth_provider;

