/*

Package egoscale is a mapping for with the CloudStack API (http://cloudstack.apache.org/api.html) from Go. It has been designed against the Exoscale (https://www.exoscale.com/) infrastructure but should fit other CloudStack services.

Requests and Responses

To build a request, construct the adequate struct. This library expects a pointer for efficiency reasons only. The response is a struct corresponding to the request itself. E.g. DeployVirtualMachine gives DeployVirtualMachineResponse, as a pointer as well to avoid big copies.

Then everything within the struct is not a pointer. Find below some examples of how egoscale may be used to interact with a CloudStack endpoint, especially Exoscale itself. If anything feels odd or unclear, please let us know: https://github.com/exoscale/egoscale/issues

	req := &egoscale.DeployVirtualMachine{
		Size:              10,
		ServiceOfferingID: "...",
		TemplateID:        "...",
		ZoneID:            "...",
	}

	fmt.Println("Deployment started")
	resp, err := cs.Request(req)
	if err != nil {
		panic(err)
	}

	vm := resp.(*egoscale.DeployVirtualMachineResponse).VirtualMachine
	fmt.Printf("Virtual Machine ID: %s\n", vm.ID)

This exemple deploys a virtual machine while controlling the job status as it goes. It enables a finer control over errors, e.g. HTTP timeout, and eventually a way to kill it of (from the client side).

	req := &egoscale.DeployVirtualMachine{
		Size:              10,
		ServiceOfferingID: "...",
		TemplateID:        "...",
		ZoneID:            "...",
	}
	resp := &egoscale.DeployVirtualMachineResponse{}

	fmt.Println("Deployment started")
	cs.AsyncRequest(req, func(jobResult *egoscale.AsyncJobResult, err error) bool {
		if err != nil {
			// any kind of error
			panic(err)
		}

		// Keep waiting
		if jobResult.JobStatus == egoscale.Pending {
			fmt.Println("wait...")
			return true
		}

		// Unmarshal the response into the response struct
		if err := jobResult.Response(resp); err != nil {
			// JSON unmarshaling error
			panic(err)
		}

		// Stop waiting
		return false
	})

	fmt.Printf("Virtual Machine ID: %s\n", resp.VirtualMachine.ID)


APIs

All the available APIs on the server and provided by the API Discovery plugin

	cs := egoscale.NewClient("https://api.exoscale.ch/compute", "EXO...", "...")

	resp, err := cs.Request(&egoscale.ListAPIs{})
	if err != nil {
		panic(err)
	}

	for _, api := range resp.(*egoscale.ListAPIsResponse).API {
		fmt.Printf("%s %s\n", api.Name, api.Description)
	}
	// Output:
	// listNetworks Lists all available networks
	// ...

Security Groups

Security Groups provide a way to isolate traffic to VMs. Rules are added via the two Authorization commands.

	resp, err := cs.Request(&egoscale.CreateSecurityGroup{
		Name: "Load balancer",
		Description: "Opens HTTP/HTTPS ports from the outside world",
	})
	securityGroup := resp.(*egoscale.CreateSecurityGroupResponse).SecurityGroup

	resp, err = cs.Request(&egoscale.AuthorizeSecurityGroupIngress{
		Description:     "SSH traffic",
		SecurityGroupID: securityGroup.ID,
		CidrList:        []string{"0.0.0.0/0"},
		Protocol:        "tcp",
		StartPort:       22,
		EndPort:         22,
	})
	// The modified SecurityGroup is returned
	securityGroup := resp.(*egoscale.AuthorizeSecurityGroupResponse).SecurityGroup

	// ...
	err = client.BooleanRequest(&egoscale.DeleteSecurityGroup{
		ID: securityGroup.ID,
	})
	// ...

Security Group also implement the generic List, Get and Delete interfaces (Listable, Gettable and Deletable).

	// List all Security Groups
	sgs, err := cs.List(new(egoscale.SecurityGroup))
	for _, s := range sgs {
		sg := s.(egoscale.SecurityGroup)
		// ...
	}

	// Get a Security Group
	sg := &egoscale.SecurityGroup{Name: "Load balancer"}
	if err := cs.Get(sg); err != nil {
		...
	}
	// The SecurityGroup struct has been loaded with the SecurityGroup informations

	if err := cs.Delete(sg); err != nil {
		...
	}
	// The SecurityGroup has been deleted

See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/networking_and_traffic.html#security-groups

Zones

A Zone corresponds to a Data Center. You may list them. Zone implements the Listable interface, which let you perform a list in two different ways. The first exposes the underlying CloudStack request while the second one hide them and you only manipulate the structs of your interest.

	// Using ListZones request
	req := &egoscale.ListZones{}
	resp, err := client.Request(req)
	if err != nil {
		panic(err)
	}

	for _, zone := range resp.(*egoscale.ListZonesResponse) {
		...
	}

	// Using client.List
	zone := &egoscale.Zone{}
	zones, err := client.List(zone)
	if err != nil {
		panic(err)
	}

	for _, z := range zones {
		zone := z.(egoscale.Zone)
		...
	}

Elastic IPs

An Elastic IP is a way to attach an IP address to many Virtual Machines. The API side of the story configures the external environment, like the routing. Some work is required within the machine to properly configure the interfaces.

See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/latest/networking_and_traffic.html#about-elastic-ips

*/
package egoscale
