// The majority of the code in this file is reused from the Native.go file in hegedustibor's github.com/hegedustibor/htgo-tts
// I implement a custom interface here to modify the sample rate when playing back a mp3 file, to make it play back faster.

// reference: https://sound.stackexchange.com/questions/39568/what-sample-rates-can-an-mp3-file-have
package pastas

import (
	"bytes"
	"os"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
)

const defaultScaleFactor = 1.20

type MP3Interface struct {
	sampleRateScale float32
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

	if m.sampleRateScale == 0 {
		sr = int(float32(sr) * defaultScaleFactor)
	} else {
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
