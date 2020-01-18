
# TODO:
# - build zips with this before yaml files (so will be read earlier)
# - invoke initially with no argument (before applying yaml files)
# - invoke a second time with '-post' argument (after applying yaml files)

# Setup steps if any prior to applying yaml files:
if [ "$1" == "--pre-yaml" ];then
    exit 0
fi

# Setup steps if any after applying yaml files:
if [ "$1" == "--post-yaml" ];then
    # Just to warn if CONTEXT is not k8scenario: (if not set user commands must include '-n k8scenario')
    CHECK_CONTEXT
    exit 0
fi

exit 0

