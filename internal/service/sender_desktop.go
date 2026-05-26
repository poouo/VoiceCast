//go:build !android

package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/poouo/VoiceCast/internal/audio"
	"github.com/poouo/VoiceCast/pkg/brand"
	"github.com/poouo/VoiceCast/pkg/protocol"
)

type SendStats struct {
	Packets uint64
	Bytes   uint64
}

type Sender struct {
	mu      sync.Mutex
	cancel  context.CancelFunc
	conn    *net.UDPConn
	capture audio.CaptureSource
	target  *net.UDPAddr
	seq     atomic.Uint32
	onStat  func(SendStats)
	stats   SendStats
}

func NewSender(onStat func(SendStats)) *Sender {
	return &Sender{onStat: onStat}
}

func (s *Sender) Start(ip string, port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		return errors.New("推送已经启动")
	}
	target, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp4", nil, target)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	capture := audio.NewSystemCapture()
	s.cancel = cancel
	s.conn = conn
	s.capture = capture
	s.target = target
	s.seq.Store(0)
	s.stats = SendStats{}
	err = capture.Start(ctx, func(pcm []byte) {
		_ = s.sendFrame(pcm)
	})
	if err != nil {
		cancel()
		_ = conn.Close()
		s.cancel = nil
		s.conn = nil
		s.capture = nil
		s.target = nil
		return err
	}
	return nil
}

func (s *Sender) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	conn := s.conn
	capture := s.capture
	s.cancel = nil
	s.conn = nil
	s.capture = nil
	s.target = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if capture != nil {
		_ = capture.Stop()
	}
	if conn != nil {
		_ = conn.Close()
	}
}

func (s *Sender) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cancel != nil
}

func (s *Sender) sendFrame(pcm []byte) error {
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()
	if conn == nil {
		return nil
	}
	pkt := protocol.Packet{
		Sequence:  s.seq.Add(1),
		Timestamp: protocol.TimestampNow(),
		Format: protocol.AudioFormat{
			SampleRate:  brand.SampleRate,
			Channels:    brand.Channels,
			FrameMillis: brand.FrameMillis,
			Codec:       protocol.CodecPCM16,
		},
		Payload: pcm,
	}
	data, err := protocol.Marshal(pkt)
	if err != nil {
		return err
	}
	if _, err := conn.Write(data); err != nil {
		return err
	}
	s.mu.Lock()
	s.stats.Packets++
	s.stats.Bytes += uint64(len(pcm))
	stats := s.stats
	s.mu.Unlock()
	if s.onStat != nil {
		s.onStat(stats)
	}
	return nil
}
