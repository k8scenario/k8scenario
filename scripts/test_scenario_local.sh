#!/bin/bash

# Usage:
#
# - To run a particular scenario, provide it's number:
#     $0 <num>
#   e.g.
#     $0 21
#
# - To run all scenarios in a loop: waiting for each to be fixed:
#   (to be used with -f option below)
#     $0 -a
#
# - To fix the current scenario (stored in tmp/.tofix):
#   (to be used with -a option above)
#     $0 -f -a
#

OPTS=""
TEST_ALL=""
FIX_ALL=0

#SET_X=""

die() {
    echo "die: $0 - $*" >&2
    exit 1
}

press() {
    echo "$*"
    echo "Press <enter> to continue"
    read DUMMY
    [ "$DUMMY" = "q" ] && exit 0
    [ "$DUMMY" = "Q" ] && exit 0
}

while [ ! -z "$1" ]; do
    #echo case $1
    case $1 in
        [0-9]*)
            scenario=scenario${1}
            SCENARIO_ZIP=SCENARII/${scenario}.zip
            #set -x
            [ -f ${SCENARIO_ZIP} ] && rm ${SCENARIO_ZIP}
            cp -a SCENARII/TEMPLATE/functions.rc SCENARII/${scenario}/.functions.rc
            zip -r9 ${SCENARIO_ZIP} SCENARII/${scenario}/ -x '*/EXCLUDE_*'
            #set +x

	    # Remove again - don't want to archive
            rm SCENARII/${scenario}/.functions.rc
            ;;

        #-x) SET_X="set -x";;
        -d|--debug) OPTS+="--debug" ;;
        -k|--keepns) OPTS+="--keepns" ;;

	ALL_IGNORE) ;;
        -f|--fix) FIX_ALL=1;;
        -a|--all) TEST_ALL="ALL";
		[ ! -z "$2" ] && {
		    shift
		    [ "$1" == "-f" ] && die "Use option -f before -a"
		    [ "$1" == "--fix" ] && die "Use option --fix before -a"
		    TEST_ALL="$*"
		    set -- ALL_IGNORE
	        };;

        *)
            [ -f "$1" ] && { SCENARIO_ZIP=$1; break; }
    
            [ -d "$1" ] && die "TODO: dir handling"
    
            die "Unknown option <$1>"
        ;;
    esac
    shift
done

GET_ALL_SCENARII() {
    TEST_ALL=$(ls -1d SCENARII/scenario*/ | sed -e 's/.*scenario//' -e 's?/??')
    echo "TEST_ALL=<"$TEST_ALL">"
}

APPLY_SCENARII() {
    TEST_ALL="$*"

    for SCENARIO in $TEST_ALL; do
        echo "$SCENARIO" > tmp/.tofix
        $0 $SCENARIO
        echo "" > tmp/.tofix
    done
}

FIX_SCENARII() {
    TEST_ALL="$*"

    while true; do
        while [ ! -f tmp/.tofix ]; do echo "Waiting for tmp/.tofix to appear"; sleep 1; done

        while [ -f tmp/.tofix ]; do
            NS_STATUS=$(kubectl get ns/k8scenario -o custom-columns=STATUS:.metadata.labels.status --no-headers)
            NS_SCENARIO=$(kubectl get ns/k8scenario -o custom-columns=STATUS:.metadata.labels.scenario --no-headers)

	    # Skip this loop if no labels set yet on namespace:
            [ "$NS_STATUS" = "<none>" ] && continue
            [ "$NS_SCENARIO" = "<none>" ] && continue

            SCENARIO=$(cat tmp/.tofix)
            [ "$SCENARIO" = "" ] && continue

            [ "scenario$SCENARIO" != "$NS_SCENARIO" ] && die "Expected scenario$SCENARIO, but Namespace has labels{status=$NS_STATUS,scenario=$NS_SCENARIO}"
            echo "Namespace has labels{status=$NS_STATUS,scenario=$NS_SCENARIO}"

            if [ "$NS_STATUS" = "readytofix" ]; then
                echo
                echo "./scripts/setup_check_fix_scenario.sh -f $SCENARIO"
                ./scripts/setup_check_fix_scenario.sh -f $SCENARIO
                press ""
            else
                echo "waiting for namespace status <$NS_STATUS>"
                sleep 1
            fi

        done
    done
    rm tmp/.tofix
}

#echo "TEST_ALL=<$TEST_ALL>"
[ "$TEST_ALL" == "ALL" ] && GET_ALL_SCENARII

if [ ! -z "$TEST_ALL" ]; then
    #echo "FIX_ALL=<$FIX_ALL>"
    if [ $FIX_ALL -eq 0 ];then
        # Setup each scenario and loop until fixed:
        APPLY_SCENARII $TEST_ALL
    else
        # Fix each of the scenarii:
        FIX_SCENARII $TEST_ALL
    fi
    exit 0
fi

#./bin/k8scenario.public --keepns --zip SCENARII/scenario21.zip 
[ -z "$SCENARIO_ZIP" ]   && die "No scenario file specified"
[ ! -f "$SCENARIO_ZIP" ] && die "No such scenario file <$SCENARIO_ZIP>"

set -x
#./bin/k8scenario.public $OPTS --zip $SCENARIO_ZIP
./bin/k8scenario.private $OPTS --zip $SCENARIO_ZIP
#./bin/k8scenario.private $OPTS --menu --dir $SCENARIO_ZIP

