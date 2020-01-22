#!/bin/sh

PORT=$(GET_SVC_CLUSTERIP_PORT ckad-demo)
IP=$(GET_SVC_CLUSTERIP_IP ckad-demo)

EP=http://${IP}:${PORT}/1

LOOPS=10

VERSION1() {
    for i in $(seq $LOOPS); do
        echo "[loop $i/$LOOPS] Checking endpoint <$EP>"
        TEST_POD_SHELL "wget -qO - --timeout 4 $EP; exit \$?"

        # Bail out on single failure
        [ $? -ne 0 ] && { echo "Error seen on loop $i/$LOOPS"; exit 1; }
    done
}

VERSION2() {
    # Faster detection on pass case by looping in container:
    echo "Checking endpoint <$EP> ... (for $LOOPS loops) "

    ## Bail out on single failure
    #kubectl run --restart=Never --rm -it --generator=run-pod/v1 --image=alpine:latest $TESTPOD -- /bin/sh -c "\
    TEST_POD_SHELL "\
	    for i in \$(seq 10); do
		    echo -n \"loop\$i/$LOOPS: \";
		    wget -qO - --timeout 1 $EP;
		    [ \$? -ne 0 ] && { echo \"Error seen on loop \$i\"; exit 1; };
	    done; exit 0
	    "
}

kubectl get pod/$TESTPOD 2>/dev/null && kubectl delete pod/$TESTPOD
VERSION2
RET=$?
[ $RET -ne 0 ] && { echo "Error - $RET - launching Pod"; exit $RET; }

echo "No errors seen on $LOOPS loops"
exit 0


