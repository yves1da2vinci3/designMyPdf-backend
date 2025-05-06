package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Backblaze/blazer/b2"
)

// Cache pour stocker les URLs fréquemment demandées
type urlCache struct {
	cache map[string]string
	mu    sync.RWMutex
}

func newURLCache() *urlCache {
	return &urlCache{
		cache: make(map[string]string),
	}
}

func (c *urlCache) get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, exists := c.cache[key]
	return val, exists
}

func (c *urlCache) set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = value
}

type BackblazeStorage struct {
	client     *b2.Client
	bucket     *b2.Bucket
	bucketName string
	baseURL    string
	urlCache   *urlCache
	clientPool sync.Pool // Pool de clients pour réutiliser les connexions
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

	baseURL := bucket.BaseURL()
	
	// Initialiser le pool de clients
	clientPool := sync.Pool{
		New: func() interface{} {
			c, _ := b2.NewClient(context.Background(), accountID, applicationKey)
			return c
		},
	}

	return &BackblazeStorage{
		client:     client,
		bucket:     bucket,
		bucketName: bucketName,
		baseURL:    baseURL,
		urlCache:   newURLCache(),
		clientPool: clientPool,
	}, nil
}

// getClient récupère un client du pool
func (s *BackblazeStorage) getClient() *b2.Client {
	return s.clientPool.Get().(*b2.Client)
}

// releaseClient remet un client dans le pool
func (s *BackblazeStorage) releaseClient(client *b2.Client) {
	s.clientPool.Put(client)
}

// UploadFile télécharge un fichier vers Backblaze B2 avec optimisations avancées
func (s *BackblazeStorage) UploadFile(ctx context.Context, filePath string, objectName string) (string, error) {
	// Vérifier si l'URL est déjà en cache
	if cachedURL, found := s.urlCache.get(objectName); found {
		return cachedURL, nil
	}

	// Timeout court pour une meilleure expérience utilisateur
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("erreur lors de l'ouverture du fichier: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("erreur lors de l'obtention des informations du fichier: %v", err)
	}

	obj := s.bucket.Object(objectName)
	writer := obj.NewWriter(ctx)
	
	// Optimisations spécifiques selon la taille
	fileSize := fileInfo.Size()
	if fileSize > 5*1024*1024 { // Plus de 5MB
		// Calculer le nombre optimal d'uploads concurrents (1 par 5MB avec max de 8)
		concurrentUploads := int(fileSize / (5 * 1024 * 1024))
		if concurrentUploads < 1 {
			concurrentUploads = 1
		} else if concurrentUploads > 8 {
			concurrentUploads = 8
		}
		writer.ConcurrentUploads = concurrentUploads

		// Taille de chunk optimale pour maximiser la vitesse vs consommation mémoire
		chunkSize := 5 * 1024 * 1024 // 5MB (minimum recommandé pour Backblaze)
		writer.ChunkSize = chunkSize
	}
	
	// Buffer optimisé pour la copie
	bufSize := 4 * 1024 * 1024 // 4MB buffer
	_, err = io.CopyBuffer(writer, file, make([]byte, bufSize))
	if err != nil {
		writer.Close()
		return "", fmt.Errorf("erreur lors de l'upload du fichier: %v", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("erreur lors de la fermeture du writer: %v", err)
	}

	// Construire l'URL et la mettre en cache
	url := fmt.Sprintf("%s/file/%s/%s", s.baseURL, s.bucketName, objectName)
	s.urlCache.set(objectName, url)
	
	return url, nil
}

// DownloadFile télécharge un fichier depuis Backblaze B2 (optimisé)
func (s *BackblazeStorage) DownloadFile(ctx context.Context, objectName string, destinationPath string) error {
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	obj := s.bucket.Object(objectName)
	reader := obj.NewReader(ctx)
	defer reader.Close()

	// Optimisation agressive pour les téléchargements
	reader.ConcurrentDownloads = 8

	// Créer le dossier de destination si nécessaire
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
		return fmt.Errorf("erreur lors de la création du dossier de destination: %v", err)
	}

	file, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du fichier de destination: %v", err)
	}
	defer file.Close()

	// Buffer optimisé pour la copie
	bufSize := 4 * 1024 * 1024 // 4MB buffer
	_, err = io.CopyBuffer(file, reader, make([]byte, bufSize))
	if err != nil {
		return fmt.Errorf("erreur lors du téléchargement du fichier: %v", err)
	}

	return nil
}

// DeleteFile supprime un fichier de Backblaze B2
func (s *BackblazeStorage) DeleteFile(ctx context.Context, objectName string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Supprimer l'URL du cache
	s.urlCache.mu.Lock()
	delete(s.urlCache.cache, objectName)
	s.urlCache.mu.Unlock()

	obj := s.bucket.Object(objectName)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("erreur lors de la suppression du fichier: %v", err)
	}
	return nil
}

// GetFileURL génère une URL pour un fichier (optimisé avec cache)
func (s *BackblazeStorage) GetFileURL(ctx context.Context, objectName string) (string, error) {
	// Vérifier si l'URL est déjà en cache
	if cachedURL, found := s.urlCache.get(objectName); found {
		return cachedURL, nil
	}

	// Construire l'URL
	url := fmt.Sprintf("%s/file/%s/%s", s.baseURL, s.bucketName, objectName)
	
	// Stocker dans le cache
	s.urlCache.set(objectName, url)
	
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
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var files []string
	var options []b2.ListOption
	
	if prefix != "" {
		options = append(options, b2.ListPrefix(prefix))
	}

	// Limiter le nombre maximum de fichiers retournés pour éviter des listes trop grandes
	// Note: Utilisons une approche différente car ListMaxCount n'existe pas
	const maxResults = 1000
	
	iterator := s.bucket.List(ctx, options...)
	count := 0
	for iterator.Next() && count < maxResults {
		files = append(files, iterator.Object().Name())
		count++
	}
	
	if err := iterator.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de la liste des fichiers: %v", err)
	}
	return files, nil
} 