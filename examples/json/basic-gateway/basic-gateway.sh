#!/bin/bash

function get_test_cases {
    init ;
    local my_list=( testcase1 )
    echo "${my_list[@]}"
}

function init {
    flogo create -f flogo.json  > /tmp/test 2>&1
    cd MyProxy
    flogo build  > /tmp/test 2>&1
    cd ..
}

function testcase1 {
    ./MyProxy/bin/MyProxy > /tmp/basic1.log 2>&1 &
    pId=$!
    sleep 5
    response=$(curl --request GET http://localhost:9096/pets/1 --write-out '%{http_code}' --silent --output /dev/null) 
    curl http://localhost:9096/pets/1 > /tmp/test.log 2>&1
    if [ $response -eq 200 ] && [[ "echo $(cat /tmp/basic1.log)" =~ "Code identified in response output: 200" ]]
        then 
            echo "PASS"
        else
            echo "FAIL"
    fi
    kill -9 $pId
    rm -rf /tmp/basic1.log
}