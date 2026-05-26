//go:build windows

package main

import (
	"log"

	"github.com/poouo/VoiceCast/internal/ui"
)

func main() {
	appUI, err := ui.New(ui.ModeWindows)
	if err != nil {
		log.Printf("读取配置失败：%v", err)
	}
	appUI.Run()
}
