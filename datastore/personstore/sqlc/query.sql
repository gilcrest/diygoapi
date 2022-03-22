-- name: CreatePerson :execrows
INSERT INTO person (person_id, org_id, create_app_id, create_user_id,
                    create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: FindPersonProfileByID :one
SELECT * FROM person_profile
WHERE person_id = $1 LIMIT 1;

-- name: CreatePersonProfile :execrows
INSERT INTO person_profile (person_profile_id, person_id, name_prefix, first_name, middle_name, last_name, name_suffix,
                            nickname, company_name, company_dept, job_title, birth_date, birth_year, birth_month,
                            birth_day, language_id,
                            create_app_id, create_user_id, create_timestamp, update_app_id,
                            update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22);

-- name: DeletePersonProfile :execrows
DELETE FROM person_profile
WHERE person_id = $1;
