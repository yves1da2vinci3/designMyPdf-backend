package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Backblaze/blazer/b2"
)

type BackblazeStorage struct {
	client     *b2.Client
	bucket     *b2.Bucket
	bucketName string
}

// NewBackblazeStorage crée une nouvelle instance de stockage Backblaze
func NewBackblazeStorage(accountID, applicationKey, bucketName string) (*BackblazeStorage, error) {
	ctx := context.Background()
	client, err := b2.NewClient(ctx, accountID, applicationKey)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du client B2: %v", err)
	}

	bucket, err := client.Bucket(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'accès au bucket: %v", err)
	}

	return &BackblazeStorage{
		client:     client,
		bucket:     bucket,
		bucketName: bucketName,
	}, nil
}

// UploadFile télécharge un fichier vers Backblaze B2
func (s *BackblazeStorage) UploadFile(ctx context.Context, filePath string, objectName string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("erreur lors de l'ouverture du fichier: %v", err)
	}
	defer file.Close()

	obj := s.bucket.Object(objectName)
	writer := obj.NewWriter(ctx)
	
	// Configuration pour les fichiers volumineux (>100MB)
	writer.ConcurrentUploads = 3 // Nombre d'uploads concurrents pour les fichiers volumineux
	
	if _, err := io.Copy(writer, file); err != nil {
		writer.Close()
		return "", fmt.Errorf("erreur lors de l'upload du fichier: %v", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("erreur lors de la fermeture du writer: %v", err)
	}

	// Obtenir l'URL de base du bucket
	baseURL := s.bucket.BaseURL()
	// Construire l'URL complète
	url := fmt.Sprintf("%s/file/%s/%s", baseURL, s.bucketName, objectName)
	return url, nil
}

// DownloadFile télécharge un fichier depuis Backblaze B2
func (s *BackblazeStorage) DownloadFile(ctx context.Context, objectName string, destinationPath string) error {
	obj := s.bucket.Object(objectName)
	reader := obj.NewReader(ctx)
	defer reader.Close()

	// Configuration pour les téléchargements volumineux
	reader.ConcurrentDownloads = 3 // Nombre de téléchargements concurrents

	// Créer le dossier de destination si nécessaire
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
		return fmt.Errorf("erreur lors de la création du dossier de destination: %v", err)
	}

	file, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du fichier de destination: %v", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("erreur lors du téléchargement du fichier: %v", err)
	}

	return nil
}

// DeleteFile supprime un fichier de Backblaze B2
func (s *BackblazeStorage) DeleteFile(ctx context.Context, objectName string) error {
	obj := s.bucket.Object(objectName)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("erreur lors de la suppression du fichier: %v", err)
	}
	return nil
}

// GetFileURL génère une URL pour un fichier
func (s *BackblazeStorage) GetFileURL(ctx context.Context, objectName string) (string, error) {
	baseURL := s.bucket.BaseURL()
	url := fmt.Sprintf("%s/file/%s/%s", baseURL, s.bucketName, objectName)
	return url, nil
}

// GetTemporaryAuthToken génère un token d'authentification temporaire pour un préfixe
func (s *BackblazeStorage) GetTemporaryAuthToken(ctx context.Context, prefix string, duration time.Duration) (string, error) {
	token, err := s.bucket.AuthToken(ctx, prefix, duration)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la génération du token d'authentification: %v", err)
	}
	return token, nil
}

// ListFiles liste tous les fichiers dans un préfixe donné
func (s *BackblazeStorage) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	var files []string
	iterator := s.bucket.List(ctx)
	for iterator.Next() {
		obj := iterator.Object()
		if prefix == "" || obj.Name()[:len(prefix)] == prefix {
			files = append(files, obj.Name())
		}
	}
	if err := iterator.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de la liste des fichiers: %v", err)
	}
	return files, nil
} 