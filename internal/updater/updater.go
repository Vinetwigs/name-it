// Package updater checks the public GitHub releases feed for a newer build of
// Name It. It is intentionally tiny and best-effort: any network or parsing
// failure is reported as an error and never blocks the application.
package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// repo is the GitHub repository that hosts the release feed.
const repo = "Vinetwigs/name-it"

// requestTimeout keeps a slow or unreachable network from holding the app open.
const requestTimeout = 8 * time.Second

// Release describes an available update as understood by the UI.
type Release struct {
	Version string `json:"version"`
	URL     string `json:"url"`
	Notes   string `json:"notes"`
}

// releaseResponse mirrors the subset of the GitHub API response we care about.
type releaseResponse struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
}

// Check queries the latest published release for the repository. It returns a
// Release only when the published version is strictly newer than current;
// otherwise it returns (nil, nil), meaning the app is up to date.
func Check(current string) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// GitHub rejects requests without a User-Agent and prefers the v3 media type.
	request.Header.Set("User-Agent", "name-it-updater")
	request.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: requestTimeout}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update check returned %s", response.Status)
	}

	var latest releaseResponse
	if err := json.NewDecoder(response.Body).Decode(&latest); err != nil {
		return nil, err
	}

	published := strings.TrimPrefix(strings.TrimSpace(latest.TagName), "v")
	if published == "" || !IsNewer(published, current) {
		return nil, nil
	}

	return &Release{
		Version: "v" + published,
		URL:     latest.HTMLURL,
		Notes:   strings.TrimSpace(latest.Body),
	}, nil
}

// IsNewer reports whether candidate is a higher semantic version than current.
// Both inputs are expected without a leading "v" and may omit minor or patch
// components, which are treated as zero.
func IsNewer(candidate, current string) bool {
	a := parse(candidate)
	b := parse(current)

	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}

	return false
}

// parse turns a loose version string into a fixed [major, minor, patch] tuple.
// Unparsable or missing components fall back to zero.
func parse(v string) [3]int {
	var out [3]int

	// Strip anything from the first dash or plus so that suffixes such as
	// "2.1.0-beta.1" or "2.1.0+build" still compare on the numeric core.
	v = strings.TrimSpace(strings.TrimPrefix(v, "v"))
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}

	for i, part := range strings.Split(v, ".") {
		if i >= len(out) {
			break
		}
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return out
		}
		out[i] = n
	}

	return out
}
