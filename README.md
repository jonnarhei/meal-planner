# Meal Planner

A full-stack meal planning application that generates personalized weekly meal plans using the Spoonacular API. Built with Go and React.

**Live app:** [meal-planner-mu-gilt.vercel.app](https://meal-planner-mu-gilt.vercel.app)

> **Note:** The backend is hosted on Railway's free tier and may take a few seconds to wake up after a period of inactivity, and is not as fast as hosting it yourself. 


## Features

- **Authentication** - User registration and JWT-based authentication
- **Weekly meal plan** - Automatic weekly meal plan generation based on dietary preferences
- **Recipe swapping** - Swap out individual recipes you don't like
- **Regenerate** - Regenerate your entire meal plan with one click
- **Dietary preferences** - Dietary preference support (vegetarian, vegan, gluten free, and more)
- **Shopping list** - Auto-generated from your meal plan. Adding, checking off and deleting items
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

**Infrastructure**
- [Railway](https://railway.app/) — Backend hosting + managed PostgreSQL
- [Vercel](https://vercel.com/) — Frontend hosting
- [Docker](https://www.docker.com/) — Backend containerization
- [GitHub Actions](https://github.com/features/actions) — CI/CD

## Prerequisites

- [Go](https://go.dev/dl/) 1.21+
- [Node.js](https://nodejs.org/) 18+
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Goose](https://github.com/pressly/goose) — `go install github.com/pressly/goose/v3/cmd/goose@latest`
- A [Spoonacular API key](https://spoonacular.com/food-api)

## Getting Started for local hosting

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
ALLOWED_ORIGINS=http://localhost:5173
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
go run ./cmd/api
```

### 6. Start the frontend

In a new terminal window
```bash
cd frontend
npm install
npm run dev
```

The app will be available at `http://localhost:5173` by default. It may run on a different port if 5173 is occupied


## Project Structure

```
meal-planner/
├── .github/
│   └── workflows/
│       ├── backend.yml       # Backend CI
│       └── frontend.yml      # Frontend CI
├── backend/
│   ├── cmd/
│   │   └── api/              # Handlers, middleware, routing
│   ├── internal/
│   │   ├── auth/             # JWT generation and validation
│   │   ├── database/         # DB connection and migrations
│   │   ├── env/              # Environment variable helpers
│   │   ├── jsonutil/         # JSON response helpers
│   │   ├── recipeclient/     # Recipe client interface
│   │   ├── spoonacular/      # Spoonacular API client
│   │   └── store/            # Database layer and models
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── frontend/
│   └── src/
│       ├── api/              # Axios API calls and types
│       ├── components/       # Reusable components
│       ├── context/          # Auth context
│       └── pages/            # Page components
├── docker-compose.yml
└── README.md
```

## API Endpoints
 
### Auth
| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/users` | Register a new user |
| `POST` | `/users/login` | Login and receive JWT |
| `GET` | `/users/me` | Get current user |
| `PUT` | `/users/me/preferences` | Update dietary preferences |
 
### Meal Plans
| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/meal-plans/current` | Get or generate current meal plan |
| `PATCH` | `/meal-plans/current/recipe` | Swap a recipe for a specific day |
| `POST` | `/meal-plans/current/regenerate` | Generate a new meal plan |
 
### Shopping List
| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/shopping-list` | Get all shopping list items |
| `POST` | `/shopping-list/items` | Add items manually |
| `POST` | `/shopping-list/from-meal-plan` | Refresh ingredients from meal plan |
| `PATCH` | `/shopping-list/items/{id}` | Toggle item checked state |
| `DELETE` | `/shopping-list/items/{id}` | Delete a single item |
| `DELETE` | `/shopping-list/checked` | Delete all checked items |

## Architecture Notes
 
- **Recipe client interface** — the backend uses a `recipeclient.Client` interface, making it straightforward to swap out Spoonacular for another recipe API in the future
- **Ingredient normalization** — units are normalized and converted to base units before saving (e.g. tablespoons → teaspoons, cups → ml) so duplicate ingredients from different recipes are correctly combined
- **Bulk inserts** — shopping list items are inserted in a single query to minimize database round trips
- **Rate limiting** — all endpoints are rate limited to 100 requests per minute per IP, with stricter limits on auth endpoints

## Limitations
- The data from spoonacular is not optimal always, so the ingredients for the shopping list may look weird at times. 
- These are mostly american recipes, so they could refer to ingredients that don't really exist in your country