import { Route, Routes } from "react-router-dom"
import Login from "./pages/Login"
import MealPlan from "./pages/MealPlan"
import Register from "./pages/Register"


function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route path="/meal-plan" element={<MealPlan />} />
    </Routes>
  )
}

export default App