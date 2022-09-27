package main

import (
	"time"

	"github.com/sgryczan/strava-uploader/pkg/dropbox"
	"github.com/sgryczan/strava-uploader/pkg/strava"
	"github.com/sgryczan/strava-uploader/pkg/util"
)

var (
	dropBoxAccessToken = "sl.BP2hlZPZCPdHbI8NZ4Xbnwi8871vbIwlmKft0N-bV4uEvJe6GYqx4DLsga71pl9Vd9KaodyNT7r3dFjEYIavebyf5KHILktc-o0Bc72J3vR8-gafhhmnrWferMmfvJKd6EDb8u8ZHaVg"
	stravaClientID     = "94369"
	stravaClientSecret = "aa84a8d15e8b6f5a04dc00c0a3f5fb0b7629cfd4"
)

func main() {
	// Start the auth server in case we need to do the oauth2
	// song and dance to get a token
	go strava.StartAuthServer()

	// start dropbox auth server
	go dropbox.StartAuthServer()
	for !dropbox.AuthIsGood() {
		time.Sleep(time.Second * 1)
	}

	// Run collection
	util.StartPeriodicCollection(stravaClientID, time.Hour*1)
}
