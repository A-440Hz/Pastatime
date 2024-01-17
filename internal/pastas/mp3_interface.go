// The majority of the code in this file is reused from the Native.go file in hegedustibor's github.com/hegedustibor/htgo-tts
// I implement a custom interface here to modify the sample rate when playing back a mp3 file, to make it play back faster.

// reference: https://sound.stackexchange.com/questions/39568/what-sample-rates-can-an-mp3-file-have
package pastas

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
)

type MP3Interface struct {
	sampleRateScale float32
	sampleRate      int
}

// TODO: is this the best pattern for this? I'm thinking for the UI, maybe setSampleRate should be accessed by
// checking a bool first, so maybe I should make 3 struct methods to change the fields? (1 to enable overwrite sample rate, and the current 2)

// m.sampleRate always overwrites m.sampleRateScale if m.sampleRate != 0
func (m *MP3Interface) setSampleRate(r int) error {
	// TODO: what do i limit this to?
	if r < 0 {
		return fmt.Errorf("%d is not a valid sample rate", r)
	}
	m.sampleRate = r
	return nil
}

func (m *MP3Interface) setSampleRateScale(s float32) error {
	if s > 3 || s < 0 {
		return fmt.Errorf("%f is not a valid scale factor", s)
	}
	m.sampleRateScale = s
	return nil
}

func (m *MP3Interface) Play(fileName string) error {
	// Read the mp3 file into memory
	fileBytes, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	fileBytesReader := bytes.NewReader(fileBytes)

	// Decode file
	decodedMp3, err := mp3.NewDecoder(fileBytesReader)
	if err != nil {
		return err
	}

	numOfChannels := 2
	audioBitDepth := 2
	sr := decodedMp3.SampleRate()
	// either overwrite or rescale the original sample rate, if specified
	switch {
	case m.sampleRate != 0:
		sr = m.sampleRate
	case m.sampleRateScale != 0:
		sr = int(float32(sr) * m.sampleRateScale)
	}

	otoCtx, readyChan, err := oto.NewContext(sr, numOfChannels, audioBitDepth)
	if err != nil {
		return err
	}
	<-readyChan

	player := otoCtx.NewPlayer(decodedMp3)

	player.Play()

	for player.IsPlaying() {
		time.Sleep(time.Millisecond)
	}

	return player.Close()
}
