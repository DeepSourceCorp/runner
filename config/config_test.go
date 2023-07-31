package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("TASK_IMAGE_PULL_SECRET_NAME", "default")
	t.Setenv("TASK_IMAGE_REGISTRY_URL", "example.com")
	t.Setenv("TASK_NAMESPACE", "default")
	t.Setenv("TASK_NODE_SELECTOR", "foo=bar")
	input := `
runner:
  id: "runner-id"
  clientId: "client-id"
  clientSecret: "client-secret"
  webhookSecret: "webhook-secret"
  host: "https://example.com"
  privateKey: |
    -----BEGIN PRIVATE KEY-----
    MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQCr26gJM8hgt8i5
    SPTZg4Io+OioIJtQg7thJeDUMBLB32ycXB98INj6zFteX//PkuqD8lCENxbsOQDV
    CYUq0H+aky4TVFKXHif3p6gQTXkU/Yqp39EWm0DdMMerOI0r7J1zMsyYw/6jvSUW
    HAZYp4sEahWmn/T8ZC+mJZZ789FVunLyCbWtYl4DmEkkuA1olFacXf6hyKEjBZ3B
    Ow+dkJHfDUpCNmnb5j00Ev3Zb7mPCshoy2/NX5/e3+KZDZGDL9zbWkNsYoD1n910
    sVXGegPLU9TMUz5FQoSVPwaxf9j4sQTNPOp0Ilib8x8pWOfs7Iy3A3XXZNDlm9mQ
    jNhUvN+7AgMBAAECggEBAJRs1BaGg4OMlq33ZYhKPOrX9k/mQV1rODTx+tgnYLvS
    E8KDCaox0FPilPLQJGYIs8QbThCyZ3jCzoYvf7R3eA1vGbcV93KOV+RbBxp1XqKT
    SuPl6nYExiOCkp+86qfJ5j3s3Kj/dPfDTrlmoNCGetjoKiTLN1GX0VNEWVBaRiwr
    vJ0ptEPKYH60hoWr9yYb3ZC4Qp8uzy+eLa60SA+dDIGcsmsjEi/mINig5DOCLS4A
    xLpnmoR8n0ym4feJXWrgeRsxJDOu7fmsDstWCxNEdV685btsrMnXJsdfEG1i7jBU
    CgHJl+/6R3EXWwgNWO3sez5MiRjGd3/fpRlZUX3n31ECgYEA5J3FZJh9JYjWkO/M
    2jU5G8XUoHlmBOwVc9od5/fqESb9tTi1ZWGN95lBld7flMFkzqBUMAwitVK9XbBU
    3Kj4P/V3bbFo3fffh9f8h7i/otBlut1ykE7uT3TWK9wH1O5RrneIl/FO9eva+pv5
    Aei19f5qdpqI4IzCx0uDMeJTcekCgYEAwHF3FrdODQOh/re37eoFrk4CbQbp85J/
    cop4wzQHyiLTxqB3e1XETDa7cHPttKTX1AVMEutF5k2+y7Vd/I/MUi1HW9t/psK1
    mXjFzNAdZmpzIYHrYA2anx9lrReD0oueEExzzfcYOfruUa8udXoKqo4QvkPZVkJ/
    67wVLyw9+gMCgYEAvoxB2na+2GoVbPhyZe22i894Scjln3Sm7Mj/5Dhef61gCYwa
    pUWKbrTuVVxOPk5zF0XK5cE3rKop68zs7n5na+fMg0E7hsbzKOZ9NSJnl+za3cV1
    l5IyT0eyuxvJ61A4BJLc5sfaaF8NRZR7F3w/LanAUtq6+25XaoUl9I4PvwECgYEA
    jux2Fr/iztWQ3V1S0/aHa5HySUjmPgjicI4Y7FjbJCvDfvQ0aLwlArlvcjAXLZ9z
    z7pzamWjz0yUVDSJ7gZaJ/oK0lTttEtNlgLVXKx/+U073nnf9sGDwYQO/oPFWnxo
    0xAEvcYzDvSnRLFHXuZZv5utIbHAW0keOlTAov1HtkMCgYA34EftvNMyq4up3h3H
    El0Vx+HPWmn3Ze7wFSpgWf+ZwpCNQOpS9te06DlFSfWDk7c5VFIwFDpd9FgTPEGl
    7dONidHCAyyjLjQzj2VHT6Ph0Hs5PwlCzdHww/Q48HcWKvZy9cqNGAJRQ2p8IlrI
    1lOFqJmY0YtbcBp4Rw/nxEg8qg==
    -----END PRIVATE KEY-----
deepsource:
  host: https://deepsource.io
  publicKey: |
    -----BEGIN PUBLIC KEY-----
    MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAq9uoCTPIYLfIuUj02YOC
    KPjoqCCbUIO7YSXg1DASwd9snFwffCDY+sxbXl//z5Lqg/JQhDcW7DkA1QmFKtB/
    mpMuE1RSlx4n96eoEE15FP2Kqd/RFptA3TDHqziNK+ydczLMmMP+o70lFhwGWKeL
    BGoVpp/0/GQvpiWWe/PRVbpy8gm1rWJeA5hJJLgNaJRWnF3+ocihIwWdwTsPnZCR
    3w1KQjZp2+Y9NBL92W+5jwrIaMtvzV+f3t/imQ2Rgy/c21pDbGKA9Z/ddLFVxnoD
    y1PUzFM+RUKElT8GsX/Y+LEEzTzqdCJYm/MfKVjn7OyMtwN112TQ5ZvZkIzYVLzf
    uwIDAQAB
    -----END PUBLIC KEY-----
apps:
  - name: "app-1"
    provider: "github"  
    github:
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
        -----END RSA PRIVATE KEY-----
kubernetes:
  namespace: default
  nodeSelector:
    foo: bar
rqlite:
  host: "localhost"
  port: 4001
saml:
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
    -----END PRIVATE KEY-----
objectStorage:
  backend: gcs
  bucket: my-bucket
  credential: |
    {
      "private_key_id": "ABCDEF",
      "private_key": "Bag Attributes\n    friendlyName: key\n    localKeyID: 22 7E 04 FC 64 32 20 83 1E C1 BD E3 F5 2F 44 7D EA 99 A5 BC\nKey Attributes: <No Attributes>\n-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDh6PSnttDsv+vi\fd+CtKP8UJOQico+p\noJHSAPsrzSr6YsGs3c9SQOslBmm9Fkh9/f/GZVTVZ6u5AsUmOcVvZ2q7Sz8Vj/aR\naIm0EJqRe9cQ5vvN9sg25rIv4xKwIZJ1VixKWJLmpCmDINqn7xvl+ldlUmSr3aGt\nw21uSDuEJhQlzO3yf2FwJMkJ9SkCm9oVDXyl77OnKXj5bOQ/rojbyGeIxDJSUDWE\nGKyRPuqKi6rSbwg6h2G/Z9qBJkqM5NNTbGRIFz/9/LdmmwvtaqCxlLtD7RVEryAp\n+qTGDk5hAgMBAAECggEBAMYYfNDEYpf4A2SdCLne/9zrrfZ0kphdUkL48MDPj5vN\nTzTRj6f9s5ixZ/+QKn3hdwbguCx13QbH5mocP0IjUhyqoFFHfasdfYAWxyyaZfpjM8tO4\nQoEYxby3BpjLe62UXESUzChQSytJZFwIDasdasdaXKcdIPNO3zvVzufEJcfG5no2b9cIvsG\nDy6J1FNILWxCtDIqBM+G1B1is9DhZnUDgn0iKzINiZmh1I1l7k/4tMnozVIKAfwo\nf1kYjG/d2IzDM02mTeTElz3IKeNriaOIYTZgI26xLJxTkiFnBV4JOWFAZw15X+yR\n+DrjGSIkTfhzbLa20Vt3AFM+LFK0ZoXT2dRnjbYPjQECgYEA+9XJFGwLcEX6pl1p\nIwXAjXKJdju9DDn4lmHTW0Pbw25h1EXONwm/NPafwsWmPll9kW9IwsxUQVUyBC9a\nc3Q7rF1e8ai/qqVFRIZof275MI82ciV2Mw8Hz7FPAUyoju5CvnjAEH4+irt1VE/7\nSgdvQ1gDBQFegS69ijdz+cOhFxkCgYEA5aVoseMy/gIlsCvNPyw9+Jz/zBpKItX0\njGzdF7lhERRO2cursujKaoHntRckHcE3P/Z4K565bvVq+VaVG0T/asfaBcBKPmPHrLmY\niuVXidltW7Jh9/RCVwb5+BvqlwlC470PEwhqoUatY/fPJ74srztrqJHvp1L29FT5\nsdmlJW8YwokCgYAUa3dMgp5C0knKp5RY1KSSU5E11w4zKZgwiWob4lq1dAPWtHpO\nGCo63yyBHImoUJVP75gUw4Cpc4EEudo5tlkIVuHV8nroGVKOhd9/Rb5K47Hke4kk\nBrn5a0Ues9qPDF65Fw1ryPDFSwHufjXAAO5SpZZJF51UGDgiNvDedbBgMQKBgHSk\nt7DjPhtW69234eCckD2fQS5ijBV1p2lMQmCygGM0dXiawvN02puOsCqDPoz+fxm2\nDwPY80cw0M0k9UeMnBxHt25JMDrDan/iTbxu++T/jlNrdebOXFlxlI5y3c7fULDS\nLZcNVzTXwhjlt7yp6d0NgzTyJw2ju9BiREfnTiRBAoGBAOPHrTOnPyjO+bVcCPTB\nWGLsbBd77mVPGIuL0XGrvbVYPE8yIcNbZcthd8VXL/38Ygy8SIZh2ZqsrU1b5WFa\nXUMLnGEODSS8x/GmW3i3KeirW5OxBNjfUzEF4XkJP8m41iTdsQEXQf9DdUY7X+CB\nVL5h7N0VstYhGgycuPpcIUQa\n-----END PRIVATE KEY-----\n",
      "client_email": "dummy@google.com",
      "client_id": "123",
      "type": "service_account"
    }`

	var c Config
	err := yaml.Unmarshal([]byte(input), &c)
	assert.NoError(t, err)
}
