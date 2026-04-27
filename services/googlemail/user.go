package googlemail

import (
	"errors"
	"goclaw/services/email"
)

var ErrNoSentEmails = errors.New("there are no sent emails in your account - we can only infer your email address from your sent emails")

func (c *Client) User() (email.Person, error) {
	results, err := c.Emails(email.WithQuery("in:sent"), email.WithMaxN(1))
	if err != nil {
		return email.Person{}, err
	}
	return results.Emails[0].From, nil
}
