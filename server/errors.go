package tiki

import "errors"

var (
	errRoomFull      = errors.New("room is already full")
	errRoomNotFound  = errors.New("room not found")
	errRoomNotReady  = errors.New("room is not ready")
	errRoomIsSame    = errors.New("player is already in this room")
	errInvalidTurn   = errors.New("not player turn yet")
	errInvalidPlayer = errors.New("player is not supposed to be here")
	errInvalidIndex  = errors.New("invalid index")
	errInvalidParams = errors.New("invalid parameters")
	errAlreadyMarked = errors.New("cell already marked")
	errPlayerNoRoom  = errors.New("player is not in a room")
)
