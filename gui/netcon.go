package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// syncGracePeriod is how long after a switch we ignore the "currently spectated
// player" reported by GSI, because it still shows the previous player briefly.
const syncGracePeriod = 2 * time.Second

// NetconPort is the TCP port of CS2's network console (launch option
// -netconport <port>). Disabled by default: the CS2 client currently accepts
// the option but never opens the listener. Kept for experiments (-netconport flag).
var NetconPort = 0

// NetconClient sends console commands to CS2 over its network console.
// This is faster and more reliable than simulated key presses: it takes a few
// milliseconds, needs no window focus and never misses a key.
type NetconClient struct {
	mu          sync.Mutex
	port        int
	conn        net.Conn
	lastAttempt time.Time
	warned      bool
}

// NewNetconClient creates a client for the given port (0 = disabled)
func NewNetconClient(port int) *NetconClient {
	return &NetconClient{port: port}
}

// connect opens the TCP connection. Failed attempts are retried at most every
// 5 seconds so a missing -netconport doesn't slow every switch down.
func (n *NetconClient) connect() error {
	if n.conn != nil {
		return nil
	}
	if time.Since(n.lastAttempt) < 5*time.Second {
		return fmt.Errorf("netcon not connected (retrying later)")
	}
	n.lastAttempt = time.Now()

	addr := fmt.Sprintf("127.0.0.1:%d", n.port)
	conn, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
	if err != nil {
		if !n.warned {
			LogInfo("⌨️  Netcon not available on %s - using simulated key presses.", addr)
			LogInfo("   For faster, focus-independent switching add this CS2 launch option: -netconport %d", n.port)
			n.warned = true
		}
		return err
	}
	n.conn = conn
	n.warned = false
	LogInfo("🔌 Connected to CS2 console via netcon (%s) - camera switches no longer need key presses", addr)

	// CS2 streams its console output over this socket. Drain it so the
	// buffer never fills up; when the read fails, the connection is gone.
	go func(c net.Conn) {
		io.Copy(io.Discard, c)
		n.mu.Lock()
		if n.conn == c {
			n.conn = nil
		}
		n.mu.Unlock()
	}(conn)
	return nil
}

// Send sends one console command. Returns an error if netcon is unavailable.
func (n *NetconClient) Send(command string) error {
	if n == nil || n.port <= 0 {
		return fmt.Errorf("netcon disabled")
	}
	n.mu.Lock()
	defer n.mu.Unlock()

	for attempt := 0; attempt < 2; attempt++ {
		if err := n.connect(); err != nil {
			return err
		}
		n.conn.SetWriteDeadline(time.Now().Add(300 * time.Millisecond))
		if _, err := n.conn.Write([]byte(command + "\n")); err != nil {
			// Connection dropped (e.g. CS2 restarted): reconnect once right away
			n.conn.Close()
			n.conn = nil
			n.lastAttempt = time.Time{}
			continue
		}
		return nil
	}
	return fmt.Errorf("netcon write failed")
}
