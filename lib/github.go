package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type GitHub struct {
	endpoint   string
	repository string
	token      string
}

type GitHubIssue struct {
	HtmlURL string `json:"html_url"`
	ApiURL  string `json:"url"`
	Title   string `json:"title"`
	Content string `json:"body"`
	github  *GitHub
}

type GitHubIssueComment struct {
	HtmlURL  string `json:"html_url"`
	ApiURL   string `json:"url"`
	IssueURL string `json:"issue_url"`
	Body     string `json:"body"`
}

//
// NewGitHub returns a GitHub accessor. Currently it's based on AWS KMS decryption for
// secret value (GitHub token). It's planned to be replaced with AWS SecretManager
//
func NewGitHub(endpoint, repository, token string) (*GitHub, error) {
	g := GitHub{
		endpoint:   endpoint,
		repository: repository,
		token:      token,
	}

	return &g, nil
}

//
// NewIssue creates an new issue on github with title and issue's body
//
func (x *GitHub) NewIssue(title, content string) (*GitHubIssue, error) {
	issueReq := struct {
		TItle string `json:"title"`
		Body  string `json:"body"`
	}{
		TItle: title,
		Body:  content,
	}

	binData, err := json.Marshal(issueReq)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to create JSON message")
	}

	client := &http.Client{}
	url := fmt.Sprintf("%s/repos/%s/issues", x.endpoint, x.repository)

	req, err := http.NewRequest("POST", url, bytes.NewReader(binData))
	if err != nil {
		return nil, errors.Wrap(err, "Fail to build a request to create github issue")
	}
	// log.Println("req = ", req)
	// log.Println("token len = ", len(x.token))
	req.Header.Add("Authorization", fmt.Sprintf("token %s", x.token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to create a github issue")
	} else if resp.StatusCode != 201 {
		return nil, errors.New(fmt.Sprintf("Fail to create a github issue, code: %d",
			resp.StatusCode))
	}

	return x.respToIssue(resp, nil)
}

func (x *GitHub) respToIssue(resp *http.Response, issue *GitHubIssue) (*GitHubIssue, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to read body data of creating issue")
	}

	if issue == nil {
		issue = &GitHubIssue{github: x}
	}

	err = json.Unmarshal(body, &issue)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to parse a result of creating issue")
	}
	return issue, nil
}

//
// GetIssue returns existing an issue from github by URL for API
//
func (x *GitHub) GetIssue(apiURL string) (*GitHubIssue, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to build a request to get github issue")
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", x.token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to get a github issue")
	} else if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Fail to get a github issue, code: %d",
			resp.StatusCode))
	}

	return x.respToIssue(resp, nil)
}

//
// AppendContent appends additional body to existing issue
//
func (x *GitHubIssue) AppendContent(content string) error {
	tempalte := "%s\n\n- - - - - - - - - -\n\n%s"
	newBody := fmt.Sprintf(tempalte, x.Content, content)
	updateReq := struct {
		Body string `json:"body"`
	}{
		Body: newBody,
	}

	binData, err := json.Marshal(updateReq)
	if err != nil {
		return errors.Wrap(err, "Fail to create JSON message")
	}

	client := &http.Client{}
	req, err := http.NewRequest("PATCH", x.ApiURL, bytes.NewReader(binData))
	if err != nil {
		return errors.Wrap(err, "Fail to build a request to create github issue")
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", x.github.token))

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "Fail to patch the issue")
	} else if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Fail to patch the issue, code: %d",
			resp.StatusCode))
	}

	x.github.respToIssue(resp, x)
	return nil
}

func (x *GitHubIssue) AddComment(comment string) (*GitHubIssueComment, error) {
	commentData := struct {
		Body string `json:"body"`
	}{
		Body: comment,
	}
	binData, err := json.Marshal(commentData)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/comments", x.ApiURL),
		bytes.NewReader(binData))
	if err != nil {
		return nil, errors.Wrap(err, "Fail to build a request to create github issue")
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", x.github.token))

	resp, err := client.Do(req)

	if err != nil {
		return nil, errors.Wrap(err, "Fail to post a comment")
	} else if resp.StatusCode != 201 {
		return nil, errors.New(fmt.Sprintf("Fail to post a comment, code: %d",
			resp.StatusCode))
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	c := GitHubIssueComment{}
	err = json.Unmarshal(respData, &c)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to parse issue comment result")
	}
	// fmt.Println(c)

	return &c, nil
}

func (x *GitHubIssue) FetchComments() ([]string, error) {
	results := []string{}

	url := fmt.Sprintf("%s/comments", x.ApiURL)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return results, errors.Wrap(err, "Fail to build get request of comment")
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", x.github.token))

	resp, err := client.Do(req)
	if err != nil {
		return results, errors.Wrap(err, "Fail to get the issues")
	} else if resp.StatusCode != 200 {
		return results, errors.New(fmt.Sprintf("Fail to patch the issue, code: %d",
			resp.StatusCode))
	}

	binData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return results, errors.Wrap(err, "Fail to read body data of github comments")
	}
	type comment struct {
		Body string `json:"body"`
	}
	comments := make([]comment, 0)
	err = json.Unmarshal(binData, &comments)
	if err != nil {
		return results, errors.Wrap(err, "Fail to parse json of github comments")
	}

	for _, c := range comments {
		results = append(results, c.Body)
	}
	return results, nil
}
