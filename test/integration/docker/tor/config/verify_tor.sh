#!/bin/bash

# Wait for Tor to start up
sleep 10

# Check exit IP using Tor
echo "Checking exit IP using Tor..."
curl --socks5-hostname localhost:9050 https://httpbin.org/ip

# Connect to the control port and authenticate
(
    echo 'authenticate "password"'
    echo 'getinfo status/circuit-established'
) | nc localhost 9051