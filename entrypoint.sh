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
    -stderrthreshold)
      LOG_LEVEL=$2
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
  -stderrthreshold $LOG_LEVEL \
  -concurrency $CONCURRENCY \
  -sleep $SLEEP \
  -targets $TARGETS
