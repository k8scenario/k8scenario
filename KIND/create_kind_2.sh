

cd $(dirname $0)

kind delete cluster --name kind
kind create cluster --config kind_2.yaml 
kind get clusters


