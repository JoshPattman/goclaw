package mail

import (
	"io"

	"goclaw/services/email"

	"github.com/nelsonken/emlparser-go"
)

func ReadEML(f io.Reader) (email.Email, error) {
	raw, err := io.ReadAll(f)
	if err != nil {
		return email.Email{}, err
	}

	m, err := emlparser.Parse(raw)
	if err != nil {
		return email.Email{}, err
	}

	// Helper to convert addresses
	convert := func(addrs []emlparser.Address) []email.Person {
		out := make([]email.Person, 0, len(addrs))
		for _, a := range addrs {
			out = append(out, email.Person{
				Name:  a.Name(),
				Email: a.Email(),
			})
		}
		return out
	}

	var from email.Person
	if len(m.From) > 0 {
		from = email.Person{
			Name:  m.From[0].Name(),
			Email: m.From[0].Email(),
		}
	}

	// Prefer HTML, fallback to text
	body := m.Html
	if body == "" {
		body = m.Text
	}

	e := email.Email{
		ID:      m.Id, // nice stable hash from library
		From:    from,
		To:      convert(m.To),
		Subject: m.Subject,
		Content: body,
		Labels:  []email.Label{email.InboxLabel}, // ingestion = inbox by default
	}

	return e, nil
}
