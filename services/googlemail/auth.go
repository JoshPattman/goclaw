package googlemail

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Read credentials JSON file - you can get this from google cloud dashboard.
func readConfigFromFile(loc string, redirectPort int) (*oauth2.Config, error) {
	credentials, err := os.ReadFile(loc)
	if err != nil {
		return nil, errors.Join(errors.New("failed to read credentials file"), err)
	}
	config, err := google.ConfigFromJSON(credentials, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, errors.Join(errors.New("failed to parse credentials file"), err)
	}
	config.RedirectURL = fmt.Sprintf("http://localhost:%d/authentication-complete", redirectPort)
	return config, nil
}

// Read the token file. If the file does not exist.
func readTokenFromFile(loc string) (*oauth2.Token, error) {
	f, err := os.Open(loc)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func getAuthURL(config *oauth2.Config) string {
	return config.AuthCodeURL(
		"state-token",
		oauth2.AccessTypeOffline,
	)
}

func listenForAuth(config *oauth2.Config, port int) (*oauth2.Token, error) {
	tokenCh := make(chan *oauth2.Token)
	errCh := make(chan error)

	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	mux.HandleFunc("/authentication-complete", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing code", http.StatusBadRequest)
			errCh <- fmt.Errorf("missing code in callback")
			return
		}

		token, err := config.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
			errCh <- err
			return
		}

		fmt.Fprintln(w, "Authentication complete! You can close this tab.")
		tokenCh <- token
	})

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for result
	select {
	case tok := <-tokenCh:
		_ = server.Shutdown(context.Background())
		return tok, nil
	case err := <-errCh:
		_ = server.Shutdown(context.Background())
		return nil, err
	}
}

func saveTokenToFile(token *oauth2.Token, loc string) error {
	f, err := os.OpenFile(loc, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func buildService(config *oauth2.Config, token *oauth2.Token) (*gmail.Service, error) {
	client := config.Client(context.Background(), token)
	return gmail.NewService(context.Background(), option.WithHTTPClient(client))
}
