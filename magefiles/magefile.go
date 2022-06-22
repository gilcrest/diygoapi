package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"

	"github.com/gilcrest/diy-go-api/command"
)

// NewKey generates a new encryption key,
// example: mage -v newkey
func NewKey() {
	command.NewEncryptionKey()
}

// CueGenerateConfig generates configuration files for the given environment,
// example: mage -v cueGenerateConfig local.
// The files are run through cue vet to ensure they are acceptable given
// the schema and are then run through cue "fmt" to format the files
//
// Acceptable environment values are: local, staging, production
func CueGenerateConfig(env string) (err error) {

	var paths command.ConfigCueFilePaths
	paths, err = command.CUEPaths(command.ParseEnv(env))
	if err != nil {
		return err
	}

	// Vet input files
	vetArgs := []string{"vet"}
	vetArgs = append(vetArgs, paths.Input...)
	err = sh.Run("cue", vetArgs...)
	if err != nil {
		return err
	}

	// format input files
	fmtArgs := []string{"fmt"}
	fmtArgs = append(fmtArgs, paths.Input...)
	err = sh.Run("cue", fmtArgs...)
	if err != nil {
		return err
	}

	// Export output files
	exportArgs := []string{"export"}
	exportArgs = append(exportArgs, paths.Input...)
	exportArgs = append(exportArgs, "--force", "--out", "json", "--outfile", paths.Output)

	err = sh.Run("cue", exportArgs...)
	if err != nil {
		return err
	}

	return nil
}

// CueGenerateGenesisConfig generates the Genesis configuration file,
// example: mage -v cueGenerateGenesisConfig.
// The files are run through cue vet to ensure they are acceptable given
// the schema and are then run through cue "fmt" to format the files
func CueGenerateGenesisConfig() (err error) {

	paths := command.CUEGenesisPaths()

	// Vet input files
	vetArgs := []string{"vet"}
	vetArgs = append(vetArgs, paths.Input...)
	err = sh.Run("cue", vetArgs...)
	if err != nil {
		return err
	}

	// format input files
	fmtArgs := []string{"fmt"}
	fmtArgs = append(fmtArgs, paths.Input...)
	err = sh.Run("cue", fmtArgs...)
	if err != nil {
		return err
	}

	// Export output files
	exportArgs := []string{"export"}
	exportArgs = append(exportArgs, paths.Input...)
	exportArgs = append(exportArgs, "--force", "--out", "json", "--outfile", paths.Output)

	err = sh.Run("cue", exportArgs...)
	if err != nil {
		return err
	}

	return nil
}

// DBUp uses the psql cli to execute DDL scripts in the up directory
// and creates all required DB objects,
// example: mage -v dbup local.
// All files will be executed, regardless of errors within an individual
// file. Check output to determine if any errors occurred. Eventually,
// I will write this to stop on errors, but for now it is what it is.
func DBUp(env string) (err error) {
	var args []string

	err = command.LoadEnv(command.ParseEnv(env))
	if err != nil {
		return err
	}

	args, err = command.PSQLArgs(true)
	if err != nil {
		return err
	}

	err = sh.Run("psql", args...)
	if err != nil {
		return err
	}

	return nil
}

// DBDown uses the psql cli to execute DDL scripts
// in the down directory and drops all project-specific DB objects,
// example: mage -v dbdown local.
// All files will be executed, regardless of errors within
// an individual file. Check output to determine if any errors occurred.
// Eventually, I will write this to stop on errors, but for now it is
// what it is.
func DBDown(env string) (err error) {
	var args []string

	err = command.LoadEnv(command.ParseEnv(env))
	if err != nil {
		return err
	}

	args, err = command.PSQLArgs(false)
	if err != nil {
		return err
	}

	err = sh.Run("psql", args...)
	if err != nil {
		return err
	}

	return nil
}

// Genesis runs all tests including executing the Genesis service,
// example: mage -v genesis local
func Genesis(env string) (err error) {
	err = command.LoadEnv(command.ParseEnv(env))
	if err != nil {
		return err
	}

	err = command.Genesis()
	if err != nil {
		return err
	}

	return nil
}

// TestAll runs all tests for the app,
// example: mage -v testall false local.
// If verbose is passed, tests will be run in verbose mode.
func TestAll(verbose bool, env string) (err error) {
	err = command.LoadEnv(command.ParseEnv(env))
	if err != nil {
		return err
	}

	args := []string{"test"}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, "./...")

	err = sh.Run("go", args...)
	if err != nil {
		return err
	}

	return nil
}

// Run runs program using the given environment configuration,
// example: mage -v run local
func Run(env string) (err error) {
	err = command.LoadEnv(command.ParseEnv(env))
	if err != nil {
		return err
	}

	err = sh.Run("go", "run", "main.go")
	if err != nil {
		return err
	}

	return nil
}

// GCP builds the app as a Docker container image to GCP Artifact Registry
// and then deploys it to Google Cloud Run, example: mage -v gcp staging
func GCP(env string) error {

	f, err := command.NewConfigFile(command.ParseEnv(env))
	if err != nil {
		return err
	}

	image := command.GCPArtifactRegistryContainerImage{
		ProjectID:          f.Config.GCP.ProjectID,
		RepositoryLocation: f.Config.GCP.ArtifactRegistry.RepoLocation,
		RepositoryName:     f.Config.GCP.ArtifactRegistry.RepoName,
		ImageName:          f.Config.GCP.ArtifactRegistry.ImageID,
		ImageTag:           f.Config.GCP.ArtifactRegistry.Tag,
	}

	err = gcpArtifactRegistryBuild(image)
	if err != nil {
		return err
	}

	args := command.GCPCloudRunDeployImage(f, image)
	if err != nil {
		return err
	}

	err = sh.Run("gcloud", args...)
	if err != nil {
		return err
	}

	return nil
}

func gcpArtifactRegistryBuild(image command.GCPArtifactRegistryContainerImage) error {
	const (
		dockerfileOrigin      = "./magefiles/Dockerfile"
		dockerfileDestination = "Dockerfile"
	)
	var err error

	// move the Dockerfile to the project root directory
	err = os.Rename(dockerfileOrigin, dockerfileDestination)
	if err != nil {
		return err
	}
	var cwd string
	cwd, err = os.Getwd()
	if err != nil {
		return err
	}
	fmt.Printf("Dockerfile moved from %s to %s\n", dockerfileOrigin, cwd)

	// defer moving the Dockerfile back
	defer func() {
		deferErr := os.Rename(dockerfileDestination, dockerfileOrigin)
		if deferErr != nil {
			if err != nil {
				err = errors.Wrap(err, deferErr.Error())
				return
			}
			err = deferErr
			return
		}
		fmt.Printf("Dockerfile moved back to %s\n", dockerfileOrigin)
	}()

	// args for gcloud
	args := []string{"builds", "submit", "--tag", image.String()}

	err = sh.Run("gcloud", args...)
	if err != nil {
		return err
	}

	return nil
}

// StartGCPDB starts the GCP Cloud SQL database for the environment/config given,
// example: mage -v startgcpdb staging
func StartGCPDB(env string) (err error) {
	var f command.ConfigFile
	f, err = command.NewConfigFile(command.ParseEnv(env))
	if err != nil {
		return err
	}

	args := []string{"sql", "instances", "patch", f.Config.GCP.CloudSQL.InstanceName, "--activation-policy=ALWAYS"}

	err = sh.Run("gcloud", args...)
	if err != nil {
		return err
	}

	return nil
}

// StopGCPDB stops the GCP Cloud SQL database for the environment/config given,
// example: mage -v stopgcpdb staging
func StopGCPDB(env string) (err error) {
	var f command.ConfigFile
	f, err = command.NewConfigFile(command.ParseEnv(env))
	if err != nil {
		return err
	}

	args := []string{"sql", "instances", "patch", f.Config.GCP.CloudSQL.InstanceName, "--activation-policy=NEVER"}

	err = sh.Run("gcloud", args...)
	if err != nil {
		return err
	}

	return nil
}
