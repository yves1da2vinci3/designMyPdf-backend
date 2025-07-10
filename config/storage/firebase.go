package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func InitializeFirebaseStorage() (*storage.BucketHandle, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	config := &firebase.Config{
		StorageBucket: fmt.Sprintf("%s.appspot.com", os.Getenv("BUCKET_NAME")),
	}
	opt := option.WithCredentialsFile(os.Getenv("FIREBASE_CONFIG_PATH"))
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
		return nil, err
	}

	client, err := app.Storage(context.Background())
	if err != nil {
		log.Fatalf("error getting Storage client: %v\n", err)
		return nil, err
	}

	bucketName := fmt.Sprintf("%s.appspot.com", os.Getenv("BUCKET_NAME"))
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		log.Fatalf("error getting bucket: %v\n", err)
		return nil, err
	}
	return bucket, nil
}

func UploadFile(bucket *storage.BucketHandle, localFilePath, storagePath string) (string, error) {
	ctx := context.Background()
	file, err := os.Open(localFilePath)
	if err != nil {
		return "", fmt.Errorf("error opening local file: %v", err)
	}
	defer file.Close()

	obj := bucket.Object(storagePath)
	wc := obj.NewWriter(ctx)
	if _, err = io.Copy(wc, file); err != nil {
		return "", fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("wc.Close: %v", err)
	}

	storageURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucket.BucketName(), storagePath)
	fmt.Printf("File uploaded successfully to Firebase Storage at %s\n", storageURL)
	return storageURL, nil
}

func listFiles(bucket *storage.BucketHandle) {
	ctx := context.Background()
	it := bucket.Objects(ctx, nil)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("error listing objects: %v\n", err)
		}
		fmt.Println(objAttrs.Name)
	}
}
