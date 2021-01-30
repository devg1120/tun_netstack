#[sudo] ip tuntap add user <username> mode tun <device-name>
#[sudo] ip link set <device-name> up
#[sudo] ip addr add <ipv4-address>/<mask-length> dev <device-name>

#sudo ip link set gusa down
#sudo ip addr del 10.0.0.0/8 dev gusa
sudo ip tuntap del  mode tun gusa
