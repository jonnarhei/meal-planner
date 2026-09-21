import client from "./client";
import type { RecipeInput, UserRecipe } from "./types";

export async function getRecipes(): Promise<UserRecipe[]> {
    const response = await client.get('/recipes')
    return response.data
}

export async function getRecipe(id: number): Promise<UserRecipe> {
    const response = await client.get(`/recipes/${id}`)
    return response.data
}

export async function createRecipe(recipe: RecipeInput): Promise<UserRecipe> {
    const response = await client.post('/recipes', recipe)
    return response.data
}

export async function updateRecipe(id: number, recipe: RecipeInput): Promise<UserRecipe> {
    const response = await client.put(`/recipes/${id}`, recipe)
    return response.data
}

export async function deleteRecipe(id: number): Promise<void> {
    await client.delete(`/recipes/${id}`)
}