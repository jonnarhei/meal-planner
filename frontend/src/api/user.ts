import client from "./client";
import type { User } from "./types";

export async function getCurrentUser(): Promise<User> {
    const response = await client.get('/users/me')
    return response.data 
}