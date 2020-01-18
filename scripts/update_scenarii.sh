#!/bin/bash

. .setup.rc

BASE_DIR=$(readlink -f $(dirname $0)/..)
SCENARII_DIR=$(dirname $0)/../SCENARII

URL_BASE=$DEFAULT_PUBURL
PUB_URL_DIR=$DEFAULT_PUBDIR
URL_SCENARII_DIR=$PUB_URL_DIR/static/k8scenarii
URL_BIN_DIR=$PUB_URL_DIR/static/bin
#REPO_DIR=$DEFAULT_PUBDIR
#PUB_URL_DIR=$REPO_DIR/static/k8scenarii/

WEB_UPLOAD=1
WEB_UPLOAD=0

die() {
    echo "$0: die - $*" >&2
    exit 1
}

create_zips() {
    #cd $SCENARII_DIR/static/k8scenarii
    cd $SCENARII_DIR/
    echo "[$PWD] Creating zips ..."

    for scenario in scenario[0-9]*/; do
        scenario=${scenario%/}

        # Skip if not scenario[0-9]+:
        [ ${scenario%[0-9]} = $scenario ] && continue

        create_zip $scenario
    done

    ls -altr *.zip

    #cd -
    cd $BASE_DIR
    echo "[$PWD] Creating zips ... DONE"
    echo "Zips copied to $PUB_URL_DIR/"
}

create_zip() {
    [ ! -d $scenario ] && die "Expected directory <$scenario> in $SCENARII_DIR"

    cp -a TEMPLATE/functions.rc $scenario/.functions.rc

#set -x
    [ -f ${scenario}.zip ] && rm -f ${scenario}.zip
    zip -r9 ${scenario}.zip ${scenario}/ -x '*/*EXCLUDE_*' 2>&1 >/dev/null

    # Remove - don't want it in the .git archive
    rm $scenario/.functions.rc

    [ ! -f ${scenario}.zip ] && die "Failed to create zip <${scenario}.zip>"
    [ ! -d $URL_SCENARII_DIR ] && mkdir -p $URL_SCENARII_DIR
    [ ! -d $URL_BIN_DIR ]      && mkdir -p $URL_BIN_DIR
    cp -a ${scenario}.zip index.list $URL_SCENARII_DIR/
    cp -a ../bin/$K8SCENARIO_BINARY $URL_BIN_DIR/k8scenario
#set +x
}

upload_zips() {
    #cd $REPO_DIR
    cd $PUB_URL_DIR
    echo "[$PWD] Uploading zips to github ..."

    #git add static/k8scenarii/
    git add .
    git commit -m "Adding latest k8scenarii"
    git push

    #cd -
    cd $BASE_DIR
    echo "[$PWD] Uploading zips to github ... DONE"
}

#DOWNLOAD="wget --no-check-certificate --no-cache --no-cookies --post-data='action=purge'"
DOWNLOAD="wget --no-check-certificate --no-cache --no-cookies -q -O -"
DOWNLOAD="curl -sL -o -"

check_zips() {
    cd $SCENARII_DIR

    for scenario in scenario*/; do
        scenario=${scenario%/}

        # Skip if not scenario[0-9]+:
        [ ${scenario%[0-9]} = $scenario ] && continue

        [ ! -d $scenario ] && die "Expected directory <$scenario> in $SCENARII_DIR"

        cksum=$(cksum < ${scenario}.zip)
        URL=$URL_BASE/${scenario}.zip

	echo; echo "==== Getting zip file from $URL"
        wcksum=$($DOWNLOAD $URL | cksum)
        while [ "$cksum" != "$wcksum" ]; do
            SLEEP=10
            echo "Sleeping $SLEEP secs [waiting for ${scenario}.zip file to be updated <local> $cksum != <online> $wcksum]"
	    sleep $SLEEP
            wcksum=$($DOWNLOAD $URL | cksum)
	done
        echo "$scenario: cksum OK $cksum == $wcksum [$URL]"

        #if [ "$cksum" != "$wcksum" ]; then
        #    echo "$scenario: $cksum != $wcksum [$URL]"
        #else
        #    echo "$scenario: cksum OK $cksum == $wcksum [$URL]"
        #fi
    done

    #cd -
    cd $BASE_DIR
}

function rebuild_index {
    cd $SCENARII_DIR
    echo "[$PWD] Rebuilding index"

    cp /dev/null index.list
    for scenario in scenario[0-9]*/; do
        scenario=${scenario%/}

        # Skip if not scenario[0-9]+:
        [ ${scenario%[0-9]} = $scenario ] && continue

        echo "# $scenario: "
        [ -f ${scenario}.fix/README.txt ] && cat ${scenario}.fix/README.txt
	NUM=${scenario#scenario}
	NUM=${NUM%/}
	echo ${NUM} >> index.list
    done > INDEX.md

    #cd -
    cd $BASE_DIR
    echo "[$PWD] Rebuilding index ... DONE"
}

[ ! -d $SCENARII_DIR ] && die "No such scenario dir <$SCENARII_DIR>"

[ ! -f .setup.rc ] && die "No .setup.rc in $PWD"
. .setup.rc

[ -z "$DEFAULT_PUBURL" ] && die "\$DEFAULT_PUBURL not set in .setup.rc"
[ -z "$DEFAULT_PUBDIR" ] && die "\$DEFAULT_PUBDIR not set in .setup.rc"
[ ! -d "$DEFAULT_PUBDIR" ] && die "No such \$DEFAULT_PUBDIR dir as <$DEFAULT_PUBDIR>"

while [ ! -z "$1" ]; do
    case $1 in
	# Run server as:
	#    ./SCENARII/local_server.sh
        -l|--local) URL_BASE=http://127.0.0.1:9000;;

        -u|--up*|--pub|-pub)       WEB_UPLOAD=1;;
        -c|--check) check_zips;    RET=$?; echo "Exiting ... $RET"; exit $RET;;
        -i|--index) rebuild_index; RET=$?; echo "Exiting ... $RET"; exit $RET;;

    esac
    shift
done

#rebuild_index
create_zips

if [ $WEB_UPLOAD -ne 0 ];then
    upload_zips
    check_zips
fi
#exit 0

#exit 0



