create table demo.org_user
(
    user_id           uuid      not null,
    username          varchar   not null,
    org_id            uuid      not null,
    person_profile_id uuid      not null,
    create_app_id     uuid      not null,
    create_user_id    uuid,
    create_timestamp  timestamp not null,
    update_app_id     uuid      not null,
    update_user_id    uuid,
    update_timestamp  timestamp not null,
    constraint user_pk
        primary key (user_id),
    constraint user_self_ref_fk1
        foreign key (create_user_id) references demo.org_user,
    constraint user_self_ref_fk2
        foreign key (update_user_id) references demo.org_user,
    constraint org_user_create_app_id_fk
        foreign key (create_app_id) references demo.app
            deferrable initially deferred,
    constraint org_user_update_app_id_fk
        foreign key (update_app_id) references demo.app
            deferrable initially deferred
);

alter table demo.org_user
    owner to demo_user;

alter table demo.app
    add constraint org_user_fk1
        foreign key (create_user_id) references demo.org_user
            deferrable initially deferred;

alter table demo.app
    add constraint org_user_fk2
        foreign key (update_user_id) references demo.org_user
            deferrable initially deferred;

create unique index user_org_uindex
    on demo.org_user (username, org_id);

