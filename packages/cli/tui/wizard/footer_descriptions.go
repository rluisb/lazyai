package wizard

import (
	"strings"

	"charm.land/huh/v2"

	"github.com/rluisb/lazyai/packages/cli/internal/theme"
)

type selectWithFooterDescription struct {
	*huh.Select[string]
	description func() string
}

func selectFooterDescription(field *huh.Select[string], description func() string) *selectWithFooterDescription {
	return &selectWithFooterDescription{Select: field, description: description}
}

func (f *selectWithFooterDescription) View() string {
	return viewWithFooterDescription(f.Select.View(), f.description)
}

type multiSelectWithFooterDescription struct {
	*huh.MultiSelect[string]
	description func() string
}

func multiSelectFooterDescription(field *huh.MultiSelect[string], description func() string) *multiSelectWithFooterDescription {
	return &multiSelectWithFooterDescription{MultiSelect: field, description: description}
}

func (f *multiSelectWithFooterDescription) View() string {
	return viewWithFooterDescription(f.MultiSelect.View(), f.description)
}

func viewWithFooterDescription(view string, description func() string) string {
	if description == nil {
		return view
	}
	text := strings.TrimSpace(description())
	if text == "" {
		return view
	}
	return view + "\n" + theme.DimText("Info: "+text)
}
