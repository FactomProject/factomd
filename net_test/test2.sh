#!/bin/bash

factomd -count=1 -net=tree -folder="test2-" -port=8090 -connect="tcp://217.0.0.1:9000" -follower=true
