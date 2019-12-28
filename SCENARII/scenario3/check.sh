#!/bin/sh

#exec 2>&1 | tee /tmp/scenario4.txt
#kubectl -n scenario4 get pods | grep critical | grep Running && echo OK; echo $?
#set -x
kubectl -n k8scenario get pods | grep ^critical | grep Running && echo "Scenario OK"
EXIT=$?
exit $EXIT
