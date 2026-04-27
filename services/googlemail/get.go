package googlemail

import (
	"encoding/base64"
	"goclaw/services/email"
	"strings"

	"google.golang.org/api/gmail/v1"
)

func extractBody(p *gmail.MessagePart) string {
	if p == nil {
		return ""
	}

	// If this part has data, decode it
	if p.Body != nil && p.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(p.Body.Data)
		if err == nil {
			return string(data)
		}
	}

	// Otherwise, recurse into parts
	for _, part := range p.Parts {
		// Prefer plain text if available
		if part.MimeType == "text/plain" {
			return extractBody(part)
		}
	}

	// Fallback: try anything
	for _, part := range p.Parts {
		if content := extractBody(part); content != "" {
			return content
		}
	}

	return ""
}

func (c *Client) GetEmail(id string) (email.Email, error) {
	msg, err := c.service.Users.Messages.Get("me", id).Do()
	if err != nil {
		return email.Email{}, err
	}

	labels := make([]email.Label, len(msg.LabelIds))
	for i, l := range msg.LabelIds {
		labels[i] = email.Label(l)
	}

	em := email.Email{ID: id, Labels: labels}

	for _, h := range msg.Payload.Headers {
		switch h.Name {
		case "Subject":
			em.Subject = h.Value
		case "From":
			person, err := parsePerson(h.Value)
			if err != nil {
				return email.Email{}, err
			}
			em.From = person
		case "To":
			for _, raw := range strings.Split(h.Value, ", ") {
				person, err := parsePerson(raw)
				if err != nil {
					return email.Email{}, err
				}
				em.To = append(em.To, person)
			}
		}
	}
	em.Content = extractBody(msg.Payload)
	return em, nil
}

func parsePerson(raw string) (email.Person, error) {
	parts := strings.SplitN(raw, "<", 2)
	var name, em string
	if len(parts) > 1 {
		name = strings.TrimSpace(parts[0])
		em = strings.Trim(parts[1], " \n\r\t>")
	} else {
		name = ""
		em = strings.TrimSpace(parts[0])
	}
	if name == "" {
		name = getNameFromEmail(em)
	}
	return email.Person{Name: name, Email: em}, nil
}

func getNameFromEmail(email string) string {
	return strings.Split(email, "@")[0]
}
