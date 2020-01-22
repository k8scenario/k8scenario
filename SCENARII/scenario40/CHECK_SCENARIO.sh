
set -x

IP=$(GET_SVC_CLUSTERIP_IP flask-app)
PORT=$(GET_SVC_CLUSTERIP_PORT flask-app)

EP=${IP}:${PORT}

echo "ENDPOINT=$EP"

# NOTE: Because of interactivity problems (seen with Kube 1.17): we use --restart=Never
#kubectl run --restart=Never --rm -it --generator=run-pod/v1 --image=alpine:latest ${NS}-test -- /bin/sh -c "wget -O - --timeout 1 $EP; exit \$?"
#kubectl run --restart=Never --rm -it --generator=run-pod/v1 --image=alpine:latest $TESTPOD -- /bin/sh -c "wget -qO - --timeout 1 $EP"
# <v1>[flask-app-696fb4674b-cch8w] Redis counter value=93

#kubectl logs tester | grep counter && return 0

TEST_POD_SHELL "wget -qO - --timeout 1 $EP"



