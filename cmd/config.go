package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/sqldb"
)

const (
	// local JSON Config File path - relative to project root
	localJSONConfigFile = "./config/local.json"
	// staging JSON Config File path - relative to project root
	stagingJSONConfigFile = "./config/staging.json"
	// production JSON Config File path - relative to project root
	productionJSONConfigFile = "./config/production.json"
	// genesisRequestFile is the local JSON Genesis Request File path
	// (relative to project root)
	genesisRequestFile = "./config/genesis/request.json"
)

// Env defines the environment
type Env uint8

const (
	// Existing environment - current environment is not overridden
	Existing Env = iota
	// Local environment (Local machine)
	Local
	// Staging environment (GCP)
	Staging
	// Production environment (GCP)
	Production

	// Invalid defines an invalid environment option
	Invalid Env = 99
)

func (e Env) String() string {
	switch e {
	case Existing:
		return "existing"
	case Local:
		return "local"
	case Staging:
		return "staging"
	case Production:
		return "production"
	case Invalid:
		return "invalid"
	}
	return "unknown_env_config"
}

// ParseEnv converts an env string into an Env value.
// returns Invalid if the input string does not match known values.
func ParseEnv(envStr string) Env {
	switch envStr {
	case "existing":
		return Existing
	case "local":
		return Local
	case "staging":
		return Staging
	case "prod":
		return Production
	default:
		return Invalid
	}
}

// ConfigFile defines the configuration file. It is the superset of
// fields for the various environments/builds. For example, when setting
// the local environment based on the ConfigFile, you do not need
// to fill any of the GCP fields.
type ConfigFile struct {
	Config struct {
		HTTPServer struct {
			ListenPort int `json:"listenPort"`
		} `json:"httpServer"`
		Logger struct {
			MinLogLevel   string `json:"minLogLevel"`
			LogLevel      string `json:"logLevel"`
			LogErrorStack bool   `json:"logErrorStack"`
		} `json:"logger"`
		Database struct {
			Host       string `json:"host"`
			Port       int    `json:"port"`
			Name       string `json:"name"`
			User       string `json:"user"`
			Password   string `json:"password"`
			SearchPath string `json:"searchPath"`
		} `json:"database"`
		EncryptionKey string `json:"encryptionKey"`
		GCP           struct {
			ProjectID        string `json:"projectID"`
			ArtifactRegistry struct {
				RepoLocation string `json:"repoLocation"`
				RepoName     string `json:"repoName"`
				ImageID      string `json:"imageID"`
				Tag          string `json:"tag"`
			} `json:"artifactRegistry"`
			CloudSQL struct {
				InstanceName           string `json:"instanceName"`
				InstanceConnectionName string `json:"instanceConnectionName"`
			} `json:"cloudSQL"`
			CloudRun struct {
				ServiceName string `json:"serviceName"`
			} `json:"cloudRun"`
		} `json:"gcp"`
	} `json:"config"`
}

// LoadEnv conditionally sets the environment from a config file
// relative to whichever environment is being set. If Existing is
// passed as EnvConfig, the current environment is used and not overridden.
func LoadEnv(env Env) (err error) {
	const op errs.Op = "cmd/LoadEnv"

	var f ConfigFile
	f, err = NewConfigFile(env)
	if err != nil {
		return errs.E(op, err)
	}

	err = overrideEnv(f)
	if err != nil {
		return errs.E(op, err)
	}
	return nil
}

// overrideEnv sets the environment
func overrideEnv(f ConfigFile) error {
	const op errs.Op = "cmd/overrideEnv"

	var err error

	// minimum accepted log level
	err = os.Setenv(logLevelMinEnv, f.Config.Logger.MinLogLevel)
	if err != nil {
		return errs.E(op, err)
	}

	// log level
	err = os.Setenv(loglevelEnv, f.Config.Logger.LogLevel)
	if err != nil {
		return errs.E(op, err)
	}

	// log error stack
	err = os.Setenv(logErrorStackEnv, fmt.Sprintf("%t", f.Config.Logger.LogErrorStack))
	if err != nil {
		return errs.E(op, err)
	}

	// server port
	err = os.Setenv(portEnv, strconv.Itoa(f.Config.HTTPServer.ListenPort))
	if err != nil {
		return errs.E(op, err)
	}

	// database host
	err = os.Setenv(sqldb.DBHostEnv, f.Config.Database.Host)
	if err != nil {
		return errs.E(op, err)
	}

	// database port
	err = os.Setenv(sqldb.DBPortEnv, strconv.Itoa(f.Config.Database.Port))
	if err != nil {
		return errs.E(op, err)
	}

	// database name
	err = os.Setenv(sqldb.DBNameEnv, f.Config.Database.Name)
	if err != nil {
		return errs.E(op, err)
	}

	// database user
	err = os.Setenv(sqldb.DBUserEnv, f.Config.Database.User)
	if err != nil {
		return errs.E(op, err)
	}

	// database user password
	err = os.Setenv(sqldb.DBPasswordEnv, f.Config.Database.Password)
	if err != nil {
		return errs.E(op, err)
	}

	// database search path
	err = os.Setenv(sqldb.DBSearchPathEnv, f.Config.Database.SearchPath)
	if err != nil {
		return errs.E(op, err)
	}

	// encryption key
	err = os.Setenv(encryptKeyEnv, f.Config.EncryptionKey)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

// NewConfigFile initializes a ConfigFile struct from a JSON file at a
// predetermined file path for each environment (paths are relative to project root)
//
// Production: ./config/production.json
//
// Staging:    ./config/staging.json
//
// Local:      ./config/local.json
func NewConfigFile(env Env) (ConfigFile, error) {
	const op errs.Op = "cmd/NewConfigFile"

	var (
		b   []byte
		err error
	)
	switch env {
	case Existing:
		return ConfigFile{}, nil
	case Local:
		b, err = os.ReadFile(localJSONConfigFile)
		if err != nil {
			return ConfigFile{}, errs.E(op, err)
		}
	case Staging:
		b, err = os.ReadFile(stagingJSONConfigFile)
		if err != nil {
			return ConfigFile{}, errs.E(op, err)
		}
	case Production:
		b, err = os.ReadFile(productionJSONConfigFile)
		if err != nil {
			return ConfigFile{}, errs.E(op, err)
		}
	default:
		return ConfigFile{}, errs.E(op, "Invalid environment")
	}

	f := ConfigFile{}
	err = json.Unmarshal(b, &f)
	if err != nil {
		return ConfigFile{}, errs.E(op, err)
	}

	return f, nil
}

// ConfigCueFilePaths defines the paths for config files processed through CUE.
type ConfigCueFilePaths struct {
	// Input defines the list of paths for files to be taken as input for CUE
	Input []string
	// Output defines the path for the JSON output of CUE
	Output string
}

// CUEPaths returns the ConfigCueFilePaths given the environment.
// Paths are relative to the project root.
func CUEPaths(env Env) (ConfigCueFilePaths, error) {
	const (
		schemaInput          = "./config/cue/schema.cue"
		localInput           = "./config/cue/local.cue"
		stagingInput         = "./config/cue/staging.cue"
		prodInput            = "./config/cue/production.cue"
		op           errs.Op = "cmd/CUEPaths"
	)

	switch env {
	case Local:
		return ConfigCueFilePaths{
			Input:  []string{schemaInput, localInput},
			Output: localJSONConfigFile,
		}, nil
	case Staging:
		return ConfigCueFilePaths{
			Input:  []string{schemaInput, stagingInput},
			Output: stagingJSONConfigFile,
		}, nil
	case Production:
		return ConfigCueFilePaths{
			Input:  []string{schemaInput, prodInput},
			Output: productionJSONConfigFile,
		}, nil
	default:
		return ConfigCueFilePaths{}, errs.E(op, fmt.Sprintf("There is no path configuration for the %s environment", env))
	}
}

// CUEGenesisPaths returns the ConfigCueFilePaths for the Genesis config.
// Paths are relative to the project root.
func CUEGenesisPaths() ConfigCueFilePaths {
	const (
		schemaInput = "./config/genesis/cue/schema.cue"
		authInput   = "./config/genesis/cue/auth.cue"
		userInput   = "./config/genesis/cue/input.cue"
	)

	return ConfigCueFilePaths{
		Input:  []string{schemaInput, authInput, userInput},
		Output: genesisRequestFile,
	}
}
