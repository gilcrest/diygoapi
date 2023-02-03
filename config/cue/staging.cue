package config

config: #GCPConfig

config: encryptionKey: "d9291b175784efbaa49f88a3891612b85889311fcbd9b3df34c7e410e9ddef7c"

config: httpServer: listenPort: 8080

config: logger: minLogLevel:   "trace"
config: logger: logLevel:      "debug"
config: logger: logErrorStack: false

config: database: host:       "/cloudsql/diygoapi:us-central1:diygoapi-db"
config: database: port:       5432
config: database: name:       "dga_staging"
config: database: user:       "demo_user"
config: database: password:   "REPLACE_ME"
config: database: searchPath: "demo"

config: gcp: projectID: "diy-nonprod"
config: gcp: cloudSQL: instanceName:           "diygoapi-db"
config: gcp: cloudSQL: instanceConnectionName: "diygoapi:us-central1:diygoapi-db"
config: gcp: cloudRun: serviceName:            "staging"
config: gcp: artifactRegistry: repoLocation:   "us-central1"
config: gcp: artifactRegistry: repoName:       "diygoapi-docker-repo"
config: gcp: artifactRegistry: imageID:        "staging"
config: gcp: artifactRegistry: tag:            "latest"
