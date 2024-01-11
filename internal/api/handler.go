package api

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

// TODO: maybe changing these file names to requestStrategy.go might be more clear. Keep in mind as I add more features

// ---\
// Env secrets scratch code for logging in a verified account via "installed app" type in the go-reddit package:
// TODO: figure out a better storage for this stuff. Maybe reference Boundary.
// TODO: actually implement this secrets stuff if i hit a rate limit while stress testing request rates.
// const envSecret = "PT_SECRET"
// const envID = "PT_ID"

// var clientSecret = os.Getenv(envSecret)
// var clientID = os.Getenv(envID)

// do api stuff manually or use go-reddit?
// it'll be good to take a look first at how it works regardless
// TODO: why am I using go-reddit instead of graw? is graw better cause it has streaming?
// need to find out the best way to proceed.
// ---/

// TODO: make these constants modifiable numbers in a config setting somewhere

var ctx = context.Background()

// requestWindowSize is the number of posts requested at a time from the reddit API. 25 seems to be a max value for the r/rising sort order
// TODO: Write a good comment note about window size and the birthday paradox (how many calls before a repeat with pool size x)
const requestWindowSize = 25

// maxWindowShifts limits the number of API requests for subsequent listings of requestWindowSize when querying reddit post listings for a valid post.
// if requestWindowSize is 25 and maxWindowShifts is 5, a max 150 posts could be requested from the reddit API until a valid one is found.
const maxWindowShifts = 5

// maxRetriesPerWindow limits the number of additional attempts made to select a valid post
// before moving on to the next pool of requestWindowSize posts
const maxRetriesPerWindow = 4

type RequestStrategy interface {
	get(subreddit string, sortOrder string) (*reddit.Post, error)
	getSFW(subreddit string, sortOrder string) (*reddit.Post, error)
	Request(subreddit string, sortOrder string, censorStrategy string) (*reddit.Post, error)
}

// RequestNewestPost prioritizes the newest reddit post from a subreddit and sort order
// not to be confused with "sort by /new"
type RequestNewestPost struct{}

func (r RequestNewestPost) get(subreddit string, sortOrder string) (*reddit.Post, error) {
	reqFunction := requestFromSortOrder(sortOrder)
	posts, resp, err := reqFunction(ctx, subreddit,
		&reddit.ListOptions{
			Limit: requestWindowSize,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("RequestNewestPost.get error with http code %q: %q", resp.Status, err.Error())
	}
	return posts[0], nil
}

func (r RequestNewestPost) getSFW(subreddit string, sortOrder string) (*reddit.Post, error) {
	reqFunction := requestFromSortOrder(sortOrder)
	posts, resp, err := reqFunction(ctx, subreddit,
		&reddit.ListOptions{
			Limit: requestWindowSize,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("RequestNewestPost.getSFW error with http code %q: %q", resp.Status, err.Error())
	}
	// TODO: iterate if NSFW
	return posts[0], nil
}

// Request specifies whether the SFW filter function is called to request posts.
func (r RequestNewestPost) Request(subreddit string, sortOrder string, censorStrategy string) (*reddit.Post, error) {
	if censorStrategy == "discard" {
		return r.getSFW(subreddit, sortOrder)
	}
	return r.get(subreddit, sortOrder)
}

// RequestRandomPost gets a random post from the top requestWindowSize posts on the reddit "trending" filter
type RequestRandomPost struct{}

func (r RequestRandomPost) get(subreddit string, sortOrder string) (*reddit.Post, error) {
	// TODO: make sort order (/rising, /hot, /top) specifiable (via options struct? factory?)
	// there is room to incorporate the SFW filtering logic into a big "get" function with many parameters. might be cleaner in the long run, maybe inconsequential
	// as is, i I have a bad feeling about building up too much repetitive code as i add more getXxxYyyy functions
	// /controversial and /top use ListPostOptions{Time : "hour, day, week, month, year, all"} rather than ListOptions.

	requestFunc := requestFromSortOrder(sortOrder)
	posts, resp, err := requestFunc(ctx, subreddit,
		&reddit.ListOptions{
			Limit: requestWindowSize,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("RequestRandomPost.get error with http code %q: %q", resp.Status, err.Error())
	}

	i := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(requestWindowSize)
	return posts[i], nil
}

// getSFW increments startPoolSize to maxWindowShifts and tries maxPoolAttempts times to draw a valid post
func (r RequestRandomPost) getSFW(subreddit string, sortOrder string) (*reddit.Post, error) {
	// The Reddit API only seems to return listings of 25 posts max at a time when querying post listings by the "/rising" sort order.
	// Ordering by "/hot" seems to return listings of 101 or 100.
	// I initially wanted to request a large number of posts and iterate through them until I find a SFW post
	// Now I will use a sliding window of 25 and try maxPoolAttempts times in each window to find a SFW post,
	// so that all sort orders can hopefully be supported.

	reqFunction := requestFromSortOrder(sortOrder)
	posts, resp, err := reqFunction(ctx, subreddit,
		&reddit.ListOptions{
			Limit: requestWindowSize,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("getRandomSFWPost error with http code %q: %q", resp.Status, err.Error())
	}

	// fmt.Print("len posts: ", len(posts))
	rs := rand.New(rand.NewSource(time.Now().UnixNano()))
	ri := rs.Intn(requestWindowSize)
	p := posts[ri]

	for i := 0; (goaway.IsProfane(p.Title) || goaway.IsProfane(p.Body)) && i < maxWindowShifts; i++ {
		// fmt.Println("len posts: ", len(posts))

		// if the content contains profanity, try maxPoolAttempts more times to find a SFW post.
		for i := 0; i < maxRetriesPerWindow; i++ {
			ri = rs.Intn(requestWindowSize)
			if !goaway.IsProfane(posts[ri].Title) && !goaway.IsProfane(posts[ri].Body) {
				return posts[ri], nil
			}
		}

		// advance p to the next maxPoolSize slice of posts
		posts, resp, err = reddit.DefaultClient().Subreddit.HotPosts(ctx, subreddit,
			&reddit.ListOptions{
				Limit: requestWindowSize,
				After: resp.After,
			},
		)
		if err != nil {
			return nil, errors.Join(
				fmt.Errorf("getRandomSFWPost error with http code %q: %q", resp.Status, err.Error()),
				fmt.Errorf("getRandomSFWPost failed to find an SFW post with maxWindowShifts: %d and maxPoolRetries: %d", maxWindowShifts, maxRetriesPerWindow),
			)
		}
		p = posts[ri]
	}
	// fmt.Printf("SFW post found with maxWindowShifts: %d and maxPoolRetries: %d\n", maxWindowShifts, maxPoolRetries)

	// Final check for no valid posts found but p was still assigned.
	// TODO: This error statement shouldn't be so
	if goaway.IsProfane(p.Title) || goaway.IsProfane(p.Body) {
		return nil, fmt.Errorf("RequestRandomPost.getSFW was not able to find a valid post with \n, requestWindowSize %d, maxWindowShifts %d, and maxRetriesPerWindow %d",
			requestWindowSize, maxWindowShifts, maxRetriesPerWindow,
		)
	}

	return p, nil
}

func (r RequestRandomPost) Request(subreddit string, sortOrder string, censorStrategy string) (*reddit.Post, error) {
	if censorStrategy == "discard" {
		return r.getSFW(subreddit, sortOrder)
	}
	return r.get(subreddit, sortOrder)
}

// requestFromSortOrder takes in a sortOrder and returns the cooresponding API request function,
// defaulting to sort by "new".
func requestFromSortOrder(sortOrder string) func(context.Context, string, *reddit.ListOptions) ([]*reddit.Post, *reddit.Response, error) {
	switch sortOrder {
	case "hot":
		return reddit.DefaultClient().Subreddit.HotPosts
	case "rising":
		return reddit.DefaultClient().Subreddit.RisingPosts
	default:
		return reddit.DefaultClient().Subreddit.NewPosts
	}
}

// TODO: figure out streaming posts too. maybe a feature for streaming newest posts whenever is posted into a channel
