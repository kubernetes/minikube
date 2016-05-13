<!--[metadata]>
+++
title = "ls"
description = "List machines"
keywords = ["machine, ls, subcommand"]
[menu.main]
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# ls

    Usage: docker-machine ls [OPTIONS] [arg...]

    List machines

    Options:

       --quiet, -q                                  Enable quiet mode
       --filter [--filter option --filter option]   Filter output based on conditions provided
       --timeout, -t "10"                           Timeout in seconds, default to 10s
       --format, -f                                 Pretty-print machines using a Go template

## Timeout

The `ls` command tries to reach each host in parallel. If a given host does not
answer in less than 10 seconds, the `ls` command will state that this host is in
`Timeout` state. In some circumstances (poor connection, high load, or while
troubleshooting), you may want to increase or decrease this value. You can use
the -t flag for this purpose with a numerical value in seconds.

### Example

    $ docker-machine ls -t 12
    NAME      ACTIVE   DRIVER       STATE     URL                         SWARM   DOCKER   ERRORS
    default   -        virtualbox   Running   tcp://192.168.99.100:2376           v1.9.1

## Filtering

The filtering flag (`--filter`) format is a `key=value` pair. If there is more
than one filter, then pass multiple flags (e.g. `--filter "foo=bar" --filter "bif=baz"`)

The currently supported filters are:

-   driver (driver name)
-   swarm  (swarm master's name)
-   state  (`Running|Paused|Saved|Stopped|Stopping|Starting|Error`)
-   name   (Machine name returned by driver, supports [golang style](https://github.com/google/re2/wiki/Syntax) regular expressions)
-   label  (Machine created with `--engine-label` option, can be filtered with `label=<key>[=<value>]`)

### Examples

    $ docker-machine ls
    NAME   ACTIVE   DRIVER       STATE     URL                         SWARM   DOCKER   ERRORS
    dev    -        virtualbox   Stopped
    foo0   -        virtualbox   Running   tcp://192.168.99.105:2376           v1.9.1
    foo1   -        virtualbox   Running   tcp://192.168.99.106:2376           v1.9.1
    foo2   *        virtualbox   Running   tcp://192.168.99.107:2376           v1.9.1

    $ docker-machine ls --filter name=foo0
    NAME   ACTIVE   DRIVER       STATE     URL                         SWARM   DOCKER   ERRORS
    foo0   -        virtualbox   Running   tcp://192.168.99.105:2376           v1.9.1

    $ docker-machine ls --filter driver=virtualbox --filter state=Stopped
    NAME   ACTIVE   DRIVER       STATE     URL   SWARM   DOCKER   ERRORS
    dev    -        virtualbox   Stopped                 v1.9.1

    $ docker-machine ls --filter label=com.class.app=foo1 --filter label=com.class.app=foo2
    NAME   ACTIVE   DRIVER       STATE     URL                         SWARM   DOCKER   ERRORS
    foo1   -        virtualbox   Running   tcp://192.168.99.105:2376           v1.9.1
    foo2   *        virtualbox   Running   tcp://192.168.99.107:2376           v1.9.1

## Formatting

The formatting option (`--format`) will pretty-print machines using a Go template.

Valid placeholders for the Go template are listed below:

| Placeholder    | Description                              |
| -------------- | ---------------------------------------- |
| .Name          | Machine name                             |
| .Active        | Is the machine active?                   |
| .ActiveHost    | Is the machine an active non-swarm host? |
| .ActiveSwarm   | Is the machine an active swarm master?   |
| .DriverName    | Driver name                              |
| .State         | Machine state (running, stopped...)      |
| .URL           | Machine URL                              |
| .Swarm         | Machine swarm name                       |
| .Error         | Machine errors                           |
| .DockerVersion | Docker Daemon version                    |
| .ResponseTime  | Time taken by the host to respond        |

When using the `--format` option, the `ls` command will either output the data exactly as the template declares or,
when using the table directive, will include column headers as well.

The following example uses a template without headers and outputs the `Name` and `Driver` entries separated by a colon
for all running machines:

    $ docker-machine ls --format "{{.Name}}: {{.DriverName}}"
    default: virtualbox
    ec2: amazonec2

To list all machine names with their driver in a table format you can use:

    $ docker-machine ls --format "table {{.Name}} {{.DriverName}}"
    NAME     DRIVER
    default  virtualbox
    ec2      amazonec2
