package leaderelection

import (
	"fmt"
	"os"

	"github.com/google/uuid"
)

func getInClusterIdentity() (string, error) {
	if v := os.Getenv("POD_NAME"); len(v) > 0 {
		return v, nil
	}
	id, err := os.Hostname()
	if err != nil {
		return "", err
	}
	id = id + "_" + uuid.New().String()
	return id, nil
}

const inClusterNamespacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

func getInClusterNamespace() (string, error) {
	if v := os.Getenv("POD_NS"); len(v) > 0 {
		return v, nil
	}
	// Check whether the namespace file exists.
	// If not, we are not running in cluster so can't guess the namespace.
	if _, err := os.Stat(inClusterNamespacePath); os.IsNotExist(err) {
		return "", fmt.Errorf("not running in-cluster, please specify LeaderElectionNamespace")
	} else if err != nil {
		return "", fmt.Errorf("error checking namespace file: %w", err)
	}
	// Load the namespace file and return its content
	namespace, err := os.ReadFile(inClusterNamespacePath)
	if err != nil {
		return "", fmt.Errorf("error reading namespace file: %w", err)
	}
	return string(namespace), nil
}
