# Meal Planner

A full-stack meal planning application that generates personalized weekly meal plans using the Spoonacular API. Built with Go and React.

## Features

- User registration and JWT-based authentication
- Automatic weekly meal plan generation based on dietary preferences
- Swap out individual recipes you don't like
- Regenerate your entire meal plan with one click
- Dietary preference support (vegetarian, vegan, gluten free, and more)
- User profile management

## Tech Stack

**Backend**
- [Go](https://go.dev/) — REST API
- [Chi](https://github.com/go-chi/chi) — HTTP router
- [PostgreSQL](https://www.postgresql.org/) — Database
- [Goose](https://github.com/pressly/goose) — Database migrations
- [JWT](https://github.com/golang-jwt/jwt) — Authentication
- [Spoonacular API](https://spoonacular.com/food-api) — Recipe data

**Frontend**
- [React](https://react.dev/) + [TypeScript](https://www.typescriptlang.org/) — UI
- [Vite](https://vitejs.dev/) — Build tool
- [Tailwind CSS](https://tailwindcss.com/) — Styling
- [Axios](https://axios-http.com/) — HTTP client
- [React Router](https://reactrouter.com/) — Navigation
- [React Hot Toast](https://react-hot-toast.com/) — Notifications

## Prerequisites

- [Go](https://go.dev/dl/) 1.21+
- [Node.js](https://nodejs.org/) 18+
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Goose](https://github.com/pressly/goose) — `go install github.com/pressly/goose/v3/cmd/goose@latest`
- A [Spoonacular API key](https://spoonacular.com/food-api)

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/jonnarhei/meal-planner.git
cd meal-planner
```

### 2. Set up environment variables

Create a `.env` file in the `backend/` folder:

```dotenv
ADDR=:8080
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_MAX_IDLE_TIME=15m
JWT_SECRET=your-long-random-secret-here
JWT_EXPIRY=86400
SPOONACULAR_API_KEY=your-spoonacular-api-key
GOOSE_DRIVER=postgres
GOOSE_DBSTRING="postgres://user:adminpassword@localhost:5432/mealplanner?sslmode=disable"
GOOSE_MIGRATION_DIR="internal/database/migrations"
```

Generate a secure JWT secret:
```bash
openssl rand -base64 32
```
Insert this into the JWT_SECRET variable in the .env file

### 3. Start the database

```bash
docker compose up -d
```

### 4. Run migrations

```bash
cd backend
goose up
```

### 5. Start the backend

```bash
go run cmd/api/*.go
```

### 6. Start the frontend

```bash
cd ..
cd frontend
npm install
npm run dev
```

The app will be available at `http://localhost:5173` by default. It may run on a different port if 5173 is occupied


## Project Structure

```
meal-planner/
├── backend/
│   ├── cmd/
│   │   └── api/          # Handlers, middleware, routing
│   ├── internal/
│   │   ├── auth/         # JWT generation and validation
│   │   ├── database/     # DB connection and migrations
│   │   ├── env/          # Environment variable helpers
│   │   ├── jsonutil/     # JSON response helpers
│   │   ├── spoonacular/  # Spoonacular API client
│   │   └── store/        # Database layer and models
│   ├── go.mod
│   └── go.sum
├── frontend/
│   └── src/
│       ├── api/          # Axios API calls
│       ├── components/   # Reusable components
│       ├── context/      # Auth context
│       └── pages/        # Page components
├── docker-compose.yml
└── README.md
```

