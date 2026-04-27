package filemail

import (
	"errors"
	"goclaw/services/email"

	"github.com/google/uuid"
)

var _ email.Client = &Client{}

func NewClient(user email.Person, loc string) (*Client, error) {
	return &Client{user, nil}, nil
}

type Client struct {
	user   email.Person
	emails []email.Email
}

// Emails implements [email.Client].
func (c *Client) Emails(opts ...email.EmailsOpt) (email.Emails, error) {
	args := email.GetEmailsArgs(opts...)

	// Include emails only before the page
	result := make([]email.Email, 0)
	for _, em := range c.emails {
		if em.ID == args.Page {
			break
		}
		result = append(result, em)
	}

	// Include emails only after the only after
	for i, em := range result {
		if em.ID == args.IgnoreAfterID {
			result = result[i+1:]
			break
		}
	}

	// Ignore filters

	// Include max n
	page := ""
	if len(result) > args.MaxResults {
		page = result[len(result)-args.MaxResults].ID
		result = result[len(result)-args.MaxResults:]
	}

	return email.Emails{Emails: result, NextPage: page}, nil
}

// GetEmail implements [email.Client].
func (c *Client) GetEmail(id string) (email.Email, error) {
	for _, em := range c.emails {
		if em.ID == id {
			return em, nil
		}
	}
	return email.Email{}, errors.New("id not found")
}

// User implements [email.Client].
func (c *Client) User() (email.Person, error) {
	return c.user, nil
}

func (c *Client) InsertEmail(e email.Email) (string, error) {
	e.ID = uuid.NewString()
	c.emails = append(c.emails, e)
	return e.ID, nil
}
