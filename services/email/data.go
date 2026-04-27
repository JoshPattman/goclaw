package email

import "fmt"

// A person is someone who an email has been sent to / has sent an email.
// If gmail provides a pretty name, this will be used as their name, otherwise, the name will be inferred from the email.
type Person struct {
	Name  string
	Email string
}

func (p Person) String() string {
	return fmt.Sprintf("%s <%s>", p.Name, p.Email)
}

// An email is a fully-fetched email resource from the API.
type Email struct {
	ID      string
	From    Person
	To      []Person
	Subject string
	Content string
	Labels  []Label
}

// Emails is a container for a set of emails, that may have another page to retrieve.
type Emails struct {
	Emails   []Email
	NextPage string
}

type Label string

const (
	InboxLabel     Label = "INBOX"
	SentLabel      Label = "SENT"
	UnreadLabel    Label = "UNREAD"
	ImportantLabel Label = "IMPORTANT"
)
