    # db-init-config

Interactive setup for `config/config.cue` — the local configuration file needed before running `task db-init` and `task db-up`.

disable-model-invocation: true

---

## Instructions

You are helping the user create their `config/config.cue` file. This file is gitignored and contains local database credentials, an encryption key, and logger settings. It validates against the schema at `config/cue/schema.cue` and is exported to `config/config.json` by `task gen-config`.

### Step 1 — Collect admin target values

Use `AskUserQuestion` to collect the **admin PostgreSQL connection** values (used by `db-init` to connect as a superuser). Present all four questions in a single `AskUserQuestion` call:

1. **Admin DB host** — Options: `localhost` (Recommended), `127.0.0.1`. Header: `Admin host`. Question: "What is the host for your admin PostgreSQL connection?"
2. **Admin DB port** — Options: `5432` (Recommended), `5433`. Header: `Admin port`. Question: "What port is your admin PostgreSQL instance running on?"
3. **Admin DB name** — Options: `postgres` (Recommended), OS username. Header: `Admin DB`. Question: "What database should the admin connection use?" (Note: this is the database psql connects to, not the one being created.)
4. **Admin DB user** — Options: OS username (Recommended), `postgres`. Header: `Admin user`. Question: "What superuser should the admin connection use?"

After receiving answers, use a **second** `AskUserQuestion` call to ask:

5. **Admin DB password** — Options: `(empty / peer auth)` (Recommended), `Enter password`. Header: `Admin pass`. Question: "What is the password for the admin PostgreSQL user? (Leave empty if using peer/trust authentication.)"

### Step 2 — Collect app target values

Use `AskUserQuestion` to collect the **application database** values (what `db-init` will create). Present all six questions in a single call:

1. **App DB host** — Options: `localhost` (Recommended), `127.0.0.1`. Header: `App host`. Question: "What host should the application database use?"
2. **App DB port** — Options: `5432` (Recommended), `5433`. Header: `App port`. Question: "What port should the application database use?"
3. **App DB name** — Options: `dga_local` (Recommended), `diygoapi`. Header: `App DB name`. Question: "What should the application database be named?"
4. **App DB user** — Options: `demo_user` (Recommended), `dga_user`. Header: `App DB user`. Question: "What database user should be created for the application?"
5. **App DB password** — Options: `REPLACE_ME` (Recommended), `Enter password`. Header: `App DB pass`. Question: "What password should the application database user have?"
6. **App DB search path** — Options: `demo` (Recommended), `public`. Header: `Schema`. Question: "What PostgreSQL schema (search_path) should be used?"

### Step 3 — Generate encryption key

Run the following command and capture its output (trimmed):

```bash
go run ./cmd/newkey/main.go
```

Store the output as the encryption key value.

### Step 4 — Write config/config.cue

Write the file `config/config.cue` using the collected values. Use the exact template below, substituting the placeholder tokens with the values collected above.

For the admin password: if the user chose empty/peer auth, use an empty string `""`. Otherwise use the password they provided.

```cue
package config

_localAdminTarget: #Target & {
	target:               "local-admin"
	server_listener_port: 8080
	logger: {
		min_log_level:   "trace"
		log_level:       "debug"
		log_error_stack: true
	}
	encryption_key: "{{ENCRYPTION_KEY}}"
	database: {
		host:        "{{ADMIN_DB_HOST}}"
		port:        {{ADMIN_DB_PORT}}
		name:        "{{ADMIN_DB_NAME}}"
		user:        "{{ADMIN_DB_USER}}"
		password:    "{{ADMIN_DB_PASSWORD}}"
		search_path: "public"
	}
}

_localTarget: #Target & {
	target:               "local"
	server_listener_port: 8080
	logger: {
		min_log_level:   "trace"
		log_level:       "debug"
		log_error_stack: true
	}
	encryption_key: "{{ENCRYPTION_KEY}}"
	database: {
		host:        "{{APP_DB_HOST}}"
		port:        {{APP_DB_PORT}}
		name:        "{{APP_DB_NAME}}"
		user:        "{{APP_DB_USER}}"
		password:    "{{APP_DB_PASSWORD}}"
		search_path: "{{APP_DB_SEARCH_PATH}}"
	}
}

#Config & {
	default_target: "local"
	targets: [_localAdminTarget, _localTarget]
}
```

**Important**: The admin target's `search_path` is always `"public"` (the default PostgreSQL schema). The admin target's `password` must be a non-empty string to pass CUE validation — if the user chose empty/peer auth, use `"peer"` as a placeholder value (psql will use peer auth regardless when PGPASSWORD is empty in the db-init command).

### Step 5 — Run gen-config

Run `task gen-config` to validate and export the config:

```bash
task gen-config
```

If this succeeds, inform the user that their config is ready, and they can now run:

```
task db-init -- --db-admin-config-target local-admin
task db-up
```

If `task gen-config` fails, show the error to the user and help them fix the `config/config.cue` file.
