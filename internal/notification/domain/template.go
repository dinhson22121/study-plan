package domain

import (
	"regexp"
	"strings"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// placeholderRe matches {variable} tokens in a template string.
var placeholderRe = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)

// NotificationTemplate is the title/body template for a notification type. Title
// and Body may contain {variable} placeholders filled at render time.
type NotificationTemplate struct {
	Code     string
	Title    string
	Body     string
	Type     NotificationType
	IsActive bool
}

// Render substitutes {variable} placeholders in the title and body using vars.
// It returns ErrValidation if any placeholder is left unfilled, so a broken
// template fails loudly instead of sending "{name}" to a student.
func (t NotificationTemplate) Render(vars map[string]string) (title, body string, err error) {
	title, missingT := substitute(t.Title, vars)
	body, missingB := substitute(t.Body, vars)
	if len(missingT) > 0 || len(missingB) > 0 {
		missing := append(missingT, missingB...)
		return "", "", shared.ErrValidation.WithMessage("missing template variables: " + strings.Join(missing, ", "))
	}
	return title, body, nil
}

func substitute(s string, vars map[string]string) (string, []string) {
	var missing []string
	out := placeholderRe.ReplaceAllStringFunc(s, func(match string) string {
		key := match[1 : len(match)-1] // strip { }
		if v, ok := vars[key]; ok {
			return v
		}
		missing = append(missing, key)
		return match
	})
	return out, missing
}
