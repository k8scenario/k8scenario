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
    set -x; cp -a bin/$K8SCENARIO_BINARY $COPY_K8SCENARIO_TO; set +x
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
install_build

