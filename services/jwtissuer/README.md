# Peergrine JWTIssuer 服務

此服務利用多種設定選項來管理其行為，包括環境變數、命令行參數和儲存在 Zookeeper 中的設定。該服務會使用預設值初始化，並且可以根據環境和 Zookeeper 設定進行自定義設定。
## 設定概述

該服務有多個設定選項可以通過環境變數或命令行參數進行設置。這些設置包括網路位址、令牌有效時間、Redis 和 Consul 設定、以及 Zookeeper 路徑。
### 設定選項
|設定欄位 |描述 |預設值 |
|--|
|**APP_ID** |服務的唯一識別碼 (可選) |隨機生成 |
|**APP_CLIENTENDPOINT_ADDR** |客戶端端點地址 (可選) |:80 |
|**APP_SERVICEENDPOINT_ADDR** |服務端點地址 (可選) |:50051 |
|**APP_REDIS_ADDR** |Redis 伺服器地址 (可選) |空值 (無 Redis 連接) |
|**APP_BEARER_TOKEN_DURATION** |Bearer 令牌有效時間（秒）(可選) |3600 |
|**APP_REFRESH_TOKEN_DURATION** |Refresh 令牌有效時間（秒）(可選) |7200 |
|**APP_CONSUL_ADDR** |Consul 服務地址 (可選) |空值 |
|**APP_SERVICE_NAME** |服務名稱 (可選) |JWTIssuer |
|**APP_SERVICE_ADDR** |服務地址 (可選) |服務自行判定 |
|**APP_SERVICE_PORT** |服務健康檢查端口 (可選) |4000 |
|**APP_ZOOKEEPER_ADDRS** |Zookeeper 伺服器地址列表（以逗號分隔）(可選) |無預設值 |
|**APP_CONFIG_PATH** |Zookeeper 中的設定路徑 (可選) |無預設值 |
> 若 **APP_CONSUL_ADDR** 為空，則 `APP_SERVICE_NAME`、`APP_SERVICE_ADDR`、`APP_SERVICE_PORT` 三項設置將不會生效。

### Zookeeper 設定

該服務還允許從 Zookeeper 獲取設定，這使得它在分散式環境中能夠進行集中化的設定管理。如果提供了 Zookeeper 地址和路徑，服務將會嘗試從 Zookeeper 獲取設定。

#### Zookeeper 路徑

- 預設的設定路徑為 `/jwtissuer`，您可以使用環境變數 `APP_CONFIG_PATH` 指定不同的路徑。
### 設定流程

1. **Zookeeper**：如果提供了 Zookeeper 地址 (`APP_ZOOKEEPER_ADDRS`)，服務會嘗試從指定的 Zookeeper 路徑讀取設定。如果該路徑不存在，將會寫入預設設定。
2. **環境變數**：可以通過環境變數設置設定值（例如 `APP_CLIENTENDPOINT_ADDR`）。
3. **命令行參數**：也可以通過命令行參數傳遞設定值（例如 `-zookeeper-addrs`）。

### Zookeeper 初始化

若設定路徑在 Zookeeper 中不存在，服務會自動創建該路徑並儲存預設設定。這樣即便未手動設定，服務也能使用預設值啟動。
### 範例

以下是如何設定服務的範例：
#### 使用環境變數

```
export APP_ID="jwt-service-1234"
export APP_CLIENTENDPOINT_ADDR=":8080"
export APP_SERVICEENDPOINT_ADDR=":9090"
export APP_BEARER_TOKEN_DURATION="3600"
export APP_REFRESH_TOKEN_DURATION="7200"
export APP_ZOOKEEPER_ADDRS="zookeeper1:2181,zookeeper2:2181"
export APP_CONFIG_PATH="/jwtissuer"
```

#### 使用命令行參數
```
./app -zookeeper-addrs "zookeeper1:2181,zookeeper2:2181" -config-path "/jwtissuer"
```