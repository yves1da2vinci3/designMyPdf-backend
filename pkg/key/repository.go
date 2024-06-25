package key

import (
	"designmypdf/pkg/entities"
	"math/rand"
	"strings"

	"gorm.io/gorm"
)

// Repository is a GORM implementation of KeyRepository
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(key *entities.Key) error {
	key.Value = generateKey()
	return r.db.Create(key).Error
}

func (r *Repository) Get(id uint) (*entities.Key, error) {
	var key entities.Key
	if err := r.db.First(&key, id).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *Repository) Update(key *entities.Key) error {
	return r.db.Save(key).Error
}

func (r *Repository) Delete(key *entities.Key) error {
	return r.db.Delete(key).Error
}

func (r *Repository) GetAll() ([]*entities.Key, error) {
	var keys []*entities.Key
	if err := r.db.Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *Repository) GetAllUserKeys(userID uint) ([]entities.Key, error) {
	var keys []entities.Key
	if err := r.db.Where("user_id = ?", userID).Preload("Logs").Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *Repository) GetKeyByValue(keyValue string) (*entities.Key, error) {
	var key entities.Key
	if err := r.db.Where("value = ?", keyValue).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

// IncreaseUsageCount increments the usage count of a key
func (r *Repository) IncreaseUsageCount(id uint) error {
	var key entities.Key
	if err := r.db.First(&key, id).Error; err != nil {
		return err
	}
	key.KeyCountUsed++
	return r.Update(&key)
}

func generateKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 32
	var sb strings.Builder
	sb.WriteString("dmp_")
	for i := 0; i < length-len("dmp_"); i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}
