package config

config: #GCPConfig

config: encryptionKey: "d9291b175784efbaa49f88a3891612b85889311fcbd9b3df34c7e410e9ddef7c"

config: httpServer: listenPort: 8080

config: logger: minLogLevel:   "trace"
config: logger: logLevel:      "debug"
config: logger: logErrorStack: true

config: database: host:       "/cloudsql/diy-go-api:us-central1:diy-go-api-db"
config: database: port:       5432
config: database: name:       "dga_staging"
config: database: user:       "demo_user"
config: database: password:   "REPLACE_ME"
config: database: searchPath: "demo"

config: gcp: projectID: "fide-nonprod"
config: gcp: cloudSQL: instanceName:            "diy-go-api-db"
config: gcp: cloudSQL: instanceConnectionName:  "diy-go-api:us-central1:diy-go-api-db"
config: gcp: cloudRun: serviceName:             "staging"
config: gcp: artifactRegistry: repoLocation: "us-central1"
config: gcp: artifactRegistry: repoName:     "diy-go-api-docker-repo"
config: gcp: artifactRegistry: imageID:      "staging"
config: gcp: artifactRegistry: tag:          "latest"
