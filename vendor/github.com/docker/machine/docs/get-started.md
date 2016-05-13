<!--[metadata]>
+++
title = "Get started with a local VM"
description = "Get started with Docker Machine and a local VM"
keywords = ["docker, machine, virtualbox, local"]
[menu.main]
parent="workw_machine"
weight=-70
+++
<![end-metadata]-->

# Get started with Docker Machine and a local VM

Let's take a look at using `docker-machine` for creating, using, and managing a
Docker host inside of <a href=
"https://www.virtualbox.org/" target="_blank">VirtualBox</a>.

## Prerequisites

* Make sure  you have <a href="https://www.virtualbox.org/wiki/Downloads" target="_blank">the latest VirtualBox</a> correctly installed
on your system. If you used <a href="https://www.docker.com/products/docker-toolbox" target="_blank">Toolbox</a> for <a href="https://docs.docker.com/engine/installation/mac/" target="_blank">Mac</a> or <a href="https://docs.docker.com/engine/installation/windows/" target="_blank">Windows</a> to install Docker Machine, VirtualBox is automatically installed.

* If you used the Quickstart Terminal to launch your first machine and set your terminal environment to point to it, a default machine was automatically created. If this is the case, you can still follow along with these steps, but create another machine and name it something other than "default" (e.g., staging or sandbox).

##  Use Machine to run Docker containers

To run a Docker container, you:

* create a new (or start an existing) Docker virtual machine
* switch your environment to your new VM
* use the docker client to create, load, and manage containers

Once you create a machine, you can reuse it as often as you like. Like any VirtualBox VM, it maintains its configuration between uses.

The examples here show how to create and start a machine, run Docker commands, and work with containers.

## Create a machine

1. Open a command shell or terminal window.

    These command examples shows a Bash shell. For a different shell, such as C Shell, the same commands are the same except where noted.

2. Use `docker-machine ls` to list available machines.

    In this example, no machines have been created yet.

        $ docker-machine ls
        NAME   ACTIVE   DRIVER   STATE   URL   SWARM   DOCKER   ERRORS

3. Create a machine.

    Run the `docker-machine create` command, passing the string `virtualbox` to the `--driver` flag. The final argument is the name of the machine. If this is your first machine, name it `default`. If you already have a "default" machine, choose another name for this new machine.

        $ docker-machine create --driver virtualbox default
        Running pre-create checks...
        Creating machine...
        (staging) Copying /Users/ripley/.docker/machine/cache/boot2docker.iso to /Users/ripley/.docker/machine/machines/default/boot2docker.iso...
        (staging) Creating VirtualBox VM...
        (staging) Creating SSH key...
        (staging) Starting the VM...
        (staging) Waiting for an IP...
        Waiting for machine to be running, this may take a few minutes...
        Machine is running, waiting for SSH to be available...
        Detecting operating system of created instance...
        Detecting the provisioner...
        Provisioning with boot2docker...
        Copying certs to the local machine directory...
        Copying certs to the remote machine...
        Setting Docker configuration on the remote daemon...
        Checking connection to Docker...
        Docker is up and running!
        To see how to connect Docker to this machine, run: docker-machine env default

      This command downloads a lightweight Linux distribution
(<a href="https://github.com/boot2docker/boot2docker" target="_blank">boot2docker</a>) with the Docker daemon installed, and creates and starts a VirtualBox VM with Docker running.

4. List available machines again to see your newly minted machine.

        $ docker-machine ls
        NAME      ACTIVE   DRIVER       STATE     URL                         SWARM   DOCKER   ERRORS
        default   *        virtualbox   Running   tcp://192.168.99.187:2376           v1.9.1

5. Get the environment commands for your new VM.

    As noted in the output of the `docker-machine create` command, you need to tell Docker to talk to the new machine. You can do this with the `docker-machine env` command.

        $ docker-machine env default
        export DOCKER_TLS_VERIFY="1"
        export DOCKER_HOST="tcp://172.16.62.130:2376"
        export DOCKER_CERT_PATH="/Users/<yourusername>/.docker/machine/machines/default"
        export DOCKER_MACHINE_NAME="default"
        # Run this command to configure your shell:
        # eval "$(docker-machine env default)"

6. Connect your shell to the new machine.

        $ eval "$(docker-machine env default)"

      **Note**: If you are using `fish`, or a Windows shell such as
      Powershell/`cmd.exe` the above method will not work as described.
      Instead, see <a href="https://docs.docker.com/machine/reference/env/" target="_blank">the `env` command's documentation</a>
      to learn how to set the environment variables for your shell.

    This sets environment variables for the current shell that the Docker client will read which specify the TLS settings. You need to do this each time you open a new shell or restart your machine.

    You can now run Docker commands on this host.

## Run containers and experiment with Machine commands

Run a container with `docker run` to verify your set up.

1. Use `docker run` to download and run `busybox` with a simple 'echo' command.

        $ docker run busybox echo hello world
        Unable to find image 'busybox' locally
        Pulling repository busybox
        e72ac664f4f0: Download complete
        511136ea3c5a: Download complete
        df7546f9f060: Download complete
        e433a6c5b276: Download complete
        hello world

2. Get the host IP address.

    Any exposed ports are available on the Docker host’s IP address, which you can get using the `docker-machine ip` command:

        $ docker-machine ip default
        192.168.99.100

3. Run a webserver (<a href="https://www.nginx.com/" target="_blank">nginx</a>) in a container with the following command:

        $ docker run -d -p 8000:80 nginx

    When the image is finished pulling, you can hit the server at port 8000 on the IP address given to you by `docker-machine ip`. For instance:

            $ curl $(docker-machine ip default):8000
            <!DOCTYPE html>
            <html>
            <head>
            <title>Welcome to nginx!</title>
            <style>
                body {
                    width: 35em;
                    margin: 0 auto;
                    font-family: Tahoma, Verdana, Arial, sans-serif;
                }
            </style>
            </head>
            <body>
            <h1>Welcome to nginx!</h1>
            <p>If you see this page, the nginx web server is successfully installed and
            working. Further configuration is required.</p>

            <p>For online documentation and support please refer to
            <a href="http://nginx.org/">nginx.org</a>.<br/>
            Commercial support is available at
            <a href="http://nginx.com/">nginx.com</a>.</p>

            <p><em>Thank you for using nginx.</em></p>
            </body>
            </html>

  You can create and manage as many local VMs running Docker as you please; just run `docker-machine create` again. All created machines will appear in the output of `docker-machine ls`.

## Start and stop machines

If you are finished using a host for the time being, you can stop it with `docker-machine stop` and later start it again with `docker-machine start`.

        $ docker-machine stop default
        $ docker-machine start default

## Operate on machines without specifying the name

Some `docker-machine` commands will assume that the given operation should be run on a machine named `default` (if it exists) if no machine name is specified.  Because using a local VM named `default` is such a common pattern, this allows you to save some typing on the most frequently used Machine commands.

For example:

          $ docker-machine stop
          Stopping "default"....
          Machine "default" was stopped.

          $ docker-machine start
          Starting "default"...
          (default) Waiting for an IP...
          Machine "default" was started.
          Started machines may have new IP addresses.  You may need to re-run the `docker-machine env` command.

          $ eval $(docker-machine env)

          $ docker-machine ip
            192.168.99.100

Commands that follow this style are:

        - `docker-machine config`
        - `docker-machine env`
        - `docker-machine inspect`
        - `docker-machine ip`
        - `docker-machine kill`
        - `docker-machine provision`
        - `docker-machine regenerate-certs`
        - `docker-machine restart`
        - `docker-machine ssh`
        - `docker-machine start`
        - `docker-machine status`
        - `docker-machine stop`
        - `docker-machine upgrade`
        - `docker-machine url`

For machines other than `default`, and commands other than those listed above, you must always specify the name explicitly as an argument.

## Start local machines on startup

In order to ensure that the Docker client is automatically configured at the start of each shell session, some users like to embed `eval $(docker-machine env default)` in their shell profiles (e.g., the `~/.bash_profile` file). However, this fails if the `default` machine is not running. If desired, you can configure your system to start the `default` machine automatically.

Here is an example of how to configure this on OS X.

Create a file called `com.docker.machine.default.plist` under `~/Library/LaunchAgents` with the following content:

```
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
    <dict>
        <key>EnvironmentVariables</key>
        <dict>
            <key>PATH</key>
            <string>/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/bin</string>
        </dict>
        <key>Label</key>
        <string>com.docker.machine.default</string>
        <key>ProgramArguments</key>
        <array>
            <string>/usr/local/bin/docker-machine</string>
            <string>start</string>
            <string>default</string>
        </array>
        <key>RunAtLoad</key>
        <true/>
    </dict>
</plist>
```

You can change the `default` string above to make this `LaunchAgent` start any  machine(s) you desire.

## Where to go next

-   Provision multiple Docker hosts [on your cloud provider](get-started-cloud.md)
-   [Understand Machine concepts](concepts.md)
-   <a href="https://docs.docker.com/machine/drivers/index.md" target="_blank">Docker Machine driver reference</a>
-   <a href="https://docs.docker.com/machine/reference/" target="_blank">Docker Machine subcommand reference</a>
