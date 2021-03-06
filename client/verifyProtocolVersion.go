package client

// verifyProtocolVersion requests the endpoint metadata
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// to verify the server is running a supported protocol version
func (clt *Client) verifyProtocolVersion() error {
	// Initialize HTTP client
	var httpClient = &http.Client{
		Timeout: time.Second * 10,
	}

	request, err := http.NewRequest(
		"WEBWIRE", "http://"+clt.serverAddr+"/", nil,
	)
	if err != nil {
		return fmt.Errorf("Couldn't create HTTP metadata request: %s", err)
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("Endpoint metadata request failed: %s", err)
	}

	// Read response body
	defer response.Body.Close()
	encodedData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Couldn't read metadata response body: %s", err)
	}

	// Unmarshal response
	var metadata struct {
		ProtocolVersion string `json:"protocol-version"`
	}
	if err := json.Unmarshal(encodedData, &metadata); err != nil {
		return fmt.Errorf(
			"Couldn't parse HTTP metadata response ('%s'): %s",
			string(encodedData),
			err,
		)
	}

	// Verify metadata
	if metadata.ProtocolVersion != supportedProtocolVersion {
		return fmt.Errorf(
			"Unsupported protocol version: %s (%s is supported by this client)",
			metadata.ProtocolVersion,
			supportedProtocolVersion,
		)
	}

	return nil
}
