package figma

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

// BlameResult is the outcome of finding which version introduced the current
// text at a node. Versions are searched newest-first backward from toVersion.
type BlameResult struct {
	// IntroducedIn is the oldest version whose text matches toVersion.
	IntroducedIn api.Version `json:"introducedIn"`
	// Previous is the version immediately older than IntroducedIn (the last
	// version with the prior text). Nil when the change predates visible history.
	Previous *api.Version `json:"previous,omitempty"`
	// Changes is the text diff Previous -> IntroducedIn. Empty when the change
	// predates visible history.
	Changes extract.TextOutput `json:"changes"`
	// PredatesHistory is true when the oldest visible version already has the
	// target text, so the introducing version is outside the fetched history.
	PredatesHistory bool `json:"predatesHistory"`
}

// blameFetcher is the dependency FindTextChange needs: the version list and a
// version's text nodes. Injecting it keeps the binary search testable without
// HTTP and without coupling to the generated document types.
type blameFetcher interface {
	Versions() ([]api.Version, error)
	TextAt(versionID string) ([]extract.TextNode, error)
}

// blameClientFetcher adapts a Figma Client + file/node scope to blameFetcher.
type blameClientFetcher struct {
	client  *Client
	fileID  string
	nodeIDs []string
}

func (f *blameClientFetcher) Versions() ([]api.Version, error) {
	return FetchAllVersions(f.client, f.fileID)
}

func (f *blameClientFetcher) TextAt(versionID string) ([]extract.TextNode, error) {
	doc, err := FetchDocument(f.client, f.fileID, f.nodeIDs, versionID, "")
	if err != nil {
		return nil, err
	}
	return extract.ExtractTextNodes(doc), nil
}

// FetchAllVersions pages through a file's version history (newest-first) until
// the end. It uses the largest page size to minimise requests.
func FetchAllVersions(client *Client, fileID string) ([]api.Version, error) {
	var all []api.Version
	after := ""
	for {
		u, err := BuildVersionsURL(fileID, VersionsQuery{PageSize: 50, After: after})
		if err != nil {
			return nil, err
		}
		resp, err := FetchVersions(client, u)
		if err != nil {
			return nil, err
		}
		if len(resp.Versions) == 0 {
			break
		}
		all = append(all, resp.Versions...)
		if resp.Pagination.NextPage == nil {
			break
		}
		// The next (older) page is reached via --after <oldest id in this page>.
		after = resp.Versions[len(resp.Versions)-1].Id
	}
	return all, nil
}

// FindTextChange searches the version history backward from toVersion to find
// the version that introduced toVersion's current text at the given nodeIDs.
// It binary-searches the history: each probe fetches a candidate version and
// compares its text to toVersion's. If fromVersion is set, the search stops
// there; otherwise it searches to the oldest visible version.
func FindTextChange(client *Client, fileID string, nodeIDs []string, toVersion, fromVersion string) (BlameResult, error) {
	return findTextChange(&blameClientFetcher{client: client, fileID: fileID, nodeIDs: nodeIDs}, toVersion, fromVersion)
}

func findTextChange(f blameFetcher, toVersion, fromVersion string) (BlameResult, error) {
	versions, err := f.Versions()
	if err != nil {
		return BlameResult{}, err
	}

	toIdx := indexByVersion(versions, toVersion)
	if toIdx < 0 {
		return BlameResult{}, fmt.Errorf("version %s not found in file history", toVersion)
	}

	hi := len(versions) - 1
	if fromVersion != "" {
		fromIdx := indexByVersion(versions, fromVersion)
		if fromIdx < 0 {
			return BlameResult{}, fmt.Errorf("--from version %s not found in file history", fromVersion)
		}
		if fromIdx <= toIdx {
			return BlameResult{}, fmt.Errorf("--from must be older than --to")
		}
		hi = fromIdx
	}

	targetText, err := f.TextAt(toVersion)
	if err != nil {
		return BlameResult{}, err
	}

	// If the oldest version in range already has the target text, the change
	// predates the visible history.
	oldestText, err := f.TextAt(versions[hi].Id)
	if err != nil {
		return BlameResult{}, err
	}
	if extract.TextEqual(oldestText, targetText) {
		return BlameResult{IntroducedIn: versions[hi], PredatesHistory: true}, nil
	}

	// Binary search for the largest index k in [toIdx, hi] whose text matches.
	// Invariant: versions[lo] matches (lo starts at toIdx, the target itself).
	lo := toIdx
	for lo < hi {
		mid := lo + (hi-lo+1)/2 // round up so lo advances and the loop terminates
		midText, err := f.TextAt(versions[mid].Id)
		if err != nil {
			return BlameResult{}, err
		}
		if extract.TextEqual(midText, targetText) {
			lo = mid // mid has the new text -> introducer is mid or older
		} else {
			hi = mid - 1 // mid has the old text -> introducer is newer
		}
	}

	introduced := versions[lo]
	// A previous version with old text is guaranteed: the oldest-in-range check
	// above returned early when hi matched, so the boundary sits below hi.
	previous := versions[lo+1]

	prevText, err := f.TextAt(previous.Id)
	if err != nil {
		return BlameResult{}, err
	}
	introText, err := f.TextAt(introduced.Id)
	if err != nil {
		return BlameResult{}, err
	}
	changes := extract.DiffText(prevText, introText)
	return BlameResult{IntroducedIn: introduced, Previous: &previous, Changes: changes}, nil
}

func indexByVersion(versions []api.Version, id string) int {
	for i, v := range versions {
		if v.Id == id {
			return i
		}
	}
	return -1
}
