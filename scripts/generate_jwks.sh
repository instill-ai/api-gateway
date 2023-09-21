#!/bin/sh
mkdir -p /instill
/go/bin/jwx jwk generate --type RSA --set --template '{"kid":"instill"}' > /instill/instill.jwks
/go/bin/jwx jwk format --public-key --set /instill/instill.jwks > /instill/instill.jwks.pub