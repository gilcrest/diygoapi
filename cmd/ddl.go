package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gilcrest/diygoapi/errs"
)

// ddlFile represents a Data Definition Language (DDL) file
// Given the file naming convention 001-user.sql, the numbers up to
// the first dash are extracted, converted to an int and added to the
// fileNumber field to make the struct sortable using the sort package.
type ddlFile struct {
	filename   string
	fileNumber int
}

// newDDLFile initializes a DDLFile struct. File naming convention
// should be 001-user.sql where 001 represents the file number order
// to be processed
func newDDLFile(f string) (ddlFile, error) {
	const op errs.Op = "cmd/newDDLFile"

	i := strings.Index(f, "-")
	fileNumber := f[:i]
	fn, err := strconv.Atoi(fileNumber)
	if err != nil {
		return ddlFile{}, errs.E(op, err)
	}

	return ddlFile{filename: f, fileNumber: fn}, nil
}

func (df ddlFile) String() string {
	return fmt.Sprintf("%s: %d", df.filename, df.fileNumber)
}

// readDDLFiles reads and returns sorted DDL files from the
// up or down directory
func readDDLFiles(dir string) ([]ddlFile, error) {
	const op errs.Op = "cmd/readDDLFiles"

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, errs.E(op, err)
	}

	var ddlFiles []ddlFile
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		var df ddlFile
		df, err = newDDLFile(file.Name())
		if err != nil {
			return nil, errs.E(op, err)
		}
		ddlFiles = append(ddlFiles, df)
	}

	sort.Sort(byFileNumber(ddlFiles))

	return ddlFiles, nil
}

// byFileNumber implements sort.Interface for []ddlFile based on
// the fileNumber field.
type byFileNumber []ddlFile

// Len returns the length of elements in the ByFileNumber slice for sorting
func (bfn byFileNumber) Len() int { return len(bfn) }

// Swap sets up the elements to be swapped for the ByFileNumber slice for sorting
func (bfn byFileNumber) Swap(i, j int) { bfn[i], bfn[j] = bfn[j], bfn[i] }

// Less is the sorting logic for the ByFileNumber slice
func (bfn byFileNumber) Less(i, j int) bool { return bfn[i].fileNumber < bfn[j].fileNumber }

// PSQLArgs takes a slice of DDL files to be executed and builds a
// sequence of command line arguments using the appropriate flags
// psql needs to execute files. The arguments returned for psql are as follows:
//
// -w flag is set to never prompt for a password as we are running this as a script
//
// -d flag sets the database connection using a Connection URI string.
//
// -f flag is sent before each file to tell it to process the file
func PSQLArgs(up bool) ([]string, error) {
	const op errs.Op = "cmd/PSQLArgs"

	dir := "./scripts/db/migrations"
	if up {
		dir += "/up"
	} else {
		dir += "/down"
	}

	// readDDLFiles reads and returns sorted DDL files from the up or down directory
	ddlFiles, err := readDDLFiles(dir)
	if err != nil {
		return nil, errs.E(op, err)
	}

	if len(ddlFiles) == 0 {
		return nil, errs.E(op, fmt.Sprintf("there are no DDL files to process in %s", dir))
	}

	// newFlags will retrieve the database info from the environment using ff
	flgs, err := newFlags([]string{"server"})
	if err != nil {
		return nil, errs.E(op, err)
	}

	// command line args for psql are constructed
	args := []string{"-w", "-d", newPostgreSQLDSN(flgs).ConnectionURI(), "-c", "select current_database(), current_user, version()"}

	for _, file := range ddlFiles {
		args = append(args, "-f")
		args = append(args, dir+"/"+file.filename)
	}

	return args, nil
}
