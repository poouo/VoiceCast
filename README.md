# VoiceCast

VoiceCast 是一个中文界面的局域网音频推流工具，用 UDP 把 Windows 设备的声音发送到同一局域网里的电脑或安卓手机。

## 目标

- Windows 客户端使用 Go 编写，负责采集、发送、接收和播放局域网音频。
- Windows UI 使用 Go 原生桌面方案，不使用前端语言。
- Android 端仅作为接收端，负责接收 UDP 音频并播放。
- 软件显示当前内网地址。
- 保存上次使用过的目标地址到配置文件，启动时自动读取。
- 输入 IP 地址时进行有效性判定，避免错误地址导致推流失败。
- 用户无需输入端口，客户端统一使用默认 UDP 端口 `39000`。

## 核心能力

- Windows -> Windows/Android：采集本机系统声音或麦克风音频，经 UDP 推送到局域网目标设备。
- Windows <- Windows：监听局域网 UDP 音频流并实时播放。
- Android <- Windows：监听局域网 UDP 音频流并实时播放。

## 项目名称和图标

- 项目名：`VoiceCast`
- Go module：`github.com/poouo/VoiceCast`
- 应用 ID：`com.poouo.voicecast`
- LOGO 源文件：[assets/logo/voicecast.svg](assets/logo/voicecast.svg)
- 客户端文案：仅中文

## 架构文档

完整软件架构图、模块划分、数据流和配置设计见：

- [docs/architecture.md](docs/architecture.md)

## 当前实现

- Windows 客户端入口：`cmd/voicecast-windows`
- Android 客户端入口：`cmd/voicecast-android`
- 配置文件：系统用户配置目录下的 `VoiceCast/config.json`
- 默认端口：`39000`
- Android APK 默认构建 `arm64-v8a` 单架构，安装包比四架构通用包更小。
- Android 打开后不会自动监听，需要手动点击“开始监听”。
- UDP 协议：`pkg/protocol`
- 地址校验：`internal/validate`
- 音频播放：`oto`
- Windows 声音采集：`malgo/miniaudio`，优先使用 WASAPI loopback

## 建议技术栈

| 端 | 技术 |
| --- | --- |
| Windows 主程序 | Go |
| Windows UI | Fyne 或 Gio |
| Windows 音频采集/播放 | WASAPI 封装或 PortAudio |
| UDP 网络层 | Go `net.UDPConn` |
| 编解码 | PCM 起步，后续可扩展 Opus |
| Android 接收端 | Go mobile 绑定或 Android 原生壳 + Go 音频网络核心 |
| 配置文件 | JSON/TOML/YAML |

## 构建

Windows：

```powershell
.\scripts\build-windows.ps1
```

Android 需要先安装 Android NDK，并设置 `ANDROID_NDK_HOME`：

```powershell
.\scripts\package-android.ps1
```

当前环境已验证 Windows 客户端可以编译并启动。Android 打包需要 NDK 环境；没有 NDK 时会停在 `android/log.h` 这类系统头文件缺失错误。

如果本机没有 Android 环境，可以直接使用 GitHub Actions 构建：

1. 推送代码到 `github.com/poouo/VoiceCast`。
2. 在 GitHub Actions 手动运行 `Build and Release`，或推送 `v*` 标签。
3. 推送标签时会自动创建 GitHub Release，并上传：
   - `VoiceCast-Windows-amd64.exe`
   - `VoiceCast-Android.apk`

Windows 构建使用内部链接器，避免部分 CGO 外部链接产物在 Windows 下提示“此应用无法在你的电脑上运行”。

## Git 隐私建议

为了避免提交记录暴露正式姓名和邮箱，仓库内建议使用：

```powershell
git config user.name "poouo"
git config user.email "poouo@users.noreply.github.com"
```

## 推荐仓库结构

```text
.
├── cmd/
│   ├── voicecast-windows/
│   └── voicecast-android/
├── internal/
│   ├── audio/
│   ├── config/
│   ├── network/
│   ├── ui/
│   └── validate/
├── pkg/
│   └── protocol/
├── docs/
│   └── architecture.md
└── README.md
```
