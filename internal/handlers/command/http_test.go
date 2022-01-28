package command

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	payload := strings.NewReader(`{"command": "turn on PC","created": "January 22, 2022 at 11:10AM"}`)
	decoded, err := decodeCommand(payload)

	assert.NoError(t, err)
	assert.Equal(t, "turn on pc", decoded.Command)
	tz, err := time.LoadLocation(timeZone)
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2022, 1, 22, 11, 10, 0, 0, tz), decoded.Created.Time)
}
