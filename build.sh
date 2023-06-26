#!/bin/bash
set -e

TARGET_USER=pi
TARGET_PWD=raspberry
TARGET_HOST=192.168.2.224
TARGET_DIR=dev
ARM_VERSION=6

# Executable name is assumed to be same as current directory name
EXECUTABLE=ShazPi

echo "Building for Raspberry Pi..."
env CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=6 CC="/home/ub/arm-cross-comp-env/arm-raspbian-linux-gnueabihf/bin/arm-linux-gnueabihf-gcc" CGO_LDFLAGS="-latomic" go build -o ShazPi ./src/

echo "Stopping ShazPi if running..."
PIDS=$(ssh $TARGET_USER@$TARGET_HOST "pgrep ShazPi")

if [ -n "$PIDS" ]; then
  # The program is running, so stop each PID
  for PID in $PIDS; do
    ssh $TARGET_USER@$TARGET_HOST "sudo kill $PID"
    echo "Stopped ShazPi (PID: $PID) on remote_server"
  done
else
  echo "ShazPi is not running on remote_server"
fi

echo "Uploading to Raspberry Pi..."
ssh $TARGET_USER@$TARGET_HOST "mkdir -p /home/pi/dev/static"
ssh $TARGET_USER@$TARGET_HOST "mkdir -p /home/pi/dev/temp"
scp -i /home/ub/.ssh/id_rsa.pub $EXECUTABLE $TARGET_USER@$TARGET_HOST:$TARGET_DIR/$EXECUTABLE
scp -i /home/ub/.ssh/id_rsa.pub static/* $TARGET_USER@$TARGET_HOST:$TARGET_DIR/static/
scp -i /home/ub/.ssh/id_rsa.pub creds.toml $TARGET_USER@$TARGET_HOST:$TARGET_DIR/creds.toml
scp -i /home/ub/.ssh/id_rsa.pub launcher.sh $TARGET_USER@$TARGET_HOST:$TARGET_DIR/launcher.sh

# Function to ask the yes or no question
ask_question() {
    local question="$1"
    local response=""

    while true; do
        read -p "$question (y/n): " response
        case $response in
            [Yy]*)
                return 0
                ;;
            [Nn]*)
                return 1
                ;;
            *)
                echo "Invalid response. Please answer with 'y' or 'n'."
                ;;
        esac
    done
}

if ask_question "Do you want to reboot the Pi?"; then
  echo "Rebooting the Pi..."
  ssh $TARGET_USER@$TARGET_HOST "sudo reboot"
fi
echo "Done"
