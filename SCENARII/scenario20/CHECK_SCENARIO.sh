#!/bin/sh

SVC_NAME=ckad-demo

# To test ClusterIP:
CHECK_SVC_CLUSTERIP $SVC_NAME || exit 1
CHECK_SVC_NODEPORT  $SVC_NAME || exit 1

#TEST_POD_SHELL tester "set -x; wget -O - --timeout 4 $ENDPOINT; RET=\$?; exit \$RET"
#TEST_POD_SHELL tester
#kubectl -n k8scenario delete pods/tester


