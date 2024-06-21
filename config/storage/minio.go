package storage

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// InitializeMinio initializes the Minio storage service
func InitializeMinio(config map[string]string) (*minio.Client, error) {
	endpoint := config["STORAGE_ENDPOINT"]
	accessKeyID := config["STORAGE_ACCESS_KEY"]
	secretAccessKey := config["STORAGE_SECRET_KEY"]
	useSSL := config["STORAGE_USE_SSL"] == "true"

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return minioClient, nil
}
