package config

config: #LocalConfig

config: encryptionKey: "31f8cbffe80df0067fbfac4abf0bb76c51d44cb82d2556743e6bf1a5e25d4e06"

config: httpServer: listenPort: 8080

config: logger: minLogLevel:   "trace"
config: logger: logLevel:      "debug"
config: logger: logErrorStack: false

config: database: host:       "localhost"
config: database: port:       5432
config: database: name:       "dga_local"
config: database: user:       "demo_user"
config: database: password:   "REPLACE_ME"
config: database: searchPath: "demo"
