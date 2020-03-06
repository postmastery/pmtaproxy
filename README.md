# pmtaproxy

Forward proxy that supports the [proxy protocol version 1](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt).

## Installing

Download pmtaproxy for Linux from [https://postmastery.egnyte.com/dl/can2yhUBi8](https://postmastery.egnyte.com/dl/can2yhUBi8).

Copy pmtaproxy to the server in /usr/local/sbin.

Create systemd configuration as /lib/systemd/system/pmtaproxy.service:

    [Unit]
    Description=Systemd configuration for pmtaproxy

    [Service]
    ExecStart=/usr/local/sbin/pmtaproxy -l=:5000 -a=10.0.0.0/8
    Restart=on-failure

    [Install]
    WantedBy=multi-user.target

Allow incoming connections to the port used by the proxy. On Ubuntu/Debian:

    ufw allow any to any port 5000 proto tcp

On Redhat/CentOS:

    firewall-cmd --zone=public --add-port=5000/tcp --permanent
    firewall-cmd --reload

Start pmtaproxy:

    systemctl start pmtaproxy

View log:

    journalctl -f -u pmtaproxy

## Configuration

Configure proxy in PowerMTA (>= 5.0r1):

    <proxy vps2>
      server 136.144.181.209:5000  # proxy listener address
      client 136.144.181.209:0 vps2.postmastery.net  # source for connections from proxy to destination
    </proxy>

    <virtual-mta vps2>
      smtp-source-host 149.210.243.38 vps4.postmastery.net  # source for connections from powermta to proxy
      use-proxy vps2 # connect to final destination via proxy
    </virtual-mta>

NOTE: The hostname in \<proxy\>.client is used as HELO/EHLO name, instead of \<virtual-mta\>.smtp-source-host.
