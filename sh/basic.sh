#!/bin/bash
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

#echo $LIN
#echo $LOU
#echo $TIN
#echo $TOU
#echo $RIN
#echo $ROU
#echo $BIN
#echo $BOU
#echo $LBI
#echo $TBI
#echo $RBI
#echo $BBI

#let the listeners start up
sleep 1s

#run the socats in parallel
(socat tcp:localhost:$LOU tcp:localhost:$TIN)&
(socat tcp:localhost:$TOU tcp:localhost:$RIN)&
(socat tcp:localhost:$ROU tcp:localhost:$BIN)&
(socat tcp:localhost:$BOU tcp:localhost:$LIN)&

# run the commands in parallel
(socat tcp:localhost:$TBI exec:$3,pty,ctty)&
(socat tcp:localhost:$BBI exec:$4,pty,ctty)&

# idle waiting for abort from user
read -r -d '' _ </dev/tty
trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT

# kill the processes 
# do we need all of these kills?
#fuser -k $LIN/tcp
#fuser -k $LOU/tcp
#fuser -k $TIN/tcp
#fuser -k $TOU/tcp
#fuser -k $RIN/tcp
#fuser -k $ROU/tcp
#fuser -k $BIN/tcp
#fuser -k $BOU/tcp
#fuser -k $LBI/tcp
#fuser -k $TBI/tcp
#fuser -k $RBI/tcp
#fuser -k $BBI/tcp
# how do we kill command socats?
# how do we kill client-only socats?
