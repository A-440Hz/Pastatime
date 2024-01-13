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
// TODO: what about paramters for these configs? limits and suggested values?
// TODO: have a user-facing README section explaining these config values clearly.

var ctx = context.Background()

// requestWindowSize is the number of posts requested at a time from the reddit API. 25 seems to be a max value for the r/rising sort order
// this value cannot be < 1
// TODO: Write a good comment note about window size and the birthday paradox (how many calls before a repeat with pool size x)
const requestWindowSize = 25

// maxWindowShifts limits the number of API requests for subsequent listings of requestWindowSize when querying reddit post listings for a valid post.
// if requestWindowSize is 25 and maxWindowShifts is 5, a max 150 posts could be requested from the reddit API until a valid one is found.
// TODO: should I add an upper limit for this?
const maxWindowShifts = 5

// maxRandSamplesPerWindow limits the number of additional attempts made when randomly selecting a valid post
// before decrementing maxWindowShifts and calling the API to request a new listing of posts to sample from.
const maxRandSamplesPerWindow = 4

// maxPollAttempts limits the maximum attempts made to find a valid post
// This value helps separate the paramters used when random sampling and linear sampling for posts.
// Linearly it could take at most maxWindowShifts * requestWindowSize operations to find (or not find) a valid post.
// One could say maxWindowShifts * requestWindowSize is close enough to maxPollAttempts, but I think it's better to have an explicit limit.
// TODO: should I make this cap apply for random sampling too? it should be trivial
const maxPollAttempts = 100

type RequestStrategy interface {
	get(subreddit string, sortOrder string) (*reddit.Post, error)
	getSFW(subreddit string, sortOrder string) (*reddit.Post, error)
	Request(subreddit string, sortOrder string, censorStrategy string) (*reddit.Post, error)
}

// RequestNewestPost returns the most recent reddit post from a subreddit using sortOrder "new, rising, hot"
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
	for i := 0; i < maxPollAttempts; i++ {
		k := i % requestWindowSize

		// request new posts if needed
		if k == 0 && i > 0 {
			posts, resp, err = reqFunction(ctx, subreddit,
				&reddit.ListOptions{
					Limit: requestWindowSize,
					After: resp.After,
				})
			if err != nil {
				return nil, errors.Join(fmt.Errorf("RequestNewestPost.getSFW error with http code %q: %q", resp.Status, err.Error()),
					fmt.Errorf("RequestNewestPost.getSFW did not find a valid post within maxPollAttempts %d", maxPollAttempts))
			}
		}

		// return valid post immediately
		if !goaway.IsProfane(posts[k].Title) && !goaway.IsProfane(posts[k].Body) {
			return posts[i], nil
		}

	}
	return nil, fmt.Errorf("RequestNewestPost.getSFW did not find a valid post within maxPollAttempts %d", maxPollAttempts)
}

// Request specifies whether the SFW filter function is called to request posts.
func (r RequestNewestPost) Request(subreddit string, sortOrder string, censorStrategy string) (*reddit.Post, error) {
	if censorStrategy == "discard" {
		return r.getSFW(subreddit, sortOrder)
	}
	return r.get(subreddit, sortOrder)
}

// RequestRandomPost picks a random post frum subreddit using sortOrder "new, rising, or hot"
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

// getSFW increments tries maxPollAttempts times to pick a valid post, requesting requestWindowSize posts from the API maxWindowShifts+1 times,
// and sampling maxRandSamplesPerWindow each time, returning the first SFW post found or an error
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
		return nil, fmt.Errorf("RequestRandomPost.getSFW error with http code %q: %q", resp.Status, err.Error())
	}

	// fmt.Print("len posts: ", len(posts))
	rs := rand.New(rand.NewSource(time.Now().UnixNano()))
	ri := rs.Intn(requestWindowSize)
	p := posts[ri]

	for w := 0; w < maxWindowShifts; w++ {
		// fmt.Println("len posts: ", len(posts))
		// Try maxPoolAttempts times to find and return a valid SFW post.
		for i := 0; i < maxRandSamplesPerWindow; i++ {
			if !goaway.IsProfane(p.Title) && !goaway.IsProfane(p.Body) {
				return p, nil
			}
			ri = rs.Intn(requestWindowSize)
			p = posts[ri]
		}

		// shift posts window to the next slice if no valid post found yet
		posts, resp, err = reddit.DefaultClient().Subreddit.HotPosts(ctx, subreddit,
			&reddit.ListOptions{
				Limit: requestWindowSize,
				After: resp.After,
			},
		)
		if err != nil {
			return nil, errors.Join(
				fmt.Errorf("RequestRandomPost.getSFW error with http code %q: %q", resp.Status, err.Error()),
				fmt.Errorf("RequestRandomPost.getSFW did not find a valid post:\nrequestWindowSize %d, maxWindowShifts %d, and maxRetriesPerWindow %d",
					requestWindowSize, maxWindowShifts, maxRandSamplesPerWindow),
			)
		}
		p = posts[ri]
	}
	// fmt.Printf("SFW post found with maxWindowShifts: %d and maxPoolRetries: %d\n", maxWindowShifts, maxPoolRetries)

	// If above loop did not return, no valid post was found within the window polling parameters
	return nil, fmt.Errorf("RequestRandomPost.getSFW did not find a valid post:\nrequestWindowSize %d, maxWindowShifts %d, and maxRetriesPerWindow %d",
		requestWindowSize, maxWindowShifts, maxRandSamplesPerWindow)
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
		// TODO: there's no way with the current reddit package to specify "hot" and a global region
		// if i want that functionality i'd need to unwrap the wrapper a bit
		return reddit.DefaultClient().Subreddit.HotPosts
	case "rising":
		return reddit.DefaultClient().Subreddit.RisingPosts
	default:
		return reddit.DefaultClient().Subreddit.NewPosts
	}
}

// TODO: figure out streaming posts too. maybe a feature for streaming newest posts whenever is posted into a channel
