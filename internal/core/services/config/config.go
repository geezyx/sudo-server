package config

import (
	"os"
	"path/filepath"

	"github.com/geezyx/sudo-server/internal/core/domain"
	"gopkg.in/yaml.v2"
)

type service struct{}

func New() *service {
	return &service{}
}

func (srv *service) Load(path string) (domain.Config, error) {
	var cfg domain.Config
	filename, _ := filepath.Abs(path)
	f, err := os.ReadFile(filename)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(f, &cfg)
	return cfg, err
}
