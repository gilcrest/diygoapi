package config

#Base: {
	encryptionKey: !="" // must be specified and non-empty
}

#HTTPServer: {
	listenPort: >=8080 & <=10080
}

#Logger: {
	// minimum accepted log level
	minLogLevel: #LogLevels
	// log level
	logLevel: #LogLevels
	// log error stack
	logErrorStack: bool
}

#Database: {
	host:       !="" // must be specified and non-empty
	port:       !=0  // must be specified and non-empty
	name:       !="" // must be specified and non-empty
	user:       !="" // must be specified and non-empty
	password:   !="" // must be specified and non-empty
	searchPath: !="" // must be specified and non-empty
}

#GCP: {
	// Google Cloud project ID
	projectID:        !="" // must be specified and non-empty
	artifactRegistry: #ArtifactRegistry
	cloudSQL:         #CloudSQL
	cloudRun:         #CloudRun
}

#ArtifactRegistry: {
	// Regional or multi-regional location of the Artifact Registry repository
	repoLocation: !="" // must be specified and non-empty

	// Artifact Registry Repository name
	repoName: !="" // must be specified and non-empty

	// Build Image ID
	imageID: !="" // must be specified and non-empty

	// Build Image Tag
	tag: !="" // must be specified and non-empty
}

#CloudSQL: {
	// Instance Name
	instanceName: !="" // must be specified and non-empty
	// Instance Connection Name
	instanceConnectionName: !="" // must be specified and non-empty
}

#CloudRun: {
	// Service Name
	serviceName: !="" // must be specified and non-empty
}

#LogLevels: "trace" | "debug" | "info" | "warn" | "error" | "fatal" | "panic" | "disabled"

#LocalConfig: {
	#Base
	httpServer: #HTTPServer
	logger:     #Logger
	database:   #Database
}

#GCPConfig: {
	#Base
	httpServer: #HTTPServer
	logger:     #Logger
	database:   #Database
	gcp:        #GCP
}
