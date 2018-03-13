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

## Logging

Logging is set up to use the ELK (*Elasticsearch* + *Logstash* + *Kibana*) stack
for gathering and searching the logs:
 - a *factomd* instance sends its log entries to *Logstash*
 - *Logstash* transforms the log entries and forwards them to
   *Elasticsearch* database
 - a *Kibana* instance connects to the same *Elasticsearch* instance and allows
   you to search the logs.

### Setting up factomd

The node needs to be instructed to use Logstash as its log target. Use the
following startup parameters:

```
factomd -loglvl info -logstash -logurl=localhost:8345
```

Once *factomd* is started it will immediately start sending the logs. Note that
you need to start the *docker-compose* first, otherwise *factomd* will refuse to
start.

### Setting up Kibana

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

### Filtering the logs

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

### Details

#### Factomd

*Factomd* is instrumented using the
[logrus](https://github.com/sirupsen/logrus) library for logging, but it is not
enabled by default, the following parameters enable logging and start sending
the logs to `localhost:8345`:

 - `-loglvl info` - enables logging via the logging framework (by default the
   log level is set to `none`)
 - `-logstash` - enables sending logs to a *Logstash* instance (by default it
   logs to *stdout*)
 - `-logurl` - sets the location of the *Logstash* instance, since the docker
   container exposes port 8345 to the local machine, you can use `localhost` as
   the target.

See `factomd -h` output for more information about these parameters.

#### Logstash

*Logstash* serves as a target for *factomd* to log to, it also allows
performing transformations on log entries. The provided configuration creates
a single pipeline (see [logstash.conf](./logstash/pipeline/logstash.conf))
that listens on port 8345 (this port is mapped from the container to the same
port on host) and forwards entries to *Elasticsearch*.

Ports:
 - `8345` - port to sends logs to.

Configs:
 - `./logstash/config/logstash.yml` - main *Logstash* configuration
 - `./logstash/pipeline/logstash.yml` - the pipeline for *factomd*

##### Multiple factomd instances

All the entries received by *Logstash* automatically have the `host` field which
contains the IP address of the node that sent the logs, this is however not very
helpful if you're running multiple nodes on the same machine. There are multiple
ways to fix this, one would be to have different instances connect to different
ports and tag the entries based on the port it connects to.

Example:

 - modify the pipeline configuration file to allow multiple ports:
   ```
   input {
       tcp {
           port => 8345
           tags => ["factomd_1"]
       }

       tcp {
           port => 8346
           tags => ["factomd_2"]
       }
   }
   ```
 - in `docker-compose.yml`, add a new port to map it to the port on the host:
   ```
   logstash:
       ports:
           - "8345:8345"
           - "8346:8346"
   ```
 - start different *factomd* instances with different ports:

   ```
   factomd -loglvl info -logstash -logurl=localhost:8345
   ```
   ```
   factomd -loglvl info -logstash -logurl=localhost:8346
   ```

The entries from different nodes will now be tagged with the instance name.

#### Elasticsearch

The environment creates an *Elasticsearch* instance that is used for
storing and searching logs generated by the network.

A single *Elasticseach* node container is started, since you should not need to
connect to it directly, the ports are not mapped to the host, but are available
to other containers.

Configs:
 - `./elasticsearch/config/elasticsearch.yml` - main *Elasticsearch* configuration

#### Kibana

A single *Kibana* node with a default configuration is created and connected to
the same *Elasticsearch* instance that *Logstash* forwards to, so that you can
use it to analyze the logs. Note that we're not creating any default mappings
for the log entries, so you won't see the fields used in the log entries
directly, but they will be available in the `message` field.

Ports:
 - `5601` - the *Kibana* Web UI

Configs:
 - `./kibana/config/kibana.yml` - main *Kibana* configuration

## Metrics

All *factomd* nodes have a built-in mechanism that gather various metrics about
the node using
[prometheus/client_golang](https://github.com/prometheus/client_golang). The
*Prometheus* instance periodically runs a scrape job that pull the metrics
exposed by *factomd* and stores it in a database.

### Setting up Prometheus

*Factomd* supports collecting metrics out-of-the-box, so no additional setup is
necessary. *Factomd* listens on port 9876 for *Prometheus* connections and
exposes them using the default `/metrics` path.

### Viewing collected metrics
To view the collected metrics open the *Prometheus* web UI at:
http://localhost:9090/, select one of the metrics from the dropdown and hit the
*Execute* button. If everything works, you should see lots of `factomd_*`
metrics in the dropdown. The metrics for all instances are labeled using the
hostname/port that was used for scraping the metrics, e.g.
`instance=factomd_1:9876`.

### Details

*Prometheus* container is currently configured to get the list of services to
monitor from a list of instances in the
[instances.json](./prometheus/config/instances.json) file. This file is mapped
to the container and watched by the *Prometheus* instance, so whenever you
change its contents, *Prometheus* should pick this up and start monitoring
a new node without restarting.

Since *Prometheus* connects to an instance of *factomd* from its container, the
solution differs depending on the your host system:
 - on Linux - the *factomd* instance is available at `localhost:9876`
 - on Docker for Mac/Windows - there is a special DNS entry that allows you to
   connect to the host from within the container, so your instance is available
   at e.g. `docker.for.mac.host.internal:9876`.

Note that both of these are added by default in `instances.json` to allow it to
work out-of-the-box, so without changes you might see *Prometheus* reporting
that one of the instances is down (see `up` metric).

Ports:
 - `9090` - *Prometheus* Web UI

Configs:
 - `./prometheus/config/prometheus.yml` - main *Prometheus* config
 - `./prometheus/config/instances.json` - list of factomd instances to connect to

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
