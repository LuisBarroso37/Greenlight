package data

import (
	"database/sql"
	"errors"
	"time"
)

// We'll return this from our Get() method when
// looking up a movie that doesn't exist in our database.
var ErrRecordNotFound = errors.New("record not found")

// We'll return this from our Update() method when
// a movie is being updated by more than 1 person at the the same time - data race.
var ErrEditConflict = errors.New("edit conflict")

type Models struct {
	Movie interface {
		Insert(movie *Movie) error
		Get(id int64) (*Movie, error)
		Update(movie *Movie) error
		Delete(id int64) error
		GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error)
	}
	User interface {
		Insert(user *User) error
		GetByEmail(email string) (*User, error)
		Update(user *User) error
		GetForToken(tokenScope, tokenPlaintext string) (*User, error)
	}
	Token interface {
		New(userID int64, ttl time.Duration, scope string) (*Token, error)
		Insert(token *Token) error
		DeleteAllForUser(scope string, userID int64) error
	}
	Permissions interface {
		GetAllForUser(userID int64) (Permissions, error)
		AddForUser(userID int64, codes ...string) error
	}
}

// Method used to initialize `Models` struct
func NewModels(db *sql.DB) Models {
	return Models{
		Movie:       MovieModel{DB: db},
		User:        UserModel{DB: db},
		Token:       TokenModel{DB: db},
		Permissions: PermissionModel{DB: db},
	}
}

// Method used to initialize mock of `Models` struct
func NewMockModels(db *sql.DB) Models {
	return Models{
		Movie:       MockMovieModel{},
		User:        MockUserModel{},
		Token:       MockTokenModel{},
		Permissions: MockPermissionsModel{},
	}
}
