create table if not exists org_kind
(
    org_kind_id      uuid                     not null,
    org_kind_extl_id varchar                  not null,
    org_kind_desc    varchar                  not null,
    create_app_id    uuid                     not null,
    create_user_id   uuid,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null,
    update_user_id   uuid,
    update_timestamp timestamp with time zone not null
);

comment on table org_kind is 'Organization Kind is a reference table denoting an organization''s (org) classification. Examples are Genesis, Test, Standard';

comment on column org_kind.org_kind_id is 'Organization Kind ID - pk for table';

comment on column org_kind.org_kind_extl_id is 'A short code denoting the organization kind';

comment on column org_kind.org_kind_desc is 'A longer descriptor of the organization kind';

comment on column org_kind.create_app_id is 'The application which created this record.';

comment on column org_kind.create_user_id is 'The user which created this record.';

comment on column org_kind.create_timestamp is 'The timestamp when this record was created.';

comment on column org_kind.update_app_id is 'The application which performed the most recent update to this record.';

comment on column org_kind.update_user_id is 'The user which performed the most recent update to this record.';

comment on column org_kind.update_timestamp is 'The timestamp when the record was updated most recently.';

create unique index if not exists org_kind_org_extl_id_uindex
    on org_kind (org_kind_extl_id);

alter table org_kind
    add constraint org_kind_pk
        primary key (org_kind_id);

alter table org_kind
    add constraint org_kind_create_app_fk
        foreign key (create_app_id) references app
            deferrable initially deferred;

alter table org_kind
    add constraint org_kind_create_user_fk
        foreign key (create_user_id) references users
            deferrable initially deferred;

alter table org_kind
    add constraint org_kind_update_app_fk
        foreign key (update_app_id) references app
            deferrable initially deferred;

alter table org_kind
    add constraint org_kind_update_user_fk
        foreign key (update_user_id) references users
            deferrable initially deferred;

