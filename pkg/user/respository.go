package user

// UserRepository defines the methods that any repository implementation should have
type UserRepository interface {
	Create(user interface{}) error
	Get(id interface{}) (interface{}, error)
	Update(user interface{}) error
	Delete(user interface{}) error
	GetAll() ([]interface{}, error)
	GetByEmail(email string) (interface{}, error)
	GetByUserName(userName string) (interface{}, error)
	GetByUserNameAndPassword(userName string, password string) (interface{}, error)
}
