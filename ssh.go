package sshclient
//TODO rename
// 	sshclient to shh
// 	SshClient to Client
//	import cryptossh "golang.org/x/crypto/ssh"

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"strings"
)

type SshClient struct {
	config    *ssh.ClientConfig
	Client    *ssh.Client // TODO rename to SSHClient? CryptoSSHClient? or embed?
	skipOnErr bool
	err       error
}

func DailWithPassword(ip, username, password string) (*SshClient, error) {
	cfg := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		//TODO HostKeyCallback: ssh.FixedHostKey(hostKey),
		//TODO HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	return dailSSH(ip, cfg)
}

func DailSSHWithKey(ip, username, passphrase string, key []byte) (*SshClient, error) {
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %v", err)
	}

	cfg := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return dailSSH(ip, cfg)
}

func dailSSH(ip string, cfg *ssh.ClientConfig) (*SshClient, error) {
	adrs := ip + ":22"
	cl, err := ssh.Dial("tcp", adrs, cfg)
	if err != nil {
		return nil, err
	}

	return &SshClient{
		config: cfg,
		Client: cl,
	}, nil
}

// SkipOnErr sets how errors are handled;
// when false the caller should handle err != nil as usual,
// when true the client records errors and only performs ssh commands when there are no previous errors.
func (c *SshClient) SkipOnErr(b bool) {
	c.skipOnErr = b
	c.err = nil
}

// Err returns the last error (if any) and clears the internal error state.
func (c *SshClient) Err() error {
	e := c.err
	c.err = nil
	return e
}

// Close the connection with the remote host.
func (c *SshClient) Close() error {
	return c.Client.Close()
}

// Exec executes a cmd on the remote host.
func (c *SshClient) Exec(cmd string, args ...string) (string, error) {
	if c.skipOnErr && c.err != nil {
		// Skip due to previous error
		return "", nil
	}
	// Each client can support multiple interactive sessions, represented by a Session.
	session, err := c.Client.NewSession()
	if err != nil {
		c.err = fmt.Errorf("new session %s %s: %v", cmd, strings.Join(args, " "), err)
		return "", c.err
	}
	defer session.Close()

	l := []string{cmd}
	l = append(l, args...)

	var b bytes.Buffer
	session.Stdout = &b
	err = session.Run(strings.Join(l, " "))
	if err != nil {
		c.err = fmt.Errorf("run %s %s: %v", cmd, strings.Join(args, " "), err)
		return "", c.err
	}

	return b.String(), nil
}

// ScpTo copies data to a remote host.
func (c *SshClient) ScpTo(data []byte, targetPath string, targetPerm int) error {
	if c.skipOnErr && c.err != nil {
		// Skip due to previous error
		return nil
	}
	c.err = ScpTo(c.Client, data, targetPath, targetPerm)
	return c.err
}

// ScpFrom copies data from a remote host.
func (c *SshClient) ScpFrom(sourcePath string) ([]byte, error) {
	if c.skipOnErr && c.err != nil {
		// Skip due to previous error
		return nil, nil
	}
	b,e := ScpFrom(c.Client, sourcePath)
	c.err = e
	return b, c.err
}