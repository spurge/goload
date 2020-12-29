#!/bin/sh

while true; do
  case "$1" in
    -host)
      HOST=$2
      shift 2
      ;;
    -port)
      PORT=$2
      shift 2
      ;;
    -loglevel)
      LOG_LEVEL=$2
      shift 2
      ;;
    -logformat)
      LOG_FORMAT=$2
      shift 2
      ;;
    -concurrency)
      CONCURRENCY=$2
      shift 2
      ;;
    -sleep)
      SLEEP=$2
      shift 2
      ;;
    -repeat)
      REPEAT=$2
      shift 2
      ;;
    -targets)
      TARGETS=$2
      shift 2
      ;;
    *)
      break
      ;;
  esac
done

exec goload \
  -host $HOST \
  -port $PORT \
  -loglevel $LOG_LEVEL \
  -logformat $LOG_FORMAT \
  -concurrency $CONCURRENCY \
  -sleep $SLEEP \
  -repeat $REPEAT \
  -targets $TARGETS
