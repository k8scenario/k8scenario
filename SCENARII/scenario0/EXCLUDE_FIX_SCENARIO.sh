# Note: any image will do

# -------- 2 types of fix:

# Fix by running a Pod directly:
kubectl run --generator=run-pod/v1 --image=mjbright/ckad-demo:1 basictest-xxx

# Fix by creating a Deployment
kubectl create deploy --image=mjbright/ckad-demo:1 basictest 

# Wait up to 20 secs for all Pods to be up
sleep 1
MAX_LOOPS=10;
SLEEP=1
CONDITION_WAIT $MAX_LOOPS $SLEEP "CHECK_PODS_PREFIXED basictest"

