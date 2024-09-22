package main

import (
	"os"
	"strings"

	"github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
	"github.com/slayercat/GoSNMPServer"
	"github.com/slayercat/GoSNMPServer/mibImps"
	"github.com/urfave/cli/v2"
)

func makeApp() *cli.App {
	return &cli.App{
		Name:        "gosnmpserver",
		Description: "an example server of gosnmp",
		Commands: []*cli.Command{
			{
				Name:    "RunServer",
				Aliases: []string{"run-server"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "logLevel", Value: "info"},
					&cli.StringFlag{Name: "community", Value: "public"},
					&cli.StringFlag{Name: "bindTo", Value: "localhost:1161"},
					&cli.StringFlag{Name: "v3Username", Value: "testuser"},
					&cli.StringFlag{Name: "v3AuthenticationPassphrase", Value: "testauth"},
					&cli.StringFlag{Name: "v3PrivacyPassphrase", Value: "testpriv"},
					&cli.BoolFlag{Name: "v3Only", Value: false},
					&cli.BoolFlag{Name: "quicMode", Usage: "Listening with QUIC", Value: false},
					&cli.StringFlag{Name: "certPEM", Usage: "Specify FilePath of certPEM", Value: "localhost/cert.pem"},
					&cli.StringFlag{Name: "keyPEM", Usage: "Specify FilePath of keyPEM", Value: "localhost/key.pem"},
				},
				Action: runServer,
			},
		},
	}
}

func main() {
	app := makeApp()
	app.Run(os.Args)
}

func runServer(c *cli.Context) error {
	logger := GoSNMPServer.NewDefaultLogger()
	switch strings.ToLower(c.String("logLevel")) {
	case "fatal":
		logger.(*GoSNMPServer.DefaultLogger).Level = logrus.FatalLevel
	case "error":
		logger.(*GoSNMPServer.DefaultLogger).Level = logrus.ErrorLevel
	case "info":
		logger.(*GoSNMPServer.DefaultLogger).Level = logrus.InfoLevel
	case "debug":
		logger.(*GoSNMPServer.DefaultLogger).Level = logrus.DebugLevel
	case "trace":
		logger.(*GoSNMPServer.DefaultLogger).Level = logrus.TraceLevel
	}
	mibImps.SetupLogger(logger)

	master := GoSNMPServer.MasterAgent{
		Logger: logger,
		SecurityConfig: GoSNMPServer.SecurityConfig{
			AuthoritativeEngineBoots: 1,
			SnmpV3Only:               c.Bool("v3Only"),
			Users: []gosnmp.UsmSecurityParameters{
				{
					UserName:                 c.String("v3Username"),
					AuthenticationProtocol:   gosnmp.MD5,
					PrivacyProtocol:          gosnmp.DES,
					AuthenticationPassphrase: c.String("v3AuthenticationPassphrase"),
					PrivacyPassphrase:        c.String("v3PrivacyPassphrase"),
				},
			},
		},
		SubAgents: []*GoSNMPServer.SubAgent{
			{
				CommunityIDs: []string{c.String("community")},
				OIDs:         mibImps.All(),
			},
		},
	}

	logger.Infof("V3 Users:")
	for _, val := range master.SecurityConfig.Users {
		logger.Infof(
			"\tUserName:%v\n\t -- AuthenticationProtocol:%v\n\t -- PrivacyProtocol:%v\n\t -- AuthenticationPassphrase:%v\n\t -- PrivacyPassphrase:%v",
			val.UserName,
			val.AuthenticationProtocol,
			val.PrivacyProtocol,
			val.AuthenticationPassphrase,
			val.PrivacyPassphrase,
		)
	}
	server := GoSNMPServer.NewSNMPServer(master)
	if c.Bool("quicMode") == true {
		err := server.ListenQUIC(c.String("bindTo"), GoSNMPServer.GenerateTLSConfig(c.String("certPEM"), c.String("keyPEM")))
		if err != nil {
			logger.Error("Error in listen: %+v", err)
		}
		// server.ServeForever()
		return nil
	} else {
		err := server.ListenUDP("udp", c.String("bindTo"))
		if err != nil {
			logger.Errorf("Error in listen: %+v", err)
		}
		server.ServeForever()
		return nil
	}
}
