package repository

import (
	"blog/internal/models"
	"errors"
)

var (
	ErrNotExists             = errors.New("not found")
	ErrOperationFailed       = errors.New("database operation failed")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("userame already exists")
	ErrForeignKeyFailed = errors.New("foreign key not found")
)

type UserRepository interface {
	InitUserDatabase() error
	CreateUser(user *models.User) (int64, error)
	GetUserByID(id int64) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id int64) error
	ListUsers(limit, offset int) ([]*models.User, error)
}

type PostRepository interface {
	InitPostDatabase() error
	GetPostByID(id int64) (*models.Post, error)
	CreatePost(post *models.Post) (int64, error)
	UpdatePost(post *models.Post) error
	DeletePost(id int64) error
	GetPostsByAuthor(authorID int64) ([]*models.Post, error)
	ListPosts(limit, offset int) ([]*models.Post, error)
}

type CommentRepository interface {

}

type Repository interface {
	User() UserRepository
	Post() PostRepository
	Comment() CommentRepository
}
