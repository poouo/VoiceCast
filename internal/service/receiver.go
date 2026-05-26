package service

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/poouo/VoiceCast/internal/audio"
	"github.com/poouo/VoiceCast/pkg/brand"
	"github.com/poouo/VoiceCast/pkg/protocol"
)

type ReceiveStats struct {
	Packets      uint64
	Bytes        uint64
	LastSequence uint32
	LastSeen     time.Time
}

type Receiver struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	conn   *net.UDPConn
	player *audio.Player
	stats  ReceiveStats
	onStat func(ReceiveStats)
}

func NewReceiver(onStat func(ReceiveStats)) *Receiver {
	return &Receiver{onStat: onStat}
}

func (r *Receiver) Start(port int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		return errors.New("监听已经启动")
	}
	addr := &net.UDPAddr{IP: net.IPv4zero, Port: port}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return err
	}
	player, err := audio.NewPlayer(brand.SampleRate, brand.Channels)
	if err != nil {
		_ = conn.Close()
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	r.conn = conn
	r.player = player
	r.stats = ReceiveStats{}
	go r.loop(ctx, conn, player)
	return nil
}

func (r *Receiver) Stop() {
	r.mu.Lock()
	cancel := r.cancel
	conn := r.conn
	player := r.player
	r.cancel = nil
	r.conn = nil
	r.player = nil
	r.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if conn != nil {
		_ = conn.Close()
	}
	if player != nil {
		_ = player.Close()
	}
}

func (r *Receiver) Running() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.cancel != nil
}

func (r *Receiver) loop(ctx context.Context, conn *net.UDPConn, player *audio.Player) {
	buf := make([]byte, protocol.HeaderSize+65535)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				continue
			}
			continue
		}
		pkt, err := protocol.Unmarshal(buf[:n])
		if err != nil || pkt.Format.Codec != protocol.CodecPCM16 {
			continue
		}
		player.WritePCM(pkt.Payload)
		r.addStats(pkt)
	}
}

func (r *Receiver) addStats(pkt protocol.Packet) {
	r.mu.Lock()
	r.stats.Packets++
	r.stats.Bytes += uint64(len(pkt.Payload))
	r.stats.LastSequence = pkt.Sequence
	r.stats.LastSeen = time.Now()
	stats := r.stats
	r.mu.Unlock()
	if r.onStat != nil {
		r.onStat(stats)
	}
}
