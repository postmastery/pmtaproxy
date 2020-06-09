# pmtaproxy

Forward proxy that supports the [proxy protocol version 1](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt).

See [PowerMTA 5.0: Using a proxy for email delivery](https://www.postmastery.com/powermta-5-0-using-a-proxy-for-email-delivery/) for possible use cases.

## Requirements

The proxy is a lean, native binary written in Go. It does mostly network I/O. A (virtual) Linux server with 1 CPU core and 1 GB RAM is sufficient to run the proxy.

## Features

The main feature of pmtaproxy is it's simplicity. It's sole purpose is to proxy connections from PowerMTA to the destination MX. PowerMTA will tell the proxy from which IP, and to which IP a connection should be made. Thus little configuration of the proxy is needed.

The proxy supports the following settings:

* -l: host:port for listening. Use this setting to specify the IPs and port to listen for incoming connections fromm PowerMTA.
* -a: allowed connection sources. Use this setting to allow incoming connections from the specified CIDR range and deny all other.

## Installing

Download pmtaproxy for Linux from [https://postmastery.egnyte.com/dl/can2yhUBi8](https://postmastery.egnyte.com/dl/can2yhUBi8).

Copy pmtaproxy to the server in /usr/local/sbin. Set permissions to 0755.

Create systemd configuration as /lib/systemd/system/pmtaproxy.service:

    [Unit]
    Description=Systemd configuration for pmtaproxy

    [Service]
    ExecStart=/usr/local/sbin/pmtaproxy -l=:5000 -a=10.0.0.0/8
    Restart=on-failure

    [Install]
    WantedBy=multi-user.target

Set permissions of /lib/systemd/system/pmtaproxy.service to 0644.

Allow incoming connections to the port used by the proxy. On Ubuntu/Debian:

    ufw allow any to any port 5000 proto tcp

On Redhat/CentOS:

    firewall-cmd --zone=public --add-port=5000/tcp --permanent
    firewall-cmd --reload

Start pmtaproxy:

    systemctl start pmtaproxy

Monitor log:

    journalctl -f -u pmtaproxy

Enable start on system startup:

    sudo systemctl enable pmtaproxy

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

## Troubleshooting

### Too many open files

If the log shows "too many open files" errors, check the limits for the pmtaproxy process:

    cat /proc/`pidof pmtaproxy`/limits

Add LimitNOFILE setting to /lib/systemd/system/pmtaproxy.service:

    [Service]
    ...
    LimitNOFILE=4096

Reload services configuration and restart pmtaproxy:

    systemctl daemon-reload
    systemctl restart pmtaproxy


