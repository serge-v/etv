package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

func getActivationCode() (string, error) {
	u := codeURL +
		"?client_id=" + clientID +
		"&client_secret=" + clientSecret +
		"&scope=" + strings.Join(scope, "%20")

	var err error
	var resp activationResp
	if err := os.Remove(cacheDir + "activation.json"); err != nil {
		log.Println(err)
	}
	if err = fetch(u, "activation.json", &resp); err != nil {
		return "", err
	}
	if *debug {
		log.Printf("Activation code: %s\n", resp.UserCode)
		log.Println("Open http://etvnet.com/device/ and enter activation code.")
		log.Println("Then run: etv -auth")
	}
	return resp.UserCode, nil
}

func authorize() error {
	var aresp activationResp
	if err := os.Remove(cacheDir + "auth.json"); err != nil {
		log.Println(err)
	}
	if err := fetch("", "activation.json", &aresp); err != nil {
		return err
	}

	u := tokenURL +
		"?client_id=" + clientID +
		"&client_secret=" + clientSecret +
		"&scope=" + strings.Join(scope, "%20") +
		"&grant_type=http%3A%2F%2Foauth.net%2Fgrant_type%2Fdevice%2F1.0" +
		"&code=" + aresp.DeviceCode
	var resp authorizationResp
	if err := fetch(u, "", &resp); err != nil {
		return err
	}
	fmt.Printf("+%v", resp)
	var newconf = cfg
	newconf.AccessToken = resp.AccessToken
	newconf.RefreshToken = resp.RefreshToken
	newconf.ExpiresIn = resp.ExpiresIn
	f, err := os.Create("etvrc.tmp")
	if err != nil {
		return err
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(&newconf); err != nil {
		return err
	}
	if err = os.Rename("etvrc", "etvrc.old"); err != nil {
		return err
	}
	if err = os.Rename("etvrc.tmp", "etvrc"); err != nil {
		return err
	}
	log.Println("etvrc updated")
	return nil
}

func refreshToken() error {
	log.Println("refresh token")
	u := tokenURL +
		"?client_id=" + clientID +
		"&client_secret=" + clientSecret +
		"&scope=" + strings.Join(scope, "%20") +
		"&grant_type=refresh_token" +
		"&refresh_token=" + cfg.RefreshToken
	var resp authorizationResp
	if err := fetch(u, "", &resp); err != nil {
		return err
	}
	log.Printf("+%v", resp)
	var newconf = cfg

	newconf.AccessToken = resp.AccessToken
	newconf.RefreshToken = resp.RefreshToken
	newconf.ExpiresIn = resp.ExpiresIn
	cfg.AccessToken = resp.AccessToken
	cfg.RefreshToken = resp.RefreshToken
	cfg.ExpiresIn = resp.ExpiresIn

	f, err := os.Create("etvrc.tmp")
	if err != nil {
		return err
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(&newconf); err != nil {
		return err
	}
	if err = os.Rename("etvrc", "etvrc.old"); err != nil {
		return err
	}
	if err = os.Rename("etvrc.tmp", "etvrc"); err != nil {
		return err
	}
	log.Println("etvrc updated")
	return nil
}

type config struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}
