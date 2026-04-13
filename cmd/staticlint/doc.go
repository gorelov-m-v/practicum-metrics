// staticlint is a custom multichecker that combines multiple static analysis tools
// for Go code quality and correctness checking.
//
// # Usage
//
// Build and run the multichecker:
//
//	go build -o staticlint ./cmd/staticlint
//	./staticlint ./...
//
// Or run directly:
//
//	go run ./cmd/staticlint ./...
//
// # Included Analyzers
//
// The multichecker includes analyzers from four categories:
//
// # Standard analyzers (golang.org/x/tools/go/analysis/passes)
//
// These are the official Go analysis passes covering common correctness issues:
//
//   - appends: detects missing values in append calls
//   - asmdecl: checks assembly declarations against Go function signatures
//   - assign: detects useless assignments
//   - atomic: checks for common mistakes using sync/atomic
//   - bools: detects common mistakes involving boolean operators
//   - buildtag: checks //go:build and // +build directives
//   - cgocall: detects violations of cgo pointer passing rules
//   - composite: checks for unkeyed composite literals
//   - copylock: checks for locks erroneously passed by value
//   - defers: checks for common mistakes in defer statements
//   - directive: checks Go toolchain directives
//   - errorsas: checks that errors.As is called with the correct type
//   - framepointer: reports assembly that clobbers the frame pointer
//   - httpresponse: checks for mistakes using HTTP responses
//   - ifaceassert: detects impossible interface-to-interface type assertions
//   - loopclosure: checks for references to enclosing loop variables
//   - lostcancel: checks for failure to cancel a context
//   - nilfunc: checks for useless comparisons between functions and nil
//   - printf: checks consistency of Printf format strings and arguments
//   - shift: checks for shifts that equal or exceed the width of the integer
//   - sigchanyzer: checks for misuse of os/signal.Notify on unbuffered channels
//   - slog: checks for invalid structured logging calls
//   - sortslice: checks for misuse of sort.Slice
//   - stdmethods: checks for misspellings in signatures of well-known interfaces
//   - stringintconv: flags type conversions from integers to strings
//   - structtag: checks that struct field tags conform to reflect.StructTag.Get
//   - testinggoroutine: checks for calls to Fatal from a test goroutine
//   - tests: checks for common mistaken usages of tests and examples
//   - timeformat: checks for time.Format/time.Parse with bad patterns
//   - unmarshal: checks for passing non-pointer types to unmarshal
//   - unreachable: checks for unreachable code
//   - unsafeptr: checks for invalid conversions of uintptr to unsafe.Pointer
//   - unusedresult: checks for unused results of calls to some functions
//   - unusedwrite: checks for unused writes to struct fields or array elements
//
// # Staticcheck SA analyzers (honnef.co/go/tools/staticcheck)
//
// All SA-class analyzers that detect bugs and suspicious code constructs,
// including but not limited to:
//
//   - SA1*: Various checks for incorrect API usage
//   - SA2*: Checks for incorrect usage of concurrency primitives
//   - SA3*: Checks for correctness issues in tests
//   - SA4*: Checks for code that is probably not what the author intended
//   - SA5*: Checks for correctness issues
//   - SA6*: Performance-related checks
//   - SA9*: Checks for dubious code constructs
//
// # Staticcheck ST analyzers (honnef.co/go/tools/stylecheck)
//
// ST-class analyzers enforce Go style conventions:
//
//   - ST1*: Style checks including naming conventions, error strings, etc.
//
// # Public third-party analyzers
//
//   - ineffassign (github.com/gordonklaus/ineffassign): detects ineffectual assignments in Go code
//   - bodyclose (github.com/timakin/bodyclose): checks whether HTTP response body is closed
//
// # Custom analyzers
//
//   - exitcheck: reports direct calls to os.Exit in the main function of package main.
//     Using os.Exit bypasses deferred functions and prevents proper resource cleanup.
//     The recommended pattern is to use a run() function that returns an exit code.
package main
