package GoSNMPServer

import "crypto/tls"

func GenerateTLSConfig() *tls.Config {
	certPEM := "cert/localhost/cert.pem"

	keyPEM := "cert/localhost/key.pem"

	tlsCert, err := tls.LoadX509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}
}
