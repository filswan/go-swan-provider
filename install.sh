#!/bin/bash

BINARY_NAME=swan-provider-2.1.0-rc1-linux-amd64
TAG_NAME=v2.1.0-rc1
URL_PREFIX=https://github.com/filswan/go-swan-provider/releases/download

wget --no-check-certificate ${URL_PREFIX}/${TAG_NAME}/${BINARY_NAME}
wget --no-check-certificate ${URL_PREFIX}/${TAG_NAME}/aria2.conf
wget --no-check-certificate ${URL_PREFIX}/${TAG_NAME}/aria2c.service
wget --no-check-certificate ${URL_PREFIX}/${TAG_NAME}/config.toml.example

sudo install -C $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

CONF_FILE_DIR=${HOME}/.swan/provider
SWAN_PATH=$(echo ${SWAN_PATH})
if [ -n "$SWAN_PATH" ]; then
  CONF_FILE_DIR=$SWAN_PATH/provider
fi

mkdir -p ${CONF_FILE_DIR}

current_create_time=`date +"%Y%m%d%H%M%S"`

if [ -f "${CONF_FILE_DIR}/config.toml"  ]; then
    mv ${CONF_FILE_DIR}/config.toml  ${CONF_FILE_DIR}/config.toml.${current_create_time}
    echo "The previous configuration files have been backed up: ${CONF_FILE_DIR}/config.toml.${current_create_time}"

    cp ./config.toml.example ${CONF_FILE_DIR}/config.toml
    echo "${CONF_FILE_DIR}/config.toml created"
    ARIA2_DOWNLOAD_DIR=${CONF_FILE_DIR}/download
    sed -i 's@%%ARIA2_DOWNLOAD_DIR%%@'${ARIA2_DOWNLOAD_DIR}'@g' ${CONF_FILE_DIR}/config.toml   # Set aria2 download dir

    if [ ! -d "${ARIA2_DOWNLOAD_DIR}" ]; then
        mkdir -p ${ARIA2_DOWNLOAD_DIR}
        echo "${ARIA2_DOWNLOAD_DIR} created"
    else
        echo "${ARIA2_DOWNLOAD_DIR} exists"
    fi

else
   cp ./config.toml.example ${CONF_FILE_DIR}/config.toml
   echo "${CONF_FILE_DIR}/config.toml created"
   ARIA2_DOWNLOAD_DIR=${CONF_FILE_DIR}/download
   sed -i 's@%%ARIA2_DOWNLOAD_DIR%%@'${ARIA2_DOWNLOAD_DIR}'@g' ${CONF_FILE_DIR}/config.toml   # Set aria2 download dir

   if [ ! -d "${ARIA2_DOWNLOAD_DIR}" ]; then
       mkdir -p ${ARIA2_DOWNLOAD_DIR}
       echo "${ARIA2_DOWNLOAD_DIR} created"
   else
       echo "${ARIA2_DOWNLOAD_DIR} exists"
   fi
fi

sed -i 's/%%USER%%/'${USER}'/g' ./aria2c.service   # Set User & Group to value of $USER

if [ ! -d "/etc/aria2" ]; then
    sudo mkdir /etc/aria2
    echo "/etc/aria2 created"
else
    echo "/etc/aria2 exists"
fi

sudo chown $USER:$USER /etc/aria2/             # Change user authority to current user

ARIA2_SESSION=/etc/aria2/aria2.session
if [ -f "${ARIA2_SESSION}" ]; then
    echo ${ARIA2_SESSION} exists
else
    sudo touch ${ARIA2_SESSION}            # Create a session file
fi

ARIA2_CONF=/etc/aria2/aria2.conf
if [ -f "${ARIA2_CONF}" ]; then
    echo ${ARIA2_CONF} exists
else
    sudo cp ./aria2.conf /etc/aria2/               # Copy config file
fi

ARIA2_SERVICE=/etc/systemd/system/aria2c.service
if [ -f "${ARIA2_SERVICE}" ]; then
    echo ${ARIA2_SERVICE} exists
else
    sudo cp ./aria2c.service /etc/systemd/system/    # Copy service file
fi

sudo systemctl enable aria2c.service           # Set to start Aria2 automatically
sudo systemctl restart aria2c.service            # Start Aria2
