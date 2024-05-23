package main

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"tblview/data"
)

// Command-line mode commands. that can be parsed from the command line.
type ClmCommand interface {
	// Regex match of the command string.
	Match(text string) (parsed bool, commandError error, command Command)
}

type SortColStrAscClmCommand struct {
	*SortColStrAscCommand
}

func NewSortColStrAscClmCommand() *SortColStrAscClmCommand {
	return &SortColStrAscClmCommand{}
}

func (clm *SortColStrAscClmCommand) Match(text string) (parsed bool, commandError error, command Command) {
	var matches []string
	re, err := regexp.Compile(`sortasc (\d+)`)
	if err != nil {
		return
	}

	// Find submatch returns the entire match and any parenthesized submatches
	matches = re.FindStringSubmatch(text)
	if matches == nil || len(matches) < 2 {
		return
	}

	column, err := strconv.Atoi(matches[1])

	if err != nil {
		return
	}

	parsed = true
	command = NewSortColStrAscCommand(column + 1)
	return
}

type ReplaceClmCommand struct {
	selection *data.Selection
}

func NewReplaceClmCommand() *ReplaceClmCommand {
	return &ReplaceClmCommand{selection: selection}
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

func (clm *ReplaceClmCommand) Match(text string) (parsed bool, commandError error, command Command) {
	ok, search, replace := clm.regex(text)

	if !ok {
		parsed = false
		return
	}

	if selection == nil {
		parsed = true
		commandError = errors.New("no cells selected")
		return
	}

	parsed = true
	command = NewReplaceTextCommand(selection, search, replace)
	return
}

type WriteFileClmCommand struct {
}

func NewWriteFileClmCommand() *WriteFileClmCommand {
	return &WriteFileClmCommand{}
}

func (clm *WriteFileClmCommand) Match(text string) (parsed bool, commandError error, command Command) {
	args := strings.Split(text, " ")
	if len(args) < 2 || args[0] != "w" {
		parsed = false
		return
	}

	return true, nil, NewWriteFileCommand(args[1])
}
type LoadFileClmCommand struct {
}

func NewLoadFileClmCommand() *LoadFileClmCommand {
	return &LoadFileClmCommand{}
}

func (clm *LoadFileClmCommand) Match(text string) (parsed bool, commandError error, command Command) {
	args := strings.Split(text, " ")
	if len(args) < 2 || args[0] != "e" {
		parsed = false
		return
	}

	return true, nil, NewLoadFileCommand(args[1])
}
