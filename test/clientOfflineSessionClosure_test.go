package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	webwire "github.com/qbeon/webwire-go"
	webwireClient "github.com/qbeon/webwire-go/client"
)

// TestClientOfflineSessionClosure verifies session restoration
func TestClientOfflineSessionClosure(t *testing.T) {
	sessionStorage := make(map[string]*webwire.Session)

	currentStep := 1
	var createdSession *webwire.Session

	// Initialize webwire server
	server := setupServer(
		t,
		webwire.Hooks{
			OnRequest: func(ctx context.Context) ([]byte, *webwire.Error) {
				// Extract request message and requesting client from the context
				msg := ctx.Value(webwire.MESSAGE).(webwire.Message)

				if currentStep == 2 {
					// Expect the session to have been automatically restored
					if msg.Client.Session != nil {
						t.Errorf("Expected client to be anonymous")
					}
					return nil, nil
				}

				// Try to create a new session
				if err := msg.Client.CreateSession(nil); err != nil {
					return nil, &webwire.Error{
						Code:    "INTERNAL_ERROR",
						Message: fmt.Sprintf("Internal server error: %s", err),
					}
				}

				// Return the key of the newly created session
				return nil, nil
			},
			// Permanently store the session
			OnSessionCreated: func(client *webwire.Client) error {
				sessionStorage[client.Session.Key] = client.Session
				return nil
			},
			// Find session by key
			OnSessionLookup: func(key string) (*webwire.Session, error) {
				// Expect the key of the created session to be looked up
				if key != createdSession.Key {
					err := fmt.Errorf(
						"Expected and looked up session keys differ: %s | %s",
						createdSession.Key,
						key,
					)
					t.Fatalf("Session lookup mismatch: %s", err)
					return nil, err
				}

				if session, exists := sessionStorage[key]; exists {
					return session, nil
				}

				// Expect the session to be found
				t.Fatalf(
					"Expected session (%s) not found in: %v",
					createdSession.Key,
					sessionStorage,
				)
				return nil, nil
			},
			// Define dummy hook to enable sessions on this server
			OnSessionClosed: func(_ *webwire.Client) error { return nil },
		},
	)
	go server.Run()

	// Initialize client
	client := webwireClient.NewClient(
		server.Addr,
		webwireClient.Hooks{},
		5*time.Second,
		os.Stdout,
		os.Stderr,
	)

	if err := client.Connect(); err != nil {
		t.Fatalf("Couldn't connect: %s", err)
	}

	/*****************************************************************\
		Step 1 - Create session and disconnect
	\*****************************************************************/

	// Create a new session
	if _, err := client.Request([]byte("auth")); err != nil {
		t.Fatalf("Auth request failed: %s", err)
	}

	tmp := client.Session()
	createdSession = &tmp

	// Disconnect client without closing the session
	client.Close()

	// Ensure the session isn't lost
	if client.IsConnected() {
		t.Fatal("Client is expected to be disconnected")
	}
	if client.Session().Key == "" {
		t.Fatal("Session lost after disconnection")
	}

	/*****************************************************************\
		Step 2 - Close session, reconnect and verify
	\*****************************************************************/
	currentStep = 2

	if err := client.CloseSession(); err != nil {
		t.Fatal("Offline session closure failed: %s", err)
	}

	// Ensure the session is removed
	if client.Session().Key != "" {
		t.Fatal("Session not removed")
	}

	// Reconnect
	if err := client.Connect(); err != nil {
		t.Fatalf("Couldn't reconnect: %s", err)
	}

	// Verify whether the previous session was restored automatically
	// and the server authenticates the user
	if _, err := client.Request([]byte("isrestored?")); err != nil {
		t.Fatalf("Second request failed: %s", err)
	}
}
