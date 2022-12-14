package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	dht "github.com/libp2p/go-libp2p-kad-dht"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
)

const echoProtocol = "/echo/1.0.0"

func main() {
	ctx := context.Background()

	// 	var peers peerstore.Peerstore
	// 	{
	// 		datastore, err := badger.NewDatastore("./peerstore", nil)
	// 		check(err)
	// 		peers, err = pstoreds.NewPeerstore(ctx, datastore, pstoreds.DefaultOpts())
	// 		check(err)
	// 	}

	node, err := libp2p.New(
		// libp2p.Peerstore(peers),
		libp2p.Routing(func(node host.Host) (routing.PeerRouting, error) {
			return dht.New(ctx, node)
		}),
	)
	check(err)

	node.SetStreamHandler(echoProtocol, func(s network.Stream) {
		fmt.Println("received new stream")
		buf := bufio.NewReader(s)
		err := func() error {
			str, err := buf.ReadString('\n')
			if err != nil {
				return err
			}
			fmt.Println(str)
			_, err = io.WriteString(s, str)
			return err
		}()
		if err != nil {
			fmt.Println(err)
			s.Reset()
		} else {
			s.Close()
		}
	})

	fmt.Println("my id:", node.ID())
	fmt.Println("listen addresses:", node.Addrs())
	{
		addrs, err := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{
			ID:    node.ID(),
			Addrs: node.Addrs(),
		})
		check(err)
		fmt.Println("my multiaddrs?:", addrs)
	}
	// fmt.Printf("%v known peers\n", len(peers.Peers()))

	defer func() { check(node.Close()) }()

	if len(os.Args) > 1 {

		fmt.Println(os.Args[1])

		addr, err := peer.AddrInfoFromString(os.Args[1])
		check(err)
		check(node.Connect(ctx, *addr))
		fmt.Println("successfully connected")

		stream, err := node.NewStream(ctx, addr.ID, echoProtocol)
		check(err)
		defer stream.Reset()
		fmt.Println("successfully opened echo stream")

		_, err = io.WriteString(stream, "hello from sender\n")
		check(err)

		out, err := io.ReadAll(stream)
		check(err)

		fmt.Printf("reply: %s\n", out)

	} else {

		// wait for a SIGINT or SIGTERM signal
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, os.Kill)
		<-ch
		fmt.Println("Received signal, shutting down...")

	}

	// 	dht, err := dht.New(ctx, node)
	// 	check(err)
	// 	check(dht.Bootstrap(ctx))

	// routingDiscovery := discovery.NewRoutingDiscovery(dht)

}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
