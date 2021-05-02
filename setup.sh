sudo apt-get update
sudo apt-get -y upgrade
wget https://dl.google.com/go/go1.16.3.linux-amd64.tar.gz
sudo tar -xvf go1.16.3.linux-amd64.tar.gz -C /usr/local
export PATH=$PATH:/usr/local/go/bin
go version
wget https://dist.ipfs.io/go-ipfs/v0.8.0/go-ipfs_v0.8.0_linux-amd64.tar.gz
tar -xvf go-ipfs_v0.8.0_linux-amd64.tar.gz
sudo go-ipfs/install.sh
rm -rf go-ipfs
rm go1.16.3.linux-amd64.tar.gz
rm go-ipfs_v0.8.0_linux-amd64.tar.gz

export HOME=/home/ubuntu
ipfs init -p server
echo '/key/swarm/psk/1.0.0/
/base16/
cd1bb5ad82071f9cddda00ab0999e5e7f0bfc96d8e62ea69ffe9e8ecf6197cf3' >> "$HOME"/.ipfs/swarm.key
export LIBP2P_FORCE_PNET=1
ipfs bootstrap add /ip4/52.4.216.226/tcp/4001/p2p/12D3KooWGgdQXmEhvsCwcFLrNuJSTbV3k5EYKnksyNzcvCx8Ld2v
sudo bash -c 'cat >/lib/systemd/system/ipfs.service <<EOL
[Unit]
Description=ipfs daemon
[Service]
ExecStart=/usr/local/bin/ipfs daemon --enable-gc
Restart=always
User=ubuntu
Group=ubuntu
[Install]
WantedBy=multi-user.target
EOL'
sudo systemctl daemon-reload
sudo systemctl enable ipfs.service
sudo systemctl start ipfs
sudo systemctl status ipfs
sudo chown -R ubuntu $HOME/.ipfs