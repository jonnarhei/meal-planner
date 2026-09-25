import { useEffect, useState } from "react"

type Props = {
    src?: string
    alt?: string
    className?: string
    /** Width of one stripe in the placeholder gradient. */
    stripe?: number
}

/** A recipe image, or the striped orange placeholder when there is none. */
function RecipeThumb({ src, alt = '', className = '', stripe = 5 }: Props) {
    const [failed, setFailed] = useState(false)

    useEffect(() => setFailed(false), [src])

    if (src && !failed) {
        return (
            <img
                src={src}
                alt={alt}
                onError={() => setFailed(true)}
                className={`object-cover object-center ${className}`}
            />
        )
    }

    return (
        <div
            className={className}
            style={{
                background: `repeating-linear-gradient(135deg, #ffedd5 0 ${stripe}px, #fff7ed ${stripe}px ${stripe * 2}px)`
            }}
        />
    )
}

export default RecipeThumb
