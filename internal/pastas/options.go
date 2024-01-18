package pastas

import (
	"pastatime/internal/api"
	"strings"
)

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
	withSortOrder       string
	withCensorStrategy  string
	withSampleRate      int
	withSampleRateScale float32
}

func getDefaultOptions() options {
	return options{
		withLanguageKey:     "en",
		withSubreddit:       "copypasta",
		withRequestStrategy: &api.RequestRandomPost{},
		withSortOrder:       "new",
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

// WithRequestStrategy specifies how posts are requested. See api package for current options.
func WithRequestStrategy(r api.RequestStrategy) Option {
	return func(o *options) {
		o.withRequestStrategy = r
	}
}

// WithCensorStrategy specifies censor strategies used when getting a post.
/*
	default:		none
	"discard":		only selects posts which do not contain any profanity
	"censor":		replaces profanity in posts with exact amount of *** characters
	"translate":	translates profanity to other languages
*/
func WithCensorStrategy(c string) Option {
	return func(o *options) {
		o.withCensorStrategy = strings.ToLower(c)
	}
}

// WithSortOrder specifies the sort order sent to the reddit API post listing request. Defaults to "new".
/*
	valid options:
	"new" (default)
	"rising"
	"hot"
*/
func WithSortOrder(s string) Option {
	return func(o *options) {
		o.withSortOrder = strings.ToLower(s)
	}
}

// WithSampleRate sets the mp3 handler sample rate. This always overides the scale option below.
func WithSampleRate(r int) Option {
	return func(o *options) {
		o.withSampleRate = r
	}
}

// WithSampleRateScale sets the mp3 handler sample rate scaling. This is always overriden by the set option above.
func WithSampleRateScale(s float32) Option {
	return func(o *options) {
		o.withSampleRateScale = s
	}
}
