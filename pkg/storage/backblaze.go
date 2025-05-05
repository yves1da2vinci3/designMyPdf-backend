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
	baseURL    string // Stockage du baseURL pour éviter de le recalculer
}

// NewBackblazeStorage crée une nouvelle instance de stockage Backblaze
func NewBackblazeStorage(accountID, applicationKey, bucketName string) (*BackblazeStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	client, err := b2.NewClient(ctx, accountID, applicationKey)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du client B2: %v", err)
	}

	bucket, err := client.Bucket(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'accès au bucket: %v", err)
	}

	// Récupérer et stocker l'URL de base une seule fois
	baseURL := bucket.BaseURL()

	return &BackblazeStorage{
		client:     client,
		bucket:     bucket,
		bucketName: bucketName,
		baseURL:    baseURL,
	}, nil
}

// UploadFile télécharge un fichier vers Backblaze B2 avec des optimisations de performance
func (s *BackblazeStorage) UploadFile(ctx context.Context, filePath string, objectName string) (string, error) {
	// Ajout d'un timeout pour l'opération d'upload
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("erreur lors de l'ouverture du fichier: %v", err)
	}
	defer file.Close()

	// Obtenir les stats du fichier pour optimiser selon la taille
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("erreur lors de l'obtention des informations du fichier: %v", err)
	}

	obj := s.bucket.Object(objectName)
	writer := obj.NewWriter(ctx)
	
	// Optimisation pour les fichiers volumineux
	fileSize := fileInfo.Size()
	if fileSize > 5*1024*1024 { // Plus de 5MB
		// Nombre d'uploads concurrents basé sur la taille du fichier
		concurrentUploads := 6 // Maximum pour les performances
		if fileSize < 20*1024*1024 { // Moins de 20MB
			concurrentUploads = 3
		}
		writer.ConcurrentUploads = concurrentUploads

		// Taille du tampon optimisée
		writer.ChunkSize = 5 * 1024 * 1024 // 5MB par chunk au lieu de 100MB
	}
	
	if _, err := io.Copy(writer, file); err != nil {
		writer.Close()
		return "", fmt.Errorf("erreur lors de l'upload du fichier: %v", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("erreur lors de la fermeture du writer: %v", err)
	}

	// Construire l'URL directement (pré-calculée)
	url := fmt.Sprintf("%s/file/%s/%s", s.baseURL, s.bucketName, objectName)
	return url, nil
}

// DownloadFile télécharge un fichier depuis Backblaze B2 (optimisé)
func (s *BackblazeStorage) DownloadFile(ctx context.Context, objectName string, destinationPath string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	obj := s.bucket.Object(objectName)
	reader := obj.NewReader(ctx)
	defer reader.Close()

	// Optimisation pour les téléchargements
	reader.ConcurrentDownloads = 6

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
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	obj := s.bucket.Object(objectName)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("erreur lors de la suppression du fichier: %v", err)
	}
	return nil
}

// GetFileURL génère une URL pour un fichier (optimisé, pas d'appel à l'API)
func (s *BackblazeStorage) GetFileURL(ctx context.Context, objectName string) (string, error) {
	url := fmt.Sprintf("%s/file/%s/%s", s.baseURL, s.bucketName, objectName)
	return url, nil
}

// GetTemporaryAuthToken génère un token d'authentification temporaire pour un préfixe
func (s *BackblazeStorage) GetTemporaryAuthToken(ctx context.Context, prefix string, duration time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	token, err := s.bucket.AuthToken(ctx, prefix, duration)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la génération du token d'authentification: %v", err)
	}
	return token, nil
}

// ListFiles liste tous les fichiers dans un préfixe donné (optimisé)
func (s *BackblazeStorage) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var files []string
	var options []b2.ListOption
	
	if prefix != "" {
		options = append(options, b2.ListPrefix(prefix))
	}
	
	iterator := s.bucket.List(ctx, options...)
	for iterator.Next() {
		files = append(files, iterator.Object().Name())
	}
	
	if err := iterator.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de la liste des fichiers: %v", err)
	}
	return files, nil
} 