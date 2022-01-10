-- create user is synonymous with create role. A role is an entity that can own database objects
-- and have database privileges; a role can be considered a “user”, a “group”, or both depending
-- on how it is used.
create user demo_user with createdb password 'REPLACE_ME';

alter user demo_user with nosuperuser;

-- create database for the environment (gab_local, gab_nonprod, gab_prod, etc.)
-- gab = go-api-basic :)
create database gab_local with owner demo_user;

-- create schema demo with owner demo_user
create schema if not exists demo authorization demo_user;

-- get list of users and roles. Not necessary to run, just a helpful bit from
-- https://www.postgresqltutorial.com/postgresql-list-users/ if you want to validate users
SELECT usename AS role_name,
       CASE
           WHEN usesuper AND usecreatedb THEN
               CAST('superuser, create database' AS pg_catalog.text)
           WHEN usesuper THEN
               CAST('superuser' AS pg_catalog.text)
           WHEN usecreatedb THEN
               CAST('create database' AS pg_catalog.text)
           ELSE
               CAST('' AS pg_catalog.text)
           END role_attributes
FROM pg_catalog.pg_user
ORDER BY role_name desc;