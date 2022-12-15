package main

import (
	"context"
	"fmt"
	"os"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/event"
)

func main() {
	ctx := context.Background()
	node, err := libp2p.New()
	check(err)
	defer func() { check(node.Close()) }()

	sub, err := node.EventBus().Subscribe(new(event.EvtPeerIdentificationCompleted))
	check(err)
	defer sub.Close()
	go func() {
		for e := range sub.Out() {
			p := e.(event.EvtPeerIdentificationCompleted).Peer
			fmt.Printf("identified %v\n", p)
			fmt.Println(node.Peerstore().GetProtocols(p))
		}
	}()

	for i, p := range dht.GetDefaultBootstrapPeerAddrInfos() {
		if err := node.Connect(ctx, p); err != nil {
			fmt.Printf("failed to connect to bootstrap node #%v\n", i)
		} else {
			fmt.Println("connected to bootstrap node")
		}
	}
	fmt.Println("done with bootstrapping")

	fmt.Println("sleeping 2 seconds then exiting")
	time.Sleep(2 * time.Second)
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
