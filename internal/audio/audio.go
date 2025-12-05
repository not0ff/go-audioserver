package audio

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
	"github.com/not0ff/go-audioserver/pkg/message"
)

const ConstSampleRate beep.SampleRate = 44100

var (
	ErrStreamerNotExist = errors.New("streamer does not exist")
	ErrStreamerExist    = errors.New("streamer already exists")
)

type AudioPlayer struct {
	controls map[int]*beep.Ctrl
}

func NewAudioPlayer() *AudioPlayer {
	// Initialize speaker with constant sample rate
	speaker.Init(ConstSampleRate, ConstSampleRate.N(time.Second/10))
	return &AudioPlayer{controls: map[int]*beep.Ctrl{}}
}

// Return reader for audio data specified in payload
func (a *AudioPlayer) readAudio(p *message.PlayPayload) (io.Reader, error) {
	if len(p.Path) > 0 {
		r, err := os.Open(p.Path)
		if err != nil {
			return nil, err
		}
		return r, nil
	} else if len(p.Data) > 0 {
		// Decompresses encoded audio data
		b := bytes.NewReader(p.Data)
		return gzip.NewReader(b)
	} else {
		return nil, errors.New("no audio data provided")
	}
}

// Decodes audio using a proper coding format
func (a *AudioPlayer) decodeAudio(r io.Reader, format string) (beep.StreamSeekCloser, *beep.Format, error) {
	switch format {
	case "mp3":
		s, f, err := mp3.Decode(io.NopCloser(r))
		if err != nil {
			return nil, nil, err
		}
		return s, &f, nil

	case "wav":
		s, f, err := wav.Decode(r)
		if err != nil {
			return nil, nil, err
		}
		return s, &f, nil

	default:
		return nil, nil, fmt.Errorf("unsupported format \"%s\"", format)
	}
}

func (a *AudioPlayer) Pause(id int) error {
	if ctrl, ok := a.controls[id]; !ok {
		return ErrStreamerNotExist
	} else {
		speaker.Lock()
		ctrl.Paused = true
		speaker.Unlock()
	}
	return nil
}

func (a *AudioPlayer) Resume(id int) error {
	if ctrl, ok := a.controls[id]; !ok {
		return ErrStreamerNotExist
	} else {
		speaker.Lock()
		ctrl.Paused = false
		speaker.Unlock()
	}
	return nil
}

func (a *AudioPlayer) Quit(id int) error {
	if ctrl, ok := a.controls[id]; !ok {
		return ErrStreamerNotExist
	} else {
		speaker.Lock()
		ctrl.Streamer = nil
		a.cleanup(id)
		speaker.Unlock()
	}
	return nil
}

func (a *AudioPlayer) cleanup(id int) {
	delete(a.controls, id)
}

// Plays audio from message payload
func (a *AudioPlayer) Play(p *message.PlayPayload) error {
	if _, exists := a.controls[p.Id]; exists {
		return ErrStreamerExist
	}

	r, err := a.readAudio(p)
	if err != nil {
		return err
	}

	streamer, format, err := a.decodeAudio(r, p.Format)
	if err != nil {
		return err
	}
	defer streamer.Close()

	// Set optional infinite looping
	l := 1
	if p.Loop {
		l = -1
	}

	// Add controls for pausing/resuming
	ctrl := &beep.Ctrl{Streamer: beep.Loop(l, streamer), Paused: false}
	a.controls[p.Id] = ctrl

	// Apply volume options
	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   float64(p.Volume),
		Silent:   false,
	}

	// Resample if audio doesn't match constant sample rate
	var st beep.Streamer
	if format.SampleRate != ConstSampleRate {
		st = beep.Resample(6, format.SampleRate, ConstSampleRate, volume)
	} else {
		st = volume
	}

	speaker.Play(beep.Seq(st, beep.Callback(func() {
		a.cleanup(p.Id)
	})))
	return nil
}
