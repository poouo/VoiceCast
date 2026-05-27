package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2/dialog"

	"github.com/poouo/VoiceCast/internal/config"
	"github.com/poouo/VoiceCast/internal/validate"
	"github.com/poouo/VoiceCast/pkg/brand"
)

func (ui *AppUI) startListening() {
	if ui.mode == ModeWindows && ui.sender != nil && ui.sender.Running() {
		ui.stopSending()
	}
	port := brand.DefaultPort
	if err := validate.ListenPort(port); err != nil {
		ui.showError(err)
		return
	}
	if err := ui.receiver.Start(port); err != nil {
		ui.showError(fmt.Errorf("启动监听失败：%w", err))
		return
	}
	ui.cfg.ListenPort = port
	if ui.mode == ModeAndroid {
		ui.cfg.AutoListen = false
	}
	_ = config.Save(ui.cfg)
	ui.startListen.Disable()
	ui.stopListen.Enable()
	ui.statusLabel.SetText(fmt.Sprintf("监听中：%d", port))
}

func (ui *AppUI) stopListening() {
	if ui.receiver != nil {
		ui.receiver.Stop()
	}
	if ui.startListen != nil {
		ui.startListen.Enable()
	}
	if ui.stopListen != nil {
		ui.stopListen.Disable()
	}
	if ui.statusLabel != nil {
		ui.statusLabel.SetText("已停止监听")
	}
	if ui.recvStatLabel != nil {
		ui.recvStatLabel.SetText("未接收")
	}
}

func (ui *AppUI) startSending() {
	if ui.sender == nil {
		return
	}
	if ui.receiver != nil && ui.receiver.Running() {
		ui.stopListening()
	}
	ip := strings.TrimSpace(ui.ipEntry.Text)
	port := brand.DefaultPort
	if err := validate.Target(ip, port); err != nil {
		ui.showError(err)
		return
	}
	if err := ui.sender.Start(ip, port); err != nil {
		ui.showError(fmt.Errorf("启动推送失败：%w", err))
		return
	}
	ui.cfg.LastTargetIP = ip
	ui.cfg.LastTargetPort = port
	_ = config.Save(ui.cfg)
	ui.startSend.Disable()
	ui.stopSend.Enable()
	ui.ipEntry.Disable()
	ui.statusLabel.SetText(fmt.Sprintf("推送中：%s", ip))
}

func (ui *AppUI) stopSending() {
	if ui.sender != nil {
		ui.sender.Stop()
	}
	if ui.startSend != nil {
		ui.startSend.Enable()
	}
	if ui.stopSend != nil {
		ui.stopSend.Disable()
	}
	if ui.ipEntry != nil {
		ui.ipEntry.Enable()
	}
	if ui.statusLabel != nil {
		ui.statusLabel.SetText("已停止推送")
	}
	if ui.sendStatLabel != nil {
		ui.sendStatLabel.SetText("未推送")
	}
}

func (ui *AppUI) showError(err error) {
	if err == nil {
		return
	}
	ui.statusLabel.SetText(err.Error())
	dialog.ShowError(err, ui.window)
}
