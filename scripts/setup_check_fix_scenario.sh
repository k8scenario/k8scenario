#!/bin/bash

cd $(dirname $0)/..

export NS=k8scenario

ACTION="check"
ACTION_SCRIPT="CHECK_SCENARIO.sh"

SCENARIO=""

die() {
    echo "$0: die - $*" >&2
    exit 1
}

CREATE_AND_RUN_TMP_SCRIPT() {
    ACTION_SCRIPT=$1; shift
    ACTION=$1;        shift

    [ ! -x "$ACTION_SCRIPT" ] && die "[$PWD] No such '$ACTION' script <$ACTION_SCRIPT>"

    TMP_SH=tmp/${ACTION}_scenario.sh
    cat > $TMP_SH << EOF
#!/bin/bash

$SET_X

export NS=$NS

EOF

    cat SCENARII/TEMPLATE/functions.rc $ACTION_SCRIPT >> $TMP_SH
    chmod +x $TMP_SH

    echo "$TMP_SH"
    $TMP_SH
    echo "==> returned exit code $?"

    [ "$ACTION" = "setup" ] && { kubectl get all -n $NS; }

    [ -z "$SCENARIO" ] && return
    echo "SCENARIO=<$SCENARIO>"

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

