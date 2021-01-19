#!/bin/bash

VERSION=$(cat version | sed 's/:.*//')
DATE_VERSION=$(date +%Y-%b-%d_%02Hh%02Mm%02S)
BUILD_VERSION="public"

# Unneeded?
export CGO_ENABLED=0 

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

    VERSION_INFO="--ldflags=\"-X 'main.BUILD_TIME=$DATE_VERSION' \
                              -X 'main.BUILD_VERSION=${BUILD_VERSION}' \
                              -X 'main.VERSION=${VERSION}' \
                              -X 'main.DEFAULT_URL=$DEFAULT_PUBURL'\""
    VERSION_INFO=$(echo $VERSION_INFO | sed 's/  */ /g')
    echo "VERSION_INFO=$VERSION_INFO"

    # Use eval to be able to treat quotes in $VERSION_INFO:
    CMD="time CGO_ENABLED=0 go build $VERSION_INFO -a -o bin/$K8SCENARIO_BINARY k8scenario.go"
    echo "---- $CMD"
    eval $CMD || exit 1
    #time go build -a -o bin/$K8SCENARIO_BINARY k8scenario.go || exit 1

    set -x; cp -a bin/$K8SCENARIO_BINARY $COPY_K8SCENARIO_TO; set +x
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
        *)    TAG=$1;;
    esac
    shift
done

build

