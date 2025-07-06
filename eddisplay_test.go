package main

import (
	"testing"

	"github.com/pellux-network/EDx52display/lcdformat"
)

// LCD Formatter Testing
func TestSpaceBetweenEqual(t *testing.T) {
	println("Testing SpaceBetween function with only two strings")
	strings := []string{"Hello", "World"}
	width := 16
	result := lcdformat.SpaceBetween(strings, width)
	// display output of string format
	header := lcdformat.NRawChar(width, "#")
	println(header)
	println(result)
	println(header)
	got := result
	want := "Hello      World"

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}
func TestSpaceBetweenLopsided(t *testing.T) {
	println("Testing SpaceBetween function with two unequal strings")
	strings := []string{"DST:", "CloseLand22"}
	width := 16
	result := lcdformat.SpaceBetween(strings, width)
	// display output of string format
	header := lcdformat.NRawChar(width, "#")
	println(header)
	println(result)
	println(header)
	got := result
	want := "DST: CloseLand22"

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestSpaceBetweenMulti(t *testing.T) {
	println("Testing SpaceBetween function with multiple different-length strings")
	strings := []string{"J:2", "P:47", "X:4", "E:1"}
	width := 16
	result := lcdformat.SpaceBetween(strings, width)
	// display output of string format
	header := lcdformat.NRawChar(width, "#")
	println(header)
	println(result)
	println(header)
	got := result
	want := "J:2 P:47 X:4 E:1"

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}
