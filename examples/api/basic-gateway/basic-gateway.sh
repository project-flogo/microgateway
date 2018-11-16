#!/bin/bash

function get_test_cases {
    local my_list=( testcase1 )
    echo "${my_list[@]}"
}

function testcase1 {
    go run main.go > /tmp/basic1.log 2>&1 &
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
    sleep 5
    # kill -9 $(lsof -i:9096) 
    var=$(ps --ppid $pId)
    pId7=$(echo $var | awk '{print $5}')
    kill -9 $pId7
    rm -rf /tmp/basic1.log
}