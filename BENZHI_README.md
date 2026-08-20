# ballast-watch Docker 打包说明

镜像用于船舶压载水处理监测服务。

```bash
./build_benzhi_docker.sh ballast-watch-bug-N linux/amd64
./build_benzhi_docker.sh ballast-watch-bug-N linux/arm64
```

基础镜像固定为 `golang:1.22-bookworm`，工具链使用 `GOTOOLCHAIN=local`，数据库使用容器外挂载的 SQLite 文件。
