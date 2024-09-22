package GoSNMPServer

import (
	"crypto/tls"
	"fmt"
)

func GenerateTLSConfig(certFile string, keyFile string) *tls.Config {

	certPEM := fmt.Sprintf("cert/%s", certFile)

	keyPEM := fmt.Sprintf("cert/%s", keyFile)

	tlsCert, err := tls.LoadX509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"snmp-quic"},
	}
}
