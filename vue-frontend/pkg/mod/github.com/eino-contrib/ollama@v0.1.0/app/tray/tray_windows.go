package tray

import (
	"github.com/eino-contrib/ollama/app/tray/commontray"
	"github.com/eino-contrib/ollama/app/tray/wintray"
)

func InitPlatformTray(icon, updateIcon []byte) (commontray.OllamaTray, error) {
	return wintray.InitTray(icon, updateIcon)
}
