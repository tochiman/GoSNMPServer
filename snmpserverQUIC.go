package GoSNMPServer

import (
	"crypto/tls"

	"github.com/pkg/errors"
)

func (server *SNMPServer) ListenQUIC(address string, tlsConfig *tls.Config) error {
	if server.wconnStream != nil {
		return errors.New("Listened")
	}
	i, err := NewQUICListener(address, tlsConfig)
	if err != nil {
		return err
	}
	server.logger.Infof("ListenQUIC: address=%s", address)
	i.SetupLogger(server.logger)
	server.wconnStream = i
	return nil
}
