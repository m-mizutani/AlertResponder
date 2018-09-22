package lib_test

import (
	"fmt"
	"testing"

	"github.com/m-mizutani/AlertResponder/lib"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubIssue(t *testing.T) {
	cfg := lib.LoadTestConfig()
	ghe, err := lib.NewGitHub(cfg.GithubEndpoint, cfg.GithubRepository, cfg.GithubToken)

	require.NotNil(t, ghe)
	assert.Nil(t, err)

	u1 := uuid.NewV4().String()
	u2 := uuid.NewV4().String()

	title := "TITLE:" + u1
	body := "BODY:" + u1
	comment := "COMMENT:" + u2

	// Create an issue
	issue, err := ghe.NewIssue(title, body)
	require.NotNil(t, issue)
	assert.Nil(t, err)

	// Get the issue by ApiURL
	issue2, err := ghe.GetIssue(issue.ApiURL)
	require.NotNil(t, issue2)
	assert.Nil(t, err)

	assert.Equal(t, title, issue2.Title)
	assert.Equal(t, body, issue2.Content)

	// Append body of the issue
	err = issue.AppendContent("Test, Test, Test!!")
	require.Nil(t, err)
	expectedNewBody := fmt.Sprintf("%s\n\n- - - - - - - - - -\n\nTest, Test, Test!!", body)

	issue3, err := ghe.GetIssue(issue.ApiURL)
	require.Nil(t, err)
	assert.NotEqual(t, body, issue3.Content)
	assert.Equal(t, expectedNewBody, issue3.Content)

	// Add/Get comments
	comments, err := issue.FetchComments()
	require.Equal(t, 0, len(comments))

	commentResp, err := issue.AddComment(comment)
	require.Nil(t, err)
	assert.NotEqual(t, "", commentResp.Body)

	comments, err = issue.FetchComments()
	require.Equal(t, 1, len(comments))
	require.Equal(t, comment, comments[0])

}
