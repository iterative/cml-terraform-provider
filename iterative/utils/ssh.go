package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func PrivatePEM() (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", err
	}

	var privKeyBuf strings.Builder
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := pem.Encode(&privKeyBuf, privateKeyPEM); err != nil {
		return "", err
	}

	return privKeyBuf.String(), nil
}

func PublicFromPrivatePEM(privateKey string) (string, error) {
	block, rest := pem.Decode([]byte(privateKey))
	if block == nil {
		return "", fmt.Errorf("Invalid PEM on the SSH private key: %#v", rest)
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	pub, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		return "", err
	}

	var pubKeyBuf strings.Builder
	pubKeyBuf.Write(ssh.MarshalAuthorizedKey(pub))

	return pubKeyBuf.String(), nil
}

func RunCommand(command string, timeout time.Duration, hostAddress string, userName string, privateKey string) (string, error) {
	parsedPrivateKey, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return "", err
	}

	configuration := &ssh.ClientConfig{
		User: userName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(parsedPrivateKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Doesn't matter for this use case, but isn't a good practice either.
		Timeout:         timeout,
	}

	client, err := ssh.Dial("tcp", hostAddress, configuration)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
