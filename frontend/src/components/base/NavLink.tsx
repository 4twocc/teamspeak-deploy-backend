// src/components/NavLink.tsx
import { Link, useLocation } from 'react-router-dom'
import { cn } from '@/lib/utils'

export function NavLink({ to, children }: { to: string; children: React.ReactNode }) {
	const { pathname } = useLocation()
	const isActive = pathname === to || pathname.startsWith(`${to}/`)

	return (
		<Link
			to={to}
			className={cn(
				'inline-flex items-center px-3 py-2 text-sm font-medium rounded-md',
				isActive ? 'bg-gray-900 text-white' : 'text-gray-700 hover:bg-gray-100 hover:text-gray-900'
			)}
		>
			{children}
		</Link>
	)
}
