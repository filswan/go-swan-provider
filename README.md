# Swan Provider Tool Guide

## Features:

This provider tool listens to the tasks that come from Swan platform. It provides the following functions:

* Download tasks automatically using Aria2 for downloading service.
* Import deals once download tasks completed.
* Synchronize deal status with Swan platform so that client will know the status changes in realtime.

## Prerequisite
- lotus-miner
- aria2
- go 1.16

## Installation

Install aria2
```shell
sudo apt install aria2
```

## How to use

### Step 1. Run Aria2 as System Service

#### Step 1.1 Set up Aria2:

```shell
sudo mkdir /etc/aria2
# Change user authority to current user
sudo chown $USER:$USER /etc/aria2/
# Create a session file
touch /etc/aria2/aria2.session
# Checkout the source and install 
git clone https://github.com/filswan/go-swan-provider.git
cd go-swan-provider
git checkout release-0.1.0

# Copy config file and service file
cp config/aria2.conf /etc/aria2/
sudo cp aria2c.service /etc/systemd/system/
# Modify /etc/systemd/system/aria2c.service, set User & Group to value of $USER  
sudo vi /etc/systemd/system/aria2c.service

# Set to start Aria2 automatically
sudo systemctl enable aria2c.service
# Start Aria2
sudo systemctl start aria2c.service
```
If modify user is nessecary while the service has been started, service should be reloaded before start.
```shell
sudo systemctl daemon-reload
sudo systemctl start aria2c.service
```

#### Step 1.2 Test Aria2 service from log (Optional)

Check if Aria2 service is successfully started

```shell
journalctl -u aria2c.service -f
```
The output will be like:

```shell
Sep 29 11:07:29 systemd[1]: Started Aria2c download manager.
Sep 29 11:07:29 aria2c[2008]: 09/29 11:07:29 [NOTICE] IPv4 RPC: listening on TCP port 6800
```

The Aira2 service will listen on certain port if installed and started correctly.

### Step 2. Compile Code
```shell
make   # generate binary file and config file to ./build folder
```

### Step 3. Start Swan Provider
```shell
cd build
vi ~/.swan/provider/config.toml   # update configuration
./swan-provider
```

#### Note
- Logs are in directory ./logs
- You can add **nohup** before **./swan-provider** to ignore the HUP (hangup) signal and therefore avoid stop when you log out.
- You can add **&** after **./swan-provider** to let the program run in background.

```shell
nohup ./swan-provider &
```


#### Config Explanation
- **port：** the port for restful api

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
- **api_key:** Your api key. Acquire from Filswan -> "My Profile"->"Developer Settings". You can also check the Guide.
- **access_token:** Your access token. Acquire from Filswan -> "My Profile"->"Developer Settings". You can also check the Guide.
- **api_heartbeat_interval:** 300 seconds or 5 minutes. Time interval to send heartbeat.

##### [bid]
- **bid_mode:** 0: manual, 1: auto
- **expected_sealing_time:** 1920 epoch or 16 hours. The time expected for sealing deals. Deals starting too soon will be rejected.
- **start_epoch:** 2880 epoch or 24 hours. Relative value to current epoch
- **auto_bid_task_per_day:** auto-bid task limit per day for your miner defined above

The deal status will be synchronized on the filwan.com, both client and miner will know the status changes in realtime.
