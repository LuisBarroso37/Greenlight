package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/LuisBarroso37/Greenlight/internal/data"
	"github.com/LuisBarroso37/Greenlight/internal/validator"
)

// Handler for the "POST /v1/movies" endpoint
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the
	// request body. This struct will be our *target decode destination*
	// (The field names and types in the struct are a subset of the Movie struct)
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	// Read request body and decode it into the input struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the values from the input struct to a new Movie struct
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// Initialize a new Validator instance
	v := validator.New()

	// Perform validation checks
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Create a movie record in the database and update the movie struct with the system-generated information
	err = app.models.Movie.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// Write a JSON response with a 201 Created status code, the movie data in the
	// response body, and the Location header
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Handler for the "GET /v1/movies/:id" endpoint
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract id parameter from request URL parameters
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch movie by given id
	movie, err := app.models.Movie.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// Write the fetched movie record in a JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Handler for the "PUT /v1/movies/:id" endpoint
func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract id parameter from request URL parameters
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch movie by given id
	movie, err := app.models.Movie.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// We use pointers so that we get a nil value when decoding these values from JSON.
	// This way we can check if a user has provided the key/value pair in the JSON or not.
	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	// Read request body and decode it into the input struct
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the values from the input struct to the fetched movie if they exist
	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	// Initialize a new Validator instance
	v := validator.New()

	// Perform validation checks
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Update movie
	err = app.models.Movie.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// Write the updated movie record in a JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Handler for the "DEL /v1/movies/:id" endpoint
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract id parameter from request URL parameters
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete movie with given id
	err = app.models.Movie.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// Write the updated movie record in a JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Handler for the "GET /v1/movies" endpoint
func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	// Initialize a new validator instance
	v := validator.New()

	// Call r.URL.Query() to get the url.Values map containing the query string data
	queryString := r.URL.Query()

	// Parse query string values and store them in `input` struct
	input.Title = app.readString(queryString, "title", "")
	input.Genres = app.readCSV(queryString, "genres", []string{})
	input.Page = app.readInt(queryString, "page", 1, v)
	input.PageSize = app.readInt(queryString, "page_size", 20, v)

	// Extract the sort query string value, falling back to "id" if it is not provided
	// by the client (which will imply a ascending sort on movie ID)
	input.Sort = app.readString(queryString, "sort", "id")

	// Add the supported sort values for this endpoint to the sort safelist
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	// Check the Validator instance for any errors and use the failedValidationResponse()
	// helper to send the client a response if necessary
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Fetch all movies that
	movies, metadata, err := app.models.Movie.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Write the list of movies in a JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
