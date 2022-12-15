package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ipfs/go-log"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
)

const echoProtocol = "/echo/1.0.0"

const staticRelays = `
{"ID":"Qmda892EoFn8iS2NaXcGe6CmhFF7xxfC5o4ddJAjjEEzRb","Addrs":["/ip4/127.0.0.1/tcp/45447","/ip4/93.75.110.16/tcp/45447"]}
{"ID":"QmeNKSnjPmecj2Hbakjv237vQRqAxtn8AxCaRZPc4GQb4L","Addrs":["/ip6/::1/tcp/4002/ws","/ip4/10.100.248.14/tcp/1080/ws","/ip6/::1/tcp/4001","/ip4/147.182.216.128/tcp/4002/ws","/ip6/64:ff9b::93b6:d880/tcp/4001","/ip6/2604:a880:400:d0::1cb0:a001/tcp/4002/ws","/ip6/2604:a880:400:d0::1cb0:a001/tcp/4001","/ip4/127.0.0.1/tcp/4002/ws","/ip4/147.182.216.128/tcp/4001","/ip4/127.0.0.1/tcp/4001"]}
{"ID":"QmVH9CYDkYqkx45pSJqxn61PLdKwrD8ejFq9DatCtPF2pU","Addrs":["/ip4/177.197.228.248/tcp/51819","/ip4/127.0.0.1/tcp/51819"]}
{"ID":"QmYkQ9SxH71iT6AttBajMqrrkPx1rmm8eBnRuZuqWDLBxB","Addrs":["/ip4/10.10.0.15/tcp/4001","/ip4/127.0.0.1/tcp/4001","/ip4/10.10.0.15/udp/4001/quic","/ip4/127.0.0.1/udp/4001/quic","/ip4/79.160.225.140/tcp/4001","/ip4/79.160.225.140/udp/4001/quic","/ip6/64:ff9b::4fa0:e18c/tcp/4001","/ip6/::1/tcp/4001"]}
{"ID":"12D3KooWMDUSYdmEC35h92UsPFaKZZFNyg9DJh5o3Frcvxb2s8fo","Addrs":["/ip4/39.122.246.132/tcp/4001","/ip6/64:ff9b::277a:f684/udp/4001/quic","/ip6/::1/tcp/4001","/ip4/127.0.0.1/udp/4001/quic","/ip6/::1/udp/4001/quic","/ip4/127.0.0.1/tcp/4001","/ip4/39.122.246.132/udp/4001/quic"]}
{"ID":"12D3KooWJ8EZcGnRHoCHPYxoCbzQBJXLh6VdHCbmUFhgfMfGWQtv","Addrs":["/ip4/127.0.0.1/udp/4001/quic","/ip6/::1/udp/4001/quic","/ip6/::1/tcp/4001","/ip4/154.53.33.82/tcp/4001","/ip4/154.53.33.82/udp/4001/quic","/ip4/127.0.0.1/tcp/4001"]}
{"ID":"12D3KooWBC9MveRPaP7hT1eeFuuexWyfB6Ywj1rQoHxEGwoXNLzB","Addrs":["/ip4/127.0.0.1/tcp/30006","/ip4/18.221.152.177/udp/30006/quic","/ip4/18.221.152.177/tcp/30006","/ip4/172.31.13.112/udp/30006/quic","/ip4/127.0.0.1/udp/30006/quic","/ip4/172.31.13.112/tcp/30006"]}
{"ID":"12D3KooWGKU9P4KTw7j2SL2TdKM9yd5dqXmVuipCFAD86XC9Dqyv","Addrs":["/ip4/152.44.39.163/tcp/4001","/ip6/64:ff9b::982c:27a3/tcp/4001","/ip4/152.44.39.163/udp/4001/quic","/ip6/64:ff9b::982c:27a3/udp/4001/quic","/ip6/::1/udp/4001/quic","/ip4/127.0.0.1/tcp/4001","/ip4/127.0.0.1/udp/4001/quic","/ip4/10.16.2.102/tcp/4001","/ip6/::1/tcp/4001"]}
{"ID":"12D3KooWPu4L5wTA9bE84q3Vawn6Uq4WsQwitNKys6LrgnP8x6Qz","Addrs":["/ip4/172.31.13.79/udp/30008/quic","/ip4/127.0.0.1/udp/30008/quic","/ip4/127.0.0.1/tcp/30008","/ip4/3.17.176.117/udp/30008/quic","/ip4/3.17.176.117/tcp/30008","/ip4/172.31.13.79/tcp/30008"]}
{"ID":"12D3KooWNr7iF4E9UgbxJYBN8K3nJs45W4QENGzBi1ovy7pATS3Q","Addrs":["/ip6/::1/udp/4001/quic","/ip6/64:ff9b::6dec:548d/udp/4001/quic","/ip4/109.236.84.141/tcp/4001","/ip4/127.0.0.1/tcp/4001","/ip6/::1/tcp/4001","/ip4/109.236.84.141/udp/4001/quic","/ip4/127.0.0.1/udp/4001/quic"]}
{"ID":"QmY6jyH3WpAGHvQdNwmTYoDT6KTBLRi42K8cA2a5hgNWzZ","Addrs":["/ip6/2a02:c207:2027:7100:3a09::19/tcp/4001","/ip6/2a02:c207:2027:7100:6ece::6/tcp/4001","/ip6/2a02:c207:2027:7100:cc1b::9/tcp/4001","/ip6/2a02:c207:2027:7100:8e88::19/tcp/4001","/ip6/::1/udp/4001/quic","/ip6/2a02:c207:2027:7100:5c05::8/tcp/4001","/ip6/2a02:c207:2027:7100:37eb::16/tcp/4001","/ip6/2a02:c207:2027:7100:e6de::3/tcp/4001","/ip4/173.249.15.183/tcp/4001","/ip6/2a02:c207:2027:7100:9efc::16/tcp/4001","/ip6/2a02:c207:2027:7100:7dfc::3/tcp/4001","/ip6/::1/tcp/4001","/ip4/173.249.15.183/udp/4001/quic","/ip4/127.0.0.1/tcp/4001","/ip6/2a02:c207:2027:7100:1e74::16/tcp/4001","/ip6/2a02:c207:2027:7100:9efc::16/udp/4001/quic","/ip6/2a02:c207:2027:7100:2950::17/tcp/4001","/ip6/2a02:c207:2027:7100:760b::10/tcp/4001","/ip4/127.0.0.1/udp/4001/quic"]}
{"ID":"QmZsZkjTaztuVNVHqDp9ytMKGqNWR8JtwcvbUzDVT4CP2t","Addrs":["/ip4/177.234.186.86/tcp/63642","/ip4/10.0.0.134/tcp/43152","/ip4/127.0.0.1/tcp/43152"]}
`

const remoteID = "12D3KooWHze5GkkkCjx8dbCJFzoo7dxvEhEjYWnpkHeNF76GdtQC"

const remote = `
/ip4/152.44.39.163/tcp/4001/p2p/12D3KooWGKU9P4KTw7j2SL2TdKM9yd5dqXmVuipCFAD86XC9Dqyv/p2p-circuit
/ip4/152.44.39.163/udp/4001/quic/p2p/12D3KooWGKU9P4KTw7j2SL2TdKM9yd5dqXmVuipCFAD86XC9Dqyv/p2p-circuit
/ip6/64:ff9b::982c:27a3/udp/4001/quic/p2p/12D3KooWGKU9P4KTw7j2SL2TdKM9yd5dqXmVuipCFAD86XC9Dqyv/p2p-circuit
/ip6/64:ff9b::982c:27a3/tcp/4001/p2p/12D3KooWGKU9P4KTw7j2SL2TdKM9yd5dqXmVuipCFAD86XC9Dqyv/p2p-circuit
/ip6/64:ff9b::6dec:548d/udp/4001/quic/p2p/12D3KooWNr7iF4E9UgbxJYBN8K3nJs45W4QENGzBi1ovy7pATS3Q/p2p-circuit
/ip4/109.236.84.141/tcp/4001/p2p/12D3KooWNr7iF4E9UgbxJYBN8K3nJs45W4QENGzBi1ovy7pATS3Q/p2p-circuit
/ip4/109.236.84.141/udp/4001/quic/p2p/12D3KooWNr7iF4E9UgbxJYBN8K3nJs45W4QENGzBi1ovy7pATS3Q/p2p-circuit
/ip4/154.53.33.82/udp/4001/quic/p2p/12D3KooWJ8EZcGnRHoCHPYxoCbzQBJXLh6VdHCbmUFhgfMfGWQtv/p2p-circuit
/ip4/154.53.33.82/tcp/4001/p2p/12D3KooWJ8EZcGnRHoCHPYxoCbzQBJXLh6VdHCbmUFhgfMfGWQtv/p2p-circuit
/ip4/79.160.225.140/tcp/4001/p2p/QmYkQ9SxH71iT6AttBajMqrrkPx1rmm8eBnRuZuqWDLBxB/p2p-circuit
/ip4/79.160.225.140/udp/4001/quic/p2p/QmYkQ9SxH71iT6AttBajMqrrkPx1rmm8eBnRuZuqWDLBxB/p2p-circuit
/ip6/64:ff9b::4fa0:e18c/tcp/4001/p2p/QmYkQ9SxH71iT6AttBajMqrrkPx1rmm8eBnRuZuqWDLBxB/p2p-circuit
`

func main() {
	ctx := context.Background()

	var relays []peer.AddrInfo
	{
		for _, s := range strings.Split(strings.TrimSpace(staticRelays), "\n") {
			var a peer.AddrInfo
			check(a.UnmarshalJSON([]byte(s)))
			relays = append(relays, a)
		}
	}

	// 	var peers peerstore.Peerstore
	// 	{
	// 		datastore, err := badger.NewDatastore("./peerstore", nil)
	// 		check(err)
	// 		peers, err = pstoreds.NewPeerstore(ctx, datastore, pstoreds.DefaultOpts())
	// 		check(err)
	// 	}

	log.SetAllLoggers(log.LevelInfo)

	// 	libp2p.SetDefaultServiceLimits(&rcmgr.InfiniteLimits)

	// 	r, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits))
	// 	check(err)

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
			autorelay.WithStaticRelays(relays),
		),
		libp2p.ForceReachabilityPrivate(),

		// libp2p.Routing(func(node host.Host) (routing.PeerRouting, error) {
		// 	r, err := dht.New(
		// 		ctx, node,
		// 		// dht.BootstrapPeers(dht.GetDefaultBootstrapPeerAddrInfos()...),
		// 	)
		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	if err := r.Bootstrap(ctx); err != nil {
		// 		return nil, err
		// 	}

		// 	for i, p := range dht.GetDefaultBootstrapPeerAddrInfos() {
		// 		if err := node.Connect(ctx, p); err != nil {
		// 			fmt.Printf("failed to connect to bootstrap node #%v\n", i)
		// 		} else {
		// 			fmt.Println("connected to bootstrap node")
		// 		}
		// 	}
		// 	fmt.Println("done with bootstrapping")

		// 	// err = r.Bootstrap(ctx)
		// 	// if err == nil {
		// 	// err = <-r.ForceRefresh()
		// 	// }
		// 	return r, nil
		// }),

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
		for i := 0; i < 3; i++ {
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
		} else if os.Args[1] == "remote" {
			id, err := peer.Decode(remoteID)
			check(err)
			peerid = id
			var addrs []ma.Multiaddr
			for _, s := range strings.Split(strings.TrimSpace(remote), "\n") {
				a, err := ma.NewMultiaddr(s)
				check(err)
				addrs = append(addrs, a)
			}
			check(node.Connect(ctx, peer.AddrInfo{
				ID:    id,
				Addrs: addrs,
			}))
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
