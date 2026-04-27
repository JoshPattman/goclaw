package googlemail

import "goclaw/services/email"

func (c *Client) Emails(opts ...email.EmailsOpt) (email.Emails, error) {
	args := email.GetEmailsArgs(opts...)
	request := c.service.Users.Messages.List("me").MaxResults(int64(args.MaxResults))
	if args.Query != "" {
		request = request.Q(args.Query)
	}
	if args.Page != "" {
		request = request.PageToken(args.Page)
	}
	res, err := request.Do()
	if err != nil {
		return email.Emails{}, err
	}

	emails := make([]email.Email, 0)
	nextPageToken := res.NextPageToken

	for _, m := range res.Messages {
		if m.Id == args.IgnoreAfterID {
			nextPageToken = ""
			break
		}
		em, err := c.GetEmail(m.Id)
		if err != nil {
			return email.Emails{}, err
		}
		emails = append(emails, em)
	}

	return email.Emails{Emails: emails, NextPage: nextPageToken}, nil
}
