package game

import (
	"encoding/json"
	"errors"
)

var ErrPlayerInfoRequired = errors.New("invalid player info")

type GameInfo struct {
	Name             string `json:"player_name"`
	Level            int    `json:"level"`
	ScreenWidth      int    `json:"screen_width"`
	ScreenHeight     int    `json:"screen_height"`
	FieldBorderWidth int    `json:"field_border_width"`
	MaxScore         int8   `json:"max_score"`
}

func (p GameInfo) Validate() error {
	if p.Name == "" {
		return ErrPlayerInfoRequired
	}

	return nil
}

func PlayerInfoFromMsg(msg []byte) (GameInfo, error) {
	var info GameInfo
	if err := json.Unmarshal(msg, &info); err != nil {
		return GameInfo{}, err
	}

	if err := info.Validate(); err != nil {
		return GameInfo{}, err
	}

	return info, nil
}
