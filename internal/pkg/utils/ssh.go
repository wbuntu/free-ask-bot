package utils

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"strconv"
	"strings"
)

// ExeSshCmd runs command inside a VM, via SSH, and returns the command output.
func ExeSshCmd(ip string, port int, password, command string) (string, error) {
	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	hostAddress := strings.Join([]string{ip, strconv.Itoa(port)}, ":")

	connection, err := ssh.Dial("tcp", hostAddress, sshConfig)
	if err != nil {
		return "", fmt.Errorf("ssh: failed to dial IP %s, error: %s", hostAddress, err.Error())
	}

	session, err := connection.NewSession()
	if err != nil {
		return "", fmt.Errorf("ssh: failed to create session, error: %s", err.Error())
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		return "", fmt.Errorf("ssh: failed to run command `%s`, error: %s", command, err.Error())
	}

	output := strings.Trim(b.String(), "\n")

	return output, nil
}
