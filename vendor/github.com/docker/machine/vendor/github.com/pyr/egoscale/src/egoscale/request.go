package egoscale

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
)

func rawValue(b json.RawMessage) (json.RawMessage, error) {
	var m map[string]json.RawMessage

	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	for _, v := range m {
		return v, nil
	}
	return nil, fmt.Errorf("Unable to extract raw value from:\n\n%s\n\n", string(b))
}

func (exo *Client) Request(command string, params url.Values) (json.RawMessage, error) {

	mac := hmac.New(sha1.New, []byte(exo.apiSecret))

	params.Set("apikey", exo.apiKey)
	params.Set("command", command)
	params.Set("response", "json")

	s := strings.Replace(strings.ToLower(params.Encode()), "+", "%20", -1)
	mac.Write([]byte(s))
	signature := url.QueryEscape(base64.StdEncoding.EncodeToString(mac.Sum(nil)))

	s = params.Encode()
	url := exo.endpoint + "?" + s + "&signature=" + signature

	resp, err := exo.client.Get(url)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	b, err = rawValue(b)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		var e Error
		if err := json.Unmarshal(b, &e); err != nil {
			return nil, err
		}
		return nil, e.Error()
	}
	return b, nil
}
