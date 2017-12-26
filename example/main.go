package main

import (
	"log"

	"github.com/davecgh/go-spew/spew"
	tw "github.com/tensei/twitch-clip"
)

func main() {
	twitch, err := tw.NewClient(
		"clientid",
		"clientsecret",
		"accesstoken",
		"refreshtoken",
	)
	if err != nil {
		log.Fatalf("hmm %v", err)
	}

	a, err := twitch.RefreshAuthToken()
	if err != nil {
		log.Fatalf("hmm 1 %v", err)
	}
	spew.Dump(a)

	// Destiny
	clipid, err := twitch.CreateClip("18074328")
	if err != nil {
		log.Fatalf("hmm 2 %v", err)
	}
	log.Println(clipid)

	clip, err := twitch.GetClip(clipid)
	if err != nil {
		log.Fatalf("hmm 3 %v", err)
	}
	log.Panicln(clip.Data[0].URL)
}
