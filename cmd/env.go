package cmd

import "os"

// getFigmaToken returns FIGMA_ACCESS_TOKEN.
// If the variable is not set or empty, it returns an error.
func getFigmaToken() (string, error) {
	token := os.Getenv("FIGMA_ACCESS_TOKEN")
	if token == "" {
		return "", &errTokenNotSet{}
	}
	return token, nil
}

type errTokenNotSet struct{}

func (e *errTokenNotSet) Error() string {
	return "FIGMA_ACCESS_TOKEN environment variable not set"
}
