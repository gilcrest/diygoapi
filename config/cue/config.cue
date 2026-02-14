package config

#Config & {
	default_target: "local"
	targets: [_localTarget, _stagingTarget]
}

_localTarget: {#Target & {
	target:               "local"
	server_listener_port: 8080
	logger: {#Logger & {
		min_log_level:   "trace"
		log_level:       "debug"
		log_error_stack: true
	}}
	encryption_key: "31f8cbffe80df0067fbfac4abf0bb76c51d44cb82d2556743e6bf1a5e25d4e06"
	database: {
		host:        "localhost"
		port:        5432
		name:        "dga_local"
		user:        "demo_user"
		password:    "REPLACE_ME"
		search_path: "demo"
	}
}}

_stagingTarget: {#Target & {
	target:               "staging"
	server_listener_port: 8080
	logger: {#Logger & {
		min_log_level:   "trace"
		log_level:       "debug"
		log_error_stack: false
	}}
	encryption_key: "d9291b175784efbaa49f88a3891612b85889311fcbd9b3df34c7e410e9ddef7c"
	database: {
		host:        "/cloudsql/diygoapi:us-central1:diygoapi-db"
		port:        5432
		name:        "dga_staging"
		user:        "demo_user"
		password:    "REPLACE_ME"
		search_path: "demo"
	}
	_gcp: {
		project_id: "diy-nonprod"
		cloud_sql: {
			instance_name:            "diygoapi-db"
			instance_connection_name: "diygoapi:us-central1:diygoapi-db"
		}
		cloud_run: {
			service_name: "staging"
		}
		artifact_registry: {
			repo_location: "us-central1"
			repo_name:     "diygoapi-docker-repo"
			image_id:      "staging"
			tag:           "latest"
		}
	}
}}
