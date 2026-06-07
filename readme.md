*this file & project are a work in progress until v1 is released*

## What is Gopher?

Gopher is CLI-based web introspection tool that's slightly adjacent to web recon. Gopher takes a URL and crawls depth-first recursively to note links. Using found links, Gopher also looks for directory listings and resource fil servers, then attempts to crawl those and their parent links as well. 

## How it works

Gopher performs these steps:
1. Send GET request to URL
2. Parse & tokenize HTML, then search for href attributes
3. Note down information, then return to step 1 for each link found (if configuration allows)

Gopher keeps track of visited URLs to avoid infinite loops and redundant requests. It also handles the following edge cases:
- 302 redirects (follows them to fetch the final content)
- Mixed URL formats consisting of HTTP and HTTPs schemes, IP addresses, port numbers, and query parameters (normalizes URLs to avoid duplicates) 
- Subdomains (treats them as separate domains unless they share the same root domain, in which case they are considered internal links)

For external links, Gopher uses a simple heuristic to determine if a link is external: if the link's TLD is an ICANN registered public suffix and doesn't match parent domain, it's considered external. However, some TLDs such as `.md` and `.sh` are commonly used as file extensions, so Gopher treats those as internal resource links. For example, if Gopher finds a link to `something.md`, it will determine that `something.md` is an internal resource rather than an external link.

## Configuration

Gopher can be configured with the following options (you may see `internal/cli.go` for more details):
- [WIP] `Workers (-w [number], default=1)` The number of concurrent workers to use for crawling 
- `LogLevel (-l 0|1|2, default=1)` The level of logging to output (0=error, 1=info, 2=debug)
- `External (-e true/false, default=false)` Whether or not to crawl external links (links that do not share the same domain as the original URL)
- [WIP] `Output (-o 0|1, default=0)` Where to write the output (0=stdout, 1=sqlite)
- [WIP] `Proxies (-p [proxy], default=none)` A path to a `.txt` file containing a list of proxies to use for requests (one proxy per line, in the format `http://ip:port` or `https://ip:port`)

## Planned
- [x] Basic crawling functionality (fetching pages, parsing links, tracking visited URLs)
- [x] Logging with different levels (error, info, debug)
- [x] Handling edge cases (redirects, mixed URL formats, subdomains)
- [ ] Better README.md and deeper documentation on https://rasoolabbas.com
- [ ] Support concurrency via Goroutines
- [ ] Support for outputting results to a SQLite database
- [ ] Support for outputting results to an interactive graph-based web interface
- [ ] Support for using proxies to make requests
- [ ] Support for delaying requests to avoid rate-limiting
- [ ] Deeper crawling by detecting resource file servers and crawling them, while subtracting paths (e.g. if we find `example.com/assets/files2024/images/photo.jpg`, we can crawl `assets/files2024/`, then `/assets`, and so on)
