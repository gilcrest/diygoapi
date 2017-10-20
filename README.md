# go-API-template

An attempt at creating a RESTful API template (built with Go).

- I'm learning Go as a hobby and absolutely love it.  I'm hoping to start to use Go professionally soon and in order to do so, I want to make sure I'm able to build "industrial strength" REST APIs, in order to mentor my team. The goal of this app is to make an example/template of relational database-backed APIs that have characteristics needed to ensure success in a high volume environment. 

- This is a work in progress - you'll notice most things below are not checked, but I will get there!  Any feedback and/or support are welcome. I do not purport to be an expert and have very thick skin, so please feel free to tell me how bad something is and I'll make it better.

## Critical components of any API (in no particular order)

- [ ] Unit Testing (with reasonably high coverage %)
- [x] Verbose Code Documentation
- Instrumentation
  - request/response logging (ability to turn on and off logging type based on some type of flag)
    - [ ] logStyle 1: structured (JSON), leveled (debug, error, info, etc.) logging to stdout
    - [ ] logStyle 2: relational database logging (certain data points broken out into standard column datatypes, request/response stored in TEXT/CLOB datatype columns)
    - [ ] API Metrics
    - [ ] Performance Monitoring
    - [ ] Helpful debug logging
- [ ] Fault tolerant - Proper Error Raising/Handling
- [ ] RESTful service versioning
- [ ] Security/Authentication/Authorization using HTTPS/OAuth2, etc.
- [ ] Properly Vendored dependencies (I have no idea how to do this yet...)
- [ ] Containerized
- [ ] Generated Client examples
- [ ] Extensive API Documentation for Clients of the API (see [twilio](https://www.twilio.com/docs/api/rest), [Uber](https://developer.uber.com/docs/riders/ride-requests/tutorials/api/introduction), [Stripe](https://stripe.com/docs/api/go#intro) and [mailchimp](http://developer.mailchimp.com/documentation/mailchimp/) as good examples)

## Helpful Resources I've used in this app (outside of the standard, yet amazing blog.golang.org and golang.org/doc/, etc.)

websites/youtube
- [JustforFunc](https://www.youtube.com/channel/UC_BzFbxG2za3bp5NRRRXJSw)
- [Go By Example](https://gobyexample.com/)

Books
- [Go in Action](https://www.amazon.com/Go-Action-William-Kennedy/dp/1617291781)
- [The Go Programming Language](https://www.amazon.com/Programming-Language-Addison-Wesley-Professional-Computing/dp/0134190440/ref=pd_lpo_sbs_14_t_0?_encoding=UTF8&psc=1&refRID=P9Z5CJMV36NXRZNXKG1F)

Blog/Medium Posts
- [The http Handler Wrapper Technique in #golang, updated -- by Mat Ryer](https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702)
- [How to correctly use context.Context in Go 1.7 -- by Jack Lindamood](https://medium.com/@cep21/how-to-correctly-use-context-context-in-go-1-7-8f2c0fafdf39)
- [Writing middleware in #golang and how Go makes it so much fun.](https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81)
- [Standard Package Layout](https://medium.com/@benbjohnson/standard-package-layout-7cdbc8391fc1)
- [Practical Persistence in Go: Organising Database Access](http://www.alexedwards.net/blog/organising-database-access)
- [Writing a Go client for your RESTful API](https://medium.com/@marcus.olsson/writing-a-go-client-for-your-restful-api-c193a2f4998c)

----

> Release roadmap below is to help me stay on track.  With my ADD brain, I often lose focus...g

## Items to complete for Release 0.0.3

- Relational DB Request logging

## Items to complete for Release 0.0.4

- Add unique Request-ID to response headers using util

## Items to complete for Release 0.0.5

- Response JSON logging and httputil.DumpResponse
- Response Relational DB Logging