#!/bin/sh


# kubectl -n k8scenario delete pod/scen2test

PORT=$(kubectl -n k8scenario get service/ckad-demo | awk '/:/ { FS=":"; $0=$5; FS="/"; $0=$2; print $1; }')
#CLUSTER_IP=$(kubectl -n k8scenario get service/ckad-demo | awk '/:/ { print $3; }')
#POD_IP=$(kubectl -n k8scenario get pods --no-headers -o wide | awk '{ print $6; exit(0); }')

# Any node would do:
NODE_IP=$(kubectl -n k8scenario get pods --no-headers -o wide | awk '{ print $7; exit(0); }')

case $NODE_IP in
    [1-9]*) # OK we got an IP address
        ;;
    *) # Assume we got a node-name
        NODE_IP=$(kubectl get nodes -o custom-columns=IP:.status.addresses[0].address --no-headers)
        ;;
esac

EP=${NODE_IP}:${PORT}

echo Checking endpoint $EP

kubectl run --restart=Never --rm -n k8scenario -it --generator=run-pod/v1 --image=alpine:latest ${NS}-test -- /bin/sh -c "set -x; wget -O - --timeout 4 $EP; RET=\$?; exit \$RET"


