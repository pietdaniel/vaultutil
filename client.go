package vaultutil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

// Constants (var for unit-testing)
var (
	KubernetesServiceAccountTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

// InClusterClient returns a vault API client using environment variables passed
// to pods within a kubernetes cluster
func InClusterClient() (*api.Client, error) {
	// make sure environment variables exist
	for _, name := range []string{
		"VAULT_ADDR",
		"VAULT_AUTH_PATH",
		"VAULT_ROLE",
	} {
		if os.Getenv(name) == "" {
			return nil, errors.Errorf("missing environment variable `%s`", name)
		}
	}

	bs, err := ioutil.ReadFile(KubernetesServiceAccountTokenFile)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading kubernetes service account token")
	}

	payload := fmt.Sprintf(`{"jwt": "%s", "role": "%s"}`,
		string(bs), os.Getenv("VAULT_ROLE"))

	u := fmt.Sprintf("%s/v1/auth/%s/login",
		os.Getenv("VAULT_ADDR"), os.Getenv("VAULT_AUTH_PATH"))
	res, err := http.Post(u, "application/json", strings.NewReader(payload))
	if err != nil {
		return nil, errors.Wrapf(err, "error retrieving vault auth token")
	}
	defer res.Body.Close()
	defer ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		bs, _ = ioutil.ReadAll(res.Body)
		return nil, errors.Errorf("error retrieving vault auth token: %d %s %s",
			res.StatusCode,
			res.Status,
			string(bs))
	}

	// { "auth": {
	//    "client_token": "abcd1234"
	// }}
	var authresult struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
	}
	err = json.NewDecoder(res.Body).Decode(&authresult)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading client token")
	}

	// Create a client using DefaultConfig which already supports the VAULT_ADDR
	// environment variable
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, errors.Wrapf(err, "error retrieving vault client")
	}
	client.SetToken(authresult.Auth.ClientToken)
	return client, nil
}

// InClusterSecret is a helper function to retrieve a secret from vault from
// within kubernetes
func InClusterSecret(path, field string) (string, error) {
	client, err := InClusterClient()
	if err != nil {
		return "", err
	}

	secret, err := client.Logical().Read(path)
	if err != nil {
		return "", err
	}
	if secret == nil || secret.Data == nil {
		return "", errors.Errorf("field `%s` not found in secret `%s`",
			field, path)
	}

	obj, ok := secret.Data[field]
	if !ok {
		return "", errors.Errorf("field `%s` not found in secret `%s`",
			field, path)
	}
	return fmt.Sprint(obj), nil
}
