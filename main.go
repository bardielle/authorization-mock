package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var AuthorizedDevices []string

func main() {
	caCert, err := ioutil.ReadFile("ca.crt")

	http.HandleFunc("/", HelloServer)

	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	ExtractAuthorizedDeviecs(err)

	// Create a Server instance to listen on port 8443 with the TLS config
	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
	}

	log.Fatal(server.ListenAndServeTLS("./server.crt", "./server.key"))
}

func ExtractAuthorizedDeviecs(err error) {
	file, err := os.Open("authorized_devices.txt")
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		AuthorizedDevices = append(AuthorizedDevices, scanner.Text())
	}
	file.Close()
}


func HelloServer(w http.ResponseWriter, r *http.Request) {
	var DeviceName = r.TLS.VerifiedChains[0][0].Subject.CommonName
	fmt.Println("Request using DeviceID", DeviceName, "certificate", r.URL)

	if deviceAuthorized(DeviceName) {
		println("^^^^^^ device ID Authorized")
	} else{
		println("****** Error - device ID is not as expected")
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
}

func deviceAuthorized(deviceCN string) bool{
	for _, autoz := range AuthorizedDevices {
		if autoz == deviceCN{
			return true
		}
	}

	return false
}

