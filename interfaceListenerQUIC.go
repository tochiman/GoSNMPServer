package GoSNMPServer

import (
	"context"
	"crypto/tls"
	"log"
	"net"

	"github.com/pkg/errors"

	q "github.com/quic-go/quic-go"
)

type QUICListener struct {
	conn   q.Connection
	logger ILogger
}

// func NewQUICListener(address string, tlsConfig *tls.Config) (ISnmpServerListener, error) {
func NewQUICListener(address string, tlsConfig *tls.Config) (<-chan ISnmpServerListener, error) {
	ret := new(QUICListener)
	ret.logger = NewDefaultLogger()
	listener, err := q.ListenAddr(address, tlsConfig, nil)
	if err != nil {
		return nil, errors.Wrap(err, "[QUIC]ListenAddr Error")
	}
	serverChan := make(chan ISnmpServerListener)
	go func() {
		for {
			conn, err := listener.Accept(context.Background())
			if err != nil {
				log.Fatalln("Error Accepting Connection", err)
				continue
			}
			ret.conn = conn
			serverChan <- ret
		}
	}()
	return serverChan, nil
}

func (quic *QUICListener) SetupLogger(i ILogger) {
	quic.logger = i
}

func (quic *QUICListener) Address() net.Addr {
	return quic.conn.LocalAddr()
}

func (quic *QUICListener) NextSnmp() ([]byte, IReplyer, error) {
	var msg [4096]byte // Buffer
	var remoteAddr net.Addr = quic.conn.RemoteAddr()
	if quic.conn == nil {
		return nil, nil, errors.New("Connection Not Listen")
	}
	stream, err := quic.conn.AcceptStream(context.Background())
	if err != nil {
		return nil, nil, errors.Wrap(err, "[QUIC]AcceptStream Error")
	}

	counts, err := stream.Read(msg[:])
	if err != nil {
		return nil, nil, errors.Wrap(err, "[QUIC]Can't Read Stream")
	}
	quic.logger.Infof("quic request from %v. size=%v", remoteAddr, counts)
	return msg[:counts], &QUICReplyer{remoteAddr, quic.conn, stream}, nil
}

func (quic *QUICListener) Shutdown() {
	if quic.conn != nil {
		quic.conn.CloseWithError(0, "Close Connection of QUIC....")
		quic.conn = nil
	}
}

type QUICReplyer struct {
	target net.Addr
	conn   q.Connection
	stream q.Stream
}

func (r *QUICReplyer) ReplyPDU(i []byte) error {
	var err error

	_, err = r.stream.Write(i)
	if err != nil {
		return errors.Wrap(err, "[QUIC]Can't Write Stream")
	}

	return nil
}

func (r *QUICReplyer) Shutdown() {}
