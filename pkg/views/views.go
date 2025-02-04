package views

import "github.com/sgaunet/gitlab-token-expiration/pkg/dto"

type Renderer interface {
	Render(tokens []dto.Token) error
}
