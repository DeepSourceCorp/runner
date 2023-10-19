package main

import (
	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcecorp/runner/rqlite"
)

func GetDB(c *config.Config) (interface{}, error) {
	conn, err := rqlite.Connect(c.RQLite.Host, c.RQLite.Port)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
