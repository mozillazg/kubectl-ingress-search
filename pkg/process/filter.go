package process

import (
	"regexp"

	"github.com/mozillazg/kubectl-ingress-search/pkg/ingress"
)

type Filter interface {
	Filter([]ingress.Rule) []ingress.Rule
}

type FieldValueFilter struct {
	Name    string
	Exp     *regexp.Regexp
	NoColor bool
}

func (f FieldValueFilter) Filter(rules []ingress.Rule) []ingress.Rule {
	var newRuels []ingress.Rule
	for _, r := range rules {
		var match bool
		switch f.Name {
		case "namespace":
			if newV, ok := f.apply(r.Namespace.Render()); ok {
				v := r.Namespace
				v.Rendered = newV
				r.Namespace = v
				match = true
			}
			break
		case "name":
			if newV, ok := f.apply(r.Name.Render()); ok {
				v := r.Name
				v.Rendered = newV
				r.Name = v
				match = true
			}
			break
		case "host":
			if newV, ok := f.apply(r.Host.Render()); ok {
				v := r.Host
				v.Rendered = newV
				r.Host = v
				match = true
			}
		case "path":
			if newV, ok := f.apply(r.Path.Render()); ok {
				v := r.Path
				v.Rendered = newV
				r.Path = v
				match = true
			}
		case "backend":
			if newV, ok := f.apply(r.Backend.Render()); ok {
				v := r.Backend
				v.Rendered = newV
				r.Backend = v
				match = true
			}
		}
		if match {
			newRuels = append(newRuels, r)
		}
	}
	return newRuels
}

func (f FieldValueFilter) apply(v string) (string, bool) {
	match := len(f.Exp.FindStringIndex(v)) > 0
	newStr := v
	if match && !f.NoColor {
		newStr = highlight(v, f.Exp)
	}
	return newStr, match
}

type HighlightDupServiceFilter struct {
}

func (f HighlightDupServiceFilter) Filter(rules []ingress.Rule) []ingress.Rule {
	services := map[string]int{}
	for _, r := range rules {
		if b, ok := r.Backend.V.(ingress.Backend); ok && b.Service != nil {
			services[b.Service.Name] += 1
		}
	}
	var newRules []ingress.Rule
	for _, r := range rules {
		if b, ok := r.Backend.V.(ingress.Backend); ok && b.Service != nil {
			num := services[b.Service.Name]
			if num > 1 {
				newV, _ := FieldValueFilter{
					Name:    "service",
					Exp:     regexp.MustCompile("Service/" + b.Service.Name),
					NoColor: false,
				}.apply(r.Backend.Render())
				v := r.Backend
				v.Rendered = newV
				r.Backend = v
			}
		}
		newRules = append(newRules, r)
	}
	return newRules
}
