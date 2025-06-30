package sqliterepo

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"blog/internal/repository"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

func New(databasePath string, log logger.Logger) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		log.Error("error connection database", sl.Error(err))
		return nil, err
	}

	if err = db.Ping(); err != nil {
		log.Error("error ping database", sl.Error(err))
		return nil, err
	}

	return db, nil
}

type SQLiteRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewSQLiteRepository(log logger.Logger, db *sql.DB) repository.Repository {
	return &SQLiteRepository{
		db:  db,
		log: log,
	}
}

func (r *SQLiteRepository) User() repository.UserRepository {
	return &SQliteUserRepo{log: r.log, db: r.db}
}

func (r *SQLiteRepository) Post() repository.PostRepository {
	return &SQlitePostRepo{log: r.log, db: r.db}
}

func (r *SQLiteRepository) Comment() repository.CommentRepository {
	return &SQliteCommentRepo{log: r.log, db: r.db}
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
