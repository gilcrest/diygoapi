create table if not exists users_lang_prefs
(
    user_id          uuid                     not null,
    language_tag     varchar                  not null,
    create_app_id    uuid                     not null,
    create_user_id   uuid,
    create_timestamp timestamp with time zone not null,
    update_app_id    uuid                     not null,
    update_user_id   uuid,
    update_timestamp timestamp with time zone not null
);

comment on table users_lang_prefs is 'The users_lang_prefs table stores the list of language tag preferences for the user.';

comment on column users_lang_prefs.user_id is 'The user having a language tag preference.';

comment on column users_lang_prefs.language_tag is 'The BCP 47 Language Tag which identifies a language both spoken and written.';

comment on column users_lang_prefs.create_app_id is 'The application which created this record.';

comment on column users_lang_prefs.create_user_id is 'The user which created this record.';

comment on column users_lang_prefs.create_timestamp is 'The timestamp when this record was created.';

comment on column users_lang_prefs.update_app_id is 'The application which performed the most recent update to this record.';

comment on column users_lang_prefs.update_user_id is 'The user which performed the most recent update to this record.';

comment on column users_lang_prefs.update_timestamp is 'The timestamp when the record was updated most recently.';

alter table users_lang_prefs
    add constraint users_lang_prefs_pk
        primary key (user_id, language_tag);

alter table users_lang_prefs
    add constraint users_lang_prefs_user_id_fk
        foreign key (user_id) references users;

alter table users_lang_prefs
    add constraint users_lang_prefs_create_app_fk
        foreign key (create_app_id) references app
            deferrable initially deferred;

alter table users_lang_prefs
    add constraint users_lang_prefs_update_app_fk
        foreign key (update_app_id) references app
            deferrable initially deferred;

alter table users_lang_prefs
    add constraint users_lang_prefs_create_user_fk
        foreign key (create_user_id) references users
            deferrable initially deferred;

alter table users_lang_prefs
    add constraint users_lang_prefs_update_user_fk
        foreign key (update_user_id) references users
            deferrable initially deferred;

