package gmailplugin

import "errors"

var ErrNoSentEmails = errors.New("there are no sent emails in your account - we can only infer your email address from your sent emails")

func (c *Client) User() (Person, error) {
	results, err := c.Emails(WithQuery("in:sent"), WithMaxN(1))
	if err != nil {
		return Person{}, err
	}
	return results.Emails[0].From, nil
}
