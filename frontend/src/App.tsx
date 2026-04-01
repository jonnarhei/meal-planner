import { Navigate, Route, Routes } from "react-router-dom"
import Login from "./pages/Login"
import MealPlanPage from "./pages/MealPlan"
import Register from "./pages/Register"
import { useAuth } from "./context/AuthContext"
import ProtectedRoute from "./components/ProtectedRoutes"
import Profile from "./pages/Profile"


function App() {
  const { isAuthenticated } = useAuth()
  
  return (
    <Routes>
      <Route path="/" element={
        isAuthenticated ? <Navigate to="/meal-plan"/> : <Navigate to="/login"/>
        } />
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route path="/meal-plan" element={
        <ProtectedRoute>
          <MealPlanPage />
        </ProtectedRoute>
      } />
      <Route path="/profile" element={
        <ProtectedRoute>
          <Profile />
        </ProtectedRoute>
      }
      />
    </Routes>
  )
}

export default App