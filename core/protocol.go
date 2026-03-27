package core

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net"
	"os"
	"runtime/debug"
	"time"

	tls "github.com/refraction-networking/utls"
)

func DumpHex(buf []byte) {
	stdoutDumper := hex.Dumper(os.Stdout)
	defer stdoutDumper.Close()
	stdoutDumper.Write(buf)
}

func TLSConn(server string) (*tls.UConn, error) {
	return TLSConnWithSessionId(server, []byte{'L', '3', 'I', 'P', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}

func TLSConnWithSessionId(server string, sessionId []byte) (*tls.UConn, error) {
	// dial vpn server
	dialConn, err := net.Dial("tcp", server)
	if err != nil {
		return nil, err
	}

	// Enable TCP keepalive to prevent idle connection from being dropped
	if tcpConn, ok := dialConn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	log.Println("socket: connected to: ", dialConn.RemoteAddr())

	// using uTLS to construct a weird TLS Client Hello (required by Sangfor)
	// The VPN and HTTP Server share port 443, Sangfor uses a special SessionId to distinguish them. (which is very stupid...)
	conn := tls.UClient(dialConn, &tls.Config{InsecureSkipVerify: true}, tls.HelloCustom)

	random := make([]byte, 32)
	rand.Read(random) // Ignore the err
	conn.SetClientRandom(random)
	conn.SetTLSVers(tls.VersionTLS11, tls.VersionTLS11, []tls.TLSExtension{})
	conn.HandshakeState.Hello.Vers = tls.VersionTLS11
	conn.HandshakeState.Hello.CipherSuites = []uint16{tls.TLS_RSA_WITH_RC4_128_SHA, tls.FAKE_TLS_EMPTY_RENEGOTIATION_INFO_SCSV}
	conn.HandshakeState.Hello.CompressionMethods = []uint8{0}
	conn.HandshakeState.Hello.SessionId = sessionId

	log.Println("tls: connected to: ", conn.RemoteAddr())

	return conn, nil
}

// TLS 1.0 connection for heartbeats (TIMQ/JJYY)
func TLSConnHeartbeat(server string, sessionId []byte) (*tls.UConn, error) {
	dialConn, err := net.Dial("tcp", server)
	if err != nil {
		return nil, err
	}

	if tcpConn, ok := dialConn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	log.Println("socket: connected to: ", dialConn.RemoteAddr())

	conn := tls.UClient(dialConn, &tls.Config{InsecureSkipVerify: true}, tls.HelloCustom)

	random := make([]byte, 32)
	rand.Read(random)
	conn.SetClientRandom(random)
	conn.SetTLSVers(tls.VersionTLS10, tls.VersionTLS10, []tls.TLSExtension{})
	conn.HandshakeState.Hello.Vers = tls.VersionTLS10
	conn.HandshakeState.Hello.CipherSuites = []uint16{tls.TLS_RSA_WITH_RC4_128_SHA, tls.FAKE_TLS_EMPTY_RENEGOTIATION_INFO_SCSV}
	conn.HandshakeState.Hello.CompressionMethods = []uint8{0}
	conn.HandshakeState.Hello.SessionId = sessionId

	log.Println("tls: connected to: ", conn.RemoteAddr())

	return conn, nil
}

func QueryIp(server string, token *[48]byte) ([]byte, *tls.UConn, error) {
	conn, err := TLSConn(server)
	if err != nil {
		debug.PrintStack()
		return nil, nil, err
	}
	// defer conn.Close()
	// Query IP conn CAN NOT be closed, otherwise tx/rx handshake will fail

	// QUERY IP PACKET
	message := []byte{0x00, 0x00, 0x00, 0x00}
	message = append(message, token[:]...)
	message = append(message, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff}...)

	n, err := conn.Write(message)
	if err != nil {
		debug.PrintStack()
		return nil, nil, err
	}

	log.Printf("query ip: wrote %d bytes", n)
	DumpHex(message[:n])

	reply := make([]byte, 0x80)
	n, err = conn.Read(reply)
	if err != nil {
		debug.PrintStack()
		return nil, nil, err
	}

	log.Printf("query ip: read %d bytes", n)
	DumpHex(reply[:n])

	if reply[0] != 0x00 {
		debug.PrintStack()
		return nil, nil, errors.New("unexpected query ip reply")
	}

	return reply[4:8], conn, nil
}

func BlockRXStream(server string, token *[48]byte, ipRev *[4]byte, ep *EasyConnectEndpoint, debug bool) error {
	conn, err := TLSConn(server)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// RECV STREAM START
	message := []byte{0x06, 0x00, 0x00, 0x00}
	message = append(message, token[:]...)
	message = append(message, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}...)
	message = append(message, ipRev[:]...)

	n, err := conn.Write(message)
	if err != nil {
		return err
	}
	log.Printf("recv handshake: wrote %d bytes", n)
	DumpHex(message[:n])

	reply := make([]byte, 1500)
	n, err = conn.Read(reply)
	if err != nil {
		return err
	}
	log.Printf("recv handshake: read %d bytes", n)
	DumpHex(reply[:n])

	if reply[0] != 0x01 {
		return errors.New("unexpected recv handshake reply")
	}

	for {
		n, err = conn.Read(reply)

		if err != nil {
			return err
		}

		ep.WriteTo(reply[:n])

		if debug {
			log.Printf("recv: read %d bytes", n)
			DumpHex(reply[:n])
		}
	}
}

func BlockTXStream(server string, token *[48]byte, ipRev *[4]byte, ep *EasyConnectEndpoint, debug bool) error {
	conn, err := TLSConn(server)
	if err != nil {
		return err
	}
	defer conn.Close()

	// SEND STREAM START
	message := []byte{0x05, 0x00, 0x00, 0x00}
	message = append(message, token[:]...)
	message = append(message, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}...)
	message = append(message, ipRev[:]...)

	n, err := conn.Write(message)
	if err != nil {
		return err
	}
	log.Printf("send handshake: wrote %d bytes", n)
	DumpHex(message[:n])

	reply := make([]byte, 1500)
	n, err = conn.Read(reply)
	if err != nil {
		return err
	}
	log.Printf("send handshake: read %d bytes", n)
	DumpHex(reply[:n])

	if reply[0] != 0x02 {
		return errors.New("unexpected send handshake reply")
	}

	errCh := make(chan error)

	ep.OnRecv = func(buf []byte) {
		var n, err = conn.Write(buf)
		if err != nil {
			errCh <- err
			return
		}

		if debug {
			log.Printf("send: wrote %d bytes", n)
			DumpHex([]byte(buf[:n]))
		}
	}

	return <-errCh
}

func BlockTIMQHeartbeat(server string, token *[48]byte, serverSessionId []byte, debug bool) error {
	log.Println("TIMQ: starting heartbeat connection...")

	// TIMQ SessionId = twfId(16B ASCII) + raw ServerHello SessionId(16B)
	twfId := token[32:48]
	timqSessionId := make([]byte, 32)
	copy(timqSessionId[0:16], twfId)
	copy(timqSessionId[16:32], serverSessionId)

	conn, err := TLSConnHeartbeat(server, timqSessionId)
	if err != nil {
		log.Printf("TIMQ: TLSConn failed: %s", err.Error())
		return err
	}
	defer conn.Close()
	log.Println("TIMQ: TLS connection established")

	tokenAscii := token[32:48]
	seq := uint32(0)

	// Initial handshake (type 0x04)
	message := []byte{'T', 'I', 'M', 'Q', 0x00, 0x00, 0x00, 0x04}
	message = append(message, byte(seq>>24), byte(seq>>16), byte(seq>>8), byte(seq))
	message = append(message, tokenAscii...)

	n, err := conn.Write(message)
	if err != nil {
		return err
	}
	if debug {
		log.Printf("TIMQ handshake: wrote %d bytes", n)
		DumpHex(message[:n])
	}

	reply := make([]byte, 100)
	n, err = conn.Read(reply)
	if err != nil {
		return err
	}
	if debug {
		log.Printf("TIMQ handshake: read %d bytes", n)
		DumpHex(reply[:n])
	}

	if n < 4 || string(reply[0:4]) != "ACKQ" {
		log.Printf("TIMQ: unexpected handshake reply (%d bytes): %x", n, reply[:n])
		return errors.New("unexpected TIMQ handshake reply")
	}
	log.Println("TIMQ: handshake successful, starting heartbeat loop")

	// Heartbeat loop (type 0x03, every 5 seconds)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		seq += 5
		message = []byte{'T', 'I', 'M', 'Q', 0x00, 0x00, 0x00, 0x03}
		message = append(message, byte(seq>>24), byte(seq>>16), byte(seq>>8), byte(seq))
		message = append(message, tokenAscii...)

		n, err = conn.Write(message)
		if err != nil {
			return err
		}

		n, err = conn.Read(reply)
		if err != nil {
			return err
		}

		if debug {
			log.Printf("TIMQ heartbeat: seq=%d, wrote/read %d bytes", seq, n)
		}
	}

	return nil
}

func BlockJJYYHeartbeat(server string, token *[48]byte, serverSessionId []byte, debug bool) error {
	log.Println("JJYY: starting heartbeat connection...")

	// JJYY uses the same SessionId format as RX/TX: L3IP prefix + serverSessionId suffix
	jjyySessionId := make([]byte, 32)
	copy(jjyySessionId[0:4], []byte{'L', '3', 'I', 'P'})
	copy(jjyySessionId[16:32], serverSessionId)

	conn, err := TLSConnHeartbeat(server, jjyySessionId)
	if err != nil {
		log.Printf("JJYY: TLSConn failed: %s", err.Error())
		return err
	}
	defer conn.Close()
	log.Println("JJYY: TLS connection established")

	tokenAscii := token[32:48]

	// Initial handshake (40 bytes: magic + type + zeros + tokenAscii)
	message := []byte{'J', 'J', 'Y', 'Y', 0x00, 0x00, 0x00, 0x00}
	message = append(message, make([]byte, 16)...)
	message = append(message, tokenAscii...)

	log.Printf("JJYY: sending handshake (%d bytes)", len(message))

	n, err := conn.Write(message)
	if err != nil {
		return err
	}
	if debug {
		log.Printf("JJYY handshake: wrote %d bytes", n)
		DumpHex(message[:n])
	}

	reply := make([]byte, 100)
	n, err = conn.Read(reply)
	if err != nil {
		return err
	}
	if debug {
		log.Printf("JJYY handshake: read %d bytes", n)
		DumpHex(reply[:n])
	}

	if n < 4 || string(reply[0:4]) != "AABB" {
		log.Printf("JJYY: unexpected handshake reply (%d bytes): %x", n, reply[:n])
		return errors.New("unexpected JJYY handshake reply")
	}
	log.Println("JJYY: handshake successful, starting heartbeat loop")

	// Heartbeat loop (every 35 seconds)
	ticker := time.NewTicker(35 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		message = []byte{'J', 'J', 'Y', 'Y', 0x00, 0x00, 0x00, 0x03}
		message = append(message, make([]byte, 32)...)
		message = append(message, tokenAscii...)
		message = append(message, make([]byte, 12)...)

		n, err = conn.Write(message)
		if err != nil {
			return err
		}

		n, err = conn.Read(reply)
		if err != nil {
			return err
		}

		if debug {
			log.Printf("JJYY heartbeat: wrote/read %d bytes", n)
		}
	}

	return nil
}

func StartProtocol(endpoint *EasyConnectEndpoint, server string, token *[48]byte, serverSessionId []byte, ipRev *[4]byte, debug bool) {
	RX := func() {
		counter := 0
		for counter < 5 {
			err := BlockRXStream(server, token, ipRev, endpoint, debug)
			if err != nil {
				log.Printf("Error occurred while recv (attempt %d/5), retrying in %ds: %s", counter+1, (counter+1)*2, err.Error())
				time.Sleep(time.Duration((counter+1)*2) * time.Second)
			}
			counter += 1
		}
		panic("recv retry limit exceeded.")
	}

	go RX()

	TX := func() {
		counter := 0
		for counter < 5 {
			err := BlockTXStream(server, token, ipRev, endpoint, debug)
			if err != nil {
				log.Printf("Error occurred while send (attempt %d/5), retrying in %ds: %s", counter+1, (counter+1)*2, err.Error())
				time.Sleep(time.Duration((counter+1)*2) * time.Second)
			}
			counter += 1
		}
		panic("send retry limit exceeded.")
	}

	go TX()

	// Heartbeats are currently disabled as the server rejects TLS 1.0 connections
	// TCP keepalive (30s) is sufficient to prevent EOF disconnections
	/*
	TIMQ := func() {
		counter := 0
		for counter < 5 {
			err := BlockTIMQHeartbeat(server, token, serverSessionId, debug)
			if err != nil {
				log.Printf("Error occurred while TIMQ heartbeat (attempt %d/5), retrying in %ds: %s", counter+1, (counter+1)*2, err.Error())
				time.Sleep(time.Duration((counter+1)*2) * time.Second)
			}
			counter += 1
		}
		log.Println("TIMQ heartbeat retry limit exceeded.")
	}

	go TIMQ()

	JJYY := func() {
		counter := 0
		for counter < 5 {
			err := BlockJJYYHeartbeat(server, token, serverSessionId, debug)
			if err != nil {
				log.Printf("Error occurred while JJYY heartbeat (attempt %d/5), retrying in %ds: %s", counter+1, (counter+1)*2, err.Error())
				time.Sleep(time.Duration((counter+1)*2) * time.Second)
			}
			counter += 1
		}
		log.Println("JJYY heartbeat retry limit exceeded.")
	}

	go JJYY()
	*/
}
