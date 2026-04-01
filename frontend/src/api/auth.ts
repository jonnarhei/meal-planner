import client from "./client"

interface RegisterPayload {
    email: string
    password: string
}

interface LoginResponse {
    token: string
}

export async function register(payload: RegisterPayload) {
    const response = await client.post('/users', payload)
    return response.data
}

export async function login(payload: RegisterPayload): Promise<LoginResponse> {
    const response = await client.post('/users/login', payload)
    return response.data
}

