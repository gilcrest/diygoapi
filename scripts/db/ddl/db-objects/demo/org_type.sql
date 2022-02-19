create table org_type
(
    org_type_id      uuid                     not null,
    org_type         varchar                  not null,
    org_type_desc    varchar                  not null,
    create_app_id    uuid                     not null,
    create_user_id   uuid,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null,
    update_user_id   uuid,
    update_timestamp timestamp with time zone not null,
    constraint org_type_pk
        primary key (org_type_id),
    constraint org_type_create_app_fk
        foreign key (create_app_id) references app
            deferrable initially deferred,
    constraint org_type_create_user_fk
        foreign key (create_user_id) references app_user
            deferrable initially deferred,
    constraint org_type_update_app_fk
        foreign key (update_app_id) references app
            deferrable initially deferred,
    constraint org_type_update_user_fk
        foreign key (update_user_id) references app_user
            deferrable initially deferred
);

comment on table org_type is 'Organization Type is a reference table denoting an organization''s (org) type. Examples are Genesis, Test, Standard';

comment on column org_type.org_type_id is 'Organization Type ID - pk for table';

comment on column org_type.org_type is 'A short description of the organization type';

comment on column org_type.org_type_desc is 'A long description of the organization type';

comment on column org_type.create_app_id is 'The application which created this record.';

comment on column org_type.create_user_id is 'The user which created this record.';

comment on column org_type.create_timestamp is 'The timestamp when this record was created.';

comment on column org_type.update_app_id is 'The application which performed the most recent update to this record.';

comment on column org_type.update_user_id is 'The user which performed the most recent update to this record.';

comment on column org_type.update_timestamp is 'The timestamp when the record was updated most recently.';

alter table org_type
    owner to demo_user;

create unique index org_type_org_type_uindex
    on org_type (org_type);

