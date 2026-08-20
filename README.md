# ballast-watch

船舶压载水处理监测服务，面向 BWTS 处理周期、采样读数、排放合规和设备告警。

```bash
go build ./...
```

默认监听 `:8080`，SQLite 文件为 `ballast.db`，看板地址为 `/web/index.html`。
