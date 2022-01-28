package command

import (
	"testing"

	"github.com/geezyx/sudo-server/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func mockAction() error {
	return nil
}

func TestAdd(t *testing.T) {
	assert := assert.New(t)

	for name, tt := range map[string]struct {
		svc             *service
		command         string
		expectedErr     bool
		expectedErrType error
	}{
		"empty command on new instance": {
			svc:         New(),
			command:     "",
			expectedErr: false,
		},
		"one word command on new instance": {
			svc:         New(),
			command:     "hi",
			expectedErr: false,
		},
		"two word command on new instance": {
			svc:         New(),
			command:     "hi there",
			expectedErr: false,
		},
		"command on existing instance": {
			svc: &service{
				commands: &domain.Commands{
					Children: map[string]*domain.Commands{
						"hi": {
							Children: map[string]*domain.Commands{
								"there": {
									IsLeaf: true,
									Action: mockAction,
								},
							},
						},
					},
				},
			},
			command:         "hi there",
			expectedErr:     true,
			expectedErrType: ErrDuplicateCommand,
		},
		"overlapping command on existing instance": {
			svc: &service{
				commands: &domain.Commands{
					Children: map[string]*domain.Commands{
						"hi": {
							Children: map[string]*domain.Commands{
								"there": {
									IsLeaf: true,
									Action: mockAction,
								},
							},
						},
					},
				},
			},
			command:         "hi there bob",
			expectedErr:     true,
			expectedErrType: ErrCommandOverlap,
		},
		"command with other children on existing instance": {
			svc: &service{
				commands: &domain.Commands{
					Children: map[string]*domain.Commands{
						"hi": {
							Children: map[string]*domain.Commands{
								"there": {
									IsLeaf: true,
									Action: mockAction,
								},
							},
						},
					},
				},
			},
			command:         "hi",
			expectedErr:     true,
			expectedErrType: ErrCommandHasChildren,
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := tt.svc.Add(tt.command, mockAction)
			if tt.expectedErr {
				assert.Error(err)
				assert.ErrorIs(err, tt.expectedErrType)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	assert := assert.New(t)

	exampleSvc := New()
	exampleSvc.Add("who is alice", mockAction)
	exampleSvc.Add("who is bob", mockAction)
	exampleSvc.Add("who is the noid", mockAction)

	for name, tt := range map[string]struct {
		command         string
		expectedErr     bool
		expectedErrType error
	}{
		"empty command": {
			command:         "",
			expectedErr:     true,
			expectedErrType: ErrEmptyCommand,
		},
		"unknown command": {
			command:         "hi",
			expectedErr:     true,
			expectedErrType: ErrUnknownCommand,
		},
		"unknown command after leaf": {
			command:         "who is bob barker",
			expectedErr:     true,
			expectedErrType: ErrUnknownCommand,
		},
		"unknown command mid trie": {
			command:         "who is the one",
			expectedErr:     true,
			expectedErrType: ErrUnknownCommand,
		},
		"valid command alice": {
			command:     "who is alice",
			expectedErr: false,
		},
		"valid command bob": {
			command:     "who is bob",
			expectedErr: false,
		},
		"valid command the noid": {
			command:     "who is the noid",
			expectedErr: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			handler, err := exampleSvc.GetAction(tt.command)
			if tt.expectedErr {
				assert.Error(err)
				assert.ErrorIs(err, tt.expectedErrType)
			} else {
				assert.NotNil(handler)
				assert.NoError(err)
			}
		})
	}
}
