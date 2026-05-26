package audio

import (
	"fmt"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
)

var (
	playbackMu       sync.Mutex
	playbackCtx      *oto.Context
	playbackRate     int
	playbackChannels int
)

type Player struct {
	queue  *PCMQueue
	player *oto.Player
}

func NewPlayer(sampleRate, channels int) (*Player, error) {
	ctx, err := sharedPlaybackContext(sampleRate, channels)
	if err != nil {
		return nil, err
	}
	queue := NewPCMQueue(sampleRate * channels * 2)
	p := ctx.NewPlayer(queue)
	p.SetBufferSize(sampleRate * channels * 2 / 10)
	p.Play()
	return &Player{queue: queue, player: p}, nil
}

func sharedPlaybackContext(sampleRate, channels int) (*oto.Context, error) {
	playbackMu.Lock()
	defer playbackMu.Unlock()
	if playbackCtx != nil {
		if playbackRate != sampleRate || playbackChannels != channels {
			return nil, fmt.Errorf("音频播放参数不一致：已有 %dHz/%d 声道", playbackRate, playbackChannels)
		}
		return playbackCtx, nil
	}
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
	playbackCtx = ctx
	playbackRate = sampleRate
	playbackChannels = channels
	return playbackCtx, nil
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
		p.player.Reset()
	}
	if p.queue != nil {
		return p.queue.Close()
	}
	return nil
}
