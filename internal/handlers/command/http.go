// ./internal/handlers/gamehdl/http.go

package command

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/geezyx/sudo-server/internal/core/domain"
)

type Commander interface {
	Add(cmd string, handler domain.ActionFunc) error
	GetAction(command string) (domain.ActionFunc, error)
}

type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(err error, msg string, keysAndValues ...interface{})
}

type HTTPHandler struct {
	command Commander
	log     Logger
}

func NewHTTPHandler(c Commander, l Logger) *HTTPHandler {
	return &HTTPHandler{
		command: c,
		log:     l,
	}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := decodeCommand(r.Body)
	if err != nil {
		h.log.Error(err, "invalid payload")
		http.Error(w, fmt.Sprintf("invalid payload: %v", err), http.StatusBadRequest)
		return
	}

	action, err := h.command.GetAction(c.Command)
	if err != nil {
		h.log.Error(err, "command", c.Command)
		http.Error(w, fmt.Sprintf("%v: %v", err, c.Command), http.StatusUnprocessableEntity)
		return
	}

	err = action()
	if err != nil {
		h.log.Error(err, "invalid payload")
		http.Error(w, fmt.Sprintf("error running command: %v", err), http.StatusInternalServerError)
	}

	h.log.Info("success", "command", c.Command, "created", c.Created)
	w.WriteHeader(http.StatusOK)
}

const (
	// IFTTT example: January 22, 2022 at 11:10AM
	iftttTimeFormat = `January 2, 2006 at 3:04PM`
	timeZone        = "America/New_York"
)

type IFTTTTime struct {
	time.Time
}

func (t *IFTTTTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	tz, err := time.LoadLocation(timeZone)
	if err != nil {
		return err
	}
	ret, err := time.ParseInLocation(iftttTimeFormat, s, tz)
	if err != nil {
		return err
	}
	t.Time = ret
	return nil
}

func (t IFTTTTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Format(iftttTimeFormat) + `"`), nil
}

func (t IFTTTTime) String() string {
	return t.Format(iftttTimeFormat)
}

type SudoContent struct {
	Command string    `json:"command"`
	Created IFTTTTime `json:"created"`
}

func decodeCommand(body io.Reader) (SudoContent, error) {
	var c SudoContent
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&c); err != nil {
		return c, err
	}
	c.Command = strings.ToLower(c.Command)
	return c, nil
}
