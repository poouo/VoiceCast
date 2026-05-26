//go:build cgo && !android

package audio

import (
	"context"
	"sync"

	"github.com/gen2brain/malgo"
	"github.com/poouo/VoiceCast/pkg/brand"
)

type SystemCapture struct {
	mu     sync.Mutex
	ctx    *malgo.AllocatedContext
	device *malgo.Device
}

func NewSystemCapture() *SystemCapture {
	return &SystemCapture{}
}

func (s *SystemCapture) Start(runCtx context.Context, onPCM func([]byte)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.device != nil {
		return nil
	}

	mctx, err := malgo.InitContext([]malgo.Backend{malgo.BackendWasapi}, malgo.ContextConfig{}, nil)
	if err != nil {
		mctx, err = malgo.InitContext(nil, malgo.ContextConfig{}, nil)
		if err != nil {
			return err
		}
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Loopback)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = brand.Channels
	deviceConfig.Capture.ShareMode = malgo.Shared
	deviceConfig.SampleRate = brand.SampleRate
	deviceConfig.PeriodSizeInMilliseconds = brand.FrameMillis

	callbacks := malgo.DeviceCallbacks{
		Data: func(_, input []byte, _ uint32) {
			if len(input) == 0 {
				return
			}
			frame := make([]byte, len(input))
			copy(frame, input)
			onPCM(frame)
		},
	}
	device, err := malgo.InitDevice(mctx.Context, deviceConfig, callbacks)
	if err != nil {
		_ = mctx.Uninit()
		mctx.Free()
		return err
	}
	if err := device.Start(); err != nil {
		device.Uninit()
		_ = mctx.Uninit()
		mctx.Free()
		return err
	}

	s.ctx = mctx
	s.device = device
	go func() {
		<-runCtx.Done()
		_ = s.Stop()
	}()
	return nil
}

func (s *SystemCapture) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.device != nil {
		s.device.Uninit()
		s.device = nil
	}
	if s.ctx != nil {
		_ = s.ctx.Uninit()
		s.ctx.Free()
		s.ctx = nil
	}
	return nil
}
