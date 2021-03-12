#!/usr/bin/env bash

# TODO: Polish this later using test framework

printf "Test: stack install works with specified version.\n"
./install.sh v0.24.0
which stack
stack -v
sudo rm /usr/local/bin/stack

printf "Test: stack install works with latest version.\n"
./install.sh
which stack
stack -v
