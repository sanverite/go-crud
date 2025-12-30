package data

import (
	"database/sql"
	"errors"
)

// A custom ErrRecordNotFound error to return from Get()
// method when looking up a movie that doesn't exist in
// the databse.
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// A Models struct which wraps all the other models.
type Models struct {
	Movies MovieModel
}

// NewModels returns a Models struct containing the
// initialized MovieModel.
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}
