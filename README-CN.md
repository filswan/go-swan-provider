# Swan Provider
[![Made by FilSwan](https://img.shields.io/badge/made%20by-FilSwan-green.svg)](https://www.filswan.com/)
[![Chat on Slack](https://img.shields.io/badge/slack-filswan.slack.com-green.svg)](https://filswan.slack.com)
[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg)](https://github.com/RichardLitt/standard-readme)

- 加入FilSwan的[Slack频道](https://filswan.slack.com)，了解新闻、讨论和状态更新。 
- 查看FilSwan的[Medium](https://filswan.medium.com)，获取最新动态和公告。

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
* 来自FilSwan竞价市场的自动竞价任务。

## 前提
- lotus-miner
- aria2
### 安装arial2
```shell
sudo apt install aria2
```
### Lotus Miner 令牌生成
Lotus miner令牌用于为Swan Provider导入交易
```shell
lotus-miner auth create-token --perm write
```
注意，Lotus Miner需要在后台运行！生成的令牌位于 `$LOTUS_MINER_PATH/token` 

参考: [Lotus: API tokens](https://docs.filecoin.io/build/lotus/api-tokens/#obtaining-tokens)

## 安装
### 选择:one: **预构建软件包**: 参照 [release assets](https://github.com/filswan/go-swan-provider/releases)
```shell
mkdir swan-provider
cd swan-provider
wget --no-check-certificate https://github.com/filswan/go-swan-provider/releases/download/v0.2.1/install.sh
chmod +x ./install.sh
./install.sh
```

### 选择:two: 从源代码构建
:bell: 需要 **go 1.16+** 
```shell
git clone https://github.com/filswan/go-swan-provider.git
cd go-swan-provider
git checkout <release_branch>
./build_from_source.sh
```

### :bangbang: 重要事项
安装后，swan-provider可能会由于缺少配置文件而退出。此时，需要
- :one: 通过编辑配置文件 **~/.swan/provider/config.toml** 来解决
- :two: 用下面其中一个命令执行 **swan-provider**
```shell
./swan-provider-0.2.1-linux-amd64   #从选择一安装完以后
./build/swan-provider               #从选择二安装完以后
```


### 注意
- 日志位于目录 ./logs
- 可以使用如下方式运行以防止程序退出，并将日志打印到单独的文件中：

```shell
nohup ./swan-provider-0.2.1-linux-amd64 >> swan-provider.log 2>&1 &   #从选择一安装完以后
nohup ./build/swan-provider >> swan-provider.log 2>&1 &               #从选择二安装完以后
```


## 配置
- **port：** 默认 `8888`，未来扩展的 web api 端口
- **release：** 默认为 `true`，在release模式下工作时设置为true；否则为false，此时环境变量`GIN_MODE`不会释放

### [lotus]
- :bangbang:**client_api_url:** lotus client web api的Url, 如: `http://[ip]:[port]/rpc/v0`, 一般 `[port]` 为 `1234`。 参照 [Lotus API](https://docs.filecoin.io/reference/lotus-api/)
- :bangbang:**market_api_url:** lotus market web api的Url, 如: `http://[ip]:[port]/rpc/v0`, 一般 `[port]` 为 `2345`。 当market和miner没有分离时，这也是miner的web api的url。参照 [Lotus API](https://docs.filecoin.io/reference/lotus-api/)
- :bangbang:**market_access_token:** lotus market web api访问令牌。当market和miner没有分离时，这也是miner访问令牌的访问令牌。参照 [Obtaining Tokens](https://docs.filecoin.io/build/lotus/api-tokens/#obtaining-tokens)

### [aria2]
- **aria2_download_dir:** 离线交易文件进行下载以供导入的目录
- **aria2_host:** Aria2 服务器地址
- **aria2_port:** Aria2 服务器端口
- **aria2_secret:** 必须与 `aria2.conf` 的rpc-secret值相同

### [主要配置]
- **api_url:** Swan API 地址 "https://api.filswan.com"。
- :bangbang:**miner_fid:** filecoin 矿工 ID。
- **import_interval:** 600秒或10分钟。每笔交易之间的导入间隔。
- **scan_interval:** 600秒或10分钟。在Swan平台上扫描所有正在进行的交易并更新状态的时间间隔。
- :bangbang:**api_key:**  api key。可以通过 [Swan Platform](https://www.filswan.com/) -> "我的资料"->"开发者设置" 获得，也可以访问操作指南。
- :bangbang:**access_token:** 访问令牌。可以通过 [Swan Platform](https://www.filswan.com/) -> "我的资料"->"开发者设置"获得，也可以访问操作指南。
- **api_heartbeat_interval:** 300秒或5分钟。发送心跳的时间间隔。
- **purge_interval:** 600秒或10分钟。清除交易状态为“完成”、“导入失败”或“交易过期”的已下载的car文件的时间间隔。

### [竞价]
- **bid_mode:** 0: 手动，1: 自动
- **expected_sealing_time:** 1920 epoch或16小时。达成交易的预期时间。过早开始交易将被拒绝。
- **start_epoch:** 2880 epoch或24小时。当前epoch的相对值。
- **auto_bid_task_per_day:** 上述定义的矿工每天的自动竞价任务限制。 


## 许可证

[Apache](https://github.com/filswan/go-swan-provider/blob/main/LICENSE)


