#!/bin/bash

DATE_VERSION=$(date +%Y-%b-%d_%02Hh%02Mm%02S)

PROMPTS=0

function die {
    echo "$0: die - $*" >&2
    exit 1
}

function press {
    echo $*
    echo "Press <return>"
    read DUMMY
    [ "$DUMMY" = "q" ] && exit 0
    [ "$DUMMY" = "Q" ] && exit 0
}

function build {
    # builds dynamic binary:
    # go build -o k8scenario k8scenario.go || exit 1
    # builds static binary:

    set -x
    sed -i.bak \
	    -e "s/__K8SCENARIO_VERSION__.*=.*/__K8SCENARIO_VERSION__=\"$K8SCENARIO_BINARY\"/g" \
	    -e "s/__DATE_VERSION__.*=.*/__DATE_VERSION__=\"$DATE_VERSION\"/g" \
	    -e "s?__DEFAULT_PUBURL__.*=.*?__DEFAULT_PUBURL__=\"$DEFAULT_PUBURL\"?g" \
	    k8scenario.go 

    time CGO_ENABLED=0 go build -a -o bin/$K8SCENARIO_BINARY k8scenario.go || exit 1
}

function install_build {
    mkdir -p BAK/
    ls -altrh bin/k8scenario 
    cp -a k8scenario.go BAK/k8scenario.go.$DATE_VERSION

    if [ -d ~/usr/bin_lin64/ ]; then
        cp -a bin/$K8SCENARIO_BINARY ~/usr/bin_lin64/${K8SCENARIO_BINARY}.$DATE_VERSION

        ls -altrh ~/usr/bin_lin64/${K8SCENARIO_BINARY}.$DATE_VERSION

        # Copy file from symlink: (no -a):
        cp  /home/mjb/usr/bin_lin64/kubectl ./bin/
    else
        if [ -d ~/BINARIES/bin_lin64/ ]; then
            cp -a bin/$K8SCENARIO_BINARY ~/BINARIES/bin_lin64/${K8SCENARIO_BINARY}.$DATE_VERSION
            ls -altrh ~/BINARIES/bin_lin64/${K8SCENARIO_BINARY}.$DATE_VERSION

            # Copy file from symlink: (no -a):
            cp  /home/mjb/BINARIES/bin_lin64/kubectl ./bin/
        fi
    fi
}

function docker_build_and_push {

    [ $PROMPTS -ne 0 ] && press "About to build docker image"


    docker build -t mjbright/k8scenario:latest .

    if [ -z "$TAG" ];then
        TAG="latest"
        docker push mjbright/k8scenario:latest
    else
        TAG="latest"
        docker tag mjbright/k8scenario:latest mjbright/k8scenario:$TAG
        docker push mjbright/k8scenario:latest
        docker push mjbright/k8scenario:$TAG
    fi
}

function CLEAN_k8scenario {
    echo "---- Deleting k8scenario pod if present:"
    kubectl get pod |& grep k8scenario && kubectl delete pod k8scenario

    echo "---- Deleting scenario namespaces if present:"
    NAMESPACES=$(kubectl get ns |& grep ^scenario | awk '{ print $1; }')
    if [ ! -z "$NAMESPACES" ]; then
        for NAMESPACE in $NAMESPACES; do
            echo kubectl delete ns/$NAMESPACE
            kubectl delete ns/$NAMESPACE
        done
    fi

    #kubectl get ns |& grep ^scenario && kubectl delete ns $(kubectl get ns | grep ^scenario | awk '{ print $1; }')
}

function run_k8scenario {
    sed "s/__TAG__/$TAG/g" < k8scenario.template.yaml > k8scenario.yaml
    sed "s/__TAG__/$TAG/g" < k8scenario_menu.template.yaml > k8scenario_menu.yaml

    # CLEANUP:
    CLEAN_k8scenario

    #kubectl create -f k8scenario.yaml
    kubectl create -f k8scenario_menu.yaml

    while kubectl get pods k8scenario | grep Running; do
	echo "Waiting for pod/k8scenario to be Running"
        sleep 2
        kubectl get pods k8scenario
    done
    kubectl get pods k8scenario

    #while ! kubectl logs pod/k8scenario | grep -v "Detect doesn't exists"; do
    #done
    echo "Waiting for pod/k8scenario logs ..."
    sleep 5
    kubectl logs pod/k8scenario

    while true; do
        echo "---- $(date) ----------------------"
        LAST_SCENARIO=$(kubectl get ns | awk '/^scenario/ { print $1; }' | tail -1)
        kubectl -n $LAST_SCENARIO get all
	sleep 5
    done
}

[ ! -f .setup.rc ] && die "No .setup.rc in $PWD"
. .setup.rc

echo "BUILDING $K8SCENARIO_BINARY for <$DEFAULT_PUBURL>"

[ -z "$DEFAULT_PUBURL" ] && die "\$DEFAULT_PUBURL not set in .setup.rc"
[ -z "$DEFAULT_PUBDIR" ] && die "\$DEFAULT_PUBDIR not set in .setup.rc"
[ ! -d "$DEFAULT_PUBDIR" ] && die "No such \$DEFAULT_PUBDIR dir as <$DEFAULT_PUBDIR>"

# Add TAG:
TAG=""

while [ ! -z "$1" ]; do
    case $1 in
        -k8s|-run)
	      run_k8scenario; exit $?;;
        *)    TAG=$1;;
    esac
    shift
done

build
install_build
#docker_build_and_push

if [ $PROMPTS -ne 0 ]; then
    press "About to run k8scenario"

    run_k8scenario
fi



