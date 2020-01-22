
set -x
# Check 'critical*' Pod is Running
# Check sleep has been changed to other than 'sleep 10'
kubectl get pods -o yaml | grep -qE "'sleep *10([^\d]|$)"
if [ $? -eq 0 ];then
    exit 1
else
    exit 0
fi


#CHECK_PODS_PREFIXED "critical" && 
#    kubectl get pods -o yaml | grep -E "'sleep *10([^\d]|$)" &&
#        exit 1

