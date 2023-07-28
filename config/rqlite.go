package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type RQLite struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (r *RQLite) ParseFromEnv() error {
	if os.Getenv("RQLITE_HOST") == "" || os.Getenv("RQLITE_PORT") == "" {
		return errors.New("config: failed to parse rqlite from env")
	}

	r.Host = os.Getenv("RQLITE_HOST")
	port, err := strconv.Atoi(os.Getenv("RQLITE_PORT"))
	if err != nil {
		return fmt.Errorf("config: failed to parse rqlite port from env: %v", err)
	}
	r.Port = port

	return nil
}
