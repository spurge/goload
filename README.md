goload
======

Goload is a blackbox-exporter, which is a monitoring tool exposing metrics with status and latency from its targets. It's written to do a series of requests against a set of targets with the ability to pass data between requests. It'll then export it's metrics for Prometheus to scrape.

The easiest way to run it, it by using docker. You'll then configure it by a set of environment variables and a file with your targets.

Environment varables
--------------------

* `HOST` the host to listen on, default is `localhost`
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
