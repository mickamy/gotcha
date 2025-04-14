package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"github.com/mickamy/gotcha/internal/config"
)

func TestLoadByPath(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadByPath("./testdata/.gotcha.yaml")
	require.NoError(t, err)
	assert.DeepEqual(t, []string{"./..."}, cfg.Include)
	assert.DeepEqual(t, []string{"vendor/", "mocks/"}, cfg.Exclude)
	assert.DeepEqual(t, []string{"-v"}, cfg.TestFlags)
}
