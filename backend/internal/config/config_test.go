package config

import "testing"

func TestGetenvIntStrictParsing(t *testing.T) {
	t.Setenv("TEST_INT", "3")
	if got := getenvInt("TEST_INT", 0); got != 3 {
		t.Fatalf("got %d", got)
	}
	t.Setenv("TEST_INT", "3junk")
	if got := getenvInt("TEST_INT", 7); got != 7 {
		t.Fatalf("invalid suffix should use default, got %d", got)
	}
	t.Setenv("TEST_INT", "-1")
	if got := getenvInt("TEST_INT", 7); got != 7 {
		t.Fatalf("negative DB should use default, got %d", got)
	}
}
