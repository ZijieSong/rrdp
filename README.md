SIMPLE-PROXY
===========

SIMPLE-PROXY 是一个小巧的代理，用以将本地流量访问代理到远程，或暴露本地端口到远程实例

## Proxy
sudo rrdp --remote 10.211.55.5:22300 proxy --localPorts 8004 --localPorts 8005

将本地的8004，8005端口映射到10.211.55.5，即访问localhost:8004相当于访问10.211.55.5:8004

## Expose
sudo rrdp --remote 10.211.55.5:22300 expose --exposedPorts 8008:8008 --exposedPorts 8009:8009

暴露本地8008，8009 到 10.211.55.5 的8008，8009，这样client访问10.211.55.5:8008时相当于访问本地的8008端口