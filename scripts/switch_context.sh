#!/bin/bash

SAVED=~/.k8scenario.saved.context

# USAGE: switch_context.sh
# - if current context is 'k8scenario' AND file $SAVED (above) is present, put back the saved context
# - else
#   - create k8scenario context using namespace k8scenario (assumes KIND cluster)
#   - set context to k8scenario


use_k8scenario_context() {
    kubectl config set-context k8scenario --cluster kind-kind --user kind-kind --namespace=k8scenario
    kubectl config use-context k8scenario
}

CONTEXT=$(kubectl config get-contexts | awk '/^* / { print $2; }')
echo "Current context is <$CONTEXT>"

if [ ! -z "$CONTEXT" ]; then
    if [ "$CONTEXT" = "k8scenario" ]; then
        [ -f $SAVED ] && kubectl config use-context $(cat $SAVED)
    else
        echo $CONTEXT > $SAVED
        use_k8scenario_context
    fi
fi

CONTEXT=$(kubectl config get-contexts | awk '/^* / { print $2; }')
echo "Current context is <$CONTEXT>"

