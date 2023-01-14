package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	kzgceremony "github.com/arnaucube/eth-kzg-ceremony-alt"
)

type Client struct {
	url string
	c   *http.Client
}

func (c *Client) postWithAuth(url, contentType string, body io.Reader, bearer string) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Add("Authorization", bearer)
	return c.c.Do(req)
}

func NewClient(sequencerURL string) *Client {
	httpClient := &http.Client{}
	return &Client{
		url: sequencerURL,
		c:   httpClient,
	}
}

func (c *Client) GetCurrentStatus() (*MsgStatus, error) {
	resp, err := c.c.Get(
		c.url + "/info/status")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var msg MsgStatus
	err = json.Unmarshal(body, &msg)
	return &msg, err
}

func (c *Client) GetCurrentState() (*kzgceremony.State, error) {
	resp, err := c.c.Get(
		c.url + "/info/current_state")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var state *kzgceremony.State
	err = json.Unmarshal(body, &state)
	return state, err
}

func (c *Client) GetRequestLink() (*MsgRequestLink, error) {
	resp, err := c.c.Get(
		c.url + "/auth/request_link")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return nil, fmt.Errorf("Invalid request. Missing parameters.")
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("Eth address doesn't match message signer, or account nonce is too low")
		case http.StatusForbidden:
			return nil, fmt.Errorf("Invalid HTTP method")
		default:
			return nil, fmt.Errorf("unexpected http code: %d", resp.StatusCode)
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var msg MsgRequestLink
	err = json.Unmarshal(body, &msg)
	return &msg, err
}

func (c *Client) PostAuthCallback() (*MsgRequestLink, error) {
	resp, err := c.c.Get(
		c.url + "/auth/request_link")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return nil, fmt.Errorf("Invalid request. Missing parameters.")
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("Eth address doesn't match message signer, or account nonce is too low")
		case http.StatusForbidden:
			return nil, fmt.Errorf("Invalid HTTP method")
		default:
			return nil, fmt.Errorf("unexpected http code: %d", resp.StatusCode)
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var msg MsgRequestLink
	err = json.Unmarshal(body, &msg)
	return &msg, err
}

type Status int

const (
	StatusReauth = Status(iota)
	StatusError
	StatusWait
	StatusProceed
)

func (c *Client) PostTryContribute(sessionID string) (*kzgceremony.BatchContribution, Status, error) {
	bearer := "Bearer " + sessionID
	resp, err := c.postWithAuth(
		c.url+"/lobby/try_contribute", "application/json", nil, bearer)
	if err != nil {
		return nil, StatusError, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, StatusError, err
	}

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return nil, StatusWait, fmt.Errorf("call came to early. rate limited")
		case http.StatusUnauthorized:
			return nil, StatusReauth, fmt.Errorf("unkown session id. unauthorized access")
		default:
			return nil, StatusWait, fmt.Errorf("unexpected http code: %d", resp.StatusCode)
		}
	}

	// note: a 200 (Ok) code by the Sequencer on try_contribute doesn't
	// mean that the contributor has been selected. It could mean that the
	// Sequencer is returning the error AnotherContributionInProgress in a
	// json msg (see
	// https://github.com/ethereum/kzg-ceremony-sequencer/blob/2538f2f08d4db880d7f4608e964df0b695bc7d2f/src/api/v1/error_response.rs#L105
	// )

	// check if body contains the error message of "another contribution in
	// progress" (despite http statuscode being 200 (Ok))
	if strings.Contains(string(body), "another contribution in progress") {
		return nil, StatusWait, fmt.Errorf("another contribution in progress")
	}

	err = ioutil.WriteFile("prevBatchContribution.json", body, 0600)
	if err != nil {
		return nil, StatusError, err
	}
	bc := &kzgceremony.BatchContribution{}
	err = json.Unmarshal(body, bc)
	return bc, StatusProceed, err
}

func (c *Client) PostAbortContribution(sessionID string) ([]byte, error) {
	bearer := "Bearer " + sessionID
	resp, err := c.postWithAuth(
		c.url+"/contribution/abort", "application/json", nil, bearer)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println("body", string(body))

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return nil, fmt.Errorf("invalid request")
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("unkown session id. unauthorized access")
		default:
			return nil, fmt.Errorf("unexpected http code: %d", resp.StatusCode)
		}
	}

	return body, nil
}

func (c *Client) PostContribute(sessionID string, bc *kzgceremony.BatchContribution) (*MsgContributeReceipt, error) {
	bearer := "Bearer " + sessionID

	jsonBC, err := json.Marshal(bc)
	if err != nil {
		return nil, err
	}

	resp, err := c.postWithAuth(
		c.url+"/contribute", "application/json", bytes.NewBuffer(jsonBC), bearer)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return nil, fmt.Errorf("invalid request")
		default:
			return nil, fmt.Errorf("unexpected http code: %d", resp.StatusCode)
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var msg MsgContributeReceipt
	err = json.Unmarshal(body, &msg)
	return &msg, err
}
