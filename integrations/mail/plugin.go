package mail

import (
	"fmt"
	"goclaw/services/email"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/JoshPattman/cg"

	markdown "github.com/JohannesKaufmann/html-to-markdown"
)

func NewPlugin(name string, buildClient func() (email.Client, error), opts ...pluginOpt) *gmailPlugin {
	kwargs := pluginKwarg{}
	for _, o := range opts {
		o(&kwargs)
	}
	return &gmailPlugin{
		name,
		buildClient,
		kwargs.ingestionRoot,
	}
}

type pluginKwarg struct {
	ingestionRoot string
}

type pluginOpt func(*pluginKwarg)

func WithIngestionFromFolder(root string) pluginOpt {
	return func(pk *pluginKwarg) {
		pk.ingestionRoot = root
	}
}

type gmailPlugin struct {
	name          string
	buildClient   func() (email.Client, error)
	ingestionRoot string
}

func (p *gmailPlugin) Name() string {
	return p.name
}

func (p *gmailPlugin) Internal() bool { return false }

type emailSaver interface {
	InsertEmail(e email.Email) (string, error)
}

func (p *gmailPlugin) Load() ([]cg.Tool, <-chan cg.Event, func(), error) {
	client, err := p.buildClient()
	if err != nil {
		return nil, nil, nil, err
	}
	tools := []cg.Tool{
		&gmailListTool{client},
		&gmailViewTool{client},
		&gmailViewTool{client},
	}
	events := make(chan cg.Event)

	stop := make(chan struct{}, 1)
	go func() {
		latestEmails, err := client.Emails(email.WithMaxN(1))
		if err != nil {
			return // TODO: Should probably send an error message event
		}
		latestID := ""
		if len(latestEmails.Emails) > 0 {
			latestID = latestEmails.Emails[0].ID
		}
		for {
			select {
			case <-stop:
				return
			default:
			}
			// This will drop emails if you get more than 10 in the interval, but that seems unlikely.
			newEmails, err := client.Emails(email.WithMaxN(10), email.WithNoEmailsAfter(latestID))
			if err != nil {
				continue // TODO: Should probably send an error message event
			}
			if len(newEmails.Emails) > 0 {
				latestID = newEmails.Emails[0].ID
				for _, em := range newEmails.Emails {
					if !slices.Contains(em.Labels, email.InboxLabel) {
						continue
					}
					events <- emailRecvEvent{em}
				}
			}
			time.Sleep(time.Second * 10)
		}
	}()

	stop2 := make(chan struct{}, 1)
	clientEmailSaver, ok := client.(emailSaver)
	if p.ingestionRoot != "" && ok {
		go func() {
			for {
				select {
				case <-stop2:
					return
				case <-time.After(time.Second * 5):
				}
				entries, err := os.ReadDir(p.ingestionRoot)
				if err != nil {
					continue // swallow error for now
				}

				for _, entry := range entries {
					if entry.IsDir() {
						continue
					}

					path := filepath.Join(p.ingestionRoot, entry.Name())

					func() {
						f, err := os.Open(path)
						if err != nil {
							return
						}
						defer f.Close()

						// Parse email
						em, err := ReadEML(f)
						if err != nil {
							return // ignore bad files
						}

						// Insert
						_, err = clientEmailSaver.InsertEmail(em)
						if err != nil {
							return // ignore insert failures
						}
					}()

					// Delete regardless of success/failure (hopper semantics)
					_ = os.Remove(path)
				}
			}
		}()
	}

	return tools, events, func() {
		stop <- struct{}{}
		stop2 <- struct{}{}
	}, nil
}

type emailRecvEvent struct {
	email email.Email
}

func (e emailRecvEvent) Kind() string {
	return "gmail_email_recv"
}

func (e emailRecvEvent) Content() cg.JsonObject {
	return cg.JsonObject{
		"email_id":    e.email.ID,
		"from":        e.email.From,
		"to":          e.email.To,
		"subject":     e.email.Subject,
		"explanation": "An email was received. You can use the gmail_view tool to view the content of this email and other recent emails. The email content is not included in this event because it could be very long, but you can view it with the tool. Alternatively, no action may be required here - you may simply choose to call end iteration.",
	}
}

type gmailListTool struct {
	client email.Client
}

func (t *gmailListTool) Def() cg.ToolDef {
	return cg.ToolDef{
		Name: "gmail_list",
		Desc: "Search for some emails (excluding bodies) from gmail. Optionally filter to specific emails. This is paginated - at most 5 emails will be returned at a time, you can then call the tool again with the same arguments but the follow up page ID to get the next page. The entire emails (including bodies) will be returned, so to search through large volumes of emails you should use the filter feature instead. Args: `filter`: [string, required] The filter text to filter by. This can be plaintext, but also supports most search options that the gmail app does (e.g. 'in:sent Dogs' searches for emails related to dogs that are in the sent box. '-category:promotions -category:social -category:updates' searches for emails that are not in those categories). Specify this to be '' (empty string) if no filter (just lists all emails). `page`: [string, required] The page ID to get. This only applies if you are getting the next page of results for a previous query. If you want the first page, specify this to be '' (empty string). After calling you first (and subsequent) calls to this tool, you will recieve a next page id, which you can use to call the tool again with the same args but with that page id, to return the next page of results.",
	}
}

func (t *gmailListTool) Call(args map[string]any) (string, error) {
	filterArg, ok := args["filter"]
	if !ok {
		return "", fmt.Errorf("missing required argument: filter")
	}
	filter, ok := filterArg.(string)
	if !ok {
		return "", fmt.Errorf("argument 'filter' must be a string")
	}
	pageArg, ok := args["page"]
	if !ok {
		return "", fmt.Errorf("missing required argument: page")
	}
	page, ok := pageArg.(string)
	if !ok {
		return "", fmt.Errorf("argument 'page' must be a string")
	}

	resp, err := t.client.Emails(email.WithMaxN(5), email.WithPage(page), email.WithQuery(filter))
	if err != nil {
		return "", err
	}

	emails := make([]string, len(resp.Emails))
	for i, email := range resp.Emails {
		recip := make([]string, len(email.To))
		for j, to := range email.To {
			recip[j] = fmt.Sprint(to)
		}
		emails[i] = fmt.Sprintf("- Email with ID %v from %v to %s. Subject = %s", email.ID, email.From, strings.Join(recip, ", "), email.Subject)
	}

	var nextPageMessage string
	if resp.NextPage != "" {
		nextPageMessage = fmt.Sprintf("There are more pages of results. To view the next page, call the tool with the same arguments again, but specify \"page\": \"%s\"", resp.NextPage)
	} else {
		nextPageMessage = "There are no further pages of results to view after this page. You cannot call the tool to get a next page of this query."
	}

	return fmt.Sprintf("This tool response has %d emails:\n\n%s\n\n\n> %s If you would like to see the bodies of any of these emails, please call the gmail_view tool with the email ID.", len(emails), strings.Join(emails, "\n\n"), nextPageMessage), nil
}

type gmailViewTool struct {
	client email.Client
}

func (t *gmailViewTool) Def() cg.ToolDef {
	return cg.ToolDef{
		Name: "gmail_view",
		Desc: "Fetch a single email by ID and return its full content in markdown format, including headers (from, to, subject) and body.",
	}
}

func (t *gmailViewTool) Call(args map[string]any) (string, error) {
	idArg, ok := args["email_id"]
	if !ok {
		return "", fmt.Errorf("missing required argument: email_id")
	}

	id, ok := idArg.(string)
	if !ok {
		return "", fmt.Errorf("argument 'email_id' must be a string")
	}

	email, err := t.client.GetEmail(id)
	if err != nil {
		return "", err
	}

	recip := make([]string, len(email.To))
	for i, to := range email.To {
		recip[i] = fmt.Sprint(to)
	}

	body := email.Content
	mdBody, err := htmlToMarkdown(body)
	if err != nil {
		mdBody = body
	}
	result := fmt.Sprintf(`# 📧 Email

**ID:** %s  
**From:** %v  
**To:** %s  
**Subject:** %s  

---

%s
`, email.ID, email.From, strings.Join(recip, ", "), email.Subject, mdBody)

	return result, nil
}

func htmlToMarkdown(html string) (string, error) {
	converter := markdown.NewConverter("", true, nil)
	md, err := converter.ConvertString(html)
	if err != nil {
		return "", err
	}
	return md, nil
}
