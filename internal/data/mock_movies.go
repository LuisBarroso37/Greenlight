package data

// Define a mock of the `MovieModel` struct type
type MockMovieModel struct{}

// Inserts a new record in the `movies` table
func (m MockMovieModel) Insert(movie *Movie) error {
	return nil
}

// Fetches a specific record from the `movies` table
func (m MockMovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

// Updates a specific record from the `movies` table
func (m MockMovieModel) Update(movie *Movie) error {
	return nil
}

// Deletes a specific record from the `movies` table
func (m MockMovieModel) Delete(id int64) error {
	return nil
}

// Fetches all movie records from the `movies` table
func (m MockMovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	return nil, Metadata{}, nil
}
