package data

// Define a mock of the `UserModel` struct type
type MockUserModel struct{}

// Inserts a new record in the `users` table
func (m MockUserModel) Insert(user *User) error {
	return nil
}

// Fetches a specific record from the `users` table by given email
func (m MockUserModel) GetByEmail(email string) (*User, error) {
	return nil, nil
}

// Updates a specific record from the `users` table
func (m MockUserModel) Update(user *User) error {
	return nil
}

// Fetch user linked to given token
func (m MockUserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	return nil, nil
}
