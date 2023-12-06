# Swan Provider
[![Discord](https://img.shields.io/discord/770382203782692945?label=Discord&logo=Discord)](https://discord.gg/MSXGzVsSYf)
[![Twitter Follow](https://img.shields.io/twitter/follow/0xfilswan)](https://twitter.com/0xfilswan)
[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg)](https://github.com/RichardLitt/standard-readme)

- 加入我们的 [Discord](https://discord.com/invite/KKGhy8ZqzK)  频道，参与讨论、了解Swan Provider的更多动态和更新。
- 关注查看我们的 [Blog](https://blog.filswan.com/),了解最新发布和公告。

## 目录

- [功能](#功能)
- [前提条件](#前提条件)
- [安装](#安装)
- [配置并运行](#配置并运行)
- [许可证](#许可证)

## 功能

Swan Provider 接收来自 Swan Platform 的离线订单。提供以下功能：

* 使用 aria2 作为下载服务自动下载离线订单；
* 发布 `PublishStorageDeals` 消息，将资金从抵押钱包转移到 StorageMarketActor 的托管账户（当设置 `Market_version=1.2` 时）
* 使用 Market(version=1.1 和 1.2) 导入下载完成的订单；
* 同步订单状态到 [Swan Platform](https://console.filswan.com/#/dashboard)，以便用户和存储提供商了解订单状态的实时变化；
* 从 FilSwan 竞价市场自动竞拍任务；
* 从 FilSwan Platform 获得手动竞价任务。

## 前提条件
- 启动 Lotus-miner
- Aria2 服务

#### 启动 Lotus-miner
在启动 `swan-provider`之前，请确保 `Lotus-miner` 正常运行。您需要使用 `Lotus-miner` 令牌导入订单。
```shell
lotus-miner auth create-token --perm write
```
注意，请保持 `Lotus-miner` 在后台运行!
生成的令牌位于 `$LOTUS_MINER_PATH/token`
请参考：[Lotus: API tokens](https://lotus.filecoin.io/reference/basics/api-access/)

#### Aria2 服务
```shell
sudo apt install aria2
```

## 安装
您可以通过环境变量设置 `$SWAN_PATH`，默认 `~/.swan`：

```shell
export SWAN_PATH="/data/.swan"
```

### 选项:one: **预构建包**: 参照 [release assets](https://github.com/filswan/go-swan-provider/releases)
####  构建指南
```shell
wget --no-check-certificate https://github.com/filswan/go-swan-provider/releases/download/v2.3.0/install.sh
chmod +x ./install.sh
./install.sh
```
#### 配置和运行
- 编辑配置文件 **~/.swan/provider/config.toml**, 参考 [此处](#配置并运行)
- 在后台运行 `swan-provider`

```
ulimit -SHn 1048576
export SWAN_PATH="/data/.swan"
nohup swan-provider-2.3.0-linux-amd64 daemon >> swan-provider.log 2>&1 & 
```
### 选项:two: 从源代码构建
构建 `swan-provider` 需要安装以下依赖包:
```
curl -sL https://deb.nodesource.com/setup_16.x | sudo -E bash -
```
```
sudo apt-get install -y nodejs
```
```
sudo apt install mesa-opencl-icd ocl-icd-opencl-dev gcc git bzr jq pkg-config curl clang build-essential hwloc libhwloc-dev wget -y && sudo apt upgrade -y
```
- Go(需要 **1.20+**)
```
wget -c https://golang.org/dl/go1.21.4.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
```
```
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc && source ~/.bashrc
```
- Rustup
```
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

####  构建指南
```shell
git clone https://github.com/filswan/go-swan-provider.git
cd go-swan-provider
git checkout release-2.3.0
./build_from_source.sh
```

## 配置并运行

您需要根据不同的 `market_version` 进行配置。
```
port = 8888
release = true                               	# 在 release 模式下运行时: 设置为 true, 否则设置为 false and 环境变量 GIN_MODE not to release

[lotus]
client_api_url = "http://[ip]:[port]/rpc/v0"    # lotus 客户端 web API 的 Url, 通常 [port] 为 1234
client_api_token = ""                           # lotus 客户端 web API 的 token (lotus auth api-info --perm=admin)
market_api_url = "http://[ip]:[port]/rpc/v0"   	# lotus market web API 的 Url，通常 [port] 为 2345，当 market 和 miner 没有分离的时候, 它也是 miner web API 的 URL
market_access_token = ""                        # lotus market web API 的 token，当 market 和 miner 没有分离的时候, 它也是 miner web API 的 token
max_sealing = 5                                 # Limit the number of concurrently executing tasks for sealing sectors.
max_addPiece = 2                                # Limit the number of concurrently executing tasks for addPiece.

[aria2]
aria2_download_dir = "%%ARIA2_DOWNLOAD_DIR%%"   # 离线订单文件的下载目录
aria2_candidate_dirs = ["/tmp"]                 # 离线订单所需 CAR 文件的查询目录
aria2_host = "127.0.0.1"                        # Aria2 服务地址
aria2_port = 6800                               # Aria2 服务端口
aria2_secret = "my_aria2_secret"                # 必须与 aria2.conf 中的 RPC-secure 为同一个值
aria2_auto_delete_car_file= false               # 当订单状态变为 Active 或 Error 时，CAR文件将被自动删除
aria2_max_downloading_tasks = 10                # Aria2 最大并行下载数 默认：10

[main]
market_version = "1.2"                          # 订单版本为 1.1 或 1.2, 配置(market_version=1.1) 将被弃用，很快会被删除 (默认: "1.2")，如果设置为 1.2，需要设置 [market] 部分
api_url = "https://go-swan-server.filswan.com"  # Swan API 地址。生产环境地址为 "https://go-swan-server.filswan.com"
api_key = ""                                    # api 密钥。从 Filswan -> "My Profile"->"Developer Settings"获得
access_token = ""                               # Token，从 Filswan -> "My Profile"->"Developer Settings"获得
miner_fid = "f0xxxx"                            # Filecoin MinerID, 此 ID 必须被添加到 Swan Storage providers 列表，添加方式：Swan Platform -> "My Profile" -> "As Storage Provider" -> "Manage" -> "Add"
import_interval = 600                           # 600 秒或 10 分钟。导入每个订单的时间间隔
scan_interval = 600                             # 600 秒或 10 分钟。扫描所有进行中的订单并在 Swan Platform 上更新状态的时间间隔
api_heartbeat_interval = 300                    # 300 秒或 5 分钟。发送心跳的时间间隔。

[bid]
bid_mode = 1					# 0: 手动, 1: 自动
expected_sealing_time = 1920			# 1920 epoch 或 16 小时。 订单的预期封装时长。过早开始将会被拒绝。
start_epoch = 2880            			# 2880 epoch 或 24 小时。 当前 epoch 的相对值
auto_bid_deal_per_day = 600   			# 上面配置的 miner_fid 每日可接受自动竞价订单的最大数量

[market]
collateral_wallet = ""                          # 质押订单用到的钱包
publish_wallet = ""                             # 发送 PublishStorageDeals 消息的钱包地址
```
**(1) `market_version = "1.1"` 时**，存储提供商会使用 lotus 内置的 Market 导入订单。因此，无需设置 `[market]` 部分。

**(2) `market_version = "1.2(推荐)"` 时**, 存储提供商会使用 `Boost` 中的 Market 导入订单, 因此须确保存储提供商状态是可接入的。具体的配置步骤如下：
- 在 miner 配置中禁用 market 子系统：
```
vi $LOTUS_MINER_PATH/config.toml
```
```
[Subsystems] 
 EnableMarkets = false
```
- 配置 `$SWAN_PATH/provider/config.toml` 中的 `[market]` 部分
- 初始化 Market repo 到 `$SWAN_PATH/provider/boost`：
```
export SWAN_PATH="/data/.swan"
swan-provider daemon 
```
- 配置 `[Libp2p]` 部分

  (1) 确保 `swan-provider` 和 `boostd` 没有运行
  ```
  kill -9 $(ps -ef | grep -E 'swan-provider|boostd' | grep -v grep | awk '{print$2}' )
  ```
  (2) 编辑 boost 的配置文件`$SWAN_PATH/provider/boost/config.toml`：
  ```
  [Libp2p]
      ListenAddresses = ["/ip4/0.0.0.0/tcp/24001", "/ip6/::/tcp/24001"]   # Binding address for the libp2p host
    AnnounceAddresses = ["/ip4/209.94.92.3/tcp/24001"]                  # Addresses to explicitly announce to other peers. If not specified, all interface addresses are announced
  ```
  (3) 在后台运行 `swan-provider`
  ```
  ulimit -SHn 1048576
  export SWAN_PATH="/data/.swan"
  nohup swan-provider daemon >> swan-provider.log 2>&1 & 
  ```
- 发布存储提供商的 Multiaddrs 和 PeerID:
	- 获取方式： `boostd --boost-repo=$SWAN_PATH/provider/boost net listen`
  ```
  lotus-miner actor set-addrs /ip4/<ip>/tcp/<port>   
  ```
	- 获取方式： `boostd --boost-repo=$SWAN_PATH/provider/boost net id`
  ```
  lotus-miner actor set-peer-id <PeerID> 
  ```
- 设置接单条件
 ```
 export SWAN_PATH="/data/.swan"
 swan-provider set-ask --price=0 --verified-price=0 --min-piece-size=1048576 --max-piece-size=34359738368
 ```
- 设置 `[market].publish_wallet` 为控制地址：
 ```
 export OLD_CONTROL_ADDRESS=`lotus-miner actor control list  --verbose | awk '{print $3}' | grep -v key | tr -s '\n'  ' '`
 ``` 
 ```
 lotus-miner actor control set --really-do-it $[market].publish_wallet $OLD_CONTROL_ADDRESS
 ```
- 给 `collateral_wallet` Market Actor 充值
 ```
 lotus wallet market add --from=<YOUR_WALLET> --address=<collateral_wallet> <amount>
 ```
>#### **注意**:
>- 日志位于 `./logs` 目录下

## 与 Swan Provider 交互
`swan-provider` 命令让您可以与运行中的 Swan Provider 进行交互。
检查您当前使用的 swan-provider 版本
```
swan-provider version
```
## 常见问题及解决方案
* `My aria is not downloaded`

  请检查 aria2 是否在运行中
  ```shell
  ps -ef | grep aria2
  ```

* `error msg="no response from swan platform”`

  请检查您的 `api_url` 是否正确, 应为 `https://go-swan-server.filswan.com`

## 帮助

有关使用问题，请在 [Discord 频道](http://discord.com/invite/KKGhy8ZqzK) 中联系 Swan Provider 团队，或在 GitHub 上创建一个新Issue。

## 许可证

[Apache](https://github.com/filswan/go-swan-provider/blob/main/LICENSE)
