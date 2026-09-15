package tunnel

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

func RunServer(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer ln.Close()

	var mu sync.Mutex
	nextID := 1
	pending := make(map[uint32]chan net.Conn)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go func(c net.Conn) {
			defer c.Close()

			var first uint32
			if err := binary.Read(c, binary.BigEndian, &first); err != nil {
				return
			}

			mu.Lock()
			ch, isDataBack := pending[first]
			mu.Unlock()

			if isDataBack {
				select {
				case ch <- c:
					return
				default:
					// Buffer already holds a pending connection and no
					// consumer is ready yet; refuse rather than block the
					// handler goroutine forever.
					return
				}
			}

			mu.Lock()
			id := nextID
			nextID++
			publicPort := uint32(20000 + id)
			ch = make(chan net.Conn, 1)
			pending[publicPort] = ch
			mu.Unlock()

			if err := binary.Write(c, binary.BigEndian, uint32(publicPort)); err != nil {
				mu.Lock()
				delete(pending, publicPort)
				mu.Unlock()
				return
			}

			fmt.Printf("tunnel: %s -> port %d\n", c.RemoteAddr(), publicPort)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			defer func() {
				mu.Lock()
				delete(pending, publicPort)
				mu.Unlock()
			}()

			go func() {
				buf := make([]byte, 1)
				for {
					if _, err := io.ReadFull(c, buf); err != nil {
						cancel()
						return
					}
				}
			}()

			dataLn, err := net.Listen("tcp", fmt.Sprintf(":%d", publicPort))
			if err != nil {
				return
			}
			defer dataLn.Close()

			for {
				peerConn, err := dataLn.Accept()
				if err != nil {
					return
				}

				select {
				case <-ctx.Done():
					peerConn.Close()
					return
				default:
				}

				if _, err := c.Write([]byte{1}); err != nil {
					peerConn.Close()
					return
				}

				select {
				case dataBack := <-ch:
					go func() {
						defer peerConn.Close()
						defer dataBack.Close()
						var wg sync.WaitGroup
						wg.Add(2)
						go func() {
							io.Copy(dataBack, peerConn)
							wg.Done()
						}()
						go func() {
							io.Copy(peerConn, dataBack)
							wg.Done()
						}()
						wg.Wait()
					}()
				case <-ctx.Done():
					peerConn.Close()
					return
				case <-time.After(30 * time.Second):
					peerConn.Close()
				}
			}
		}(conn)
	}
}

func RunClient(serverAddr string, localPort int) (int, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return 0, fmt.Errorf("connect to tunnel server: %w", err)
	}

	if err := binary.Write(conn, binary.BigEndian, uint32(localPort)); err != nil {
		conn.Close()
		return 0, fmt.Errorf("send local port: %w", err)
	}

	var publicPort uint32
	if err := binary.Read(conn, binary.BigEndian, &publicPort); err != nil {
		conn.Close()
		return 0, fmt.Errorf("read public port: %w", err)
	}

	go func() {
		defer conn.Close()
		buf := make([]byte, 1)
		for {
			if _, err := io.ReadFull(conn, buf); err != nil {
				return
			}

			go func() {
				dataConn, err := net.Dial("tcp", serverAddr)
				if err != nil {
					return
				}
				defer dataConn.Close()

				if err := binary.Write(dataConn, binary.BigEndian, publicPort); err != nil {
					return
				}

				localConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
				if err != nil {
					return
				}
				defer localConn.Close()

				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					io.Copy(dataConn, localConn)
					wg.Done()
				}()
				go func() {
					io.Copy(localConn, dataConn)
					wg.Done()
				}()
				wg.Wait()
			}()
		}
	}()

	return int(publicPort), nil
}
