// Package version holds the single source of truth for the Name It release string.
package version

// Version follows semantic versioning, without the leading "v".
const Version = "2.1.1"

// Display returns the version as shown in the UI and used in comparisons
// against published Git tags.
func Display() string {
	return "v" + Version
}
