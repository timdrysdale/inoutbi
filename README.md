![alt text][logo]

# inoutbi
 
A three-way port connector that can be combined with ```socat v1``` to simplify setting up direction-dependent message modification, while we wait for ```socat v2``` (which might very well work already, but was proving a bit recalcitrant in my hands).

## Background
Socat's [exec1](http://www.dest-unreach.org/socat/doc/socat-exec.html) feature, part of [address chaining](http://www.dest-unreach.org/socat/doc/socat-addresschain.html), is still in development and I have not managed to get clean pty behaviour out of it with per-line filters. That's probably down to me missing an important option, but I figured it would be good to have a verbose but robust solution in the meantime (plus I'd already written this before fully getting to the bottom of the issue with ```exec1``` in my use case). This helper tool allows per-line filters to be used in ```exec2 mode``` from the current stable version 1 ```socat```. 


## Usage

## Back-to-back demonstration

In separate terminals, run these commands (starting them in the order shown)

```$ inoutbi -in 5000 -out 5001 -bi 5002```

```$ inoutbi -in 6000 -out 6001 -bi 6002```

```$ socat -u tcp:localhost:5001 tcp:localhost:6000```

```$ socat -u tcp:localhost:6001 tcp:localhost:5000```

```$ socat - tcp:localhost:6002```

```$ socat - tcp:localhost:5002```

What is typed into the stdin for one ```inoutbi``` appears at the other, and vice versa.

![alt text][use1]

## With filtering

Put the scripts ```freeport```, ```jq-addtime``` and ```jq-clean``` into a location that's on your path.

Then, from ```./sh```, run these commands in this order, each in separate terminals

```./basic.sh 3000 4000 jq-addtime jq-clean```

```socat - tcp:localhost:3000```
 
```socat - tcp:localhost:4000```

JSON objects entered into port 3000 have the time added, and are reported at port 4000. In the opposite direction, non-JSON objects are replaced with {}.

![alt text][use2]

## Limitations

The basic ```inoutbi``` is limited in scope because it is a temporary fix while we await ```socat v2```.

### JQ

The ```jq-clean``` is not really very complete at the moment, and both scripts need modifying to suppress errors.

### Singleton clients

One client connected to the outward port will get all messages send into the bidirectional port, which is all I need. If you want to distribute those messages to multiple clients, you will need to arrange your own ```tee```. I can use [vw](https://github.com/timdrysdale/vw)'s ```ws``` interface for that because clients connecting to the same path get all messages sent to that path. Note that ```vw``` cannot be used for bidirectional message modification without ```inoutbi``` because a cycle is formed that means messages would echo forever.

### Ports are TCP only

I stuck to TCP ports only because they can be proxied (to probably anything you need) with ```socat```. 


## Attempts at using socat v1 for this without inoutbi
Before I tried the beta ```socat v2``` and its rather useful looking [address chains](http://www.dest-unreach.org/socat/doc/socat-addresschain.html), I tried and failed to do this with version one only. It didn't work, but just so I have a note of what I tried, I've included this section.

With ```socat``` version one (the current stable version), it's seemingly impossible to modify the content of network traffic flowing in bidirectional streams, other than to [modifying the terminators](https://stackoverflow.com/questions/2166399/rewriting-a-tcp-stream-on-the-fly-how-difficult-is-it-how-about-taking-a-dump). Some possible workarounds fail because of [not being able to duplicate messages to all clients](https://unix.stackexchange.com/questions/195880/socat-duplicate-stdin-to-each-connected-client) means messages get lost. Other tools considered but disregarded include ````netcat``` and ```scapy``` together, which wasn't appropriate for me because it focuses on [low-level ```tcp``` modification](https://medium.com/@sarunyouwhangbunyapirat/packet-manipulation-with-netcat-and-scapy-9403ebaa82de) and complicates deployment because it runs on python. 

### Simple port, with unidirectional connections

This doesn't work because the second client is refused

bidirectional port (where stdin is mock for the outside world)

```
socat tcp-listen:1236 -
```

outward port
```
socat -u tcp:localhost:1236 -
```

That works so far - messages typed into the bidirectional port appear at the outward port, and messages typed at the outward port do NOT appear at the bidirectional port. Good.

inward port
```
$ socat -u - tcp:localhost:1236
<connection refused message>
```

So the port needs to be multiplexed.

### Multiplexed socat port with unidirectional connections

Here's what I tried with ```socat``` [modified from this SO answer](https://stackoverflow.com/questions/17480967/using-socat-to-multiplex-incoming-tcp-connection)

bidirectional port
```
socat tcp-l:42000,fork,reuseaddr -
```

outward port
```
socat -u tcp:localhost:42000 -
```

inward port
```
socat -u - tcp:localhost:42000
```

A message typed into the inward port will appear at the bidirectional port.

Every approximately second message typed into the bidirectional port will appear at the outward port.

There does not seem to be a way to share a message with all clients, so this will result in message loss which is unacceptable.

Of course, that's why ```socat v2``` has the ```exec1``` feature.

[logo]: ./img/logo.png "inoutbi logo"

[use1]: ./img/use1.png "connection schematic of two inoutbi back-to-back"

[use2]: ./img/use2.png "connection schematic of two inoutbi back-to-back with interstitial filters"
