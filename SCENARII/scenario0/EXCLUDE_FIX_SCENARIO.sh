# Note: any image will do

# Fix by running a Pod directly:
kubectl -n k8scenario run --generator=run-pod/v1 --image=mjbright/ckad-demo:1 basictest-xxx

# Fix by creating a Deployment
kubectl -n k8scenario create deploy --image=mjbright/ckad-demo:1 basictest 



