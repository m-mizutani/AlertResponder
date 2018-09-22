package lib_test

import (
	"testing"

	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretManager(t *testing.T) {
	cfg := lib.LoadTestConfig()

	values, err := lib.GetSecretValues(cfg.SecretID, cfg.AwsRegion)
	require.NoError(t, err)
	assert.NotEqual(t, "", values.GitHubToken)
	assert.NotEqual(t, "", values.GitHubToken)
}
