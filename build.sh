#!/bin/bash
TARGET_USER=pi
TARGET_PWD=vespas
TARGET_HOST=192.8.2.227
TARGET_DIR=dev
ARM_VERSION=6
 
# Executable name is assumed to be same as current directory name
EXECUTABLE=${PWD##*/} 
 
echo "Building for Raspberry Pi..."
env GOOS=linux GOARCH=arm GOARM=$ARM_VERSION go build ./src/
 
echo "Uploading to Raspberry Pi..."
scp -i /c/Users/William/.ssh/id_rsa.pub $EXECUTABLE $TARGET_USER@$TARGET_HOST:$TARGET_DIR/$EXECUTABLE