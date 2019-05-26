package tiki

import (
	"bufio"
	"net"
)

// Encapsulates a connection plus a room ID for self-tracking of joined rooms.
type player struct {
	net.Conn
	roomID string
}

// Sends a message to this connection.
func (p *player) send(msg string) error {
	writer := bufio.NewWriter(p.Conn)
	_, err := writer.WriteString(msg + "\n")
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}
