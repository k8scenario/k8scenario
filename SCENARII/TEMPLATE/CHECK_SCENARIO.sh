
# Example to check a Pod is in the Running State:
kubectl get pods | grep ^critical | grep Running && echo "Scenario OK"

# TODO: Check no matches on other Pods, not Running ...

## -- Services: ------------------------------------------------
# To test ClusterIP:
#CLUSTERIP_PORT=$(GET_SVC_CLUSTERIP_PORT)
#CLUSTERIP_IP(GET_SVC_CLUSTERIP_IP)
#ENDPOINT=${CLUSTERIP_IP}:${CLUSTERIP_PORT}

# To test NodePort: (1st node)
PORT=$(GET_SVC_NODEPORT_PORT)
NODE_IP=$(GET_NODE_IP 1)
ENDPOINT=${NODE_IP}:${PORT}/1

echo Checking endpoint $ENDPOINT

## -- TEST POD: ------------------------------------------------
TEST_POD_SHELL tester "set -x; wget -O - --timeout 4 $ENDPOINT; RET=\$?; exit \$RET"
#TEST_POD_SHELL tester
#kubectl -n k8scenario delete pods/tester


