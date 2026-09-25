import { Navigate, Route, Routes } from "react-router-dom"
import Login from "./pages/Login"
import MealPlanPage from "./pages/MealPlan"
import Register from "./pages/Register"
import { useAuth } from "./context/AuthContext"
import ProtectedRoute from "./components/ProtectedRoutes"
import AppLayout from "./components/AppLayout"
import Profile from "./pages/Profile"
import DietaryPreferences from "./pages/Preferences"
import ShoppingList from "./pages/ShoppingList"
import MyRecipes from "./pages/MyRecipes"
import RecipeForm from "./pages/RecipeForm"


function App() {
  const { isAuthenticated } = useAuth()

  return (
    <Routes>
      <Route path="/" element={
        isAuthenticated ? <Navigate to="/meal-plan" /> : <Navigate to="/login" />
      } />

      <Route path="/login" element={<Login />} />

      <Route path="/register" element={<Register />} />

      {/* Protected pages that live inside the app shell */}
      <Route element={
        <ProtectedRoute>
          <AppLayout />
        </ProtectedRoute>
      }>
        <Route path="/meal-plan" element={<MealPlanPage />} />

        <Route path="/profile" element={<Profile />} />

        <Route path="/shopping-list" element={<ShoppingList />} />

        <Route path="/recipes" element={<MyRecipes />} />
      </Route>

      {/* Protected pages that render on their own */}
      <Route path="/dietary-preferences" element={
        <ProtectedRoute>
          <DietaryPreferences />
        </ProtectedRoute>
      } />

      <Route path="/recipes/new" element={
        <ProtectedRoute>
          <RecipeForm />
        </ProtectedRoute>
      } />

      <Route path="/recipes/:id/edit" element={
        <ProtectedRoute>
          <RecipeForm />
        </ProtectedRoute>
      } />
    </Routes>
  )
}

export default App
