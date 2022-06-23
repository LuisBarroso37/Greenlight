package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/LuisBarroso37/Greenlight/internal/validator"
	"github.com/lib/pq"
)

type Movie struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`    // Movie release year
	Runtime   Runtime   `json:"runtime,omitempty"` // Movie runtime (in minutes)
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"` // The version number starts at 1 and will be incremented each time the movie information is updated
	CreatedAt time.Time `json:"-"`
}

// Run validation checks on `Movie` struct
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

// Define a MovieModel struct type which wraps a sql.DB connection pool
type MovieModel struct {
	DB *sql.DB
}

// Inserts a new record in the `movies` table
func (m MovieModel) Insert(movie *Movie) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
  	INSERT INTO movies (title, year, runtime, genres) 
    VALUES ($1, $2, $3, $4)
    RETURNING id, created_at, version`

	return m.DB.QueryRowContext(
		ctx,
		query,
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
	).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Fetches a specific record from the `movies` table
func (m MovieModel) Get(id int64) (*Movie, error) {
	// To avoid making an unnecessary database call, we return an error if received id
	// is less than 1
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var movie Movie

	query := `
  	SELECT id, title, year, runtime, genres, version, created_at
    FROM movies
    WHERE id = $1`

	err := m.DB.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&movie.ID,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
		&movie.CreatedAt,
	)

	// If there was no matching movie found, Scan() will return
	// a sql.ErrNoRows error. We check for this and return our custom ErrRecordNotFound
	// error instead.
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

// Updates a specific record from the `movies` table
// JSON items with null values will be ignored and will remain unchanged
func (m MovieModel) Update(movie *Movie) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
  	UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
    WHERE id = $5 and version = $6
		RETURNING version`

	err := m.DB.QueryRowContext(
		ctx,
		query,
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	).Scan(&movie.Version)
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

// Deletes a specific record from the `movies` table
func (m MovieModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		DELETE FROM movies
		WHERE id = $1`

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// Fetches all movie records from the `movies` table
func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	totalRecords := 0
	movies := []*Movie{}

	// We also include a secondary sort on the movie ID to ensure a
	// consistent ordering
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, title, year, runtime, genres, version, created_at
		FROM movies
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
    AND (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	rows, err := m.DB.QueryContext(
		ctx,
		query,
		title,
		pq.Array(genres),
		filters.limit(),
		filters.offset(),
	)
	if err != nil {
		return nil, Metadata{}, err
	}

	// Defer a call to rows.Close() to ensure that the result set is closed
	// before GetAll() returns
	defer rows.Close()

	for rows.Next() {
		var movie Movie

		err := rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
			&movie.CreatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		movies = append(movies, &movie)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Generate a Metadata struct
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return movies, metadata, nil
}
