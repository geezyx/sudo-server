package command

import (
	"errors"
	"strings"

	"github.com/geezyx/sudo-server/internal/core/domain"
)

type service struct {
	commands *domain.Commands
}

func New() *service {
	return &service{
		&domain.Commands{
			Action:   nil,
			IsLeaf:   false,
			Children: nil,
		},
	}
}

var ErrUnknownCommand = errors.New("unknown command")
var ErrEmptyCommand = errors.New("empty command")
var ErrDuplicateCommand = errors.New("duplicate command")
var ErrCommandHasChildren = errors.New("command with children not allowed")
var ErrCommandOverlap = errors.New("command overlaps with another command")

// Add adds the provided command to the command trie with the provided handler
func (s *service) Add(cmd string, action domain.ActionFunc) error {
	elements := strings.Split(cmd, " ")
	return s.add(s.commands, elements, action)
}

// add recursively builds the command trie with handlers.
func (s *service) add(cs *domain.Commands, elements []string, action domain.ActionFunc) error {
	// we have reached the end of the command string and can register the leaf
	if len(elements) == 0 {
		// if the current node has children, it means we're trying to add a new command
		// that overlaps with an existing command which is unsupported
		if cs.Children != nil {
			return ErrCommandHasChildren
		}
		if cs.IsLeaf {
			return ErrDuplicateCommand
		}
		cs.IsLeaf = true
		cs.Action = action
		return nil
	}
	// if the current node is a leaf, it means we're trying to add a new command
	// that overlaps with an existing command which is unsupported
	if cs.IsLeaf {
		return ErrCommandOverlap
	}
	// add or update child
	if cs.Children == nil {
		cs.Children = make(map[string]*domain.Commands)
	}
	if cs.Children[elements[0]] == nil {
		cs.Children[elements[0]] = &domain.Commands{}
	}

	return s.add(cs.Children[elements[0]], elements[1:], action)
}

// GetHandler finds the handler matching the command
func (s *service) GetAction(command string) (domain.ActionFunc, error) {
	if command == "" {
		return nil, ErrEmptyCommand
	}
	elements := strings.Split(command, " ")
	return s.getAction(s.commands, elements)
}

// getHandler recursively traverses the command trie until it finds
// a matching leaf with a handler.
func (s *service) getAction(cs *domain.Commands, elements []string) (domain.ActionFunc, error) {
	if len(elements) == 0 && cs.IsLeaf {
		return cs.Action, nil
	}
	if child, found := cs.Children[elements[0]]; found {
		return s.getAction(child, elements[1:])
	}
	return nil, ErrUnknownCommand
}
