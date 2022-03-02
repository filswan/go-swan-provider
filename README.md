# Swan Provider
[![Made by FilSwan](https://img.shields.io/badge/made%20by-FilSwan-green.svg)](https://www.filswan.com/)
[![Chat on Slack](https://img.shields.io/badge/slack-filswan.slack.com-green.svg)](https://filswan.slack.com)
[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg)](https://github.com/RichardLitt/standard-readme)

- Join us on our [public Slack channel](https://filswan.slack.com) for news, discussions, and status updates. 
- [Check out our medium](https://filswan.medium.com) for the latest posts and announcements.

## Table of Contents

- [Features](#Features)
- [Prerequisite](#Prerequisite)
- [Installation](#Installation)
- [Config](#Config)
- [License](#license)

## Features

Swan Provider listens to offline deals that come from Swan platform. It provides the following functions:

* Download offline deals automatically using aria2 for downloading service.
* Import deals using lotus once download completed.
* Synchronize deal status to [Swan Platform](https://www.filswan.com/), so that both clients and miners will know the status changes in realtime. 
* Auto bid task from FilSwan bidding market.

## Prerequisite
- lotus-miner
- aria2
### arial2 installation
```shell
sudo apt install aria2
```
### Lotus Miner Token creation
Lotus miner token is used for importing deal for swan provider
```shell
lotus-miner auth create-token --perm write
```
Note that the Lotus Miner need to be running in the background!
The created token located at $LOTUS_STORAGE_PATH/token 

Reference: [Lotus: API tokens](https://docs.filecoin.io/build/lotus/api-tokens/#obtaining-tokens)

## Installation
### Option:one: **Prebuilt package**: See [release assets](https://github.com/filswan/go-swan-provider/releases)
```shell
mkdir swan-provider
cd swan-provider
wget https://github.com/filswan/go-swan-provider/releases/download/v0.2.1-rc1/install.sh
chmod +x ./install.sh
./install.sh
```

### Option:two: Source Code
:bell:**go 1.16+** is required
```shell
git clone https://github.com/filswan/go-swan-provider.git
cd go-swan-provider
git checkout <release_branch>
./build_from_source.sh
```

### :bangbang: Important
After installation, swan-provider maybe quit due to lack of configuration. Under this situation, you need
- :one: Edit config file **~/.swan/provider/config.toml** to solve this.
- :two: Execute **swan-provider** using one of the following commands
```shell
./swan-provider-0.2.1-rc1-unix   #After installation from Option 1
./build/swan-provider            #After installation from Option 2
```


### Note
- Logs are in directory ./logs
- You can add `nohup` before `./swan-provider` to ignore the HUP (hangup) signal and therefore avoid stop when you log out.
- You can add `>> swan-provider.log` in the command to let all the logs output to `swan-provider.log`.
- You can add `&` at the end of the command to let the program run in background.
- Such as:
```shell
nohup ./swan-provider-0.2.1-rc1-unix >> swan-provider.log &   #After installation from Option 1
nohup ./build/swan-provider >> swan-provider.log &            #After installation from Option 2
```


## Config
- **port:** Default `8888`, web api port for extension in future
- **release:** Default `true`, when work in release mode: set this to true, otherwise to false and enviornment variable GIN_MODE not to release

### [lotus]
- :bangbang:**client_api_url:** Url of lotus client web api, such as: `http://[ip]:[port]/rpc/v0`, generally the `[port]` is `1234`. See [Lotus API](https://docs.filecoin.io/reference/lotus-api/)
- :bangbang:**market_api_url:** Url of lotus market web api, such as: `http://[ip]:[port]/rpc/v0`, generally the `[port]` is `2345`. When market and miner are not separate, it is also the url of miner web api. See [Lotus API](https://docs.filecoin.io/reference/lotus-api/)
- :bangbang:**market_access_token:** Access token of lotus market web api. When market and miner are not separate, it is also the access token of miner access token. See [Obtaining Tokens](https://docs.filecoin.io/build/lotus/api-tokens/#obtaining-tokens)

### [aria2]
- **aria2_download_dir:** Directory where offline deal files will be downloaded for importing
- **aria2_host:** Aria2 server address
- **aria2_port:** Aria2 server port
- **aria2_secret:** Must be the same value as rpc-secret in `aria2.conf`

### [main]
- **api_url:** Swan API address. For Swan production, it is "https://api.filswan.com"
- :bangbang:**miner_fid:** Your filecoin Miner ID
- **import_interval:** 600 seconds or 10 minutes. Importing interval between each deal.
- **scan_interval:** 600 seconds or 10 minutes. Time interval to scan all the ongoing deals and update status on Swan platform.
- :bangbang:**api_key:** Your api key. Acquire from [Swan Platform](https://www.filswan.com/) -> "My Profile"->"Developer Settings". You can also check the Guide.
- :bangbang:**access_token:** Your access token. Acquire from [Swan Platform](https://www.filswan.com/) -> "My Profile"->"Developer Settings". You can also check the Guide.
- **api_heartbeat_interval:** 300 seconds or 5 minutes. Time interval to send heartbeat.

### [bid]
- **bid_mode:** 0: manual, 1: auto
- **expected_sealing_time:** 1920 epoch or 16 hours. The time expected for sealing deals. Deals starting too soon will be rejected.
- **start_epoch:** 2880 epoch or 24 hours. Relative value to current epoch
- **auto_bid_deal_per_day:** auto-bid deal limit per day for your miner defined above

# Common Isuse and solutions
* My aria is not downloaded

  Please check if aria2 is running
  ps -ef|grep aria2

* error msg="no response from swan platform”

  Please check your API endpiont is correct, it should be https://api.filswan.com/

## License

[Apache](https://github.com/filswan/go-swan-provider/blob/main/LICENSE)


