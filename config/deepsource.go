package config

import (
	"crypto/rsa"
	"net/url"

	"github.com/golang-jwt/jwt"
)

type DeepSource struct {
	Host      url.URL
	PublicKey *rsa.PublicKey
}

func (d *DeepSource) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type T struct {
		Host      string `yaml:"host"`
		PublicKey string `yaml:"publicKey"`
	}
	var v T
	if err := unmarshal(&v); err != nil {
		return err
	}
	host, err := url.Parse(v.Host)
	if err != nil {
		return err
	}
	d.Host = *host
	pk, err := jwt.ParseRSAPublicKeyFromPEM([]byte(v.PublicKey))
	if err != nil {
		return err
	}
	d.PublicKey = pk
	return nil
}
