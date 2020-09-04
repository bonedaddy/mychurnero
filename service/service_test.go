package service

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testNetRPC    = "http://127.0.0.1:6061/json_rpc"
	testNetWallet = "testnetwallet123"
)

func TestService(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("test_path")
	})
	srv, err := New(context.Background(), 1, "test_path", testNetWallet, testNetRPC)
	require.NoError(t, err)
	srv.MC()
	srv.DB()
	srv.Context()
	err = srv.Close()
	require.NoError(t, err)
}
