package pastas

import "pastatime/internal/api"

// the Pastas object uses the go functional options pattern https://golang.cafe/blog/golang-functional-options-pattern.html
// to allow configurable parameters when creating a new Pasta, i.e. func(language=voices.English)

// I think ideally the languages in the tts/voices package are their own type, so that users cannot enter invalid options.
// However, it looks like the htgo speech package enters the language code into a query to google translate, and there are
// a lot of options in the following lists of supported language codes, so it may not be feasible/scalable/reasonable to catalogue
// every possible code as a const.

// From htgo-tts/voices/languages.go:
//     full tables
//     http://www.lingoes.net/en/translator/langcode.htm
//     https://www.science.co.il/language/Codes.php
//     https://cloud.google.com/text-to-speech/docs/voices

func getOpts(opt ...Option) options {
	opts := getDefaultOptions()
	for _, o := range opt {
		o(&opts)
	}
	return opts
}

// Option - how Options are passed as arguments
type Option func(*options)

// options = how options are represented
type options struct {
	withLanguageKey     string
	withSubreddit       string
	withRequestStrategy api.RequestStrategy
	withCensorStrategy  string
}

func getDefaultOptions() options {
	return options{
		withLanguageKey:     "en",
		withSubreddit:       "copypasta",
		withRequestStrategy: &api.RequestRandomPost{},
		withCensorStrategy:  "",
	}
}

// WithLanguageKey specifies the language code that will be sent to google translate.
func WithLanguageKey(lang string) Option {
	return func(o *options) {
		o.withLanguageKey = lang
	}
}

// WithSubreddit specifies the subreddit to request posts from.
func WithSubreddit(r string) Option {
	return func(o *options) {
		o.withSubreddit = r
	}
}

// WithRequestStrategy specifies how posts are requested.
func WithRequestStrategy(r api.RequestStrategy) Option {
	return func(o *options) {
		o.withRequestStrategy = r
	}
}

// WithCensorStrategy specifies censor strategies used when getting a post.
//		default:			none
//		"only-select-sfw":	only selects posts which do
func WithCensorStrategy(c string) Option {
	return func(o *options) {
		o.withCensorStrategy = c
	}
}
