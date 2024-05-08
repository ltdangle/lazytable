package main

// Command-line mode commands. that can be parsed from the command line.
type ClmCommand interface {
	// Regex match of the command string.
	Match(text string) (ok bool, matches []string)
	// Undoable command.
	Command
}

type SortColStrAscClmCommand struct {
	*SortColStrAscCommand
}

func NewSortColStrAscClmCommand() *SortColStrAscClmCommand {
	return &SortColStrAscClmCommand{SortColStrAscCommand: NewSortColStrAscCommand(1)}
}

func (clm *SortColStrAscClmCommand) Match(text string) (ok bool, matches []string) {
	if text == "sortasc" {
		ok = true
	}
	return
}
