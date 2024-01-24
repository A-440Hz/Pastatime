package main

import (
	"fmt"
	"net/http"
	"pastatime/internal/api"
	"pastatime/internal/pastas"
	"pastatime/views"
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

func getRandomPostSFW(speak bool) (*pastas.Pasta, error) {
	np, err := pastas.NewPasta([]pastas.Option{
		pastas.WithSortOrder("hot"),
		pastas.WithCensorStrategy("discard"),
		pastas.WithRequestStrategy(api.RequestRandomPost{}),
	}...)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	fmt.Println("Random SFW r/copypasta post:\n", np.GetTitle(), "\n", np.GetBody())
	fmt.Print("\n")
	if speak {
		np.Speak()
	}
	return np, nil
}

func getBatchSFW(n int, speak bool) ([]*pastas.Pasta, error) {
	var wg sync.WaitGroup
	pastaCh := make(chan *pastas.Pasta)
	errCh := make(chan error)
	var pastas []*pastas.Pasta

	// asynchronously request n posts
	for i := n; i > 0; i-- {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p, err := getRandomPostSFW(speak)
			errCh <- err
			pastaCh <- p

		}()
	}

	go func() {
		wg.Wait()
		close(errCh)
		close(pastaCh)
	}()

	// convert results into lists (and return error if needed)
	for i := 0; i < n; i++ {
		if err := <-errCh; err != nil {
			return nil, err
		}
		pastas = append(pastas, <-pastaCh)
	}

	return pastas, nil
}

func main() {

	// getMostRecentPost(true)
	// getRandomPostSFW()
	// p, e := getBatchSFW(5, false)
	// if e != nil {
	// 	fmt.Print(e.Error())
	// }
	// fmt.Print(p != nil)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", views.HomeFunc)
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		fmt.Println("main.go: ", err)
	}

}
