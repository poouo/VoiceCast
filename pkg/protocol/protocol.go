package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

const (
	Magic       = "VCST"
	Version     = 1
	CodecPCM16  = 0
	HeaderSize  = 24
	maxPayload  = 60 * 1024
	headerMagic = 0
)

var (
	ErrShortPacket = errors.New("packet is too short")
	ErrBadMagic    = errors.New("packet magic mismatch")
	ErrBadVersion  = errors.New("packet version mismatch")
	ErrPayloadSize = errors.New("packet payload size mismatch")
)

type AudioFormat struct {
	SampleRate  uint16
	Channels    uint8
	FrameMillis uint8
	Codec       uint8
}

type Packet struct {
	Sequence  uint32
	Timestamp uint64
	Format    AudioFormat
	Payload   []byte
}

func Marshal(p Packet) ([]byte, error) {
	if len(p.Payload) > maxPayload {
		return nil, fmt.Errorf("payload too large: %d", len(p.Payload))
	}
	out := make([]byte, HeaderSize+len(p.Payload))
	copy(out[headerMagic:4], Magic)
	out[4] = Version
	out[5] = p.Format.Codec
	binary.BigEndian.PutUint32(out[6:10], p.Sequence)
	binary.BigEndian.PutUint64(out[10:18], p.Timestamp)
	binary.BigEndian.PutUint16(out[18:20], p.Format.SampleRate)
	out[20] = p.Format.Channels
	out[21] = p.Format.FrameMillis
	binary.BigEndian.PutUint16(out[22:24], uint16(len(p.Payload)))
	copy(out[HeaderSize:], p.Payload)
	return out, nil
}

func Unmarshal(data []byte) (Packet, error) {
	if len(data) < HeaderSize {
		return Packet{}, ErrShortPacket
	}
	if string(data[headerMagic:4]) != Magic {
		return Packet{}, ErrBadMagic
	}
	if data[4] != Version {
		return Packet{}, ErrBadVersion
	}
	payloadSize := int(binary.BigEndian.Uint16(data[22:24]))
	if payloadSize != len(data)-HeaderSize {
		return Packet{}, ErrPayloadSize
	}
	payload := make([]byte, payloadSize)
	copy(payload, data[HeaderSize:])
	return Packet{
		Sequence:  binary.BigEndian.Uint32(data[6:10]),
		Timestamp: binary.BigEndian.Uint64(data[10:18]),
		Format: AudioFormat{
			SampleRate:  binary.BigEndian.Uint16(data[18:20]),
			Channels:    data[20],
			FrameMillis: data[21],
			Codec:       data[5],
		},
		Payload: payload,
	}, nil
}

func TimestampNow() uint64 {
	return uint64(time.Now().UnixMilli())
}
