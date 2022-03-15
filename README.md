# udpme

udpme = UDP must have EDNS0。从协议层面借助 EDNS0 过滤掉有问题的 UDP 报文。全平台可用。

原理: udpme 会发送带 EDNS0 的请求报文，然后过滤掉没有 EDNS0 的应答报文。如果服务器支持 EDNS0，则会回应 EDNS0。因此过滤掉没有 EDNS0 报文可以过滤掉某些有问题的(假的)回应。

## 命令参数

- `l`: 监听地址。
- `u`: UDP 上游地址。上游必需支持 EDNS0。
- `t`: 快速测试 `-u` 是否支持 EDNS0。

## 使用

正常使用:

```bash
udpme -l 127.0.0.1:5353 -u 8.8.8.8
```

测试服务器是否支持 EDNS0:

打印 `edns0 response received`，说明支持。打印 `response received without edns0` 说明不支持。

```bash
udpme -u 8.8.8.8 -t
```

或者也可以用 `dig +edns` 测试。

```txt
> dig +edns cloudflare.com @8.8.8.8

; <<>> DiG 9.16.1-Ubuntu <<>> +edns cloudflare.com @8.8.8.8
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 2173
;; flags: qr rd ra ad; QUERY: 1, ANSWER: 2, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 512    <-- 支持的服务器应答中这里会包含 EDNS0 
;; QUESTION SECTION:
;cloudflare.com.			IN	A

;; ANSWER SECTION:
cloudflare.com.		300	IN	A	104.16.132.229
cloudflare.com.		300	IN	A	104.16.133.229

;; Query time: 7 msec
;; SERVER: 8.8.8.8#53(8.8.8.8)
;; WHEN: Tue Mar 15 01:42:53 UTC 2022
;; MSG SIZE  rcvd: 75
```
