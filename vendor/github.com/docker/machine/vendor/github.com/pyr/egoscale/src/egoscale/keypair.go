package egoscale

import (
	"encoding/json"
	"net/url"
)

func (exo *Client) CreateKeypair(name string) (*CreateSSHKeyPairResponse, error) {
	params := url.Values{}
	params.Set("name", name)

	resp, err := exo.Request("createSSHKeyPair", params)
	if err != nil {
		return nil, err
	}

	var r CreateSSHKeyPairWrappedResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r.Wrapped, nil
}

func (exo *Client) DeleteKeypair(name string) (*StandardResponse, error) {
	params := url.Values{}
	params.Set("name", name)

	resp, err := exo.Request("deleteSSHKeyPair", params)
	if err != nil {
		return nil, err
	}

	var r StandardResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil

}
