#!/bin/sh

# kubectl -n k8scenario delete pod/scen3test

PORT=$(kubectl -n k8scenario get service/ckad-demo | awk '/:/ { FS=":"; $0=$5; FS="/"; $0=$2; print $1; }')
#CLUSTER_IP=$(kubectl -n k8scenario get service/ckad-demo | awk '/:/ { print $3; }')
#POD_IP=$(kubectl -n k8scenario get pods --no-headers -o wide | awk '{ print $6; exit(0); }')

# Any node would do:
NODE_IP=$(kubectl -n k8scenario get pods --no-headers -o wide | awk '{ print $7; exit(0); }')

EP=${NODE_IP}:${PORT}

LOOPS=10

for i in $(seq $LOOPS); do
    echo "[loop $i] Checking endpoint $EP"
    #kubectl run --restart=Never --rm -n k8scenario -it --generator=run-pod/v1 --image=alpine:latest ${NS}-test -- /bin/sh -c "set -x; wget -O - --timeout 4 $EP; RET=\$?; exit \$RET"
    kubectl run --restart=Never --rm -n k8scenario -it --generator=run-pod/v1 --image=alpine:latest ${NS}-test -- /bin/sh -c "wget -O - --timeout 4 $EP; exit \$?"

    # Bail out on single failure
    [ $? -ne 0 ] && { echo "Error seen on loop $i"; exit 1; }
done

echo "No errors seen on $LOOPS loops"
exit 0


