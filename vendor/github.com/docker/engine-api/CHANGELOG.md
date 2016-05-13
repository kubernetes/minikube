# Changelog

Items starting with DEPRECATE are important deprecation notices. For more information on the list of deprecated APIs please have a look at https://docs.docker.com/misc/deprecated/ where target removal dates can also be found.

## 0.3.2 (2016-03-30)

## Client

- Revert setting the ServerName in the TLS configuration at client init. See https://github.com/docker/swarm/issues/2027.

## 0.3.1 (2016-03-23)

### Client

- Ensure that API paths are properly escaped.

## 0.3.0 (2016-03-22)

### Client

- Add context to every function.
- Fix issue loading a default TLS CA.
- Allow to configure the client with a given http.Client.
- Add support for Windows named pipes.
- Set default host for Solaris.
- Add quiet flag for image load.
- Add ability to hijack connections through a proxy.
- Correctly set content type for image load.
- Add support for getting token for login.

### Types

- Add struct for update restart policy.
- Add human friendly State for container.
- Use OS specific host when DOCKER_HOST is not set.
- Rename Status in info to SystemStatus.
- Add internal flag to network inspect.
- Add disk quota field to container.
- Add EnableIPv6 fields.
- Add Mounts to container.
- Add cgroup driver to info.
- Add userns to host config.
- Remove email from AuthConfig.
- Make AuthConfig fields optional.
- Add IO resource settings for Windows.
- Add storage driver to host config.
- Update NetworkName to return proper user defined network names.
- Support joining cgroups by container id.
- Add KernelMemory to info.
- Add UsernsMode to container config.
- Add CPU resource control for Windows.
- Add AutoRemove to host config.
- Add Status field to Volume.
- Add Label to Image, Network and Volume.
- Add RootFS to container.

## 0.2.3 (2016-02-02)

### Types

- Add missing status field.

## 0.2.2 (2016-01-13)

### Client

- Fix issue configuring response hijacking with TLS enabled.


## 0.2.1 (2016-01-12)

### Client

- Fix issue detecting missing images on container creation.

### Types

- Remove invalid json tag in endpoint configuration.
- Add missing fields in info structure.

## 0.2.0 (2016-01-11)

### Client

- Allow to force network disconnection. (docker 1.10)

### Types

- Add global and local alias configuration to network endpoint.
- Add network ID to network endpoint.
- Add IPAM options.
- Add Seccomp options.
- Fix issue referencing OOMKillDisable.


## 0.1.3 (2016-01-07)

### Client

- Fix issue sending all network configurations for a per network request.


## 0.1.2 (2016-01-07)

### Client

- Add interface to represent the API client.
- Restrict the fields send to the update endpoint to only those that are used.
- Send network settings as part of the container create request. (docker 1.10)
- Send network settings as part of the network connect request. (docker 1.10)

### Types

- Add PidsLimit as part of the host configuration.
- Add PidsStats to show PID stats.
- Add graph storage options to host configuration.
- Add NetworkConfig and EndpointIPAMConfig structs. (docker 1.10)


## 0.1.1 (2016-01-06)

### Client

- Delegate shmSize units conversion to the consumer.

### Types

- Add warnings to the volume list reponse.
- Fix image build options:
	* use 0 as default value for shmSize.


## 0.1.0 (2016-01-04)

### Client

- Initial API client implementation.

### Types

- Initial API types implementation.
