package main

// Manifest schema
type Manifest struct {
	Description string            `json:"description"`
	Events      []string          `json:"default_events,omitempty"`
	Name        string            `json:"name"`
	Permissions map[string]string `json:"default_permissions,omitempty"`
	Public      bool              `json:"public"`
	RedirectURL string            `json:"redirect_url"`
	URL         string            `json:"url"`
	Webhook     GithubWebhook     `json:"hook_attributes"`
}

type GithubWebhook struct {
	Active bool   `json:"active"`
	URL    string `json:"url"`
}

func newManifest(nameSuffix, redirectURL, webhookURL string) Manifest {
	return Manifest{
		Name: "Manifestor-" + nameSuffix,
		URL:  "https://github.com/leg100/manifestor",
		Webhook: GithubWebhook{
			URL: webhookURL,
		},
		RedirectURL: redirectURL,
	}
}
