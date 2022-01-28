package configloader

import (
	"github.com/geezyx/sudo-server/internal/core/domain"
	"github.com/geezyx/sudo-server/internal/core/services/config"
)

func Load(path string) (domain.Config, error) {
	return config.New().Load(path)
}
