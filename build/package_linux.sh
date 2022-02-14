CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app/CreateBuyOrderAPP CreateBuyOrderAPP.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app/CreateSellOrderAPP CreateSellOrderAPP.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app/CancelBuyOrderAPP CancelBuyOrderAPP.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app/UpdateSellOrderAPP UpdateSellOrderAPP.go
