# Use an official lightweight Linux distribution as a parent image
FROM debian:latest

# Update and install necessary packages
RUN apt-get update && \
    apt-get install -y tor curl netcat-openbsd

# Add the torrc configuration file
ADD config/torrc /etc/tor/torrc

# Add the verification script
ADD config/verify_tor.sh /usr/local/bin/verify_tor.sh
RUN chmod +x /usr/local/bin/verify_tor.sh

# Expose the necessary ports for Tor
EXPOSE 9050 9051

# Run Tor and then the verification script
CMD ["sh", "-c", "tor & sleep 10 && /usr/local/bin/verify_tor.sh"]