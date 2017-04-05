package main

import "testing"

func TestMain(t *testing.T) {
	exit = func(int) {}
	getwd = func() (string, error) { return "", nil }
	operate = func([]byte) []byte { return []byte{} }
	main()
}
