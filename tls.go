// Copyright (C) Pagoda Box, Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package router

import (
	"crypto/tls"
	"net"
	"net/http"

	"github.com/jcelliott/lumber"
)

var tlsListener net.Listener
var certificates = []tls.Certificate{}

var keys = []KeyPair{}

type KeyPair struct {
	Cert string `json:"cert"`
	Key  string `json:"key"`
}

var address = "0.0.0.0:443"

// Start listening for secure connection.
// The web server is split out from the much simpler form of
//   http.ListenAndServeTLS(addr string, certFile string, keyFile string, handler Handler)
// because we needed to handle multiple certs all at the same time and we needed
// to be able to change the set of certs without restarting the server
// this can be done by establishing a tls listener seperate form the http Server.
func StartTLS(addr string) error {
	address = addr
	var err error
	if tlsListener != nil {
		tlsListener.Close()
	}
	if len(certificates) > 0 {
		config := &tls.Config{
			Certificates: certificates,
		}
		config.BuildNameToCertificate()
		tlsListener, err = tls.Listen("tcp", address, config)
		if err != nil {
			return err
		}

		go http.Serve(tlsListener, handler{https: true})
	}

	return nil
}

// update the stored certificates and restart the web server
func UpdateCerts(newKeys []KeyPair) {
	newCerts := []tls.Certificate{}
	for _, newKey := range newKeys {
		cert, err := tls.X509KeyPair([]byte(newKey.Cert), []byte(newKey.Key))
		if err == nil {
			newCerts = append(newCerts, cert)
		} else {
			lumber.Error("[NANOBOX-ROUTER] Failed to update certs - %v", err)
		}
	}

	mutex.Lock()
	keys = newKeys
	certificates = newCerts
	mutex.Unlock()
	StartTLS(address)
	lumber.Trace("[NANOBOX-ROUTER] Certs updated")
}

// list the cached keys.
func Keys() []KeyPair {
	return keys
}
