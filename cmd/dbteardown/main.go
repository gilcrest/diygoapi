package main

import (
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

// configFile mirrors the JSON config structure for the fields db-teardown needs.
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

	fs := flag.NewFlagSet("dbteardown", flag.ContinueOnError)
	configPath := fs.String("config-file", "./config/config.json", "path to JSON configuration file")
	dbAdminConfigTarget := fs.String("db-admin-config-target", "", "admin target name for the psql connection (required)")
	appConfigTarget := fs.String("app-config-target", "", "target whose DB values (user, database, schema) will be dropped; defaults to default_target from config")

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

	// Step 1: Drop schema (must connect to the app database)
	appDBDSN := sqldb.PostgreSQLDSN{
		Host:   admin.Host,
		Port:   admin.Port,
		DBName: app.Name,
		User:   admin.User,
	}
	if err := dropSchema(appDBDSN, admin.Password, app); err != nil {
		return errs.E(op, err)
	}

	// Step 2: Drop database (connect to admin database)
	if err := dropDatabase(adminDSN, admin.Password, app); err != nil {
		return errs.E(op, err)
	}

	// Step 3: Drop user (connect to admin database)
	if err := dropUser(adminDSN, admin.Password, app); err != nil {
		return errs.E(op, err)
	}

	fmt.Println("\ndb-teardown completed successfully")
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

func dropSchema(appDBDSN sqldb.PostgreSQLDSN, adminPassword string, app targetDB) error {
	const op errs.Op = "main/dropSchema"

	drop := psqlCmd(appDBDSN, adminPassword, "-c", fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", app.SearchPath))
	drop.Stdout = os.Stdout
	if err := drop.Run(); err != nil {
		return errs.E(op, err)
	}

	fmt.Printf("schema %q dropped\n", app.SearchPath)
	return nil
}

func dropDatabase(adminDSN sqldb.PostgreSQLDSN, adminPassword string, app targetDB) error {
	const op errs.Op = "main/dropDatabase"

	drop := psqlCmd(adminDSN, adminPassword, "-c", fmt.Sprintf("DROP DATABASE IF EXISTS %s", app.Name))
	drop.Stdout = os.Stdout
	if err := drop.Run(); err != nil {
		return errs.E(op, err)
	}

	fmt.Printf("database %q dropped\n", app.Name)
	return nil
}

func dropUser(adminDSN sqldb.PostgreSQLDSN, adminPassword string, app targetDB) error {
	const op errs.Op = "main/dropUser"

	drop := psqlCmd(adminDSN, adminPassword, "-c", fmt.Sprintf("DROP USER IF EXISTS %s", app.User))
	drop.Stdout = os.Stdout
	if err := drop.Run(); err != nil {
		return errs.E(op, err)
	}

	fmt.Printf("user %q dropped\n", app.User)
	return nil
}
