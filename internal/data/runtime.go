package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

// A custom Runtime type, which has the underlying type int32
type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	// Generate a string containing the movie runtime in the required format.
	jsonValue := fmt.Sprintf("%d mins", r)

	// Use the strconv.Quote() function on the string to wrap it in double quotes.
	// It needs to be surrounded by double quotes in order to be a valid *JSON string*.
	quotedJSONValue := strconv.Quote(jsonValue)

	// Convert the quoted string value to a byte slice and return it.
	return []byte(quotedJSONValue), nil
}

// UnmarshalJSON() method on the Runtime type so that it satisfies the json.Unmarshaler
// interface.
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// expected string format: "<runtime> mins". remove the surrounding
	// double-quotes from this string. if we cant remove, return the
	// ErrInvalidRuntimeFormat err.
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Split the string to isolate the part containing the number.
	parts := strings.Split(unquotedJSONValue, " ")

	// Sanity check the pars of the string to make sure it was in the expected format.
	// If it isnt, return the ErrInvalidRuntimeFormat err.
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// Parse the string containing the number into an int32. If this fails, return the
	// ErrInvalidRuntimeFormat err.
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Convert the int32 to a Runtime type and assign this to the receiver.
	*r = Runtime(i)

	return nil
}
