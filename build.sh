#!/bin/bash

source ./sanity.sh

# Environment variables
PRIMARYPATH=$GOPATH/src/github.com/project-flogo/microgateway

DESTDIR=artifacts
mkdir -p artifacts

x=0 y=0 z=0
fetch::example::recipeslist

echo "Total number of test cases=$z, Total number of testcases paseed=$x, Total number of testcases failed=$y"

if [[ $z == $x ]]; then
    echo "All the tests are passed"
else
    echo "Number of testcases failed with build is $y"
    exit 1
 fi   