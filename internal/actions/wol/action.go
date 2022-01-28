package wol

import (
	"fmt"
	"net"
)

type wolAction struct {
	broadcastAddress   string
	broadcastInterface string
	udpPort            string
	targetMACAddress   string
}

func New(targetMACAddress string) wolAction {
	return wolAction{
		broadcastAddress:   "255.255.255.255",
		broadcastInterface: "",
		udpPort:            "9",
		targetMACAddress:   targetMACAddress,
	}
}

func (wa *wolAction) WakeUp() error {
	// Populate the local address in the event that the broadcast interface has
	// been set.
	var localAddr *net.UDPAddr
	var err error
	if wa.broadcastInterface != "" {
		localAddr, err = ipFromInterface(wa.broadcastInterface)
		if err != nil {
			return fmt.Errorf("obtaining address from broadcast interface: %w", err)
		}
	}

	magicPacket, err := newMagicPacket(wa.targetMACAddress)
	if err != nil {
		return fmt.Errorf("invalid server configuration: %w", err)
	}
	// Grab a stream of bytes to send.
	bs, err := magicPacket.Marshal()
	if err != nil {
		return fmt.Errorf("error creating magic packet: %w", err)
	}

	bcastAddr := fmt.Sprintf("%s:%s", wa.broadcastAddress, wa.udpPort)
	udpAddr, err := net.ResolveUDPAddr("udp", bcastAddr)
	if err != nil {
		return fmt.Errorf("error resolving broadcast address: %w", err)
	}

	// Grab a UDP connection to send our packet of bytes.
	conn, err := net.DialUDP("udp", localAddr, udpAddr)
	if err != nil {
		return fmt.Errorf("error opening udp connection: %w", err)
	}
	defer conn.Close()

	n, err := conn.Write(bs)
	if err == nil && n != 102 {
		return fmt.Errorf("magic packet sent was %d bytes (expected 102 bytes sent)", n)
	}
	if err != nil {
		return fmt.Errorf("error sending magic packet: %w", err)
	}

	return nil
}

func ipFromInterface(iface string) (*net.UDPAddr, error) {
	ief, err := net.InterfaceByName(iface)
	if err != nil {
		return nil, err
	}

	addrs, err := ief.Addrs()
	if err == nil && len(addrs) <= 0 {
		err = fmt.Errorf("no address associated with interface %s", iface)
	}
	if err != nil {
		return nil, err
	}

	// Validate that one of the addrs is a valid network IP address.
	for _, addr := range addrs {
		switch ip := addr.(type) {
		case *net.IPNet:
			if !ip.IP.IsLoopback() && ip.IP.To4() != nil {
				return &net.UDPAddr{
					IP: ip.IP,
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("no address associated with interface %s", iface)
}
