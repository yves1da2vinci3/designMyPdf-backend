#!/bin/bash

# Set the file ID and destination path from environment variables
FILE_ID="$FILE_ID"
DESTINATION="$DESTINATION"

# Download the file from Google Drive
wget --load-cookies /tmp/cookies.txt "https://docs.google.com/uc?export=download&confirm=$(wget --quiet --save-cookies /tmp/cookies.txt --keep-session-cookies --no-check-certificate 'https://docs.google.com/uc?export=download&id='$FILE_ID -O- | sed -rn 's/.*confirm=([0-9A-Za-z_]+).*/\1\n/p')&id=$FILE_ID" -O "$DESTINATION" && rm -rf /tmp/cookies.txt

# Install Chromium
if [ -x "$(command -v apt-get)" ]; then
  # Ubuntu and Debian
  sudo apt-get update
  sudo apt-get install -y chromium-browser
elif [ -x "$(command -v yum)" ]; then
  # CentOS and Fedora
  sudo yum install -y chromium
elif [ -x "$(command -v pacman)" ]; then
  # Arch Linux
  sudo pacman -S --noconfirm chromium
elif [ -x "$(command -v zypper)" ]; then
  # OpenSUSE
  sudo zypper install -y chromium
else
  echo "Unsupported package manager. Please install Chromium manually."
  exit 1
fi

# Check if Chromium was installed successfully
if [ -x "$(command -v chromium-browser)" ]; then
  echo "Chromium installed successfully."
else
  echo "Failed to install Chromium."
  exit 1
fi
