# Module Hooks Example with Dependency Container
In input values of your hook, you have prebuilded clients, which you can use, if you need some additional info.

## Warning
Dependency container clients should only be used as a last resort, because requests through them load the system more than working with our hook manager in Deckhouse. 

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
go build -o dependency-example-module-hooks ./cmd/hooks 
```