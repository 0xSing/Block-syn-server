#!/bin/bash
SERVER_NAME="bsc_main"
CHAIN_ID="56"

cd /data/walletSyn_$CHAIN_ID/walletsynv2/
git pull ssh://git@gitlab.thyy.pro:5519/blockchain/walletsynv2.git dev
make walletSyn

while read -r pid; do
  if [[ -n "$pid" ]]; then
    kill -SIGINT "$pid"
    echo "PID $pid is shutdown"
  fi
done <<< $(ps -ef | grep ./build/bin/$SERVER_NAME | grep -v grep | awk '{print $2}')

sleep 3

ps -ef | grep ./build/bin/$SERVER_NAME | grep -v grep | awk '{print $2}' | xargs kill -9

screen -dmS syn_$CHAIN_ID bash -c "cd /data/walletSyn_$CHAIN_ID/walletsynv2/ && ./build/bin/$SERVER_NAME; exec bash"

screen -wipe