package gmailplugin

func (c *Client) Emails(opts ...listEmailOpt) (Emails, error) {
	args := listEmailArgs{maxResults: 10}
	for _, o := range opts {
		o(&args)
	}
	request := c.service.Users.Messages.List("me").MaxResults(int64(args.maxResults))
	if args.query != "" {
		request = request.Q(args.query)
	}
	if args.page != "" {
		request = request.PageToken(args.page)
	}
	res, err := request.Do()
	if err != nil {
		return Emails{}, err
	}

	emails := make([]Email, 0)
	nextPageToken := res.NextPageToken

	for _, m := range res.Messages {
		if m.Id == args.ignoreAfterID {
			nextPageToken = ""
			break
		}
		email, err := c.GetEmail(m.Id)
		if err != nil {
			return Emails{}, err
		}
		emails = append(emails, email)
	}

	return Emails{emails, nextPageToken}, nil
}

type listEmailOpt func(*listEmailArgs)
type listEmailArgs struct {
	maxResults    int
	query         string
	page          string
	ignoreAfterID string
}

// Specify the maximum number of emails to be returned on this page.
// You can call the same function with the next page to get further results.
func WithMaxN(maxResults int) listEmailOpt {
	return func(lea *listEmailArgs) {
		lea.maxResults = maxResults
	}
}

// Specify a query (same syntax as gmail search box in the web app) to search for.
func WithQuery(query string) listEmailOpt {
	return func(lea *listEmailArgs) {
		lea.query = query
	}
}

// Specify the page to get - this page ID is returned from previous calls to the Emails function.
func WithPage(page string) listEmailOpt {
	return func(lea *listEmailArgs) {
		lea.page = page
	}
}

// Specify an ID of an email after which we will return no more results, and no further pages.
// This is used to poll the inbox to discover which emails have just been returned.
func WithNoEmailsAfter(emailID string) listEmailOpt {
	return func(lea *listEmailArgs) {
		lea.ignoreAfterID = emailID
	}
}
