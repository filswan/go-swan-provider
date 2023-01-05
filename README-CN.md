# Swan Provider
[![Discord](https://img.shields.io/discord/770382203782692945?label=Discord&logo=Discord)](https://discord.gg/MSXGzVsSYf)
[![Twitter Follow](https://img.shields.io/twitter/follow/0xfilswan)](https://twitter.com/0xfilswan)
[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg)](https://github.com/RichardLitt/standard-readme)

`Web3 service providers`提供多种区块链服务，包括但不限于Filecoin、AR和Polygon的RPC服务。 通过FilSwan服务的提供, 我们可以连接到web3服务市场，使web3服务的提供变的更容易。
典型的Web3市场有:

 - Filecoin 网络的数据订单市场
 - Pocket 网络的RPC市场(即将实现···)


## 目录

- [特性](#Features)
- [前提条件](#Prerequisite)
- [安装](#Installation)
- [配置](#Config)
- [许可证](#license)

## 特性

Swan Provider监听来自Swan平台的离线交易。提供以下功能：

* 使用aria2作为下载服务自动下载离线交易。
* 下载完成后使用lotus导入交易。
* 同步交易状态到 [Swan Platform](https://www.filswan.com/)，让客户端和矿工了解状态的实时变化。
* 来自FilSwan竞价市场的自动或手动竞价任务。

## 前提
- lotus-miner
- aria2
### 安装 Aria2 
```shell
sudo apt install aria2
```
### Lotus Miner 令牌生成
Lotus miner令牌用于为Swan Provider导入交易
```shell
lotus-miner auth create-token --perm write
```
注意，Lotus Miner需要在后台运行！生成的令牌位于 `$LOTUS_MINER_PATH/token`

参考: [Lotus: API tokens](https://lotus.filecoin.io/reference/basics/api-access/)

## 安装
### 选择:one: **预构建软件包**: 参照 [release assets](https://github.com/filswan/go-swan-provider/releases)
####  构建方法
```shell
mkdir swan-provider
cd swan-provider
wget --no-check-certificate https://raw.githubusercontent.com/filswan/go-swan-provider/release-2.0.0/install.sh
chmod +x ./install.sh
./install.sh
```
#### 配置并运行
- 编辑配置文件 **~/.swan/provider/config.toml**, 参照 [配置](#Config) 部分
- 后台运行 `swan-provider` 

```
nohup ./swan-provider-2.0.0-linux-amd64 daemon >> swan-provider.log 2>&1 & 
```
### 选择:two: 从源代码构建
####  构建指引
```shell
git clone https://github.com/filswan/go-swan-provider.git
cd go-swan-provider
git checkout release-2.0.0
./build_from_source.sh
```
#### 配置并运行
- 编辑配置文件 **~/.swan/provider/config.toml**, 参照 [配置](#Config) 部分
- 后台运行 `swan-provider`

```
nohup ./swan-provider daemon >> swan-provider.log 2>&1 & 
```
#### 注意:
- 日志位于目录 ./logs
- 需要 **go 1.16+**
  `
## 配置
- **port:** 默认 `8888`，web api 端口
- **release:** 默认为 `true`, 在release模式下工作时设置为true；否则为false

### [lotus]
- **client_api_url:** lotus 客户端的web api对应的Url, 比如 `http://[ip]:[port]/rpc/v0`, 通常来说 `[port]` 是 `1234`. 参照 [Lotus API](https://docs.filecoin.io/reference/lotus-api/)
- **market_api_url:** lotus 客户端的web api对应的Url, 比如 `http://[ip]:[port]/rpc/v0`, 通常来说 `[port]` 是 `2345`. 当market和miner没有分离时，这也是miner访问令牌的访问令牌. 参照 [Lotus API](https://docs.filecoin.io/reference/lotus-api/)
- **market_access_token:** lotus market web api的访问令牌. 当market和miner没有分离时，这也是miner访问令牌的访问令牌. 参照 [Obtaining Tokens](https://docs.filecoin.io/build/lotus/api-tokens/#obtaining-tokens)

### [aria2]
- **aria2_download_dir:** 离线交易文件进行下载以供导入的目录
- **aria2_host:** Aria2 服务器地址
- **aria2_port:** Aria2 服务器端口
- **aria2_secret:** 必须与 `aria2.conf` 的rpc-secret值相同
- **aria2_auto_delete_car_file**: 交易状态变为 Active 或 Error 后, 对应的 CAR 文件会被自动删除， 默认: false
- **aria2_max_downloading_tasks**: Aria2 任务最大下载量, 默认: 10

### [main]
- **api_url:** Swan API 地址. 对于 Swan production, 地址为 `https://go-swan-server.filswan.com`
- :bangbang:**miner_fid:** Filecoin 矿工ID, 须被添加到 Swan Storage Providers 列表， 添加方式:  [Swan Platform](https://console.filswan.com/#/dashboard) -> "个人信息" -> "作为存储服务商" -> "管理" -> "添加" 。
- **import_interval:** 600秒或10分钟。每笔交易之间的导入间隔。
- **scan_interval:** 600秒或10分钟。在Swan平台上扫描所有正在进行的交易并更新状态的时间间隔。
- **api_key:** api key。可以通过 [Swan Platform](https://console.filswan.com/#/dashboard) -> "个人信息"->"开发人员设置" 获得，也可以访问操作指南。
- **access_token:** 访问令牌。可以通过 [Swan Platform](https://console.filswan.com/#/dashboard) -> "个人信息"->"开发人员设置". 可以访问操作指南查看。
- **api_heartbeat_interval:** 300 秒或 5 分钟. 发送心跳的时间间隔.

### [bid]
- **bid_mode:** 0: 手动, 1: 自动
- **expected_sealing_time:**  默认: 1920 epoch 或 16 小时. 封装交易的预期时间。过早开始交易将被拒绝。
- **start_epoch:** 默认: 2880 epoch 或 24 小时. 当前 epoch 的相对值。
- **auto_bid_deal_per_day:** 上述配置的矿工每天的自动竞价任务限制。

## Swan Provider 命令
 用 `./swan-provider` 命令，与运行中的 swan provider 进程进行交互.
检查 swan-provider 当前的版本
```
./swan-provider version
```
## 常见问题及解决方案
* aria2无法下载

  请检查aria2是否在运行
  ```shell
  ps -ef | grep aria2
  ```

* `error msg="no response from swan platform”`

  请检查你的 `api_url` 是否正确，应为  `https://go-swan-server.filswan.com`
## 帮助

如有任何使用问题，请在 [Discord 频道](http://discord.com/invite/KKGhy8ZqzK) 联系 Swan Provider 团队或在Github上创建新的问题.

## 许可证

[Apache](https://github.com/filswan/go-swan-provider/blob/main/LICENSE)
