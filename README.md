goload
======

Goload is a blackbox-exporter, which is a monitoring tool exposing metrics with status and latency from its targets. It's written to do a series of requests against a set of targets with the ability to pass data between requests. It'll then export it's metrics for Prometheus to scrape.

The easiest way to run it, is by using docker. You'll then configure it by a set of environment variables and a file with your targets.

How to run it
-------------

Goload comes shipped with Alpine as a docker image at `spurge/goload`

### With docker

```sh
docker run -d -v $PWD/targets.yml:/targets.yml -e TARGETS=/targets.yml -p 8100:8100 spurge/goload
curl localhost:8100/metrics
```

### Rolling your own docker image

```Dockerfile
FROM spurge/goload

COPY targets.yml /app

ENV TARGETS=/app/targets
```

and then ...

```sh
docker build -t goload-some-target .
docker run -d goload-some-target
curl localhost:8100/metrics
```

Environment varables
--------------------

* `HOST` the host to listen on, default is `0.0.0.0`
* `PORT` the port to listen on, default is `8100`
* `CONCURRENCY` the number of concurrent workers, doing requests against your targets, default is `0`
* `SLEEP` the time to sleep in seconds before running through your targets again, default is `1`
* `TARGETS` the path to your targets defined in an yaml-file

Targets yaml-file
-----------------

```yaml
- name: login
  url: http://some-host/login
  method: POST
  body: >
    {
      "username": "yumba",
      "password": "secret"
    }
- name: profile
  url: 'http://some-host/profile/{{ fromJson "login" "user.id" }}'
  method: GET
  headers:
    Authentication: 'Bearer {{ fromJson "login" "user.auth.token" }}'
```
