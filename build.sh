#!/bin/bash
TARGET_USER=pi
TARGET_PWD=vespas
TARGET_HOST=192.168.2.227
TARGET_DIR=dev
ARM_VERSION=6
 
# Executable name is assumed to be same as current directory name
EXECUTABLE=shaz
 
echo "Building for Raspberry Pi..."
env CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=6 CC="/opt/cross-pi-gcc/bin/arm-linux-gnueabihf-gcc" go build -o shaz ./src/
#  arm-linux-gnueabi-gcc
#  arm-linux-gnueabihf-gcc
echo "Uploading to Raspberry Pi..."
scp -i /home/ub/.ssh/id_rsa.pub $EXECUTABLE $TARGET_USER@$TARGET_HOST:$TARGET_DIR/$EXECUTABLE
scp -i /home/ub/.ssh/id_rsa.pub wifi_connected.png $TARGET_USER@$TARGET_HOST:$TARGET_DIR/wifi_connected.png
scp -i /home/ub/.ssh/id_rsa.pub wifi_unconnected.png $TARGET_USER@$TARGET_HOST:$TARGET_DIR/wifi_unconnected.png
scp -i /home/ub/.ssh/id_rsa.pub 8-BIT_WONDER.TTF $TARGET_USER@$TARGET_HOST:$TARGET_DIR/8-BIT_WONDER.TTF



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
