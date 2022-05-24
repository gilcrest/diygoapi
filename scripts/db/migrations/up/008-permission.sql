create table permission
(
    permission_id          uuid                     not null
        constraint permission_pk
            primary key,
    permission_extl_id     varchar                  not null,
    resource               varchar                  not null,
    operation              varchar                  not null,
    permission_description varchar                  not null,
    active                 boolean                  not null,
    create_app_id          uuid                     not null
        constraint permission_create_app_fk
            references app
            deferrable initially deferred,
    create_user_id         uuid
        constraint permission_create_user_fk
            references org_user
            deferrable initially deferred,
    create_timestamp       timestamp with time zone not null,
    update_app_id          uuid                     not null
        constraint permission_update_app_fk
            references app
            deferrable initially deferred,
    update_user_id         uuid
        constraint permission_update_user_fk
            references org_user
            deferrable initially deferred,
    update_timestamp       timestamp with time zone not null,
    constraint permission_ui
        unique (permission_id),
    constraint permission_resource_ui
        unique (resource, operation)
);

comment on table permission is 'The permission table stores an approval of a mode of access to a resource.';

comment on column permission.permission_id is 'The unique ID for the table.';

comment on column permission.permission_extl_id is 'Unique External ID to be given to outside callers.';

comment on column permission.resource is 'A human-readable string which represents a resource (e.g. an HTTP route or document, etc.).';

comment on column permission.operation is 'A string representing the action taken on the resource (e.g. POST, GET, edit, etc.)';

comment on column permission.permission_description is 'A description of what the permission is granting, e.g. "grants ability to edit a billing document".';

comment on column permission.active is 'A boolean denoting whether the permission is active (true) or not (false).';

comment on column permission.create_app_id is 'The application which created this record.';

comment on column permission.create_user_id is 'The user which created this record.';

comment on column permission.create_timestamp is 'The timestamp when this record was created.';

comment on column permission.update_app_id is 'The application which performed the most recent update to this record.';

comment on column permission.update_user_id is 'The user which performed the most recent update to this record.';

comment on column permission.update_timestamp is 'The timestamp when the record was updated most recently.';

create unique index permission_extl_id_uindex
    on permission (permission_extl_id);
