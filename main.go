package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
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

	// log.SetAllLoggers(log.LevelInfo)

	// 	libp2p.SetDefaultServiceLimits(&rcmgr.InfiniteLimits)

	// 	r, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits))
	// 	check(err)

	nextPeer := make(chan peer.AddrInfo)

	node, err := libp2p.New(
		// libp2p.ResourceManager(r),
		// libp2p.Peerstore(peers),

		// 		libp2p.EnableAutoRelay(autorelay.WithPeerSource(
		// 			func(ctx context.Context, numPeers int) <-chan peer.AddrInfo {
		// 			},
		// 			0,
		// 		)),

		libp2p.NATPortMap(),
		libp2p.EnableAutoRelay(
			// autorelay.WithCircuitV1Support(),
			autorelay.WithBootDelay(0),
			autorelay.WithPeerSource(
				func(ctx context.Context, numPeers int) <-chan peer.AddrInfo {
					r := make(chan peer.AddrInfo)
					go func() {
						defer close(r)
						for i := 0; i < numPeers; i++ {
							select {
							case p := <-nextPeer:
								select {
								case r <- p:
								case <-ctx.Done():
									return
								}
							case <-ctx.Done():
								return
							}
						}
					}()
					return r
				},
				0,
			),
		),
		libp2p.ForceReachabilityPrivate(),

		libp2p.Routing(func(node host.Host) (routing.PeerRouting, error) {
			r, err := dht.New(
				ctx, node,
				// dht.BootstrapPeers(dht.GetDefaultBootstrapPeerAddrInfos()...),
			)
			if err != nil {
				return nil, err
			}

			if err := r.Bootstrap(ctx); err != nil {
				return nil, err
			}

			for i, p := range dht.GetDefaultBootstrapPeerAddrInfos() {
				if err := node.Connect(ctx, p); err != nil {
					fmt.Printf("failed to connect to bootstrap node #%v\n", i)
				} else {
					fmt.Println("connected to bootstrap node")
				}
			}
			fmt.Println("done with bootstrapping")

			// use dht to get new peers for autorelay
			go func() {
				for {
					peers, err := r.GetClosestPeers(ctx, node.ID().String())
					if err != nil {
						fmt.Println("GetClosestPeers error:", err)
						time.Sleep(time.Second)
						continue
					}
					for _, p := range peers {
						addrs := node.Peerstore().Addrs(p)
						if len(addrs) == 0 {
							continue
						}
						nextPeer <- peer.AddrInfo{
							ID:    p,
							Addrs: addrs,
						}
					}
				}
			}()

			// err = r.Bootstrap(ctx)
			// if err == nil {
			// err = <-r.ForceRefresh()
			// }
			return r, nil
		}),
	)
	check(err)
	fmt.Println("\n\n\ncreated node\n\n\n")

	if false {
		kad, err := dht.New(ctx, node)
		check(err)
		check(kad.Bootstrap(ctx))
		for i, p := range dht.GetDefaultBootstrapPeerAddrInfos() {
			if err := node.Connect(ctx, p); err != nil {
				fmt.Printf("failed to connect to bootstrap node #%v\n", i)
			} else {
				fmt.Println("connected to bootstrap node")
			}
		}
		fmt.Println("done with bootstrapping")
		peers, err := kad.GetClosestPeers(ctx, node.ID().String())
		check(err)
		fmt.Println("peers:")
		for _, p := range peers {
			ps, err := node.Peerstore().SupportsProtocols(
				p,
				"/libp2p/circuit/relay/0.1.0",
				"/libp2p/circuit/relay/0.2.0/stop",
				"/libp2p/circuit/relay/0.2.0/hop",
			)
			check(err)
			if len(ps) == 0 {
				continue
			}
			b, err := peer.AddrInfo{
				ID:    p,
				Addrs: node.Peerstore().Addrs(p),
			}.MarshalJSON()
			check(err)
			fmt.Println(string(b))
		}
		fmt.Println("/peers")
	}

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

	defer func() { check(node.Close()) }()

	fmt.Println("my id:", node.ID())
	// fmt.Println("listen addresses:", node.Addrs())

	if len(os.Args) == 1 {
		for i := 0; i < 10; i++ {
			fmt.Println()
			addrs, err := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{
				ID:    node.ID(),
				Addrs: node.Addrs(),
			})
			check(err)
			_ = addrs
			fmt.Println("my multiaddrs:")
			for _, a := range addrs {
				fmt.Printf("\t%v\n", a)
			}
			fmt.Println("my listen addresses:")
			for _, a := range node.Addrs() {
				fmt.Printf("\t%v\n", a)
			}
			fmt.Println()
			time.Sleep(2 * time.Second)
		}
		fmt.Println("done looping")
	}
	// fmt.Printf("%v known peers\n", len(peers.Peers()))

	if len(os.Args) > 1 {

		peerid, err := peer.Decode(os.Args[1])
		if err == nil {
			fmt.Println("got peer id")
		} else {
			addr, err := peer.AddrInfoFromString(os.Args[1])
			check(err)
			fmt.Println("got addrinfo")
			check(node.Connect(ctx, *addr))
			fmt.Println("successfully connected")
			peerid = addr.ID
		}

		// stream, err := node.NewStream(ctx, addr.ID, echoProtocol)
		// peerid, err := peer.Decode(os.Args[1])
		// check(err)
		stream, err := node.NewStream(
			network.WithUseTransient(
				ctx,
				"this is a testing connection",
			),
			peerid,
			echoProtocol,
		)

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
