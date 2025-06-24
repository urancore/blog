package sqliterepo

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"blog/internal/models"
	"blog/internal/repository"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

type SQlitePostRepo struct {
	log logger.Logger
	db  *sql.DB
}

func (r *SQlitePostRepo) InitPostDatabase() error {
	stmt := `
	CREATE TABLE IF NOT EXISTS post (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		author_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (author_id) REFERENCES user(id)
	);
	`
	if _, err := r.db.Exec(stmt); err != nil {
		return err
	}

	return nil
}
func (r *SQlitePostRepo) GetPostByID(id int64) (*models.Post, error) {
	log := r.log.With("fn", "repository.sqliterepo.GetPostByID")
	query := `
		SELECT id, title, content, author_id, created_at, updated_at
		FROM post
		WHERE id = ?
    	`

	var post models.Post
	row := r.db.QueryRow(query, id)
	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.AuthorID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("post not found", slog.Int64("id", id))
			return nil, repository.ErrNotExists
		}
		log.Error("failed to get post", "error", sl.Error(err), slog.Int64("id", id))
		return nil, fmt.Errorf("query error: %w", repository.ErrOperationFailed)
	}

	return &post, nil
}

func (r *SQlitePostRepo) CreatePost(post *models.Post) (int64, error) {
	log := r.log.With("fn", "repository.sqliterepo.CreatePost")
	query := `
		INSERT INTO post (title, content, author_id)
		VALUES (?, ?, ?)
		RETURNING id, created_at, updated_at
    	`

	var id int64
	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(query, post.Title, post.Content, post.AuthorID).Scan(
		&id,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err.Error() == "FOREIGN KEY constraint failed" {
			return 0, repository.ErrForeignKeyFailed
		}
		log.Error("failed to create post", sl.Error(err))
		return 0, repository.ErrOperationFailed
	}

	post.ID = id
	post.CreatedAt = createdAt
	post.UpdatedAt = updatedAt
	return id, nil
}

func (r *SQlitePostRepo) UpdatePost(post *models.Post) error {
	log := r.log.With(slog.String("fn", "repository.sqliterepo.UpdatePost"))
	query := `
		UPDATE post
		SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND author_id = ?`

	res, err := r.db.Exec(query,
		post.Title,
		post.Content,
		post.ID,
		post.AuthorID)

	if err != nil {
		log.Error("update query failed",
			"error", err,
			"id", post.ID,
			"author_id", post.AuthorID)
		return fmt.Errorf("update error: %w", repository.ErrOperationFailed)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error("rows affected check failed", "error", err)
		return fmt.Errorf("update error: %w", repository.ErrOperationFailed)
	}

	if rowsAffected == 0 {
		log.Info("post not found or author mismatch",
			"id", post.ID,
			"author_id", post.AuthorID)
		return repository.ErrNotExists
	}
	return nil
}

func (r *SQlitePostRepo) DeletePost(id int64) error {
	log := r.log.With("fn", "repository.sqliterepo.DeletePost")
	query := `DELETE FROM post WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		log.Error("failed to delete post", "error", err, "id", id)
		return fmt.Errorf("delete error: %w", repository.ErrOperationFailed)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		log.Info("post not found during deletion", "id", id)
		return repository.ErrNotExists
	}
	return nil
}

func (r *SQlitePostRepo) GetPostsByAuthor(authorID int64) ([]*models.Post, error) {
	log := r.log.With("fn", "repository.sqliterepo.GetPostsByAuthor")
	query := `
		SELECT id, title, content, author_id, created_at, updated_at
		FROM post
		WHERE author_id = ?
		ORDER BY created_at DESC
    	`
	rows, err := r.db.Query(query, authorID)
	if err != nil {
		log.Error("failed to fetch posts", "error", err, "author_id", authorID)
		return nil, fmt.Errorf("query error: %w", repository.ErrOperationFailed)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.AuthorID,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			log.Error("failed to scan post", "error", err)
			return nil, fmt.Errorf("scan error: %w", repository.ErrOperationFailed)
		}
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		log.Error("iteration error", "error", err)
		return nil, fmt.Errorf("row iteration error: %w", repository.ErrOperationFailed)
	}
	return posts, nil
}

func (r *SQlitePostRepo) ListPosts(limit, offset int) ([]*models.Post, error) {
	log := r.log.With("fn", "repository.sqliterepo.ListPosts")
	query := `
		SELECT id, title, content, author_id, created_at, updated_at
		FROM post
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
    	`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		log.Error("failed to list posts", "error", err)
		return nil, fmt.Errorf("query error: %w", repository.ErrOperationFailed)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.AuthorID,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			log.Error("failed to scan post", "error", err)
			return nil, fmt.Errorf("scan error: %w", repository.ErrOperationFailed)
		}
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		log.Error("iteration error", "error", err)
		return nil, fmt.Errorf("row iteration error: %w", repository.ErrOperationFailed)
	}
	return posts, nil
}
