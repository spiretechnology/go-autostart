# go-autostart

Go library to register your app to autostart on startup (supports Linux, macOS, and Windows).

## Basic example

```go
// Define your app's autostart behavior
app := autostart.New(autostart.Options{
    Label: "com.myapp.MyApp",
    Vendor: "Company"
    Name: "My App",
    Description: "My app description",
    Mode: autostart.ModeUser,
    Arguments: []string{ /* ... */ },
})

// Enable, disable, or check the status
app.Enable()
app.Disable()
app.IsEnabled()

// To get other useful data
app.DataDir()
app.StdOutPath()
app.StdErrPath()
```

## Supported platforms

- Linux (systemd)
- macOS (launchd)
- Windows (Service Manager)

## Logging

When running your process as a service, it's a good idea to make a deliberate decision about where to send logs.

With `go-autostart`, you can specify file paths for both stdout and stderr. If you don't specify a path, a platform-specific default location will be chosen based on your app details.

```go
app := autostart.New(autostart.Options{
    // ...
    StdoutPath: "/path/to/myapp.log",
    StderrPath: "/path/to/myapp.err",
})
```

You can create writers to your custom log files. This will override the global `os.Stdout` and `os.Stderr` with custom writers
that send logs to both the program's stdio and the custom log files.

Run the following to override the global `os.Stdout` and `os.Stderr` with your custom log files. This function returns a new, wrapped version of stdout. If you still want to write logs to the console, in addition to the log file, writing to this new stdout will do both.

```go
stdout, err := app.Stdio()
```

Then, to write logs to both the console and the log file:

```go
// Write directly
fmt.Fprintln(stdout, "Hello, world!")

// Set the output for the `log` package
log.SetOutput(stdout)
log.Println("Hello, world!")
```

## Full Example

```go
// Define your app's autostart behavior
app := autostart.New(autostart.Options{
    Label: "com.myapp.MyApp",
    Vendor: "Company"
    Name: "My App",
    Description: "My app description",
    Mode: autostart.ModeUser,
    Arguments: []string{ /* ... */ },
})

// Setup logging to the log files
stdout, err := app.Stdio()
log.SetOutput(stdout)

// Enable auto-start for the app
app.Enable()
```
