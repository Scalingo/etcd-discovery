package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Run("hostname should be set", func(t *testing.T) {
		require.NotNil(t, hostname)
	})
}
