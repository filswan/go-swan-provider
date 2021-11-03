#!/bin/bash

CONF_FILE_DIR=${HOME}/.swan/provider
mkdir ${CONF_FILE_DIR}

CONF_FILE_PATH=${CONF_FILE_DIR}/config.toml
echo $CONF_FILE_PATH

if [ -f "${CONF_FILE_PATH}" ]; then
    echo "${CONF_FILE_PATH} exists"
else
    cp ./config/config.toml.example $CONF_FILE_PATH
    sed -i 's/%%ARIA2_DOWNLOAD_DIR%%/'${HOME}'/g' $CONF_FILE_PATH   # Set User & Group to value of $USER
    echo "${CONF_FILE_PATH} created"
fi

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
chmod +x ./build/go-swan-provider
./build/go-swan-provider                          # Run swan provider

