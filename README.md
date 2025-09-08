# Expense Management System

A backend system for managing employee expenses with approval workflows and payment processing.

## Features

- User authentication with JWT
- Expense submission with validation
- Manager approval workflow
- Auto-approval for small expenses
- Payment processing with idempotency
- Role-based access control
- Rate limiting and CORS support

## Setup

1. Clone the repository
2. Copy `.env.example` to `.env` and configure your environment variables
3. Run `docker-compose up -d` to start the database

## API Documentation

### Authentication

- `POST /api/auth/login` - Login with email and password

### Expenses

- `POST /api/expenses` - Create a new expense
- `GET /api/expenses` - List user's expenses
- `GET /api/expenses/{id}` - Get expense details
- `PUT /api/expenses/{id}/approve` - Approve expense (managers only)
- `PUT /api/expenses/{id}/reject` - Reject expense (managers only)
- `GET /api/expenses/pending` - Get pending approvals (managers only)

### Health

- `GET /api/health` - Health check endpoint

## Default Users

- Manager: `manager@example.com` / `password`
- Employee: `employee@example.com` / `password`

## Business Rules

- Minimum expense amount: IDR 10,000
- Maximum expense amount: IDR 50,000,000
- Approval threshold: IDR 1,000,000
- Expenses below threshold are auto-approved
- Expenses above threshold require manager approval
