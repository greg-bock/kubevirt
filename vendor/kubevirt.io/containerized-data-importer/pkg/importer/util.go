package importer

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"kubevirt.io/containerized-data-importer/pkg/common"
)

// ParseEnvVar provides a wrapper to attempt to fetch the specified env var
func ParseEnvVar(envVarName string, decode bool) (string, error) {
	value := os.Getenv(envVarName)
	if decode {
		v, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return "", errors.Errorf("error decoding environment variable %q", envVarName)
		}
		value = fmt.Sprintf("%s", v)
	}
	return value, nil
}

// ParseEndpoint parses the required endpoint and return the url struct.
func ParseEndpoint(endpt string) (*url.URL, error) {
	var err error
	if endpt == "" {
		endpt, err = ParseEnvVar(common.ImporterEndpoint, false)
		if err != nil {
			return nil, err
		}
		if endpt == "" {
			return nil, errors.Errorf("endpoint %q is missing or blank", common.ImporterEndpoint)
		}
	}
	return url.Parse(endpt)
}

// StreamDataToFile provides a function to stream the specified io.Reader to the specified local file
func StreamDataToFile(dataReader io.Reader, filePath string) error {
	// Attempt to create the file with name filePath.  If it exists, fail.
	outFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.ModePerm)
	defer outFile.Close()
	if err != nil {
		return errors.Wrapf(err, "could not open file %q", filePath)
	}
	glog.V(1).Infof("begin import...\n")
	if _, err = io.Copy(outFile, dataReader); err != nil {
		os.Remove(outFile.Name())
		return errors.Wrapf(err, "unable to write to file")
	}
	return nil
}
