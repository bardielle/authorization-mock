# Authorization validation mock
## prerequisite
The running directory should contains those files: 
* Server.crt
* Server.key
* CA.crt
* ca.key
* authorization_devices.txt

Use the secret from flotta namespace to get ca.crt, ca.key, server.crt, server.key

```
kubectl get secret -n flotta flotta-ca  -o json | jq '.data."ca.crt"| @base64d' -r >ca.crt
kubectl get secret -n flotta flotta-ca  -o json | jq '.data."ca.key"| @base64d' -r >ca.key

kubectl -n flotta get secret flotta-host-certificate -o json | jq -r '.data."server.crt" | @base64d' >server.crt
kubectl -n flotta get secret flotta-host-certificate -o json | jq -r '.data."server.key" | @base64d' >server.key
```


## create client certificate and key:
```
openssl genrsa -out client123.key 4096

fill in more info
openssl req -new -subj '/CN=device123' -key client123.key -out client123.req

openssl x509 -req \
-in client123.req \
-CA ca.crt \
-CAkey ca.key \
-CAcreateserial \
-out client123.crt \
-days 10 -sha256
```

## run a test 
```
curl -v  --cacert ca.crt   --cert client123.crt   --key client123.key -v   -X GET   https://127.0.0.1:8443/123
```

# Upgrade a device - use a reverse proxy
## Open port in firewalld
In the Builder VM open the port 8443
``` 
firewall-cmd --add-port=8443/tcp
```

## Use the dns name of the server
It can be found using the command: `openssl x509 -in /home/repo/authorization-mock/server.crt -text`
Update it in `/etc/hosts`

## Config the remote
In the remotes.d file config those flags (please store the certificates in `/etc/pki/tls`) 
* tls-client-cert-path
* tls-client-key-path
* tls-ca-path
* url (should be looked like `https://project-flotta.io:8443/device-worker-upgrade/repo`)

