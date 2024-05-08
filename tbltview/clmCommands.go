package main

import (
	"regexp"
	"strconv"
)

// Command-line mode commands. that can be parsed from the command line.
type ClmCommand interface {
	// Regex match of the command string.
	Match(text string) bool
	// Undoable command.
	Command
}

type SortColStrAscClmCommand struct {
	*SortColStrAscCommand
}

func NewSortColStrAscClmCommand() *SortColStrAscClmCommand {
	return &SortColStrAscClmCommand{}
}

func (clm *SortColStrAscClmCommand) Match(text string) bool {
	var matches []string
	re, err := regexp.Compile(`sortasc (\d+)`)
	if err != nil {
		return false
	}

	// Find submatch returns the entire match and any parenthesized submatches
	matches = re.FindStringSubmatch(text)
	if matches == nil || len(matches) < 2 {
		return false
	}

	column, err := strconv.Atoi(matches[1])

	if err != nil {
		return false
	}

	clm.SortColStrAscCommand = NewSortColStrAscCommand(column + 1)

	return true
}
