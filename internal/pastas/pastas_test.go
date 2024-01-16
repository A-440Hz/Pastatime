package pastas

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sample_body_general_a = "You do realise OP that the rich need money to find a way of transcending in afearal light beings, that can leave this plain. Greece and the gods of old, are the key to it all. I got confirmed to me today by a Freemason that we live in a matrix.\nI also told him that people above himself( he was high in set but not at the top by a long measure, he knew things lower members wouldn’t but didn’t know the could shots were depopulation events, he said he told his Covid shots but I asked him, if he’d he’d then before I mentioned how they was a tool of thinning the heard.\nAll humans live in a dream, within a dream, a simulation that goes many layers down. I’ve seen a tree pixelate in front of me, a missing girl poster and the missing girl was sat in front of it. I’ve had 10 NDE, that I shouldn’t have survived at all.\nDo I believe I have a purpose? Yeh I do but don’t know that it is at all."
var sample_body_random_chars = "d](qMCSm.F(;yXZ_{}-T{:/m2n%c89b!d@.qPrA+}YLR&4;9463%vz(ZwQU:_Zm46jL)U=MCrXg*6}$$Zj,5(XJ3zQ51@,h4x*6Ax1TMnpW4KHTNE&2@N25yba61d2_t=uXn%:1+AnhK_-Hy{6/*%WBN.J&9k$+8JWj/)jYzpV9%!V&9tdyAd2P%K[vK64c*izp"
var sample_body_emojis = "howie thought 🤔 brass 🎺 was 🕜 the height ⬆️ of style 👔 now ⏱️ he's got 🎁 something 🌳➰ of a 🅰️ steampunk 🚂 smile 😁 and it's 🤡 alright 👍👍 it's 🤡 alright 😊 it's 🤡 alright 👍\nsally got 🎁 a dagger 🗡️ hung ➰ from her septum 🫀 omalley cut ✂️ his ears 🌽 off, but wishes ⭐ that he kept 'em and it's 🤡 all right ▶️ (it's alright to look cool 😎) (you 🫵 do what you 🫵 do what you 👦 do) rooney got 🎁 his skull 💀 exposed, doggone it 🤡! soon ⌚ he's gonna get 🎁 scrimshaw ☠️ carved 🔪 on it and it's all right (it's alright to be cool 🆒) it's alright, it's alright 👍 do what you want with you 🫵 be nonchalant 😐 with screws 🪛 stuck 😝 through your 👶 eyelids 👁️_👁️ new wave of pirates 🦜 modify 👩‍🔧 (modify whatever 🙄) modify 🔧 (modify ⚒️ and sever 🪚) modify 🔨 may nothing 😶 get rejected ❌ may nothing 😑 get infected 🤧🦠"

var sample_body_general_b = "I think contracting aids from a girl would actually be kinda hot and romantic in a fucked up way. Like yeah you are literally dying for the chance to clap her cheeks and that adds a little something. And if it's a girl you really love then it's kinda romantic bc you are deciding that if she's gonna die from it then you'll go with her. I think I would come a lot harder knowing my fate was sealed the moment I slid inside her, and knowing it's too late to back out. I think that if people stopped being such cowards and just tanked the crotch rot, we'd build up and immunity to it and not have to worry about it at all anymore. But nobody wants to make the hard sacrifices to get us to that point. This generation is so sexually closed minded it's unreal. Like they're too scared to even eat ass bc \"muh strep throat\". Cowards the lot of you"

// the google translate api seems to not return a file when the query is too long
// let's try to separate long queries into chunks of <= 195 characters at the nearest space

func TestNewPasta(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		options []Option
	}{
		// TODO: implement options testcases

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPasta(tt.options...)
			if tt.wantErr {
				// TODO: implement err message checking
				assert.Error(t, err)
				return
			}
			require.Nil(t, err)
			assert.NotNil(t, p)
		})
	}
}

func TestSpeak(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name:    "general",
			body:    sample_body_general_a,
			wantErr: true,
			// test should fail because input is too long;
			// the unofficial google translate url does not return a file.
			// long input string will be split in normal use cases
			// TODO: write a more specific fail case with error msg checking and comment why it should fail
		},
		{
			name:    "random-chars",
			body:    sample_body_random_chars,
			wantErr: false,
		},
		{
			name:    "emojis",
			body:    sample_body_emojis,
			wantErr: true,
			// test should fail because input is too long.
		},
		{
			name:    "empty",
			body:    "",
			wantErr: true,
		},
		{
			name:    "one-char",
			body:    "a",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: this timeout go function doesn't work in VS Code.
			// it gets overridden somewhere and my workaround is to hard set a value in settings.json
			timeout := time.After(3 * time.Minute)
			done := make(chan bool)
			go func() {
				err := speak(tt.body)
				if tt.wantErr {
					assert.Error(t, err)
					done <- true
					return
				}
				require.NoError(t, err)
				done <- true
				return
			}()

			select {
			case <-timeout:
				t.Fatal("Test didn't finish in time")
			case <-done:
			}
		})
	}

}

func TestSlice(t *testing.T) {
	// TODO: it seems like these body texts are getting cut off a bit before the end.
	// Figure out if its a string size limit thing or my code dropping lines somehow
	// is the max char length 690?
	tests := []struct {
		name  string
		title string
		body  string
	}{
		{
			name:  "test-slice",
			title: "We all live in a matrix, a Freemason told me so.",
			body:  sample_body_general_a,
		},
		{
			name:  "test-slice-emojis",
			title: "here is an example with emojis",
			body:  sample_body_emojis,
		},
		{
			name:  "test-slice-emojis",
			title: "here is an example with random chars",
			body:  sample_body_random_chars,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPasta(tt.title, tt.body)
			require.NotEmpty(t, p.GetTitle())
			for _, ln := range p.GetBody() {
				fmt.Println(ln)
				assert.LessOrEqual(t, len(ln), strSplitLen)
				if len(ln) < strSplitLen {
					assert.Contains(t, punctuations, rune(ln[len(ln)-1]))
				}
			}
			assert.Equal(t, tt.title, p.GetTitle())
		})
	}
}

// TODO: it's bad form to have this test here because it tests elements of the api package, so I should remove this test at some point
// probably when I get main working.
func TestGetRandomPost(t *testing.T) {
	p, err := NewPasta()
	require.NoError(t, err)
	fmt.Println(p.GetTitle())
	for _, ln := range p.GetBody() {
		fmt.Println(ln)
		require.LessOrEqual(t, len(ln), strSplitLen)
	}
	require.NoError(t, p.Speak())
}
