package audio

import (
	"time"

	"github.com/ebitengine/oto/v3"
)

type Player struct {
	queue  *PCMQueue
	player *oto.Player
}

func NewPlayer(sampleRate, channels int) (*Player, error) {
	queue := NewPCMQueue(sampleRate * channels * 2)
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: channels,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   80 * time.Millisecond,
	})
	if err != nil {
		return nil, err
	}
	<-ready
	p := ctx.NewPlayer(queue)
	p.SetBufferSize(sampleRate * channels * 2 / 10)
	p.Play()
	return &Player{queue: queue, player: p}, nil
}

func (p *Player) WritePCM(data []byte) {
	if p == nil || p.queue == nil || len(data) == 0 {
		return
	}
	_, _ = p.queue.Write(data)
	if p.player != nil && !p.player.IsPlaying() {
		p.player.Play()
	}
}

func (p *Player) Close() error {
	if p == nil {
		return nil
	}
	if p.player != nil {
		p.player.Pause()
		_ = p.player.Close()
	}
	if p.queue != nil {
		return p.queue.Close()
	}
	return nil
}
