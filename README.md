# pmtaproxy

Forward proxy that supports the [proxy protocol version 1](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt).

See [PowerMTA 5.0: Using a proxy for email delivery](https://www.postmastery.com/powermta-5-0-using-a-proxy-for-email-delivery/) for possible use cases.

## Requirements

The proxy is a lean, native binary written in Go. It does mostly network I/O. A (virtual) Linux server with 1 CPU core and 1 GB RAM is sufficient to run the proxy.

Any up-to-date Linux distribution can be used. CentOS/RHEL 5.x is not supported.

PowerMTA 5.0r2 or higher should be used. PowerMTA 5.0r1 has a bug in the proxy support implementation.

## Features

The main feature of pmtaproxy is it's simplicity. It's sole purpose is to proxy connections from PowerMTA to the destination MX. PowerMTA will tell the proxy from which IP, and to which IP a connection should be made. Thus little configuration of the proxy is needed.

The proxy supports the following settings:

* -l: host:port for listening. Use this setting to specify the IP (optional) and port to listen for incoming connections fromm PowerMTA.
* -a: allowed connection sources. Use this setting to allow incoming connections from the specified CIDR range and deny all other.

## Installing

Compile pmtaproxy from source or download a precompiled binary from [Releases](https://github.com/postmastery/pmtaproxy/releases).

Copy pmtaproxy to the server in /usr/local/sbin. Set permissions to 0755.

Create systemd configuration as /lib/systemd/system/pmtaproxy.service:

    [Unit]
    Description=Systemd configuration for pmtaproxy

    [Service]
    ExecStart=/usr/local/sbin/pmtaproxy -l=:5000 -a=10.0.0.0/8
    Restart=on-failure
    LimitNOFILE=4096

    [Install]
    WantedBy=multi-user.target

Set permissions of /lib/systemd/system/pmtaproxy.service to 0644.

If a firewall is used, allow incoming connections to the port used by the proxy. 

On Ubuntu/Debian:

    ufw allow 5000/tcp

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

### PowerMTA 5.0r8 or later

A \<virtual-mta\> can be configured with smtp-proxy-source-host instead of smtp-source-host which routes all connections via the specified proxy and the specified IP on the proxy. For example:

    <virtual-mta prox01>
      # smtp-proxy-source-host LOCAL-IP PROXY-SERVER:PORT PROXY-CLIENT-IP[:PORT] HOSTNAME
      smtp-proxy-source-host 140.201.38.1 65.30.42.1:5000 65.30.42.1 prox01.example.com
    </virtual-mta>

NOTE: The hostname must refer to the proxy client IP, not the local IP.

### PowerMTA 5.0r2 to 5.0r7

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

### Testing connection

    nc -vz proxy_server_ip 5000

    echo -n "PROXY TCP4 proxy_source_ip destination_ip 0 25" | nc -s local_source_ip proxy_server_ip 5000

### Too many open files

If the log shows "too many open files" errors, check the limits for the pmtaproxy process:

    cat /proc/`pidof pmtaproxy`/limits

Increase the LimitNOFILE setting in /lib/systemd/system/pmtaproxy.service:

    [Service]
    ...
    LimitNOFILE=8092

Reload services configuration and restart pmtaproxy:

    systemctl daemon-reload
    systemctl restart pmtaproxy


