package ui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/poouo/VoiceCast/internal/config"
	"github.com/poouo/VoiceCast/internal/netutil"
	"github.com/poouo/VoiceCast/internal/service"
	"github.com/poouo/VoiceCast/pkg/brand"
)

type Mode int

const (
	ModeWindows Mode = iota
	ModeAndroid
)

type AppUI struct {
	mode    Mode
	app     fyne.App
	window  fyne.Window
	cfg     config.Config
	cfgPath string

	receiver *service.Receiver
	sender   *service.Sender

	ipEntry       *widget.Entry
	lanLabel      *widget.Label
	portLabel     *widget.Label
	statusLabel   *widget.Label
	recvStatLabel *widget.Label
	sendStatLabel *widget.Label

	startListen *widget.Button
	stopListen  *widget.Button
	startSend   *widget.Button
	stopSend    *widget.Button
	quitting    bool
}

func New(mode Mode) (*AppUI, error) {
	cfg, path, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}
	a := app.NewWithID(brand.AppID)
	a.SetIcon(IconResource())
	w := a.NewWindow(brand.AppName)
	w.SetIcon(IconResource())
	ui := &AppUI{
		mode:     mode,
		app:      a,
		window:   w,
		cfg:      cfg,
		cfgPath:  path,
		receiver: service.NewReceiver(nil),
	}
	ui.receiver = service.NewReceiver(func(stats service.ReceiveStats) {
		fyne.Do(func() {
			ui.recvStatLabel.SetText(fmt.Sprintf("已接收：%d 包 / %.1f KB，最后序号：%d", stats.Packets, float64(stats.Bytes)/1024, stats.LastSequence))
			ui.statusLabel.SetText("正在接收音频")
		})
	})
	if mode == ModeWindows {
		ui.sender = service.NewSender(func(stats service.SendStats) {
			fyne.Do(func() {
				ui.sendStatLabel.SetText(fmt.Sprintf("已推送：%d 包 / %.1f KB", stats.Packets, float64(stats.Bytes)/1024))
				ui.statusLabel.SetText("正在推送本机声音")
			})
		})
	}
	ui.build()
	return ui, err
}

func (ui *AppUI) Run() {
	ui.window.ShowAndRun()
}

func (ui *AppUI) build() {
	ui.lanLabel = widget.NewLabel(strings.Join(netutil.LANIPv4s(), " / "))
	if ui.lanLabel.Text == "" {
		ui.lanLabel.SetText("未检测到局域网 IPv4 地址")
	}
	ui.portLabel = widget.NewLabel(fmt.Sprintf("默认端口：%d", brand.DefaultPort))
	ui.statusLabel = widget.NewLabel("就绪")
	ui.recvStatLabel = widget.NewLabel("已接收：0 包 / 0 KB")
	ui.sendStatLabel = widget.NewLabel("已推送：0 包 / 0 KB")

	ui.startListen = widget.NewButtonWithIcon("开始监听", theme.MediaPlayIcon(), ui.startListening)
	ui.startListen.Importance = widget.HighImportance
	ui.stopListen = widget.NewButtonWithIcon("停止监听", theme.MediaStopIcon(), ui.stopListening)
	ui.stopListen.Disable()

	title := canvas.NewText(brand.AppName, theme.Color(theme.ColorNameForeground))
	title.TextSize = 24
	title.TextStyle.Bold = true
	logo := canvas.NewImageFromResource(IconResource())
	logo.SetMinSize(fyne.NewSize(52, 52))
	header := container.NewBorder(nil, nil, logo, nil, container.NewVBox(title, widget.NewLabel("局域网语音接收与推送工具")))

	localCard := widget.NewCard("本机地址", "", container.NewVBox(
		ui.lanLabel,
		ui.portLabel,
		container.NewHBox(ui.startListen, ui.stopListen),
		ui.recvStatLabel,
	))

	content := container.NewVBox(header, localCard)
	if ui.mode == ModeWindows {
		content.Add(ui.senderCard())
	} else {
		content.Add(widget.NewCard("接收模式", "", widget.NewLabel("Android 客户端仅支持接收局域网音频。")))
	}
	content.Add(widget.NewCard("状态", "", container.NewVBox(ui.statusLabel, widget.NewLabel("配置文件："+ui.cfgPath))))

	ui.window.SetContent(container.NewPadded(content))
	if ui.mode == ModeWindows {
		ui.window.Resize(fyne.NewSize(680, 520))
	} else {
		ui.window.Resize(fyne.NewSize(420, 640))
	}
	ui.setupTray()
	ui.window.SetCloseIntercept(func() {
		if ui.quitting {
			ui.window.Close()
			return
		}
		ui.window.Hide()
		ui.statusLabel.SetText("已最小化到系统托盘")
	})
	if ui.mode == ModeAndroid || ui.cfg.AutoListen {
		go func() {
			time.Sleep(300 * time.Millisecond)
			fyne.Do(ui.startListening)
		}()
	}
}

func (ui *AppUI) setupTray() {
	deskApp, ok := ui.app.(desktop.App)
	if !ok {
		return
	}
	menu := fyne.NewMenu(brand.AppName,
		fyne.NewMenuItem("打开主窗口", func() {
			ui.showWindow()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("退出 VoiceCast", func() {
			ui.quit()
		}),
	)
	deskApp.SetSystemTrayMenu(menu)
	deskApp.SetSystemTrayIcon(IconResource())
	deskApp.SetSystemTrayWindow(ui.window)
}

func (ui *AppUI) showWindow() {
	ui.window.Show()
	ui.window.RequestFocus()
}

func (ui *AppUI) quit() {
	ui.quitting = true
	ui.stopListening()
	if ui.sender != nil {
		ui.sender.Stop()
	}
	ui.app.Quit()
}

func (ui *AppUI) senderCard() fyne.CanvasObject {
	ui.ipEntry = widget.NewEntry()
	ui.ipEntry.SetText(ui.cfg.LastTargetIP)
	ui.ipEntry.SetPlaceHolder("例如：192.168.1.25")
	ui.startSend = widget.NewButtonWithIcon("开始推送", theme.UploadIcon(), ui.startSending)
	ui.startSend.Importance = widget.HighImportance
	ui.stopSend = widget.NewButtonWithIcon("停止推送", theme.MediaStopIcon(), ui.stopSending)
	ui.stopSend.Disable()
	return widget.NewCard("推送到设备", "", container.NewVBox(
		widget.NewForm(widget.NewFormItem("目标 IP", ui.ipEntry)),
		widget.NewLabel(fmt.Sprintf("推送端口：%d", brand.DefaultPort)),
		container.NewHBox(ui.startSend, ui.stopSend),
		ui.sendStatLabel,
		widget.NewLabel("推送使用 Windows 系统声音回环采集，接收端保持监听即可。"),
	))
}
