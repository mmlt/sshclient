package sshclient_test

import (
	"fmt"
	"testing"
	"github.com/mmlt/sshclient"
)

var (
	ip string = "10.10.10.10"
	user string = "xxx"
	pass string = "yyy"
)

func TestSCPTo(t *testing.T) {
	cl, err := sshclient.DailWithPassword(ip, user, pass)
	if err != nil {
		fmt.Printf("failed to dail: %v\n", err)
	}
	defer cl.Close()

	err = sshclient.scpTo(cl.Client, []byte("this is a test3"), "/var/tmp/test03", 0600)
	if err != nil {
		fmt.Printf("err %v\n", err)
	}
}

func TestSCPFrom(t *testing.T) {
	cl, err := sshclient.DailWithPassword(ip, user, pass)
	if err != nil {
		fmt.Printf("failed to dail: %v\n", err)
	}
	defer cl.Close()

	data, err := sshclient.ScpFrom(cl.Client, "/var/tmp/test01")
	if err != nil {
		fmt.Printf("err %v\n", err)
	}
	fmt.Println(string(data))
}
