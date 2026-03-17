#!/bin/bash
export SSHPASS='If%2zvElra'
sshpass -e ssh -o StrictHostKeyChecking=no root@91.218.113.247 'mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGF+ix7XmKOLG3wSP6TDvkdfWWLb1FXzkuo6CT3jCPDI codex-mudro-vps2" >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo "KEY ADDED SUCCESSFULLY"'
