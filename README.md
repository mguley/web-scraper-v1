# Web Scraper Project
This project is a web scraping application designed to gather and manage data from various sources while maintaining anonymity through the use of the Tor network. The project is built with Go and utilizes Docker for containerization.

## Features
- **Tor Integration**: Ensures anonymity by routing web requests through the Tor network.
- **In-Memory Caching**: Provides efficient data retrieval and storage using an in-memory cache.
- **MongoDB Repository**: Utilizes MongoDB for persistent data storage.
- **Configuration Management**: Manages application configurations via environment files.
- **Dockerized Environment**: Containerizes the application and its dependencies for consistent and reproducible deployments.
- **Integration Tests**: Tests to ensure the functionality and reliability of the system.
- **RabbitMQ Integration**: Manages message queues for efficient task processing.
- **Kubernetes Support**: Deployment configurations for Kubernetes clusters.

## Recent Changes
- **Expanded Integration Tests**: Added more comprehensive integration tests for different components, including message consumers and publishers.
- **Enhanced Kubernetes Deployments**: Added YAML configurations for deploying MongoDB, RabbitMQ, and the Tor network within a Kubernetes cluster.
- **Improved Configuration Management**: Enhanced the management of environment variables and configuration files.
- **Refined Tor Integration**: Improved the implementation of Tor-related commands and the client facade.

## Project Structure

The project is organized into several key directories:

- **config/**: Contains configuration files and scripts.
- **docker/**: Docker-related files and configurations, including Dockerfiles and Docker Compose files.
- **internal/cache/**: Implementation of the in-memory cache.
- **internal/crawler/**: Contains the dispatcher and worker logic for the web scraper.
- **internal/message/**: Message queue management, including consumers and publishers for RabbitMQ.
- **internal/model/**: Definition of data models used within the application.
- **internal/repository/**: MongoDB repository implementation.
- **internal/tor/**: Tor client and facade implementation, along with related commands and builders.
- **internal/utils/**: Utility functions, including logging.
- **test/integration/**: Integration tests for various components of the application.
- **.github/workflows/**: GitHub Actions workflows for CI/CD and integration tests.
