package data

import "time"

// Define a mock of the `TokenModel` struct type
type MockTokenModel struct{}

// The New() method is a shortcut which creates a new Token struct and then inserts the
// data in the tokens table
func (m MockTokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	return nil, nil
}

// Insert() adds the data for a specific token to the tokens table
func (m MockTokenModel) Insert(token *Token) error {
	return nil
}

// DeleteAllForUser() deletes all tokens for a specific user and scope
func (m MockTokenModel) DeleteAllForUser(scope string, userID int64) error {
	return nil
}
