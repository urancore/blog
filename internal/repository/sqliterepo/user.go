package sqliterepo

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	sqlite "github.com/mattn/go-sqlite3"

	"blog/internal/models"
	"blog/internal/repository"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

type SQliteUserRepo struct {
	log logger.Logger
	db  *sql.DB
}

func (r *SQliteUserRepo) InitUserDatabase() error {
	stmt := `
	CREATE TABLE IF NOT EXISTS user (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := r.db.Exec(stmt)
	return err
}

func (r *SQliteUserRepo) CreateUser(user *models.User) (int64, error) {
	log := r.log.With("fn", "repository.sqlite.CreateUser")
	query := `
		INSERT INTO user (username, password, email)
		VALUES (?, ?, ?)
	`

	res, err := r.db.Exec(query, user.Username, user.Password, user.Email)
	if err != nil {
		if errors.Is(err, sqlite.ErrConstraintUnique) {
			if strings.Contains(err.Error(), "username") {
				return 0, repository.ErrUsernameAlreadyExists
			}
			if strings.Contains(err.Error(), "email") {
				return 0, repository.ErrEmailAlreadyExists
			}
		}
		log.Error("error creating user", "error", sl.Error(err))
		return 0, fmt.Errorf("failed to create user: %w", repository.ErrOperationFailed)
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Error("failed to get last insert ID", "error", sl.Error(err))
		return 0, fmt.Errorf("failed to get user ID: %w", repository.ErrOperationFailed)
	}

	return id, nil
}

func (r *SQliteUserRepo) GetUserByID(id int64) (*models.User, error) {
	log := r.log.With("fn", "repository.sqlite.GetUserByID")
	query := `
		SELECT id, username, password, email, created_at, updated_at
		FROM user
		WHERE id = ?
	`

	var user models.User
	row := r.db.QueryRow(query, id)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("user not found", "id", id)
			return nil, repository.ErrNotExists
		}
		log.Error("failed to get user", "error", sl.Error(err), "id", id)
		return nil, fmt.Errorf("query error: %w", repository.ErrOperationFailed)
	}

	return &user, nil
}

func (r *SQliteUserRepo) GetUserByEmail(email string) (*models.User, error) {
	log := r.log.With("fn", "repository.sqlite.GetUserByEmail")
	query := `
		SELECT id, username, password, email, created_at, updated_at
		FROM user
		WHERE email = ?
	`

	var user models.User
	row := r.db.QueryRow(query, email)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("user not found", "email", email)
			return nil, repository.ErrNotExists
		}
		log.Error("failed to get user", "error", sl.Error(err), "email", email)
		return nil, fmt.Errorf("query error: %w", repository.ErrOperationFailed)
	}

	return &user, nil
}

func (r *SQliteUserRepo) UpdateUser(user *models.User) error {
	log := r.log.With("fn", "repository.sqlite.UpdateUser")
	query := `
		UPDATE user
		SET
			username = ?,
			password = ?,
			email = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	res, err := r.db.Exec(
		query,
		user.Username,
		user.Password,
		user.Email,
		user.ID,
	)
	if err != nil {
		if errors.Is(err, sqlite.ErrConstraintUnique) {
			if strings.Contains(err.Error(), "username") {
				return repository.ErrUsernameAlreadyExists
			}
			if strings.Contains(err.Error(), "email") {
				return repository.ErrEmailAlreadyExists
			}
		}
		log.Error("failed to update user", "error", err, "id", user.ID)
		return fmt.Errorf("update error: %w", repository.ErrOperationFailed)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		log.Info("user not found during update", "id", user.ID)
		return repository.ErrNotExists
	}

	return nil
}

func (r *SQliteUserRepo) DeleteUser(id int64) error {
	log := r.log.With("fn", "repository.sqlite.DeleteUser")
	query := `DELETE FROM user WHERE id = ?`

	res, err := r.db.Exec(query, id)
	if err != nil {
		log.Error("failed to delete user", "error", err, "id", id)
		return fmt.Errorf("delete error: %w", repository.ErrOperationFailed)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		log.Info("user not found during deletion", "id", id)
		return repository.ErrNotExists
	}

	return nil
}

func (r *SQliteUserRepo) ListUsers(limit, offset int) ([]*models.User, error) {
	log := r.log.With("fn", "repository.sqlite.ListUsers")
	query := `
		SELECT
			id,
			username,
			email,
			created_at,
			updated_at
		FROM user
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		log.Error("failed to list users", "error", err)
		return nil, fmt.Errorf("query error: %w", repository.ErrOperationFailed)
	}
	defer rows.Close()

	users := []*models.User{}
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			log.Error("failed to scan user", "error", err)
			return nil, fmt.Errorf("scan error: %w", repository.ErrOperationFailed)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		log.Error("iteration error", "error", err)
		return nil, fmt.Errorf("row iteration error: %w", repository.ErrOperationFailed)
	}

	return users, nil
}
