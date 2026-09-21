export interface MealPlanRecipe {
    id: number
    meal_plan_id: number
    recipe_id: number
    recipe_title: string
    image: string
    source_url: string
    day: number
    source: 'spoonacular' | 'user'
    user_recipe_id: number | null
}

export interface MealPlan {
    id: number
    user_id: number
    start_date: string
    end_date: string
    created_at: string
    recipes: MealPlanRecipe[]
}

export interface User {
    id: number
    email: string
    dietary_preferences: string[]
    intolerances: string[]
    excluded_ingredients: string[]
}

export const DIETARY_OPTIONS = [
    'vegetarian',
    'vegan',
    'gluten free',
    'dairy free',
    'ketogenic',
    'paleo',
] as const

export const INTOLERANCE_OPTIONS = [
    'dairy',
    'egg',
    'gluten',
    'grain',
    'peanut',
    'seafood',
    'sesame',
    'shellfish',
    'soy',
    'sulfite',
    'tree nut',
    'wheat'
] as const

export interface ShoppingListItem {
    id: number
    user_id: number
    name: string
    amount: number
    unit: string
    checked: boolean
    source: string
    created_at: string
}

export interface UserRecipeIngredient {
    id: number
    name: string
    amount: number
    unit: string
    position: number
}

export interface UserRecipe {
    id: number
    user_id: number
    title: string
    image: string
    source_url: string
    instructions: string
    servings: number
    ingredients: UserRecipeIngredient[] | null
    created_at: string
    updated_at: string
}

export interface RecipeInput {
    title: string
    image: string
    source_url: string
    instructions: string
    servings: number
    ingredients: { name: string; amount: number; unit: string }[]
}