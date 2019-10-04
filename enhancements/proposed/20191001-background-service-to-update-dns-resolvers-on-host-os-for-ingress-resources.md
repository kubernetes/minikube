# Background service to update dns resolvers on Host OS for ingress resources

First proposed: 2019-10-01
Authors: Josh Woodcock

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?
*   Are there usability, reliability, or technical debt concerns?

Please leave the above text in your proposal as instructions to the reader.

## Summary

Some capabilities in k8s can only be tested when using a domain name. An example would be if a developer is using a router like 
Istio or Traefik that matches inbound requests to a particular domain name. When running a development cluster that has 
some type of server, they can use [external dns](https://github.com/kubernetes-incubator/external-dns) to automatically 
configure their DNS to point to the right ip address. However, when they are using minikube they can only resolve host names locally 
since the cluster ip is only available to their Host OS.

If this proposal is implemented it will enable a developer to access `redis1.myproject1.test` on the right minikube
instance regardless of what the minikube ip is and their only configuration is the ingress resource that is installed in
the cluster.

## Goals

* Enable ingress host names configured in any minikube cluster to be automatically resolved to the right minikube ip
* After a user installs an ingress in minikube they can access the hostname for the ingress on their host web browser
  or command line with no additional steps
* This capability can be enabled with a maximum of 2 steps and never repeated even while new minikube instances are 
  created and torn down

## Non-Goals

* Resolve ingress host names for ingress resources configured on Kubernetes clusters that are not running on minikube
* Make the user do an additional step after adding an ingress resource in order to access a host name on their
  host OS web browser
* Create web browser plugins
* Resolve host names external addresses or ExternalName type services
* Use `minikube tunnel` with LoadBalancer services in order to somehow resolve host names
* Use NodePort without an ingress configuration
* Address ways of accessing services from the host which do not require using a specific host name
* Enabling this capability can be enabled with a single step
* Install local ingress ssl certificates on the Host OS with this service

## Design Details

### Design summary

There are 3 aspects of making the ideal solution work seamlessly for the end user although #1 and #3 are not part 
of this proposal

#### 1. A DNS service that runs in the cluster and resolves ingress host names to the minikube cluster ip
Use minikube as a DNS server for and have a DNS server which will resolve the DNS queries for ingress hosts to the IP 
address of the minikube instance. This is the most preferable solution since the ingress resources within the cluster 
would be limited to the cluster itself. For DNS resolution the A record for each DNS query would be the hostname of 
the ingress and the ip address would be the cluster ip. There is a working solution for a plugin that currently does 
this here: https://github.com/kubernetes/minikube/pull/5507

#### 2. A background service that runs on the host OS that monitors changes to host names and minikube ips
Run a background service that will update `/etc/resolver` or similar files on the Host OS as each new minikube instance
is launched and stopped and as there are updates, removals, or additions to the ingress resources installed in any 
minikube cluster

The service can be started with the command:
```bash
minikube ingress resolve
```

This service will need to request privileged access from the user to be installed and started since updating of resolver
configurations required administrative access on all known operating systems

`resolve` is a sub-command of ingress because I think there is a strong use case for a different proposal which could be started with:

```bash
minikube ingress ssl
```

which will install local certificates on the host OS automatically for ingress resources. That is not covered in this
proposal.

#### 3. The background service can be automatically stopped and started when an addon is enabled or disabled
This can reduce the steps required to enable this capability from 2 to 1. This is not covered as part of this proposal.

## Alternatives Considered

### Ways to access services
* Don't use ingress at all and just use NodePort then match the DNS name using one of the alternatives described below
  * Some services won't have a node port available. Especially if they are in a helm chart
  * Now I have to remember the NodePort every time I want to access my service which maybe I can commit that 5 digit 
    number to my memory for every service I need to access. If I don't have a good memory then this is really terrible
    for me as I have to keep a reference sheet of where everything is running or I have to look it up with `kubectl`
  * If the static ip address is somehow configured then I only have to create the DNS entries one time
* Use `minikube tunnel` and an ingress dns resolver like CoreDNS then add a resolver file for the internal host IP 
address
  * So then if I have 4 minikube clusters running for 4 different projects I have to stop, start, stop, start
    minikube tunnel and enter my sudo password every time I do. 
  * I can only resolve 1 cluster's load balancers entries at any given time and must use the internal service name
    like myservice.default.svc.cluster.local
* Don't use a host at all and trick the target services and configurations being tested that a host is being used 
utilizing some type of reverse proxy configuration that runs in the cluster
   * Probably possible although I didn't try it. In this case I still have to then enter a different IP address in my 
     web browser which will change every time I tear down and start my minikube cluster. 
   * Even if I can configure a static ip I still then have to remember which static ip belongs to which project

### Ways to resolve the host names on the host OS
* Have the user manually update the `/etc/hosts` file each time they have an updated, new, or removed host name for 
any of their ingresses or the ip of the minikube instance has changed
  * This is what users are doing today. This leads to pollution of the `/etc/hosts` file for every project a developer is
    on.
  * Even if a static ip address can be configured per profile the host names still must be updated any time there
    is an addition, deletion, or update
* Use a plugin that utilizes minikube instance as a dns service for ingress resources. Then have the user manually 
update the `/etc/resolver` configuration file each time the minikube ip changes for the cluster
  * This would be potentially substantially less painful although still somewhat painful if there was also a way to 
    configure static ip's for a profile since hypothetically this would be a one time step each time there is a new 
    minikube profile
* Create this service completely outside of minikube along with a plugin that utilizes the minikube instance as dns 
service for ingress resources and have the user manually start and stop this third party tool when the plugin is being 
used or not being used
  * This would not provide an ideal solution for the developer as they would have to install a tool outside of minikube
   in order to gain the complete experience of automatic host name resolution for ingress resources but it could work
  * A possible benefit of this would be that the tool could be used for clusters outside of minikube and with minikube
   alternatives like micro k8s, etc.
* Have a user add a `/etc/resolver` file for each and every host for each and every ingress running in each and every
one of their minikube instances
  * This actually feels worse than `/etc/hosts` pollution because of how many files I will have to create
  * If static ip addresses are configurable this is not nearly as bad
  * At least with this solution the configuration files for all my projects wouldn't be in a single file. Still 
    very manual and not much better than a `/etc/hosts` file
* Have a service running in my minikube instance which will update a public dns server to map local domain names to 
  the internal IP address of my minikube instance. 
  * Hey technically this should work for 1 developer. Although some people would be confused why redis-internal.mysite.com 
    points to an address that doesn't exist on their local machine like `192.168.1.200`
  * This would also require each and every developer to have their own domain name which is also possible. Then you have to 
    somehow create a website that runs dns servers like route53 and create authentication mechanism that allows
    a developer to update those sites.
  * This is a huge bit of overhead and certainly more work than updating an `/etc/hosts` file
* Use some solution like this guy made https://www.mailgun.com/blog/creating-development-environments-with-kubernetes-devgun
  to wrap minikube and just resolve the internal dns resources to a cidr block which is somehow mapped to only a single
  minikube instance at a time. 
  * There's no code for this so I don't really know how it works in order to re-create it
  * If it was open source some extra stuff for mailgun would get installed which I don't really want
  * This approach would presumably only work for a single minikube instance at a time. 
  * I have to access my services with a long service name which means I need to know all the internal details
    about which namespace the service is installed in. 
  * I have to type `someservice.somenamespace.svc.cluster.local` in my browser over and over again which is long and obtuse. Not
    terrible but not great. 
* Use DNSmasq with this tool https://github.com/superbrothers/minikube-ingress-dns/blob/master/README.md which 
  routes dns queries through a local DNSmasq service running on 127.0.0.1. 
  * Unfortunately this tool hasn't been updated in quite awhile. 
  * I tried to get this working and couldn't
  * Its written in ruby so I guess this only runs on mac? Or I have to install ruby?
* Use a service like nip.io or xip.io to resolve a generic hostname to an IP address. 
  * This is workable but it doesn't provide a great user experience since the IP address of minikube is constantly 
    changing every time it is recreated. Even if we can solve the problem of a constantly changing ip, developers also 
    would have to memorize an obscure IP address converted to a domain.
  * This also creates a dependency on having an internet connection even though the rest of my services can be run 
    independent of an internet connection.
* Use a scheme of config files which identify which profile should have which domain names then minikube command line
  can update the resolver files each time there is a change to one of these configs. Or write a third party tool that 
  does the same. 
  * This could work but would require the user to update configured host names which could get out of sync with the 
    ingress resources in their cluster
  * In addition to having to manage these configurations there needs to be someway to view and update them.
  * Its nearly not really better than manually configuring `/etc/hosts` or `resolver` files. Some might say its worse 
    because at least I know how to update `/etc/hosts`
* Use a static IP address for each minikube instance so that at least the ip address isn't changing
  * This doesn't address the issue of host names being added, removed, updated over the lifecycle of a project and all 
    the manual work that has to be done, remembered, troubleshooted, in order to make work on a larger team