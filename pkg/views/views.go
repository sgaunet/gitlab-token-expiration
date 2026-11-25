package views

import (
	"errors"

	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
)

// ErrUnknownFormat is returned for an unsupported output format.
var ErrUnknownFormat = errors.New("unknown output format")

// Renderer is an interface for rendering token information.
type Renderer interface {
	Render(tokens []dto.Token) error
}
