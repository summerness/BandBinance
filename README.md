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


## 修改说明

#### go1.16版本中time包不支持 UnixMilli  已将其改为 ..UnixNano() / 1e6

#### 数据迁移 将 store/*.go 中 init() 做了合并, 并补上了 TradeOrder 表的迁移
    func init() {
        err := DB.AutoMigrate(&domain.GridSymbolConfig{})
        if err != nil {
            panic(err)
        }
        err = DB.AutoMigrate(&domain.GridTrade{})
        if err != nil {
            panic(err)
        }
        err = DB.AutoMigrate(&domain.TradeOrder{})
        if err != nil {
            panic(err)
        }
    }

#### 添加了 钉钉通知  notify/DDNotify.go

#### 将 UpdateSellOrderAPP.go 中 time.sleep 修改为了 time.NewTicker 用于当for中的任务耗时较长时 任务的周期时长不会随之增加

#### 将 strategy/Grid.go 中:
    ProcessGridTrades 方法里全部使用配置项symbolConfig  删除了参数 grid *domain.Grid
    方法调用者中删除了newGrid()
    step2: 调用 GetTrades 的地方直接使用了 ProcessGridTrades 删除了 GetTrades
    step3: ProcessGridTrades 频繁调用 使用了sync.Once 保证只执行一次

#### 部分 exchange.go 中 priceMap等map类型的读取只需要读锁 lockPriceMap.RLock() 可提高并发时的效率

#### exchange.go中UpdateBalance()里 个人觉得 for i := range balances  前不需要用len() 判断对象的长度 如果len为0 会自动跳过for循环

#### 添加了config.yaml 作为配置文件 使得程序在打包后也可以更改一些参数的配置

#### 取消订单中 3小时以前 的写法有点繁琐 替换为了 time.Now().Add(-3 * time.Hour)

## 小结
    关于代码修改:大致读完了4个app的流程并顺手改了些代码  其中有一些是运行逻辑问题 但大部分只是个人习惯... 
    改动比较大 不一定对 可能也有我理解不到位的地方 如果有问题还望指出
    
    关于策略: 看了代码后对网格的理解更进了一步 大致的理解是:
        在一定价格区间内 按照等百分比划定区间 
        价格触发区间下限的时候买入 若成功则挂一个区间上限价格的卖出单

        优点:价格上下波动时有部分可观收益
        缺点:遇到单边行情 损失会比较大
    
    关于改进:   
        对于网格打算将基础货币由busd改为任意币种 比如做eth和btc间的 
        将增加多币种的动态平衡策略    
    

