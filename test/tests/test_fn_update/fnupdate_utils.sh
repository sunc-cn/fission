#!/bin/bash

#
# Common methods used for testing function update
#

set -euo pipefail

test_fn() {
    echo `date +%Y/%m/%d:%H:%M:%S` "Doing an HTTP GET on the function's route"
    echo `date +%Y/%m/%d:%H:%M:%S` "Checking for valid response"

    while true; do
      response0=$(curl http://$FISSION_ROUTER/$1)
      echo `date +%Y/%m/%d:%H:%M:%S` "Dumping response : $response0"
      echo $response0 | grep -i $2
      if [[ $? -eq 0 ]]; then
        break
      fi
      sleep 1
    done
}
export -f test_fn

dump_function_pod_logs() {
    ns=$1
    fns=$2

    functionPods=$(kubectl -n $fns get pod -o name -l functionName)
    for p in $functionPods
    do
	echo "--- function pod logs $p ---"
	containers=$(kubectl -n $fns get $p -o jsonpath={.spec.containers[*].name} --ignore-not-found)
	for c in $containers
	do
	    echo "--- function pod logs $p: container $c ---"
	    kubectl -n $fns logs $p $c || true
	    echo "--- end function pod logs $p: container $c ---"
	done
	echo "--- end function pod logs $p ---"
    done
}
export -f dump_function_pod_logs