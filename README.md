# TOOLS

### upnp

gupnp 是 upnp 操作工具,构建二进制可执行文件。

```shell
cd gupnp
#构建arm
CGO_ENABLED=0 GOARCH=arm64 GOOS=linux CGO_ENABLED=0 go build
#构建x86
CGO_ENABLED=0 GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build
```

- 查看

```shell
./gupnp --help
```

### nat-checker

网卡 NAT 类型检测

```shell
python3 nat_checker_std.py --help
```
