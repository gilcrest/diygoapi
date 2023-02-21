create table if not exists org
(
    org_id           uuid                     not null,
    org_extl_id      varchar                  not null,
    org_name         varchar                  not null,
    org_description  varchar                  not null,
    org_kind_id      uuid                     not null,
    create_app_id    uuid                     not null,
    create_user_id   uuid,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null,
    update_user_id   uuid,
    update_timestamp timestamp with time zone not null
);

comment on column org.org_id is 'Organization ID - Unique ID for table';

comment on column org.org_extl_id is 'Organization Unique External ID to be given to outside callers.';

comment on column org.org_name is 'Organization Name - a short name for the organization';

comment on column org.org_description is 'Organization Description - several sentences to describe the organization';

comment on column org.org_kind_id is 'Foreign Key to org_kind table.';

comment on column org.create_app_id is 'The application which created this record.';

comment on column org.create_user_id is 'The user which created this record.';

comment on column org.create_timestamp is 'The timestamp when this record was created.';

comment on column org.update_app_id is 'The application which performed the most recent update to this record.';

comment on column org.update_user_id is 'The user which performed the most recent update to this record.';

comment on column org.update_timestamp is 'The timestamp when the record was updated most recently.';

create unique index if not exists org_org_id_uindex
    on org (org_id);

create unique index if not exists org_org_name_uindex
    on org (org_name);

create unique index if not exists org_org_extl_id_uindex
    on org (org_extl_id);

alter table org
    add constraint org_pk
        primary key (org_id);

alter table org
    add constraint org_create_user_fk
        foreign key (create_user_id) references users
            deferrable initially deferred;

alter table org
    add constraint org_update_user_fk
        foreign key (update_user_id) references users
            deferrable initially deferred;

alter table org
    add constraint org_create_app_fk
        foreign key (create_app_id) references app
            deferrable initially deferred;

alter table org
    add constraint org_update_app_fk
        foreign key (update_app_id) references app
            deferrable initially deferred;

alter table org
    add constraint org_org_kind_fk
        foreign key (org_kind_id) references org_kind
            deferrable initially deferred;

