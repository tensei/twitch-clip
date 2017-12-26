package main

import (
	"context"
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
	ctx := context.Background()

	a, err := twitch.RefreshAuthToken(ctx)
	if err != nil {
		log.Fatalf("hmm 1 %v", err)
	}
	spew.Dump(a)

	// Destiny
	clipid, err := twitch.CreateClip(ctx, "44445592")
	if err != nil {
		log.Fatalf("hmm 2 %v", err)
	}
	log.Println(clipid)

	clip, err := twitch.GetClip(ctx, clipid)
	if err != nil {
		log.Fatalf("hmm 3 %v", err)
	}
	log.Println(clip.Data[0].URL)
}
