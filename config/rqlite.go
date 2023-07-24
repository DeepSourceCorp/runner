package config

import (
	"os"
	"strconv"
)

type RQLite struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (r *RQLite) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type T struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
	var v T
	if err := unmarshal(&v); err != nil {
		return err
	}
	if os.Getenv("RQLITE_HOST") != "" {
		r.Host = os.Getenv("RQLITE_HOST")
	} else {
		r.Host = v.Host
	}
	if os.Getenv("RQLITE_PORT") != "" {
		p, err := strconv.Atoi(os.Getenv("RQLITE_PORT"))
		if err != nil {
			return err
		}
		r.Port = p
	} else {
		r.Port = v.Port
	}
	return nil
}
