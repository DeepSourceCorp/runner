# runner
DeepSource self-hosted runners.  The power of DeepSource Cloud with the safety of your infrastructure.

## Notes:

### Generate Runner key pair:
```
openssl genrsa 2048 | openssl pkcs8 -topk8 -nocrypt > private_key.pem
openssl rsa -in private_key.pem -pubout -out public_key.pem
```

### Generate SAML Certificate:
```
openssl req -x509 -newkey rsa:2048 -keyout myservice.key -out myservice.cert -days 365 -nodes -subj "/CN=myservice.example.com"
```