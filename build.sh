#!/bin/bash

TARGET_USER=pi
TARGET_HOST=192.168.2.224
TARGET_DIR=dev
EXECUTABLE=ShazPi
CC_PATH="/home/ub/arm-cross-comp-env/arm-raspbian-linux-gnueabihf/bin/arm-linux-gnueabihf-gcc"

# ASCII art banner
echo "
   _____ _               _____ _ 
  / ____| |             |  __ (_)
 | (___ | |__   __ _ ___| |__) | 
  \___ \| '_ \ / _\` |_  /  ___/ |
  ____) | | | | (_| |/ /| |   | |
 |_____/|_| |_|\\__,_/___|_|   |_|
                                                             
"

# Function to exit the script if a command fails
check_exit_status() {
    if [ $? -ne 0 ]; then
        echo "Error: Command failed. Exiting..."
        exit 1
    fi
}

echo "Building for Raspberry Pi..."
env CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=6 CC="$CC_PATH" CGO_LDFLAGS="-latomic" go build -o ShazPi ./src/
check_exit_status

echo "Stopping ShazPi if running..."
if ssh $TARGET_USER@$TARGET_HOST "pgrep ShazPi"; then
  # The program is running, so stop it
  ssh $TARGET_USER@$TARGET_HOST "sudo pkill ShazPi"
  check_exit_status
  echo "Stopped ShazPi on remote server"
else
  echo "ShazPi is not running on remote server"
fi

echo "Uploading to Raspberry Pi..."
# ssh $TARGET_USER@$TARGET_HOST "[ ! -f $TARGET_DIR/$EXECUTABLE ]"
ssh $TARGET_USER@$TARGET_HOST "mkdir -p $TARGET_DIR/static"
check_exit_status
ssh $TARGET_USER@$TARGET_HOST "mkdir -p $TARGET_DIR/temp"
check_exit_status
scp $EXECUTABLE $TARGET_USER@$TARGET_HOST:$TARGET_DIR/$EXECUTABLE
check_exit_status
scp static/* $TARGET_USER@$TARGET_HOST:$TARGET_DIR/static/
check_exit_status
scp creds.toml $TARGET_USER@$TARGET_HOST:$TARGET_DIR/creds.toml
check_exit_status
scp launcher.sh $TARGET_USER@$TARGET_HOST:$TARGET_DIR/launcher.sh
check_exit_status
# else
#   echo "Skipping uploading as ShazPi executable already exists on remote server"
# fi

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
                exit 1
                ;;
        esac
    done
}

if ask_question "Do you want to reboot the Pi?"; then
  echo "Rebooting the Pi..."
  ssh $TARGET_USER@$TARGET_HOST "sudo reboot"
  check_exit_status
fi
echo "Done"


