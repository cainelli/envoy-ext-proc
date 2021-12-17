# Envoy External Processing Filter

Experiment with [Envoy External Processing Filter](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_proc/v3/ext_proc.proto)

```shell
make up
```

```shell
curl http://127.0.0.1:10000/headers -v

*   Trying 127.0.0.1:10000...
* Connected to 127.0.0.1 (127.0.0.1) port 10000 (#0)
> GET /headers HTTP/1.1
> Host: 127.0.0.1:10000
> User-Agent: curl/7.77.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< server: envoy
< date: Wed, 15 Dec 2021 19:05:44 GMT
< content-type: application/json
< content-length: 187
< access-control-allow-origin: *
< access-control-allow-credentials: true
< x-envoy-upstream-service-time: 2
< x-request-id: ec55e255-f363-9706-95b6-16ba7d08df10
<
{
  "headers": {
    "Accept": "*/*",
    "Host": "127.0.0.1:10000",
    "User-Agent": "curl/7.77.0",
    "X-Custom-Header": "ok",
    "X-Envoy-Expected-Rq-Timeout-Ms": "15000"
  }
}
* Connection #0 to host 127.0.0.1 left intact
```
