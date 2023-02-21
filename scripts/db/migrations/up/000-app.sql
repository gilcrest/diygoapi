create table if not exists app
(
    app_id                  uuid                     not null,
    app_extl_id             varchar                  not null,
    org_id                  uuid                     not null,
    app_name                varchar                  not null,
    app_description         varchar                  not null,
    auth_provider_id        integer,
    auth_provider_client_id varchar,
    create_app_id           uuid                     not null,
    create_user_id          uuid,
    create_timestamp        timestamp with time zone not null,
    update_app_id           uuid                     not null,
    update_user_id          uuid,
    update_timestamp        timestamp with time zone not null
);

comment on table app is 'app stores data about applications that interact with the system';

comment on column app.app_id is 'The Unique ID for the table.';

comment on column app.app_extl_id is 'The unique application External ID to be given to outside callers.';

comment on column app.org_id is 'The Foreign key for the organization that the app belongs to.';

comment on column app.app_name is 'The application name is a short name for the application.';

comment on column app.app_description is 'The application description is several sentences to describe the application.';

comment on column app.auth_provider_id is 'unique identifier representing authorization provider (e.g. Google, Github, etc.)';

comment on column app.auth_provider_client_id is 'Unique identifer of client ID given by an authentication provider. For example, GCP supports cross-client identity - see https://developers.google.com/identity/protocols/oauth2/cross-client-identity for a great explanation.';

comment on column app.create_app_id is 'The application which created this record.';

comment on column app.create_user_id is 'The user which created this record.';

comment on column app.create_timestamp is 'The timestamp when this record was created.';

comment on column app.update_app_id is 'The application which performed the most recent update to this record.';

comment on column app.update_user_id is 'The user which performed the most recent update to this record.';

comment on column app.update_timestamp is 'The timestamp when the record was updated most recently.';

create unique index if not exists app_app_extl_id_uindex
    on app (app_extl_id);

create unique index if not exists app_name_uindex
    on app (app_name);

create unique index if not exists auth_provider_client_id_ui
    on app (auth_provider_client_id);

alter table app
    add constraint app_pk
        primary key (app_id);

alter table app
    add constraint app_self_ref1
        foreign key (create_app_id) references app;

alter table app
    add constraint app_self_ref2
        foreign key (update_app_id) references app;

