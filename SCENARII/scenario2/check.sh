#!/bin/sh

#exec 2>&1 | tee /tmp/scenario2.txt
#kubectl -n scenario2 get pods | grep critical | grep Running && echo OK; echo $?
#set -x
kubectl -n k8scenario get pods | grep ^critical | grep Running && echo "Scenario OK"
EXIT=$?
exit $EXIT
