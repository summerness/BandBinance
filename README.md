安装好go环境


```shell
# 下载依赖
go mod tidy
go mod download 
# 设置自己的ak sk

# 设置自己的飞书告警链接

# 打包
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app/CreateBuyOrderAPP CreateBuyOrderAPP.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app/CreateSellOrderAPP CreateSellOrderAPP.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app/CancelBuyOrderAPP CancelBuyOrderAPP.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app/UpdateSellOrderAPP UpdateSellOrderAPP.go

# 运行二进制程序



```

均是实盘交易
