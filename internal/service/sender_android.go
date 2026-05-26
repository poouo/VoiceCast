//go:build android

package service

import "errors"

type SendStats struct {
	Packets uint64
	Bytes   uint64
}

type Sender struct{}

func NewSender(func(SendStats)) *Sender {
	return &Sender{}
}

func (s *Sender) Start(string, int) error {
	return errors.New("Android 客户端仅支持接收")
}

func (s *Sender) Stop() {}

func (s *Sender) Running() bool {
	return false
}
