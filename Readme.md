# Turbo Deploy 🚀

Turbo Deploy is an enterprise-grade deployment platform that automates the entire process of deploying React/Vite applications. With a single Git URL, it orchestrates building, hosting, and serving your application through a sophisticated AWS infrastructure pipeline. This project is built using Go, Node.js, PostgreSQL, and several AWS services, including RDS, S3, ElasticCache, SES, SQS, ECS, and EC2.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker](https://img.shields.io/badge/Docker-Enabled-blue)](https://www.docker.com/)
[![AWS](https://img.shields.io/badge/AWS-Powered-orange)](https://aws.amazon.com/)

## Table of Contents 📑

- [Features](#features)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Project Structure](#Project-Structure)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Reference](#api-reference)
- [Infrastructure](#infrastructure)
- [Monitoring](#monitoring)
- [Security](#security)
- [Contributing](#contributing)
- [License](#license)

## Features

### Core Capabilities
- **Automated Deployments**
  - Zero-configuration deployment pipeline
  - Automatic build optimization
  - Smart caching strategies
  - Rolling updates with zero downtime

### Build System
- **Isolated Build Environment**
  - Containerized builds using AWS ECS
  - Concurrent build support
  - Custom build parameter configuration
  - Build artifact caching

### Hosting & Delivery
- **S3-Powered Static Hosting**
  - Automatic asset optimization
  - Content compression
  - Cache control headers
  - Custom domain support

### Monitoring & Logging
- **Comprehensive Observability**
  - Real-time deployment logs
  - Performance metrics
  - Error tracking
  - Resource utilization monitoring

### Security
- **Enterprise-Grade Security**
  - SSL/TLS encryption
  - Sanitization of Build Commands

## Architecture

### System Components

#### Frontend Layer (yet to be built)
- React/Vite application hosting
- Custom domain management
- SSL/TLS termination
- Static asset serving

#### Application Services

1. **Build Server (Nodejs)**
   - Handles build requests
   - Manages ECS task lifecycle
   - Coordinates with S3 for artifact storage
   - Implements build caching

2. **Reverse Proxy (Go, Node.js)**
   - Routes requests to appropriate S3 buckets
   - Handles custom domain mapping
   - Manages SSL/TLS certificates
   - Implements caching strategies

3. **Log Consumer Service (Go)**
   - Processes SQS messages
   - Aggregates build logs
   - Handles log retention
   - Provides real-time log streaming

4. **Email Service (Go)**
   - Manages notification templates
   - Handles deployment notifications
   - Processes email queues
   - Tracks delivery status

#### Infrastructure Services

1. **Database Layer (PostgreSQL)**
   - Deployment records
   - User management
   - Configuration storage
   - Audit logs

2. **Cache Layer (Redis)**
   - Session management
   - Rate limiting
   - Build cache
   - Temporary data storage

3. **Message Queue (SQS)**
   - Service communication
   - Event processing
   - Log aggregation
   - Email notifications

4. **Storage Layer (S3)**
   - Static file hosting
   - Build artifacts
   - Log archives
   - Backup storage

## Prerequisites

### System Requirements
- Docker Engine 20.10+
- Docker Compose 2.0+
- Node.js 16.x+
- Go 1.19+
- PostgreSQL 13+
- Redis 6+

### AWS Requirements
- AWS Account with admin access
- Configured AWS CLI
- Required AWS services enabled:
  - EC2
  - ECS
  - S3
  - RDS
  - ElastiCache
  - SQS
  - SES
  - IAM

### Network Requirements
- Public IP address
- Domain name
- SSL certificate


## Project Structure

The Turbo-Deploy project is built with a modular structure to ensure scalability, maintainability, and ease of development. Below is a breakdown of the key directories and their roles:

### **/controller**  
Contains the logic for handling incoming requests and orchestrating business processes.  
- **prometheus**: Custom Prometheus metrics for real-time monitoring and observability.  
- **user**: Handles user-specific operations like authentication, user profile management, and access control.  

### **/docs**  
Houses project documentation, API references, and technical guides.  

### **/errorHandler**  
Centralized error-handling mechanism used throughout the application to ensure consistency and structured error responses.  

### **/functions**  
A collection of utility modules to handle common tasks across services:  
- **general**: Generic utilities used in various modules.  
- **logger**: Custom logging utilities for structured and level-based logging.  
- **retry**: Helper functions for implementing retry logic for failed operations.  
- **tracer**: OpenTelemetry tracing utilities for distributed system observability.  
- **uploader**: Tools for handling file uploads to S3.  
- **validator**: Input validation functions to enforce data integrity.  

### **/infra**  
Infrastructure-specific modules that abstract interactions with external services:  
- **db**: PostgreSQL database connection and query abstraction.  
- **redis**: Integration with AWS ElasticCache for caching.  
- **s3**: S3 bucket management and operations for file storage.  
- **ses**: Email sending functionalities via AWS SES.  
- **sqs**: Integration with AWS SQS for event queue management.  

### **/migrations**  
Scripts for managing database schema migrations. Ensures smooth transitions when modifying the database schema.  

### **/models**  
Defines the data models and business logic for various entities:  
- **deployment**: Manages deployment processes and state.  
- **deployment_log**: Tracks deployment history and logs.  
- **error**: Standardized error model for the system.  
- **project**: Handles project-related data and logic.  
- **user**: User-specific data and operations.  

### **/routes**  
Defines API routes and their mapping to controller functions. Acts as the entry point for incoming requests.  

### **/services**  
Contains microservices, each performing specific roles:  
- **build-server**: Executes ECS tasks to clone repositories, build projects, and upload artifacts to S3.  
- **email-consumer**: Listens to SQS queues and sends email notifications using AWS SES.  
- **logs-sqs-consumer**: Processes logs/events from the SQS queues and updates deployment logs or metrics.  
- **reverse-proxy**: Dynamically serves files from S3 based on subdomains.  

---


## Installation

### Local Development Setup

1. **Clone the Repository**
```bash
git clone https://github.com/swarajkumarsingh/turbo-deploy.git
cd turbo-deploy
```

3. **Env Setup**
```bash
# edit the run.local.sh files accordingly to your config
nano ./run.local.sh
```

2. **Run .sh file**
```bash
./run.local.sh
```

3. **Run all services(consumers)**
```bash
# migrate to root/services
make start
```

## Configuration

### Environment Variables

```env
DB_HOST=http://127.0.0.1/
DB_PORT=5432
DB_USER=user
DB_PASSWORD=postgres
DB_NAME=turbo-deploy
REDIS_HOST=
REDIS_PORT=
REDIS_USER=
REDIS_PASSWORD=
SENTRY_DSN=
DD_AGENT_HOST=
S3_BUCKET=

STAGE=dev
AWS_SQS_URL=
AWS_REGION=
AWS_ACCESS_KEY=
AWS_SECRET_ACCESS_KEY=
DB_URL=
```


## Usage

### API Examples
- check .postman folder postman.json --> import into Postman Client

## Infrastructure 🏗️

### AWS Resources

1. **Compute**
   - ECS Cluster
   - EB elastic beanstalk
   - EC2 Auto Scaling Group
   - Application Load Balancer
   - ECR - Registry for code storage

2. **Storage**
   - S3 Buckets
   - RDS Instance
   - ElastiCache Cluster

3. **Networking**
   - VPC
   - Subnets
   - Security Groups

4. **Services**
   - SES Email Service
   - SQS Queue Service

### Resource Specifications

```hcl
# ECS Task Definition
resource "aws_ecs_task_definition" "build" {
  family                   = "build"
  requires_compatibilities = ["FARGATE"]
  network_mode            = "awsvpc"
  cpu                     = 1024
  memory                  = 2048
  
  container_definitions = jsonencode([
    {
      name  = "build"
      image = "${aws_ecr_repository.build.repository_url}:latest"
      # ... additional configuration
    }
  ])
}
```

## Monitoring

### Available Metrics

1. **System Metrics**
   - CPU Usage
   - Memory Usage
   - Network I/O
   - Disk Usage

2. **Application Metrics**
   - Build Success Rate
   - Build Duration
   - Cache Hit Rate
   - Error Rate

3. **Business Metrics**(yet to be built)
   - Active Deployments
   - Total Users
   - Resource Consumption
   - Cost Analysis

### Prometheus Configuration

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'turbo-deploy'
    static_configs:
      - targets: ['localhost:9090']
```

## Security

### Authentication

- JWT-based authentication
- Rate limiting
- IP whitelisting

### Command Sanitization

- npm build sanitization for security

## Contributing

### Development Process

1. Fork the repository
2. Create a feature branch
3. Implement changes
4. Add tests
5. Submit pull request

### Code Standards

- Follow Go style guide
- Use ESLint for JavaScript
- Write unit tests
- Document changes

## License

MIT License

Copyright (c) 2024 Swaraj Kumar Singh

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.


## Support

- Email: sswaraj169@gmail.com
- Medium: [Join our community](https://swarajkumarsingh.medium.com/)
- Documentation: [Read the docs](https://github.com/swarajkumarsingh/turbo-deploy/blob/main/README.md)
- Issues: [GitHub Issues](https://github.com/username/turbo-deploy/issues)

---

Made with ❤️ by Swaraj Kumar Singh