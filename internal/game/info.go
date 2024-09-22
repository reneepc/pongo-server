package game

import (
	"encoding/json"
	"errors"
)

var ErrPlayerInfoRequired = errors.New("invalid player info")

type Info struct {
	Name         string `json:"name"`
	Level        int    `json:"level"`
	ScreenWidth  int    `json:"screen_width"`
	ScreenHeight int    `json:"screen_height"`
	MaxScore     int8   `json:"max_score"`
}

func (p Info) Validate() error {
	if p.Name == "" {
		return ErrPlayerInfoRequired
	}

	return nil
}

func PlayerInfoFromMsg(msg []byte) (Info, error) {
	var info Info
	if err := json.Unmarshal(msg, &info); err != nil {
		return Info{}, err
	}

	if err := info.Validate(); err != nil {
		return Info{}, err
	}

	return info, nil
}
