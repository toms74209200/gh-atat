package output

import (
	"fmt"
	"io"
	"os"
)

// Println writes a message to stdout and optionally to an additional writer.
// This function is designed for testability and mirrors the behavior of the reference implementation.
//
// Parameters:
//   - message: The message to write
//   - writer: Optional additional writer (can be nil)
//
// Returns an error if writing to the additional writer fails.
// Errors writing to stdout are logged to stderr but do not cause the function to fail.
func Println(message string, writer io.Writer) error {
	// Write to stdout
	if _, err := fmt.Fprintln(os.Stdout, message); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to stdout: %v\n", err)
	}

	// Write to additional writer if provided
	if writer != nil {
		if _, err := fmt.Fprintln(writer, message); err != nil {
			return err
		}
	}

	return nil
}
