package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/sqldb"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

// configFile mirrors the JSON config structure for the fields db-init needs.
type configFile struct {
	DefaultTarget string `json:"default_target"`
	Targets       []struct {
		Target   string `json:"target"`
		Database struct {
			Host       string `json:"host"`
			Port       int    `json:"port"`
			Name       string `json:"name"`
			User       string `json:"user"`
			Password   string `json:"password"`
			SearchPath string `json:"search_path"`
		} `json:"database"`
	} `json:"targets"`
}

// targetDB holds the database config for a single target.
type targetDB struct {
	Host       string
	Port       int
	Name       string
	User       string
	Password   string
	SearchPath string
}

func findTarget(cf configFile, name string) (targetDB, error) {
	const op errs.Op = "main/findTarget"

	for _, t := range cf.Targets {
		if t.Target == name {
			return targetDB{
				Host:       t.Database.Host,
				Port:       t.Database.Port,
				Name:       t.Database.Name,
				User:       t.Database.User,
				Password:   t.Database.Password,
				SearchPath: t.Database.SearchPath,
			}, nil
		}
	}
	return targetDB{}, errs.E(op, fmt.Sprintf("target %q not found in config file", name))
}

func run() error {
	const op errs.Op = "main/run"

	fs := flag.NewFlagSet("dbinit", flag.ContinueOnError)
	configPath := fs.String("config-file", "./config/config.json", "path to JSON configuration file")
	dbAdminConfigTarget := fs.String("db-admin-config-target", "", "admin target name for the psql connection (required)")
	appConfigTarget := fs.String("app-config-target", "", "target whose DB values (user, password, db name, search path) will be created; defaults to default_target from config")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return errs.E(op, err)
	}

	if *dbAdminConfigTarget == "" {
		return errs.E(op, "--db-admin-config-target flag is required (admin target for psql connection)")
	}

	// Read and decode config file
	f, err := os.Open(*configPath)
	if err != nil {
		return errs.E(op, err)
	}
	defer f.Close()

	var cf configFile
	if err := json.NewDecoder(f).Decode(&cf); err != nil {
		return errs.E(op, err)
	}

	// Look up admin target
	admin, err := findTarget(cf, *dbAdminConfigTarget)
	if err != nil {
		return errs.E(op, err)
	}

	// Look up app target
	appTargetName := *appConfigTarget
	if appTargetName == "" {
		appTargetName = cf.DefaultTarget
	}
	if appTargetName == "" {
		return errs.E(op, "no --app-config-target specified and no default_target in config")
	}

	app, err := findTarget(cf, appTargetName)
	if err != nil {
		return errs.E(op, err)
	}

	// Build admin DSN for psql connection
	adminDSN := sqldb.PostgreSQLDSN{
		Host:   admin.Host,
		Port:   admin.Port,
		DBName: admin.Name,
		User:   admin.User,
	}

	fmt.Printf("Admin connection: %s@%s:%d/%s\n", admin.User, admin.Host, admin.Port, admin.Name)
	fmt.Printf("App target: %s (user=%s, db=%s, schema=%s)\n\n", appTargetName, app.User, app.Name, app.SearchPath)

	// Step 1: Check/create user
	if err := createUser(adminDSN, admin.Password, app); err != nil {
		return errs.E(op, err)
	}

	// Step 2: Alter user
	if err := alterUser(adminDSN, admin.Password, app); err != nil {
		return errs.E(op, err)
	}

	// Step 3: Check/create database
	if err := createDatabase(adminDSN, admin.Password, app); err != nil {
		return errs.E(op, err)
	}

	// Step 4: Create schema (connect to the new database)
	appDBDSN := sqldb.PostgreSQLDSN{
		Host:   admin.Host,
		Port:   admin.Port,
		DBName: app.Name,
		User:   admin.User,
	}
	if err := createSchema(appDBDSN, admin.Password, app); err != nil {
		return errs.E(op, err)
	}

	fmt.Println("\ndb-init completed successfully")
	return nil
}

// psqlCmd creates an exec.Command for psql with the given DSN and SQL arguments.
// If adminPassword is non-empty, it is set via the PGPASSWORD environment variable.
func psqlCmd(dsn sqldb.PostgreSQLDSN, adminPassword string, sqlArgs ...string) *exec.Cmd {
	args := append([]string{"-w", "-d", dsn.ConnectionURI()}, sqlArgs...)
	c := exec.Command("psql", args...)
	c.Stderr = os.Stderr
	if adminPassword != "" {
		c.Env = append(os.Environ(), "PGPASSWORD="+adminPassword)
	}
	return c
}

func createUser(adminDSN sqldb.PostgreSQLDSN, adminPassword string, app targetDB) error {
	const op errs.Op = "main/createUser"

	var out bytes.Buffer
	check := psqlCmd(adminDSN, adminPassword, "-tAc", fmt.Sprintf("SELECT 1 FROM pg_roles WHERE rolname='%s'", app.User))
	check.Stdout = &out
	if err := check.Run(); err != nil {
		return errs.E(op, err)
	}

	if out.Len() > 0 {
		fmt.Printf("user %q already exists, skipping\n", app.User)
		return nil
	}

	create := psqlCmd(adminDSN, adminPassword, "-c", fmt.Sprintf("CREATE USER %s WITH CREATEDB PASSWORD '%s'", app.User, app.Password))
	create.Stdout = os.Stdout
	if err := create.Run(); err != nil {
		return errs.E(op, err)
	}

	fmt.Printf("user %q created\n", app.User)
	return nil
}

func alterUser(adminDSN sqldb.PostgreSQLDSN, adminPassword string, app targetDB) error {
	const op errs.Op = "main/alterUser"

	alter := psqlCmd(adminDSN, adminPassword, "-c", fmt.Sprintf("ALTER USER %s WITH NOSUPERUSER", app.User))
	alter.Stdout = os.Stdout
	if err := alter.Run(); err != nil {
		return errs.E(op, err)
	}

	fmt.Printf("user %q altered (NOSUPERUSER)\n", app.User)
	return nil
}

func createDatabase(adminDSN sqldb.PostgreSQLDSN, adminPassword string, app targetDB) error {
	const op errs.Op = "main/createDatabase"

	var out bytes.Buffer
	check := psqlCmd(adminDSN, adminPassword, "-tAc", fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname='%s'", app.Name))
	check.Stdout = &out
	if err := check.Run(); err != nil {
		return errs.E(op, err)
	}

	if out.Len() > 0 {
		fmt.Printf("database %q already exists, skipping\n", app.Name)
		return nil
	}

	create := psqlCmd(adminDSN, adminPassword, "-c", fmt.Sprintf("CREATE DATABASE %s WITH OWNER %s", app.Name, app.User))
	create.Stdout = os.Stdout
	if err := create.Run(); err != nil {
		return errs.E(op, err)
	}

	fmt.Printf("database %q created\n", app.Name)
	return nil
}

func createSchema(appDBDSN sqldb.PostgreSQLDSN, adminPassword string, app targetDB) error {
	const op errs.Op = "main/createSchema"

	schema := psqlCmd(appDBDSN, adminPassword, "-c", fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s AUTHORIZATION %s", app.SearchPath, app.User))
	schema.Stdout = os.Stdout
	if err := schema.Run(); err != nil {
		return errs.E(op, err)
	}

	fmt.Printf("schema %q ensured\n", app.SearchPath)
	return nil
}
