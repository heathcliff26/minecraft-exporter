package save

import "fmt"

type ErrNoWorldDirectory struct {
	details string
}

func NewErrNoWorldDirectory(details string) *ErrNoWorldDirectory {
	return &ErrNoWorldDirectory{
		details: details,
	}
}

func (e *ErrNoWorldDirectory) Error() string {
	return "No valid world directory provided: " + e.details
}

func NewErrFailedToParseStat(name string, value int) *ErrFailedToParseStat {
	return &ErrFailedToParseStat{
		Name:  name,
		Value: value,
	}
}

type ErrFailedToParseStat struct {
	Name  string
	Value int
}

func (e *ErrFailedToParseStat) Error() string {
	return fmt.Sprintf("Failed to parse the stat (\"%s\": %d)", e.Name, e.Value)
}
