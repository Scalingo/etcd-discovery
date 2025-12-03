package service

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Run("hostname should be set", func(t *testing.T) {
		require.NotNil(t, hostname)
	})

	t.Run("hostname should be set", func(t *testing.T) {
		assert.Equal(t, "[etcd-discovery] ", logger.Prefix())
		assert.Equal(t, log.LstdFlags, logger.Flags())
	})
}
