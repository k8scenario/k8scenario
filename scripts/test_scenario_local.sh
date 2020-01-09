#!/bin/bash

OPTS=""

die() {
    echo "die: $0 - $*" >&2
    exit 1
}

while [ ! -z "$1" ]; do
	echo case $1
    case $1 in
        [0-9]*) SCENARIO=
            SCENARIO_FILE=SCENARII/scenario${1}.zip
            # SCENARIO_FILE=SCENARII/scenario${1}/ ????
            ;;

        -k) OPTS+="--keepns"
	    ;;

	*)
	    [ -f "$1" ] && { SCENARIO_FILE=$1; break; }

	    [ -d "$1" ] && die "TODO: dir handling"

	    die "Unknown option <$1>"
	    ;;
    esac
    shift
done

#./bin/k8scenario.public --keepns --zip SCENARII/scenario21.zip 
[ -z "$SCENARIO_FILE" ]   && die "No scenario file specified"
[ ! -f "$SCENARIO_FILE" ] && die "No such scenario file <$SCENARIO_FILE>"

set -x
./bin/k8scenario.public $OPTS --zip $SCENARIO_FILE


