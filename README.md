![alt text][logo]

# inoutbi
 
A three-way port connector that can be combined with ```socat``` to simplify setting up direction-dependent message modification

## Background

I can't for the life of me figure out how to modify the content of network traffic flowing in bidirectional streams using ```socat```, which seems to be limited to [modifying only the terminators](https://stackoverflow.com/questions/2166399/rewriting-a-tcp-stream-on-the-fly-how-difficult-is-it-how-about-taking-a-dump). Perhaps there is a trick with [address chains](http://www.dest-unreach.org/socat/doc/socat-addresschain.html) or [multiplexing](https://stackoverflow.com/questions/17480967/using-socat-to-multiplex-incoming-tcp-connection), but [not being able to duplicate messages to all clients](https://unix.stackexchange.com/questions/195880/socat-duplicate-stdin-to-each-connected-client) means messages get lost. No doubt someone will be along soon to raise an issue to show how to do it - please do - my failed attempts are at the end of the doc. Other tools considered but disregarded include ````netcat``` and ```scapy``` together, which wasn't appropriate for me because it focuses on [low-level ```tcp``` modification](https://medium.com/@sarunyouwhangbunyapirat/packet-manipulation-with-netcat-and-scapy-9403ebaa82de) and complicates deployment because it runs on python. So it seemed more efficient just to write a helper that I could use along with ```socat`` that did what I needed. Actually, it's taken longer to write this README than the actual code and (minimal) tests.

## Usage

## Basic connection test

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

I have a small microcontroller with no realtime clock that is producing JSON messages. To aid in reconstructing the data at the other end of the connection, I want to add the current time to the outgoing JSON objects. However, I don't want to handle timekeeping on a microncontroller without a real time clock unit because of the hassle of drift, synchronisation, daylight saving awareness etc. All the messages are passing through a linux single board computer, so mounting a MITM attack on my own data stream seems like the most efficient solution. Also, I'd like to have the option of sanity-checking incoming messages for errors or being oversize. These are two different filters.

The time can be added by using [jq](https://stedolan.github.io/jq/), which is "like ```sed``` for JSON":

```
$ echo -n '{"foo":"bar"}' | jq -c '.time=now'
{"foo":"bar","time":1573242308.612527}
```

Invalid JSON can be filtered out using [try/catch](https://stackoverflow.com/questions/41035458/suppress-jq-parse-error-messages-in-linux)
```
echo -n 'foo' | jq -n 'try inputs catch {}'
{}
```

These filters can be worked into the data flows using pipes (or possibly address chains). 

```$ inoutbi -in 5000 -out 5001 -bi 5002```

```$ inoutbi -in 6000 -out 6001 -bi 6002```

```$ socat -u tcp:localhost:5001 - |  jq -c '.time=now' | tcp:localhost:6000```

```$ socat -u tcp:localhost:6001 tcp:localhost:5000```

```$ socat - tcp:localhost:6002```

```$ socat - tcp:localhost:5002```

![alt text][use2]


## Limitations

### Singleton clients

One client connected to the outward port will get all messages send into the bidirectional port, which is all I need. If you want to distribute those messages to multiple clients, you will need to arrange your own ```tee```. I can use [vw](https://github.com/timdrysdale/vw)'s ```ws``` interface for that because clients connecting to the same path get all messages sent to that path. Note that ```vw``` cannot be used for bidirectional message modification without ```inoutbi``` because a cycle is formed that means messages would echo forever.

### Ports are TCP only

I stuck to TCP ports only because they can be proxied (to probably anything you need) with ```socat```. 


## Attempts at using socat for this (spoiler: failed)

For those who are thinking surely it can't be that hard to do it with ```socat``` here's what I found:

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

### And something that is really close ...

There is some very encouraging info [here](http://www.dest-unreach.org/socat/doc/socat-exec.html) but the examples throw errors. It is dated from 2009.

```

Executing programs using socat
Introduction
From its very beginning socat provided the EXEC and SYSTEM address types for executing programs or scripts and exchanging data with them. Beginning with version 2 - with implementation of the address chain feature (inter addresses) - these address types were enhanced to allow execution of programs also in inter address context.

Program context types
Currently socat provides three contexts (interfaces) for program or script execution:

The endpoint context: this is the classical situation where socat equips the program with stdin and stdout. It does not matter if the program uses other external communication channels. Address keywords: EXEC, SYSTEM. This variant should be easy to understand in terms of socat version 1 functionality and is therefore not further discussed here.
The bidirectional inter address context: socat expects the program to use two bidirectional channels: stdin and stdout on its "left" side and file descriptors 3 and 4 on its "right" side. This allows to provide nearly arbitrary data manipulations within the context of socat chains. Address keywords: EXEC, SYSTEM.
The unidirectional inter address context: for easy integration of standard UNIX programs socat provides the EXEC1 and SYSTEM1 address types where socat provides stdin on the "left" side and stdout on the "right" side of the program, or vice versa for right-to-left transfers.
Note: The endpoint and the unidirectional inter address contexts both just use the program's stdio to communicate with it. However, in practice the last form will in most cases just manipulate and transfer data, while the first form will usually have side effects by communicating with exteral resources or by writing to the file system etc.

Executing programs in bidirectional inter address context
socat address chains concatenate internal modules that communicate bidirectionally. For example, a chain that establishes encrypted connection to a socks server might look something like this (parameters and options dropped):

"SOCKS:... | OPENSSL-CLIENT | TCP:..."

If you have a program that implements a new encryption protocol the chain could be changed to:

"SOCKS:... | EXEC:myencrypt.sh | TCP:..."

The complete example:

socat - "SOCKS:www.domain.com:80 | EXEC:myencrypt.sh | TCP:encrypt.secure.org:444"

The myencrypt.sh script would be a wrapper around some myencrypt program. It must adhere a few rules: It reads and writes cleartext data on its left side (FDs 0 and 1), and it reads and writes encrypted data on its right side (FDs 3 and 4). Thus, cleartext data would come from the left on FD 0, be encrypted, and sent to the right side through FD 4. Encrypted data would come from the the right on FD 3, be unencrypted, and sent to the left side through FD 1. It does not matter if the encryption protocol would required negotiations or multiple packets on the right side.

The myencrypt.sh script might log to syslog, its own log file, or to stderr - this is independend of socat. It might have almost arbitrary side effects.

For optimal integration the script should be able to perform half-close and should be able work with different file descriptor types (sockets, pipes, ptys).

The socat source distribution contains two example scripts that focus on partial aspects:

predialog.sh implements an initial dialog on the "right" script side and, after successful completion, begins to transfer data unmodified in both directions.
cat2.sh shows how to use two programs unidirectionally to gain bidirectional transfer. The important aspects here are the shell based input / output redirections that are necessary for half-close.
Using unidirectional inter addresses
There exist lots of UNIX programs that perform data manipulation like compression or encoding from stdin to stdout, while related programs perform the reverse operation (decompression, decoding...) also from stdin to stdout. socat makes it easy to use those programs directly, i.e. without the need to write a bidirectional wrapper shell script.

socat - "exec1:gzip % exec1:gunzip | tcp:remotehost:port"

The % character creates a dual communication context where different inter addresses are used for left-to-right and right-to-left transer (see socat-addresschain.html. socat generates stdin/stdout file descriptors for both programs independently.
```
Except, exec1 seems to have been removed ... what's the new way of expressing that?

[logo]: ./img/logo.png "inoutbi logo"

[use1]: ./img/use1.png "connection schematic of two inoutbi back-to-back"

[use2]: ./img/use2.png "connection schematic of two inoutbi back-to-back with interstitial filters"
