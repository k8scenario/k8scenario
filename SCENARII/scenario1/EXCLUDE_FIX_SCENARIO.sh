
kubectl set image deploy/critical critical=mjbright/ckad-demo:1

# Wait up to 20 secs for all Pods to be up
sleep 1
MAX_LOOPS=10;
SLEEP=1
CONDITION_WAIT $MAX_LOOPS $SLEEP "CHECK_PODS_PREFIXED critical"


