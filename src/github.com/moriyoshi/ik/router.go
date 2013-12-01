package ik

type Pattern struct {
	chunks []string
	prefix string
}

type Router struct {
	patterns []Pattern
}

func (router *Router) AddPattern(Pattern

func (router *Router) Emit(records []FluentRecord) error {
	for _, record := range records {
		if record.Tag
	}
}

