package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var AuthorizedDevices []string

func main() {
	caCert, err := ioutil.ReadFile("ca.crt")
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
		MinVersion: tls.VersionTLS12,
	}

	ExtractAuthorizedDevices(err)

	// Create a Server instance to listen on port 8443 with the TLS config
	server := &http.Server{
		Addr:      ":8443",
		Handler: http.HandlerFunc(HelloServer),
		TLSConfig: tlsConfig,
	}

	log.Fatal(server.ListenAndServeTLS("./server.crt", "./server.key"))
}

func HelloServer(rw http.ResponseWriter, req *http.Request){
	fmt.Printf("[reverse proxy server] received request at: %s\n", time.Now())

	var DeviceName = req.TLS.VerifiedChains[0][0].Subject.CommonName
	fmt.Println("Request using DeviceID", DeviceName, "certificate", req.URL)

	if deviceAuthorized(DeviceName) {
		fmt.Println("Device ID ", DeviceName, "is authorized")

		originServerURL, err := url.Parse("http://127.0.0.1:80/device-worker-upgrade/repo")
		if err != nil {
			log.Fatal("invalid origin server URL")
		}


		// set req Host, URL and Request URI to forward a request to the origin server
		req.Host = originServerURL.Host
		req.URL.Host = originServerURL.Host
		req.URL.Scheme = originServerURL.Scheme
		req.RequestURI = ""

		// send a request to the origin server
		originServerResponse, err := http.DefaultClient.Do(req)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(rw, err)
			return
		}

		// return response to the client
		copyHeader(rw.Header(), originServerResponse.Header)
		rw.WriteHeader(originServerResponse.StatusCode)
		io.Copy(rw, originServerResponse.Body)
		originServerResponse.Body.Close()

	} else{
		fmt.Println("ERROR: Device ID ", DeviceName, "is not authorized")
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}
}

// copyHeader and singleJoiningSlash are copy from "/net/http/httputil/reverseproxy.go"
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func ExtractAuthorizedDevices(err error) {
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


func deviceAuthorized(deviceCN string) bool{
	for _, autoz := range AuthorizedDevices {
		if autoz == deviceCN{
			return true
		}
	}

	return false
}
