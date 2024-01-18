package main

import (
	"fmt"
	"pastatime/internal/api"
	"pastatime/internal/pastas"
	"sync"
)

// Here are some basic examples

func getMostRecentPost(speak bool) {

	np, err := pastas.NewPasta([]pastas.Option{
		pastas.WithSortOrder("new"),
		pastas.WithRequestStrategy(api.RequestNewestPost{})}...)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Newest r/copypasta post:\n", np.GetTitle())
	for _, b := range np.GetBody() {
		fmt.Print(b)
		// fmt.Println(fmt.Sprintf("%q|%d", b, len(b)))
	}
	if !speak {
		return
	}
	err = np.Speak([]pastas.Option{
		pastas.WithLanguageKey("en-UK"),
		pastas.WithSampleRate(31000)}...)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func getRandomPostSFW(speak bool) {
	np, err := pastas.NewPasta([]pastas.Option{
		pastas.WithSortOrder("hot"),
		pastas.WithCensorStrategy("discard"),
		pastas.WithRequestStrategy(api.RequestRandomPost{}),
	}...)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Random SFW r/copypasta post:\n", np.GetTitle(), "\n", np.GetBody())
	if speak {
		np.Speak()
	}
}

func getBatchSFW(n int, speak bool) {
	var wg sync.WaitGroup
	for i := n; i > 0; i-- {
		wg.Add(1)
		go func() {
			defer wg.Done()
			getRandomPostSFW(speak)
			fmt.Print("\n")
		}()
	}
	wg.Wait()
}

func main() {

	// getMostRecentPost(true)
	// getRandomPostSFW()
	getBatchSFW(10, false)

}
