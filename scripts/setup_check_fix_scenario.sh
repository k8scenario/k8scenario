#!/bin/bash

cd $(dirname $0)/..

export NS=k8scenario

ACTION="check"
ACTION_SCRIPT="CHECK_SCENARIO.sh"

SCENARIO=""

RED='\e[00;31m';      B_RED='\e[01;31m';      BG_RED='\e[07;31m'
GREEN='\e[00;32m';    B_GREEN='\e[01;32m'     BG_GREEN='\e[07;32m'
YELLOW='\e[00;33m';   B_YELLOW='\e[01;33m'    BG_YELLOW='\e[07;33m'
NORMAL='\e[00m'

die() {
    echo -e "$0: ${RED}die - $*${NORMAL}" >&2
    exit 1
}

OK() { echo -e "${GREEN}${*}${NORMAL}"; }

ERROR() { echo -e "${RED}${*}${NORMAL}"; }

WARN() { echo -e "${YELLOW}${*}${NORMAL}"; }

RUN_SCRIPT_CHECK_EXPECTED_RETURN_CODE() {
    SCRIPT_PLUS_ARGS="$1"; shift
    EXPECTED_RET=$1; shift

    $SCRIPT_PLUS_ARGS; RET=$?
    echo -n "$SCRIPT_PLUS_ARGS ==> "
    if [ $RET -eq $EXPECTED_RET ];then
        echo -e "${GREEN}GOOD  ==>${NORMAL} '$ACTION' returned exit code $RET"
    else
        echo -e "${RED}ERROR ==>${NORMAL} '$ACTION' returned exit code $RET"
    fi
}

CREATE_AND_RUN_TMP_SCRIPT() {
    ACTION_SCRIPT=$1; shift
    ACTION=$1;        shift

    [ ! -f "$ACTION_SCRIPT" ] && die "[$PWD] No such '$ACTION' script <$ACTION_SCRIPT>"
    [ ! -x "$ACTION_SCRIPT" ] && die "[$PWD] '$ACTION' script <$ACTION_SCRIPT> is not executable"

    TMP_SH=tmp/${ACTION}_scenario.sh
    cat > $TMP_SH << EOF
#!/bin/bash

$SET_X

export NS=$NS

EOF

    EXPECTED_RET=0
    case $ACTION in
        setup|check|fix) EXPECTED_RET=0;;
        check-broken)    EXPECTED_RET=1;;
    esac

    cat SCENARII/TEMPLATE/functions.rc $ACTION_SCRIPT >> $TMP_SH
    chmod +x $TMP_SH

    echo "$TMP_SH"
    case $ACTION in
        setup)
           RUN_SCRIPT_CHECK_EXPECTED_RETURN_CODE "$TMP_SH --pre-yaml" $EXPECTED_RET

	   NUM_YAML=$(find SCENARII/scenario${SCENARIO}/ -iname '*.y*ml' | wc -l)
	   [ $NUM_YAML -gt 0 ] && {
	       CMD="kubectl -n $NS create -f SCENARII/scenario${SCENARIO}/"
	       echo "-- $CMD [ $NUM_YAML files ]"
	       $CMD
           }

           RUN_SCRIPT_CHECK_EXPECTED_RETURN_CODE "$TMP_SH --post-yaml" $EXPECTED_RET
	   ;;

        *) RUN_SCRIPT_CHECK_EXPECTED_RETURN_CODE $TMP_SH $EXPECTED_RET;;
    esac

    [ "$ACTION" = "setup" ] && { kubectl get all -n $NS; }

    [ -z "$SCENARIO" ] && return

    [ "$ACTION" = "fix" ] && {
        echo
        echo "==== Relaunching script in check node:"

        echo "$0 --check $SCENARIO"
        $0 --check $SCENARIO
    }
}

SET_X=""

while [ ! -z "$1" ]; do
  case $1 in
    -x) SET_X="set -x";;
    -f|--fix)
        ACTION="fix"
        ACTION_SCRIPT="EXCLUDE_FIX_SCENARIO.sh"
	;;
    -c|--check)
        ACTION="check"
        ACTION_SCRIPT="CHECK_SCENARIO.sh"
	;;
    -b|--check-broken)
        ACTION="check-broken"
        ACTION_SCRIPT="CHECK_SCENARIO.sh"
	;;
    -s|--setup)
        ACTION="setup"
        ACTION_SCRIPT="SETUP_SCENARIO.sh"
	;;
    [0-9]*)
        SCENARIO=$1
	;;
    *)
        [ ! -f "$1" ] && die "Unknown argument <$1>"

	ACTION_SCRIPT="$1"
        [ -z "$ACTION" ] && ACTION="check"
        CREATE_AND_RUN_TMP_SCRIPT $ACTION_SCRIPT $ACTION
	exit $?
        ;;
  esac
  shift
done

[ -z "$SCENARIO" ] && die "Missing scenario number: 
    Usage:
        $0 [--check|--fix|--setup] <num>
"

[ ! -d "SCENARII/scenario$SCENARIO" ] && die "No such dir <SCENARII/scenario$SCENARIO>"

ACTION_SCRIPT="SCENARII/scenario$SCENARIO/$ACTION_SCRIPT"

CREATE_AND_RUN_TMP_SCRIPT $ACTION_SCRIPT $ACTION

