#!/bin/bash

DATE_VERSION=$(date +%Y-%b-%d_%02Hh%02Mm%02S)

# Disabled for now as causing error message on github action build:
# - https://github.com/k8scenario/k8scenario/runs/1713449601?check_suite_focus=true
#    can't load package: build constraints exclude all Go files
#    - https://github.com/golang/go/issues/24433
#    - https://github.com/golang/go/issues/24068

# export CGO_ENABLED=0 

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

    time go build -a -o bin/$K8SCENARIO_BINARY k8scenario.go || exit 1
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

