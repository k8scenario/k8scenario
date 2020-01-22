#!/bin/bash

# Usage:
#
# - To run a particular scenario, provide it's number:
#     $0 <num>
#   e.g.
#     $0 21
#
# - To test all scenarios:
#   In 2 separate consiles:
#     - 1st run all scenarios in a loop: waiting for each to be fixed:
#       (to be used with -f option below)
#         $0 -a
#
#     - To fix the current scenario (stored in $TMP_TOFIX):
#       (to be used with -a option above)
#         $0 -f -a
#

# DEFAULT ACTION: launch each scenarii:
[ ! "$1" ] && set -- -a

TMP_TOFIX=tmp/.tofix
NEW_ALL=new.scenarii.txt

OPTS=""
TEST_ALL=""
FIX_ALL=0
DO_REGRESSION_TESTS=0

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

RUN_REGRESSION_TESTS() {
    NS=k8scenario

    for SCENARIO in $TEST_ALL; do
        echo "Test scenario $SCENARIO - regenerating zip file"
        REBUILD_SCENARIO_ZIP $SCENARIO
	ls -altr SCENARII/scenario${SCENARIO}.zip

	echo; press "more SCENARII/scenario${SCENARIO}/EXCLUDE_SOLUTION.txt"
	more SCENARII/scenario${SCENARIO}/EXCLUDE_SOLUTION.txt

	echo; press "About to setup scenario $SCENARIO [will (delete/re)create namespace $NS]"
        kubectl get ns | grep -q $NS && kubectl delete ns $NS
        kubectl create ns $NS

	# TODO: pre/post setup !!
        scripts/setup_check_fix_scenario.sh --setup $SCENARIO

	echo; press "About to check-broken scenario $SCENARIO"
        scripts/setup_check_fix_scenario.sh --check-broken $SCENARIO

	echo; press "About to fix scenario $SCENARIO"
        scripts/setup_check_fix_scenario.sh --fix $SCENARIO

	echo; press "About to check-fixed scenario $SCENARIO"
        scripts/setup_check_fix_scenario.sh --check $SCENARIO
    done
}

GET_ALL_SCENARII() {
    TEST_ALL=$(ls -1d SCENARII/scenario*/ | sed -e 's/.*scenario//' -e 's?/??')
    echo "TEST_ALL=<"$TEST_ALL">"
}

APPLY_SCENARII() {
    TEST_ALL="$*"

    for SCENARIO in $TEST_ALL; do
        echo "$SCENARIO" > $TMP_TOFIX
        $0 $SCENARIO
        echo "" > $TMP_TOFIX
    done
}

FIX_SCENARII() {
    TEST_ALL="$*"

    while true; do
        while [ ! -f $TMP_TOFIX ]; do echo "Waiting for $TMP_TOFIX to appear"; sleep 1; done

        while [ -f $TMP_TOFIX ]; do
            NS_STATUS=$(kubectl get ns/k8scenario -o custom-columns=STATUS:.metadata.labels.status --no-headers)
            NS_SCENARIO=$(kubectl get ns/k8scenario -o custom-columns=STATUS:.metadata.labels.scenario --no-headers)

	    # Skip this loop if no labels set yet on namespace:
            [ "$NS_STATUS" = "<none>" ] && continue
            [ "$NS_SCENARIO" = "<none>" ] && continue

            SCENARIO=$(cat $TMP_TOFIX)
            [ "$SCENARIO" = "" ] && continue

            [ "scenario$SCENARIO" != "$NS_SCENARIO" ] && die "Expected scenario$SCENARIO, but Namespace has labels{status=$NS_STATUS,scenario=$NS_SCENARIO}"
            echo "Namespace has labels{status=$NS_STATUS,scenario=$NS_SCENARIO}"

            if [ "$NS_STATUS" = "readytofix" ]; then
                echo
                echo "./scripts/setup_check_fix_scenario.sh -f $SCENARIO"
                ./scripts/setup_check_fix_scenario.sh -f $SCENARIO
                echo "fix for scenario$SCENARIO applied"
		echo
                press "Moving to next scenario ..."
            else
                echo "waiting for namespace status <$NS_STATUS>"
                sleep 1
            fi

        done
    done
    rm $TMP_TOFIX
}

VALIDATE_ALL_SCENARII_YAML() {
    local SCENARIO_DIR=SCENARII/
    find $SCENARIO_DIR/ -maxdepth 2 -iname '*.y*ml' -exec kubeval {} \; | grep -v valid && die "Yaml validation failed"
}

VALIDATE_SCENARIO_YAML() {
    local SCENARIO_DIR=$1; shift
    find $SCENARIO_DIR/ -maxdepth 0 -iname '*.y*ml' -exec kubeval {} \; | grep -v valid && die "Yaml validation failed"
}

REBUILD_SCENARIO_ZIP() {
    SCENARIO=${1}
    TEMPLATE_DIR=SCENARII/TEMPLATE
    SCENARIO_DIR=SCENARII/scenario${SCENARIO}
    SCENARIO_ZIP=${SCENARIO_DIR}.zip

    VALIDATE_SCENARIO_YAML $SCENARIO_DIR
    #set -x
    [ -f ${SCENARIO_ZIP} ] && rm ${SCENARIO_ZIP}
    cp -a $TEMPLATE_DIR/functions.rc ${SCENARIO_DIR}/.functions.rc
    zip -r9 ${SCENARIO_ZIP} ${SCENARIO_DIR}/ -x '*/EXCLUDE_*' -x '*/.EXCLUDE*'
    #set +x

    # Remove again - don't want to archive
    rm ${SCENARIO_DIR}/.functions.rc
}

while [ ! -z "$1" ]; do
    #echo case $1
    case $1 in
        [0-9]*)
            REBUILD_SCENARIO_ZIP $1
            ;;

        #-x) SET_X="set -x";;
        -d|--debug) OPTS+="--debug" ;;
        -k|--keepns) OPTS+="--keepns" ;;

        -r|-rt|-nrt|--rt|--nrt)
		TEST_ALL="ALL";
		DO_REGRESSION_TESTS=1
		[ ! -z "$2" ] && { shift; TEST_ALL="$*"; set -- DUMMY_ARG; }
		;;

        -f|--fix) FIX_ALL=1;;
        -na|--new-all) TEST_ALL=$(cat $NEW_ALL);;
        -a|--all) TEST_ALL="ALL";
		[ ! -z "$2" ] && {
		    shift
		    [ "$1" == "-f" ] && die "Use option -f before -a"
		    [ "$1" == "--fix" ] && die "Use option --fix before -a"
		    TEST_ALL="$*"
		    set -- DUMMY_ARG;
	        }
		;;

        *)
            [ -f "$1" ] && { SCENARIO_ZIP=$1; break; }
    
            [ -d "$1" ] && die "TODO: dir handling"
    
            die "Unknown option <$1>"
        ;;
    esac
    shift
done

#echo "TEST_ALL=<$TEST_ALL>"
[ "$TEST_ALL" == "ALL" ] && GET_ALL_SCENARII

[ $DO_REGRESSION_TESTS -ne 0 ] && { RUN_REGRESSION_TESTS; exit $?; }

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

if [ $FIX_ALL -ne 0 ];then
    # Fix each of the scenarii:
    FIX_SCENARII $TEST_ALL
    exit 0
fi

#./bin/k8scenario.public --keepns --zip SCENARII/scenario21.zip 
[ -z "$SCENARIO_ZIP" ]   && die "No scenario file specified"
[ ! -f "$SCENARIO_ZIP" ] && die "No such scenario file <$SCENARIO_ZIP>"

echo "$SCENARIO" > $TMP_TOFIX
#echo "\$SCENARIO=$(cat $TMP_TOFIX)"

set -x
#./bin/k8scenario.public $OPTS --zip $SCENARIO_ZIP
export TESTMODE_VERBOSE=1
./bin/k8scenario.private $OPTS --zip $SCENARIO_ZIP
#./bin/k8scenario.private $OPTS --menu --dir $SCENARIO_ZIP

