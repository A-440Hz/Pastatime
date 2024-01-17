package api

import (
	"fmt"
	"testing"

	goaway "github.com/TwiN/go-away"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests do not have fixed results because they query the current reddit API
// One way to fix this is to use a fixed .json file to emulate query responses
// I would have to greatly modify the existing code to have that working.
// In a corporate environment I would probably have to make that change for the sake of scalability (and to better text my error cases working properly)
// but for my current needs these tests are fine. Note that they (especially TestGetRandomSFWPost) may not pass 100% of the time for this reason.

// TODO: also test the "maxWindowShift,... etc", parameters. Maybe in a refactored separate config file.
// TODO: test all the options
// reconsider how the options are passed through at New() and Speak()..

func TestGetNewestPost(t *testing.T) {
	rs := RequestNewestPost{}
	// the fact that my input parameters look so awkward indicates that I should be looking to improve my domain model here
	p, err := rs.Request("copypasta", "new", "")
	require.NoError(t, err)
	assert.NotEmpty(t, p)
	fmt.Println(p.Title, "\n", p.Body)
}

func TestRequestNewestPost(t *testing.T) {
	// There is room to do parallel testing here, but
	rs := RequestNewestPost{}
	p1, err := rs.Request("copypasta", "rising", "")
	require.NoError(t, err)
	assert.NotEmpty(t, p1.Title)
	assert.NotEmpty(t, p1.Body)
	fmt.Println("newest from /rising:\n", p1.Title, "\n", p1.Body)

	p2, err := rs.Request("copypasta", "hot", "")
	require.NoError(t, err)
	assert.NotEmpty(t, p2.Title)
	assert.NotEmpty(t, p2.Body)
	fmt.Println("newest from /hot:\n", "\n", p2.Title, "\n", p2.Body)

	p3, err := rs.Request("copypasta", "new", "")
	require.NoError(t, err)
	assert.NotEmpty(t, p3.Title)
	assert.NotEmpty(t, p3.Body)
	fmt.Println("newest from /new:\n", "\n", p3.Title, "\n", p3.Body)

	p4, err := rs.Request("copypasta", "new", "sfw")
	require.NoError(t, err)
	assert.NotEmpty(t, p4.Title)
	assert.NotEmpty(t, p4.Body)
	// can't compare with p3 because they are the same if the newest post is sfw
	assert.NotEqual(t, p2.Body, p4.Body)
	assert.False(t, goaway.IsProfane(p4.Title) || goaway.IsProfane(p4.Body))
	fmt.Println("newest sfw from /new:\n", "\n", p4.Title, "\n", p4.Body)

	// TODO: I need to put extra words in the profanity filter lmao

}

func TestRequestRandomPost_SFW(t *testing.T) {
	tests := []struct {
		name            string
		subreddit       string
		RequestStrategy RequestStrategy
		censorStrategy  string
		errContains     string
	}{
		{
			name:            "default-startPoolSize",
			subreddit:       "copypasta",
			RequestStrategy: RequestRandomPost{},
		},
		{
			name:            "SFW-subreddit",
			subreddit:       "test",
			RequestStrategy: RequestRandomPost{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := tt.RequestStrategy.Request(tt.subreddit, "new", tt.censorStrategy)
			if tt.errContains != "" {
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, p)
			fmt.Println(p.Title, "\n", p.Body)

		})
	}
}

func TestRequestNewestPost_SFW(t *testing.T) {

	rs := RequestNewestPost{}
	p, err := rs.Request("copypasta", "new", "discard")

	assert.NoError(t, err)
	assert.NotEmpty(t, p.Body)
	assert.NotEmpty(t, p.Title)
	fmt.Println("\nrequested newest sfw post:\n", p.Title, "\n", p.Body)

	p2, err := rs.Request("copypasta", "new", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, p2.Body)
	assert.NotEmpty(t, p2.Title)
	fmt.Println("\nrequested newest non-sfw post:\n", p2.Title, "\n", p2.Body)

	// TODO: there's some sort of error here maybe involving max str length
	// try search on Error: "Should not be: " in assert.NotEqual
	assert.NotEqual(t, p.Body, p2.Body)

	// request a post from a different subreddit
	p3, err := rs.Request("lifehacks", "new", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, p3.Body)
	assert.NotEmpty(t, p3.Title)
	fmt.Println("\nrequested a post from r/test:\n", p3.Title, "\n", p3.Body)

	// this testcase has a SLA of 99.9%
	assert.NotEqual(t, p2.Body, p3.Body)
}
