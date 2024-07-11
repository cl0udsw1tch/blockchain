#!/bin/bash



mkdir -p $HOME/.terium/.data
mkdir -p $HOME/.terium/.data/index
mkdir -p $HOME/.terium/wallets
mkdir -p $HOME/.terium/.tmp
mkdir -p $HOME/.terium/.state/mempool
mkdir -p $HOME/.terium/.state/utxoset




# Update package index
sudo apt-get update

# Install MySQL server if not installed
if ! command -v mysql >/dev/null 2>&1; then
    echo "mysql not installed, aborting"
    exit 1
    #sudo apt-get install -y mysql-server
fi

read -sp 'Enter MySQL root password: ' root_password
echo

read -p 'Enter new MySQL username: ' mysql_username

# Prompt for new MySQL user password
read -sp 'Enter new MySQL user password: ' mysql_user_password
echo

database_name=terium_blocks

# Start MySQL service
sudo service mysql start

mysql -u root -p"$root_password" -e "\
CREATE USER '${mysql_username}'@'localhost' IDENTIFIED BY '${mysql_user_password}';\
CREATE DATABASE ${database_name};\
GRANT PRIVILEGES ON ${database_name}.* TO ${mysql_username};\
"

mysql -u ${mysql_username} -p${mysql_user_password} -e "\
source $(pwd)/internal/store/blocks.sql;\
source $(pwd)/internal/blockchain/mempool/mempool.sql;\
source $(pwd)/internal/blockchain/utxo/utxo.sql;\
"