# factomd Docker Helper

The factomd Docker Helper is a simple tool to help build and run factomd as a container

## Prerequisites

You must have at least Docker v17 installed on your system.

Having this repo cloned helps too 😇

## Build
From wherever you have cloned this repo, run

`docker build -t factomd_container .`

(yes, you can replace **factomd_container** with whatever you want to call the container.  e.g. **factomd**, **foo**, etc.)

#### Cross-Compile
To cross-compile for a different target, you can pass in a `build-arg` as so

`docker build -t factomd_container --build-arg GOOS=darwin .`

## Run
#### No Persistence
`docker run --rm -p 8090:8090 factomd_container`
  
* This will start up **factomd** with no flags.
* The Control Panel is accessible at port 8090  
* When the container terminates, all data will be lost
* **Note** - In the above, replace **factomd_container** with whatever you called it when you built it - e.g. **factomd**, **foo**, etc.

#### With Persistence
1. `docker volume create factomd_volume`
2. `docker run --rm -v $(PWD)/factomd.conf:/source -v factomd_volume:/destination busybox /bin/cp /source /destination/factomd.conf`
3. `docker run --rm -p 8090:8090 -v factomd_volume:/root/.factom/m2 factomd_container`

* This will start up **factomd** with no flags.
* The Control Panel is accessible at port 8090  
* When the container terminates, the data will remain persisted in the volume **factomd_volume**
* The above copies **factomd.conf** from the local directory into the container. Put _your_ version in there, or change the path appropriately.
* **Note**.  In the above
   * replace **factomd_container** with whatever you called it when you built it - e.g. **factomd**, **foo**, etc.
   * replace **factomd_volume** with whatever you might want to call it - e.g. **myvolume**, **barbaz**, etc.

#### Additional Flags
In all cases, you can startup with additional flags by passing them at the end of the docker command, e.g.

`docker run --rm -p 8090:8090 factomd_container -port 9999`


## Copy
So yeah, you want to get your binary _out_ of the container. To do so, you basically mount your target into the container, and copy the binary over, like so


`docker run --rm --entrypoint='' -v <FULLY_QUALIFIED_PATH_TO_TARGET_DIRECTORY>:/destination factomd_container /bin/cp /go/bin/factomd /destination`

e.g.

`docker run --rm --entrypoint='' -v /tmp:/destination factomd_container /bin/cp /go/bin/factomd /destination`

which will copy the binary to `/tmp/factomd`

**Note** : You should replace ** factomd_container** with whatever you called it in the **build** section above  e.g. **factomd**, **foo**, etc.

#### Cross-Compile
If you cross-compiled to a different target, your binary will be in `/go/bin/<target>/factomd`.  e.g. If you built with `--build-arg GOOS=darwin`, then you can copy out the binary with

`docker run --rm --entrypoint='' -v <FULLY_QUALIFIED_PATH_TO_TARGET_DIRECTORY>:/destination factomd_container /bin/cp /go/bin/darwin_amd64/factomd /destination`

e.g.

`docker run --rm --entrypoint='' -v /tmp:/destination factomd_container /bin/cp /go/bin/darwin_amd64/factomd /destination` 

which will copy the darwin_amd64 version of the binary to `/tmp/factomd`

**Note** : You should replace ** factomd_container** with whatever you called it in the **build** section above  e.g. **factomd**, **foo**, etc.
