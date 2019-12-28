


## TODO:
#  kubectl -n k8scenario get pods | grep ^critical | grep Running && echo "Scenario OK"

getServicePort () {
    SERVICE=$1
    PORT=$(kubectl -n k8scenario get service/$SERVICE | awk '/:/ { FS=":"; $0=$5; FS="/"; $0=$2; print $1; }')
}

getClusterIP() {
    SERVICE=$1
    CLUSTER_IP=$(kubectl -n k8scenario get service/$SERVICE | awk '/:/ { print $3; }')
}

getPodIP() {
    POD_MATCH=""
    [ ! -z "$1" ] && POD_MATCH="/$1"
    POD_IP=$(kubectl -n k8scenario get pod${POD_MATCH} --no-headers -o wide | awk '{ print $6; exit(0); }')
}

getANodeIPForRunningPod() {
    POD_MATCH=""
    [ ! -z "$1" ] && POD_MATCH="/$1"
    # Any node would do: but let's take one on which our Pod is running on
    NODE_IP=$(kubectl -n k8scenario get pod${POD_MATCH} --no-headers -o wide | awk '{ print $7; exit(0); }')
}

getANodeIP() {
    getANodeIPForRunningPod
}

set -x

getANodeIP
getServicePort flask-app

EP=${NODE_IP}:${PORT}

echo "ENDPOINT=$EP"
kubectl run --restart=Never --rm -n k8scenario -it --generator=run-pod/v1 --image=alpine:latest ${NS}-test -- /bin/sh -c "wget -O - --timeout 4 $EP; exit \$?"



