package twitchClip

import (
	"context"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestTwitch_RefreshAuthToken(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tw, _ := NewClient("", "", "", "")
	tests := []struct {
		name    string
		t       *Twitch
		args    args
		wantErr bool
	}{
		{t: tw, args: args{context.Background()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.t.RefreshAuthToken(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Twitch.RefreshAuthToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				t.Logf("access_token: %s, refresh_token: %s, scope: %v", got.AccessToken, got.RefreshToken, got.Scope)
				return
			}
		})
	}
}

func TestTwitch_CreateClip(t *testing.T) {
	type args struct {
		ctx           context.Context
		broadcasterid string
	}
	tw, _ := NewClient("", "", "", "")
	tests := []struct {
		name    string
		t       *Twitch
		args    args
		wantErr bool
	}{
		{t: tw, args: args{context.Background(), "18074328"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.t.CreateClip(tt.args.ctx, tt.args.broadcasterid)
			if (err != nil) != tt.wantErr {
				t.Errorf("Twitch.CreateClip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == "" {
				t.Errorf("Twitch.CreateClip() = %v, error %v", got, err)
			}
		})
	}
}

func TestTwitch_GetClip(t *testing.T) {
	type args struct {
		ctx    context.Context
		clipid string
	}
	tw, _ := NewClient("", "", "", "")
	tests := []struct {
		name    string
		t       *Twitch
		args    args
		wantErr bool
	}{
		{t: tw, args: args{context.Background(), "ChillyGrossOtterFailFish"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.t.GetClip(tt.args.ctx, tt.args.clipid)
			if (err != nil) != tt.wantErr {
				t.Errorf("Twitch.GetClip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				spew.Dump(got)
			}
		})
	}
}
