package storage

import (
	"fmt"
)

// InitializeStorage initializes the storage service based on the provided type
func InitializeStorage(storageType string, config map[string]string) (interface{}, error) {
	switch storageType {
	case "Minio":
		return InitializeMinio(config)
	case "Firebase":
		return InitializeFirebase(config)
	case "GoogleDrive":
		// Add Google Drive initialization here
		return nil, fmt.Errorf("Google Drive initialization not implemented")
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}
