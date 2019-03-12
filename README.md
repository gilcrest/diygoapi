# go-API-template

A RESTful API template (built with Go) - work in progress...

- The goal of this app is to make an example/template of relational database-backed APIs that have characteristics needed to ensure success in a high volume environment.

- This is a work in progress - you'll notice most things below are not checked.  Any feedback and/or support are welcome. I have thick skin, so please feel free to tell me how bad something is and I'll make it better.

## Critical components of any API (in no particular order)

- [ ] Unit Testing (with reasonably high coverage %)
- [x] Verbose Code Documentation
- [ ] Instrumentation
  - [x] [configurable http request/response logging](#configurable-logging)
  - [ ] Helpful debug logging
  - [ ] API Metrics
  - [ ] Performance Monitoring
- [x] [Fault tolerant - Proper Error Raising/Handling](#http-json-error-responses)
- [ ] RESTful service versioning
- [ ] Security/Authentication/Authorization using HTTPS/OAuth2, etc.
- [ ] Containerized
- [ ] Generated Client examples
- [ ] Extensive API Documentation for Clients of the API (see [twilio](https://www.twilio.com/docs/api/rest), [Uber](https://developer.uber.com/docs/riders/ride-requests/tutorials/api/introduction), [Stripe](https://stripe.com/docs/api/go#intro) and [mailchimp](http://developer.mailchimp.com/documentation/mailchimp/) as good examples - potentially use [Docusaurus](http://docusaurus.io/)

----

### Configurable HTTP Request Logging

Go-API-Template uses [httplog](https://github.com/gilcrest/httplog) to allow for "configurable request/response logging". With **httplog** you can enable request and response logging, get a unique ID for each request as well as set certain request elements into the context for retrieval later in the response body.

> Note the audit section of the response body below, this is provided by [httplog](https://github.com/gilcrest/httplog). Check out the repo for more details in the README.

```json
{
    "username": "gopher",
    "mobile_id": "617-445-2778",
    "email": "repoman@alwaysintense.com",
    "first_name": "Otto",
    "last_name": "Maddox",
    "update_user_id": "gilcrest",
    "created": 1539138260,
    "audit": {
        "id": "beum5l708qml02e3hvag",
        "url": {
            "host": "127.0.0.1",
            "port": "8080",
            "path": "/api/v1/adapter/user",
            "query": "qskey1=fake&qskey2=test"
        }
    }
}
```

----

### HTTP JSON Error Responses

For error responses, the api sends a simple structured JSON message in the response body, similar to [Stripe](https://stripe.com/docs/api#errors), [Uber](https://developer.uber.com/docs/riders/guides/errors) and many others, e.g.:

```json
{
    "error": {
        "kind": "input_validation_error",
        "param": "Year",
        "message": "The first film was in 1878, Year must be >= 1878"
    }
}
```

This is achieved using a carve-out of the [Upspin project](https://github.com/upspin/upspin) and specifically the error handling functionality summarized so eloquently by Rob Pike in [this article](https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html). My copy of the errors functionality has changes that suit my use cases. Namely, I have built a new function `errors.RE` (for **R**esponse **E**rror), which allows the engineer to easily build an HTTP response error.

This is achieved by sending a custom `HTTPErr` error type as a parameter to the `httpError` function, which then writes then replies to the request using the `http.Error` function. The structure of `HTTPErr` is based on Matt Silverlock's blog post [here](https://elithrar.github.io/article/http-handler-error-handling-revisited/), but I ended up not using the wrapper/handler technique instead opting for a different approach. You'll also note the use of constants from a custom lib/errors package. This is pretty much lifted straight from the [upspin.io](https://upspin.io/) project, with minor tweaks to suit `go-API-template` purposes.

The package makes error handling fun - you’ll always return a pretty good looking error and setting up errors is pretty easy.

When creating errors within your app, you should not have to setup every error as an `HTTPErr`— you can return "normal" errors lower down in the code and, depending on how you organize your code, you can catch the error and form the `HTTPErr` at the highest level so you’re not having to deal with populating a cumbersome struct all throughout your code. The below example is taken from the handler, which in this case is the highest level. In this case, any errors returned from the SetUsername method can be qualified as validation errors and the error caught from this method would be a "normal" Go error.

```go
    err = usr.SetUsername(rqst.Username)
    if err != nil {
        err = HTTPErr{
            Code: http.StatusBadRequest,
            Kind: errors.Validation,
            Err:  err,
        }
        httpError(w, err)
        return
    }
```

----

### Helpful Resources I've used in this app (outside of the standard, yet amazing blog.golang.org and golang.org/doc/, etc.)

websites/youtube

- [JustforFunc](https://www.youtube.com/channel/UC_BzFbxG2za3bp5NRRRXJSw)

- [Go By Example](https://gobyexample.com/)

Books

- [Go in Action](https://www.amazon.com/Go-Action-William-Kennedy/dp/1617291781)
- [The Go Programming Language](https://www.amazon.com/Programming-Language-Addison-Wesley-Professional-Computing/dp/0134190440/ref=pd_lpo_sbs_14_t_0?_encoding=UTF8&psc=1&refRID=P9Z5CJMV36NXRZNXKG1F)

Blog/Medium Posts

- [How I write Go HTTP services after seven years](https://medium.com/statuscode/how-i-write-go-http-services-after-seven-years-37c208122831)
- [The http Handler Wrapper Technique in #golang, updated -- by Mat Ryer](https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702)
- [Writing middleware in #golang and how Go makes it so much fun. -- by Mat Ryer](https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81)
- [http.Handler and Error Handling in Go -- by Matt Silverlock](https://elithrar.github.io/article/http-handler-error-handling-revisited/)
- [How to correctly use context.Context in Go 1.7 -- by Jack Lindamood](https://medium.com/@cep21/how-to-correctly-use-context-context-in-go-1-7-8f2c0fafdf39)
- [Standard Package Layout](https://medium.com/@benbjohnson/standard-package-layout-7cdbc8391fc1)
- [Practical Persistence in Go: Organising Database Access](http://www.alexedwards.net/blog/organising-database-access)
- [Writing a Go client for your RESTful API](https://medium.com/@marcus.olsson/writing-a-go-client-for-your-restful-api-c193a2f4998c)
