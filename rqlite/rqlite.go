package rqlite

import (
	"fmt"
	"net/http"

	"github.com/rqlite/gorqlite"
	"golang.org/x/exp/slog"
)

var conn *gorqlite.Connection

func Connect(host string, port int) (*gorqlite.Connection, error) {
	if conn != nil {
		return conn, nil
	}
	dsn := fmt.Sprintf("http://%s:%d/?disableClusterDiscovery=true", host, port)
	var err error
	conn, err = gorqlite.Open(dsn)
	if err != nil {
		slog.Error("failed to connect to rqlite", slog.Any("error", err))
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprint("http://", host, ":", port, "/readyz"), http.NoBody)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		slog.Error("failed to connect to rqlite, ping failed", slog.Any("status", res.StatusCode))
		return nil, fmt.Errorf("rqlite is not ready")
	}
	return conn, nil
}
