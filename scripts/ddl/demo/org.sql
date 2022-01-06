create table org
(
    org_id           uuid                     not null,
    org_extl_id      varchar                  not null,
    org_name         varchar                  not null,
    org_description  varchar                  not null,
    create_app_id    uuid                     not null,
    create_user_id   uuid,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null,
    update_user_id   uuid,
    update_timestamp timestamp with time zone not null,
    constraint org_pk
        primary key (org_id),
    constraint org_create_user_fk
        foreign key (create_user_id) references app_user
            deferrable initially deferred,
    constraint org_update_user_fk
        foreign key (update_user_id) references app_user
            deferrable initially deferred,
    constraint org_create_app_fk
        foreign key (create_app_id) references app
            deferrable initially deferred,
    constraint org_update_app_fk
        foreign key (update_app_id) references app
            deferrable initially deferred
);

alter table org
    owner to postgres;

create unique index org_org_id_uindex
    on org (org_id);

create unique index org_org_name_uindex
    on org (org_name);

create unique index org_org_extl_id_uindex
    on org (org_extl_id);

