// Package diff contains pure design-diff use cases.
package diff

import (
	"fmt"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/extract"
)

// Version is the history data needed to explain a text change.
type Version struct {
	ID        string
	CreatedAt time.Time
	User      string
}

// TextHistory supplies ordered version history and text at a version.
type TextHistory interface {
	Versions() ([]Version, error)
	TextAt(versionID string) ([]extract.TextNode, error)
}

// BlameResult explains which version introduced a target version's text.
type BlameResult struct {
	IntroducedIn    Version            `json:"introducedIn"`
	Previous        *Version           `json:"previous,omitempty"`
	Changes         extract.TextOutput `json:"changes"`
	PredatesHistory bool               `json:"predatesHistory"`
}

// FindTextChange finds the oldest version with the target version's text.
func FindTextChange(history TextHistory, toVersion, fromVersion string) (BlameResult, error) {
	versions, err := history.Versions()
	if err != nil {
		return BlameResult{}, err
	}

	toIndex := indexByVersion(versions, toVersion)
	if toIndex < 0 {
		return BlameResult{}, fmt.Errorf("version %s not found in file history", toVersion)
	}

	high := len(versions) - 1
	if fromVersion != "" {
		fromIndex := indexByVersion(versions, fromVersion)
		if fromIndex < 0 {
			return BlameResult{}, fmt.Errorf("--from version %s not found in file history", fromVersion)
		}
		if fromIndex <= toIndex {
			return BlameResult{}, fmt.Errorf("--from must be older than --to")
		}
		high = fromIndex
	}

	targetText, err := history.TextAt(toVersion)
	if err != nil {
		return BlameResult{}, err
	}
	oldestText, err := history.TextAt(versions[high].ID)
	if err != nil {
		return BlameResult{}, err
	}
	if extract.TextEqual(oldestText, targetText) {
		return BlameResult{IntroducedIn: versions[high], PredatesHistory: true}, nil
	}

	low := toIndex
	for low < high {
		middle := low + (high-low+1)/2
		text, err := history.TextAt(versions[middle].ID)
		if err != nil {
			return BlameResult{}, err
		}
		if extract.TextEqual(text, targetText) {
			low = middle
			continue
		}
		high = middle - 1
	}

	introduced := versions[low]
	previous := versions[low+1]
	previousText, err := history.TextAt(previous.ID)
	if err != nil {
		return BlameResult{}, err
	}
	introducedText, err := history.TextAt(introduced.ID)
	if err != nil {
		return BlameResult{}, err
	}
	return BlameResult{
		IntroducedIn: introduced,
		Previous:     &previous,
		Changes:      extract.DiffText(previousText, introducedText),
	}, nil
}

func indexByVersion(versions []Version, id string) int {
	for index, version := range versions {
		if version.ID == id {
			return index
		}
	}
	return -1
}
