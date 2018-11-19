#!/bin/bash

source ./sanity.sh

# Environment variables
PRIMARYPATH=$GOPATH/src/github.com/project-flogo/microgateway

x=0 y=0 z=0
fetch::example::recipeslist

a=0 b=0 c=0
fetch::activity::recipeslist
echo -----------------------------------------
echo "Examples-------Total number of test cases=$z, Total number of testcases paseed=$x, Total number of testcases failed=$y"
echo "Activity-------Total number of test cases=$c, Total number of testcases paseed=$a, Total number of testcases failed=$b"

if [[ $z == $x ]] && [[ $c == $a ]]; then
    echo "All the tests are passed"
else
    echo "Number of testcases failed with for examples=$y and for activity testcases=$b"
    exit 1
 fi 
 echo -----------------------------------------