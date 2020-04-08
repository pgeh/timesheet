#!/bin/bash

set -euo pipefail

LOGFILE=~/.timelogger

if [ $# -ne 1 ]
then
    echo "Command needed"
    exit 1
fi

case $1 in
    start)
        echo "`date -Iseconds` Start" >> $LOGFILE
    ;;
    stop)
        echo "`date -Iseconds` Stop" >> $LOGFILE
    ;;
esac