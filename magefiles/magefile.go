package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"

	"github.com/gilcrest/go-api-basic/command"
)

// DBUp uses the psql command line interface to execute DDL scripts
// in the up directory and create all required DB objects. All files
// will be executed, regardless of errors within an individual file.
// Check output to determine if any errors occurred.
// Eventually, I will write this to stop on errors, but for now it is
// what it is.
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

// DBDown uses the psql command line interface to execute DDL scripts
// in the down directory and drops all project-specific DB objects.
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

// TestAll runs all tests for the app
func TestAll(env string) (err error) {
	err = command.LoadEnv(command.ParseEnv(env))
	if err != nil {
		return err
	}

	err = sh.Run("go", "test", "-v", "./...")
	if err != nil {
		return err
	}

	return nil
}

// Run runs program using the given environment configuration
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

// Genesis runs all tests including executing the Genesis service
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

// NewKey generates a new encryption key
func NewKey() {
	command.NewEncryptionKey()
}

// GCP deploys the app to Google Cloud Run
func GCP(env string) error {

	f, err := command.NewConfigFile(command.ParseEnv(env))
	if err != nil {
		return err
	}

	image := command.GCPArtifactRegistryContainerImage{
		ProjectID:          f.Config.GCP.ProjectID,
		RepositoryLocation: f.Config.GCP.ArtifactoryRegistry.RepoLocation,
		RepositoryName:     f.Config.GCP.ArtifactoryRegistry.RepoName,
		ImageName:          f.Config.GCP.ArtifactoryRegistry.ImageID,
		ImageTag:           f.Config.GCP.ArtifactoryRegistry.Tag,
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

// GenConfig generates configuration files for the given environment.
// The files are run through cue vet first to ensure they are acceptable
// given the schema.
//
// Acceptable environment values are: local, staging, production
func GenConfig(env string) (err error) {

	var paths command.ConfigCueFilePaths
	paths, err = command.CUEPaths(command.ParseEnv(env))
	if err != nil {
		return err
	}

	// Vet input files
	vetArgs := []string{"vet"}
	for _, path := range paths.Input {
		vetArgs = append(vetArgs, path)
	}
	err = sh.Run("cue", vetArgs...)
	if err != nil {
		return err
	}

	// Export output files
	exportArgs := []string{"export"}
	for _, path := range paths.Input {
		exportArgs = append(exportArgs, path)
	}
	exportArgs = append(exportArgs, "--force", "--out", "json", "--outfile", paths.Output)

	err = sh.Run("cue", exportArgs...)
	if err != nil {
		return err
	}

	return nil
}

// StartGCPDB starts the GCP Cloud SQL database for the environment/config given
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

// StopGCPDB stops the GCP Cloud SQL database for the environment/config given
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
