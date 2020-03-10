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

Monitor log:

    journalctl -f -u pmtaproxy

## Configuration

Each IP address used on the proxy needs to be configured in PowerMTA using the \<proxy\> tag. This tag specifies the listener address and the source address for connections to the destination. For example:

    <proxy prox01>
      server 65.30.42.1:5000  # proxy listener address
      client 65.30.42.1:0 prox01.example.com  # source for connections from proxy to destination
    </proxy>

Then a \<virtual-mta\> can be configured with the use-proxy directive which routes all connections via the specified proxy. For example:

    <virtual-mta prox01>
      smtp-source-host 140.201.38.1 pmta01.example.com  # source for connections from powermta to proxy
      use-proxy prox01 # connect to final destination via proxy
    </virtual-mta>

NOTE: The hostname in \<proxy\>.client is used as HELO/EHLO name, instead of \<virtual-mta\>.smtp-source-host.
