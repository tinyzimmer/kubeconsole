package server

import (
	"crypto/rand"
	"crypto/rsa"
	"io/ioutil"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

const bitSize = 2048

type Server interface {
	Listen() error
}

type server struct {
	Server
	key       ssh.Signer
	incluster bool
}

func New(incluster bool, serverkey string) (Server, error) {
	s := &server{incluster: incluster}
	var key ssh.Signer
	var err error
	if serverkey == "" {
		key, err = generateKey()
	} else {
		key, err = loadKey(serverkey)
	}
	if err != nil {
		return nil, err
	}
	s.key = key
	return s, nil
}

func (s *server) Listen() (err error) {
	config := &ssh.ServerConfig{}
	config.AddHostKey(s.key)
	config.NoClientAuth = true

	log.Println("Listening for channels on :2022")
	ln, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		return
	}
	for {
		nConn, err := ln.Accept()
		if err != nil {
			log.Println("Failed to accept connection from client:", err)
			continue
		}
		conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
		if err != nil {
			log.Println("failed to handshake: ", err)
			continue
		}
		log.Printf("New SSH connection from %s (%s)", conn.RemoteAddr(), conn.ClientVersion())
		go ssh.DiscardRequests(reqs)

		for newChannel := range chans {
			go s.handleChannel(newChannel)
		}
	}
}

func generateKey() (signer ssh.Signer, err error) {
	reader := rand.Reader
	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return
	}
	signer, err = ssh.NewSignerFromKey(key)
	return
}

func loadKey(keyfile string) (key ssh.Signer, err error) {
	privateBytes, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return
	}
	key, err = ssh.ParsePrivateKey(privateBytes)
	return
}
