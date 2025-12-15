# Module Hooks Example
In this example you can build your hook binary with settings checkiung and check how it works.

It can be usefull to understand how to do helm values validation.

### Run

To get list of your registered hooks
```bash
go run . hook list
```

To get configs of your registered hooks
```bash
go run . hook config
```

To run settings check
```bash
go run . hook check
```

To dump configs of your registered hooks in file
```bash
go run . hook dump
```

To run registered hook with index '0' (you can see index of your hook in output of "hook list" command)
```bash
go run . hook run 0
```

By default, all logs in hooks are suppressed and he waiting for files in default folders. 
To make them available, you must add env variable LOG_LEVEL and CREATE_FILES.
```bash
CREATE_FILES=true LOG_LEVEL=INFO go run . hook run 0
```

### Build
```bash
go build -o example-module-hooks .
```