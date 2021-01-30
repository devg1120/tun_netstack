package main

import (
	"fmt"
	//"flag"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	//"strings"
	"time"

	"github.com/google/netstack/tcpip"
	"github.com/google/netstack/tcpip/link/fdbased"
	"github.com/google/netstack/tcpip/link/rawfile"
	"github.com/google/netstack/tcpip/link/tun"
	//"github.com/google/netstack/tcpip/network/arp"
	"github.com/google/netstack/tcpip/network/ipv4"
	"github.com/google/netstack/tcpip/network/ipv6"
	"github.com/google/netstack/tcpip/stack"
	"github.com/google/netstack/tcpip/transport/tcp"
	//"github.com/google/netstack/tcpip/transport/icmp"

	"github.com/google/netstack/waiter"

	//"github.com/songgao/water"
	"os/exec"
	"test/tcpsetup"
	//"errors"
)

func myecho(wq *waiter.Queue, ep tcpip.Endpoint) {

	//info := fmt.Sprintf("%s", ep.Info())
	defer ep.Close()

	// Create wait queue entry that notifies a channel.
	waitEntry, notifyCh := waiter.NewChannelEntry(nil)

	wq.EventRegister(&waitEntry, waiter.EventIn)
	defer wq.EventUnregister(&waitEntry)

	for {
		//v, _, err := ep.Read(nil)
		_, _, err := ep.Read(nil)
		if err != nil {
			if err == tcpip.ErrWouldBlock {
				<-notifyCh
				continue
			}

			return
		}
		//ep.Write(tcpip.SlicePayload(v), tcpip.WriteOptions{})
		ep.Write(tcpip.SlicePayload("MY ECHO OK!"), tcpip.WriteOptions{})
		//ep.Write(tcpip.SlicePayload(info), tcpip.WriteOptions{})
	}
}

func fatalIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func runBin(bin string, args ...string) {
	cmd := exec.Command(bin, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	fatalIf(cmd.Run())
}

func tundel() {
	fmt.Println("\ntundel...")
	runBin("/bin/ip", "link", "set", "dev", "gusa", "down")
	runBin("/bin/ip", "link", "delete", "dev", "gusa")
	runBin("/bin/ip", "tuntap", "del", "mode", "tun", "gusa")
}

func tunsetup() {
	time.Sleep(1000)
	runBin("/bin/ip", "tuntap", "add", "user", "gusa1120", "mode", "tun", "gusa")
	runBin("/bin/ip", "link", "set", "dev", "gusa", "mtu", "1500")
	runBin("/bin/ip", "addr", "add", "192.168.100.1/24", "dev", "gusa")
	runBin("/bin/ip", "link", "set", "dev", "gusa", "up")
	//runBin2("/bin/ip",fmt.Sprintf("route add 10.0.0.0/8 dev %s", iface.Name()))
	runBin("/bin/ip", "route", "add", "10.0.0.0/8", "via", "192.168.100.254")
}

func exit() {
	tundel()
	os.Exit(3)
}

func set_addr(addrName string) (tcpip.Address, tcpip.NetworkProtocolNumber, error) {

	// Parse the IP address. Support both ipv4 and ipv6.
	parsedAddr := net.ParseIP(addrName)
	if parsedAddr == nil {
		tundel()
		log.Fatalf("Bad IP address: %v", addrName)
	}

	var addr tcpip.Address
	var proto tcpip.NetworkProtocolNumber
	if parsedAddr.To4() != nil {
		addr = tcpip.Address(parsedAddr.To4())
		proto = ipv4.ProtocolNumber
	} else if parsedAddr.To16() != nil {
		addr = tcpip.Address(parsedAddr.To16())
		proto = ipv6.ProtocolNumber
	} else {
		tundel()
		log.Fatalf("Unknown IP type: %v", addrName)
	}

	return addr, proto, nil
}

func make_stack(
	proto tcpip.NetworkProtocolNumber,
	addr tcpip.Address,
	addr2 tcpip.Address,
	linkEP stack.LinkEndpoint) *stack.Stack {

	s := stack.New(stack.Options{
		NetworkProtocols: []stack.NetworkProtocol{ipv4.NewProtocol()},
		//TransportProtocols: []stack.TransportProtocol{tcp.NewProtocol(),icmp.NewProtocol4()},
		TransportProtocols: []stack.TransportProtocol{tcp.NewProtocol()},
	})
	if err := s.CreateNIC(1, linkEP); err != nil {
		tundel()
		log.Fatal(err)
	}

	if err := s.AddAddress(1, proto, addr); err != nil {
		tundel()
		log.Fatal(err)
	}
	if err := s.AddAddress(1, proto, addr2); err != nil {
		tundel()
		log.Fatal(err)
	}
	return s

}

func make_stack2(
	proto tcpip.NetworkProtocolNumber,
	addr tcpip.Address,
	linkEP stack.LinkEndpoint) *stack.Stack {

	s := stack.New(stack.Options{
		NetworkProtocols: []stack.NetworkProtocol{ipv4.NewProtocol()},
		//TransportProtocols: []stack.TransportProtocol{tcp.NewProtocol(),icmp.NewProtocol4()},
		TransportProtocols: []stack.TransportProtocol{tcp.NewProtocol()},
	})
	if err := s.CreateNIC(1, linkEP); err != nil {
		tundel()
		log.Fatal(err)
	}

	if err := s.AddAddress(1, proto, addr); err != nil {
		tundel()
		log.Fatal(err)
	}
	return s

}

func main() {

	tunsetup()
	tunName := "gusa"
	addrName := "10.1.1.1"
	addrName2 := "10.1.1.2"
	portName := "7000"
	portName2 := "7001"

	rand.Seed(time.Now().UnixNano())

	addr, proto, err := set_addr(addrName)
	if err != nil {
		tundel()
		log.Fatalf("Unable to convert port %v: %v", portName, err)
	}
	addr2, proto, err := set_addr(addrName2)
	if err != nil {
		tundel()
		log.Fatalf("Unable to convert port %v: %v", portName, err)
	}

	localPort, err := strconv.Atoi(portName)
	if err != nil {
		tundel()
		log.Fatalf("Unable to convert port %v: %v", portName, err)
	}

	localPort2, err := strconv.Atoi(portName2)
	if err != nil {
		tundel()
		log.Fatalf("Unable to convert port %v: %v", portName, err)
	}
	/*
		// Create the stack with ip and tcp protocols, then add a tun-based
		// NIC and address.
		s := stack.New(stack.Options{
			//NetworkProtocols:   []stack.NetworkProtocol{ipv4.NewProtocol(), ipv6.NewProtocol(), arp.NewProtocol()},
			//NetworkProtocols:   []stack.NetworkProtocol{ipv4.NewProtocol(),  arp.NewProtocol()},

			NetworkProtocols: []stack.NetworkProtocol{ipv4.NewProtocol()},
			//TransportProtocols: []stack.TransportProtocol{tcp.NewProtocol(),icmp.NewProtocol4()},
			TransportProtocols: []stack.TransportProtocol{tcp.NewProtocol()},
		})
	*/
	mtu, err := rawfile.GetMTU(tunName)
	if err != nil {
		tundel()
		log.Fatal(err)
	}

	var fd int
	fd, err = tun.Open(tunName)
	if err != nil {
		tundel()
		log.Fatal(err)
	}

	linkEP, err := fdbased.New(&fdbased.Options{
		FDs: []int{fd},
		MTU: mtu,
		//EthernetHeader: *tap,
		//Address:        tcpip.LinkAddress(maddr),
	})
	if err != nil {
		tundel()
		log.Fatal(err)
	}
	/*
		if err := s.CreateNIC(1, linkEP); err != nil {
			tundel()
			log.Fatal(err)
		}

		if err := s.AddAddress(1, proto, addr); err != nil {
			tundel()
			log.Fatal(err)
		}

		if err := s.AddAddress(1, proto, addr2); err != nil {
			tundel()
			log.Fatal(err)
		}
	*/

	s := make_stack(proto, addr, addr2, linkEP)
	go tcpsetup.TcpPortBind(s, proto, localPort, myecho)
	go tcpsetup.TcpPortBind(s, proto, localPort2, tcpsetup.Echo)

        /*
	s1 := make_stack2(proto, addr, linkEP)
	s2 := make_stack2(proto, addr2, linkEP2)
	go tcpsetup.TcpPortBind(s1, proto, localPort, myecho)
	go tcpsetup.TcpPortBind(s2, proto, localPort2, tcpsetup.Echo)
        */

	// シグナル用のチャネル定義
	quit := make(chan os.Signal)

	// 受け取るシグナルを設定
	signal.Notify(quit, os.Interrupt)

	<-quit // ここでシグナルを受け取るまで以降の処理はされない
	tundel()
}
