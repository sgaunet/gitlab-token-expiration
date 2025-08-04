package views

import "github.com/sgaunet/gitlab-token-expiration/pkg/dto"

// Renderer is an interface for rendering token information.
type Renderer interface {
	Render(tokens []dto.Token) error
}
