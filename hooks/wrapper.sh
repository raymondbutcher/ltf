#!/bin/bash

# Redirect 3 (a new file descriptor) to stdout, and the original stdout to stderr.
# 0=stdin
# 1=stdout
# 2=stderr
# TODO: why can't I reverse the following lines?
exec 3>&1
exec 1>&2

echo "Hello from wrapper.sh"
export LTF_WRAPPER=1
source embedded.sh
env | grep LTF >&3
