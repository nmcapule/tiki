package tiki

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
)

const (
	actionJoin = iota
	actionQuit
	actionLeave
	actionMarkBoard
	actionResetBoard
	actionPrintBoard
	actionStartGame
	actionWaitForInput
)

// Encapsulates an actionable event for a room instance.
type action struct {
	kind int
	src  *player
	idx  int
	msg  string
}

// Encapsulates a single tictactoe room implementation limited for 2 players.
type room struct {
	mu      sync.Mutex
	id      string
	board   [9]byte
	players []*player
	turn    int
	actch   chan action
	stop    chan bool
}

func newRoom(id string) *room {
	return &room{
		id:    id,
		actch: make(chan action, 10),
	}
}

// Implements the logic loop for the room instance.
func (r *room) run() error {
	for {
		select {
		case act := <-r.actch:
			err := r.execute(act)
			if err != nil {
				if act.src != nil {
					act.src.send(err.Error())
				}
				log.Println("error:", err)
			}
		case <-r.stop:
			break
		}
	}
	return nil
}

// Single point of entry for executing given room action.
func (r *room) execute(act action) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch act.kind {
	case actionResetBoard:
		r.broadcast("Resetting the board.")
		r.reset()
		r.actch <- action{kind: actionStartGame}
	case actionStartGame:
		if !r.checkReady() {
			return errRoomNotReady
		}
		r.shuffle()
		pidx := r.turn
		oidx := (r.turn + 1) % 2
		r.players[pidx].send(fmt.Sprintf("You mark with '%c'", r.charOf(pidx)))
		r.players[oidx].send(fmt.Sprintf("You mark with '%c'", r.charOf(oidx)))
		r.broadcast(r.render())
		r.actch <- action{kind: actionWaitForInput}
	case actionWaitForInput:
		if !r.checkReady() {
			return errRoomNotReady
		}
		pidx := r.turn
		oidx := (r.turn + 1) % 2
		r.players[pidx].send("It's your turn.")
		r.players[oidx].send("Waiting for the other player to move...")
	case actionJoin:
		if err := r.join(act.src); err != nil {
			return err
		}
		if r.checkReady() {
			r.actch <- action{kind: actionStartGame}
		} else {
			for _, p := range r.players {
				p.send(fmt.Sprintf("Waiting for the other player to join..."))
			}
		}
	case actionLeave, actionQuit:
		if err := r.leave(act.src); err != nil {
			return err
		}
		r.broadcast("The other player left the game.")
		r.actch <- action{kind: actionResetBoard}
	case actionMarkBoard:
		if !r.checkReady() {
			return errRoomNotReady
		}
		if r.checkTurn() != act.src {
			return errInvalidTurn
		}
		pidx, ok := r.indexOf(act.src)
		if !ok {
			return errInvalidPlayer
		}
		err := r.mark(act.idx, r.charOf(pidx))
		if err != nil {
			return err
		}
		err = r.broadcast(r.render())
		if err != nil {
			return err
		}

		// Index of the other player.
		oidx := (pidx + 1) % 2
		_, win := r.checkWin()
		if win {
			r.players[pidx].send("Oh hey you won! Congrats!")
			r.players[oidx].send("Oops! Sorry bud better luck next time!")
			r.actch <- action{kind: actionResetBoard}
		} else {
			// Set next turn.
			r.turn = oidx
			r.actch <- action{kind: actionWaitForInput}
		}
	}
	return nil
}

// Creates a string rendering of the currrent state of the board.
func (r *room) render() string {
	var args []interface{}
	for _, c := range r.board {
		// Display uninitialized cells as '.'.
		if c == 0 {
			c = '.'
		}
		args = append(args, c)
	}
	return fmt.Sprintf("%c%c%c\n%c%c%c\n%c%c%c\n", args...)
}

// Randomize character assignment and first turn player.
func (r *room) shuffle() {
	// Randomize turn.
	r.turn = rand.Int() % 2
	// Randomize list to randomize assigned tic tac.
	rand.Shuffle(len(r.players), func(i, j int) {
		r.players[i], r.players[j] = r.players[j], r.players[i]
	})
}

// Resetsthe state of the board.
func (r *room) reset() {
	r.board = [9]byte{}
	r.turn = 0
}

func (r *room) broadcast(msg string) error {
	for _, p := range r.players {
		if err := p.send(msg); err != nil {
			return err
		}
	}
	return nil
}

func (r *room) join(p *player) error {
	if len(r.players) >= 2 {
		return errRoomFull
	}
	if _, ok := r.indexOf(p); ok {
		return errRoomIsSame
	}
	r.players = append(r.players, p)
	return nil
}

func (r *room) leave(p *player) error {
	for i, rp := range r.players {
		if rp == p {
			r.players = append(r.players[:i], r.players[i+1:]...)
		}
	}
	return nil
}

func (r *room) indexOf(p *player) (int, bool) {
	for idx, rp := range r.players {
		if rp == p {
			return idx, true
		}
	}
	return -1, false
}

func (r *room) charOf(pidx int) byte {
	switch pidx {
	case 0:
		return 'o'
	case 1:
		return 'x'
	default:
		return '?'
	}
}

// Marks the given cell index (uses 1-based indices, 1-9).
func (r *room) mark(idxp int, char byte) error {
	if idxp < 1 || idxp > 9 {
		return errInvalidIndex
	}
	if r.board[idxp-1] != 0 {
		return errAlreadyMarked
	}
	r.board[idxp-1] = char
	return nil
}

func (r *room) checkReady() bool {
	return len(r.players) == 2
}

func (r *room) checkTurn() *player {
	if !r.checkReady() {
		return nil
	}
	return r.players[r.turn]
}

func (r *room) checkWin() (byte, bool) {
	patterns := [][]int{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{0, 3, 6},
		{1, 4, 7},
		{2, 5, 8},
		{0, 4, 8},
		{6, 4, 2},
	}
	for _, p := range patterns {
		i, j, k := p[0], p[1], p[2]
		if r.board[i] == 0 {
			continue
		}
		if r.board[i] == r.board[j] && r.board[j] == r.board[k] {
			return r.board[i], true
		}
	}
	return 0, false
}
