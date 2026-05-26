//go:build android

package main

import (
	"log"

	"github.com/poouo/VoiceCast/internal/ui"
)

func main() {
	appUI, err := ui.New(ui.ModeAndroid)
	if err != nil {
		log.Printf("读取配置失败：%v", err)
	}
	appUI.Run()
}
