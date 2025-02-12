package GoSNMPServer

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"strconv"

	"github.com/gosnmp/gosnmp"
	"github.com/pkg/errors"

	q "github.com/quic-go/quic-go"
)

type QUICListener struct {
	conn       q.Connection
	logger     ILogger
	snmpserver *gosnmp.GoSNMP
}

// func NewQUICListener(address string, tlsConfig *tls.Config) (ISnmpServerListener, error) {
func NewQUICListener(address string, tlsConfig *tls.Config, filepath string, server *gosnmp.GoSNMP) (<-chan ISnmpServerListener, error) {
	ret := new(QUICListener)
	ret.snmpserver = server
	ret.logger = NewDefaultLogger(filepath)
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

	host, port, err := net.SplitHostPort(remoteAddr.String())
	if err != nil {
		return nil, nil, errors.Wrap(err, "[QUIC]SplitHostPort Error")
	}
	portConv, err := strconv.Atoi(port)
	if err != nil {
		return nil, nil, errors.Wrap(err, "[QUIC]Strconv.Atoi Error")
	}

	quic.snmpserver.Target = host
	quic.snmpserver.Port = uint16(portConv)

	sP, err := quic.snmpserver.SnmpDecodePacket(msg[:counts])
	if err != nil {
		return nil, nil, errors.Wrap(err, "[QUIC]SNMPDecodePacket Error")
	}
	var oids []string
	for _, vb := range sP.Variables {
		oids = append(oids, vb.Name)

	}
	quic.logger.Infof("quic request: %v", oids)

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
