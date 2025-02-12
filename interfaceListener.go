package GoSNMPServer

import (
	"net"
	"strconv"

	"github.com/gosnmp/gosnmp"
	"github.com/pkg/errors"
)

type ISnmpServerListener interface {
	SetupLogger(ILogger)
	Address() net.Addr
	NextSnmp() (snmpbytes []byte, replyer IReplyer, err error)
	Shutdown()
}

type IReplyer interface {
	ReplyPDU([]byte) error
	Shutdown()
}

type UDPListener struct {
	conn       *net.UDPConn
	logger     ILogger
	snmpserver *gosnmp.GoSNMP
}

func NewUDPListener(l3proto, address string, server *gosnmp.GoSNMP) (ISnmpServerListener, error) {
	ret := new(UDPListener)
	ret.logger = NewDiscardLogger()
	ret.snmpserver = server
	udpaddr, err := net.ResolveUDPAddr(l3proto, address)
	if err != nil {
		return nil, errors.Wrap(err, "ResolveUDPAddr Error")
	}
	conn, err := net.ListenUDP(l3proto, udpaddr)
	if err != nil {
		return nil, errors.Wrap(err, "UDP Listen Error")
	}
	ret.conn = conn
	return ret, nil
}

func (udp *UDPListener) SetupLogger(i ILogger) {
	udp.logger = i
}
func (udp *UDPListener) Address() net.Addr {
	return udp.conn.LocalAddr()
}

func (udp *UDPListener) NextSnmp() ([]byte, IReplyer, error) {
	var msg [4096]byte
	if udp.conn == nil {
		return nil, nil, errors.New("Connection Not Listen")
	}
	counts, udpAddr, err := udp.conn.ReadFromUDP(msg[:])
	if err != nil {
		return nil, nil, errors.Wrap(err, "UDP Read Error")
	}
	udp.logger.Infof("udp request from %v. size=%v", udpAddr, counts)

	host, port, err := net.SplitHostPort(udpAddr.String())
	if err != nil {
		return nil, nil, errors.Wrap(err, "SplitHostPort Error")
	}
	portConv, err := strconv.Atoi(port)
	if err != nil {
		return nil, nil, errors.Wrap(err, "[QUIC]Strconv.Atoi Error")
	}

	udp.snmpserver.Target = host
	udp.snmpserver.Port = uint16(portConv)

	// g := &gosnmp.GoSNMP{
	// 	Target:    host,
	// 	Port:      uint16(portConv),
	// 	Community: "public",
	// 	Version:   gosnmp.Version2c,
	// 	Timeout:   time.Duration(time.Second * 3),
	// 	Retries:   0,
	// }

	sP, err := udp.snmpserver.SnmpDecodePacket(msg[:counts])
	if err != nil {
		return nil, nil, errors.Wrap(err, "SNMPDecodePacket Error")
	}
	var oids []string
	for _, vb := range sP.Variables {
		oids = append(oids, vb.Name)
	}
	udp.logger.Infof("udp request: %v", oids)

	return msg[:counts], &UDPReplyer{udpAddr, udp.conn}, nil
}

func (udp *UDPListener) Shutdown() {
	if udp.conn != nil {
		udp.conn.Close()
		udp.conn = nil
	}
}

type UDPReplyer struct {
	target *net.UDPAddr
	conn   *net.UDPConn
}

func (r *UDPReplyer) ReplyPDU(i []byte) error {
	conn := r.conn
	_, err := conn.WriteToUDP(i, r.target)
	if err != nil {
		return errors.Wrap(err, "WriteToUDP")
	}
	return nil
}

func (r *UDPReplyer) Shutdown() {}
