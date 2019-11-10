#!/bin/bash
# set up bidirectional modification
# args: leftport rightport leftToRightCommand rightToLeftCommand
#
# assumes freeport shell is on the path
# freeport:
# #/bin/bash
# comm -23 <(seq 49152 65535) <(ss -tan | awk '{print $4}' | cut -d':' -f2 | grep "[0-9]\{1,5\}" | sort | uniq) | shuf | head -n 1
#
# the commands should use stdin/stdout and operate in a while loop like this:-
# jq-addtime:
# #!/bin/bash
# while true; 
# do 
# jq -c '.time=now'
# done
#
# jq-clean:
# #!/bin/bash
# while true; 
# do 
# jq -n 'try inputs catch {}'
# done 
#
# example: add timestamp to JSON objects using jq in one direction
# and sanitising in the other
# 
# ./basic.sh 3000 3001 jq-addtime jq-clean


LIN=$(freeport)
LOU=$(freeport)
TIN=$(freeport)
TOU=$(freeport)
RIN=$(freeport)
ROU=$(freeport)
BIN=$(freeport)
BOU=$(freeport)
LBI=$1
TBI=$(freeport)
RBI=$2
BBI=$(freeport)

#run in parallel
(inoutbi -in $LIN -out $LOU -bi $LBI)&
(inoutbi -in $TIN -out $TOU -bi $TBI)&
(inoutbi -in $RIN -out $ROU -bi $RBI)&
(inoutbi -in $BIN -out $BOU -bi $BBI)&

#let the listeners start up
sleep 1s

#run the socats in parallel
(socat tcp:localhost:$LOU tcp:localhost:$TIN)&
(socat tcp:localhost:$TOU tcp:localhost:$RIN)&
(socat tcp:localhost:$ROU tcp:localhost:$BIN)&
(socat tcp:localhost:$BOU tcp:localhost:$LIN)&

# run the commands in parallel
(socat tcp:localhost:$TBI exec:$3,pty,ctty,echo=0)&
(socat tcp:localhost:$BBI exec:$4,pty,ctty,echo=0)&

# idle waiting for abort from user
read -r -d '' _ </dev/tty
trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT

