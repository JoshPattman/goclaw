package gmailplugin

import (
	"fmt"

	"google.golang.org/api/gmail/v1"
)

// Create a new client, using the provided credential file.
// If a token already exists at the token location, will use that, otherwise will get a new token and store it there.
// The oath callback will be directed to the listen port (if running in a container, need to expose that port).
// To show the token to the user, call showToken (for example, this function could print the token to CLI, bring up an alert, or directly open the browser).
func BuildClient(
	credentialLocation string,
	tokenLocation string,
	listenPort int,
	showToken func(string),
) (*Client, error) {
	config, err := readConfigFromFile(credentialLocation, listenPort)
	if err != nil {
		return nil, err
	}
	token, err := readTokenFromFile(tokenLocation)
	if err != nil {
		tokenURL := getAuthURL(config)
		fmt.Printf("Go to the following URL to login: %s\n", tokenURL)
		token, err = listenForAuth(config, listenPort)
		if err != nil {
			return nil, err
		}
		err = saveTokenToFile(token, tokenLocation)
		if err != nil {
			return nil, err
		}
	}
	service, err := buildService(config, token)
	if err != nil {
		return nil, err
	}
	return &Client{service}, nil
}

// A client is a wrapper around the gmail api that provides convienient functions for interacting with your inbox.
type Client struct {
	service *gmail.Service
}

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
