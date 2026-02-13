package config

#Config: {
	default_target: string
	targets: [...#Target]
}

#Target: {
	target:               string
	server_listener_port: >=8080 & <=10080
	logger:               #Logger
	encryption_key:       !="" // must be specified and non-empty
	database:             #Database
	_gcp:                 #GCP
}

#LogLevel: "trace" | "debug" | "info" | "warn" | "error" | "fatal" | "panic" | "disabled"

#Logger: {
	// minimum accepted log level
	min_log_level: #LogLevel
	// log level
	log_level: #LogLevel
	// log error stack
	log_error_stack: bool
}

#Database: {
	host:        !="" // must be specified and non-empty
	port:        !=0  // must be specified and non-empty
	name:        !="" // must be specified and non-empty
	user:        !="" // must be specified and non-empty
	password:    !="" // must be specified and non-empty
	search_path: !="" // must be specified and non-empty
}

#GCP: {
	// Google Cloud project ID
	project_id:        !="" // must be specified and non-empty
	artifact_registry: #ArtifactRegistry
	cloud_sql:         #CloudSQL
	cloud_run:         #CloudRun
}

#ArtifactRegistry: {
	// Regional or multi-regional location of the Artifact Registry repository
	repo_location: !="" // must be specified and non-empty

	// Artifact Registry Repository name
	repo_name: !="" // must be specified and non-empty

	// Build Image ID
	image_id: !="" // must be specified and non-empty

	// Build Image Tag
	tag: !="" // must be specified and non-empty
}

#CloudSQL: {
	// Instance Name
	instance_name: !="" // must be specified and non-empty
	// Instance Connection Name
	instance_connection_name: !="" // must be specified and non-empty
}

#CloudRun: {
	// Service Name
	service_name: !="" // must be specified and non-empty
}
