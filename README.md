# TicTacToe

## Requirements

You'll probably need at least Go 1.11 since this uses Go modules.

## How to Setup

To instantiate a TicTacToe server at your local machine at port 1234 (default), you can either:

* Type the following in the command line:

    ```
    $ go run main.go
    ```

* Build and run the binary:
    
    ```
    $ go build
    $ ./tiki
    ```

## How to Test

Since the inputs and outputs are line-based only, you can test the server responses by using netcat or telnet, e.g:

```
$ telnet localhost 1234
Trying ::1...
Connected to localhost.
Escape character is '^]'.
help

Commands are:

JOIN <r>  join room <r> (and quit the current one if already joined)
MARK <n>  mark square <n>, where squares are numbered like in the following diagram:                                                                                   
           1 | 2 | 3
          ---+---+---
           4 | 5 | 6
          ---+---+---
           7 | 8 | 9
QUIT      close the current connection

^]
telnet>
Connection closed.
```

```
$ nc localhost 1234
help

Commands are:

JOIN <r>  join room <r> (and quit the current one if already joined)
MARK <n>  mark square <n>, where squares are numbered like in the following diagram:
           1 | 2 | 3
          ---+---+---
           4 | 5 | 6
          ---+---+---
           7 | 8 | 9
QUIT      close the current connection

quit
```

There's also a bundled-in dumb client under `exampleclient/`. To run, try in the command line:

```
$ cd exampleclient
$ go run client.go
```
