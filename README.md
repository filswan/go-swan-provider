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
- [License](#license)

## Features:

Swan Provider tool listens to offline deals that come from Swan platform. It provides the following functions:

* Download offline deals automatically using aria2 for downloading service.
* Import deals using lotus once download completed.
* Synchronize deal status to [Swan Platform](https://www.filswan.com/), so that both clients and miners will know the status changes in realtime.

## Prerequisite
- lotus-miner
- aria2
```shell
sudo apt install aria2
```
- go 1.16

## Installation
### Option 1.  **Prebuilt package**: See [release assets](https://github.com/filswan/go-swan-provider/releases)
```shell
wget https://github.com/filswan/go-swan-provider/releases/download/release-0.1.0-beta-rc1/install.sh
chmod +x ./install.sh
./install.sh
```

After installation, swan-provider maybe quit due to lack of configuration. Under this situation, you need:
- Step 1. edit the config file **~/.swan/provider/config.toml** to solve this.
- Step 2. execute **swan-provider** using the following command
```shell
./swan-provider-0.1.0-rc-unix
```

### Option 2.  Source Code
```shell
wget https://github.com/filswan/go-swan-provider/releases/download/release-0.1.0-beta-rc1/installFromSourceCode.sh
chmod +x ./installFromSourceCode.sh
./installFromSourceCode.sh
```

After installation, swan-provider maybe quit due to lack of configuration. Under this situation, you need:
- Step 1. edit the config file **~/.swan/provider/config.toml** to solve this.
- Step 2. execute **swan-provider** using the following command
```shell
./build/swan-provider
```


#### Note
- Logs are in directory ./logs
- You can add **nohup** before **./swan-provider** to ignore the HUP (hangup) signal and therefore avoid stop when you log out.
- You can add **&** after **./swan-provider** to let the program run in background.
- Such as:
```shell
nohup ./swan-provider &
```


#### Config Explanation
- **port：** Default 8888, web api port for extension in future

##### [aria2]
- **aria2_download_dir:** Directory where offline deal files will be downloaded for importing
- **aria2_host:** Aria2 server address
- **aria2_port:** Aria2 server port
- **aria2_secret:** Must be the same value as rpc-secure in aria2.conf

##### [main]
- **api_url:** Swan API address. For Swan production, it is "https://api.filswan.com"
- **miner_fid:** Your filecoin Miner ID
- **import_interval:** 600 seconds or 10 minutes. Importing interval between each deal.
- **scan_interval:** 600 seconds or 10 minutes. Time interval to scan all the ongoing deals and update status on Swan platform.
- **api_key:** Your api key. Acquire from [Swan Platform](https://www.filswan.com/) -> "My Profile"->"Developer Settings". You can also check the Guide.
- **access_token:** Your access token. Acquire from [Swan Platform](https://www.filswan.com/) -> "My Profile"->"Developer Settings". You can also check the Guide.
- **api_heartbeat_interval:** 300 seconds or 5 minutes. Time interval to send heartbeat.

##### [bid]
- **bid_mode:** 0: manual, 1: auto
- **expected_sealing_time:** 1920 epoch or 16 hours. The time expected for sealing deals. Deals starting too soon will be rejected.
- **start_epoch:** 2880 epoch or 24 hours. Relative value to current epoch
- **auto_bid_task_per_day:** auto-bid task limit per day for your miner defined above


## License

[Apache](https://github.com/filswan/go-swan-provider/blob/main/LICENSE)


