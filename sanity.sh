#!/bin/bash

# Split recipes list into an array
function get::recipes::list() {
    IFS=\  read -a RECIPE <<<"$samples" ;
    set | grep ^IFS= ;
    # separating arrays by line
    IFS=$' \t\n' ;
    # fetching Gateway
    set | grep ^RECIPE=\\\|^samples= ;
}

# Fatch types(api's and json) list
function fetch::example::recipeslist() {
    pushd $PRIMARYPATH/examples
    TYPES=(*)
    for ((p=0; p<${#TYPES[@]}; p++));
    do
        TYPES[$p]=${TYPES[$p]}
        cd ${TYPES[$p]}
        ls -d * > $GOPATH/${TYPES[$p]}
        cat $GOPATH/${TYPES[$p]}
        cd ..
        samples=$(echo $(cat $GOPATH/${TYPES[$p]}));
        unset RECIPE
        get::recipes::list
        for ((k=0; k<"${#RECIPE[@]}"; k++));
        do
            echo ${RECIPE[$k]}
            microgateway::examples::sanity::test
        done
    done
    popd
}

# run sanity tests
function microgateway::examples::sanity::test() {
    if [[ -f $PRIMARYPATH/examples/${TYPES[$p]}/${RECIPE[$k]}/${RECIPE[$k]}.sh ]]; then
        pushd $PRIMARYPATH/examples/${TYPES[$p]}/${RECIPE[$k]};
        source ./${RECIPE[$k]}.sh
        value=($(get_test_cases))
        sleep 10        
        for ((i=0;i < ${#value[@]};i++))
        do
            value1=$(${value[i]})
            sleep 10
            if [[ $value1 == *"PASS"* ]];  then
                echo "${value[i]}-${RECIPE[$k]}":"Passed"
                x=$((x+1))
            else
                echo "${value[i]}-${RECIPE[$k]}":"Failed"
                y=$((y+1))
            fi
            z=$((z+1))
        done
        popd
    else
        echo "Sanity file does not exist"
    fi    
}