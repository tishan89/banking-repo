# Banking APIs Repository

This repository is a collection of banking-related APIs, designed to support a variety of financial services and use cases. It is structured to allow for modular development and deployment of different banking domains, such as investment banking, lending, and more.

## Project Overview

The goal of this repository is to provide a unified platform for developing, testing, and deploying banking APIs. Each API is organized in its own backend directory, with clear documentation and OpenAPI specifications for easy integration and maintenance.

## APIs in this Repository

### 1. Investment Banking API
- **Directory:** `investment-backend/`
- **Description:** Provides CRUD operations for investment banking clients and their savings accounts. Supports filtering clients by investment value and is designed for easy extension and integration.
- **OpenAPI Spec:** See `investment-backend/openapi.yaml`

### 2. Lending Data API (Planned)
- **Directory:** `lending-backend/` (to be added)
- **Description:** Will provide endpoints for managing lending products, loan applications, repayments, and customer credit data.

### 3. Additional APIs
- More banking APIs (e.g., payments, customer onboarding, compliance) will be added as the project evolves.

## Getting Started

### Prerequisites
- Go (for backend services)
- Git

### Running the Investment Banking API
1. Navigate to the backend directory:
   ```sh
   cd investment-backend
   ```
2. Run the server:
   ```sh
   go run main.go
   ```
3. The API will be available at `http://localhost:8080`.

### Running Tests
```sh
cd investment-backend
go test
```

## API Documentation
Each API includes an OpenAPI (Swagger) definition for easy reference and integration. See the respective backend directory for details.

## Contributing
- Fork the repo and create a feature branch for your API or enhancement.
- Follow the structure of existing APIs for consistency.
- Add or update OpenAPI specs and tests for your changes.
- Submit a pull request with a clear description.

## Future Plans
- Add more banking APIs (lending, payments, etc.)
- Integrate with CI/CD pipelines
- Add Docker and Choreo deployment guides
- Enhance security and authentication across APIs

---

For questions or suggestions, please open an issue or contact the maintainers.