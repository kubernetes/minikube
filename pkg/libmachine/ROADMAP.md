# Machine Roadmap

Machine currently works really well for development and test environments. The
goal is to make it work better for provisioning and managing production
environments.

This is not a simple task -- production is inherently far more complex than
development -- but there are three things which are big steps towards that goal:
**client/server architecture**, **swarm integration** and **flexible
provisioning**.

(Note: this document is a high-level overview of where we are taking Machine.
For what is coming in specific releases, see our [upcoming
milestones](https://github.com/docker/machine/milestones).)

### Docker Engine / Swarm Configuration

Currently there are only a few things that can be configured in the Docker Engine and Swarm.  This will enable more operations such as Engine labels and Swarm strategies.

### Boot2Docker Migration Support

Currently both Machine and Boot2Docker provider similar functionality.  This will enable users to migrate from boot2docker to machine.

### Expand Provisioner

Machine currently supports running Boot2Docker for "local" providers and Ubuntu for "remote" providers.  This will expand the provisioning capabilities to include other base operating systems such as Red Hat-like distributions and possibly other "just enough" operating systems.

### Windows Experience

Currently, the Machine on Windows experience is not as good as the Mac / Linux.  There is no "recommended" path to use Machine and there are several inconsistencies on Windows such as logging and output formatting.

# Project Planning

An [Open-Source Planning Process](https://github.com/docker/machine/wiki/Open-Source-Planning-Process) is used to define the Roadmap. [Project Pages](https://github.com/docker/machine/wiki) define the goals for each Milestone and identify current progress.
