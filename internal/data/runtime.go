package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Define an error that our UnmarshalJSON() method can return if we're unable to parse
// or convert the JSON string successfully
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type Runtime int32

// We implement a MarshalJSON() method on the Runtime type so that it satisfies the
// json.Marshaler interface. This should return the JSON-encoded value for the movie
// runtime (in our case, it will return a string in the format "<runtime> mins").
func (r Runtime) MarshalJSON() ([]byte, error) {
	// Generate a string containing the movie runtime in the required format
	jsonValue := fmt.Sprintf("%d mins", r)

	// Wrap string in double quotes
	quotedJsonValue := strconv.Quote(jsonValue)

	// Convert the quoted string value to a byte slice and return it
	return []byte(quotedJsonValue), nil
}

// Implement a UnmarshalJSON() method on the Runtime type so that it satisfies the
// json.Unmarshaler interface. IMPORTANT: Because UnmarshalJSON() needs to modify the
// receiver (our Runtime type), we must use a pointer receiver for this to work
// correctly. Otherwise, we will only be modifying a copy (which is then discarded when
// this method returns)
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// We expect that the incoming JSON value will be a string in the format
	// "<runtime> mins", and the first thing we need to do is remove the surrounding
	// double-quotes from this string. If we can't unquote it, then we return the
	// ErrInvalidRuntimeFormat error.
	unquotedJsonValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Split the string to isolate the part containing the number of minutes
	parts := strings.Split(unquotedJsonValue, " ")

	// Sanity check the parts of the string to make sure it was in the expected format
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// Parse the string containing the number of minutes into an int32
	minutes, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Convert the int32 to a Runtime type and assign this to the receiver
	*r = Runtime(minutes)

	return nil
}
