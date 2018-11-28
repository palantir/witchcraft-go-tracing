witchcraft-go-tracing
=====================
`witchcraft-go-tracing` defines interfaces for implementing `zipkin`-style tracing and provides an implementation that 
uses [openzipkin/zipkin-go](https://github.com/openzipkin/zipkin-go). The defined APIs mirror the `zipkin-go` APIs
closely, but are defined separately so that the underlying implementation can be changed and so that `witchcraft` 
projects can write tracing-related code using a common interface (while still allowing different projects/components to
use different underlying implementations if needed).

License
-------
This project is made available under the [Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0).
