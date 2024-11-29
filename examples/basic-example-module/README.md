# Module Hooks Basic Example
In this example you can build your basic hook binary and check how it works.

It can be usefull to understand how to register hooks in different layouts.

### Run

To get list of your registered hooks
```bash
go run ./cmd/hooks hook list
```

To get configs of your registered hooks
```bash
go run ./cmd/hooks hook config
```

To dump configs of your registered hooks in file
```bash
go run ./cmd/hooks hook dump
```

To run registered hook with index '0' (you can see index of your hook in output of "hook list" command)
```bash
go run ./cmd/hooks hook run 0
```

By default, all logs in hooks are suppressed and he waiting for files in default folders. 
To make them available, you must add env variable LOG_LEVEL and CREATE_FILES.
```bash
CREATE_FILES=true LOG_LEVEL=INFO go run ./cmd/hooks hook run 0
```

### Build
```bash
go build -o basic-example-module-hooks ./cmd/hooks 
```