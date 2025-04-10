package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const GitlabApiEndpoint = "https://gitlab.com/api/v4"

type GitlabService struct {
	gitlabApiEndpoint string
	token             string
	httpClient        *http.Client
	// rateLimitProjectAPI *rate.Limiter TODO: implement rate limiting
	// rateLimitGroupAPI   *rate.Limiter TODO: implement rate limiting
}

// getNextLink parses the link header and returns the next page url
// gitlab API returns a link header with the next page url
// we need to parse this header to get the next page url
func getNextLink(linkHeader string) string {
	// linkHeader has the format: url1; rel="first", url2; rel="prev", url3; rel="next", url4; rel="last"
	// so we split the string with the , separator
	// and take the first element
	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		// link is formatted like this:
		// <https://gitlab.com/api/v4/groups/1234/projects?page=2&per_page=100>; rel="next"
		// we only need the next page url
		// so we split the string with the ; separator
		// and take the first element
		if strings.Contains(link, "rel=\"next\"") {
			nextPageUrl := strings.Split(link, ";")[0]
			// remove the < and > characters
			nextPageUrl = strings.Trim(nextPageUrl, "<> ")
			return nextPageUrl
		}
	}
	return ""
}

// NewGitlabService returns a new GitlabService
func NewGitlabService() *GitlabService {
	gs := &GitlabService{
		gitlabApiEndpoint: GitlabApiEndpoint,
		token:             os.Getenv("GITLAB_TOKEN"),
		httpClient:        &http.Client{},
		// // implement rate limiting https://docs.gitlab.com/ee/administration/settings/rate_limit_on_projects_api.html
		// rateLimitProjectAPI: rate.NewLimiter(rate.Every(600*time.Second), 2000),
		// // https://docs.gitlab.com/ee/administration/settings/rate_limit_on_groups_api.html
		// rateLimitGroupAPI: rate.NewLimiter(rate.Every(60*time.Second), 200),
	}
	return gs
}

// SetGitlabEndpoint sets the Gitlab API endpoint
// default: https://gitlab.com/v4/api/
func (r *GitlabService) SetGitlabEndpoint(gitlabApiEndpoint string) {
	r.gitlabApiEndpoint = gitlabApiEndpoint
}

// SetToken sets the Gitlab API token
// default: GITLAB_TOKEN env variable
func (r *GitlabService) SetToken(token string) {
	r.token = token
}

// SetHttpClient sets the http client
// default: http.Client{}
func (r *GitlabService) SetHttpClient(httpClient *http.Client) {
	if httpClient != nil {
		r.httpClient = httpClient
	}
}

// Get sends a GET request to the Gitlab API to the given path
func (r *GitlabService) get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", r.token)
	req.Header.Set("Content-Type", "application/json")
	return r.httpClient.Do(req)
}

// Post sends a POST request to the Gitlab API to the given path
func (r *GitlabService) post(url string) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", r.token)
	return r.httpClient.Do(req)
}

// GetGroup returns the gitlab group from the given ID
func (s *GitlabService) GetGroup(groupID int) (res GitlabGroup, err error) {
	url := fmt.Sprintf("%s/groups/%d", s.gitlabApiEndpoint, groupID)
	resp, err := s.get(url)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}
	// Unmarshal the response
	if err = json.Unmarshal(body, &res); err != nil {
		// If the response is an error message, unmarshal it
		return res, UnmarshalErrorMessage(body)
	}
	return res, err
}

// GetProject returns informations of the project that matches the given ID
func (s *GitlabService) GetProject(projectID int) (res GitlabProject, err error) {
	url := fmt.Sprintf("%s/projects/%d", s.gitlabApiEndpoint, projectID)
	resp, err := s.get(url)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}
	// Unmarshal the response
	if err = json.Unmarshal(body, &res); err != nil {
		// If the response is an error message, unmarshal it
		return res, UnmarshalErrorMessage(body)
	}
	return res, err
}
