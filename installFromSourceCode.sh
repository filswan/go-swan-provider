#!/bin/bash

git clone https://github.com/filswan/go-swan-provider.git
cd go-swan-provider
git checkout main

sed -i 's/%%USER%%/'${USER}'/g' ./aria2c.service   # Set User & Group to value of $USER

if [ ! -d "/etc/aria2" ]; then
    sudo mkdir /etc/aria2
    echo "/etc/aria2 created"
else
    echo "/etc/aria2 exists"
fi

sudo chown $USER:$USER /etc/aria2/             # Change user authority to current user
sudo touch /etc/aria2/aria2.session            # Create a session file
sudo cp ./config/aria2.conf /etc/aria2/        # Copy config file
sudo cp ./aria2c.service /etc/systemd/system/  # Copy service file
sudo systemctl enable aria2c.service           # Set to start Aria2 automatically
sudo systemctl start aria2c.service            # Start Aria2

make
chmod +x ./build/swan-provider
./build/swan-provider                          # Run swan provider

