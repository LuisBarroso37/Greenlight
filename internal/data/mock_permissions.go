package data

// Define a mock of the `TokenModel` struct type
type MockPermissionsModel struct{}

// This method returns all permission codes for a specific user in a
// Permissions slice
func (m MockPermissionsModel) GetAllForUser(userID int64) (Permissions, error) {
	return nil, nil
}

// Add the provided permission codes for a specific user
func (m MockPermissionsModel) AddForUser(userID int64, codes ...string) error {
	return nil
}
