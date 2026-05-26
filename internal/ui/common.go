package ui

import (
	"fmt"
	"image/color"
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
	if mode == ModeAndroid {
		a.Settings().SetTheme(theme.DarkTheme())
	}
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
	ui.lanLabel = widget.NewLabel("")
	ui.refreshLANAddresses()
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
		container.NewBorder(nil, nil, nil, widget.NewButtonWithIcon("刷新", theme.ViewRefreshIcon(), ui.refreshLANAddresses), ui.lanLabel),
		ui.portLabel,
		container.NewHBox(ui.startListen, ui.stopListen),
		ui.recvStatLabel,
	))

	if ui.mode == ModeWindows {
		content := container.NewVBox(header, localCard)
		content.Add(ui.senderCard())
		content.Add(widget.NewCard("状态", "", container.NewVBox(ui.statusLabel, widget.NewLabel("配置文件："+ui.cfgPath))))
		ui.window.SetContent(container.NewPadded(content))
		ui.window.Resize(fyne.NewSize(680, 520))
	} else {
		ui.window.SetContent(ui.androidPlayerContent())
		ui.window.Resize(fyne.NewSize(420, 760))
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
	if ui.mode == ModeWindows && ui.cfg.AutoListen {
		go func() {
			time.Sleep(300 * time.Millisecond)
			fyne.Do(ui.startListening)
		}()
	}
	go func() {
		time.Sleep(1200 * time.Millisecond)
		fyne.Do(ui.refreshLANAddresses)
	}()
}

func (ui *AppUI) refreshLANAddresses() {
	text := strings.Join(netutil.LANIPv4s(), " / ")
	if text == "" {
		text = "未检测到局域网 IPv4 地址，请确认已连接 Wi-Fi"
	}
	ui.lanLabel.SetText(text)
}

func (ui *AppUI) androidPlayerContent() fyne.CanvasObject {
	topSafe := spacer(56)
	title := canvas.NewText(brand.AppName, color.NRGBA{R: 245, G: 248, B: 252, A: 255})
	title.TextSize = 28
	title.TextStyle.Bold = true
	subtitle := canvas.NewText("局域网音频接收", color.NRGBA{R: 156, G: 170, B: 188, A: 255})
	subtitle.TextSize = 15

	disc := canvas.NewCircle(color.NRGBA{R: 32, G: 160, B: 155, A: 255})
	logo := canvas.NewImageFromResource(IconResource())
	logo.SetMinSize(fyne.NewSize(100, 100))
	cover := container.NewCenter(container.NewGridWrap(fyne.NewSize(190, 190),
		container.NewStack(disc, container.NewCenter(logo)),
	))

	ui.startListen.Importance = widget.HighImportance
	buttons := container.NewGridWithColumns(2, ui.startListen, ui.stopListen)
	address := container.NewBorder(nil, nil, nil, widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), ui.refreshLANAddresses), ui.lanLabel)

	panel := widget.NewCard("", "", container.NewVBox(
		widget.NewLabel("本机地址"),
		address,
		ui.portLabel,
		widget.NewSeparator(),
		ui.recvStatLabel,
		ui.statusLabel,
	))

	body := container.NewVBox(
		topSafe,
		container.NewCenter(title),
		container.NewCenter(subtitle),
		spacer(24),
		cover,
		spacer(22),
		buttons,
		spacer(14),
		panel,
	)
	bg := canvas.NewRectangle(color.NRGBA{R: 14, G: 18, B: 24, A: 255})
	return container.NewStack(bg, container.NewPadded(body))
}

func spacer(height float32) fyne.CanvasObject {
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(1, height))
	return rect
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
