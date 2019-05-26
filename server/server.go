// Package tiki implements a simple tic-tac-toe server that accepts line-by-line
// inputs from a tcp client.
package tiki

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

const helpQuote = `
Commands are:

JOIN <r>  join room <r> (and quit the current one if already joined)
MARK <n>  mark square <n>, where squares are numbered like in the following diagram:
           1 | 2 | 3
          ---+---+---
           4 | 5 | 6
          ---+---+---
           7 | 8 | 9
QUIT      close the current connection
`

// Server implements a multi-room, multi-tenant tictactoe server.
type Server struct {
	rooms sync.Map
	msgch chan string
}

// NewServer creates a new tiki.Server instance.
func NewServer() *Server {
	return &Server{
		msgch: make(chan string, 10),
	}
}

// ListenAndServe creates a TCP listener and waits for new connections. This
// method is blocking.
func (s *Server) ListenAndServe(address string) error {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("error: accept connection:", err)
			continue
		}

		log.Println(conn, "has connected")
		go func(conn net.Conn, out chan<- string) {
			scanner := bufio.NewScanner(conn)
			p := &player{Conn: conn}

			for scanner.Scan() {
				line := strings.Trim(scanner.Text(), "\n\t ")
				err := s.parseAndExecute(p, line)
				if err != nil {
					p.send(err.Error())
					log.Println("error:", err)
				}
			}

			// Ensure proper closure if client did an abrupt close (e.g. CTRL+C).
			s.parseAndExecute(p, "QUIT")
			log.Println(conn, "has disconnected")
		}(conn, s.msgch)
	}
}

func (s *Server) createRoom(roomID string) *room {
	room := newRoom(roomID)
	go room.run()
	s.rooms.Store(roomID, room)
	return room
}

func (s *Server) parseAndExecute(p *player, msg string) error {
	tokens := strings.Split(msg, " ")
	if len(tokens) == 0 {
		return nil
	}

	instrc := strings.ToUpper(tokens[0])
	switch instrc {
	case "JOIN":
		if len(tokens) != 2 {
			return errInvalidParams
		}

		// If joining the same room, don't even bother.
		if p.roomID == tokens[1] {
			return errRoomIsSame
		}

		// Leave current room, if it exists.
		if p.roomID != "" {
			r, ok := s.rooms.Load(p.roomID)
			if ok {
				err := r.(*room).execute(action{kind: actionLeave, src: p})
				if err != nil {
					return err
				}
				p.roomID = ""
			}
		}

		// Enter the new room.
		r, ok := s.rooms.Load(tokens[1])
		if !ok {
			r = s.createRoom(tokens[1])
		}

		err := r.(*room).execute(action{kind: actionJoin, src: p})
		if err != nil {
			return err
		}
		p.roomID = tokens[1]
	case "MARK":
		if len(tokens) != 2 {
			return errInvalidParams
		}
		r, ok := s.rooms.Load(p.roomID)
		if !ok {
			return errPlayerNoRoom
		}

		idx, err := strconv.Atoi(tokens[1])
		if err != nil {
			return err
		}

		err = r.(*room).execute(action{kind: actionMarkBoard, src: p, idx: idx})
		if err != nil {
			return err
		}
	case "QUIT":
		if p.roomID != "" {
			r, ok := s.rooms.Load(p.roomID)
			if !ok {
				return errRoomNotFound
			}
			r.(*room).execute(action{kind: actionQuit, src: p})
		}
		p.Close()
	case "!":
		r, ok := s.rooms.Load(p.roomID)
		if !ok {
			return errPlayerNoRoom
		}
		r.(*room).broadcast(fmt.Sprintf("%v: %s", p, msg[1:]))
	default:
		p.send(helpQuote)
	}
	return nil
}
