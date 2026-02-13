package cmd

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterbourgon/ff/v3"

	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/sqldb"
)

const (
	configFileFlagName        = "config-file"
	configFileFlagNameDefault = "./config.json"
	configFileFlagNameEnvVar  = "CONFIG_FILE"

	targetFlagName       = "target"
	targetFlagDefault    = "dev"
	targetFlagEnvVarName = "TARGET"

	logLevelMinFlagName       = "log-level-min"
	logLevelMinFlagDefault    = "trace"
	logLevelMinFlagEnvVarName = "LOG_LEVEL_MIN"

	loglevelFlagName       = "log-level"
	loglevelFlagDefault    = "info"
	loglevelFlagEnvVarName = "LOG_LEVEL"

	logErrorStackFlagName       = "log-error-stack"
	logErrorStackFlagDefault    = false
	logErrorStackFlagEnvVarName = "LOG_ERROR_STACK"

	listenPortFlagName       = "port"
	listenPorFlagDefault     = 8080
	listenPortFlagEnvVarName = "PORT"

	dbHostFlagName    = "db-host"
	dbHostFlagDefault = "localhost"

	dbPortFlagName    = "db-port"
	dbPortFlagDefault = 5432

	dbNameFlagName    = "db-name"
	dbNameFlagDefault = ""

	dbUserFlagName    = "db-user"
	dbUserFlagDefault = ""

	dbPasswordFlagName    = "db-password"
	dbPasswordFlagDefault = ""

	dbSearchPathFlagName    = "db-search-path"
	dbSearchPathFlagDefault = ""

	encryptKeyFlagName       = "encrypt-key"
	encryptKeyFlagDefault    = ""
	encryptKeyFlagEnvVarName = "ENCRYPT_KEY"
)

type flags struct {

	// target is the deployment target name, e.g. dev, test, prod
	target string

	// log-level flag allows for setting logging level, e.g. to run the server
	// with level set to debug, it'd be: ./server -log-level=debug
	// If not set, defaults to error
	loglvl string

	// log-level-min flag sets the minimum accepted logging level
	// - e.g. in production, you may have a policy to never allow logs at
	// trace level. You could set the minimum log level to Debug. Even
	// if the Global log level is set to Trace, only logs at Debug
	// and above would be logged. Default level is trace.
	logLvlMin string

	// logErrorStack flag determines whether a full error stack
	// should be logged. If true, error stacks are logged, if false,
	// just the error is logged
	logErrorStack bool

	// port flag is what http.ListenAndServe will listen on. default is 8080 if not set
	port int

	// dbhost is the database host
	dbhost string

	// dbport is the database port
	dbport int

	// dbname is the database name
	dbname string

	// dbuser is the database user
	dbuser string

	// dbpassword is the database user's password
	dbpassword string

	// dbsearchpath is the database search path
	dbsearchpath string

	// encryptkey is the encryption key
	encryptkey string
}

func (f *flags) Validate() error {
	const op errs.Op = "cmd/flags.Validate"

	// validate target is not empty
	if f.target == "" {
		return errs.E(op, "target is required")
	}

	// validate port in acceptable range
	err := portRange(f.port)
	if err != nil {
		return err
	}

	// validate encryption key is not empty
	if f.encryptkey == "" {
		return errs.E(op, "encryption key is required")
	}

	// validate log level is not empty
	if f.loglvl == "" {
		return errs.E(op, "log level is required")
	}

	// validate minimum log level is not empty
	if f.logLvlMin == "" {
		return errs.E(op, "minimum log level is required")
	}

	// validate database host is not empty
	if f.dbhost == "" {
		return errs.E(op, "database host is required")
	}

	// validate database name is not empty
	if f.dbname == "" {
		return errs.E(op, "database name is required")
	}

	// validate database user is not empty
	if f.dbuser == "" {
		return errs.E(op, "database user is required")
	}

	return nil
}

// newFlags parses the command line flags using ff and returns
// a flags struct or an error
func newFlags(args []string) (flags, error) {
	const op errs.Op = "cmd/newFlags"
	// create new FlagSet using the program name being executed (args[0])
	// as the name of the FlagSet
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	var (
		target        = fs.String(targetFlagName, targetFlagDefault, fmt.Sprintf("target to run (also via %s)", targetFlagEnvVarName))
		logLvlMin     = fs.String(logLevelMinFlagName, logLevelMinFlagDefault, fmt.Sprintf("sets minimum log level (trace, debug, info, warn, error, fatal, panic, disabled), (also via %s)", logLevelMinFlagEnvVarName))
		loglvl        = fs.String(loglevelFlagName, loglevelFlagDefault, fmt.Sprintf("sets log level (trace, debug, info, warn, error, fatal, panic, disabled), (also via %s)", loglevelFlagEnvVarName))
		logErrorStack = fs.Bool(logErrorStackFlagName, logErrorStackFlagDefault, fmt.Sprintf("if true, log full error stacktrace using github.com/pkg/errors, else just log error, (also via %s)", logErrorStackFlagEnvVarName))
		port          = fs.Int(listenPortFlagName, listenPorFlagDefault, fmt.Sprintf("listen port for server (also via %s)", listenPortFlagEnvVarName))
		dbhost        = fs.String(dbHostFlagName, dbHostFlagDefault, fmt.Sprintf("postgresql database host (also via %s)", sqldb.DBHostEnv))
		dbport        = fs.Int(dbPortFlagName, dbPortFlagDefault, fmt.Sprintf("postgresql database port (also via %s)", sqldb.DBPortEnv))
		dbname        = fs.String(dbNameFlagName, dbNameFlagDefault, fmt.Sprintf("postgresql database name (also via %s)", sqldb.DBNameEnv))
		dbuser        = fs.String(dbUserFlagName, dbUserFlagDefault, fmt.Sprintf("postgresql database user (also via %s)", sqldb.DBUserEnv))
		dbpassword    = fs.String(dbPasswordFlagName, dbPasswordFlagDefault, fmt.Sprintf("postgresql database password (also via %s)", sqldb.DBPasswordEnv))
		dbsearchpath  = fs.String(dbSearchPathFlagName, dbSearchPathFlagDefault, fmt.Sprintf("postgresql database search path (also via %s)", sqldb.DBSearchPathEnv))
		encryptkey    = fs.String(encryptKeyFlagName, encryptKeyFlagDefault, fmt.Sprintf("encryption key (also via %s)", encryptKeyFlagEnvVarName))
		configFile    = fs.String(configFileFlagName, configFileFlagNameDefault, fmt.Sprintf("JSON configuration file (also via %s)", configFileFlagNameEnvVar))
	)

	// Parse the command line flags from above
	err := ff.Parse(fs, args[1:],
		ff.WithEnvVars(),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser))
	if err != nil {
		return flags{}, errs.E(op, err)
	}

	return flags{
		target:        *target,
		loglvl:        *loglvl,
		logLvlMin:     *logLvlMin,
		logErrorStack: *logErrorStack,
		port:          *port,
		dbhost:        *dbhost,
		dbport:        *dbport,
		dbname:        *dbname,
		dbuser:        *dbuser,
		dbpassword:    *dbpassword,
		dbsearchpath:  *dbsearchpath,
		encryptkey:    *encryptkey,
	}, nil
}

type ConfigFile struct {
	DefaultTarget string `json:"default_target"`
	Targets       []struct {
		Target             string `json:"target"`
		ServerListenerPort int    `json:"server_listener_port"`
		Logger             struct {
			MinLogLevel   string `json:"min_log_level"`
			LogLevel      string `json:"log_level"`
			LogErrorStack bool   `json:"log_error_stack"`
		} `json:"logger"`
		EncryptionKey string `json:"encryption_key"`
		Database      struct {
			Host       string `json:"host"`
			Port       int    `json:"port"`
			Name       string `json:"name"`
			User       string `json:"user"`
			Password   string `json:"password"`
			SearchPath string `json:"search_path"`
		} `json:"database"`
	} `json:"targets"`
}

// NewConfigFile initializes a ConfigFile struct from a JSON file which
// must be located in the same directory as the executable. The executable
// path is determined by os.Executable() and returned as a string.
func NewConfigFile() (ConfigFile, string, error) {
	const op errs.Op = "cmd/NewConfigFile"

	var (
		b   []byte
		s   string
		err error
	)
	s, err = os.Executable()
	if err != nil {
		return ConfigFile{}, "", errs.E(op, err)
	}

	p := filepath.Dir(s)
	p = filepath.Join(p, configFilename)

	b, err = os.ReadFile(p)
	if err != nil {
		return ConfigFile{}, "", errs.E(op, err)
	}

	f := ConfigFile{}
	err = json.Unmarshal(b, &f)
	if err != nil {
		return ConfigFile{}, "", errs.E(op, err)
	}

	return f, p, nil
}

// ConfigParser is a parser for config files in a specifig JSON format.
func ConfigParser(r io.Reader, set func(name, value string) error) error {

	// Decode io.Reader into a Decoder type and unmarshal that into the
	// CreateMovieRequest struct (rb)
	err := json.NewDecoder(r).Decode(&rb)

	s := bufio.NewScanner(r)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue // skip empties
		}

		if line[0] == '#' {
			continue // skip comments
		}

		var (
			name  string
			value string
			index = strings.IndexRune(line, ' ')
		)
		if index < 0 {
			name, value = line, "true" // boolean option
		} else {
			name, value = line[:index], strings.TrimSpace(line[index:])
		}

		if i := strings.Index(value, " #"); i >= 0 {
			value = strings.TrimSpace(value[:i])
		}

		if err := set(name, value); err != nil {
			return err
		}
	}
	return nil
}
