package validate

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

var (
	ErrInvalidIP   = errors.New("IP 地址格式不正确")
	ErrInvalidPort = errors.New("端口必须在 1 到 65535 之间")
	ErrNotLAN      = errors.New("请输入局域网私有地址")
	ErrBadTarget   = errors.New("该地址不能作为推送目标")
)

func Target(ipText string, port int) error {
	ip := net.ParseIP(strings.TrimSpace(ipText))
	if ip == nil {
		return ErrInvalidIP
	}
	ip = ip.To4()
	if ip == nil {
		return fmt.Errorf("%w：当前版本仅支持 IPv4", ErrInvalidIP)
	}
	if ip.IsLoopback() || ip.IsUnspecified() || ip.IsMulticast() {
		return ErrBadTarget
	}
	if !ip.IsPrivate() {
		return ErrNotLAN
	}
	if port < 1 || port > 65535 {
		return ErrInvalidPort
	}
	return nil
}

func ListenPort(port int) error {
	if port < 1 || port > 65535 {
		return ErrInvalidPort
	}
	return nil
}
