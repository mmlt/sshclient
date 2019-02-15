package sshclient

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"path"
	"time"
)

// SCP
// http://web.archive.org/web/20170215184048/https://blogs.oracle.com/janp/entry/how_the_scp_protocol_works
//

func ScpTo(client *ssh.Client, data []byte, targetPath string, targetPerm int) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %v", err)
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("connect stdin: %v", err)
	}

	rerr := make(chan error, 2)

	go func() {
		_, err = fmt.Fprintf(stdin, "C%04o %d %s\n%s\x00", targetPerm, len(data), path.Base(targetPath), string(data))
		rerr <- err
		stdin.Close() // tells scp to exit
	}()

	go func() {
		err = session.Run("/usr/bin/scp -qt " + path.Dir(targetPath))
		rerr <- err
	}()

	for i := 0; i < 2; i++ {
		select {
		case <-time.After(10*time.Second): //TODO fix hardcoded timeout
			err = fmt.Errorf("timeout")
			break
		case err = <-rerr:
			if err != nil {
				break
			}
		}
	}

	return err
}

func ScpFrom(client *ssh.Client, sourcePath string) ([]byte, error) {
	session, err := client.NewSession()
	if err != nil {
		return []byte{}, fmt.Errorf("new session: %v", err)
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return []byte{}, fmt.Errorf("connect stdin: %v", err)
	}
	defer stdin.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return []byte{}, fmt.Errorf("connect stdout: %v", err)
	}

	r := make(chan []byte)
	rerr := make(chan error, 2)

	go func() {
		ack := []byte{0}
		// Write 00 to initiate transfer
		stdin.Write(ack)
		// Read C<perm> <size> <filename>\n
		var perm, size int
		var filename string
		fmt.Fscanf(stdout, "C%04o %d %s", &perm, &size, &filename)
		stdin.Write(ack)
		// Read <data>
		data := make([]byte, size)
		_, err := io.ReadFull(stdout, data)
		stdin.Write(ack)

		r <- data
		rerr <- err
	}()

	go func() {
		rerr <- session.Run("/usr/bin/scp -qf " + sourcePath)
	}()

	var result []byte

	for i := 0; i < 2; i++ {
		select {
		case <-time.After(10 * time.Second): //TODO fix hardcoded timeout
			err = fmt.Errorf("timeout")
			break
		case err = <-rerr:
			if err != nil {
				break
			}
		case result = <-r:
			break
		}
	}
	return result, err
}


type SourceItem struct {
	filename string
	perm int
	size int
	//reader *io.Reader
}

type SourceFn func() (*SourceItem, error)

func ScpToRecursive(client *ssh.Client, targetDir string, targetDirPerm int, source SourceFn) error { //TODO implement
	return nil
}

type SinkItem struct {
	filename string
	perm int
	size int
	//writer *io.Writer
}

type SinkFn func() (*SourceItem, error)

func ScpFromRecursive(client *ssh.Client, sourcePath string, sink SinkFn) error { //TODO implement
	return nil
}

