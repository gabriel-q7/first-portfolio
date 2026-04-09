# my-monorepo

A monorepo containing the frontend, backend, infrastructure and automation scripts for the project.

## Structure

```
my-monorepo/
├── apps/
│   ├── frontend/      # Frontend application
│   └── backend/       # Backend application
│
├── infra/             # Docker, Terraform, etc.
├── scripts/           # Automation scripts
├── docs/              # Project documentation
│
├── docker-compose.yml
├── Makefile
└── README.md
```

## Getting Started

### Prerequisites

- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/)
- [Make](https://www.gnu.org/software/make/)

### Running the project

```bash
# Build all services
make build

# Start all services in detached mode
make up

# View logs
make logs

# Stop all services
make down
```