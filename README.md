# short.ly

shortly is a simple POC service that exposes an HTTP API for generating and storing short URLs. In addition to generating
them, it can also accept an existing short URL and redirect the consumer to the original long URL.

```http
Endpoint for redirecting a short URL --> original URL
GET /{ID}

Endpoint for generating a short URL
POST /api/v1/shorten
```

## Usage

To interact with the service, take the following steps in your preferred terminal:

```bash
cd ~/your_workdir
export GO111MODULE=on
git clone https://github.com/williamhgough/shortly.git
cd shortly && go mod vendor
cd cmd/shortly && go build && ./shortly
```

Then, to use the service, execute the following commands in a new terminal window:

```bash
curl -i -X POST http://localhost:8080/api/v1/shorten -d '{"original_url":"http://google.com"}'

HTTP/1.1 200 OK
Content-Type: application/json
Date: Sun, 24 Feb 2019 15:26:16 GMT
Content-Length: 95

{"id":"O6YMaXo","original_url":"http://google.com","short_url":"http://localhost:8080/O6YMaXo"}
```

Then, to use the short URL, simply paste the `short_url` provided by the API in to the browser of your choice.. et violà!

## Testing

To run the project tests with coverage, execute the following:

```bash
go test ./... -cover
?       github.com/williamhgough/shortly/cmd/shortly    [no test files]
?       github.com/williamhgough/shortly/pkg/adding     [no test files]
ok      github.com/williamhgough/shortly/pkg/hashing    0.003s  coverage: 80.0% of statements
ok      github.com/williamhgough/shortly/pkg/http/rest  0.007s  coverage: 75.0% of statements
?       github.com/williamhgough/shortly/pkg/redirect   [no test files]
ok      github.com/williamhgough/shortly/pkg/storage/memory     0.009s  coverage: 100.0% of statements
```

## Design Choices

My primary goal with this service was to create a simple implementation of a URL shortener. To do this I chose to only use the standard library _where appropriate_. The service has only one dependency on a third party library, [go-hashids](https://github.com/speps/go-hashids). Initially I looked at implementing that functionality myself, but that seemed counter-productive and unneccessary. However, to provide for any changes to the hashing implementation in the future, I created the `Hasher` interface. Meaning that alternate options can be added and swapped in easily.

Secondly I decided that I should consider storage, so I created a `Repository` interface that would also help future-proof the service against any changes in requirements down the line, thus improving the overall maintainability. My implementation of the interface `mapRepository` uses a `sync.RWMutex` to access the `map[string]*URLObject` concurrently.

Each `Service` instance of shortly contains a `Hasher` and `Repository` implementation. I chose to use one data structure for incoming requests to the API and for it's response to simplify it further.

You may also notice that the `Repository` interface includes an `Exists` method, this is used in the `generateURLHandler` to check if the URL passed into the request already exists in the database. This approach was taken on the assumption that the service is for internal use. If the service were to be public, you would not want to use this functionality as it would mean each short URL is not unique per consumer, and therefore if there were any metrics associated with it, they would be incorrect. To remove this functionality, you need only uncomment the following section on `shortly.go L129-133`, and the service is ready for per-consumer use again:

```go
if res, exists := s.db.Exists(req.OriginalURL); exists {
    log.Printf("ID %s already exists for URL: %s", res.ID, res.OriginalURL)
    respond(w, res)
    return
}
```

I chose to use a versioned API when deciding on exposed endpoints, having recently read [Designing Distributed Systems](https://www.amazon.co.uk/Designing-Distributed-Systems-Brendan-Burns/dp/1491983647), Brendan Burns stated:

> "It may not seem logical, but it costs very little to version your API when you initially define it. Refactoring versioning onto an API without it, on the other hand, is very expensive. Consequently it is a best practice to always add versions to your APIs even if you're not sure they'll ever change, better safe than sorry."

## Next Steps for production

In order to take this service to production, I would consider the following steps:

- Add CI integration (Travis, Circle, Gitlab etc)
- Benchmark and look at optimising the service with pprof
- Add better support for logging, e.g different verbosities (error, critical, debug, info etc...)
- Add service metric functionality, for reporting (ELK for logs and prometheus & grafana for metrics)
- Add metric information to the `URLObject` (e.g hit count)
- Add a UI for creating URLs and reviewing URL metrics
- Look at moving to a FaaS approach utilising [Open-Faas](https://www.openfaas.com/) or [Kubeless](https://kubeless.io/)

Due to it's on-demand nature, I think this service could be a great candidate for a FaaS approach, especially as an internal tool, as it could leverage the libraries above to manage scaling up by avoiding steep linear costs of pay-per-request model and instead benefitting from the pricing models of virtual machines on AWS' EKS or GCP's GKE.