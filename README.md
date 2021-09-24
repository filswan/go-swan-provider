# Swan Provider Tool Guide

## Features:

This provider tool listens to the tasks that come from Swan platform. It provides the following functions:

* Download tasks automatically using Aria2 for downloading service.
* Import deals once download tasks completed.
* Synchronize deal status with Swan platform so that client will know the status changes in realtime.

## Prerequisite
- Lotus-miner
- Aria2

## Config
* ./config/config.toml.example
```shell
port = 8888

[aria2]
aria2_download_dir = ""   # Directory where offline deal files will be downloaded for importing
aria2_host = "127.0.0.1"  # Aria2 server address
aria2_port = 6800         # Aria2 server port
aria2_secret = "my_aria2_secret"  # Must be the same value as rpc-secure in aria2.conf

[main]
api_url = "https://api.filswan.com"  # Swan API address. For Swan production, it is "https://api.filswan.com"
miner_fid = "f0xxxx"          # Your filecoin Miner ID
import_interval = 600         # 600 seconds or 10 minutes. Importing interval between each deal.
scan_interval = 600           # 600 seconds or 10 minutes. Time interval to scan all the ongoing deals and update status on Swan platform.
api_key = ""                  # Your api key. Acquire from Filswan -> "My Profile"->"Developer Settings". You can also check the Guide.
access_token = ""             # Your access token. Acquire from Filswan -> "My Profile"->"Developer Settings". You can also check the Guide.
api_heartbeat_interval = 600  # 600 seconds or 10 minutes. Time interval to send heartbeat.

[bid]
bid_mode = 1                  # 0: manual, 1: auto
expected_sealing_time = 1920  # 1920 epoch or 16 hours. The time expected for sealing deals. Deals starting too soon will be rejected.
start_epoch = 2880            # 2880 epoch or 24 hours. Relative value to current epoch
auto_bid_task_per_day = 20    # auto-bid task limit per day for your miner defined above
```

## Installation

Install Aria2
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
git clone https://github.com/filswan/swan-provider

cd swan-provider

# Copy config file and service file
cp config/aria2.conf /etc/aria2/
sudo cp aria2c.service /etc/systemd/system/
# Modify the aria2c.service file in /etc/systemd/system/

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
Jul 30 03:00:00 systemd[1]: Started Aria2c download manager.
Jul 30 03:00:00 aria2c[2433312]: 07/30 03:00:00 [NOTICE] IPv4 RPC: listening on TCP port 6800
```

The Aira2 service will listen on certain port if installed and started correctly.

### Step 2. Download code
```shell
git clone https://github.com/filswan/go-swan-provider.git
```
### Step 3. Compile Code
```shell
cd go-swan-provider
make help    # view how to use make tool
make clean   # remove generated binary file and config file
make test    # Run unit tests
make build   # generate binary file and config file
```

### Step 4. Start Swan Provider
```shell
cd build
vi ./config/config.toml   # fill valid configuration
nohup ./swan-provider > ./swan-provider.log &
```

The deal status will be synchronized on the filwan.com, both client and miner will know the status changes in realtime.
