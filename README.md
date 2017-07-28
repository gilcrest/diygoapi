# go-API-template
An attempt at creating a RESTful API template (built with Go)

### Critical components of any API (in no particular order)
- [ ] Unit Testing
- [ ] Verbose Code Documentation
- Instrumentation
    - request/response logging (ability to turn on and off logging type based on some type of flag)
        - [ ] logStyle 1: structured (JSON), leveled (debug, error, info, etc.) logging to stdout
        - [ ] logStyle 2: relational database logging (certain data points broken out, request/response in a TEXT datatype)
    - [ ] API Metrics
    - [ ] Performance Monitoring
- [ ] Helpful logging
- [ ] Proper Error Handling
- [ ] RESTful service versioning
- [ ] Properly Vendored dependencies (I have no idea how to do this yet...)
