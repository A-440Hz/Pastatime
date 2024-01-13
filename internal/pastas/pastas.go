package pastas

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	tts "github.com/hegedustibor/htgo-tts"
	"github.com/hegedustibor/htgo-tts/handlers"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

// TODO: why is this by str len and not word count? is there some better metric?
const strSplitLen = 195

const punctuations = ".,?!:\n"

// TODO: this pattern creates a separate audio folder when running go tests. Maybe this is desirable?
var audioFolder string

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("couldn't get cwd")
	}
	f := filepath.Join(cwd, "/audio")

	audioFolder = f
}

type Pasta struct {
	title          string
	body           []string
	body_sfw       []string
	censorStrategy string
}

// TODO: do I need error checks here?
func (p *Pasta) GetTitle() string {
	return p.title
}

func (p *Pasta) GetBody() []string {
	return p.body
}

// TODO: try using this https://github.com/TwiN/go-away/blob/master/goaway.go

// NewPasta creates a new pasta object.
// Maybe have options for "newest" or specify subreddit or others in the future
func NewPasta(...Option) (*Pasta, error) {
	opts := getDefaultOptions()

	// is this a bad design? too many opts?
	// i think it's harder to understand the code this way for sure
	// maybe im missing a struct that goes into the api package or doing certain things in the wrong places
	// but i do need to know which posts to discard for sure, and i'd like to request listings efficiently if possible
	p, err := opts.withRequestStrategy.Request(opts.withSubreddit, opts.withSortOrder, opts.withCensorStrategy)
	if err != nil {
		return nil, err
	}
	title, body := titleAndBody(p)
	return newPasta(title, body), nil
}

// this function exists so I can test without requesting a non-static live post
func newPasta(title string, body string) *Pasta {
	return &Pasta{
		title: title,
		body:  sliceTo(body),
	}
}

// sliceTo converts a string s into a slice of strings of at most size maxStrLen, preferring to split on punctuation.
func sliceTo(s string) []string {
	head, tail := 0, 0
	sliced := make([]string, 0)
	for i, cur := range s {

		// we append a new slice when it is no longer possible to have a slice under size len
		// there is an edge case of when head==tail but i don't know an elegant solution to write it
		// i ended up just using >= to account for the extra character
		// there is also an edge case of when i is len(s)

		if i-head >= strSplitLen && head != tail {
			sliced = append(sliced, strings.Clone(s[head:tail]))
			head = tail
		} else if i-head >= strSplitLen && head == tail {
			sliced = append(sliced, strings.Clone(s[head:i]))
			head = i
			tail = i
		} else if i == len(s)-1 {
			sliced = append(sliced, strings.Clone(s[head:]))
		}

		// we increment tail if the current rune is a punctuation mark
		if strings.Contains(punctuations, string(cur)) {
			tail = i + 1
		}

	}
	return sliced
}

// Speak calls the speak function on the title of a pasta and each line in the body.
func (p *Pasta) Speak(opt ...Option) error {
	err := speak(p.title, opt...)
	if err != nil {
		return err
	}
	for _, b := range p.body {
		err = speak(b, opt...)
		// if there is an err we should keep trying the next line because it might just be a newline or certain input that is wrong.
		// How do we test for errors in this case? Is there a problem with using a specific code at the end?
		if err != nil {
			return err
		}
	}
	return nil
}

func titleAndBody(p *reddit.Post, opt ...Option) (string, string) {
	// TODO: implement optional SFW filter and figure out storage location later
	return p.Title, p.Body
}

// TODO: maybe have a paths or settings file with a list of constants
// Configure a new TTS object and pass in however many strings

// speak takes in a string and uses it to create and play an mp3 file.
// opts.WithLanguageKey is used to specify a google translate language code to playback the string.
func speak(str string, opt ...Option) error {
	// TODO: I call getOptions twice when making a new pasta and calling speak.
	// this is probably a fixable design redundancy. I probably have to store something in the new struct itself.
	opts := getOpts(opt...)

	// ideally I check here if opts.withLanguageKey is valid, but idk how without compiling a giant list
	// of all supported google translate language codes, which is not very scalable

	// TODO: actually it looks like I can get a list by calling https://texttospeech.googleapis.com/v1/voices
	// https://cloud.google.com/text-to-speech/docs/reference/rest/v1/voices/list
	// https://developers.google.com/explorer-help/

	// TODO: move the audio directory within this package when I want to design replay/saved/favorite pastas
	// maybe have a pastas package with structs containing the files, some metadata, favorite bool
	// cleanup after x files played...
	// remember to use path/filepath & os.PathSeparator ('/' is fine for modern OS) for cross platform compatibility

	speech := tts.Speech{
		Folder:   audioFolder,
		Language: opts.withLanguageKey,
		Handler:  &handlers.Native{},
	}

	// TODO: implement my own speech package for miniscule speed improvements (no checking the folder if it exists every time,
	// storage of audio files with the Pasta object??)
	// I'm not sure

	// speech.Speak checks if the file already exists in the audio folder, and requests it from google
	// the request is this url:
	// fmt.Sprintf("http://translate.google.com/translate_tts?ie=UTF-8&total=1&idx=0&textlen=32&client=tw-ob&q=%s&tl=%s", url.QueryEscape(text), speech.Language)
	err := speech.Speak(str)
	if err != nil {
		// get the first 6 chars of the string for the error message
		i := 6
		if len(str) < i {
			i = len(str)
		}
		return fmt.Errorf("could not playback \"%q...\" with lang \"%q\": %q", str[:i], opts.withLanguageKey, err.Error())
	}
	return nil
}
