#!/bin/sh

### BEGIN INIT INFO
# Provides:             box
# Required-Start:       $all
# Required-Stop:        $all
# Default-Start:        2 3 4 5
# Default-Stop:         0 1 6
# Short-Description:    box daemon
### END INIT INFO

NAME="boxd"
DESC="box upload processing daemon"

set -e

test -f /etc/default/box && . /etc/default/box
[ $BOX_BASE ] || exit 0

test -x $BOX_BASE/bin/boxd || exit 0

. /lib/lsb/init-functions

RETVAL=0


start() {
  echo -n "Starting $DESC: "
  start-stop-daemon --start \
                    --background \
                    --quiet \
                    --pidfile /var/run/box_boxd.pid \
                    --make-pidfile \
                    --chuid box \
                    --exec /bin/sh \
                    -- -c "$BOX_BASE/bin/boxd $BOX_OPTS >> /var/log/box/boxd.log 2>&1"
 
  RETVAL=$?
  echo "$NAME."
}

stop() {
  echo -n "Stopping $DESC: "
  
  start-stop-daemon --stop \
                    --quiet \
                    --oknodo \
                    --pidfile /var/run/box_boxd.pid
    
  RETVAL=$?
  echo "$NAME."
}


case "$1" in
  start)
    start
    ;;
  
  stop)
    stop
    ;;
  
  restart)
    stop
    start
    ;;
  
  *)
    echo "Usage: $NAME {start|stop|restart}"
    exit 1
    ;;
esac

exit $RETVAL
