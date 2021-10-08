#!/bin/bash

wget https://github.com/filswan/go-swan-provider/releases/download/release-0.1.0-beta-rc1/swan-provider-0.1.0-rc-unix
wget https://github.com/filswan/go-swan-provider/releases/download/release-0.1.0-beta-rc1/aria2.conf
wget https://github.com/filswan/go-swan-provider/releases/download/release-0.1.0-beta-rc1/aria2c.service

sed -i 's/%%USER%%/'${USER}'/g' ./aria2c.service   # Set User & Group to value of $USER

if [ ! -d "/etc/aria2" ]; then
    sudo mkdir /etc/aria2
    echo "/etc/aria2 created"
else
    echo "/etc/aria2 exists"
fi

sudo chown $USER:$USER /etc/aria2/             # Change user authority to current user
sudo touch /etc/aria2/aria2.session            # Create a session file
sudo cp ./aria2.conf /etc/aria2/               # Copy config file
sudo cp ./aria2c.service /etc/systemd/system/    # Copy service file
sudo systemctl enable aria2c.service           # Set to start Aria2 automatically
sudo systemctl start aria2c.service            # Start Aria2

chmod +x ./swan-provider-0.1.0-rc-unix
./swan-provider-0.1.0-rc-unix                  # Run swan provider

