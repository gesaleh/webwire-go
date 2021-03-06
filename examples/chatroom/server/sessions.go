package main

import (
	"log"

	webwire "github.com/qbeon/webwire-go"
)

var connectedClients = make(map[*webwire.Client]bool)
var sessions = make(map[string]*webwire.Client)

func onClientConnected(newClient *webwire.Client) {
	connectedClients[newClient] = true
	log.Printf("New client connected: %s | %s", newClient.RemoteAddr(), newClient.UserAgent())
}

func onClientDisconnected(client *webwire.Client) {
	delete(connectedClients, client)
	log.Printf("Client %s disconnected", client.RemoteAddr())
}

func onSessionCreated(client *webwire.Client) error {
	sessions[client.Session.Key] = client
	return nil
}

func onSessionLookup(key string) (*webwire.Session, error) {
	if clt, ok := sessions[key]; ok {
		return clt.Session, nil
	}
	return nil, nil
}

func onSessionClosed(client *webwire.Client) error {
	delete(sessions, client.Session.Key)
	log.Printf("Client %s closed the session", client.RemoteAddr())
	return nil
}
