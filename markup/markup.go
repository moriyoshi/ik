package markup

import "github.com/moriyoshi/ik"

type MarkupRenderer interface {
    Render(markup *ik.Markup)
}
