package util

import (
	"context"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
)

const (
	DefaultICMPProtocol = "ip4:icmp"
	DefaultIPProtocol   = "ip4"

	DefaultLocalAddress = "0.0.0.0"

	DefaultICMPDataPacket     = "ping"
	DefaultICMPPacketCode     = 0
	DefaultICMPPacketID       = 1
	DefaultICMPPacketSequence = 1
	DefaultICMPResponseBuffer = 1500
)

// todo(): adjust thsi function so that ctx is propagated to the net.ResolveIPAddr func too
// todo(): make this func more readable, right now is kinda all over the place without a clear structure
// Ping sends an ICMP Echo Request and waits for a reply with context support.
func Ping(ctx context.Context, address string) error {
	// Open a raw ICMP connection for receiving replies
	conn, err := icmp.ListenPacket(DefaultICMPProtocol, DefaultLocalAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Resolve the destination address
	dst, err := net.ResolveIPAddr(DefaultIPProtocol, address)
	if err != nil {
		return err
	}

	// Use a channel to handle cancellation
	errChan := make(chan error, 1)

	go func() {
		errChan <- sendICMPEchoRequest(conn, dst)
	}()

	// Make sure that the request sent does not block the context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		if err != nil {
			return err
		}
	}

	// Wait for the reply
	go func() {
		errChan <- func() error {
			parsedMsg, err := receiveICMPEchoReply(conn)
			if err != nil {
				return err
			}

			if parsedMsg.Type == ipv4.ICMPTypeEchoReply {
				return nil
			}
			return fmt.Errorf("received unexpected ICMP message: %v", parsedMsg)
		}()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

// sendICMPEchoRequest sends an ICMP Echo Request to the given destination.
func sendICMPEchoRequest(conn *icmp.PacketConn, dst *net.IPAddr) error {
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: DefaultICMPPacketCode,
		Body: &icmp.Echo{
			ID:   DefaultICMPPacketID,
			Seq:  DefaultICMPPacketSequence,
			Data: []byte(DefaultICMPDataPacket),
		},
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return err
	}

	_, err = conn.WriteTo(msgBytes, dst)
	return err
}

// receiveICMPEchoReply waits for an ICMP Echo Reply and parses it.
func receiveICMPEchoReply(conn *icmp.PacketConn) (*icmp.Message, error) {
	reply := make([]byte, DefaultICMPResponseBuffer)

	n, _, err := conn.ReadFrom(reply)
	if err != nil {
		return nil, err
	}

	return icmp.ParseMessage(1, reply[:n])
}
