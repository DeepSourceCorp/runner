package config

import (
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"strings"
)

type SAML struct {
	Enabled     bool             `json:"enabled"`
	Certificate *tls.Certificate `json:"-"`
	MetadataURL url.URL          `json:"metadataUrl"`
}

func (s *SAML) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type T struct {
		Enabled        bool   `yaml:"enabled"`
		MetadataURLStr string `yaml:"metadataUrl"`
		CertificateStr string `yaml:"certificate"`
		Key            string `yaml:"key"`
	}
	var v T
	if err := unmarshal(&v); err != nil {
		return err
	}
	if !v.Enabled {
		return nil
	}
	metadataURL, err := url.Parse(v.MetadataURLStr)
	if err != nil {
		return err
	}
	v.CertificateStr = strings.TrimSpace(v.CertificateStr)
	v.Key = strings.TrimSpace(v.Key)
	cert, err := tls.X509KeyPair([]byte(v.CertificateStr), []byte(v.Key))
	if err != nil {
		return err
	}
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return err
	}
	s.Enabled = v.Enabled
	s.MetadataURL = *metadataURL
	s.Certificate = &cert
	return nil
}
