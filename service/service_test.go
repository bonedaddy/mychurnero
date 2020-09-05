package service

import (
	"context"
	"os"
	"testing"

	"github.com/bonedaddy/mychurnero/config"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("test_path")
	})
	srv, err := New(context.Background(), config.DefaultConfig())
	require.NoError(t, err)
	srv.MC()
	srv.DB()
	srv.Context()
	err = srv.Close()
	require.NoError(t, err)
}
