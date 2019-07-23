# watcher

The watcher module provides functions for 'watching' a file for modifications and processing any new bytes in Go.

## Installation

To install the module, simply run:

`$ go get github.com/walkert/watcher`


## Usage

The `watcher` module can be used in two modes: standard and channel.

### Standard

Simply call `New` with the name of the file to be watched. You can then call `WasModified` which simply returns a boolean based on whether or not the file was modified or `GetNewBytes` which will return a byte slice containing any new bytes seen.

### Channel mode

If you supply the `WithChannelMonitor` option to `New`, the `watcher` will operate in channel mode. At every interval specified, the `watcher` will get any new bytes using `GetNewBytes` and send them to the `ByteChannel`. Any errors will be sent to the `ErrChannel`.

The example below shows a simple `watcher` usage in channel mode

```go
package main

import (
    "fmt"
    "log"

    "github.com/walkert/watcher"
)

func main() {
    w, err := watcher.New("/path/to/file", watcher.WithChannelMonitor(25))
    if err != nil {
        log.Fatal(err)
    }
    for {
         select {
         case bytes := <-w.ByteChannel:
             fmt.Printf("Received %d bytes from file\n", len(bytes))
         case err := <-w.ErrChannel:
             log.Fatalf("Received error from watcher: %v\n", err)
             return
         }
     }
}
```
