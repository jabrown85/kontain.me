# kontain.me

[kontain.me](https://kontain.me) serves Docker containter images generated
on-demand at the time they are requested.

`docker pull kontain.me/random:latest` serves an image containing random data.
By default the image contains one layer containing 10 MB of random bytes.  You
can request a specific size and shape of random image. For example,
`kontain.me/random:4x100` generates a random image of 4 layers of 100 random
bytes each.

`docker pull kontain.me/ko/[import path]` serves an image containing a Go
binary fetched using `go get` and built into a container image using
[`ko`](https://github.com/google/go-containerregistry/blob/master/cmd/ko/README.md).
For example, `docker pull
kontain.me/ko/github.com/google/go-containerregistry/cmd/ko` will fetch, build
and (eventually) serve a Docker image containing `ko` itself. _Koception!_

The registry does not accept pushes and does not handle requests for images by
digest. This is a silly hack and probably isn't stable. Don't rely on it for
anything serious.

## How it works

The backend is implemented using [Google Serverless
Containers](https://g.co/serverlesscontainers), with an App Engine Go 1.11
frontend to provide SSL on a custom domain (source in [`fwd/`](./fwd/)).

When the app receives a request for an image manifest, it parses the request
and generates layers for the requested image, writing the blobs to [Google
Cloud Storage](https://cloud.google.com/storage/). After it receives the
manifest, `docker pull` fetches the blobs. The app simply redirects to Cloud
Storage to serve the blobs. Blobs are deleted after 10 days.