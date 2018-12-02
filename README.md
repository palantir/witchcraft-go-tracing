witchcraft-go-tracing
=====================
`witchcraft-go-tracing` defines interfaces for implementing `zipkin`-style tracing and provides an implementation that 
uses [openzipkin/zipkin-go](https://github.com/openzipkin/zipkin-go). The defined APIs mirror the `zipkin-go` APIs
closely, but are defined separately so that the underlying implementation can be changed and so that `witchcraft` 
projects can write tracing-related code using a common interface (while still allowing different projects/components to
use different underlying implementations if needed).

Tracer
------
Any program that wants to generate spans must create a tracer. A tracer is the mechanism that is used to create spans
(both root spans and child spans) and coordinates things such as whether or not a newly generated root span should be
sampled and how a completed span should be recorded. The tracer interface is `wtracing.Tracer`.

Reporter
--------
A reporter is an interface that receives information on a span that is marked as finished and reports or records it in
some manner -- for example, by writing it to a log or sending it to a remote system. In the `witchcraft` ecosystem, the
most commonly used reporter is a trace logger that writes a span as a trace log entry to a trace log file or to STDOUT.
The reporter interface is `wtracing.Reporter`.

Span
----
A span corresponds to a single section of an operation that is being traced. A span stores information such as the name
of the operation, the trace ID, the span ID, the parent ID (if the span has a parent), when the span started, etc. A 
span is created using a tracer and, when marked as finished, it is provided to a reporter that handles recording the 
span. The span interface is `wtracing.Span`. It consists of the core span data (stored in the `SpanContext` type and
accessible via the `Context()` function of the interface) and the `Finish()` function, which is called to signal that
the span is finished (which then sends the span information to the associated reporter).

Extractor/Injector
------------------
Tracing is typically used to track operations that span multiple different services/processes. In order for this to be
possible, there must be a mechanism to propagate spans across service boundaries. When a process makes a request to 
another process and it needs to be traced, it must inject its span information into the request. Similarly, when a 
process receives a request, it must extract any span information contained in that request. The `wtracing.SpanInjector`
and `wtracing.SpanExtractor` types define types that inject and extract spans, respectively.

The most common example of propagation is propagating spans in HTTP requests. The 
[B3 header propagation specification](https://github.com/openzipkin/b3-propagation) defines HTTP headers for 
representing spans. The `b3` package contains functions that return an injector and extractor that inject and extract
spans from an `*http.Request`.  

Usage
-----
### Tracer
Programs that intend to create and/or record spans must create a tracer. The tracer must be provided with a reporter,
which handles the completion of spans created by the tracer (it is possible to specify a no-op reporter). The tracer can
also be configured with a sampling policy (which is used to determine whether or not newly created root spans should be
sampled) and information on the local endpoint (service name, IP address, port).

The `wtracing` package defines the `Tracer` interface, but does not provide a concrete implementation of the interface.
The `wzipkin` package provides a `Tracer` implementation that is implemented using the `open-zipkin/zipkin-go` library.

The following creates a new tracer using the `wzipkin` tracer implementation and a no-op reporter:

```go
tracer, err := wzipkin.NewTracer(wtracing.NewNoopReporter())
```

Tracer creation functions typically support configuring the tracer using `wtracing.TracerOption` configuration 
functions. For example, the following creates a tracer that never samples:

```go
tracer, err := wzipkin.NewTracer(wtracing.NewNoopReporter(), wtracing.WithSampler(func(id uint64) bool { return false }))
```

In the most common use case, a program will instantiate a single tracer configured properly and then make it available
to the rest of the code in the program, either by passing it as an argument or by setting it on a context that is used
by program logic.

The `wtracing.ContextWithTracer` function can be used to create a context with the provided tracer set on it, and the
`wtracing.TracerFromContext` function can be used to retrieve a tracer that is set on a context (if one exists). Note
that, although setting a tracer on a context can be a useful pattern, it introduces an implicit API dependency on the 
state of the context, so this is something that should be kept in mind -- if this approach is taken, then care should be
taken to ensure that the value is always set on the contexts provided to program logic and logic that extracts the 
tracer from the context should be cognizant of possible failure modes. 

### Spans
Spans are created using a tracer and require a span name to be created. Spans can also be created with various options
that configure information on the span (such as the span's kind, information on the address of a remote endpoint if a
span is capturing a network call, etc.). Conceptually, if a span is created as part of an operation that is already part
of a span, then the newly created span should set that span as its parent span.

Programs that utilize spans and use contexts typically set the span as part of the context. Thus, when creating a new
span, the following pattern is typically used:

```go
// assume "var tracer wtracing.Tracer" exists and is non-nil 
span, ctx := wtracing.StartSpanFromContext(tracer, ctx, spanName).
defer span.Finish()
```

This call creates a new span and returns the newly created span and a copy of the provided context with the newly 
created span set as its span. If the provided context already has a span set on it, the the newly created span is 
configured to be a child span of that span.

If the tracer is set on the context, it can be accessed using `TracerFromContext(ctx)`, and the call can be combined
with the above as:

```go
span, ctx := wtracing.StartSpanFromContext(wtracing.TracerFromContext(ctx), ctx, spanName)
// span will be nil if tracer is nil (not set on context). If that is known to never be the case, this check can be
// omitted (but line will panic if there is ever a situation where the context does not have a tracer set) 
if span != nil {
    defer span.Finish()
}
``` 

### Injecting/extracting spans to deal with multi-process spans
Communication that occurs at the process boundary (for example, incoming and outgoing HTTP requests to/from other 
services) should inject and extract spans from the process communication mechanism as needed.

For example, if an HTTP request is being made to another service, the span information should be injected in the request 
so that, if the other service supports tracing, it will create child traces that are properly associated with the trace 
ID and span ID of the request. Typically a new span would be created for the request. The following is an example of
this workflow:

```go
// assume "var tracer wtracing.Tracer" exists and is non-nil 
span, ctx := wtracing.StartSpanFromContext(tracer, ctx, spanName).
defer span.Finish()

// req is the outgoing *http.Request 
b3.SpanInjector(req)(span.Context())
```

[conjure-go-runtime](https://github.com/palantir/conjure-go-runtime) clients automatically handle this logic.

As another example, if an HTTP request is received from another service and work is done based on that, any span 
information that is set on the incoming request should be used as the current span so that any new spans created by the
current process will properly set the parent span as the incoming one and use its sampling decision. The following is an
example of this workflow:

```go
// assume "var tracer wtracing.Tracer" exists and is non-nil
// extract the SpanContext from the request and start a new span with that span as the parent 
span := tracer.StartSpan(req.Method, wtracing.WithParentSpanContext(b3.SpanExtractor(req)()))
defer span.Finish()
// update the context to be a context that has the span set on it
ctx = wtracing.ContextWithSpan(ctx, span)
```

[witchcraft-go-server](https://github.com/palantir/witchcraft-go-server) servers automatically handle this logic in its
request middleware.

License
-------
This project is made available under the [Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0).
