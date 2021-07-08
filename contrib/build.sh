#!/bin/bash

SELF=$(cd `dirname $0`; pwd)

cd $SELF/..
if [ ! -d local/src/github.com/google ]; then
   mkdir -p local/src/github.com/google
   pushd local/src/github.com/google
   ln -s ../../../.. zoekt
   popd
fi
cd local
GOPATH=`pwd` go install github.com/google/zoekt/contrib/cmd/zoekt-webserver
