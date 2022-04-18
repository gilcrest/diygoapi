package config

config: #LocalConfig

config: encryptionKey: "9e44fd332e8060025eb7de13c56c2cc260286ca22241a2ac87fc97a5e4a185ac"

config: httpServer: listenPort: 8080

config: logger: minLogLevel:   "trace"
config: logger: logLevel:      "debug"
config: logger: logErrorStack: true

config: database: host:       "localhost"
config: database: port:       5432
config: database: name:       "gab_local"
config: database: user:       "demo_user"
config: database: password:   "REPLACE_ME"
config: database: searchPath: "demo"
