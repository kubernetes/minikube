package softlayer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Client struct {
	User     string
	ApiKey   string
	Endpoint string
}

type HostSpec struct {
	Hostname                       string            `json:"hostname"`
	Domain                         string            `json:"domain"`
	Cpu                            int               `json:"startCpus"`
	Memory                         int               `json:"maxMemory"`
	Datacenter                     Datacenter        `json:"datacenter"`
	SshKeys                        []*SSHKey         `json:"sshKeys"`
	BlockDevices                   []BlockDevice     `json:"blockDevices"`
	NetworkMaxSpeeds               []NetworkMaxSpeed `json:"networkComponents"`
	InstallScript                  string            `json:"postInstallScriptUri"`
	PrivateNetOnly                 bool              `json:"privateNetworkOnlyFlag"`
	Os                             string            `json:"operatingSystemReferenceCode"`
	HourlyBilling                  bool              `json:"hourlyBillingFlag"`
	LocalDisk                      bool              `json:"localDiskFlag"`
	PrimaryNetworkComponent        *NetworkComponent `json:"primaryNetworkComponent,omitempty"`
	PrimaryBackendNetworkComponent *NetworkComponent `json:"primaryBackendNetworkComponent,omitempty"`
}

type NetworkMaxSpeed struct {
	MaxSpeed int `json:"maxSpeed"`
}

type NetworkComponent struct {
	NetworkVLAN *NetworkVLAN `json:"networkVlan"`
}

type NetworkVLAN struct {
	Id int `json:"id"`
}

type SSHKey struct {
	Key   string `json:"key,omitempty"`
	Id    int    `json:"id,omitempty"`
	Label string `json:"label,omitempty"`
}

type BlockDevice struct {
	Device    string    `json:"device"`
	DiskImage DiskImage `json:"diskImage"`
}

type DiskImage struct {
	Capacity int `json:"capacity"`
}

type Datacenter struct {
	Name string `json:"name"`
}

type sshKey struct {
	*Client
}

type VirtualGuest struct {
	*Client
}

func NewClient(user, key, endpoint string) *Client {
	return &Client{User: user, ApiKey: key, Endpoint: endpoint}
}

func (c *Client) isOkStatus(code int) bool {
	codes := map[int]bool{
		200: true,
		201: true,
		204: true,
		400: false,
		404: false,
		500: false,
		409: false,
		406: false,
	}

	return codes[code]
}

func (c *Client) newRequest(method, uri string, body interface{}) ([]byte, error) {
	var (
		client = &http.Client{}
		url    = fmt.Sprintf("%s/%s", c.Endpoint, uri)
		err    error
		req    *http.Request
	)

	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(bodyJSON))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("Error with request: %v - %q", url, err)
	}

	req.SetBasicAuth(c.User, c.ApiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if !c.isOkStatus(resp.StatusCode) {
		type apiErr struct {
			Err string `json:"error"`
		}
		var outErr apiErr
		json.Unmarshal(data, &outErr)
		return nil, fmt.Errorf("Error in response: %s", outErr.Err)
	}
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Client) SSHKey() *sshKey {
	return &sshKey{c}
}

func (c *sshKey) namespace() string {
	return "SoftLayer_Security_Ssh_Key"
}

func (c *sshKey) Create(label, key string) (*SSHKey, error) {
	var (
		method = "POST"
		uri    = c.namespace()
		body   = SSHKey{Key: key, Label: label}
	)

	data, err := c.newRequest(method, uri, map[string]interface{}{"parameters": []interface{}{body}})
	if err != nil {
		return nil, err
	}

	var k SSHKey
	if err := json.Unmarshal(data, &k); err != nil {
		return nil, err
	}

	return &k, nil
}

func (c *sshKey) Delete(id int) error {
	var (
		method = "DELETE"
		uri    = fmt.Sprintf("%s/%v", c.namespace(), id)
	)

	_, err := c.newRequest(method, uri, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) VirtualGuest() *VirtualGuest {
	return &VirtualGuest{c}
}

func (c *VirtualGuest) namespace() string {
	return "SoftLayer_Virtual_Guest"
}

func (c *VirtualGuest) PowerState(id int) (string, error) {
	type state struct {
		KeyName string `json:"keyName"`
		Name    string `json:"name"`
	}
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%v/getPowerState.json", c.namespace(), id)
	)

	data, err := c.newRequest(method, uri, nil)
	if err != nil {
		return "", err
	}
	var s state
	if err := json.Unmarshal(data, &s); err != nil {
		return "", err
	}

	return s.Name, nil
}

func (c *VirtualGuest) ActiveTransaction(id int) (string, error) {
	type transactionStatus struct {
		AverageDuration string `json:"averageDuration"`
		FriendlyName    string `json:"friendlyName"`
		Name            string `json:"name"`
	}
	type transaction struct {
		CreateDate        string            `json:"createDate"`
		ElapsedSeconds    int               `json:"elapsedSeconds"`
		GuestID           int               `json:"guestId"`
		HardwareID        int               `json:"hardwareId"`
		ID                int               `json:"id"`
		ModifyDate        string            `json:"modifyDate"`
		StatusChangeDate  string            `json:"statusChangeDate"`
		TransactionStatus transactionStatus `json:"transactionStatus"`
	}
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%v/getActiveTransaction.json", c.namespace(), id)
	)

	data, err := c.newRequest(method, uri, nil)
	if err != nil {
		return "", err
	}
	var t transaction
	if err := json.Unmarshal(data, &t); err != nil {
		return "", err
	}

	return t.TransactionStatus.Name, nil
}

func (c *VirtualGuest) Create(spec *HostSpec) (int, error) {
	var (
		method = "POST"
		uri    = c.namespace() + ".json"
	)

	data, err := c.newRequest(method, uri, map[string]interface{}{"parameters": []interface{}{spec}})
	if err != nil {
		return -1, err
	}

	type createResp struct {
		ID int `json:"id"`
	}

	var r createResp
	if err := json.Unmarshal(data, &r); err != nil {
		return -1, err
	}

	return r.ID, nil
}

func (c *VirtualGuest) Cancel(id int) error {
	var (
		method = "DELETE"
		uri    = fmt.Sprintf("%s/%v", c.namespace(), id)
	)

	_, err := c.newRequest(method, uri, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *VirtualGuest) PowerOn(id int) error {
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%v/powerOn.json", c.namespace(), id)
	)

	_, err := c.newRequest(method, uri, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *VirtualGuest) PowerOff(id int) error {
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%v/powerOff.json", c.namespace(), id)
	)

	_, err := c.newRequest(method, uri, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *VirtualGuest) Pause(id int) error {
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%v/pause.json", c.namespace(), id)
	)

	_, err := c.newRequest(method, uri, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *VirtualGuest) Resume(id int) error {
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%v/resume.json", c.namespace(), id)
	)

	_, err := c.newRequest(method, uri, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *VirtualGuest) Reboot(id int) error {
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%v/rebootSoft.json", c.namespace(), id)
	)

	_, err := c.newRequest(method, uri, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *VirtualGuest) GetPublicIP(id int) (string, error) {
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%v/getPrimaryIpAddress.json", c.namespace(), id)
	)

	data, err := c.newRequest(method, uri, nil)
	if err != nil {
		return "", err
	}
	return strings.Replace(string(data), "\"", "", -1), nil
}

func (c *VirtualGuest) GetPrivateIP(id int) (string, error) {
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%v/getPrimaryBackendIpAddress.json", c.namespace(), id)
	)

	data, err := c.newRequest(method, uri, nil)
	if err != nil {
		return "", err
	}
	return strings.Replace(string(data), "\"", "", -1), nil
}
