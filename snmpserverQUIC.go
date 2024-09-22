package GoSNMPServer

import (
	"crypto/tls"

	"github.com/pkg/errors"
)

func (server *SNMPServer) ListenQUIC(address string, tlsConfig *tls.Config) error {
	if server.wconnStream != nil {
		return errors.New("Listened")
	}
	connectionChan, err := NewQUICListener(address, tlsConfig)
	if err != nil {
		return err
	}
	for conn := range connectionChan {
		go func() {
			server.logger.Infof("ListenQUIC: address=%s", address)

			conn.SetupLogger(server.logger)
			server.wconnStream = conn
			server.ServeForever()
		}()
	}
	return nil
}
