import { createContext, useContext, useState } from "react"

//shape of auth state
interface AuthState {
    token: string | null
    setToken: (token: string | null) => void
    isAuthenticated: boolean
}

//create context with default value
const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [token, setToken] = useState<string | null>(
        localStorage.getItem('token')
    )

    return (
        <AuthContext.Provider value={{
            token,
            setToken: (newToken) => {
                if (newToken) {
                    localStorage.setItem('token', newToken)
                } else {
                    localStorage.removeItem('token')
                }
                setToken(newToken)
            },
            isAuthenticated: token !== null
        }}>
            {children}
        </AuthContext.Provider>
    )
}

export function useAuth() {
    const context = useContext(AuthContext)
    if (!context) throw new Error('useAuth must be used within AuthProvider')
    return context
}