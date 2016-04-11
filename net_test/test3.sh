#!/bin/bash

factomd -count=5 -net=tree -folder="test3-" -port=8091 -connect="tcp://217.0.0.1:9000" -follower=true
