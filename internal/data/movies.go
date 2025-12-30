package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/sanverite/greenlight/internal/validator"
)

// A MovieModel struct type which wraps a sql.DB connection pool.
type MovieModel struct {
	DB *sql.DB
}

// The Insert() method accepts a pointer to a movie struct, which
// should contain the data for the new record.
func (m MovieModel) Insert(movie *Movie) error {
	// Define the SQL query for inserting a new record in the movies
	// table and returning the system-generated data.
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	// An args slice containing the values for the parameters from the
	// movie struct. This will be used in DB.QueryRow() method to execute
	// the SQL query.
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Use the QueryRow() method to execute the SQL query on our connection
	// pool, passing in the args slice as a variadic parameter and scanning
	// the system-generated id, created_at and version values into the movie
	// struct.
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// The Get() method accepts a movie's ID to query into the movies table.
// Return the movie record if it exists for the given ID, else return an
// ErrRecordNotFound error.
func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving the movie data.
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id = $1`

	// Declare a Movie struct to hold the data returned by the query;
	var movie Movie

	// Execute the query using the QueryRow() method, passing in the
	// provided id value as a placeholder parameter, and scan the
	// response data into the fields of the Movie struct.
	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

// The Update() method will take the entire movie struct and update
// it in the movies table.
func (m MovieModel) Update(movie *Movie) error {
	// Declare the SQL query for updating the record and returning the
	// new version number.
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	// Create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice
	// as a variadic parameter and scanning the new version into the movie struct.
	// If no matching row could be found, the movie version has changed (or the
	// record has been deleted), return custom ErrConflict error.
	err := m.DB.QueryRow(query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

// The Delete() method accepts a movie's ID to delete the record in the
// movies table.
func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	// Define the SQL query for deleting a record
	query := `
		DELETE FROM movies
		where id = $1`

	// Execute the SQL query using the Exec() method, passing in the id variable
	// as the value for the placeholder parameter. The Exec() method returns a
	// sql.Result object.
	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected() method on the sql.Result object to get the number
	// of rows affected by the query.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were affected, we know that the movies table didn't contain a
	// record with the provided ID. In that case return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

type Movie struct {
	ID        int64     `json:"id"`                // Unique integer ID for the movie
	CreatedAt time.Time `json:"-"`                 // Timestamp for when the movie is added to our database
	Title     string    `json:"title"`             // Movie Title
	Year      int32     `json:"year,omitempty"`    // Movie release year
	Runtime   Runtime   `json:"runtime,omitempty"` // Movie runtime  (in minutes)
	Genres    []string  `json:"genres,omitempty"`  // Slice of genres for the movie (romance, comedy, etc)
	Version   int32     `json:"version"`           // The version number starts at 1 and will be incremented each time the movie information is updated
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}
