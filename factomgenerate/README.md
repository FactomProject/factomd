# factomgenerate

The templates and go go in this directory are used to create generated code.

See ../generated/ to see the finished product

To add a new type of generated object add a comment to any *.go  file in this project like this one:

```go
package queue

import "github.com/FactomProject/factomd/generated"

//FactomGenerate template accountedqueue typename Queue_IMsg type interfaces.IMsg import github.com/FactomProject/factomd/common/interfaces
type MsgQueue = generated.Queue_IMsg
```

NOTE: above we also add a type pointer
so we can still control the type name used throughout the codebase.

View this file: [../queue/queue.go](../queue/queue.go)
