#!/bin/bash
TARGET_USER=pi
TARGET_PWD=vespas
TARGET_HOST=192.168.2.224
TARGET_DIR=dev
ARM_VERSION=6
 
# Executable name is assumed to be same as current directory name
EXECUTABLE=shaz
 
echo "Building for Raspberry Pi..."
env CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=6 CC="/home/ub/arm-cross-comp-env/arm-raspbian-linux-gnueabihf/bin/arm-linux-gnueabihf-gcc" CGO_LDFLAGS="-latomic" go build -o shaz ./src/
#  arm-linux-gnueabi-gcc
#  arm-linux-gnueabihf-gcc
# echo "Uploading to Raspberry Pi..."


PIDS=$(ssh $TARGET_USER@$TARGET_HOST "pgrep shaz")

if [ -n "$PIDS" ]; then
  # The program is running, so stop each PID
  for PID in $PIDS; do
    ssh $TARGET_USER@$TARGET_HOST "sudo kill $PID"
    echo "Stopped shaz (PID: $PID) on remote_server"
  done
else
  echo "shaz is not running on remote_server"
fi

scp -i /home/ub/.ssh/id_rsa.pub $EXECUTABLE $TARGET_USER@$TARGET_HOST:$TARGET_DIR/$EXECUTABLE
scp -i /home/ub/.ssh/id_rsa.pub static/* $TARGET_USER@$TARGET_HOST:$TARGET_DIR/static/
scp -i /home/ub/.ssh/id_rsa.pub creds.toml $TARGET_USER@$TARGET_HOST:$TARGET_DIR/creds.toml
scp -i /home/ub/.ssh/id_rsa.pub launcher.sh $TARGET_USER@$TARGET_HOST:$TARGET_DIR/launcher.sh
# scp -i /home/ub/.ssh/id_rsa.pub -r /home/ub/Documents/Coding/ShazPi/static $TARGET_USER@$TARGET_HOST:$TARGET_DIR/static
ssh $TARGET_USER@$TARGET_HOST "sudo reboot"


# CC=arm-none-linux-gnueabi-gcc CXX=arm-none-linux-gnueabi-g++ ./configure --target=arm-none-linux-gnueabi --host=arm-none-linux-gnueabi
# CFLAGS="-march=armv7-a -mtune=cortex-a8 -mfpu=neon -mfloat-abi=softfp" ./configure --host=arm-linux-gnueabihf --prefix=/usr/local/angstrom/arm --disable-shared 

# https://medium.com/@stonepreston/how-to-cross-compile-a-cmake-c-application-for-the-raspberry-pi-4-on-ubuntu-20-04-bac6735d36df
# $ uname -r
# 6.1.21+

# $ ld --version
# GNU ld (GNU Binutils for Raspbian) 2.35.2
# Copyright (C) 2020 Free Software Foundation, Inc.
# This program is free software; you may redistribute it under the terms of
# the GNU General Public License version 3 or (at your option) a later version.
# This program has absolutely no warranty.


# $ gcc --version
# gcc (Raspbian 10.2.1-6+rpi1) 10.2.1 20210110
# Copyright (C) 2020 Free Software Foundation, Inc.
# This is free software; see the source for copying conditions.  There is NO
# warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

# $ ldd --version
# ldd (Debian GLIBC 2.31-13+rpt2+rpi1+deb11u5) 2.31
# Copyright (C) 2020 Free Software Foundation, Inc.
# This is free software; see the source for copying conditions.  There is NO
# warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
# Written by Roland McGrath and Ulrich Drepper.

# $ rsync -rzLR --safe-links pi@192.168.2.227:/lib/ ./
# rsync -rzLR --safe-links pi@192.168.2.227:/usr/include/ ./
# rsync -rzLR --safe-links pi@192.168.2.227:/usr/lib/ ./
