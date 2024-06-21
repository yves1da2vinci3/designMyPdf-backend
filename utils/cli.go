package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "config-generator",
	Short: "CLI to generate .env configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		var databaseType string
		var storageType string

		// Prompt for database type
		dbPrompt := &survey.Select{
			Message: "Choose a database:",
			Options: []string{"MongoDB", "MySQL", "PostgreSQL"},
		}
		survey.AskOne(dbPrompt, &databaseType)

		// Prompt for storage type
		storagePrompt := &survey.Select{
			Message: "Choose a storage service:",
			Options: []string{"Minio", "Firebase", "Google Drive"},
		}
		survey.AskOne(storagePrompt, &storageType)

		// Create .env file based on user input
		err := generateEnvFile(databaseType, storageType)
		if err != nil {
			log.Fatalf("Failed to generate .env file: %v", err)
		}
		fmt.Println(".env file generated successfully")
	},
}

func generateEnvFile(databaseType, storageType string) error {
	file, err := os.Create(".env")
	if err != nil {
		return err
	}
	defer file.Close()

	// Database configurations
	switch databaseType {
	case "MongoDB":
		file.WriteString("DB_TYPE=MongoDB\n")
		file.WriteString("DB_HOST=localhost\n")
		file.WriteString("DB_PORT=27017\n")
		file.WriteString("DB_USER=root\n")
		file.WriteString("DB_PASS=password\n")
	case "MySQL":
		file.WriteString("DB_TYPE=MySQL\n")
		file.WriteString("DB_HOST=localhost\n")
		file.WriteString("DB_PORT=3306\n")
		file.WriteString("DB_USER=root\n")
		file.WriteString("DB_PASS=password\n")
	case "PostgreSQL":
		file.WriteString("DB_TYPE=PostgreSQL\n")
		file.WriteString("DB_HOST=localhost\n")
		file.WriteString("DB_PORT=5432\n")
		file.WriteString("DB_USER=root\n")
		file.WriteString("DB_PASS=password\n")
	}

	// Storage configurations
	switch storageType {
	case "Minio":
		file.WriteString("STORAGE_TYPE=Minio\n")
		file.WriteString("STORAGE_ENDPOINT=http://localhost:9000\n")
		file.WriteString("STORAGE_ACCESS_KEY=minioadmin\n")
		file.WriteString("STORAGE_SECRET_KEY=minioadmin\n")
	case "Firebase":
		file.WriteString("STORAGE_TYPE=Firebase\n")
		file.WriteString("FIREBASE_CONFIG_PATH=/path/to/firebase/config.json\n")
	case "Google Drive":
		file.WriteString("STORAGE_TYPE=GoogleDrive\n")
		file.WriteString("GOOGLE_DRIVE_CLIENT_ID=your-client-id\n")
		file.WriteString("GOOGLE_DRIVE_CLIENT_SECRET=your-client-secret\n")
		file.WriteString("GOOGLE_DRIVE_REFRESH_TOKEN=your-refresh-token\n")
	}

	return nil
}

func Launch() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
