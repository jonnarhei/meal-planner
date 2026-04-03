import { useState } from "react"
import { useAuth } from "../context/AuthContext"
import { Link, useNavigate } from "react-router-dom"
import { login } from "../api/auth"
import toast from "react-hot-toast"

function Login() {
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [loading, setLoading] = useState(false)

    const { setToken } = useAuth()
    const navigate = useNavigate()

    const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        setLoading(true) 

        try {
            const data = await login({email, password})
            setToken(data.token)
            navigate('/meal-plan')
        } catch (err) {
            toast.error('Invalid email or password')
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="min-h-screen bg-orange-50 flex items-center justify-center">
            <div className="bg-white rounded-2xl shadow-md p-8 w-full max-w-md">
                <h1 className="text-3xl font-bold text-orange-600 mb-2">Meal Planner</h1>
                <p className="text-gray-500 mb-6">Sign into your account</p>

                <form onSubmit={handleSubmit} className="flex flex-col gap-4"> 
                    <input
                        type="email"
                        placeholder="Email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        className="border border-gray-200 rounded-lg px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-orange-300"
                    />
                    <input
                        type="password"
                        placeholder="Password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="border border-gray-200 rounded-lg px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-orange-300"
                    />
                    <button 
                        type="submit" 
                        disabled={loading}
                        className="bg-orange-500 hover:bg-orange-600 text-white font-semibold py-3 rounded-lg transition-colors diabled:opacity-50"
                    >
                        {loading ? 'Logging in...' : 'Log in'}
                    </button>
                </form>
                <p className="text-sm text-gray-500 mt-6 text-center">
                    Don't have an account? {' '} 
                    <Link to="/register" className="text-orange-500 hover:underline font-medium">
                        Register
                    </Link>
                </p>
            </div>
        </div>
        
    )
}


export default Login