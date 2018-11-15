# camouflage

## Features

* Max size for proxied images
* Follow redirects to a certain depth
* Restricts proxied images content-types to a whitelist
* Forward images regardless of HTTP status code

## URL Formats

Camo supports two distinct URL formats:

    http://example.org/<digest>?url=<image-url>
    http://example.org/<digest>/<image-url>

The `<digest>` is a 40 character hex encoded HMAC digest generated with a shared
secret key and the unescaped `<image-url>` value. The `<image-url>` is the
absolute URL locating an image. In the first format, the `<image-url>` should be
URL escaped aggressively to ensure the original value isn't mangled in transit.
In the second format, each byte of the `<image-url>` should be hex encoded such
that the resulting value includes only characters `[0-9a-f]`.

## Configuration

Camo is configured through environment variables.

* `PORT`: The port number Camo should listen on. (default: 8081)
* `CAMO_HEADER_VIA`: The string for Camo to include in the `Via` and `User-Agent` headers it sends in requests to origin servers. (default: `Camo Asset Proxy`)
* `CAMO_KEY`: A shared key consisting of a random string, used to generate the HMAC digest.
* `CAMO_LENGTH_LIMIT`: The maximum `Content-Length` Camo will proxy. (default: 5242880)
* `CAMO_LOGGING_ENABLED`: The logging level used for reporting debug or error information. Options are `debug` and `disabled`. (default: `disabled`)
* `CAMO_MAX_REDIRECTS`: The maximum number of redirects Camo will follow while fetching an image. (default: 4)
* `CAMO_SOCKET_TIMEOUT`: The maximum number of seconds Camo will wait before giving up on fetching an image. (default: 10)
* `CAMO_TIMING_ALLOW_ORIGIN`: The string for Camo to include in the [`Timing-Allow-Origin` header](http://www.w3.org/TR/resource-timing/#cross-origin-resources) it sends in responses to clients. The header is omitted if this environment variable is not set. (default: not set)
* `CAMO_HOSTNAME`: The `Camo-Host` header value that Camo will send. (default: `unknown`)
* `CAMO_KEEP_ALIVE`: Whether or not to enable keep-alive session. (default: `false`)


## Examples
* Ruby - https://github.com/ankane/camo
* PHP - https://github.com/willwashburn/Phpamo
* Node.js - https://github.com/pine/node-camo-url
