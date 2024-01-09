package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNewestPost(t *testing.T) {
	rs := RequestNewestPost{}
	p, err := rs.Get("copypasta")
	require.NoError(t, err)
	assert.NotEmpty(t, p)
	fmt.Println(p.Title, p.Body)
}

func TestGetRandomPost(t *testing.T) {
	rs := RequestRandomPost{}
	p1, err := rs.Get("copypasta")
	require.NoError(t, err)
	assert.NotEmpty(t, p1)
	fmt.Println(p1.Title, "\n", p1.Body)

	p2, err := rs.Get("copypasta")
	require.NoError(t, err)
	assert.NotEmpty(t, p2)
	fmt.Println("\n", p2.Title, "\n", p2.Body)
}

func TestGetRandomSFWPost(t *testing.T) {
	tests := []struct {
		name          string
		subreddit     string
		startPoolSize int
		errContains   string
	}{
		{
			name:          "default-startPoolSize",
			subreddit:     "copypasta",
			startPoolSize: 25,
		},
		{
			name:          "SFW-subreddit",
			subreddit:     "test",
			startPoolSize: 25,
		},
		{
			name:          "SFW-subreddit",
			subreddit:     "test",
			startPoolSize: 3000,
			errContains:   "getRandomPostSFW recieved a startPoolSize {3000}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rs := RequestRandomPostSFW{}
			p, err := rs.Get(tt.subreddit, tt.startPoolSize)
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
