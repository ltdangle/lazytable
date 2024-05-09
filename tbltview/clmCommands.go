package main

import (
	"regexp"
	"strconv"
)

// Command-line mode commands. that can be parsed from the command line.
type ClmCommand interface {
	// Regex match of the command string.
	Match(text string) (bool, Command)
}

type SortColStrAscClmCommand struct {
	*SortColStrAscCommand
}

func NewSortColStrAscClmCommand() *SortColStrAscClmCommand {
	return &SortColStrAscClmCommand{}
}

func (clm *SortColStrAscClmCommand) Match(text string) (bool, Command) {
	var matches []string
	re, err := regexp.Compile(`sortasc (\d+)`)
	if err != nil {
		return false, nil
	}

	// Find submatch returns the entire match and any parenthesized submatches
	matches = re.FindStringSubmatch(text)
	if matches == nil || len(matches) < 2 {
		return false, nil
	}

	column, err := strconv.Atoi(matches[1])

	if err != nil {
		return false, nil
	}

	return true, NewSortColStrAscCommand(column + 1)

}

type ReplaceClmCommand struct {
}

func NewReplaceClmCommand() *ReplaceClmCommand {
	return &ReplaceClmCommand{}
}
func (clm *ReplaceClmCommand) regex(text string) (ok bool, search string, replace string) {
	// Compile the regular expression to match the whole pattern and extract the content within the quotes
	re, err := regexp.Compile(`^replace '([^']*)' with '([^']*)'$`)
	if err != nil {
		return
	}

	// Find submatches, ensuring the whole string matches the pattern
	match := re.FindStringSubmatch(text)
	if match == nil {
		return
	}

	ok = true
	search = match[1]
	replace = match[2]
	return
}
func (clm *ReplaceClmCommand) Match(text string) (bool, Command) {
	ok, search, replace := clm.regex(text)
	if !ok {
		return false, nil
	}
	return true, NewChangeCellValueCommand(1, 1, search+"=>"+replace)

}
