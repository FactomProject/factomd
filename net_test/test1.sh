#!/bin/bash

factomd -count=5 -net=tree -folder="test1-" -port=8089 -serve=9000 -connect="tcp://217.0.0.1:9500"
