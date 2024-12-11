# <h1 align="center">Go Email Consumer with AWS SQS & AWS SES</h1>
___
<p align="center">
<a href="Java url">
    <img alt="Java" src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
</a>
<a href="License url" >
        <img alt="BSD Clause 3" src="https://img.shields.io/badge/License-MIT-yellow.svg"/>
    </a>
</p>

---

### Overview
This project provides a production-ready Golang email consumer service utilizing AWS SQS (Simple Queue Service) and AWS SES (Simple Email Service). The service is designed for high availability, scalability, and blazing fast performance. It's Dockerized for easy development and deployment.


## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
   - [Environment Variables](#environment-variables)
- [Contributing](#contributing)
- [Contact](#contact)
- [License](#license)

## Technologies <a name="technologies"></a>

- [Go](https://golang.org/)
- [Docker](https://www.docker.com/)
- [AWS SES](https://aws.amazon.com/)
- [AWS SQS](https://aws.amazon.com/)

## Features

- AWS Integration: Utilizes AWS SQS for message queuing and AWS SES for sending emails.
- High Availability: Designed to maintain service availability even during high loads or failures.
- Scalability: Easily scalable to handle varying loads of email processing.
- Blazing Fast Performance: Optimized for performance to ensure quick processing of email requests.
- Dockerized: Packaged as Docker containers for seamless development and deployment.

### Prerequisites
Before running this service, ensure you have the following prerequisites installed:

- Docker: Install Docker
- AWS Account: Obtain AWS credentials and configure them locally.

## Environment Variables

Table bellow shows the obligatory environment variables for mariadb container. You should set them based on what was also set for backend container.

Environment variable  | Default value | Optional
--- | --- | ---
STAGE | dev | `NO`
SQS_URL |  | `YES`
AWS_REGION |  | `YES`
AWS_ACCESS_KEY |  | `YES`
AWS_SECRET_ACCESS_KEY |  | `YES`

## Getting started
   ```
   First of all, correctly configure the Golang development environment on your machine, see https://go.dev/doc/install
   
   - Clone this repository:
   $ git clone https://github.com/swarajkumarsingh/go-email-consumer-aws-sqs.git

   - Enter in directory:
   $ cd go-email-consumer-aws-sqs

   - For install dependencies(optional):
   $ make run SQS_URL= AWS_REGION= AWS_ACCESS_KEY= AWS_SECRET_ACCESS_KEY=
   ```
---

### Contributing
Contributions are welcome! If you'd like to contribute to this project, please follow these guidelines:
```
1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and test thoroughly.
4. Commit your changes with clear commit messages.
5. Create a pull request against the main branch.
```

## Contact

- Swaraj Singh <br> <br>
  <a  href="https://www.linkedin.com/in/swarajkumarsingh/" target="_blank"><img alt="LinkedIn" src="https://img.shields.io/badge/linkedin%20-%230077B5.svg?&style=for-the-badge&logo=linkedin&logoColor=white" /></a>
  <a href="sswaraj169@gmail.com"><img  alt="Gmail" src="https://img.shields.io/badge/Gmail-D14836?style=for-the-badge&logo=gmail&logoColor=white" />

  feel free to contact me!

## License

> You can check out the full license [here](https://github.com/swarajkumarsingh/go-email-consumer-aws-sqs/blob/main/LICENSE)

This project is licensed under the terms of the **MIT** license

### Acknowledgments
Special thanks to AWS for providing powerful cloud services and Golang community for building robust frameworks and libraries.
