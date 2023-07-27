package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestSAML_UnmarshalYAML(t *testing.T) {
	input := `
enabled: true
metadataUrl: https://samltest.id/saml/idp
certificate: |
  -----BEGIN CERTIFICATE-----
  MIICvDCCAaQCCQC7HJDoGWaYLDANBgkqhkiG9w0BAQsFADAgMR4wHAYDVQQDDBVt
  eXNlcnZpY2UuZXhhbXBsZS5jb20wHhcNMjMwNzEzMTQzMjMxWhcNMjQwNzEyMTQz
  MjMxWjAgMR4wHAYDVQQDDBVteXNlcnZpY2UuZXhhbXBsZS5jb20wggEiMA0GCSqG
  SIb3DQEBAQUAA4IBDwAwggEKAoIBAQDpl2U9S9h/Y7ZDkp7o9dfoItwqncVptvOw
  tinwK1TLB2uSdN58QqTyuMZTgUjVv+AeSEixHa92DE1F1TxpVJRGUK6c3J+k53g3
  d33xjPiawzJxkLOFExrpxheKolNmjIjoHkY+o/dVYncgui7UDLao3yzcPdWU6400
  +/mrSc5INCSpIE2pbVFtcT/yj73hj+7ns2Z1c//aqNsWOr9U63SwpVRD6v3mMho2
  hLLXNJ3sZwMJgEPCVx0Es74KkdQoR8jyfjbKGiBbyJhU7j9EiAYleS7VP/YHdN5a
  qtT/JX0L2gjueXk5Neap/MISGrpGHCGq8K5QNia30sb+E5SONsHNAgMBAAEwDQYJ
  KoZIhvcNAQELBQADggEBAK46iwXDkDiyiaQd035Sslf6anCIqSPqFBKidRRZJtIS
  l/BAMgljUSGSagp5mTJXOU/jU1b+xrxlYZizWlrH6hpo9pdwgnZTF2V36xicz0PK
  8D2wsx+MOYGvJVkUmLz0+dsTPHzauDmjaz7WoZTMv/RnBybFV1tyXPMy00Zdh3Wo
  kHbeCsmYAJedo765nrOR+eGDMqHHWYUJzo5iU7kpLlQH/nNkqMdTJpniTfbWARow
  cC87WkoGg80l3bWOlelVkbAGBtwqw2YJtmyOHsjEauHd4OCdxriG1Yca3JtEVURn
  bGlzj9VI3XA7YIqmip+iFe/oj6PlAcU01h1ZJ00zMXQ=
  -----END CERTIFICATE-----
key: |
  -----BEGIN PRIVATE KEY-----
  MIIEwAIBADANBgkqhkiG9w0BAQEFAASCBKowggSmAgEAAoIBAQDpl2U9S9h/Y7ZD
  kp7o9dfoItwqncVptvOwtinwK1TLB2uSdN58QqTyuMZTgUjVv+AeSEixHa92DE1F
  1TxpVJRGUK6c3J+k53g3d33xjPiawzJxkLOFExrpxheKolNmjIjoHkY+o/dVYncg
  ui7UDLao3yzcPdWU6400+/mrSc5INCSpIE2pbVFtcT/yj73hj+7ns2Z1c//aqNsW
  Or9U63SwpVRD6v3mMho2hLLXNJ3sZwMJgEPCVx0Es74KkdQoR8jyfjbKGiBbyJhU
  7j9EiAYleS7VP/YHdN5aqtT/JX0L2gjueXk5Neap/MISGrpGHCGq8K5QNia30sb+
  E5SONsHNAgMBAAECggEBAMoY9zliLoyAu4eZCi2pzcQErRGd8Ne2tv3TjVNCWhlS
  cSqEPJ2rl0R8wvIqb9anLINmrKW4dj8fA5gAlkTXLXXshjYm12R381Wh53AeNFTJ
  vxHsTLU8w1Mw1NtX9+pIeobA8qttdycDiufgzXUfDsXqWMiwIuK2LTSDMQ6WS4fB
  FjU/XcNIi3QA2nM5W+3bPc/Voaqdtn64htYHOcRr2cehAVvVGMeUYEI69t1RwOqG
  3Lddg/2R0I+MKl3/NHsQAh01jhqe56hP24phapAIlaaKJoBCfF61TpgDYG7NZqto
  jntHPv1G+0m5K5v5pmcWOxSwpR+x7kaCWCcQpHLkdYECgYEA9UBe/K0PODS5n7jK
  Qpnv88ctm0h15O01ymWU51G8k4uEojyKjoEmWiZUeuv0nziZD60s0CN1VAFRgCZ5
  I9V2cFfJl8/uh6BXtkuXZD2VaKcH+OEwTmHyoPtdXp3tXI0fiZmAh6lZeAbMPC/B
  7PAJUAWVe5IEIscQLmU+6ecROqECgYEA89Qz8y0zEEZo3Z+kaL2W+s+HEtlOl44h
  bI6T+YYRT3b4ce5s3KcLTvri0Hn9miLN+gO85Cr/pXBHt5AH3TItYJTIkiwRrgTe
  gEvuDkKkqYQ9gqRNACo9VPBMkisRcg/x8g1Tl8DkpGp4nZdgJ7nXVrbaTq8n7LkJ
  8A7Z+CT6Q60CgYEA3E1khekXANAr5hPibA1HhF3o09I1RNzoMtUo+tlrYcYz8GAd
  voC46MYBoSGPbe8zXueal6UiYcGFam4k51F6wNO63MoFZINeBvzEE2FWctmHycLO
  17oYbw8dAj8u1rJWIA5pbHNtUOoaT/4+Xw4H73/0lTnGyU6zdFmyN/4+dcECgYEA
  kyBTbIO0kTh7JGek9BKaXKMGtSfs1WRM5M0vmtv77AA0r8KXa5lcKH8Yh4VksjIY
  KalBvEf51GDo1WmSZTVWzjVYxWLUFDYZ8D5g2bf61dLWrtLnJ5dVRMBu47AbKcFX
  U6AY9bPOAyu/tg/WVII93rQdDGeCZsPMrE651ZKydE0CgYEA1YtvmC7oYA+NV7Mt
  v10V1FyhcjvpvAGD+Z58FNSTbLHO+fgmsEJeepdzrVrI+v75rIgU9o0KqYXgaTdh
  lkPFLIx8+vcA+YkXonPBjL6YVu929zn8InTR50eRf57g1jbGNvavP8HUb4SComw/
  gui8lSvhx4H4omrlecZ6n+5HQPk=
  -----END PRIVATE KEY-----`
	var saml SAML
	err := yaml.Unmarshal([]byte(input), &saml)
	require.NoError(t, err)
	assert.Equal(t, "https://samltest.id/saml/idp", saml.MetadataURL.String())
	assert.NotNil(t, saml.Certificate)
}
