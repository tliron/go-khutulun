package commands

import (
	"github.com/spf13/cobra"
	hostpkg "github.com/tliron/khutulun/host"
	cobrautil "github.com/tliron/kutil/cobra"
	"github.com/tliron/kutil/util"
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

func init() {
	rootCommand.AddCommand(serverCommand)
	serverCommand.Flags().StringVarP(&statePath, "state-path", "p", "/mnt/khutulun", "state path")
	serverCommand.Flags().StringVarP(&grpcProtocol, "grpc-protocol", "", "tcp", "gRPC protocol (\"tcp\", \"tcp6\", or \"tcp4\")")
	serverCommand.Flags().StringVarP(&grpcAddress, "grpc-address", "", "::", "gRPC address")
	serverCommand.Flags().IntVarP(&grpcPort, "grpc-port", "", 8181, "gRPC port")
	serverCommand.Flags().StringVarP(&httpProtocol, "http-protocol", "", "tcp", "HTTP protocol (\"tcp\", \"tcp6\", or \"tcp4\")")
	serverCommand.Flags().StringVarP(&httpAddress, "http-address", "", "::", "HTTP address")
	serverCommand.Flags().IntVarP(&httpPort, "http-port", "", 8182, "HTTP port")
	serverCommand.Flags().StringVarP(&gossipAddress, "gossip-address", "", "::", "gossip address")
	serverCommand.Flags().IntVarP(&gossipPort, "gossip-port", "", 8183, "gossip port")
	serverCommand.Flags().StringVarP(&broadcastProtocol, "broadcast-protocol", "", "udp4", "broadcast protocol (\"udp6\", or \"udp4\")")
	serverCommand.Flags().StringVarP(&broadcastAddress, "broadcast-address", "", "239.0.0.1", "broadcast address")
	serverCommand.Flags().IntVarP(&broadcastPort, "broadcast-port", "", 8184, "broadcast port")

	cobrautil.SetFlagsFromEnvironment("KHUTULUN_CONDUCTOR_", serverCommand)
}

var serverCommand = &cobra.Command{
	Use:   "server",
	Short: "Run the server",
	Run: func(cmd *cobra.Command, args []string) {
		host, err := hostpkg.NewHost(statePath)
		util.FailOnError(err)
		util.OnExitError(host.Release)
		server := hostpkg.NewServer(host)
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
