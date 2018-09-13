package main

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"
)

func (a *api) getActivation() (activationResp, error) {
	u := codeURL +
		"?client_id=" + clientID +
		"&client_secret=" + clientSecret +
		"&scope=" + strings.Join(scope, "%20")

	var err error
	var resp activationResp
	buf, err := getURL(u)
	if err != nil {
		log.Println(err)
		return resp, err
	}
	err = checkToken(buf)
	if err != nil {
		return resp, err
	}
	if err := json.Unmarshal(buf, &resp); err != nil {
		log.Println(err, string(buf))
		return resp, err
	}
	log.Println("user code:", resp.UserCode, "device code:", resp.DeviceCode)
	return resp, nil
}

func (a *api) authorize() error {
	if a.deviceCode == "" {
		return errors.New("invalid authorize call")
	}
	u := tokenURL +
		"?client_id=" + clientID +
		"&client_secret=" + clientSecret +
		"&scope=" + strings.Join(scope, "%20") +
		"&grant_type=http%3A%2F%2Foauth.net%2Fgrant_type%2Fdevice%2F1.0" +
		"&code=" + a.deviceCode
	if err := a.fetch(u, "", &a.auth); err != nil {
		return err
	}
	a.deviceCode = ""
	log.Printf("authorization: +%v", a.auth)
	return nil
}

func (a *api) refreshToken() error {
	log.Println("refresh token")
	u := tokenURL +
		"?client_id=" + clientID +
		"&client_secret=" + clientSecret +
		"&scope=" + strings.Join(scope, "%20") +
		"&grant_type=refresh_token" +
		"&refresh_token=" + a.auth.RefreshToken
	buf, err := getURL(u)
	if err != nil {
		log.Println(err)
		return err
	}
	err = checkToken(buf)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(buf, &a.auth); err != nil {
		log.Println(err, string(buf))
		return err
	}
	a.auth.Expires = time.Now().Add(time.Second * time.Duration(a.auth.ExpiresIn))

	log.Printf("%+v", a.auth)
	return nil
}
