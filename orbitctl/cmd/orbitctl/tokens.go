package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type tokenStore struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

func tokenPath() string {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".orbitdeploy")
	_ = os.MkdirAll(base, 0o700)
	return filepath.Join(base, "tokens.json")
}

func saveAccessToken(t string) error {
	p := tokenPath()
	var store tokenStore
	_ = readJSON(p, &store)
	store.AccessToken = t
	return writeJSON0600(p, &store)
}

func saveRefreshToken(t string) error {
	p := tokenPath()
	var store tokenStore
	_ = readJSON(p, &store)
	store.RefreshToken = t
	return writeJSON0600(p, &store)
}

func loadAccessToken() string {
	p := tokenPath()
	var store tokenStore
	_ = readJSON(p, &store)
	return store.AccessToken
}

func loadRefreshToken() string {
	p := tokenPath()
	var store tokenStore
	_ = readJSON(p, &store)
	return store.RefreshToken
}

func clearTokens() {
	p := tokenPath()
	_ = os.Remove(p)
}

func readJSON(p string, v any) error {
	f, err := os.Open(p)
	if err != nil { return err }
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

func writeJSON0600(p string, v any) error {
	f, err := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil { return err }
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// debug only
func must[T any](v T, err error) T { if err != nil { panic(err) }; return v }

func ensure(err error) { if err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) } }
