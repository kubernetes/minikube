Changelog
=========

0.9.23
------

- feat: `booleanResponse` supports true booleans: https://github.com/apache/cloudstack/pull/2428

0.9.22
------

- feat: `ListUsers`, `CreateUser`, `UpdateUser`
- feat: `ListResourceDetails`
- feat: `SecurityGroup` helper `RuleByID`
- feat: `Sign` signs the payload
- feat: `UpdateNetworkOffering`
- feat: `GetVirtualMachineUserData`
- feat: `EnableAccount` and `DisableAccount` (admin stuff)
- feat: `AsyncRequest` and `AsyncRequestWithContext` to examine the polling
- fix: `AuthorizeSecurityGroupIngress` support for ICMPv6
- change: move `APIName()` into the `Client`, nice godoc
- change: `Payload` doesn't sign the request anymore
- change: `Client` exposes more of its underlying data
- change: requests are sent as GET unless it body size is too big

0.9.21
------

- feat: `Network` is `Listable`
- feat: `Zone` is `Gettable`
- feat: `Client.Payload` to help preview the HTTP parameters
- feat: generate command utility
- fix: `CreateSnapshot` was missing the `Name` attribute
- fix: `ListSnapshots` was missing the `IDs` attribute
- fix: `ListZones` was missing the `NetworkType` attribute
- fix: `ListAsyncJobs` was missing the `ListAll` attribute
- change: ICMP Type/Code are uint8 and TCP/UDP port are uint16

0.9.20
------

- feat: `Template` is `Listable`
- feat: `IPAddress` is `Listable`
- change: `List` and `Paginate` return pointers
- fix: `Template` was missing `tags`

0.9.19
------

- feat: `SSHKeyPair` is `Listable`

0.9.18
------

- feat: `VirtualMachine` is `Listable`
- feat: new `Client.Paginate` and `Client.PaginateWithContext`
- change: the inner logic of `Listable`
- remove: not working `Client.AsyncList`

0.9.17
------

- fix: `AuthorizeSecurityGroup(In|E)gress` startport may be zero

0.9.16
------

- feat: new `Listable` interface
- feat: `Nic` is `Listable`
- feat: `Volume` is `Listable`
- feat: `Zone` is `Listable`
- feat: `AffinityGroup` is `Listable`
- remove: deprecated methods `ListNics`, `AddIPToNic`, and `RemoveIPFromNic`
- remove: deprecated method `GetRootVolumeForVirtualMachine`

0.9.15
------

- feat: `IPAddress` is `Gettable` and `Deletable`
- fix: serialization of *bool

0.9.14
------

- fix: `GetVMPassword` response
- remove: deprecated `GetTopology`, `GetImages`, and al

0.9.13
------

- feat: IP4 and IP6 flags to DeployVirtualMachine
- feat: add ActivateIP6
- fix: error message was gobbled on 40x

0.9.12
------

- feat: add `BooleanRequestWithContext`
- feat: add `client.Get`, `client.GetWithContext` to fetch a resource
- feat: add `cleint.Delete`, `client.DeleteWithContext` to delete a resource
- feat: `SSHKeyPair` is `Gettable` and `Deletable`
- feat: `VirtualMachine` is `Gettable` and `Deletable`
- feat: `AffinityGroup` is `Gettable` and `Deletable`
- feat: `SecurityGroup` is `Gettable` and `Deletable`
- remove: deprecated methods `CreateAffinityGroup`, `DeleteAffinityGroup`
- remove: deprecated methods `CreateKeypair`, `DeleteKeypair`, `RegisterKeypair`
- remove: deprecated method `GetSecurityGroupID`

0.9.11
------

- feat: CloudStack API name is now public `APIName()`
- feat: enforce the mutual exclusivity of some fields
- feat: add `context.Context` to `RequestWithContext`
- change: `AsyncRequest` and `BooleanAsyncRequest` are gone, use `Request` and `BooleanRequest` instead.
- change: `AsyncInfo` is no more

0.9.10
------

- fix: typo made ListAll required in ListPublicIPAddresses
- fix: all bool are now *bool, respecting CS default value
- feat: (*VM).DefaultNic() to obtain the main Nic

0.9.9
-----

- fix: affinity groups virtualmachineIds attribute
- fix: uuidList is not a list of strings

0.9.8
-----

- feat: add RootDiskSize to RestoreVirtualMachine
- fix: monotonic polling using Context

0.9.7
-----

- feat: add Taggable interface to expose ResourceType
- feat: add (Create|Update|Delete|List)InstanceGroup(s)
- feat: add RegisterUserKeys
- feat: add ListResourceLimits
- feat: add ListAccounts

0.9.6
-----

- fix: update UpdateVirtualMachine userdata
- fix: Network's name/displaytext might be empty

0.9.5
-----

- fix: serialization of slice

0.9.4
-----

- fix: constants

0.9.3
-----

- change: userdata expects a string
- change: no pointer in sub-struct's

0.9.2
-----

- bug: createNetwork is a sync call
- bug: typo in listVirtualMachines' domainid
- bug: serialization of map[string], e.g. UpdateVirtualMachine
- change: IPAddress's use net.IP type
- feat: helpers VM.NicsByType, VM.NicByNetworkID, VM.NicByID
- feat: addition of CloudStack ApiErrorCode constants

0.9.1
-----

- bug: sync calls returns succes as a string rather than a bool
- change: unexport BooleanResponse types
- feat: original CloudStack error response can be obtained

0.9.0
-----

Big refactoring, addition of the documentation, compliance to golint.

0.1.0
-----

Initial library
