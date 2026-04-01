import { useNavigate } from "react-router-dom"
import { useAuth } from "../context/AuthContext"
import HamburgerMenu from "./HamburgerMenu"
import { useEffect, useState } from "react"
import { getCurrentUser } from "../api/user"
import type { User as UserType } from "../api/types"

function Profile() {
    const { setToken } = useAuth()
    const navigate = useNavigate()
    const [user, setUser] = useState<UserType | null>(null)
    const [error, setError] = useState<string | null>(null)
    const [loading, setLoading] = useState(true)



    useEffect(() => {
        const fetchUserData = async () => {
            try {
                const data = await getCurrentUser()
                setUser(data)
            } catch (err) {
                setError("Error fetching user data")
            } finally {
                setLoading(false)
            }
        }

        fetchUserData()
    }, [])

    const handleLogout = () => {
        setToken(null)
        navigate('/login')
    }

    if (loading) return (
        <div className="min-h-screen bg-orange-50 flex items-center justify-center">
            <p className="text-orange-500">Loading...</p>
        </div>
    )

    if (error) return (
        <div className="min-h-screen bg-orange-50 flex items-center justify-center">
            <p className="text-red-500">{error}</p>
        </div>
    )

    return (
        <div className="min-h-screen bg-orange-50">

            <div className="bg-white shadow-sm">
                <div className="max-w-screen-2xl mx-auto px-6 py-4 flex items-center gap-4">
                    <HamburgerMenu />
                    <h1 className="text-2xl font-bold text-orange-600">Profile</h1>
                </div>
            </div>

            <div className="max-w-lg mx-auto px-6 py-12">
                <div className="bg-white rounded-3x1 shadow-md border border-orange-100 p-8">
                    <div className="flex items-center gap-4 mb-8">
                        <div className="w-16 h-16 rounded-full bg-orange-200 flex items-center justify-center text-2xl">
                            👤
                        </div>
                        <div>
                            <h2 className="text-xl font-bold text-gray-800">Your Account</h2>
                        </div>
                    </div>

                    <div className="space-y-4">
                        <div className="bg-orange-100 rounded-2xl px-4 py-3">
                            <p className="text-xs font-semibold uppercase tracking-wide mb-1">Email</p>
                            <p className="text-gray-700 text-sm font-medium">{ user!.email }</p>
                        </div>
                    </div>

                    <button
                        onClick={handleLogout}
                        className="mt-8 w-full bg-red-200 hover:bg-red-300 text-red-500 font-semibold py-3 rounded-xl transition-colors"
                    >
                        Sign out
                    </button>
                </div>
            </div>

        </div>
    )
}


export default Profile