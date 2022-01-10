create table app
(
    app_id           uuid      not null,
    org_id           uuid      not null,
    app_extl_id      varchar   not null,
    app_name         varchar   not null,
    app_description  varchar   not null,
    active           boolean,
    create_app_id    uuid      not null,
    create_user_id   uuid,
    create_timestamp timestamp not null,
    update_app_id    uuid      not null,
    update_user_id   uuid,
    update_timestamp timestamp not null,
    constraint app_pk
        primary key (app_id),
    constraint app_self_ref1
        foreign key (create_app_id) references app,
    constraint app_self_ref2
        foreign key (update_app_id) references app,
    constraint app_user_fk1
        foreign key (create_user_id) references app_user
            deferrable initially deferred,
    constraint app_user_fk2
        foreign key (update_user_id) references app_user
            deferrable initially deferred,
    constraint app_org_org_id_fk
        foreign key (org_id) references org
            deferrable initially deferred
);

comment on table app is 'app stores data about applications that interact with the system';

alter table app
    owner to demo_user;

create unique index app_app_extl_id_uindex
    on app (app_extl_id);

