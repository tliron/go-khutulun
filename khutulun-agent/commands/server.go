package commands

import (
	"os"

	"github.com/spf13/cobra"
	agentpkg "github.com/tliron/go-khutulun/agent"
	cobrautil "github.com/tliron/go-kutil/cobra"
	"github.com/tliron/go-kutil/util"
)

var statePath string
var grpcProtocol string
var grpcAddress string
var grpcPort int
var httpProtocol string
var httpAddress string
var httpPort int
var gossipAddress string
var gossipPort int
var broadcastProtocol string
var broadcastAddress string
var broadcastPort int

var defaultTcpProtocol = "tcp"
var defaultUdpProtocol = "udp"
var defaultAddress = "::"
var defaultBroadcastAddress = "ff02::1"

func init() {
	switch os.Getenv("IP_VERSION") {
	case "6":
		defaultTcpProtocol = "tcp6"
		defaultUdpProtocol = "udp6"

	case "4":
		defaultTcpProtocol = "tcp4"
		defaultUdpProtocol = "udp4"
		defaultAddress = "0.0.0.0"
		defaultBroadcastAddress = "239.0.0.1"

	case "4dual":
		defaultAddress = "0.0.0.0"
		defaultBroadcastAddress = "239.0.0.1"
	}

	rootCommand.AddCommand(serverCommand)
	serverCommand.Flags().StringVarP(&statePath, "state-path", "p", "/mnt/khutulun", "state path")
	serverCommand.Flags().StringVarP(&grpcProtocol, "grpc-protocol", "", defaultTcpProtocol, "gRPC protocol (\"tcp\", \"tcp6\", or \"tcp4\")")
	serverCommand.Flags().StringVarP(&grpcAddress, "grpc-address", "", defaultAddress, "gRPC address")
	serverCommand.Flags().IntVarP(&grpcPort, "grpc-port", "", 8181, "gRPC port")
	serverCommand.Flags().StringVarP(&httpProtocol, "http-protocol", "", defaultTcpProtocol, "HTTP protocol (\"tcp\", \"tcp6\", or \"tcp4\")")
	serverCommand.Flags().StringVarP(&httpAddress, "http-address", "", defaultAddress, "HTTP address")
	serverCommand.Flags().IntVarP(&httpPort, "http-port", "", 8182, "HTTP port")
	serverCommand.Flags().StringVarP(&gossipAddress, "gossip-address", "", defaultAddress, "gossip address")
	serverCommand.Flags().IntVarP(&gossipPort, "gossip-port", "", 8183, "gossip port")
	serverCommand.Flags().StringVarP(&broadcastProtocol, "broadcast-protocol", "", defaultUdpProtocol, "broadcast protocol (\"udp\", \"udp6\", or \"udp4\")")
	serverCommand.Flags().StringVarP(&broadcastAddress, "broadcast-address", "", defaultBroadcastAddress, "broadcast address")
	serverCommand.Flags().IntVarP(&broadcastPort, "broadcast-port", "", 8184, "broadcast port")

	cobrautil.SetFlagsFromEnvironment("KHUTULUN_CONDUCTOR_", serverCommand)
}

var serverCommand = &cobra.Command{
	Use:   "server",
	Short: "Run the server",
	Run: func(cmd *cobra.Command, args []string) {
		agent, err := agentpkg.NewAgent(statePath)
		util.FailOnError(err)
		util.OnExitError(agent.Release)
		server := agentpkg.NewServer(agent)
		util.OnExit(server.Stop)
		server.GRPCProtocol = grpcProtocol
		server.GRPCAddress = grpcAddress
		server.GRPCPort = grpcPort
		server.HTTPProtocol = httpProtocol
		server.HTTPAddress = httpAddress
		server.HTTPPort = httpPort
		server.GossipAddress = gossipAddress
		server.GossipPort = gossipPort
		server.BroadcastProtocol = broadcastProtocol
		server.BroadcastAddress = broadcastAddress
		server.BroadcastPort = broadcastPort
		err = server.Start(true, true)
		util.FailOnError(err)
		select {}
	},
}
