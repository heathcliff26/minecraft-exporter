package rcon

import "fmt"

type ErrRCONMissingHost struct{}

func (e ErrRCONMissingHost) Error() string {
	return "Missing target host for RCON"
}

type ErrRCONMissingPort struct{}

func (e ErrRCONMissingPort) Error() string {
	return "Missing target port for RCON"
}

type ErrRCONMissingPassword struct{}

func (e ErrRCONMissingPassword) Error() string {
	return "Missing password for RCON"
}

type ErrRCONConnectionTimeout struct{}

func (e ErrRCONConnectionTimeout) Error() string {
	return "Timed out waiting for a response"
}

type ErrForgeTPS struct{}

func (e ErrForgeTPS) Error() string {
	return "Failed to retrieve the overall tps stats"
}

type ErrPaperTPS struct {
	Text  string
	Count int
}

func NewErrPaperTPS(text string, count int) error {
	return &ErrPaperTPS{
		Text:  text,
		Count: count,
	}
}

func (e *ErrPaperTPS) Error() string {
	return fmt.Sprintf("Expected at 3 values, got %d. Input: \"%s\"", e.Count, e.Text)
}
