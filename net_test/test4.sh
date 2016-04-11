#!/bin/bash

factomd -count=5 -net=tree -folder="test4-" -port=8092 -connect="tcp://217.0.0.1:9000" -follower=true
