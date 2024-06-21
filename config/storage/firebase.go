package storage

import (
	"context"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// InitializeFirebase initializes the Firebase storage service
func InitializeFirebase(config map[string]string) (*firebase.App, error) {
	configPath := config["FIREBASE_CONFIG_PATH"]
	opt := option.WithCredentialsFile(configPath)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, err
	}

	return app, nil
}
