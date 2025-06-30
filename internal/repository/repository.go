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
	ErrForeignKeyFailed      = errors.New("foreign key not found")
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
	CreatePost(post *models.Post) (int64, error)
	GetPostByID(id int64) (*models.Post, error)
	GetPostsByAuthor(authorID int64) ([]*models.Post, error)
	UpdatePost(post *models.Post) error
	DeletePost(id int64) error
	ListPosts(limit, offset int) ([]*models.Post, error)
}

type CommentRepository interface {
	InitCommentDatabase() error
	CreateComment(comment *models.Comment) (int64, error)
	GetCommentAuthorID(commentID int64) (int64, error)
	UpdateComment(comment *models.Comment) error
	DeleteComment(id int64) error
	ListComments(limit int, offset int, postID int64) ([]*models.Comment, error)
}

type Repository interface {
	User() UserRepository
	Post() PostRepository
	Comment() CommentRepository
}
