package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestGithub_UnmarshalYAML(t *testing.T) {
	input := `
appId: "app-id"
clientId: "client-id"
clientSecret: "client-secret"
webhookSecret: "webhook-secret"
host: "https://github.com"
apiHost: "https://api.github.com"
slug: "slug-1"
privateKey: |
  -----BEGIN RSA PRIVATE KEY-----
  MIIEpQIBAAKCAQEA6BzjvvccBOWR6iCaNrLG2GRSpEoZo12sIwm1EWgIk6m6feFP
  nIwqDVT50gpdINSCiLGdisSa5UO5S6PnJ+1kwn1TVBwIngiAGFFO6dUW4COmmVKu
  pKaTS1xotIvOXAfcMB+7QOP8uLEuAyQa/+YiqiFJFgqfALEaU76Unw195Q4WKLtk
  NgybvZfieZRsj06wXFUj/LRUlYiL5oy/gwGV6t7P9VeAi4tSmecgfrDakpd8aYxn
  GBuONnNkxwWLbbjNfI/MQJm7nRp/WvdL2wwzmWCCUo8fx4+Y7rHf4ZPRtnIJcVHA
  uosp/4ZyIrtBEeT2GOBkhQeL5dIZC9FFFAUqxwIDAQABAoIBAQDVXdktLlKfXbjo
  E9gu9+A6At7FDyjKN82I19+OhKd9tcQs+vUH3wC5CKgtIEHDcBYeOcesTFZm8f5f
  Pee7mEnLTxFOfAaf3wiBUhzMbol8uMjooEzSJh24ZNYLQYkMqF0MD98+I1WpIZY+
  ZO481fx/j+FzVYgcRrEA0mwkWW6lIrGCcYLKjUi7Z+/dWJM8TT0n4AdG47WJRoi8
  JMWf4Z+iPb4Styx9GUoJ60D48qG/1ctOoc927TL7UbPSrTE+nf3nMZGWmKZe8nT5
  PWaKysfX3k7yuFW4DUv6dK546J15oDPsxKdeSUQzxdo2awunZ6HJwoBbm5cAXvNH
  h5jNXJ/hAoGBAPpTTq/kgmfdWbkOuH0YIt0VwNVKVCTZD8nvPGyRKUf4F03w3pBB
  XIah8LcHBdQueDqOFjGVgLk+Roo0cEqf46gGOCnb9AImF+L6o6ij2q2q/mfN4tks
  WgF6Tj9SgI1WR+BC1IJP7UcTgpPXGjabk0mgK5+vYJGcUu5JmRful7/jAoGBAO1f
  5Anzrdl65o8+hxTJ1Abbx1rJYLHQcYrMM5uOMQiiCP+CJZ68Af6xULUoGq/3EVte
  Z/THyDCz9g6veXq/3SZkYM9IcF11ehY//L8+ZRDF2hgiXLm2RguMhds17VGFsCm0
  U09vfgoYy379MnIwEjvbW8/OV/HsqjTPiIW3YRbNAoGACZZTNy1bSTsTCqFjs3bP
  LwR8RC76lgayMhu1hrrwh88apWOKQqAeORHOtFPSh1PYSvXSJ8gADBg0f2qOumzx
  PSgv0nqYF9T5qTnMNtM/ttMLt1INVB/8un3CrW4tejxJuG8W0H7bKZO3to3QdTL0
  Kye1RAJlgm4oRvQOpvn+Wd0CgYEAmU7gQgku1BJLTGKu7Z84oEFb7Oe42r7sRh+C
  iUn5o0C7nQIad/2nMC6nGIlRSyq//AnqDC7nvYTNO0jbpYq7MyuLVvTLFaFk+2/S
  NlX/Aik2pXWz+4GclaLpZN3ca1VzpEvBrsEsXysKavbumM8xR5VyI7F6HVajyz3q
  R6pbO1UCgYEAimm/G/AHls6mFcpqtGQepBj6HiCn2xh1yIiVn6rgEwTIgXiOnaPk
  QSJagHDhMl0xohtpyUMON4Fwo3mZ6cP1qTykkAvn+KBZYmhHIHg/vFvW00dmzj+m
  icsf57kgDOPfVV+bLZ008gjLrgQrK17iwtAyXuqcFH2JrfLDCGA82zs=
  -----END RSA PRIVATE KEY-----`
	var github Github
	err := yaml.Unmarshal([]byte(input), &github)
	require.NoError(t, err)
	assert.Equal(t, "app-id", github.AppID)
	assert.Equal(t, "client-id", github.ClientID)
	assert.Equal(t, "client-secret", github.ClientSecret)
	assert.Equal(t, "webhook-secret", github.WebhookSecret)
	assert.Equal(t, "https://github.com", github.Host.String())
	assert.Equal(t, "https://api.github.com", github.APIHost.String())
	assert.Equal(t, "slug-1", github.Slug)
	assert.NotNil(t, github.PrivateKey)

}
