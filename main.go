package main

import (
	"fmt"
	"pastatime/internal/api"
)

// TODO: get rid of this page and have all relevant tests in the _test files.

func getRandomPost() {
	rs := api.RequestRandomPost{}
	p, err := rs.Request("copypasta", "new", "discard")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(p.Title)
	fmt.Println(p.Body)
}

func main() {

	getRandomPost()

}
