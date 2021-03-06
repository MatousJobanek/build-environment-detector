package git

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"

	"golang.org/x/crypto/ssh"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

type Secret interface {
	GitAuthMethod() (transport.AuthMethod, error)
	Client() *http.Client
	SecretType() string
	Content() string
}

type commonSecretInfo struct {
	secretType    string
	secretContent []byte
}

func (k *commonSecretInfo) SecretType() string {
	return k.secretType
}

func (k *commonSecretInfo) Content() string {
	return string(k.secretContent)
}

const (
	SshKeyType           = "SshKey"
	OauthTokenType       = "OauthToken"
	UsernamePasswordType = "UsernamePassword"
)

type SshKey struct {
	*commonSecretInfo
	passphrase []byte
}

func NewSshKey(sshKey []byte, passphrase []byte) *SshKey {
	return &SshKey{
		commonSecretInfo: &commonSecretInfo{
			secretType:    SshKeyType,
			secretContent: bytes.TrimSpace(sshKey),
		},
		passphrase: passphrase,
	}
}

func (k *SshKey) GitAuthMethod() (transport.AuthMethod, error) {
	var signer ssh.Signer
	var err error
	if len(k.passphrase) > 0 {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(k.secretContent, k.passphrase)
		if err != nil {
			return nil, err
		}
	} else {
		signer, err = ssh.ParsePrivateKey(k.secretContent)
		if err != nil {
			return nil, err
		}
	}

	return &gitssh.PublicKeys{User: "git", Signer: signer}, nil
}

func (k *SshKey) Client() *http.Client {
	return nil
}

type OauthToken struct {
	*commonSecretInfo
}

func NewOauthToken(token []byte) *OauthToken {
	return &OauthToken{
		commonSecretInfo: &commonSecretInfo{
			secretType:    OauthTokenType,
			secretContent: bytes.TrimSpace(token),
		},
	}
}

func (t *OauthToken) GitAuthMethod() (transport.AuthMethod, error) {
	return &githttp.BasicAuth{Password: string(t.secretContent)}, nil
}

func (t *OauthToken) Client() *http.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: string(t.secretContent)},
	)
	return oauth2.NewClient(ctx, ts)

}

type UsernamePassword struct {
	*commonSecretInfo
	username string
	password string
}

func NewUsernamePassword(username, password string) *UsernamePassword {
	return &UsernamePassword{
		commonSecretInfo: &commonSecretInfo{
			secretType:    UsernamePasswordType,
			secretContent: []byte(fmt.Sprintf("%s:%s", username, password)),
		},
		username: username,
		password: password,
	}
}

func (t *UsernamePassword) GitAuthMethod() (transport.AuthMethod, error) {
	return &githttp.BasicAuth{Username: t.username, Password: t.password}, nil
}

func (t *UsernamePassword) Client() *http.Client {
	return &http.Client{}
}

func ParseUsernameAndPassword(secret string) (string, string) {
	split := strings.Split(secret, ":")
	if len(split) == 2 {
		return split[0], split[1]
	}
	return "", ""
}
