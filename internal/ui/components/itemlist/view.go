package itemlist

import (
	"home-inventory-system/internal/ui/styles"
)

// RenderError renders an error message
func RenderError(err error) string {
	if err == nil {
		return ""
	}
	return styles.ErrorStyle.Render(err.Error())
}
