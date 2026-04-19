package state

import (
	"sync"
)

type PlayerState struct {
	Username string `json:"username"`
	Action   string `json:"action"`
}

var (
	// OnlinePlayers stores the current state of online players
	// Key: string (UserID or Username), Value: PlayerState
	OnlinePlayers sync.Map
)

func UpdatePlayerState(userID string, username string, action string) {
	OnlinePlayers.Store(userID, PlayerState{
		Username: username,
		Action:   action,
	})
}

func RemovePlayer(userID string) {
	OnlinePlayers.Delete(userID)
}

func GetOnlinePlayers() map[string]PlayerState {
	players := make(map[string]PlayerState)
	OnlinePlayers.Range(func(key, value interface{}) bool {
		players[key.(string)] = value.(PlayerState)
		return true // continue iteration
	})
	return players
}
