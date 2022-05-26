create table if not exists person
(
    person_id        uuid                     not null,
    org_id           uuid                     not null,
    create_app_id    uuid                     not null,
    create_user_id   uuid,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null,
    update_user_id   uuid,
    update_timestamp timestamp with time zone not null,
    constraint person_pk
        primary key (person_id),
    constraint person_create_app_fk
        foreign key (create_app_id) references app
            deferrable initially deferred,
    constraint person_update_app_fk
        foreign key (update_app_id) references app
            deferrable initially deferred,
    constraint person_create_user_fk
        foreign key (create_user_id) references org_user
            deferrable initially deferred,
    constraint person_update_user_fk
        foreign key (update_user_id) references org_user
            deferrable initially deferred
);

