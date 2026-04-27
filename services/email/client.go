package email

type Client interface {
	User() (Person, error)
	GetEmail(id string) (Email, error)
	Emails(opts ...EmailsOpt) (Emails, error)
}

type EmailsOpt func(*EmailsArgs)
type EmailsArgs struct {
	MaxResults    int
	Query         string
	Page          string
	IgnoreAfterID string
}

// Specify the maximum number of emails to be returned on this page.
// You can call the same function with the next page to get further results.
func WithMaxN(maxResults int) EmailsOpt {
	return func(lea *EmailsArgs) {
		lea.MaxResults = maxResults
	}
}

// Specify a query (same syntax as gmail search box in the web app) to search for.
func WithQuery(query string) EmailsOpt {
	return func(lea *EmailsArgs) {
		lea.Query = query
	}
}

// Specify the page to get - this page ID is returned from previous calls to the Emails function.
func WithPage(page string) EmailsOpt {
	return func(lea *EmailsArgs) {
		lea.Page = page
	}
}

// Specify an ID of an email after which we will return no more results, and no further pages.
// This is used to poll the inbox to discover which emails have just been returned.
func WithNoEmailsAfter(emailID string) EmailsOpt {
	return func(lea *EmailsArgs) {
		lea.IgnoreAfterID = emailID
	}
}

func GetEmailsArgs(opts ...EmailsOpt) EmailsArgs {
	args := EmailsArgs{MaxResults: 10}
	for _, o := range opts {
		o(&args)
	}
	return args
}
