package api

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

// ---\
// Env secrets scratch code for logging in a verified account via "installed app" type in the go-reddit package:
// TODO: figure out a better storage for this stuff. Maybe reference Boundary.
// TODO: actually implement this secrets stuff if i hit a rate limit while stress testing request rates.
// const envSecret = "PT_SECRET"
// const envID = "PT_ID"

// var clientSecret = os.Getenv(envSecret)
// var clientID = os.Getenv(envID)

var ctx = context.Background()

// do api stuff manually or use go-reddit?
// it'll be good to take a look first at how it works regardless
// TODO: why am I using go-reddit instead of graw? is graw better cause it has streaming?
// need to find out the best way to proceed.
// ---/

// TODO: make these constants modifiable numbers in a config setting somewhere

// randomPoolSize is the starting pool size used when querying x posts from a queue
// TODO: Write a good comment note about pool size and the birthday paradox (how many calls before a repeat with pool size x)
const randomPoolSize = 25

// maxPoolSize limits the max pool size cap used when querying x reddit posts into a queue
const maxPoolSize = 100

// maxPoolAttempts limits the number of attempts made before giving up on querying a post when the randomPoolSize is at maxPoolSize
const maxPoolAttempts = 10

type RequestStrategy interface {
	Get(subreddit string) (*reddit.Post, error)
}

// RequestNewestPost gets the newest reddit post from subreddit
type RequestNewestPost struct{}

func (r *RequestNewestPost) Get(subreddit string) (*reddit.Post, error) {

	posts, resp, err := reddit.DefaultClient().Subreddit.NewPosts(ctx, subreddit,
		&reddit.ListOptions{
			Limit: 1,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("GetNewestPost error with http code %q: %q", resp.Status, err.Error())
	}

	return posts[0], nil
}

// RequestRandomPost gets a random post from the top randomPoolSize posts on the reddit "trending" filter
type RequestRandomPost struct{}

func (r *RequestRandomPost) Get(subreddit string) (*reddit.Post, error) {

	return getRandomPost(subreddit, randomPoolSize)
}

func getRandomPost(subreddit string, startPoolSize int) (*reddit.Post, error) {

	posts, resp, err := reddit.DefaultClient().Subreddit.RisingPosts(ctx, subreddit,
		&reddit.ListOptions{
			Limit: startPoolSize,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("getRandomPost error with http code %q: %q", resp.Status, err.Error())
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(startPoolSize)
	return posts[r], nil
}

// RequestRandomPostSFW increments startPoolSize to maxPoolSize and tries maxPoolAttempts times to draw a random
type RequestRandomPostSFW struct{}

func (r *RequestRandomPostSFW) Get(subreddit string, startPoolSize int) (*reddit.Post, error) {
	return getRandomPostSFW(subreddit, startPoolSize)

}

func getRandomPostSFW(subreddit string, startPoolSize int) (*reddit.Post, error) {
	// TODO: the max return length from reddit is a list of 25 posts. modify constants and behavior to scroll query until a valid post is found
	// I initially wanted to query reddit for maxPoolSize posts once and reuse the list as we iterate through, but the API only seems to return
	// 25 posts at a time.
	// I want to fix this query to use a sliding window until it hits a max length.
	if startPoolSize > maxPoolSize {
		return nil, fmt.Errorf("getRandomPostSFW recieved a startPoolSize {%d} greater than maxPoolSize {%d}", startPoolSize, maxPoolSize)
	}

	posts, resp, err := reddit.DefaultClient().Subreddit.RisingPosts(ctx, subreddit,
		&reddit.ListOptions{
			Limit: maxPoolSize,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("getRandomSFWPost error with http code %q: %q", resp.Status, err.Error())
	}

	// NOTE: rand.New will call multiple times if getRandomPost is invoked while maxPoolSize is too large (> 1000?)
	r := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(startPoolSize)
	p := posts[r]
	fmt.Print(len(posts))

	for goaway.IsProfane(p.Title) || goaway.IsProfane(p.Body) {
		// if startPoolSize is at maxPoolSize, try maxPoolAttempts additional times before giving up to find a SFW post.
		if startPoolSize == maxPoolSize {
			for i := 0; i < maxPoolAttempts; i++ {
				r = rand.New(rand.NewSource(time.Now().UnixNano())).Intn(startPoolSize)
				if !goaway.IsProfane(posts[r].Title) && !goaway.IsProfane(posts[r].Body) {
					return posts[r], nil
				}
			}
			return nil, fmt.Errorf("getRandomSFWPost failed to find an SFW post with maxPoolSize: %d and maxPoolAttempts %d", maxPoolSize, maxPoolAttempts)
		}
		startPoolSize = less(maxPoolSize, startPoolSize*2)
		r = rand.New(rand.NewSource(time.Now().UnixNano())).Intn(startPoolSize)
		p = posts[r]

	}

	return p, nil

}

func less(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// TODO: figure out streaming posts too. maybe a feature for streaming newest posts whenever is posted into a channel
