package main

import (
	"testing"

	"github.com/pecoin/go-fullnode/common"
)

func TestFacebook(t *testing.T) {
	for _, tt := range []struct {
		url  string
		want common.Address
	}{
		{
			"https://www.facebook.com/fooz.gazonk/posts/2837228539847129",
			common.HexToAddress("0xDeadDeaDDeaDbEefbEeFbEEfBeeFBeefBeeFbEEF"),
		},
	} {
		_, _, gotAddress, err := authFacebook(tt.url)
		if err != nil {
			t.Fatal(err)
		}
		if gotAddress != tt.want {
			t.Fatalf("address wrong, have %v want %v", gotAddress, tt.want)
		}
	}
}
