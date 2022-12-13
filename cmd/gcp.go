package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gilcrest/diygoapi/sqldb"
)

// GCPCloudRunDeployImage builds arguments for running a service on
// Cloud Run given an Artifact Registry image.
func GCPCloudRunDeployImage(f ConfigFile, image GCPArtifactRegistryContainerImage) []string {

	var (
		// Google Cloud Run Service Name
		serviceName = f.Config.GCP.CloudRun.ServiceName
		// Google Cloud SQL Instance Name
		gcpCloudSQLInstanceConnectionName = f.Config.GCP.CloudSQL.InstanceConnectionName
		// postgresql database name
	)

	args := []string{"run", "deploy", serviceName, "--image", image.String(), "--platform", "managed", "--no-allow-unauthenticated"}

	args = append(args, "--add-cloudsql-instances", gcpCloudSQLInstanceConnectionName)

	icn := fmt.Sprintf(`INSTANCE-CONNECTION-NAME=%s`, gcpCloudSQLInstanceConnectionName)
	dbName := fmt.Sprintf(`%s=%s`, sqldb.DBNameEnv, f.Config.Database.Name)
	dbUser := fmt.Sprintf(`%s=%s`, sqldb.DBUserEnv, f.Config.Database.User)
	dbPassword := fmt.Sprintf(`%s=%s`, sqldb.DBPasswordEnv, f.Config.Database.Password)
	dbHost := fmt.Sprintf(`%s=%s`, sqldb.DBHostEnv, f.Config.Database.Host)
	dbPort := fmt.Sprintf(`%s=%s`, sqldb.DBPortEnv, strconv.Itoa(f.Config.Database.Port))
	dbSearchPath := fmt.Sprintf(`%s=%s`, sqldb.DBSearchPathEnv, f.Config.Database.SearchPath)
	encryptKey := fmt.Sprintf(`%s=%s`, encryptKeyEnv, f.Config.EncryptionKey)

	envVars := []string{icn, dbName, dbUser, dbPassword, dbHost, dbPort, dbSearchPath, encryptKey}

	args = append(args, "--set-env-vars", strings.Join(envVars, ","))

	return args
}

// GCPArtifactRegistryContainerImage defines a GCP Artifact Registry
// build image according to https://cloud.google.com/artifact-registry/docs/docker/names
// The String method prints the build string needed to build to
// Artifact Registry using gcloud as well as deploy it to Cloud Run.
type GCPArtifactRegistryContainerImage struct {
	ProjectID          string
	RepositoryLocation string
	RepositoryName     string
	ImageName          string
	ImageTag           string
}

// String outputs the Google Artifact Registry image name.
// LOCATION-docker.pkg.dev/PROJECT-ID/REPOSITORY/IMAGE:TAG
func (i GCPArtifactRegistryContainerImage) String() string {
	if i.ImageTag != "" {
		return fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s:%s", i.RepositoryLocation, i.ProjectID, i.RepositoryName, i.ImageName, i.ImageTag)
	}
	return fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s", i.RepositoryLocation, i.ProjectID, i.RepositoryName, i.ImageName)
}
