package sqliterepo

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/mattn/go-sqlite3"

	"blog/internal/models"
	"blog/internal/repository"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

type SQliteCommentRepo struct {
	log logger.Logger
	db  *sql.DB
}

func (r *SQliteCommentRepo) InitCommentDatabase() error {
	stmt := `
	CREATE TABLE IF NOT EXISTS comment (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		post_id INTEGER NOT NULL,
		author_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (post_id) REFERENCES post(id) ON DELETE CASCADE,
		FOREIGN KEY (author_id) REFERENCES user(id) ON DELETE SET NULL
	);

	`
	if _, err := r.db.Exec(stmt); err != nil {
		return err
	}

	return nil
}

func (r *SQliteCommentRepo) CreateComment(comment *models.Comment) (int64, error) {
	log := r.log.With("fn", "repository.sqliterepo.CreateComment")
	query := `
		INSERT INTO comment (content, post_id, author_id)
		VALUES (?, ?, ?)
		RETURNING id, created_at, updated_at
    	`

	var id int64
	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(query, comment.Content, comment.PostID, comment.AuthorID).Scan(
		&id,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sqlite3.ErrConstraintForeignKey) {
			return 0, repository.ErrForeignKeyFailed
		}
		log.Error("failed to create comment", sl.Error(err))
		return 0, repository.ErrOperationFailed
	}

	comment.ID = id
	comment.CreatedAt = createdAt
	comment.UpdatedAt = updatedAt
	return id, nil
}

func (r *SQliteCommentRepo) GetCommentAuthorID(commentID int64) (int64, error) {
	log := r.log.With(slog.String("fn", "repository.sqliterepo.GetCommentAuthor"), slog.Int64("comment_id", commentID))
	query := `SELECT author_id FROM comment WHERE id = ?`

	var authorID int64
	err := r.db.QueryRow(query, commentID).Scan(&authorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info("comment not found")
			return 0, repository.ErrNotExists
		}
		log.Error("failed to get comment author", sl.Error(err))
		return 0, fmt.Errorf("query error: %w", repository.ErrOperationFailed)
	}

	return authorID, nil
}

func (r *SQliteCommentRepo) UpdateComment(comment *models.Comment) error {
	log := r.log.With(slog.String("fn", "repository.sqliterepo.UpdateComment"))
	query := `
		UPDATE comment
		SET content = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND post_id = ? AND author_id = ?
	`

	res, err := r.db.Exec(query,
		comment.Content,
		comment.ID,
		comment.PostID,
		comment.AuthorID)

	if err != nil {
		log.Error("update query failed",
			sl.Error(err),
			slog.Int64("id", comment.ID),
			slog.Int64("author_id", comment.AuthorID),
			slog.Int64("post_id", comment.PostID))

		return fmt.Errorf("update error: %w", repository.ErrOperationFailed)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error("rows affected check failed", sl.Error(err))

		return fmt.Errorf("update error: %w", repository.ErrOperationFailed)
	}

	if rowsAffected == 0 {
		log.Info("comment not found or author mismatch or post mismatch",
			slog.Int64("id", comment.ID),
			slog.Int64("post_id", comment.PostID),
			slog.Int64("author_id", comment.AuthorID))

		return repository.ErrNotExists
	}

	return nil
}

func (r *SQliteCommentRepo) DeleteComment(id int64) error {
	log := r.log.With("fn", "repository.sqliterepo.DeleteComment")
	query := `DELETE FROM comment WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		log.Error("failed to delete comment", sl.Error(err), slog.Int64("id", id))
		return fmt.Errorf("delete error: %w", repository.ErrOperationFailed)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		log.Info("comment not found during deletion", slog.Int64("id", id))
		return repository.ErrNotExists
	}

	return nil
}

func (r *SQliteCommentRepo) ListComments(limit int, offset int, postID int64) ([]*models.Comment, error) {
	log := r.log.With("fn", "repository.sqliterepo.ListComment")
	query := `
		SELECT id, content, post_id, author_id, created_at, updated_at
		FROM comment
		WHERE post_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
    	`
	rows, err := r.db.Query(query, postID, limit, offset)
	if err != nil {
		log.Error("failed to list comments", sl.Error(err))
		return nil, fmt.Errorf("query error: %w", repository.ErrOperationFailed)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(
			&comment.ID,
			&comment.Content,
			&comment.PostID,
			&comment.AuthorID,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			log.Error("failed to scan comment", "error", err)
			return nil, fmt.Errorf("scan error: %w", repository.ErrOperationFailed)
		}
		comments = append(comments, &comment)
	}

	if err := rows.Err(); err != nil {
		log.Error("iteration error", "error", err)
		return nil, fmt.Errorf("row iteration error: %w", repository.ErrOperationFailed)
	}
	return comments, nil
}
