# SITCOMCHAIN
Blockchain's part of a SIT-Competence using tendermint.

## Prerequisites
0. Install required dependencies
    ```bash
    sudo apt-get update
    sudo apt install build-essential unzip
    ```

1. Install Go and Setup GOPATH
    ```bash
    ### For Ubuntu and Debian distros
    # Install Go
    wget https://dl.google.com/go/go1.13.5.linux-amd64.tar.gz
    mkdir -p /usr/local
    tar xvf go1.13.5.linux-amd64.tar.gz
    sudo mv go /usr/local
    
    # Setup Go environment variables
    mkdir -p ~/gofolder
    echo 'export GOROOT=/usr/local/go' >> ~/.bashrc
    echo 'export GOPATH=$HOME/gofolder' >> ~/.bashrc
    echo 'export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin' >> ~/.bashrc
    source ~/.bashrc
    ```
 2. Install Tendermint
    ```bash
    ### For Ubuntu and Debian distros
    wget https://github.com/tendermint/tendermint/releases/download/v0.32.7/tendermint_v0.32.7_linux_amd64.zip
    unzip tendermint_v0.32.7_linux_amd64.zip -d ~/tendermint
    echo 'export PATH=$PATH:$HOME/tendermint' >> ~/.bashrc
    ```

## Usage
1. Generate validator key, node key and genesis file
    ```bash
    tendermint init
    
    ```
2. For more than 1 node, set nodes id and their corresponding ip address and port to persistent_peers variable in **~/.tendermint/config/config.toml** in format => **persistent_peers = "{NODEID}@{IP}:{Port}"**
    ```bash
    # get node's id
    tendermint show_node_id
    
    # example in ~/.tendermint/config/config.toml
    persistent_peers = "5a3b1b228d558235d5a8c76c28ecef13e6ad55f2@10.4.56.17:26656,31c219dd725aa371052c2d9b8c1f12de13ed4591@10.4.56.22:26656,8369dfd9f8cedf85db929186fade7054175a4cf1@10.4.56.23:26656"
    ```
3. You could set **create_empty_blocks = false** in **config.toml** to prevent unnecessary producing block.
4. Run tendermint node
    ```bash
    tendermint node
    ```
5. Open another tab to run a SITCOMCHAIN smart-contract
    ```bash
    mkdir -p $GOPATH/src/github.com/saguywalker
    cd $GOPATH/src/github.com/saguywalker
    git clone https://github.com/saguywalker/sitcomchain
    cd sitcomchain
    go build
    ./sitcomchain
    ```
