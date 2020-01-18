#!/bin/sh

# kubectl delete pod/scen3test

# Original: require extra shell procs:
#PORT=$(kubectl get service/ckad-demo | awk '/:/ { FS=":"; $0=$5; FS="/"; $0=$2; print $1; }')
#CLUSTER_IP=$(kubectl get service/ckad-demo | awk '/:/ { print $3; }')

# Lighterweight:
PORT=$(kubectl get service/ckad-demo -o custom-columns=NODEPORT:.spec.ports[0].nodePort --no-headers)
#CLUSTER_IP=$(kubectl get service/ckad-demo -o custom-columns=CLUSTERIP:.spec.clusterIP --no-headers)
#POD_IPS=$(kubectl get pods --no-headers -o custom-columns=CLUSTERIP:.status.podIP)

#POD_IP=$(kubectl get pods --no-headers -o wide | grep ckad-demo | awk '{ print $6; exit(0); }')

# Any node would do:
# POD_NODE=$(kubectl get pods --no-headers -o wide | awk '{ print $7; exit(0); }')
ANY_NODE_IP=$(kubectl get pods ckad-demo-6fdb684968-hq8qs -o custom-columns=IP:.status.hostIP --no-headers)

# Could have done this - but extra process:
# ANY_NODE_IP=$(kubectl get nodes -o custom-columns=IP:.status.addresses[0].address --no-headers | head -1)

EP=http://${ANY_NODE_IP}:${PORT}/1

LOOPS=10

VERSION1() {
    for i in $(seq $LOOPS); do
        echo "[loop $i/$LOOPS] Checking endpoint <$EP>"
        kubectl run --restart=Never --rm -it --generator=run-pod/v1 --image=alpine:latest k8scenario-test -- /bin/sh -c "wget -qO - --timeout 4 $EP; exit \$?"

        # Bail out on single failure
        [ $? -ne 0 ] && { echo "Error seen on loop $i/$LOOPS"; exit 1; }
    done
}

VERSION2() {
    # Faster detection on pass case by looping in container:
    echo "Checking endpoint <$EP> ... (for $LOOPS loops) "

    ## Bail out on single failure
    #kubectl run --restart=Never --rm -it --generator=run-pod/v1 --image=alpine:latest k8scenario-test -- /bin/sh -c "for i in \$(seq 10); do echo loop\$i/$LOOPS; wget -qO - --timeout 4 $EP; [ \$? -ne 0 ] && { echo \"Error seen on loop \$i\"; exit 1; }; done"
    kubectl run --restart=Never --rm -it --generator=run-pod/v1 --image=alpine:latest k8scenario-test -- /bin/sh -c "\
	    for i in \$(seq 10); do
		    echo -n \"loop\$i/$LOOPS: \";
		    wget -qO - --timeout 4 $EP;
		    [ \$? -ne 0 ] && { echo \"Error seen on loop \$i\"; exit 1; };
	    done
	    "
}

VERSION2
echo "No errors seen on $LOOPS loops"
exit 0


