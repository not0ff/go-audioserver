package audio

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
	"github.com/not0ff/go-audioserver/internal/message"
)

const ConstSampleRate beep.SampleRate = 44100

type AudioPlayer struct {
	controls map[int]*beep.Ctrl
	quits    map[int]chan bool
}

func NewAudioPlayer() *AudioPlayer {
	// Initialize speaker with constant sample rate
	speaker.Init(ConstSampleRate, ConstSampleRate.N(time.Second/10))

	return &AudioPlayer{controls: map[int]*beep.Ctrl{}, quits: map[int]chan bool{}}
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
func (a *AudioPlayer) decodeAudio(r io.Reader, format string) (beep.StreamSeekCloser, beep.Format, error) {
	switch format {
	case "mp3":
		s, f, err := mp3.Decode(io.NopCloser(r))
		if err != nil {
			return nil, beep.Format{}, err
		}
		return s, f, nil

	case "wav":
		s, f, err := wav.Decode(r)
		if err != nil {
			return nil, beep.Format{}, err
		}
		return s, f, nil

	default:
		return nil, beep.Format{}, fmt.Errorf("unsupported format \"%s\"", format)
	}
}

func (a *AudioPlayer) Pause(id int) {
	speaker.Lock()
	if ctrl, ok := a.controls[id]; !ok {
		log.Printf("Error pausing, streamer doesnt exist")
	} else {
		ctrl.Paused = true
	}
	speaker.Unlock()
}

func (a *AudioPlayer) Resume(id int) {
	speaker.Lock()
	if ctrl, ok := a.controls[id]; !ok {
		log.Printf("Error resuming, streamer doesnt exist")
	} else {
		ctrl.Paused = false
	}
	speaker.Unlock()
}

func (a *AudioPlayer) Quit(id int) {
	if quit, ok := a.quits[id]; !ok {
		log.Printf("Error quiting, streamer doesnt exist")
	} else {
		quit <- true
		a.cleanup(id)
	}
}

func (a *AudioPlayer) cleanup(id int) {
	close(a.quits[id])
	delete(a.quits, id)
	delete(a.controls, id)
}

// Plays audio from message payload (blocking)
func (a *AudioPlayer) Play(p *message.PlayPayload) {
	r, err := a.readAudio(p)
	if err != nil {
		log.Printf("Error reading audio: %s", err)
		return
	}

	streamer, format, err := a.decodeAudio(r, p.Format)
	if err != nil {
		log.Printf("Error decoding audio: %s", err)
		return
	}
	defer streamer.Close()

	// Set optional infinite looping
	l := 1
	if p.Loop {
		l = -1
	}

	// Create callback so function returns after ending playback
	quit := make(chan bool)
	sq := beep.Seq(beep.Loop(l, streamer), beep.Callback(func() {
		quit <- true
		a.cleanup(p.Id)
	}))

	// Add controls for pausing/resuming
	ctrl := &beep.Ctrl{Streamer: sq, Paused: false}

	// Apply volume options
	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   float64(p.Volume),
		Silent:   false,
	}

	// Bind playback control and termination to provided ID
	a.controls[p.Id] = ctrl
	a.quits[p.Id] = quit

	// Resample if audio doesn't match constant sample rate
	var st beep.Streamer
	if format.SampleRate != ConstSampleRate {
		st = beep.Resample(6, format.SampleRate, ConstSampleRate, volume)
	} else {
		st = volume
	}

	speaker.Play(st)
	<-quit
}
