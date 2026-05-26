//go:build !cgo || android

package audio

import (
	"context"
	"errors"
)

type SystemCapture struct{}

func NewSystemCapture() *SystemCapture {
	return &SystemCapture{}
}

func (s *SystemCapture) Start(context.Context, func([]byte)) error {
	return errors.New("当前平台不支持音频采集")
}

func (s *SystemCapture) Stop() error {
	return nil
}
