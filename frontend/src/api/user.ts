import client from "./client";
import type { User } from "./types";

export async function getCurrentUser(): Promise<User> {
    const response = await client.get('/users/me')
    return response.data 
}

export async function updatePreferences(preferences: string[], intolerances: string[], excludedIngredients: string[]): Promise<void> {
    await client.put('/users/me/preferences', { 
        dietary_preferences: preferences,
        intolerances,
        excluded_ingredients: excludedIngredients
    })
}