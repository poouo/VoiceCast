package protocol

import "testing"

func TestMarshalUnmarshal(t *testing.T) {
	want := Packet{
		Sequence:  42,
		Timestamp: 123456,
		Format: AudioFormat{
			SampleRate:  48000,
			Channels:    2,
			FrameMillis: 20,
			Codec:       CodecPCM16,
		},
		Payload: []byte{1, 2, 3, 4},
	}
	data, err := Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Sequence != want.Sequence || got.Timestamp != want.Timestamp || got.Format != want.Format {
		t.Fatalf("packet header mismatch: got %+v want %+v", got, want)
	}
	if string(got.Payload) != string(want.Payload) {
		t.Fatalf("payload mismatch: got %v want %v", got.Payload, want.Payload)
	}
}

func TestUnmarshalRejectsBadMagic(t *testing.T) {
	data, err := Marshal(Packet{Format: AudioFormat{Codec: CodecPCM16}})
	if err != nil {
		t.Fatal(err)
	}
	data[0] = 'X'
	if _, err := Unmarshal(data); err != ErrBadMagic {
		t.Fatalf("expected ErrBadMagic, got %v", err)
	}
}
