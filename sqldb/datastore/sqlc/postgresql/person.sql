-- name: CreatePerson :execrows
INSERT INTO person (person_id, person_extl_id, create_app_id, create_user_id,
                    create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: FindPersonByUserID :one
SELECT p.person_id,
       p.person_extl_id,
       pa.user_id,
       pa.user_extl_id,
       pa.name_prefix,
       pa.first_name,
       pa.middle_name,
       pa.last_name,
       pa.name_suffix,
       pa.nickname,
       pa.email,
       pa.company_name,
       pa.company_dept,
       pa.job_title,
       pa.birth_date,
       pa.birth_year,
       pa.birth_month,
       pa.birth_day
FROM person p
         inner join users pa on pa.person_id = p.person_id
WHERE pa.user_id = $1;

-- name: FindPersonByUserExternalID :one
SELECT p.person_id,
       p.person_extl_id,
       pa.user_id,
       pa.user_extl_id,
       pa.name_prefix,
       pa.first_name,
       pa.middle_name,
       pa.last_name,
       pa.name_suffix,
       pa.nickname,
       pa.email,
       pa.company_name,
       pa.company_dept,
       pa.job_title,
       pa.birth_date,
       pa.birth_year,
       pa.birth_month,
       pa.birth_day
FROM person p
         inner join users pa on pa.person_id = p.person_id
WHERE pa.user_extl_id = $1;

-- name: DeletePerson :execrows
DELETE
FROM person
WHERE person_id = $1;

-- name: CreateUsersOrg :execrows
INSERT INTO users_org (users_org_id, org_id, user_id,
                       create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: CreateUser :execrows
INSERT INTO users (user_id, user_extl_id, person_id, name_prefix, first_name, middle_name, last_name, name_suffix,
                   nickname, email, company_name, company_dept, job_title, birth_date, birth_year, birth_month, birth_day,
                   create_app_id, create_user_id, create_timestamp,
                   update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23);

-- name: FindUserByID :one
SELECT * FROM users
WHERE user_id = $1;

-- name: FindUserByExternalID :one
SELECT * FROM users
WHERE user_extl_id = $1;

-- name: DeleteUserByID :execrows
DELETE FROM users
WHERE user_id = $1;

-- name: FindUserLanguagePreferencesByUserID :many
SELECT *
FROM users_lang_prefs
WHERE user_id = $1;

-- name: CreateUserLanguagePreference :execrows
INSERT INTO users_lang_prefs (user_id, language_tag, create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: DeleteUserLanguagePreferences :execrows
DELETE FROM users_lang_prefs
WHERE user_id = $1;
