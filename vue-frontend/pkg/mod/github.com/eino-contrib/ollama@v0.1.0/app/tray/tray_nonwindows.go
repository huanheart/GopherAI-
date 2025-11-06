//go:build !windows

package tray

import (
	"errors"

	"github.com/eino-contrib/ollama/app/tray/commontray"
)

func InitPlatformTray(icon, updateIcon []byte) (commontray.OllamaTray, error) {
	return nil, errors.New("not implemented")
}
