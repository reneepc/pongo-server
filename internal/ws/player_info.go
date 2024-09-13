package ws

import (
	"encoding/json"
	"errors"
)

var ErrPlayerInfoRequired = errors.New("invalid player info")

type PlayerInfo struct {
	Name string `json:"name"`
}

func (p PlayerInfo) Validate() error {
	if p.Name == "" {
		return ErrPlayerInfoRequired
	}

	return nil
}

func PlayerInfoFromMsg(msg []byte) (PlayerInfo, error) {
	var playerInfo PlayerInfo
	if err := json.Unmarshal(msg, &playerInfo); err != nil {
		return PlayerInfo{}, err
	}

	if err := playerInfo.Validate(); err != nil {
		return PlayerInfo{}, err
	}

	return playerInfo, nil
}
