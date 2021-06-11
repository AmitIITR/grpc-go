#!/bin/zsh

function generateCACertificate() {
  echo "Generating CA Certificate"
  openssl ecparam -name prime256v1 -genkey -noout -out certs/cakey.key
  openssl req -x509 -new -nodes -key certs/cakey.key -subj "/CN=ServerMonCA/C=SM" -days 3650 -out certs/cacert.pem
}

function generateServerCertificate() {
  echo "Generating Server Certificate"
  openssl ecparam -name prime256v1 -genkey -noout -out certs/server.key
  generateCSRConfigForServer
  openssl req -new -key certs/server.key -out certs/server.csr -config certs/csrserver.conf
  openssl x509 -req -in certs/server.csr -CA certs/cacert.pem -CAkey certs/cakey.key -CAcreateserial -out certs/server.pem -days 3650 -extfile certs/csrserver.conf -extensions req_ext
}

function generateClientCertificate() {
  echo "Generating Client Certificate"
  openssl ecparam -name prime256v1 -genkey -noout -out certs/client.key
  generateCSRConfigForClient
  openssl req -new -key certs/client.key -out certs/client.csr -config certs/csrclient.conf
  openssl x509 -req -in certs/client.csr -CA certs/cacert.pem -CAkey certs/cakey.key -CAcreateserial -out certs/client.pem -days 3650 -extfile certs/csrclient.conf -extensions req_ext
}

function generateCSRConfigForServer() {
  echo "Generating CSR config for server"
cat > certs/csrserver.conf <<EOF
  [ req ]
  default_bits = 256
  prompt = no
  default_md = sha256
  req_extensions = req_ext
  distinguished_name = dn

  [ dn ]
  C = SM
  CN = localhost

  [ req_ext ]
  keyUsage = keyEncipherment, dataEncipherment
  extendedKeyUsage = serverAuth
  subjectAltName = @alt_names

  [ alt_names ]
  DNS.1 = localhost
  IP.1 = 127.0.0.1

EOF
}

function generateCSRConfigForClient() {
  echo "Generating CSR config for Client"
cat > certs/csrclient.conf <<EOF
  [ req ]
  default_bits = 256
  prompt = no
  default_md = sha256
  req_extensions = req_ext
  distinguished_name = dn

  [ dn ]
  C = SM
  CN = client

  [ req_ext ]
  keyUsage = keyEncipherment
  extendedKeyUsage = clientAuth

EOF
}

function generateCerts() {
  rm -rf certs
  mkdir certs
  generateCACertificate
  generateServerCertificate
  generateClientCertificate
}

generateCerts
