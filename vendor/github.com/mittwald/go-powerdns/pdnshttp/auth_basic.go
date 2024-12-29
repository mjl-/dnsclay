package pdnshttp

import "net/http"

type BasicAuthenticator struct {
	Username string
	Password string
}

func (a *BasicAuthenticator) OnRequest(r *http.Request) error {
	r.SetBasicAuth(a.Username, a.Password)
	return nil
}

func (a *BasicAuthenticator) OnConnect(*http.Client) error {
	return nil
}
