#[sudo] ip tuntap add user <username> mode tun <device-name>
#[sudo] ip link set <device-name> up
#[sudo] ip addr add <ipv4-address>/<mask-length> dev <device-name>

sudo ip tuntap add user gusa1120 mode tun gusa
sudo ip link set gusa up
sudo ip addr add 10.0.0.0/8 dev gusa
