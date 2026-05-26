package audio

import "context"

type CaptureSource interface {
	Start(context.Context, func([]byte)) error
	Stop() error
}

func BytesPerFrame(channels int) int {
	return channels * 2
}
