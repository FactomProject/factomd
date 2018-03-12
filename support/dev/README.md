# Dev environment setup guide

The *docker-compose* setup included here allows creating a development
environment for all supporting services (for logging, monitoring etc.) running
in docker containers on the local machine.

Starts the following:
 - a full ELK (*Elasticsearch* + *Logstash* + *Kibana*) stack that gathers logs
   from the network
 - a *Prometheus* instance for gathering metrics from the nodes

> Warning: This setup is for development environment only and should not be
> used in production. Specifically the data for all services are not persisted
> accross restarts.

## Basic usage

This setup uses *docker* and *docker-compose* to run all supporting services in
docker containers. For more details about the usage, consult with
[docker](https://docs.docker.com/) and
[*docker-compose*](https://docs.docker.com/compose/) documentation.

### Prerequisites

To create the environment install the following components first:
 - [docker](https://www.docker.com/community-edition)
 - [docker-compose](https://docs.docker.com/compose/install/)

All commands below assume the current directory is `<REPO_ROOT>/support/dev`.

### Creating the environment

```
docker-compose up
```

This command:
 - builds all container images if necessary
 - starts all the containers
 - starts printing all *stdout* logs from the containers to the foreground.

If you prefer to run this in the background, use the `-d` option:

```
docker-compose up -d
```

After *docker-compose* finishes the startup, be sure to check the output of
`docker ps` command to make sure all the services are running.

### Starting/stopping the environment

```
docker-compose stop
```

This command stops all the containers without removing them. Note that all
non-persistent storage will be wiped out as the containers are restarted.

The containers can be started again with:

```
docker-compose start
```

### Tearing down the environment

```
docker-compose down
```

This command:
 - removes all service containers
 - removes all created networks

## Using services

Once the environment is properly built and started you can start using it for
monitoring a factomd on your local machine.

### Logging

Logging is set up to use the ELK (*Elasticsearch* + *Logstash* + *Kibana*) stack
for gathering and searching the logs:
 - a *factomd* instance sends its log entries to *Logstash*
 - *Logstash* transforms the log entries and forwards them to the
   *Elasticsearch* database for searching
 - a *Kibana* instance connects to the same *Elasticsearch* instance and allows
   you to search the logs.

#### Setting up factomd

The node needs to be instructed to use Logstash as its log target (multiple
nodes should use the same Logstash instance). Use the following startup
parameters:

```
factomd -loglvl info -logstash -logurl=localhost:8345
```

 - `loglvl info` - enables logging via the logging framework (by default the log
   level is set to `none`)
 - `logstash` - enables sending logs to a *Logstash* instance (by default it
   logs to *stdout*)
 - `logurl` - sets the location of the *Logstash* instance, since the docker
   container exposes port 8345 to the local machine, you can use `localhost` as
   the target.

Once *factomd* is started it will immediately start sending the logs.

#### Setting up Kibana

To open the *Kibana UI*, go to the following address in your browser:
http://localhost:5601/. The first time you do this after rebuilding the
environment you'll need to set up the default index pattern. Go to *Management*
-> *Index patterns*, in the first step define the *Index pattern* as `*`, in
the second step select `@timestamp` as the *Time Filter field name* and hit the
*Create index pattern* button. If you forget about this step any attemp to look
at the logs will take automatically take you to this screen.

Once the index pattern is set up, go to the *Discover* tab. You should see the
list of logs created by all *factomd* nodes in the network ordered by the
`@timestamp` field.

#### Filtering the logs

*Kibana* comes with some powerful tools for filtering the logs, some examples:

 * To filter the logs by the time range, click on the *Last 15 minutes* button
   in the top right for some filtering options.
 * To filter the logs by the instance it was created from, click on the `host`
   field in the list of *Available fields* on the left. Clicking on the `+`
   icon for one of the hosts will automatically add a filter that shows only
   logs from this host.
 * The search box allows filtering the entries by the content using the
   *Lucene* query syntax, e.g. entering `NOT eom` will filter out all messages
   the contain the string `eom` in the log message.

### Metrics

Metrics are pulled from *factomd* instances into *Prometheus*. To view the
collected metrics open the *Prometheus* web UI at: http://localhost:9090/,
select one of the metrics from the dropdown and hit the *Execute* button. The
metrics for all instances are labeled using the hostname/port that was used for
scraping the metrics, e.g. `instance=factomd_1:9876`.

### Networking

The current setup allows manipulating the connectivity between containers to
test e.g. various network conditions by inserting and removing *iptables*
rules. In addition to the commands described here, please consult the
*iptables* and *docker* documentation.

#### Linux

When hosting the setup on Linux, you should be able to manipulate the
*iptables* directly from your host machine, note however that you will most
likely need to run all the commands as root.

#### Mac OS / Windows

When hosting the setup on Mac or Windows (using *Docker for Mac* or *Docker for
Windows* respectively), you need to take into account that the docker engine is
running inside a Linux VM, so on the host machine you will not have access e.g.
the list of running processes or *iptables* setup.

On Mac OS you can attach to the VM shell using this command:

```
screen ~/Library/Containers/com.docker.docker/Data/com.docker.driver.amd64-linux/tty
```

To detach: `CTRL-a CTRL-\` and hit `y`.

Alternatively on both Windows and Mac OS you can start a priviledged container
that will give you access to the host, see this article:
https://blog.jongallant.com/2017/11/ssh-into-docker-vm-windows/.

#### Viewing the iptable rules

To view all tables and rules:

```
iptables -L
```

This will show all currently set up *iptables* rules. If everything is set up
correctly, you should see entries created by docker in the `DOCKER` chain,
e.g.:

```
Chain DOCKER (3 references)

...

ACCEPT     tcp  --  anywhere             10.7.0.3             tcp dpt:8090
ACCEPT     tcp  --  anywhere             10.7.0.2             tcp dpt:8090
ACCEPT     tcp  --  anywhere             10.7.0.1             tcp dpt:8090
```

#### Dropping connections

To drop the connections between two containers, get their IP addresses first
(see commands below) and add the following rule:

```
iptables -I FORWARD 1 -s 10.7.0.1 -d 10.7.0.2 -j DROP
```

This will drop all connections from `10.7.0.1` (`factomd_1`) to `10.7.0.2`
(`factomd_2`). It is important to use `-I`, so that the rule gets inserted
first and it will match a given network packet before it is matched by rules
created by docker.

#### Restoring connections

To restore the connectivity, remove the previously added rule. To do this you
need to first obtain the current number of the added rule

```
iptables -L --line-numbers
```

```
...

Chain FORWARD (policy DROP)
num  target     prot opt source               destination
1    DROP       all  --  10.7.0.1             10.7.0.2

...
```

The previously added rule has number 1 in the `FORWARD` chain, so we can drop
it with:

```
iptables -R FORWARD 1
```

## Useful commands

### Listing containers

```
docker ps
```
```
docker-compose ps
```

### Viewing the stdout logs

Single container:

```
docker logs <container_name>
```
```
docker logs <container_name> -f
```

All containers:

```
docker-compose logs
```
```
docker-compose logs -f
```

### Logging into a container

```
docker exec -it <container_name> bash
```

### Stopping a single container

To gracefully shutdown a container (the application receives a SIGTERM signal):

```
docker stop <container_name>
```

To kill the container:

```
docker kill <container_name>
```

### Getting IP addresses

Display all containers with all networks they belong to and their static /
assigned IP addresses in a network:

```
docker inspect -f '{{.Name}} - {{range $name, $net := .NetworkSettings.Networks}}{{$name}}:{{$net.IPAddress}} {{end}}' $(docker ps -aq)
```

## Service setup details

The environment is created using *docker* containers that are put together
using *docker-compose*.

> Warning: in this environment data are not persisted accross container
> restarts, this includes: *factomd* data, logs, metrics, settings etc.

All services get their configuration from files mapped as readonly volumes to
the corresponding containers. Changing a configuration value requires
restarting the container.

Services are currently not set up to restart automatically in case of failure,
so they will need to be started manually.

### Factom network

The created *factomd* network is using 3 *factomd* instances that are set up as
peers in a custom *factomd* network. The configuration files for all the
instances are located in the `factom` directory. Note that each of the
instances has a different identity set up in the config file, so that all nodes
can be used as leaders in the network (see `IdentityChainID`,
`LocalServerPrivKey`, `LocalServerPublicKey` in `factom/factomd_*.conf`).

All the nodes are set up to log to the provided *Logstash* instances by adding
the command line parameters: `-logstash -logurl=logstash:8345` (see the
`command` section in the `docker-compose.yml` file.

All the instances have static IP addresses assigned to them. When each of the
instances start, they connect to an *nginx* instance provided as one of the
services, which serves the list of IP addresses for all the nodes, so that all
instances connect to each other.

The default assignment of IP addresses for *factomd* instances:
 * *factomd_1* - `10.7.0.1`
 * *factomd_2* - `10.7.0.2`
 * *factomd_3* - `10.7.0.3`

All other services present in the `factomd` docker network have their IPs
assigned automatically in `10.7.1.0/24`.

### ELK stack

The ELK (*Elasticsearch* + *Logstash* + *Kibana*) stack is used to gather logs
from multiple nodes that are running in the network. The logs are sent to the
*Logstash* instances which stores it in *Elasticsearch* and can later be
explored in *Kibana*.

#### Elasticsearch

The environment creates a single *Elasticsearch* instance that is used for
storing and searching logs generated by the network.

The configuration for the *Elasticsearch* instance is copied from
`elasticsearch/config/elasticsearch.yml`.

#### Logstash

*Logstash* collects the logs from factomd instances and forwards it to
*Elasticsearch*. The connections to the *Logstash* instance are using port 500,
this port is also mapped to the same port on the local machine, so that you can
use the same setup to log from *factomd* instances that you run outside of
docker containers. The *Logstash* configuration files and the pipeline
definition for *factomd* are located in the `logstash` directory.

#### Kibana

Kibana allows exploring, searching and visualizing logs created by factomd
network. The web UI is mapped to the local port 5601.

The configuration for *Kibana* is copied during the build from
`kibana/config/kibana.yml`.

### Prometheus

*Prometheus* periodically gathers metrics from a all *factomd* instances. The
built-in web UI is exposed to the local host using port `9090`.

Note that *Prometheus* is pull-based, so it fetches metrics from *factomd*
instances, not the other way around, so to add monitoring for your local nodes,
you'll need to modify its configuration and rebuild the container.

The configuration for *Prometheus* is copied during the build from
`prometheus/config/prometheus.yml`. Currently it pull metrics from all 3
instances and labels them using the instance name.

## Known issues

* If you are running this setup in a *Docker for Mac* or a *Docker for Windows*
  environment you might want to adjust the CPU and Memory settings
  (*Preferences...* -> *Advanced*), since the services started in this setup
  may require more than the defaults. It may also be useful to bump the default
  *docker-compose* HTTP timeout setting if you see some problems during
  startup, e.g.:

  ```
  COMPOSE_HTTP_TIMEOUT=120 docker-compose up
  ```

* The `factomd` instances sometimes exit after the first build, another
  `docker-compose up` command should bring start them correctly.

* There is an issue when using the environment on Mac OS:

  ```
  ERROR: for kibana  Cannot start service kibana: driver failed programming
  external connectivity on endpoint kibana
  (7e6b3eaddf72eb60f384edff6d5c0bbac759af0cf5c24cfaff646ec558815cd5): Timed out
  proxy starting the userland proxy
  ```

  Unfortunately this is an unresolved *Docker for Mac* issue for which there is
  no good solution (restarting does not help), you'll need to retry until it
  succeeds.
