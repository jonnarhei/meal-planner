import { useEffect, useRef, useState } from "react";
import { useAuth } from "../context/AuthContext";
import { useNavigate } from "react-router-dom";

function HamburgerMenu() {
    const [open, setOpen] = useState(false)
    
    const {setToken} = useAuth()
    const navigate = useNavigate()
    const menuRef = useRef<HTMLDivElement>(null)


    useEffect(() => {
        const handleClickOutside = (e: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
                setOpen(false)
            }
        }

        document.addEventListener('mousedown', handleClickOutside)
        return () => document.removeEventListener('mousedown', handleClickOutside)
    }, [])

    const handleLogout = () => {
        setToken(null)
        navigate('/login')
    }

    return (
        <div className="relative" ref={menuRef}>
            <button
                onClick={() => setOpen(prev => !prev)}
                className="flex flex-col gap-1.5 p-2 rounded-lg hover:bg-orange-50 transition-colors"
            >
                <span className={`block w-6 h-0.5 bg-gray-600 transition-all ${open ? 'rotate-45 translate-y-2' : ''}`} />
                <span className={`block w-6 h-0.5 bg-gray-600 transition-all ${open ? 'opacity-0' : ''}`} />
                <span className={`block w-6 h-0.5 bg-gray-600 transition-all ${open ? '-rotate-45 -translate-y-2' : ''}`} />
            </button>

            {open && (
                <div className="absolute left-0 top-12 bg-white rounded shadow-lg border border-orange-100 w-48 py-2 z-50">
                    <button 
                        onClick={() => { navigate('/profile'); setOpen(false) }}
                        className="w-full text-left px-4 py-2.5 text-sm text-gray-700 hover:bg-orange-50 transition-colors"
                    >
                        Profile
                    </button>
                    <hr className="my-1 border-orange-100" />
                    <button
                        onClick={() => { navigate('/meal-plan'); setOpen(false)}}
                        className="w-full text-left px-4 py-2.5 text-sm text-gray-700 hover:bg-orange-50 transition-colors"
                    >
                        Meal Plan
                    </button>
                    <hr className="my-1 border-orange-100" />
                    <button
                        onClick={() => {navigate('/shopping-list'); setOpen(false)}}
                        className="w-full text-left px-4 py-2.5 text-sm text-gray-700 hover:bg-orange-50 transition-colors"
                    >
                        Shopping List
                    </button>
                    <hr className="my-1 border-orange-100" />
                    <button
                        onClick={handleLogout}
                        className="w-full text-left px-4 py-2.5 text-sm text-red-500 hover:bg-red-50 transition-colors"
                    >
                        Log out
                    </button>
                </div>
            )}
        </div>
    )
}

export default HamburgerMenu