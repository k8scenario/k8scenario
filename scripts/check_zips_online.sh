#!/bin/bash

. .setup.rc

VERBOSE=0

press() {
    echo $*
    echo "Press <enter>"
    read DUMMY
    [ "$DUMMY" = "q" ] && exit 0
    [ "$DUMMY" = "Q" ] && exit 0
}

echo "Using DEFAULT_PUBDIR=$DEFAULT_PUBDIR"
echo "Using DEFAULT_PUBURL=$DEFAULT_PUBURL"


for zip in SCENARII/scenario*.zip; do
    zipfile=$(basename $zip)
    [ $VERBOSE -ne 0 ] && echo; echo "-- $zipfile"
    #press $zipfile

    LOCAL_CKSUM=$(cksum - < $zip)
    [ $VERBOSE -ne 0 ] && echo "LOCAL_CKSUM=$LOCAL_CKSUM"

    REPO_ZIP=$DEFAULT_PUBDIR/static/k8scenarii/$zipfile
    REPO_CKSUM=$(cksum - < $REPO_ZIP)
    [ $VERBOSE -ne 0 ] && echo "REPO_CKSUM=$REPO_CKSUM"
    [ "$REPO_CKSUM" != "$LOCAL_CKSUM" ] && {
      echo "REPO_CKSUM='$REPO_CKSUM' != LOCAL_CKSUM='$LOCAL_CKSUM'"
      press "Skipping web check"
      continue
    }

    #URL=$DEFAULT_PUBURL/static/k8scenarii/$zipfile
    URL=$DEFAULT_PUBURL/$zipfile
    [ $VERBOSE -ne 0 ] && echo "URL=$URL"

    WEB_ZIP=~/tmp/zipfile.$zipfile
    wget --no-cache -qO - $URL 2>/dev/null > $WEB_ZIP
    [ ! -s "$WEB_ZIP" ] && {
        ls -al $WEB_ZIP
        die "Empty file $WEB_ZIP"
    }

    WEB_CKSUM=$(cksum - < $WEB_ZIP)
    rm $WEB_ZIP

    [ $VERBOSE -ne 0 ] && echo "WEB_CKSUM=$WEB_CKSUM"
    [ "$WEB_CKSUM" != "$LOCAL_CKSUM" ] && {
      echo "WEB_CKSUM='$WEB_CKSUM' != LOCAL_CKSUM='$LOCAL_CKSUM'"
      continue
    }

    echo "Identical cksums for $zipfile ('$REPO_CKSUM','$LOCAL_CKSUM''$WEB_CKSUM')"
done

